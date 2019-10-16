package utils

import (
	"context"
	"io"
	"os"
	"time"

	"cloud.google.com/go/storage"
)

const (
	GCloudBucketEnvVar = "GOOGLE_CLOUD_BUCKET"
	GCloudBucketPrefix = "GOOGLE_CLOUD_BUCKET_PREFIX"

	GCloudClientTimeout = 3*time.Minute
)

func WriteToBucket(bucket, object, fileName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), GCloudClientTimeout)
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
