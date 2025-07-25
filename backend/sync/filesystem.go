package sync

import (
	"context"
	"fmt"
	"path/filepath"
	"time"
)

// FilesystemSourceConfig holds filesystem-specific configuration
type FilesystemSourceConfig struct {
	Type     string        `fig:"type" yaml:"type"`
	Interval time.Duration `fig:"interval" yaml:"interval"`
	Path     string        `fig:"path" yaml:"path"`
	BasePath string        `fig:"base_path" yaml:"base_path,omitempty"`
}

// Validate ensures the filesystem configuration is valid
func (f *FilesystemSourceConfig) Validate() error {
	if f.Type != "filesystem" {
		return fmt.Errorf("expected type 'filesystem', got '%s'", f.Type)
	}
	if f.Path == "" {
		return fmt.Errorf("filesystem source requires path field")
	}

	interval := f.GetInterval()
	if interval < MinFilesystemInterval {
		return fmt.Errorf("filesystem source interval must be at least %v, got %v", MinFilesystemInterval, interval)
	}

	// Set default values if not provided
	if f.Type == "" {
		f.Type = "filesystem"
	}

	return nil
}

// GetInterval returns the sync interval for this source
func (f *FilesystemSourceConfig) GetInterval() time.Duration {
	if f.Interval == 0 {
		return 5 * time.Minute // default
	}
	return f.Interval
}

// GetBasePath returns the base path for this source
func (f *FilesystemSourceConfig) GetBasePath() string {
	return f.BasePath
}

// GetSourceType returns the source type
func (f *FilesystemSourceConfig) GetSourceType() string {
	return f.Type
}

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
