package sync

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/doron-cohen/argus/backend/internal/models"
)

// ComponentsFetcher defines the interface for fetching components from different sources
type ComponentsFetcher interface {
	// Fetch retrieves all components from the given source
	Fetch(ctx context.Context, source SourceConfig) ([]models.Component, error)
}

// GitFetcher implements ComponentsFetcher for git repositories
type GitFetcher struct {
	client *GitClient
	parser *models.Parser
}

// NewGitFetcher creates a new git fetcher
func NewGitFetcher() *GitFetcher {
	return &GitFetcher{
		client: NewGitClient(),
		parser: models.NewParser(),
	}
}

// Fetch retrieves all components from a git repository
func (g *GitFetcher) Fetch(ctx context.Context, source SourceConfig) ([]models.Component, error) {
	cfg := source.GetConfig()
	gitConfig, ok := cfg.(*GitSourceConfig)
	if !ok {
		return nil, fmt.Errorf("source is not a git config")
	}

	// Find all manifest files
	manifestPaths, err := g.client.FindManifests(ctx, *gitConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to find manifests: %w", err)
	}

	slog.Debug("Found manifest files", "count", len(manifestPaths), "source", gitConfig.URL)

	var components []models.Component
	for _, path := range manifestPaths {
		component, err := g.fetchComponentFromManifest(ctx, *gitConfig, path)
		if err != nil {
			slog.Warn("Failed to process manifest", "path", path, "source", gitConfig.URL, "error", err)
			continue // Skip invalid manifests, don't fail entire sync
		}
		components = append(components, component)
	}

	return components, nil
}

// fetchComponentFromManifest processes a single manifest file and returns a Component
func (g *GitFetcher) fetchComponentFromManifest(ctx context.Context, gitConfig GitSourceConfig, path string) (models.Component, error) {
	content, err := g.client.GetFileContent(ctx, gitConfig, path)
	if err != nil {
		return models.Component{}, fmt.Errorf("failed to get file content: %w", err)
	}

	manifest, err := g.parser.Parse(content)
	if err != nil {
		return models.Component{}, fmt.Errorf("failed to parse manifest: %w", err)
	}

	if err := g.parser.Validate(manifest); err != nil {
		return models.Component{}, fmt.Errorf("invalid manifest: %w", err)
	}

	return models.Component{
		Name: manifest.Name,
	}, nil
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

	return models.Component{
		Name: manifest.Name,
	}, nil
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
