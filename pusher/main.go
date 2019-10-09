package main

import (
	"bufio"
	"github.com/stackrox/CvesPusher/utils"
	"io/ioutil"

	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/clair/pkg/commonerr"
	"github.com/coreos/clair/pkg/httputil"
	log "github.com/sirupsen/logrus"
)

const (
	dataFeedURL     string = "https://nvd.nist.gov/feeds/json/cve/1.0/nvdcve-1.0-%s.json.gz"
	dataFeedMetaURL string = "https://nvd.nist.gov/feeds/json/cve/1.0/nvdcve-1.0-%s.meta"
	logDataFeedName string = "data feed name"
	filePrefix      string = "NVD"
	kubernetes      string = "k8s"
	istio           string = "istio"
	fileExt         string = ".json"
	cves            string = "CVEs"
)

type appender struct {
	localPath      string
	dataFeedHashes map[string]string
	metadata       map[string]NVDMetadata
}

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
	fileMap := make(map[string]string)

	path, err := utils.IsEnvVarNonEmpty("LOCAL_FILE_PATH")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	err = utils.IsPathOk(path)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	feedReader, _, err := getDataFeeds(fileMap, path)
	if err != nil {
		log.Errorf("error downloading from NVD DB")
	}

	var allk8sCVEs, allistioCVEs []utils.NvdCVEEntry

	for feedName, fileName := range feedReader {
		dat, err := ioutil.ReadFile(fileName)
		if err != nil {
			log.Errorf("error reading file %q, error : %+v", fileName, err)
			return
		}
		cves, err := utils.GetCVEs(dat)
		if err != nil {
			log.Errorf("error unmarshalling file %q, error : %+v", fileName, err)
			return
		}

		k8sCVEs := filterK8sCVEs(cves)
		log.Infof("found %d k8s cves for year %s", len(k8sCVEs), feedName)
		allk8sCVEs = append(allk8sCVEs, k8sCVEs...)

		istioCVEs := filterIstioCVEs(cves)
		log.Infof("found %d istio cves for year %s", len(istioCVEs), feedName)
		allistioCVEs = append(allistioCVEs, istioCVEs...)
	}

	k8sFileName := fmt.Sprintf("%s%s%s%s", filePrefix, kubernetes, cves, fileExt)
	k8sFileFullName := fmt.Sprintf("%s%s", path, k8sFileName)
	utils.StoreCVEsToFile(allk8sCVEs, k8sFileFullName)

	istioFileName := fmt.Sprintf("%s%s%s%s", filePrefix, istio, cves, fileExt)
	istioFileFullName := fmt.Sprintf("%s%s", path, istioFileName)
	utils.StoreCVEsToFile(allistioCVEs, istioFileFullName)

	err = utils.WriteToBucket(kubernetes+"/"+k8sFileName, k8sFileFullName)
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("done pushing to bucket path %q", fmt.Sprintf("%s/%s/%s", os.Getenv(utils.GcloudBucket), kubernetes, k8sFileName))

	err = utils.WriteToBucket(istio+"/"+istioFileName, istioFileFullName)
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("done pushing to bucket path %q", fmt.Sprintf("%s/%s/%s", os.Getenv(utils.GcloudBucket), istio, istioFileName))
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

func getDataFeeds(dataFeedHashes map[string]string, localPath string) (map[string]string, map[string]string, error) {
	var dataFeedNames []string
	for y := 2010; y <= time.Now().Year(); y++ {
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
		if h, ok := dataFeedHashes[dataFeedName]; ok && h == dataFeedHashes[dataFeedName] {
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
				return dataFeedReaders, dataFeedHashes, err
			}
			dataFeedReaders[dataFeedName] = fileName
		}
	}
	return dataFeedReaders, dataFeedHashes, nil
}

func downloadFeed(dataFeedName, fileName string) error {
	r, err := httputil.GetWithUserAgent(fmt.Sprintf(dataFeedURL, dataFeedName))
	if err != nil {
		log.WithError(err).WithField(logDataFeedName, dataFeedName).Error("could not download NVD data feed")
		return commonerr.ErrCouldNotDownload
	}
	defer r.Body.Close()

	if !httputil.Status2xx(r) {
		log.WithFields(log.Fields{"StatusCode": r.StatusCode, "DataFeedName": dataFeedName}).Error("Failed to download NVD data feed")
		return commonerr.ErrCouldNotDownload
	}
	err = utils.StoreHTTPResponseToFile(r, fileName, dataFeedName)
	if err != nil {
		os.Exit(1)
	}
	return nil
}

func getHashFromMetaURL(metaURL string) (string, error) {
	r, err := httputil.GetWithUserAgent(metaURL)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()

	if !httputil.Status2xx(r) {
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
