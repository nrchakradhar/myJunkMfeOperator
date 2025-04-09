// File: pkg/bundle/cdn/gcs.go
package cdn

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type GCSUploader struct {
	client     *storage.Client
	bucketName string
}

func NewGCSUploader(ctx context.Context, bucketName string) (*GCSUploader, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}
	return &GCSUploader{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (u *GCSUploader) Upload(ctx context.Context, localPath, remotePath string) error {
	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer f.Close()

	remotePath = filepath.ToSlash(remotePath)
	w := u.client.Bucket(u.bucketName).Object(remotePath).NewWriter(ctx)
	defer w.Close()

	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("failed to upload to GCS: %w", err)
	}
	return nil
}
