// File: pkg/bundle/cdn/s3.go
package cdn

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Uploader struct {
	client     *s3.S3
	bucket     string
	cdnBaseURL string
}

func NewS3Uploader(region, bucket, accessKey, secretKey string) (*S3Uploader, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
		Credentials: credentials.NewStaticCredentials(
			accessKey,
			secretKey,
			"", // token
		),
	})
	if err != nil {
		return nil, err
	}

	return &S3Uploader{
		client: s3.New(sess),
		bucket: bucket,
	}, nil
}

func (u *S3Uploader) Upload(ctx context.Context, localPath, remotePath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", localPath, err)
	}
	defer file.Close()

	_, err = u.client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(filepath.ToSlash(remotePath)),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("S3 upload failed: %w", err)
	}
	return nil
}
