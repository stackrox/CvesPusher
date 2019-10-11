package utils

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"cloud.google.com/go/storage"
)

const (
	GcloudBucket   = "GOOGLE_CLOUD_BUCKET"
	GcloudAppCreds = "GOOGLE_APPLICATION_CREDENTIALS"
)

func IsgcloudConfigOK() error {
	// Check if bucket exists
	_, err := IsEnvVarNonEmpty(GcloudBucket)
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

func WriteToBucket(bucket, object, fileName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	wc := client.Bucket(bucket).Object(object).NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}
	return nil
}
