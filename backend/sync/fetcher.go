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
	// Find all manifest files
	manifestPaths, err := g.client.FindManifests(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("failed to find manifests: %w", err)
	}

	slog.Debug("Found manifest files", "count", len(manifestPaths), "source", source.URL)

	var components []models.Component
	for _, path := range manifestPaths {
		component, err := g.fetchComponentFromManifest(ctx, source, path)
		if err != nil {
			slog.Warn("Failed to process manifest", "path", path, "source", source.URL, "error", err)
			continue // Skip invalid manifests, don't fail entire sync
		}
		components = append(components, component)
	}

	return components, nil
}

// fetchComponentFromManifest processes a single manifest file and returns a Component
func (g *GitFetcher) fetchComponentFromManifest(ctx context.Context, source SourceConfig, path string) (models.Component, error) {
	content, err := g.client.GetFileContent(ctx, source, path)
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

// NewFetcher creates the appropriate fetcher based on source type
func NewFetcher(sourceType string) (ComponentsFetcher, error) {
	switch sourceType {
	case "git":
		return NewGitFetcher(), nil
	default:
		return nil, fmt.Errorf("unsupported source type: %s", sourceType)
	}
}
