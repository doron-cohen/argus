package sync

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/doron-cohen/argus/backend/internal/models"
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

// FilesystemFetcher implements ComponentsFetcher for local filesystem
type FilesystemFetcher struct {
	client *FilesystemClient
	parser *models.Parser
}

// NewFilesystemFetcher creates a new filesystem fetcher
func NewFilesystemFetcher() *FilesystemFetcher {
	return &FilesystemFetcher{
		client: NewFilesystemClient(),
		parser: models.NewParser(),
	}
}

// Fetch retrieves all components from a filesystem path
func (f *FilesystemFetcher) Fetch(ctx context.Context, source SourceConfig) ([]models.Component, error) {
	cfg := source.GetConfig()
	filesystemConfig, ok := cfg.(*FilesystemSourceConfig)
	if !ok {
		return nil, fmt.Errorf("source is not a filesystem config")
	}

	// Find all manifest files
	manifestPaths, err := f.client.FindManifests(ctx, *filesystemConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to find manifests: %w", err)
	}

	slog.Debug("Found manifest files", "count", len(manifestPaths), "source", filesystemConfig.Path)

	var components []models.Component
	for _, path := range manifestPaths {
		component, err := f.fetchComponentFromManifest(ctx, *filesystemConfig, path)
		if err != nil {
			slog.Warn("Failed to process manifest", "path", path, "source", filesystemConfig.Path, "error", err)
			continue // Skip invalid manifests, don't fail entire sync
		}
		components = append(components, component)
	}

	return components, nil
}

// fetchComponentFromManifest processes a single manifest file and returns a Component
func (f *FilesystemFetcher) fetchComponentFromManifest(ctx context.Context, filesystemConfig FilesystemSourceConfig, path string) (models.Component, error) {
	content, err := f.client.GetFileContent(ctx, filesystemConfig, path)
	if err != nil {
		return models.Component{}, fmt.Errorf("failed to get file content: %w", err)
	}

	manifest, err := f.parser.Parse(content)
	if err != nil {
		return models.Component{}, fmt.Errorf("failed to parse manifest: %w", err)
	}

	if err := f.parser.Validate(manifest); err != nil {
		return models.Component{}, fmt.Errorf("invalid manifest: %w", err)
	}

	return manifest.ToComponent(), nil
}
