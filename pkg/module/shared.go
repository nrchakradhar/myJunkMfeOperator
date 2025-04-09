// File: pkg/module/shared.go
package module

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"mfe-operator/pkg/bundle/cdn"
)

// SharedModule represents a JS module shared by an MFE
type SharedModule struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Entry   string `json:"entry"`
}

// AnalyzeSharedModules scans an extracted bundle directory to find shared modules
func AnalyzeSharedModules(bundlePath string) ([]SharedModule, error) {
	var modules []SharedModule
	err := filepath.WalkDir(bundlePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if strings.HasSuffix(d.Name(), ".js") {
			matched, _ := filepath.Match("remoteEntry.js", d.Name())
			if matched {
				found, err := parseRemoteEntryForShared(path)
				if err != nil {
					return err
				}
				modules = append(modules, found...)
			}
		}
		return nil
	})
	return modules, err
}

// Very simple shared module detection from remoteEntry.js
func parseRemoteEntryForShared(path string) ([]SharedModule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// naive regex to catch e.g. react@18.2.0 references
	re := regexp.MustCompile(`"([a-zA-Z0-9-_]+)@([0-9]+\.[0-9]+\.[0-9]+)"`)
	matches := re.FindAllStringSubmatch(string(data), -1)
	modules := make([]SharedModule, 0, len(matches))
	for _, match := range matches {
		modules = append(modules, SharedModule{
			Name:    match[1],
			Version: match[2],
			Entry:   filepath.Base(path),
		})
	}
	return modules, nil
}

// UploadSharedModules uploads each module to the CDN under vendor/<name>@<version>/<entry>
func UploadSharedModules(ctx context.Context, uploader cdn.CDNClient, bundlePath string, modules []SharedModule) error {
	for _, m := range modules {
		srcPath := filepath.Join(bundlePath, m.Entry)
		remotePath := fmt.Sprintf("vendor/%s@%s/%s", m.Name, m.Version, m.Entry)
		if err := uploader.Upload(ctx, srcPath, remotePath); err != nil {
			return fmt.Errorf("uploading module %s: %w", m.Name, err)
		}
	}
	return nil
}

// SaveSharedModulesManifest saves shared modules metadata to a JSON file
func SaveSharedModulesManifest(path string, modules []SharedModule) error {
	data, err := json.MarshalIndent(modules, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
