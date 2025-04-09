// File: pkg/bundle/fetch.go
package bundle

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/oci"
)

// FetchOCIArtifact downloads an OCI artifact to a local tarball using the given naming strategy.
func FetchOCIArtifact(ctx context.Context, ref string, baseOutputPath, crName string, strategy TarballNamingStrategy) (string, error) {
	outDir, err := ResolveOutputPath(strategy, baseOutputPath, crName, "fetch")
	if err != nil {
		return "", fmt.Errorf("failed to resolve output path: %w", err)
	}

	filePath := filepath.Join(outDir, "bundle.tar.gz")
	fmt.Printf("Fetching OCI artifact %s -> %s\n", ref, filePath)

	target, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer target.Close()

	repo, err := remote.NewRepository(ref)
	if err != nil {
		return "", fmt.Errorf("failed to create remote repository: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "mfe-oci-pull-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := oci.NewWithTemp(tmpDir)
	if err != nil {
		return "", fmt.Errorf("failed to create temp store: %w", err)
	}

	desc, err := oras.Copy(ctx, repo, "latest", store, "latest", oras.DefaultCopyOptions)
	if err != nil {
		return "", fmt.Errorf("failed to pull OCI artifact: %w", err)
	}

	blobReader, err := store.Fetch(ctx, desc)
	if err != nil {
		return "", fmt.Errorf("failed to fetch blob: %w", err)
	}
	defer blobReader.Close()

	if _, err := io.Copy(target, blobReader); err != nil {
		return "", fmt.Errorf("failed to write to file: %w", err)
	}

	fmt.Println("OCI fetch complete.")
	return filePath, nil
}
