package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stackrox/k8s-istio-cve-pusher/utils"
)

const (
	logDataFeedName      = "data feed name"
	kubernetesBucketDir  = "k8s"
	istioBucketDir       = "istio"
	cveFileExt           = ".json"
	k8sCveFileName       = "NVDk8sCVEs.json"
	k8sCveHashFileName   = "k8sCVEsHash"
	istioCveFileName     = "NVDistioCVEs.json"
	istioCveHashFileName = "istioCVEsHash"
)

func main() {
	if err := runCmd(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func runCmd() error {
	path, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}

	feedReader, err := getDataFeeds(path)
	if err != nil {
		return errors.Wrap(err, "error downloading from NVD DB")
	}

	k8sCVEs, istioCVEs, err := getK8sAndIstioCVEs(feedReader)
	if err != nil {
		return err
	}

	if err = storeK8sAndIstioCVEsToFile(k8sCVEs, istioCVEs, path); err != nil {
		return err
	}

	if err = pushK8sAndIstioCVEsToBucket(path); err != nil {
		return err
	}

	return nil
}

func storeK8sAndIstioCVEsToFile(k8sCVEs, istioCVEs []utils.NvdCVEEntry, path string) error {
	k8sFileNameWithPath := filepath.Join(path, k8sCveFileName)
	k8sHashFileNameWithPath := filepath.Join(path, k8sCveHashFileName)
	if err := utils.StoreCVEsToFile(k8sCVEs, k8sFileNameWithPath, k8sHashFileNameWithPath); err != nil {
		return err
	}

	istioFileNameWithPath := filepath.Join(path, istioCveFileName)
	istioHashFileNameWithPath := filepath.Join(path, istioCveHashFileName)
	if err := utils.StoreCVEsToFile(istioCVEs, istioFileNameWithPath, istioHashFileNameWithPath); err != nil {
		return err
	}

	return nil
}

func pushK8sAndIstioCVEsToBucket(path string) error {
	bucket := os.Getenv(utils.GCloudBucket)
	k8sFileFullName := filepath.Join(path, k8sCveFileName)
	if err := utils.WriteToBucket(bucket, filepath.Join(kubernetesBucketDir, k8sCveFileName), k8sFileFullName); err != nil {
		return err
	}
	log.Infof("done pushing to bucket path %q", filepath.Join(bucket, kubernetesBucketDir, k8sCveFileName))

	istioFileFullName := filepath.Join(path, istioCveFileName)
	if err := utils.WriteToBucket(bucket, filepath.Join(istioBucketDir, istioCveFileName), istioFileFullName); err != nil {
		return err
	}
	log.Infof("done pushing to bucket path %q", filepath.Join(bucket, istioBucketDir, istioCveFileName))

	return nil
}

func getK8sAndIstioCVEs(feedReader map[string]string) ([]utils.NvdCVEEntry, []utils.NvdCVEEntry, error) {
	var allK8sCVEs, allIstioCVEs []utils.NvdCVEEntry

	for feedName, fileName := range feedReader {
		dat, err := ioutil.ReadFile(fileName)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "error reading file %q", fileName)
		}

		cves, err := utils.GetCVEs(dat)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "error unmarshalling file %q", fileName)
		}

		k8sCVEs := filterCVEs(cves, "kubernetes")
		log.Infof("found %d k8s cves for year %s", len(k8sCVEs), feedName)
		allK8sCVEs = append(allK8sCVEs, k8sCVEs...)

		istioCVEs := filterCVEs(cves, "istio")
		log.Infof("found %d istio cves for year %s", len(istioCVEs), feedName)
		allIstioCVEs = append(allIstioCVEs, istioCVEs...)
	}
	log.Infof("total %d k8s and %d istio CVEs found", len(allK8sCVEs), len(allIstioCVEs))
	return allK8sCVEs, allIstioCVEs, nil
}

func filterCVEs(cves *utils.NvdCVEs, project string) []utils.NvdCVEEntry {
	var cveEntries []utils.NvdCVEEntry
	for _, cve := range cves.Entries {
		if appliesToProject(cve, project) {
			cveEntries = append(cveEntries, cve)
		}
	}
	return cveEntries
}

func appliesToProject(cve utils.NvdCVEEntry, project string) bool {
	for _, vendorData := range cve.CVE.Affects.Vendor.VendorData {
		for _, productData := range vendorData.Product.ProductData {
			if productData.ProductName == project {
				return true
			}
		}
	}
	return false
}

func getDataFeeds(path string) (map[string]string, error) {
	dataFeedReaders := make(map[string]string)

	// Get hashes for these feeds.
	for y := 2014; y <= time.Now().Year(); y++ {
		dataFeedName := strconv.Itoa(y)
		if err := validateURLMeta(utils.GetDataFeedMetaURL(dataFeedName)); err != nil {
			log.WithError(err).WithField(logDataFeedName, dataFeedName).Warning("could not get NVD data feed hash")
			continue
		}

		fileName := filepath.Join(path, fmt.Sprintf("%s%s", dataFeedName, cveFileExt))
		if err := downloadFeed(dataFeedName, fileName); err != nil {
			return dataFeedReaders, err
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
			return errors.Wrapf(err, "failed to read response body, status code: %d, status: %s", r.StatusCode, r.Status)
		}
		return fmt.Errorf("failed to download NVD data feed. status code: %d, status: %s, response body: %s", r.StatusCode, r.Status, string(buf))
	}

	err = utils.StoreHTTPResponseToFile(r, fileName, dataFeedName)
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
			return errors.Wrapf(err, "failed to read response body, status code: %d, status: %s", r.StatusCode, r.Status)
		}
		return fmt.Errorf("failed to get NVD data feed meta. status code: %d, status: %s, response body: %s", r.StatusCode, r.Status, string(buf))
	}
	return nil
}
