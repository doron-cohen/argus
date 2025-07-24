package sync

import (
	"context"
	"fmt"
	"path/filepath"
)

// FilesystemClient handles filesystem operations for syncing
type FilesystemClient struct {
	manifestClient *ManifestClient
}

// NewFilesystemClient creates a new filesystem client
func NewFilesystemClient() *FilesystemClient {
	return &FilesystemClient{
		manifestClient: NewManifestClient(),
	}
}

// FindManifests finds all manifest.yaml and manifest.yml files in the filesystem path
func (f *FilesystemClient) FindManifests(ctx context.Context, config FilesystemSourceConfig) ([]string, error) {
	// Resolve absolute path
	rootPath, err := filepath.Abs(config.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for %s: %w", config.Path, err)
	}

	// Validate that the root path exists
	if err := f.manifestClient.ValidateBasePath(rootPath, ""); err != nil {
		return nil, fmt.Errorf("filesystem path %s does not exist", config.Path)
	}

	// Use shared manifest discovery logic
	return f.manifestClient.FindManifests(rootPath, config.BasePath)
}

// GetFileContent reads the content of a file from the filesystem
func (f *FilesystemClient) GetFileContent(ctx context.Context, config FilesystemSourceConfig, filePath string) ([]byte, error) {
	// Resolve absolute path
	rootPath, err := filepath.Abs(config.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for %s: %w", config.Path, err)
	}

	// Use shared file reading logic
	return f.manifestClient.GetFileContent(rootPath, filePath)
}

// GetLastModified returns a simple timestamp for filesystem sources (not as sophisticated as git commits)
// This could be enhanced to track the most recent modification time of manifest files
func (f *FilesystemClient) GetLastModified(ctx context.Context, config FilesystemSourceConfig) (string, error) {
	// For now, we'll just return a simple indicator that this is a filesystem source
	// In the future, this could track the modification times of manifest files
	return fmt.Sprintf("filesystem:%s", config.Path), nil
}
