package sync

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/doron-cohen/argus/backend/internal/models"
	"github.com/doron-cohen/argus/backend/internal/storage"
)

// Error definitions
var (
	ErrSourceNotFound     = errors.New("source not found")
	ErrSyncAlreadyRunning = errors.New("sync already running for this source")
)

// SourceStatus represents the status of a sync source
type SourceStatus struct {
	Status          Status
	LastSync        *time.Time
	LastError       *string
	ComponentsCount int
	Duration        time.Duration
}

// Status represents the sync status
type Status string

const (
	StatusIdle      Status = "idle"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

// Service orchestrates the sync process
type Service struct {
	repo     Repository // Use interface instead of concrete type
	config   Config
	fetchers map[string]ComponentsFetcher // Cache fetchers by type

	// Status tracking
	statusMutex sync.RWMutex
	statuses    map[int]*SourceStatus
	running     map[int]bool
}

// NewService creates a new sync service
func NewService(repo Repository, config Config) *Service {
	return &Service{
		repo:     repo,
		config:   config,
		fetchers: make(map[string]ComponentsFetcher),
		statuses: make(map[int]*SourceStatus),
		running:  make(map[int]bool),
	}
}

// API Methods

// GetSources returns all configured sources
func (s *Service) GetSources() []SourceConfig {
	return s.config.Sources
}

// GetSourceByIndex returns a source by its index
func (s *Service) GetSourceByIndex(index int) (SourceConfig, error) {
	if index < 0 || index >= len(s.config.Sources) {
		return SourceConfig{}, ErrSourceNotFound
	}
	return s.config.Sources[index], nil
}

// GetSourceStatus returns the status of a source by index
func (s *Service) GetSourceStatus(index int) (*SourceStatus, error) {
	if index < 0 || index >= len(s.config.Sources) {
		return nil, ErrSourceNotFound
	}

	s.statusMutex.RLock()
	defer s.statusMutex.RUnlock()

	status, exists := s.statuses[index]
	if !exists {
		// Return default status for sources that haven't been synced yet
		return &SourceStatus{
			Status: StatusIdle,
		}, nil
	}

	return status, nil
}

// TriggerSync triggers a manual sync for a source
func (s *Service) TriggerSync(index int) error {
	if index < 0 || index >= len(s.config.Sources) {
		return ErrSourceNotFound
	}

	s.statusMutex.Lock()
	defer s.statusMutex.Unlock()

	// Check if sync is already running
	if s.running[index] {
		return ErrSyncAlreadyRunning
	}

	// Mark as running
	s.running[index] = true

	// Start sync in background
	go func() {
		defer func() {
			s.statusMutex.Lock()
			s.running[index] = false
			s.statusMutex.Unlock()
		}()

		source := s.config.Sources[index]
		ctx := context.Background()

		// Update status to running
		s.updateStatus(index, &SourceStatus{
			Status: StatusRunning,
		})

		startTime := time.Now()

		// Perform sync
		err := s.SyncSource(ctx, source)

		duration := time.Since(startTime)

		// Update status based on result
		status := &SourceStatus{
			Duration: duration,
		}

		if err != nil {
			status.Status = StatusFailed
			errorMsg := err.Error()
			status.LastError = &errorMsg
		} else {
			status.Status = StatusCompleted
		}

		now := time.Now()
		status.LastSync = &now

		s.updateStatus(index, status)
	}()

	return nil
}

// updateStatus updates the status for a source (thread-safe)
func (s *Service) updateStatus(index int, status *SourceStatus) {
	s.statusMutex.Lock()
	defer s.statusMutex.Unlock()
	s.statuses[index] = status
}

// StartPeriodicSync starts the sync process if sources are configured
func (s *Service) StartPeriodicSync(ctx context.Context) {
	if len(s.config.Sources) == 0 {
		slog.Warn("No sync sources configured, skipping sync service startup")
		return
	}

	slog.Info("Starting sync service", "sources", len(s.config.Sources))

	for i, source := range s.config.Sources {
		// Initialize status for this source
		s.updateStatus(i, &SourceStatus{
			Status: StatusIdle,
		})

		go s.startSourceSync(ctx, source, i)
	}
}

// startSourceSync starts periodic sync for a single source
func (s *Service) startSourceSync(ctx context.Context, source SourceConfig, index int) {
	interval := time.Duration(0)
	if cfg := source.GetConfig(); cfg != nil {
		interval = cfg.GetInterval()
	}
	if interval == 0 {
		interval = 5 * time.Minute // fallback default
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	sourceInfo := s.getSourceInfo(source)
	slog.Info("Starting periodic sync for source", "source", sourceInfo, "interval", interval)

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
	cfg := source.GetConfig()
	sourceType := "unknown"
	if cfg != nil {
		sourceType = cfg.GetSourceType()
	}
	slog.Info("Starting sync", "source", sourceInfo, "type", sourceType)

	// Get or create fetcher for this source type
	fetcher, err := s.getFetcher(sourceType)
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
	cfg := source.GetConfig()
	switch c := cfg.(type) {
	case *GitSourceConfig:
		return c.URL
	case *FilesystemSourceConfig:
		return c.Path
	default:
		return "unknown"
	}
}
