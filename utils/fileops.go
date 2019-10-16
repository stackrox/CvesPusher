package utils

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

func StoreHTTPResponseToFile(r *http.Response, fileName string) error {
	contentReader := r.Body
	if isGzipResponse(r) {
		var err error
		contentReader, err = gzip.NewReader(r.Body)
		if err != nil {
			return err
		}
		defer contentReader.Close()
	}

	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		log.WithError(err).WithField("Filename", fileName).Warning("could not store NVD data feed to filesystem")
		return err
	}
	defer closeFile(f)

	_, err = io.Copy(f, contentReader)
	if err != nil {
		log.WithError(err).WithField("Filename", fileName).Warning("could not stream NVD data feed to filesystem")
		return err
	}
	return nil
}

func StoreCVEsToFile(cves []NvdCVEEntry, fileName, hashFileName string) error {
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		log.WithError(err).WithField("Filename", fileName).Warning("could not open file in truncate mode")
		return err
	}
	defer closeFile(f)

	m, err := json.Marshal(cves)
	if err != nil {
		log.WithError(err).WithField("Filename", fileName).Errorf("could not marshall %d CVEs", len(cves))
		return err
	}

	if _, err = f.Write(m); err != nil {
		log.WithError(err).WithField("Filename", fileName).Warning("could not write marshalled CVEs list on file", fileName)
		return err
	}
	log.Infof("done writing file: %s", fileName)

	sha256Bytes := sha256.Sum256(m)
	sha256String := hex.EncodeToString(sha256Bytes[:])

	if err := ioutil.WriteFile(hashFileName, []byte(sha256String), 0666); err != nil {
		log.WithError(err).WithField("Filename", hashFileName).Warning("could not write hash on file", hashFileName)
		return err
	}
	log.Infof("done writing file: %s", hashFileName)

	return nil
}

func closeFile(f *os.File) error {
	err := f.Close()
	if err != nil {
		return err
	}
	return nil
}
