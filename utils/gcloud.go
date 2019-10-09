package utils

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"io"
	"log"
	"os"
)

const (
	GcloudBucket   = "GOOGLE_CLOUD_BUCKET"
	GcloudProject  = "GOOGLE_CLOUD_PROJECT"
	GcloudAppCreds = "GOOGLE_APPLICATION_CREDENTIALS"
)

func isgcloudConfigOK() error {
	// Check if project exists
	_, err := IsEnvVarNonEmpty(GcloudProject)
	if err != nil {
		return err
	}

	// Check if bucket exists
	_, err = IsEnvVarNonEmpty(GcloudBucket)
	if err != nil {
		return err
	}

	// Check if creds file exists
	credsFilePath, err := IsEnvVarNonEmpty(GcloudAppCreds)
	if err != nil {
		return err
	}
	stat, err := os.Stat(credsFilePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("path %q specified by GOOGLE_APPLICATION_CREDENTIALS environment variable does not exist", credsFilePath)
	}
	if stat.IsDir() {
		return fmt.Errorf("path %q specified by GOOGLE_APPLICATION_CREDENTIALS environment variable is a dictory", credsFilePath)
	}
	return nil
}

func WriteToBucket(object, fileName string) error {
	err := isgcloudConfigOK()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	bucket := os.Getenv(GcloudBucket)
	wc := client.Bucket(bucket).Object(object).NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}
	return nil
}
