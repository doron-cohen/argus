package sync

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/doron-cohen/argus/backend/internal/models"
	"github.com/doron-cohen/argus/backend/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// getTestRepositoryURL returns the repository URL for integration testing
// Can be overridden with ARGUS_TEST_REPO_URL environment variable
func getTestRepositoryURL() string {
	if url := os.Getenv("ARGUS_TEST_REPO_URL"); url != "" {
		return url
	}
	// Default to the main argus repository - update this to your fork!
	return "https://github.com/doron-cohen/argus"
}

// skipIfRepositoryNotAccessible checks if the repository is accessible and skips the test if not
func skipIfRepositoryNotAccessible(t *testing.T) {
	ctx := context.Background()
	client := NewGitClient()

	gitConfig := GitSourceConfig{
		URL:    getTestRepositoryURL(),
		Branch: "main",
	}

	// Try to get the latest commit to check if repository is accessible
	_, err := client.GetLatestCommit(ctx, gitConfig)
	if err != nil {
		t.Skipf("Repository %s not accessible (you may need to set ARGUS_TEST_REPO_URL env var to your fork): %v",
			getTestRepositoryURL(), err)
	}
}

// IntegrationMockRepository is a mock repository for integration testing
type IntegrationMockRepository struct {
	mock.Mock
	createdComponents []storage.Component
}

func (m *IntegrationMockRepository) GetComponentByName(ctx context.Context, name string) (*storage.Component, error) {
	args := m.Called(ctx, name)

	// Check if component was already created in this test
	for _, comp := range m.createdComponents {
		if comp.Name == name {
			return &comp, nil
		}
	}

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.Component), args.Error(1)
}

func (m *IntegrationMockRepository) CreateComponent(ctx context.Context, component storage.Component) error {
	args := m.Called(ctx, component)
	if args.Error(0) == nil {
		// Track created components
		m.createdComponents = append(m.createdComponents, component)
	}
	return args.Error(0)
}

func TestExample_SyncFromRealRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	skipIfRepositoryNotAccessible(t)

	ctx := context.Background()

	// Mock repository to avoid needing real database
	mockRepo := &IntegrationMockRepository{}

	// Set up mock expectations - all components should be new
	mockRepo.On("GetComponentByName", mock.Anything, "auth-service").Return(nil, storage.ErrComponentNotFound)
	mockRepo.On("GetComponentByName", mock.Anything, "api-gateway").Return(nil, storage.ErrComponentNotFound)
	mockRepo.On("GetComponentByName", mock.Anything, "user-service").Return(nil, storage.ErrComponentNotFound)
	mockRepo.On("GetComponentByName", mock.Anything, "platform-infrastructure").Return(nil, storage.ErrComponentNotFound)

	// Expect all components to be created
	mockRepo.On("CreateComponent", mock.Anything, mock.MatchedBy(func(comp storage.Component) bool {
		return comp.Name == "auth-service"
	})).Return(nil)
	mockRepo.On("CreateComponent", mock.Anything, mock.MatchedBy(func(comp storage.Component) bool {
		return comp.Name == "api-gateway"
	})).Return(nil)
	mockRepo.On("CreateComponent", mock.Anything, mock.MatchedBy(func(comp storage.Component) bool {
		return comp.Name == "user-service"
	})).Return(nil)
	mockRepo.On("CreateComponent", mock.Anything, mock.MatchedBy(func(comp storage.Component) bool {
		return comp.Name == "platform-infrastructure"
	})).Return(nil)

	// Test configuration - sync from the examples directory
	config := Config{
		Sources: []SourceConfig{
			{
				Type:     "git",
				URL:      getTestRepositoryURL(),
				Branch:   "main",
				BasePath: "backend/tests/testdata",
				Interval: time.Minute, // Not used in single sync test
			},
		},
	}

	// Create service and run single sync
	service := NewService(mockRepo, config)

	// Sync the first (and only) source
	err := service.SyncSource(ctx, config.Sources[0])
	require.NoError(t, err)

	// Verify all expected components were created
	mockRepo.AssertExpectations(t)

	// Verify the created components
	assert.Len(t, mockRepo.createdComponents, 4, "Should have created 4 components")

	componentNames := make([]string, len(mockRepo.createdComponents))
	for i, comp := range mockRepo.createdComponents {
		componentNames[i] = comp.Name
	}

	assert.Contains(t, componentNames, "auth-service")
	assert.Contains(t, componentNames, "api-gateway")
	assert.Contains(t, componentNames, "user-service")
	assert.Contains(t, componentNames, "platform-infrastructure")
}

