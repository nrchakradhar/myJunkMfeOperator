// File: pkg/bundle/extract.go
package bundle

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ExtractTarball extracts the given tar.gz file using the desired strategy.
func ExtractTarball(ctx context.Context, tarballPath, baseOutputPath, crName string, strategy TarballNamingStrategy) (string, error) {
	destDir, err := ResolveOutputPath(strategy, baseOutputPath, crName, "extract")
	if err != nil {
		return "", err
	}

	fmt.Printf("Extracting tarball: %s to directory: %s\n", tarballPath, destDir)

	f, err := os.Open(tarballPath)
	if err != nil {
		return "", fmt.Errorf("failed to open tarball: %w", err)
	}
	defer f.Close()

	gzReader, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("error reading tar: %w", err)
		}

		targetPath := filepath.Join(destDir, hdr.Name)
		if err := ensureValidPath(destDir, targetPath); err != nil {
			return "", fmt.Errorf("invalid tar entry path: %w", err)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return "", fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
				return "", fmt.Errorf("failed to create parent directory: %w", err)
			}
			outFile, err := os.Create(targetPath)
			if err != nil {
				return "", fmt.Errorf("failed to create file: %w", err)
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return "", fmt.Errorf("failed to write file: %w", err)
			}
			outFile.Close()
		default:
			fmt.Printf("Skipping unsupported tar entry: %s\n", hdr.Name)
		}
	}

	fmt.Println("Extraction completed successfully.")
	return destDir, nil
}

// ensureValidPath ensures no directory traversal is possible
func ensureValidPath(basePath, targetPath string) error {
	realBase, err := filepath.Abs(basePath)
	if err != nil {
		return err
	}

	realTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return err
	}

	if !filepath.HasPrefix(realTarget, realBase) {
		return fmt.Errorf("path traversal detected: %s", targetPath)
	}
	return nil
}
