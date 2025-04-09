// File: pkg/bundle/cdn/cdn_test.go
package cdn_test

import (
	"cloud.google.com/go/storage"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	azblob "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"mfe-operator/pkg/bundle/cdn"
)

type MockCDNClient struct {
	mock.Mock
}

func (m *MockCDNClient) Upload(ctx context.Context, localPath, remotePath string) error {
	args := m.Called(ctx, localPath, remotePath)
	return args.Error(0)
}

func TestUploadDirectoryToCDN(t *testing.T) {
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "index.html")
	file2 := filepath.Join(tempDir, "js", "app.js")
	os.MkdirAll(filepath.Dir(file2), 0755)
	os.WriteFile(file1, []byte("<html></html>"), 0644)
	os.WriteFile(file2, []byte("console.log('hello');"), 0644)

	mockClient := new(MockCDNClient)
	mockClient.On("Upload", mock.Anything, file1, mock.MatchedBy(func(path string) bool {
		return strings.HasSuffix(path, "/index.html")
	})).Return(nil)
	mockClient.On("Upload", mock.Anything, file2, mock.MatchedBy(func(path string) bool {
		return strings.HasSuffix(path, "/js/app.js")
	})).Return(nil)

	err := cdn.UploadDirectoryToCDN(context.Background(), mockClient, tempDir, "cdn/mfe")
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestUploadFailsOnBadFile(t *testing.T) {
	tempDir := t.TempDir()
	badFile := filepath.Join(tempDir, "bad.txt")
	os.WriteFile(badFile, []byte("test"), 0644)
	os.Chmod(badFile, 0000) // make file unreadable
	defer os.Chmod(badFile, 0644) // reset permissions for cleanup

	mockClient := new(MockCDNClient)
	err := cdn.UploadDirectoryToCDN(context.Background(), mockClient, tempDir, "cdn/mfe")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, os.ErrPermission))
}

func TestS3UploaderIntegration(t *testing.T) {
	if os.Getenv("TEST_S3") != "true" {
		t.Skip("Skipping S3 integration test")
	}
	region := os.Getenv("AWS_REGION")
	bucket := os.Getenv("AWS_BUCKET")
	key := os.Getenv("AWS_ACCESS_KEY_ID")
	secret := os.Getenv("AWS_SECRET_ACCESS_KEY")
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(key, secret, ""),
	})
	assert.NoError(t, err)

	uploader := &cdn.S3Uploader{
		client: s3.New(sess),
		bucket: bucket,
	}
	tempFile := createTempFile(t)
	err = uploader.Upload(context.Background(), tempFile, fmt.Sprintf("test/%d/file.txt", time.Now().UnixNano()))
	assert.NoError(t, err)
}

func TestGCSUploaderIntegration(t *testing.T) {
	if os.Getenv("TEST_GCS") != "true" {
		t.Skip("Skipping GCS integration test")
	}
	ctx := context.Background()
	uploader, err := cdn.NewGCSUploader(ctx, os.Getenv("GCS_BUCKET"))
	assert.NoError(t, err)
	tempFile := createTempFile(t)
	err = uploader.Upload(ctx, tempFile, fmt.Sprintf("test/%d/file.txt", time.Now().UnixNano()))
	assert.NoError(t, err)
}

func TestAzureBlobUploaderIntegration(t *testing.T) {
	if os.Getenv("TEST_AZURE") != "true" {
		t.Skip("Skipping Azure integration test")
	}
	uploader, err := cdn.NewAzureBlobUploader(os.Getenv("AZURE_CONN_STR"), os.Getenv("AZURE_CONTAINER"))
	assert.NoError(t, err)
	tempFile := createTempFile(t)
	err = uploader.Upload(context.Background(), tempFile, fmt.Sprintf("test/%d/file.txt", time.Now().UnixNano()))
	assert.NoError(t, err)
}

func createTempFile(t *testing.T) string {
	tempFile, err := ioutil.TempFile("", "upload-test-*.txt")
	assert.NoError(t, err)
	_, err = tempFile.Write([]byte("upload test"))
	assert.NoError(t, err)
	tempFile.Close()
	return tempFile.Name()
}
