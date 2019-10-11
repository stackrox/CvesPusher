package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/coreos/clair/pkg/commonerr"
	"github.com/coreos/clair/pkg/httputil"
	errs "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stackrox/k8s-istio-cve-pusher/utils"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	dataFeedURL      = "https://nvd.nist.gov/feeds/json/cve/1.0/nvdcve-1.0-%s.json.gz"
	dataFeedMetaURL  = "https://nvd.nist.gov/feeds/json/cve/1.0/nvdcve-1.0-%s.meta"
	logDataFeedName  = "data feed name"
	kubernetes       = "k8s"
	istio            = "istio"
	fileExt          = ".json"
	k8sCveFileName = "NVDk8sCVEs.json"
	istioCveFileName = "NVDistioCVEs.json"
)

type NVDMetadata struct {
	CVSSv2 NVDmetadataCVSSv2
	CVSSv3 NVDmetadataCVSSv3
}

type NVDmetadataCVSSv2 struct {
	PublishedDateTime string
	Vectors           string
	Score             float64
}

type NVDmetadataCVSSv3 struct {
	Vectors             string
	Score               float64
	ExploitabilityScore float64
	ImpactScore         float64
}

func main() {
	if err := runCmd(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func runCmd() error {
	path, err := utils.IsEnvVarNonEmpty("LOCAL_FILE_PATH")
	if err != nil {
		return err
	}

	if err = utils.IsPathWritableDir(path); err != nil {
		return err
	}

	if err = utils.IsgcloudConfigOK(); err != nil {
		return err
	}

	feedReader, err := getDataFeeds(path)
	if err != nil {
		return errs.Wrap(err,"error downloading from NVD DB")
	}

	k8sCVEs, istioCVEs, err := getK8sAndIstioCVEs(feedReader)
	if err != nil {
		return err
	}

	k8sFileFullName := filepath.Join(path, k8sCveFileName)
	if err = utils.StoreCVEsToFile(k8sCVEs, k8sFileFullName); err != nil {
		return err
	}

	istioFileFullName := filepath.Join(path, istioCveFileName)
	if err = utils.StoreCVEsToFile(istioCVEs, istioFileFullName); err != nil {
		return err
	}

	bucket := os.Getenv(utils.GcloudBucket)
	if err = utils.WriteToBucket(bucket, filepath.Join(kubernetes,k8sCveFileName), k8sFileFullName); err != nil {
		return err
	}
	log.Infof("done pushing to bucket path %q", filepath.Join(bucket, kubernetes, k8sCveFileName))

	if err = utils.WriteToBucket(bucket, filepath.Join(istio, istioCveFileName), istioFileFullName); err != nil {
		return err
	}
	log.Infof("done pushing to bucket path %q", filepath.Join(bucket, istio, istioCveFileName))
	return nil
}

func getK8sAndIstioCVEs(feedReader map[string]string) ([]utils.NvdCVEEntry, []utils.NvdCVEEntry, error) {
	var allk8sCVEs, allistioCVEs []utils.NvdCVEEntry
	var totalk8sCVECount, totalistioCVECount int

	for feedName, fileName := range feedReader {
		dat, err := ioutil.ReadFile(fileName)
		if err != nil {
			return nil, nil, errs.Wrapf(err, "error reading file %q", fileName)
		}

		cves, err := utils.GetCVEs(dat)
		if err != nil {
			return nil, nil, errs.Wrapf(err, "error unmarshalling file %q", fileName)
		}

		k8sCVEs := filterK8sCVEs(cves)
		log.Infof("found %d k8s cves for year %s", len(k8sCVEs), feedName)
		allk8sCVEs = append(allk8sCVEs, k8sCVEs...)
		totalk8sCVECount += len(k8sCVEs)

		istioCVEs := filterIstioCVEs(cves)
		log.Infof("found %d istio cves for year %s", len(istioCVEs), feedName)
		allistioCVEs = append(allistioCVEs, istioCVEs...)
		totalistioCVECount += len(istioCVEs)
	}
	log.Infof("total %d k8s and %d istio CVEs found", totalk8sCVECount, totalistioCVECount)
	return allk8sCVEs, allistioCVEs, nil
}

func filterK8sCVEs(cves *utils.NvdCVEs) []utils.NvdCVEEntry {
	return filterCVEs(cves, "kubernetes")
}

func filterIstioCVEs(cves *utils.NvdCVEs) []utils.NvdCVEEntry {
	return filterCVEs(cves, "istio")
}

func filterCVEs(cves *utils.NvdCVEs, project string) []utils.NvdCVEEntry {
	var cveEntries []utils.NvdCVEEntry
	for _, cve := range cves.Entries {
		for _, vendorData := range cve.CVE.Affects.Vendor.VendorData {
			for _, productData := range vendorData.Product.ProductData {
				if productData.ProductName == project {
					cveEntries = append(cveEntries, cve)
				}
			}
		}
	}
	return cveEntries
}

func getDataFeeds(localPath string) (map[string]string, error) {
	dataFeedHashes := make(map[string]string)
	var dataFeedNames []string
	for y := 2014; y <= time.Now().Year(); y++ {
		dataFeedNames = append(dataFeedNames, strconv.Itoa(y))
	}

	// Get hashes for these feeds.
	for _, dataFeedName := range dataFeedNames {
		hash, err := getHashFromMetaURL(fmt.Sprintf(dataFeedMetaURL, dataFeedName))
		if err != nil {
			log.WithError(err).WithField(logDataFeedName, dataFeedName).Warning("could not get NVD data feed hash")
			continue
		}
		dataFeedHashes[dataFeedName] = hash
	}

	// Create map containing the name and filename for every data feed.
	dataFeedReaders := make(map[string]string)
	for _, dataFeedName := range dataFeedNames {
		fileName := filepath.Join(localPath, fmt.Sprintf("%s%s", dataFeedName, fileExt))
		if _, ok := dataFeedHashes[dataFeedName]; ok {
			// The hash is known, the disk should contains the feed. Try to read from it.
			if localPath != "" {
				if f, err := os.Open(fileName); err == nil {
					f.Close()
					dataFeedReaders[dataFeedName] = fileName
					continue
				}
			}

			err := downloadFeed(dataFeedName, fileName)
			if err != nil {
				return dataFeedReaders, err
			}
			dataFeedReaders[dataFeedName] = fileName
		}
	}
	return dataFeedReaders, nil
}

func downloadFeed(dataFeedName, fileName string) error {
	r, err := httputil.GetWithUserAgent(fmt.Sprintf(dataFeedURL, dataFeedName))
	if err != nil {
		log.WithError(err).WithField(logDataFeedName, dataFeedName).Error("could not download NVD data feed")
		return commonerr.ErrCouldNotDownload
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		log.WithFields(log.Fields{"StatusCode": r.StatusCode, "DataFeedName": dataFeedName}).Error("failed to download NVD data feed")
		return commonerr.ErrCouldNotDownload
	}

	if !isGzipResponse(r) {
		log.WithFields(log.Fields{"DataFeedName": dataFeedName}).Errorf("response is not a gzip encoded response")
		return fmt.Errorf("response for datafeed: %q is not gzip encoded", dataFeedName)
	}

	err = utils.StoreHTTPResponseToFile(r, fileName, dataFeedName)
	if err != nil {
		os.Exit(1)
	}
	return nil
}

func isGzipResponse(r *http.Response) bool {
	// Check Content-Encoding
	for _, s := range  r.Header["Content-Encoding"] {
		if s == "gzip" {
			return true
		}
	}
	// If Content-Encoding is not set, check Content-Type
	for _, s := range  r.Header["Content-Type"] {
		if strings.Contains(s,"gzip") {
			return true
		}
	}
	return false
}

func getHashFromMetaURL(metaURL string) (string, error) {
	r, err := httputil.GetWithUserAgent(metaURL)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%v failed status code: %d", metaURL, r.StatusCode)
	}

	scanner := bufio.NewScanner(r.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "sha256:") {
			return strings.TrimPrefix(line, "sha256:"), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", errors.New("invalid .meta file format")
}
