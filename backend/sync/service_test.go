package sync

import (
	"context"
	"errors"
	"testing"

	"github.com/doron-cohen/argus/backend/internal/models"
	"github.com/doron-cohen/argus/backend/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockFetcher implements ComponentsFetcher for testing
type MockFetcher struct {
	mock.Mock
}

func (m *MockFetcher) Fetch(ctx context.Context, source SourceConfig) ([]models.Component, error) {
	args := m.Called(ctx, source)
	return args.Get(0).([]models.Component), args.Error(1)
}

// MockRepository implements Repository interface for testing
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetComponentByName(ctx context.Context, name string) (*storage.Component, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.Component), args.Error(1)
}

func (m *MockRepository) CreateComponent(ctx context.Context, component storage.Component) error {
	args := m.Called(ctx, component)
	return args.Error(0)
}

func TestService_SyncSource_Success(t *testing.T) {
	// Setup
	mockRepo := &MockRepository{}
	mockFetcher := &MockFetcher{}

	service := &Service{
		repo:     mockRepo,
		config:   Config{},
		fetchers: map[string]ComponentsFetcher{"git": mockFetcher},
	}

	source := SourceConfig{
		Type: "git",
		URL:  "https://github.com/test/repo",
	}

	expectedComponents := []models.Component{
		{Name: "service-a"},
		{Name: "service-b"},
	}

	ctx := context.Background()

	// Mock expectations
	mockFetcher.On("Fetch", ctx, source).Return(expectedComponents, nil)

	// Both components are new (not found)
	mockRepo.On("GetComponentByName", ctx, "service-a").Return(nil, storage.ErrComponentNotFound)
	mockRepo.On("GetComponentByName", ctx, "service-b").Return(nil, storage.ErrComponentNotFound)

	// Both components are created successfully
	mockRepo.On("CreateComponent", ctx, storage.Component{Name: "service-a"}).Return(nil)
	mockRepo.On("CreateComponent", ctx, storage.Component{Name: "service-b"}).Return(nil)

	// Execute
	err := service.SyncSource(ctx, source)

	// Assert
	require.NoError(t, err)
	mockFetcher.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestService_SyncSource_SkipExistingComponents(t *testing.T) {
	// Setup
	mockRepo := &MockRepository{}
	mockFetcher := &MockFetcher{}

	service := &Service{
		repo:     mockRepo,
		config:   Config{},
		fetchers: map[string]ComponentsFetcher{"git": mockFetcher},
	}

	source := SourceConfig{
		Type: "git",
		URL:  "https://github.com/test/repo",
	}

	expectedComponents := []models.Component{
		{Name: "existing-service"},
		{Name: "new-service"},
	}

	ctx := context.Background()
	existingComponent := &storage.Component{Name: "existing-service"}

	// Mock expectations
	mockFetcher.On("Fetch", ctx, source).Return(expectedComponents, nil)

	// First component exists, second is new
	mockRepo.On("GetComponentByName", ctx, "existing-service").Return(existingComponent, nil)
	mockRepo.On("GetComponentByName", ctx, "new-service").Return(nil, storage.ErrComponentNotFound)

	// Only new component is created
	mockRepo.On("CreateComponent", ctx, storage.Component{Name: "new-service"}).Return(nil)

	// Execute
	err := service.SyncSource(ctx, source)

	// Assert
	require.NoError(t, err)
	mockFetcher.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestService_SyncSource_FetchError(t *testing.T) {
	// Setup
	mockRepo := &MockRepository{}
	mockFetcher := &MockFetcher{}

	service := &Service{
		repo:     mockRepo,
		config:   Config{},
		fetchers: map[string]ComponentsFetcher{"git": mockFetcher},
	}

	source := SourceConfig{
		Type: "git",
		URL:  "https://github.com/test/repo",
	}

	ctx := context.Background()
	fetchError := errors.New("failed to clone repository")

	// Mock expectations
	mockFetcher.On("Fetch", ctx, source).Return([]models.Component{}, fetchError)

	// Execute
	err := service.SyncSource(ctx, source)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch components")
	assert.Contains(t, err.Error(), "failed to clone repository")
	mockFetcher.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestService_SyncSource_CreateComponentError(t *testing.T) {
	// Setup
	mockRepo := &MockRepository{}
	mockFetcher := &MockFetcher{}

	service := &Service{
		repo:     mockRepo,
		config:   Config{},
		fetchers: map[string]ComponentsFetcher{"git": mockFetcher},
	}

	source := SourceConfig{
		Type: "git",
		URL:  "https://github.com/test/repo",
	}

	expectedComponents := []models.Component{
		{Name: "failing-service"},
		{Name: "working-service"},
	}

	ctx := context.Background()
	createError := errors.New("database connection failed")

	// Mock expectations
	mockFetcher.On("Fetch", ctx, source).Return(expectedComponents, nil)

	// Both components are new
	mockRepo.On("GetComponentByName", ctx, "failing-service").Return(nil, storage.ErrComponentNotFound)
	mockRepo.On("GetComponentByName", ctx, "working-service").Return(nil, storage.ErrComponentNotFound)

	// First component creation fails, second succeeds
	mockRepo.On("CreateComponent", ctx, storage.Component{Name: "failing-service"}).Return(createError)
	mockRepo.On("CreateComponent", ctx, storage.Component{Name: "working-service"}).Return(nil)

	// Execute
	err := service.SyncSource(ctx, source)

	// Assert - sync should complete even with individual component failures
	require.NoError(t, err)
	mockFetcher.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestService_SyncSource_UnsupportedSourceType(t *testing.T) {
	// Setup
	mockRepo := &MockRepository{}

	service := &Service{
		repo:     mockRepo,
		config:   Config{},
		fetchers: make(map[string]ComponentsFetcher),
	}

	source := SourceConfig{
		Type: "svn", // Unsupported type
		URL:  "https://svn.example.com/repo",
	}

	ctx := context.Background()

	// Execute
	err := service.SyncSource(ctx, source)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get fetcher")
	assert.Contains(t, err.Error(), "unsupported source type: svn")
}

func TestService_StartPeriodicSync_NoSources(t *testing.T) {
	// Setup
	mockRepo := &MockRepository{}

	service := &Service{
		repo:     mockRepo,
		config:   Config{Sources: []SourceConfig{}}, // Empty sources
		fetchers: make(map[string]ComponentsFetcher),
	}

	ctx := context.Background()

	// Execute - should return immediately without error
	service.StartPeriodicSync(ctx)

	// Assert - no expectations to verify since no operations should occur
	mockRepo.AssertExpectations(t)
}

func TestService_processComponent_DatabaseCheckError(t *testing.T) {
	// Setup
	mockRepo := &MockRepository{}

	service := &Service{
		repo:     mockRepo,
		config:   Config{},
		fetchers: make(map[string]ComponentsFetcher),
	}

	component := models.Component{Name: "test-service"}
	source := SourceConfig{URL: "https://github.com/test/repo"}
	ctx := context.Background()

	dbError := errors.New("database connection lost")

	// Mock expectations - database check fails with non-NotFound error
	mockRepo.On("GetComponentByName", ctx, "test-service").Return(nil, dbError)

	// Execute
	err := service.processComponent(ctx, component, source)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check existing component")
	assert.Contains(t, err.Error(), "database connection lost")
	mockRepo.AssertExpectations(t)
}

func TestService_getFetcher_Caching(t *testing.T) {
	// Setup
	service := &Service{
		repo:     &MockRepository{},
		config:   Config{},
		fetchers: make(map[string]ComponentsFetcher),
	}

	// Execute - get fetcher twice
	fetcher1, err1 := service.getFetcher("git")
	fetcher2, err2 := service.getFetcher("git")

	// Assert
	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.Same(t, fetcher1, fetcher2) // Should be same instance (cached)
}

func TestNewService(t *testing.T) {
	// Setup
	mockRepo := &MockRepository{}
	config := Config{
		Sources: []SourceConfig{
			{Type: "git", URL: "https://github.com/test/repo"},
		},
	}

	// Execute
	service := NewService(mockRepo, config)

	// Assert
	assert.NotNil(t, service)
	assert.Same(t, mockRepo, service.repo)
	assert.Equal(t, config, service.config)
	assert.NotNil(t, service.fetchers)
	assert.Empty(t, service.fetchers) // Should start empty
}
