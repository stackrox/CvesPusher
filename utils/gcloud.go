package utils

import (
	"context"
	"time"

	"cloud.google.com/go/storage"
)

const (
	GCloudClientTimeout = 3 * time.Minute
)

func WriteToBucket(bucketName string, objectName string, data []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), GCloudClientTimeout)
	defer cancel()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	wc := client.Bucket(bucketName).Object(objectName).NewWriter(ctx)
	if _, err = wc.Write(data); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}
	return nil
}