func TestExample_SyncWithBasePath_ServicesOnly(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	skipIfRepositoryNotAccessible(t)

	ctx := context.Background()

	// Mock repository
	mockRepo := &IntegrationMockRepository{}

	// Only services should be found (auth, api, user) - not platform-infrastructure
	mockRepo.On("GetComponentByName", mock.Anything, "auth-service").Return(nil, storage.ErrComponentNotFound)
	mockRepo.On("GetComponentByName", mock.Anything, "api-gateway").Return(nil, storage.ErrComponentNotFound)
	mockRepo.On("GetComponentByName", mock.Anything, "user-service").Return(nil, storage.ErrComponentNotFound)

	mockRepo.On("CreateComponent", mock.Anything, mock.MatchedBy(func(comp storage.Component) bool {
		return comp.Name == "auth-service"
	})).Return(nil)
	mockRepo.On("CreateComponent", mock.Anything, mock.MatchedBy(func(comp storage.Component) bool {
		return comp.Name == "api-gateway"
	})).Return(nil)
	mockRepo.On("CreateComponent", mock.Anything, mock.MatchedBy(func(comp storage.Component) bool {
		return comp.Name == "user-service"
	})).Return(nil)

	// Test configuration - sync only from backend/tests/testdata/services
	config := Config{
		Sources: []SourceConfig{
			{
				Type:     "git",
				URL:      getTestRepositoryURL(),
				Branch:   "main",
				BasePath: "backend/tests/testdata/services", // Only services subdirectory
				Interval: time.Minute,
			},
		},
	}

	service := NewService(mockRepo, config)
	err := service.SyncSource(ctx, config.Sources[0])
	require.NoError(t, err)

	// Verify only 3 service components were created (not platform-infrastructure)
	mockRepo.AssertExpectations(t)
	assert.Len(t, mockRepo.createdComponents, 3, "Should have created 3 service components only")

	componentNames := make([]string, len(mockRepo.createdComponents))
	for i, comp := range mockRepo.createdComponents {
		componentNames[i] = comp.Name
	}

	assert.Contains(t, componentNames, "auth-service")
	assert.Contains(t, componentNames, "api-gateway")
	assert.Contains(t, componentNames, "user-service")
	assert.NotContains(t, componentNames, "platform-infrastructure", "Platform component should not be found with services-only BasePath")
}

func TestExample_SyncWithBasePath_PlatformOnly(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	skipIfRepositoryNotAccessible(t)

	ctx := context.Background()

	// Mock repository
	mockRepo := &IntegrationMockRepository{}

	// Only platform-infrastructure should be found
	mockRepo.On("GetComponentByName", mock.Anything, "platform-infrastructure").Return(nil, storage.ErrComponentNotFound)
	mockRepo.On("CreateComponent", mock.Anything, mock.MatchedBy(func(comp storage.Component) bool {
		return comp.Name == "platform-infrastructure"
	})).Return(nil)

	// Test configuration - sync only from examples/platform
	config := Config{
		Sources: []SourceConfig{
			{
				Type:     "git",
				URL:      getTestRepositoryURL(),
				Branch:   "main",
				BasePath: "backend/tests/testdata/platform", // Only platform subdirectory
				Interval: time.Minute,
			},
		},
	}

	service := NewService(mockRepo, config)
	err := service.SyncSource(ctx, config.Sources[0])
	require.NoError(t, err)

	// Verify only 1 platform component was created
	mockRepo.AssertExpectations(t)
	assert.Len(t, mockRepo.createdComponents, 1, "Should have created 1 platform component only")
	assert.Equal(t, "platform-infrastructure", mockRepo.createdComponents[0].Name)
}

