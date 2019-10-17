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
	"github.com/stackrox/k8s-istio-cve-pusher/nvd"
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

type component int

const (
	k8s component = iota
	istio
)

type fileAndBucketPaths struct {
	cveFileNameWithPath     string
	cveHashFileNameWithPath string
	cveBucketSubPath        string
	cveHashBucketSubPath    string
}

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

func storeK8sAndIstioCVEsToFile(k8sCVEs, istioCVEs []nvd.CVEEntry, path string) error {
	k8sPaths, err := getPaths(path, k8s)
	if err != nil {
		return err
	}
	if err := utils.StoreCVEsToFile(k8sCVEs, k8sPaths.cveFileNameWithPath, k8sPaths.cveHashFileNameWithPath); err != nil {
		return err
	}

	istioPaths, err := getPaths(path, istio)
	if err != nil {
		return err
	}
	if err := utils.StoreCVEsToFile(istioCVEs, istioPaths.cveFileNameWithPath, istioPaths.cveHashFileNameWithPath); err != nil {
		return err
	}

	return nil
}

func pushK8sAndIstioCVEsToBucket(path string) error {
	if err := pushCVEsToBucket(path, k8s); err != nil {
		return err
	}

	if err := pushCVEsToBucket(path, istio); err != nil {
		return err
	}

	return nil
}

func pushCVEsToBucket(path string, c component) error {
	bucketName := os.Getenv(utils.GCloudBucketEnvVar)
	paths, err := getPaths(path, c)
	if err != nil {
		return err
	}

	if err := utils.WriteToBucket(bucketName, paths.cveHashBucketSubPath, paths.cveHashFileNameWithPath); err != nil {
		return err
	}
	log.Infof("done pushing to bucket path %q", filepath.Join(bucketName, paths.cveHashBucketSubPath))

	if err := utils.WriteToBucket(bucketName, paths.cveBucketSubPath, paths.cveFileNameWithPath); err != nil {
		return err
	}
	log.Infof("done pushing to bucket path %q", filepath.Join(bucketName, paths.cveBucketSubPath))

	return nil
}

func getPaths(path string, c component) (fileAndBucketPaths, error) {
	switch c {
	case k8s:
		return fileAndBucketPaths{
			cveFileNameWithPath:     filepath.Join(path, k8sCveFileName),
			cveHashFileNameWithPath: filepath.Join(path, k8sCveHashFileName),
			cveBucketSubPath:        filepath.Join(os.Getenv(utils.GCloudBucketPrefix), kubernetesBucketDir, k8sCveFileName),
			cveHashBucketSubPath:    filepath.Join(os.Getenv(utils.GCloudBucketPrefix), kubernetesBucketDir, k8sCveHashFileName),
		}, nil

	case istio:
		return fileAndBucketPaths{
			cveFileNameWithPath:     filepath.Join(path, istioCveFileName),
			cveHashFileNameWithPath: filepath.Join(path, istioCveHashFileName),
			cveBucketSubPath:        filepath.Join(os.Getenv(utils.GCloudBucketPrefix), istioBucketDir, istioCveFileName),
			cveHashBucketSubPath:    filepath.Join(os.Getenv(utils.GCloudBucketPrefix), istioBucketDir, istioCveHashFileName),
		}, nil

	default:
		return fileAndBucketPaths{}, fmt.Errorf("unknown component type %d", c)
	}
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
