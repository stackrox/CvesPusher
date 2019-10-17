package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stackrox/k8s-istio-cve-pusher/nvd"
	"github.com/stackrox/k8s-istio-cve-pusher/utils"
)

const (
	logDataFeedName = "data feed name"
	cveFileExt      = ".json"
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
	)
	flag.Parse()

	if !*flagDryRun && *flagGCSBucketName == "" {
		return errors.New("GCS bucket name is empty")
	}

	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}

	feedReader, err := getDataFeeds(tmpDir)
	if err != nil {
		return errors.Wrap(err, "error downloading from NVD DB")
	}

	k8sCVEs, istioCVEs, err := getK8sAndIstioCVEs(feedReader)
	if err != nil {
		return err
	}
	log.Printf("fetched %d total k8s CVEs", len(k8sCVEs))
	log.Printf("fetched %d total istio CVEs", len(istioCVEs))

	// Marshal k8s CVEs as json and compute checksum.
	k8sJson, k8sChecksum, err := marshalCVEs(k8sCVEs)
	if err != nil {
		return err
	}
	log.Printf("computed k8s checksum %s", k8sChecksum)

	// Marshal Istio CVEs as json and compute checksum.
	istioJson, istioChecksum, err := marshalCVEs(istioCVEs)
	if err != nil {
		return err
	}
	log.Printf("computed istio checksum %s", istioChecksum)

	// Stop early, is the dry run (-dry-run) flag was given.
	if *flagDryRun {
		log.Printf("skipping GCS upload since dry run was specified")
		return nil
	}

	// Upload k8s CVE json and checksum data to GCS bucket.
	if err := pushCVEsToBucket(nvd.Feeds[nvd.Kubernetes], k8sJson, k8sChecksum, *flagGCSBucketName, *flagGCSBucketPrefix); err != nil {
		return err
	}

	// Upload Istio CVE json and checksum data to GCS bucket.
	if err := pushCVEsToBucket(nvd.Feeds[nvd.Istio], istioJson, istioChecksum, *flagGCSBucketName, *flagGCSBucketPrefix); err != nil {
		return err
	}

	return nil
}

func marshalCVEs(entries []nvd.CVEEntry) ([]byte, string, error) {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return nil, "", err
	}

	hash := sha256.Sum256(data)
	checksum := hex.EncodeToString(hash[:])

	return data, checksum, nil
}

func pushCVEsToBucket(feed nvd.Feed, jsonData []byte, checksum string, gcsBucketName string, gcsBucketPrefix string) error {
	jsonObjectPath := filepath.Join(gcsBucketPrefix, feed.CVEFilename)
	if err := utils.WriteToBucket(gcsBucketName, jsonObjectPath, jsonData); err != nil {
		return err
	}
	log.Infof("pushed cve list to gs://%s/%s", gcsBucketName, jsonObjectPath)

	checksumObjectPath := filepath.Join(gcsBucketPrefix, feed.ChecksumFilename)
	if err := utils.WriteToBucket(gcsBucketName, checksumObjectPath, []byte(checksum)); err != nil {
		return err
	}
	log.Infof("pushed checksum to gs://%s/%s", gcsBucketName, checksumObjectPath)

	return nil
}

func getK8sAndIstioCVEs(feedReader map[string]string) ([]nvd.CVEEntry, []nvd.CVEEntry, error) {
	var allK8sCVEs, allIstioCVEs []nvd.CVEEntry

	for feedName, fileName := range feedReader {
		dat, err := ioutil.ReadFile(fileName)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "error reading file %q", fileName)
		}

		cves, err := nvd.Load(dat)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "error unmarshalling file %q", fileName)
		}

		k8sCVEs := cves.FilterProject("kubernetes")
		log.Infof("found %d k8s cves for year %s", len(k8sCVEs), feedName)
		allK8sCVEs = append(allK8sCVEs, k8sCVEs...)

		istioCVEs := cves.FilterProject("istio")
		log.Infof("found %d istio cves for year %s", len(istioCVEs), feedName)
		allIstioCVEs = append(allIstioCVEs, istioCVEs...)
	}
	log.Infof("total %d k8s and %d istio CVEs found", len(allK8sCVEs), len(allIstioCVEs))
	return allK8sCVEs, allIstioCVEs, nil
}

func getDataFeeds(path string) (map[string]string, error) {
	dataFeedReaders := make(map[string]string)

	for y := 2014; y <= time.Now().Year(); y++ {
		dataFeedName := strconv.Itoa(y)
		if err := validateURLMeta(utils.GetDataFeedMetaURL(dataFeedName)); err != nil {
			log.WithError(err).WithField(logDataFeedName, dataFeedName).Warning("could not get NVD data feed hash")
			return nil, err
		}

		fileName := filepath.Join(path, fmt.Sprintf("%s%s", dataFeedName, cveFileExt))
		if err := downloadFeed(dataFeedName, fileName); err != nil {
			return nil, err
		}
		dataFeedReaders[dataFeedName] = fileName
	}
	return dataFeedReaders, nil
}

func downloadFeed(dataFeedName, fileName string) error {
	r, err := utils.RunHTTPGet(utils.GetDataFeedURL(dataFeedName))
	if err != nil {
		log.WithError(err).WithField(logDataFeedName, dataFeedName).Error("could not download NVD data feed")
		return err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		log.WithFields(log.Fields{"StatusCode": r.StatusCode, "Status": r.Status, "DataFeedName": dataFeedName}).Error("failed to download NVD data feed")
		buf, err := utils.ReadNBytesFromResponse(r, 1024)
		if err != nil {
			return errors.Wrapf(err, "failed to get NVD data feed, additionally, there was an error reading the response body. status code: %d, status: %s", r.StatusCode, r.Status)
		}
		return fmt.Errorf("failed to download NVD data feed. status code: %d, status: %s, response body: %s", r.StatusCode, r.Status, string(buf))
	}

	err = utils.StoreHTTPResponseToFile(r, fileName)
	if err != nil {
		log.WithFields(log.Fields{"DataFeedName": dataFeedName}).Errorf("failed to store gzip response to file")
		return err
	}

	return nil
}

func validateURLMeta(metaURL string) error {
	r, err := utils.RunHTTPGet(metaURL)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		log.WithFields(log.Fields{"StatusCode": r.StatusCode, "Status": r.Status}).Error("failed to get NVD data feed meta")
		buf, err := utils.ReadNBytesFromResponse(r, 1024)
		if err != nil {
			return errors.Wrapf(err, "failed to get NVD data feed meta, additionally, there was an error reading the response body. status code: %d, status: %s", r.StatusCode, r.Status)
		}
		return fmt.Errorf("failed to get NVD data feed meta. status code: %d, status: %s, response body: %s", r.StatusCode, r.Status, string(buf))
	}
	return nil
}