func TestExample_GitClient_RealRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	skipIfRepositoryNotAccessible(t)

	ctx := context.Background()
	client := NewGitClient()

	gitConfig := GitSourceConfig{
		URL:      getTestRepositoryURL(),
		Branch:   "main",
		BasePath: "backend/tests/testdata",
	}

	t.Run("find manifests in examples directory", func(t *testing.T) {
		manifests, err := client.FindManifests(ctx, gitConfig)
		require.NoError(t, err)

		// Should find all 4 manifest files
		assert.Len(t, manifests, 4, "Should find 4 manifest files in examples directory")

		// Check that all expected manifests are found
		expectedManifests := []string{
			"backend/tests/testdata/services/auth/manifest.yaml",
			"backend/tests/testdata/services/api/manifest.yaml",
			"backend/tests/testdata/services/user/manifest.yml",
			"backend/tests/testdata/platform/infrastructure/manifest.yml",
		}

		for _, expected := range expectedManifests {
			assert.Contains(t, manifests, expected, "Should find manifest: %s", expected)
		}
	})

	t.Run("read manifest content", func(t *testing.T) {
		// Test reading the auth service manifest
		content, err := client.GetFileContent(ctx, gitConfig, "backend/tests/testdata/services/auth/manifest.yaml")
		require.NoError(t, err)

		// Parse the manifest
		parser := models.NewParser()
		manifest, err := parser.Parse(content)
		require.NoError(t, err)

		// Validate the content
		err = parser.Validate(manifest)
		require.NoError(t, err)

		assert.Equal(t, "auth-service", manifest.Name)
	})

	t.Run("get latest commit", func(t *testing.T) {
		commit, err := client.GetLatestCommit(ctx, gitConfig)
		require.NoError(t, err)
		assert.NotEmpty(t, commit, "Should return a valid commit hash")
		assert.Len(t, commit, 40, "Git commit hash should be 40 characters")
	})
}

func TestExample_ManifestValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	skipIfRepositoryNotAccessible(t)

	ctx := context.Background()
	client := NewGitClient()
	parser := models.NewParser()

	gitConfig := GitSourceConfig{
		URL:      getTestRepositoryURL(),
		Branch:   "main",
		BasePath: "backend/tests/testdata",
	}

	// Test all manifest files can be parsed and validated
	manifests, err := client.FindManifests(ctx, gitConfig)
	require.NoError(t, err)

	expectedComponents := map[string]string{
		"backend/tests/testdata/services/auth/manifest.yaml":          "auth-service",
		"backend/tests/testdata/services/api/manifest.yaml":           "api-gateway",
		"backend/tests/testdata/services/user/manifest.yml":           "user-service",
		"backend/tests/testdata/platform/infrastructure/manifest.yml": "platform-infrastructure",
	}

	for _, manifestPath := range manifests {
		t.Run("validate "+manifestPath, func(t *testing.T) {
			// Read content
			content, err := client.GetFileContent(ctx, gitConfig, manifestPath)
			require.NoError(t, err, "Should be able to read manifest file")

			// Parse
			manifest, err := parser.Parse(content)
			require.NoError(t, err, "Should be able to parse manifest YAML")

			// Validate
			err = parser.Validate(manifest)
			require.NoError(t, err, "Manifest should pass validation")

			// Check expected name
			expectedName := expectedComponents[manifestPath]
			assert.Equal(t, expectedName, manifest.Name, "Component name should match expected value")
		})
	}
}

// Helper function to run integration tests
func TestExample_FullEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	skipIfRepositoryNotAccessible(t)

	ctx := context.Background()

	// Test the complete flow: GitFetcher + Service integration
	mockRepo := &IntegrationMockRepository{}

	// Set up expectations for all components
	mockRepo.On("GetComponentByName", mock.Anything, mock.AnythingOfType("string")).Return(nil, storage.ErrComponentNotFound)
	mockRepo.On("CreateComponent", mock.Anything, mock.AnythingOfType("storage.Component")).Return(nil)

	// Create fetcher with real git client
	fetcher := NewGitFetcher()

	sourceConfig := SourceConfig{
		Type:     "git",
		URL:      getTestRepositoryURL(),
		Branch:   "main",
		BasePath: "backend/tests/testdata",
		Interval: time.Minute,
	}

	// Test the fetcher directly
	components, err := fetcher.Fetch(ctx, sourceConfig)
	require.NoError(t, err)

	// Verify we got all expected components
	assert.Len(t, components, 4, "Should fetch 4 components from examples directory")

	componentNames := make([]string, len(components))
	for i, comp := range components {
		componentNames[i] = comp.Name
	}

	assert.Contains(t, componentNames, "auth-service")
	assert.Contains(t, componentNames, "api-gateway")
	assert.Contains(t, componentNames, "user-service")
	assert.Contains(t, componentNames, "platform-infrastructure")

	// Now test the service with the same configuration
	config := Config{
		Sources: []SourceConfig{sourceConfig},
	}

	service := NewService(mockRepo, config)
	err = service.SyncSource(ctx, sourceConfig)
	require.NoError(t, err)

	// Verify all components were processed
	mockRepo.AssertExpectations(t)
	assert.Len(t, mockRepo.createdComponents, 4, "Service should have created all 4 components")
}
