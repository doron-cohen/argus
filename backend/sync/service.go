package sync

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/doron-cohen/argus/backend/internal/models"
	"github.com/doron-cohen/argus/backend/internal/storage"
)

// Service orchestrates the sync process
type Service struct {
	repo     Repository // Use interface instead of concrete type
	config   Config
	fetchers map[string]ComponentsFetcher // Cache fetchers by type
}

// NewService creates a new sync service
func NewService(repo Repository, config Config) *Service {
	return &Service{
		repo:     repo,
		config:   config,
		fetchers: make(map[string]ComponentsFetcher),
	}
}

// StartPeriodicSync starts the sync process if sources are configured
func (s *Service) StartPeriodicSync(ctx context.Context) {
	if len(s.config.Sources) == 0 {
		slog.Warn("No sync sources configured, skipping sync service startup")
		return
	}

	slog.Info("Starting sync service", "sources", len(s.config.Sources))

	for _, source := range s.config.Sources {
		go s.startSourceSync(ctx, source)
	}
}

// startSourceSync starts periodic sync for a single source
func (s *Service) startSourceSync(ctx context.Context, source SourceConfig) {
	ticker := time.NewTicker(source.Interval)
	defer ticker.Stop()

	sourceInfo := s.getSourceInfo(source)
	slog.Info("Starting periodic sync for source", "source", sourceInfo, "interval", source.Interval)

	// Initial sync
	if err := s.SyncSource(ctx, source); err != nil {
		slog.Error("Initial sync failed", "source", sourceInfo, "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping sync for source", "source", sourceInfo)
			return
		case <-ticker.C:
			if err := s.SyncSource(ctx, source); err != nil {
				slog.Error("Sync failed", "source", sourceInfo, "error", err)
			}
		}
	}
}

// SyncSource performs a full sync for a single source
func (s *Service) SyncSource(ctx context.Context, source SourceConfig) error {
	sourceInfo := s.getSourceInfo(source)
	slog.Info("Starting sync", "source", sourceInfo, "type", source.Type)

	// Get or create fetcher for this source type
	fetcher, err := s.getFetcher(source.Type)
	if err != nil {
		return fmt.Errorf("failed to get fetcher: %w", err)
	}

	// Fetch all components from the source
	components, err := fetcher.Fetch(ctx, source)
	if err != nil {
		return fmt.Errorf("failed to fetch components: %w", err)
	}

	slog.Info("Fetched components", "count", len(components), "source", sourceInfo)

	// Process each component
	created := 0
	for _, component := range components {
		if err := s.processComponent(ctx, component, source); err != nil {
			slog.Error("Failed to process component",
				"name", component.Name,
				"source", sourceInfo,
				"error", err)
			continue
		}
		created++
	}

	slog.Info("Sync completed",
		"source", sourceInfo,
		"total", len(components),
		"created", created)

	return nil
}

// processComponent handles a single component (create only for now)
func (s *Service) processComponent(ctx context.Context, component models.Component, source SourceConfig) error {
	// Check if component already exists
	existing, err := s.repo.GetComponentByName(ctx, component.Name)
	if err != nil && err != storage.ErrComponentNotFound {
		return fmt.Errorf("failed to check existing component: %w", err)
	}

	if existing != nil {
		// Component exists, skip for now (no updates)
		slog.Debug("Component already exists, skipping", "name", component.Name)
		return nil
	}

	// Create new component
	storageComponent := storage.Component{
		Name: component.Name,
	}

	if err := s.repo.CreateComponent(ctx, storageComponent); err != nil {
		return fmt.Errorf("failed to create component: %w", err)
	}

	slog.Info("Created new component", "name", component.Name)
	return nil
}

// getFetcher returns a cached fetcher for the given type
func (s *Service) getFetcher(sourceType string) (ComponentsFetcher, error) {
	if fetcher, exists := s.fetchers[sourceType]; exists {
		return fetcher, nil
	}

	fetcher, err := NewFetcher(sourceType)
	if err != nil {
		return nil, err
	}

	s.fetchers[sourceType] = fetcher
	return fetcher, nil
}

// getSourceInfo returns a string representation of the source for logging
func (s *Service) getSourceInfo(source SourceConfig) string {
	switch source.Type {
	case "git":
		if gitConfig, err := source.GitConfig(); err == nil {
			return gitConfig.URL
		}
		return source.URL // fallback to raw field
	case "filesystem":
		if fsConfig, err := source.FilesystemConfig(); err == nil {
			return fsConfig.Path
		}
		return source.Path // fallback to raw field
	default:
		return fmt.Sprintf("%s:%s", source.Type, source.URL)
	}
}
