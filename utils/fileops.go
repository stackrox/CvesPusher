package utils

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

func StoreHTTPResponseToFile(r *http.Response, fileName string, dataFeedName string) error {
	contentReader := r.Body
	if IsGzipResponse(r) {
		var err error
		contentReader, err = gzip.NewReader(r.Body)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{"StatusCode": r.StatusCode, "DataFeedName": dataFeedName}).Error("gzip reader could not read NVD data feed")
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

func StoreCVEsToFile(cves []NvdCVEEntry, fileName string) error {
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
	return nil
}

//func ReadFromFile(fileName string) (string, error) {
//	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDONLY, 0666)
//	if err != nil {
//		log.WithError(err).WithField("Filename", fileName).Warning("could not open file in read mode")
//		return "", err
//	}
//	defer closeFile(f)
//
//	data, err := ioutil.ReadAll(f)
//	if err != nil {
//		return "", err
//	}
//
//	return string(data), nil
//}
//
//func WriteToFile(fileName, data string) error {
//	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0666)
//	if err != nil {
//		log.WithError(err).WithField("Filename", fileName).Warning("could not open file in write mode")
//		return err
//	}
//	defer closeFile(f)
//
//	err = ioutil.WriteFile(fileName, []byte(data), 0666)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}

func closeFile(f *os.File) error {
	err := f.Close()
	if err != nil {
		return err
	}
	return nil
}
