package main

import (
	"bufio"
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
	"golang.org/x/sys/unix"
)

const (
	dataFeedURL     string = "https://nvd.nist.gov/feeds/json/cve/1.0/nvdcve-1.0-%s.json.gz"
	dataFeedMetaURL string = "https://nvd.nist.gov/feeds/json/cve/1.0/nvdcve-1.0-%s.meta"
	appenderName string = "NVD"
	logDataFeedName string = "data feed name"
)

var (
	path string = os.Getenv("CVES_FILE_PATH")
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

func writable(path string) bool {
	return unix.Access(path, unix.W_OK) == nil
}

func main() {
	fileMap := make(map[string]string)
	err := isPathOk()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	feedReader, _, err := getDataFeeds(fileMap, path)
	if err != nil {
		log.Errorf("Error downloading from NVD DB")
	}

	for feedName, fileName := range feedReader {
		dat, err := ioutil.ReadFile(fileName)
		if err != nil {
			log.Errorf("Error reading file %q, Error : %+v", fileName, err)
			return
		}
		cves, err := getCVEs(dat)
		if err != nil {
			log.Errorf("Error unmarshalling file %q, Error : %+v", fileName, err)
			return
		}
		k8sCVEs := filterK8sCVEs(cves)
		if len(k8sCVEs) > 0 {
			log.Infof("Found %d k8s cves for year %s\n", len(k8sCVEs), feedName)
			storeCVEsToFile(k8sCVEs, path + "k8sCVEs.json")
		}

		istioCVEs := filterIstioCVEs(cves)
		if len(istioCVEs) > 0 {
			log.Infof("Found %d istio cves for year %s\n", len(istioCVEs),  feedName)
			storeCVEsToFile(istioCVEs, path + "istioCVEs.json")
		}
	}
}

func filterK8sCVEs(cves *nvdCVEs) []nvdCVEEntry {
	return filterCVEs(cves, "kubernetes")
}

func filterIstioCVEs(cves *nvdCVEs) []nvdCVEEntry {
	return filterCVEs(cves, "istio")
}

func filterCVEs(cves *nvdCVEs, project string) []nvdCVEEntry {
	var cveEntries []nvdCVEEntry
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
			// It's not a big deal, no need interrupt, we're just going to download it again then.
			continue
		}
		dataFeedHashes[dataFeedName] = hash
	}

	// Create map containing the name and filename for every data feed.
	dataFeedReaders := make(map[string]string)
	for _, dataFeedName := range dataFeedNames {
		fileName := filepath.Join(localPath, fmt.Sprintf("%s.json", dataFeedName))
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
	err = storeHTTPResponseToFile(r, fileName, dataFeedName)
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