// File: pkg/bundle/cdn/upload.go
package cdn

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// UploadDirectoryToCDN walks a directory and uploads all files to the target CDN path.
func UploadDirectoryToCDN(ctx context.Context, cdn CDNClient, srcDir, cdnBasePath string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		cdnPath := filepath.ToSlash(filepath.Join(cdnBasePath, relPath))
		fmt.Printf("Uploading %s -> %s\n", path, cdnPath)
		return cdn.Upload(ctx, path, cdnPath)
	})
}
