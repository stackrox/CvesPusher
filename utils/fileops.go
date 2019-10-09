package utils

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/coreos/clair/pkg/commonerr"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"io"
	"net/http"
	"os"
	"strings"
)

func isWritable(path string) bool {
	return unix.Access(path, unix.W_OK) == nil
}

func IsPathOk(path string) error {
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
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"StatusCode": r.StatusCode, "DataFeedName": dataFeedName}).Error("could not read NVD data feed")
		return commonerr.ErrCouldNotDownload
	}

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

func StoreCVEsToFile(cves []NvdCVEEntry, fileName string) error {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		log.WithError(err).WithField("Filename", fileName).Warning("could not open file in truncate mode")
		return commonerr.ErrFilesystem
	}
	defer file.Close()
	var marshalledCVEs []string
	for _, cve := range cves {
		m, err := json.Marshal(cve)
		if err != nil {
			log.WithError(err).WithField("Filename", fileName).Warning("could not marshal CVE %q", cve.CVE.Metadata.CVEID)
			continue
		}
		marshalledCVEs = append(marshalledCVEs, string(m))
	}
	m := fmt.Sprintf("[%s]", strings.Join(marshalledCVEs, ","))
	if _, err = file.WriteString(m); err != nil {
		log.WithError(err).WithField("Filename", fileName).Warning("could not write marshalled CVEs list on file", fileName)
	}
	log.Infof("done writing file: %s", fileName)
	return nil
}
