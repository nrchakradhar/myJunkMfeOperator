// File: pkg/bundle/cdn/azure.go
package cdn

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
)

type AzureBlobUploader struct {
	client     *azblob.Client
	container  string
}

func NewAzureBlobUploader(connectionString, container string) (*AzureBlobUploader, error) {
	client, err := azblob.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure Blob client: %w", err)
	}
	return &AzureBlobUploader{
		client:    client,
		container: container,
	}, nil
}

func (u *AzureBlobUploader) Upload(ctx context.Context, localPath, remotePath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	blobPath := filepath.ToSlash(remotePath)
	blobPath = strings.TrimLeft(blobPath, "/")

	_, err = u.client.UploadFile(ctx, u.container, blobPath, file, &azblob.UploadFileOptions{
		HTTPHeaders: &azblob.HTTPHeaders{
			BlobContentType: to.Ptr("application/octet-stream"),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to upload to Azure Blob Storage: %w", err)
	}
	return nil
}
