package sync

import (
	"context"
	"errors"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

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

func (m *MockRepository) GetComponentByID(ctx context.Context, componentID string) (*storage.Component, error) {
	args := m.Called(ctx, componentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.Component), args.Error(1)
}

func (m *MockRepository) CreateComponent(ctx context.Context, component storage.Component) error {
	args := m.Called(ctx, component)
	return args.Error(0)
}

func newSourceConfigFromYAMLOrPanic(yamlSource string) SourceConfig {
	var source SourceConfig
	err := yaml.Unmarshal([]byte(yamlSource), &source)
	if err != nil {
		panic(err)
	}
	return source
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

	source := newSourceConfigFromYAMLOrPanic("type: git\nurl: https://github.com/test/repo")

	expectedComponents := []models.Component{
		{Name: "service-a"},
		{Name: "service-b"},
	}

	ctx := context.Background()

	// Mock expectations
	mockFetcher.On("Fetch", ctx, source).Return(expectedComponents, nil)

	// Both components are new (not found)
	mockRepo.On("GetComponentByID", ctx, "service-a").Return(nil, storage.ErrComponentNotFound)
	mockRepo.On("GetComponentByID", ctx, "service-b").Return(nil, storage.ErrComponentNotFound)

	// Both components are created successfully
	mockRepo.On("CreateComponent", ctx, storage.Component{ComponentID: "service-a", Name: "service-a"}).Return(nil)
	mockRepo.On("CreateComponent", ctx, storage.Component{ComponentID: "service-b", Name: "service-b"}).Return(nil)

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

	source := newSourceConfigFromYAMLOrPanic("type: git\nurl: https://github.com/test/repo")

	expectedComponents := []models.Component{
		{Name: "existing-service"},
		{Name: "new-service"},
	}

	ctx := context.Background()
	existingComponent := &storage.Component{ComponentID: "existing-service", Name: "existing-service"}

	// Mock expectations
	mockFetcher.On("Fetch", ctx, source).Return(expectedComponents, nil)

	// First component exists, second is new
	mockRepo.On("GetComponentByID", ctx, "existing-service").Return(existingComponent, nil)
	mockRepo.On("GetComponentByID", ctx, "new-service").Return(nil, storage.ErrComponentNotFound)

	// Only new component is created
	mockRepo.On("CreateComponent", ctx, storage.Component{ComponentID: "new-service", Name: "new-service"}).Return(nil)

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

	source := newSourceConfigFromYAMLOrPanic("type: git\nurl: https://github.com/test/repo")

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

	source := newSourceConfigFromYAMLOrPanic("type: git\nurl: https://github.com/test/repo")

	expectedComponents := []models.Component{
		{Name: "failing-service"},
		{Name: "working-service"},
	}

	ctx := context.Background()
	createError := errors.New("database connection failed")

	// Mock expectations
	mockFetcher.On("Fetch", ctx, source).Return(expectedComponents, nil)

	// Both components are new
	mockRepo.On("GetComponentByID", ctx, "failing-service").Return(nil, storage.ErrComponentNotFound)
	mockRepo.On("GetComponentByID", ctx, "working-service").Return(nil, storage.ErrComponentNotFound)

	// First component creation fails, second succeeds
	mockRepo.On("CreateComponent", ctx, storage.Component{ComponentID: "failing-service", Name: "failing-service"}).Return(createError)
	mockRepo.On("CreateComponent", ctx, storage.Component{ComponentID: "working-service", Name: "working-service"}).Return(nil)

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

	// Test that the service handles unsupported source types gracefully
	// Create a source config with an unsupported type
	// We need to create a SourceConfig manually since YAML unmarshaling would fail
	source := SourceConfig{
		config: &MockSourceConfig{sourceType: "svn"},
	}

	ctx := context.Background()

	// Execute
	err := service.SyncSource(ctx, source)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get fetcher")
	assert.Contains(t, err.Error(), "unsupported source type: svn")
}

// MockSourceConfig implements SourceTypeConfig for testing unsupported types
type MockSourceConfig struct {
	sourceType string
}

func (m *MockSourceConfig) Validate() error {
	return nil
}

func (m *MockSourceConfig) GetInterval() time.Duration {
	return 5 * time.Minute
}

func (m *MockSourceConfig) GetBasePath() string {
	return ""
}

func (m *MockSourceConfig) GetSourceType() string {
	return m.sourceType
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
	source := newSourceConfigFromYAMLOrPanic("type: git\nurl: https://github.com/test/repo")
	ctx := context.Background()

	dbError := errors.New("database connection lost")

	// Mock expectations - database check fails with non-NotFound error
	mockRepo.On("GetComponentByID", ctx, "test-service").Return(nil, dbError)

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
	source := newSourceConfigFromYAMLOrPanic("type: git\nurl: https://github.com/test/repo")
	config := Config{
		Sources: []SourceConfig{source},
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

func TestService_EmptySources(t *testing.T) {
	mockRepo := &MockRepository{}
	config := Config{Sources: []SourceConfig{}}
	service := NewService(mockRepo, config)
	assert.NotNil(t, service)
	assert.Empty(t, service.config.Sources)
}
