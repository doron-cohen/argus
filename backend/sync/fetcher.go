package sync

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/doron-cohen/argus/backend/internal/models"
)

// Manifest represents a loaded and parsed manifest
type Manifest struct {
	Path    string
	Content *models.Manifest
}

// LoadManifests loads all manifest.yaml and manifest.yml files from the given path
// Returns a map of file paths to their parsed manifest content
func LoadManifests(ctx context.Context, searchPath string) (map[string]Manifest, error) {
	// Check if search directory exists
	if _, err := os.Stat(searchPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory %s does not exist", searchPath)
	}

	manifests := make(map[string]Manifest)
	parser := models.NewParser()

	// Find and load manifest.yaml files
	yamlFiles, err := findManifestFiles(searchPath, "manifest.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to find manifest.yaml files: %w", err)
	}

	for _, filePath := range yamlFiles {
		content, err := os.ReadFile(filepath.Join(searchPath, filePath))
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		parsedManifest, err := parser.Parse(content)
		if err != nil {
			return nil, fmt.Errorf("failed to parse manifest %s: %w", filePath, err)
		}

		if err := parser.Validate(parsedManifest); err != nil {
			return nil, fmt.Errorf("invalid manifest %s: %w", filePath, err)
		}

		manifests[filePath] = Manifest{
			Path:    filePath,
			Content: parsedManifest,
		}
	}

	// Find and load manifest.yml files
	ymlFiles, err := findManifestFiles(searchPath, "manifest.yml")
	if err != nil {
		return nil, fmt.Errorf("failed to find manifest.yml files: %w", err)
	}

	for _, filePath := range ymlFiles {
		content, err := os.ReadFile(filepath.Join(searchPath, filePath))
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		parsedManifest, err := parser.Parse(content)
		if err != nil {
			return nil, fmt.Errorf("failed to parse manifest %s: %w", filePath, err)
		}

		if err := parser.Validate(parsedManifest); err != nil {
			return nil, fmt.Errorf("invalid manifest %s: %w", filePath, err)
		}

		manifests[filePath] = Manifest{
			Path:    filePath,
			Content: parsedManifest,
		}
	}

	return manifests, nil
}

// findManifestFiles recursively finds files with the given name using fs.WalkDir
func findManifestFiles(searchPath, fileName string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(searchPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && d.Name() == fileName {
			// Get relative path from search directory
			relPath, err := filepath.Rel(searchPath, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}

		return nil
	})

	return files, err
}

// ComponentsFetcher defines the interface for fetching components from different sources
type ComponentsFetcher interface {
	// Fetch retrieves all components from the given source
	Fetch(ctx context.Context, source SourceConfig) ([]models.Component, error)
}

// NewFetcher creates the appropriate fetcher based on source type
func NewFetcher(sourceType string) (ComponentsFetcher, error) {
	switch sourceType {
	case "git":
		return NewGitFetcher(), nil
	case "filesystem":
		return NewFilesystemFetcher(), nil
	default:
		return nil, fmt.Errorf("unsupported source type: %s", sourceType)
	}
}
