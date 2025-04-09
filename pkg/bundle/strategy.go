// File: pkg/bundle/strategy.go
package bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

type TarballNamingStrategy int

const (
	IsolatedTempDir TarballNamingStrategy = iota
	UseCRName
	UseUUID
)

// Sanitize name to be safe for filenames and folder names
func SanitizeName(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9-_]`)
	return re.ReplaceAllString(strings.ReplaceAll(name, "/", "-"), "_")
}

// ResolveOutputPath resolves an output path using the given strategy
func ResolveOutputPath(strategy TarballNamingStrategy, basePath, crName, prefix string) (string, error) {
	switch strategy {
	case IsolatedTempDir:
		fmt.Printf("Using strategy: IsolatedTempDir for %s\n", prefix)
		return os.MkdirTemp(basePath, fmt.Sprintf("mfe-%s-*", prefix))

	case UseCRName:
		fmt.Printf("Using strategy: UseCRName for %s\n", prefix)
		sanitized := SanitizeName(crName)
		if sanitized == "" {
			return ResolveOutputPath(IsolatedTempDir, basePath, crName, prefix)
		}
		dirPath := filepath.Join(basePath, sanitized)
		if err := os.MkdirAll(dirPath, 0o755); err != nil {
			return "", fmt.Errorf("failed to create named dir: %w", err)
		}
		return dirPath, nil

	case UseUUID:
		fmt.Printf("Using strategy: UseUUID for %s\n", prefix)
		dirPath := filepath.Join(basePath, fmt.Sprintf("%s-%s", prefix, uuid.NewString()))
		if err := os.MkdirAll(dirPath, 0o755); err != nil {
			return "", fmt.Errorf("failed to create uuid dir: %w", err)
		}
		return dirPath, nil

	default:
		fmt.Printf("Invalid strategy, defaulting to IsolatedTempDir for %s\n", prefix)
		return ResolveOutputPath(IsolatedTempDir, basePath, crName, prefix)
	}
}
