package sync

import (
	"context"
	"fmt"

	"github.com/doron-cohen/argus/backend/internal/models"
)

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
