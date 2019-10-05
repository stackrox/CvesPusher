package main

import (
	"compress/gzip"
	"encoding/json"
	"github.com/coreos/clair/pkg/commonerr"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
)

func isPathOk() error {
	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		return errors.Wrapf(err, "Path %q does not exists", path)
	}
	if !stat.IsDir() {
		return errors.Wrapf(err,"Path %q is not a directory", path)
	}
	if !writable(path) {
		return errors.Wrapf(err, "Path %q is not writable", path)
	}
	return nil
}

func storeHTTPResponseToFile(r *http.Response, fileName string, dataFeedName string) error {
	// Un-gzip it.
	gr, err := gzip.NewReader(r.Body)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"StatusCode": r.StatusCode, "DataFeedName": dataFeedName}).Error("could not read NVD data feed")
		return commonerr.ErrCouldNotDownload
	}

	// Store it to a file at the same time if possible.
	f, err := os.Create(fileName)
	if err != nil {
		log.WithError(err).WithField("Filename", fileName).Warning("could not store NVD data feed to filesystem")
		return commonerr.ErrFilesystem
	}
	defer f.Close()

	_, err = io.Copy(f, gr)
	if err != nil {
		log.WithError(err).WithField("Filename", fileName).Warning("could not stream NVD data feed to filesystem")
		return commonerr.ErrFilesystem
	}
	return nil
}

func storeCVEsToFile(cves []nvdCVEEntry, fileName string) error {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.WithError(err).WithField("Filename", fileName).Warning("could not open file in truncate mode")
		return commonerr.ErrFilesystem
	}
	for _, cve := range cves {
		m, err := json.Marshal(cve)
		if err != nil {
			log.WithError(err).WithField("Filename", fileName).Warning("could not marshal CVE %q", cve.CVE.Metadata.CVEID)
			continue
		}
		if _, err = file.WriteString(string(m)+"\n"); err != nil {
			log.WithError(err).WithField("Filename", fileName).Warning("could not write marshalled CVE %q on file %q", cve.CVE.Metadata.CVEID, fileName)
		}
	}
	return nil
}
