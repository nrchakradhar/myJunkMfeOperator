// File: pkg/bundle/cdn/interface.go
package cdn

import "context"

// CDNClient defines an interface for uploading files to a CDN
type CDNClient interface {
	Upload(ctx context.Context, localPath, remotePath string) error
}
