package main

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/facebookincubator/nvdtools/cvefeed/nvd/schema"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stackrox/k8s-istio-cve-pusher/nvd"
	"github.com/stackrox/k8s-istio-cve-pusher/utils"
)

var (
	// specs is the set of projects that we wish to filter CVEs for.
	specs = map[nvd.Project]string{
		nvd.Istio:      "istio",
		nvd.Kubernetes: "kubernetes",
		nvd.Openshift:  "openshift",
	}
)

func main() {
	if err := runCmd(); err != nil {
		log.Fatalf("cve-pusher: %v", err)
		os.Exit(1)
	}
}

func runCmd() error {
	var (
		flagGCSBucketName   = flag.String("gcs-bucket-name", "", "GCS bucket name to upload CVE data to")
		flagGCSBucketPrefix = flag.String("gcs-bucket-prefix", "", "GCS bucket prefix to upload CVE data under")
		flagDryRun          = flag.Bool("dry-run", false, "Skip uploading CVE data to GCS")
		flagOutputDir       = flag.String("output-dir", "", "Directory to write output, locally, if desired")
	)
	flag.Parse()

	if *flagDryRun {
		log.Info("Running in dry-run mode...")
	} else if *flagGCSBucketName == "" {
		return errors.New("-gcs-bucket-name is empty and -dry-run was not set")
	}

	// Generate a list of NVD urls for fetching CVE data.
	dataFeedURLs := generateDataFeedURLs()

	// Download all the NVD data feeds, and unmarshal CVEs into a channel.
	allCVEs := downloadDataFeeds(dataFeedURLs)

	// Split the set of all CVEs into groups based on which project they apply to.
	groups := splitCVEs(specs, allCVEs)

	// Update fixed by version for NVD CVEs that do not have it but has been determined by manual audit.
	for project, cves := range groups {
		groups[project] = nvd.FillMissingData(project, cves)
	}

	// Do some logging for the resulting groups of CVEs.
	for project, cves := range groups {
		log.Infof("%d total CVEs for project %s\n", len(cves), project)
		for index, cve := range cves {
			log.Infof("[%d/%d] - %s", index+1, len(cves), cve.CVE.CVEDataMeta.ID)
		}
	}

	// Generate json and checksum data for each group of CVEs, and upload that
	// data into a GCS bucket.
	for project, cves := range groups {
		// Marshal CVEs as json and compute checksum.
		jsonData, checksumData, err := marshalCVEs(cves)
		if err != nil {
			return err
		}
		log.Infof("computed %v checksum %s", project, checksumData)

		if !*flagDryRun {
			// Upload CVE json and checksum data to GCS bucket.
			if err := pushCVEsToBucket(nvd.Feeds[project], jsonData, checksumData, *flagGCSBucketName, *flagGCSBucketPrefix); err != nil {
				return err
			}
		}
		if *flagOutputDir != "" {
			// Write CVE json and checksum data to output-path.
			if err := writeToDir(nvd.Feeds[project], jsonData, checksumData, *flagOutputDir); err != nil {
				return err
			}
		}
	}

	return nil
}

func marshalCVEs(entries []*schema.NVDCVEFeedJSON10DefCVEItem) ([]byte, string, error) {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return nil, "", err
	}

	hash := sha256.Sum256(data)
	checksum := hex.EncodeToString(hash[:])

	return data, checksum, nil
}

func pushCVEsToBucket(feed nvd.Feed, jsonData []byte, checksum string, gcsBucketName string, gcsBucketPrefix string) error {
	log.Infof("Writing %s data to gs://%s/%s...", feed.Name, gcsBucketName, gcsBucketPrefix)

	jsonObjectPath := path.Join(gcsBucketPrefix, feed.CVEFilename)
	if err := utils.WriteToBucket(gcsBucketName, jsonObjectPath, jsonData); err != nil {
		return errors.Wrap(err, "writing JSON data to GCS")
	}

	checksumObjectPath := path.Join(gcsBucketPrefix, feed.ChecksumFilename)
	if err := utils.WriteToBucket(gcsBucketName, checksumObjectPath, []byte(checksum)); err != nil {
		return errors.Wrap(err, "writing checksum data to GCS")
	}

	return nil
}

