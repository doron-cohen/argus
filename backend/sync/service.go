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

	slog.Info("Starting periodic sync for source", "url", source.URL, "interval", source.Interval)

	// Initial sync
	if err := s.SyncSource(ctx, source); err != nil {
		slog.Error("Initial sync failed", "source", source.URL, "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping sync for source", "url", source.URL)
			return
		case <-ticker.C:
			if err := s.SyncSource(ctx, source); err != nil {
				slog.Error("Sync failed", "source", source.URL, "error", err)
			}
		}
	}
}

// SyncSource performs a full sync for a single source
func (s *Service) SyncSource(ctx context.Context, source SourceConfig) error {
	slog.Info("Starting sync", "source", source.URL, "type", source.Type)

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

	slog.Info("Fetched components", "count", len(components), "source", source.URL)

	// Process each component
	created := 0
	for _, component := range components {
		if err := s.processComponent(ctx, component, source); err != nil {
			slog.Error("Failed to process component",
				"name", component.Name,
				"source", source.URL,
				"error", err)
			continue
		}
		created++
	}

	slog.Info("Sync completed",
		"source", source.URL,
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
