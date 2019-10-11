package utils

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

func isWritable(path string) bool {
	return unix.Access(path, unix.W_OK) == nil
}

func IsPathWritableDir(path string) error {
	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		return errors.Wrapf(err, "path %q does not exists", path)
	}
	if !stat.IsDir() {
		return errors.Wrapf(err, "path %q is not a directory", path)
	}
	if !isWritable(path) {
		return errors.Wrapf(err, "path %q is not writable", path)
	}
	return nil
}

func StoreHTTPResponseToFile(r *http.Response, fileName string, dataFeedName string) error {
	gr, err := gzip.NewReader(r.Body)
	defer gr.Close()
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"StatusCode": r.StatusCode, "DataFeedName": dataFeedName}).Error("gzip reader could not read NVD data feed")
		return err
	}

	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	defer f.Close()
	if err != nil {
		log.WithError(err).WithField("Filename", fileName).Warning("could not store NVD data feed to filesystem")
		return err
	}

	_, err = io.Copy(f, gr)
	if err != nil {
		log.WithError(err).WithField("Filename", fileName).Warning("could not stream NVD data feed to filesystem")
		return err
	}
	return nil
}

func StoreCVEsToFile(cves []NvdCVEEntry, fileName string) error {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		log.WithError(err).WithField("Filename", fileName).Warning("could not open file in truncate mode")
		return err
	}
	defer file.Close()
	m, err := json.Marshal(cves)
	if err != nil {
		log.WithError(err).WithField("Filename", fileName).Errorf("could not marshall %d CVEs", len(cves))
		return err
	}
	if _, err = file.WriteString(string(m)); err != nil {
		log.WithError(err).WithField("Filename", fileName).Warning("could not write marshalled CVEs list on file", fileName)
		return err
	}
	log.Infof("done writing file: %s", fileName)
	return nil
}
