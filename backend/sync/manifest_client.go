package sync

import (
	"fmt"
	"os"
	"path/filepath"
)

// ManifestClient provides common functionality for discovering manifest files
type ManifestClient struct{}

// NewManifestClient creates a new manifest client
func NewManifestClient() *ManifestClient {
	return &ManifestClient{}
}

// FindManifests finds all manifest.yaml and manifest.yml files in the given directory
// If basePath is specified, it searches within that subdirectory and adjusts paths accordingly
func (m *ManifestClient) FindManifests(rootDir, basePath string) ([]string, error) {
	// Determine search directory based on base path
	searchDir := rootDir
	if basePath != "" {
		searchDir = filepath.Join(rootDir, basePath)
		// Check if base path exists
		if _, err := os.Stat(searchDir); os.IsNotExist(err) {
			return nil, fmt.Errorf("base path %s does not exist in directory %s", basePath, rootDir)
		}
	}

	var manifests []string

	// Find manifest.yaml files
	yamlFiles, err := m.findFiles(searchDir, "manifest.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to find manifest.yaml files: %w", err)
	}

	// If we have a base path, adjust the relative paths
	if basePath != "" {
		for i, file := range yamlFiles {
			yamlFiles[i] = filepath.Join(basePath, file)
		}
	}
	manifests = append(manifests, yamlFiles...)

	// Find manifest.yml files
	ymlFiles, err := m.findFiles(searchDir, "manifest.yml")
	if err != nil {
		return nil, fmt.Errorf("failed to find manifest.yml files: %w", err)
	}

	// If we have a base path, adjust the relative paths
	if basePath != "" {
		for i, file := range ymlFiles {
			ymlFiles[i] = filepath.Join(basePath, file)
		}
	}
	manifests = append(manifests, ymlFiles...)

	return manifests, nil
}

// GetFileContent reads the content of a file from the filesystem
func (m *ManifestClient) GetFileContent(rootDir, filePath string) ([]byte, error) {
	fullPath := filepath.Join(rootDir, filePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return content, nil
}

// ValidateBasePath checks if the base path exists within the root directory
func (m *ManifestClient) ValidateBasePath(rootDir, basePath string) error {
	if basePath == "" {
		return nil // No base path is valid
	}

	fullPath := filepath.Join(rootDir, basePath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("base path %s does not exist in directory %s", basePath, rootDir)
	}
	return nil
}

// findFiles recursively finds files with the given name
func (m *ManifestClient) findFiles(rootDir, fileName string) ([]string, error) {
	var files []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && info.Name() == fileName {
			// Get relative path from root directory
			relPath, err := filepath.Rel(rootDir, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}

		return nil
	})

	return files, err
}