func writeToDir(feed nvd.Feed, jsonData []byte, checksum, dir string) error {
	log.Infof("Writing %s data to %s...", feed.Name, dir)

	jsonPath := filepath.Join(dir, feed.CVEFilename)
	err := os.MkdirAll(filepath.Dir(jsonPath), 0755)
	if err != nil {
		return errors.Wrap(err, "creating directory for JSON data")
	}
	checksumPath := filepath.Join(dir, feed.ChecksumFilename)
	err = os.MkdirAll(filepath.Dir(checksumPath), 0755)
	if err != nil {
		return errors.Wrap(err, "creating directory for checksum data")
	}

	err = os.WriteFile(jsonPath, jsonData, 0755)
	if err != nil {
		return errors.Wrap(err, "writing JSON data")
	}
	err = os.WriteFile(checksumPath, []byte(checksum), 0755)
	if err != nil {
		return errors.Wrap(err, "writing checksum data")
	}

	return nil
}

func splitCVEs(specs map[nvd.Project]string, allCVEs []*schema.NVDCVEFeedJSON10DefCVEItem) map[nvd.Project][]*schema.NVDCVEFeedJSON10DefCVEItem {
	entries := make(map[nvd.Project][]*schema.NVDCVEFeedJSON10DefCVEItem, len(specs))
	for _, cve := range allCVEs {
		for project := range specs {
			if nvd.CVEAppliesToProject(project, cve) {
				entries[project] = append(entries[project], cve)
			}
		}
	}

	for _, cves := range entries {
		// Sort CVEs by CVE ID.
		sort.Slice(cves, func(i, j int) bool {
			return cves[i].CVE.CVEDataMeta.ID < cves[j].CVE.CVEDataMeta.ID
		})
	}

	return entries
}

func generateDataFeedURLs() []string {
	var feeds []string
	for year := 2014; year <= time.Now().Year(); year++ {
		feeds = append(feeds, fmt.Sprintf("https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-%d.json.gz", year))
	}
	return feeds
}

func downloadDataFeeds(feedURLs []string) []*schema.NVDCVEFeedJSON10DefCVEItem {
	cveChan := make(chan schema.NVDCVEFeedJSON10)
	var wg sync.WaitGroup

	for _, feedURL := range feedURLs {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			// HTTP GET the feed URL.
			resp, err := utils.RunHTTPGet(url)
			if err != nil {
				log.Errorf("failed to fetch feed %s: %v", url, err)
				return
			}
			defer resp.Body.Close()

			// Ensure that we received a 200 OK from the feed url.
			if resp.StatusCode != http.StatusOK {
				log.Errorf("failed to fetch feed %s: %s", url, resp.Status)
			}
			log.Infof("successfully fetched feed %s", url)

			gr, err := gzip.NewReader(resp.Body)
			if err != nil {
				log.Errorf("failed to gunzip feed %s: %v", url, err)
			}
			defer gr.Close()

			cveFeed, err := nvd.LoadReader(gr)
			if err != nil {
				log.Errorf("failed to parse feed %s: %v", url, err)
				return
			}

			cveChan <- *cveFeed
			log.Infof("successfully parsed feed %s: found %d CVEs", url, len(cveFeed.CVEItems))
		}(feedURL)
	}

	go func() {
		wg.Wait()
		close(cveChan)
	}()

	var allCVEs []*schema.NVDCVEFeedJSON10DefCVEItem
	for cveFeed := range cveChan {
		allCVEs = append(allCVEs, cveFeed.CVEItems...)
	}

	return allCVEs
}
