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

// GetBasePath returns the base path for this source (always empty for filesystem)
func (f *FilesystemSourceConfig) GetBasePath() string {
	return ""
}

// GetSourceType returns the source type
func (f *FilesystemSourceConfig) GetSourceType() string {
	return f.Type
}

// FilesystemFetcher implements ComponentsFetcher for local filesystem
type FilesystemFetcher struct{}

// NewFilesystemFetcher creates a new filesystem fetcher
func NewFilesystemFetcher() *FilesystemFetcher {
	return &FilesystemFetcher{}
}

// Fetch retrieves all components from a filesystem path
func (f *FilesystemFetcher) Fetch(ctx context.Context, source SourceConfig) ([]models.Component, error) {
	cfg := source.GetConfig()
	filesystemConfig, ok := cfg.(*FilesystemSourceConfig)
	if !ok {
		return nil, fmt.Errorf("source is not a filesystem config")
	}

	// Resolve absolute path
	rootPath, err := filepath.Abs(filesystemConfig.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for %s: %w", filesystemConfig.Path, err)
	}

	// Load all manifests directly
	manifests, err := LoadManifests(ctx, rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load manifests: %w", err)
	}

	slog.Debug("Found manifest files", "count", len(manifests), "source", filesystemConfig.Path)

	var components []models.Component
	for _, manifest := range manifests {
		component := manifest.Content.ToComponent()
		components = append(components, component)
	}

	return components, nil
}
