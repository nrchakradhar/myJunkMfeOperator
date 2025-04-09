// File: pkg/bundle/fetch.go
package bundle

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// TarballNamingStrategy defines how tarball filenames are generated
type TarballNamingStrategy int

const (
	IsolatedTempDir TarballNamingStrategy = iota // Default: create isolated temp dir per CR
	UseCRName                                   // Use CR name to name tarball
	UseUUID                                     // Use UUID for filename uniqueness
)

// FetchOCIArtifact pulls the OCI artifact from the registry and saves it locally as a tar.gz
func FetchOCIArtifact(ctx context.Context, reference string, outputBasePath string, strategy TarballNamingStrategy, crName string) (string, error) {
	fmt.Println("Fetching OCI artifact:", reference)

	outputPath, tarballName, err := resolveOutputPath(strategy, outputBasePath, crName)
	if err != nil {
		return "", err
	}

	// Create a local file-based content store
	tempStore, err := file.New(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create local store: %w", err)
	}
	defer tempStore.Close()

	// Pull artifact from registry
	desc, err := oras.Copy(ctx, oras.DefaultRegistry, reference, tempStore, "", oras.DefaultCopyOptions)
	if err != nil {
		return "", fmt.Errorf("failed to pull artifact: %w", err)
	}

	// Create tarball
	tarballPath := filepath.Join(outputPath, tarballName)
	f, err := os.Create(tarballPath)
	if err != nil {
		return "", fmt.Errorf("failed to create tarball: %w", err)
	}
	defer f.Close()

	if err := tempStore.SaveAsTar(ctx, desc, f); err != nil {
		return "", fmt.Errorf("failed to write tarball: %w", err)
	}

	fmt.Println("Artifact saved to:", tarballPath)
	return tarballPath, nil
}

func resolveOutputPath(strategy TarballNamingStrategy, basePath, crName string) (string, string, error) {
	var outputPath string
	var tarballName string

	switch strategy {
	case IsolatedTempDir:
		fmt.Println("Using strategy: IsolatedTempDir (create temp directory per CR)")
		tempDir, err := os.MkdirTemp(basePath, "mfe-*")
		if err != nil {
			return "", "", fmt.Errorf("failed to create temp directory: %w", err)
		}
		outputPath = tempDir
		tarballName = "bundle.tar.gz"
		fmt.Printf("Created temp directory: %s\n", outputPath)

	case UseCRName:
		fmt.Println("Using strategy: UseCRName (name tarball using CR name)")
		tarballName = fmt.Sprintf("%s.tar.gz", sanitizeName(crName))
		if tarballName == ".tar.gz" {
			fmt.Println("Invalid CR name, falling back to IsolatedTempDir")
			tempDir, err := os.MkdirTemp(basePath, "mfe-*")
			if err != nil {
				return "", "", fmt.Errorf("failed to create temp directory: %w", err)
			}
			outputPath = tempDir
			tarballName = "bundle.tar.gz"
			fmt.Printf("Created temp directory: %s\n", outputPath)
		} else {
			outputPath = basePath
			fmt.Printf("Output path: %s, Tarball name: %s\n", outputPath, tarballName)
		}

	case UseUUID:
		fmt.Println("Using strategy: UseUUID (unique tarball name using UUID)")
		outputPath = basePath
		tarballName = fmt.Sprintf("bundle-%s.tar.gz", uuid.NewString())
		fmt.Printf("Output path: %s, Tarball name: %s\n", outputPath, tarballName)

	default:
		fmt.Println("Invalid strategy, falling back to IsolatedTempDir")
		tempDir, err := os.MkdirTemp(basePath, "mfe-*")
		if err != nil {
			return "", "", fmt.Errorf("failed to create temp directory: %w", err)
		}
		outputPath = tempDir
		tarballName = "bundle.tar.gz"
		fmt.Printf("Created temp directory: %s\n", outputPath)
	}

	return outputPath, tarballName, nil
}

// sanitizeName replaces illegal filename characters
func sanitizeName(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9-_]`)
	return re.ReplaceAllString(strings.ReplaceAll(name, "/", "-"), "_")
}
