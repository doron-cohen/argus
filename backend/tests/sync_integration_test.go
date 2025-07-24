package integration

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/doron-cohen/argus/backend/api/client"
	"github.com/doron-cohen/argus/backend/internal/server"
	"github.com/doron-cohen/argus/backend/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getTestDataPath returns the absolute path to the testdata directory
func getTestDataPath(t *testing.T) string {
	wd, err := os.Getwd()
	require.NoError(t, err)
	return filepath.Join(wd, "testdata")
}

// getTestRepositoryURL returns the repository URL for git integration testing
func getTestRepositoryURL() string {
	if url := os.Getenv("ARGUS_TEST_REPO_URL"); url != "" {
		return url
	}
	// Default to the main argus repository
	return "https://github.com/doron-cohen/argus"
}

// skipIfRepositoryNotAccessible checks if the repository is accessible and skips the test if not
func skipIfRepositoryNotAccessible(t *testing.T) {
	ctx := context.Background()
	client := sync.NewGitClient()

	gitConfig := sync.GitSourceConfig{
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

func TestFilesystemSyncIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDataPath := getTestDataPath(t)

	// Create config with filesystem source pointing to testdata
	testConfig := TestConfig
	testConfig.Sync = sync.Config{
		Sources: []sync.SourceConfig{
			{
				Type:     "filesystem",
				Path:     testDataPath,
				Interval: time.Second, // Fast sync for testing
			},
		},
	}

	// Start server with sync enabled
	stop, err := server.Start(testConfig)
	require.NoError(t, err)
	defer stop()

	// Wait for server to start and initial sync to complete
	time.Sleep(3 * time.Second)

	// Create API client
	apiClient, err := client.NewClientWithResponses("http://localhost:8080/api")
	require.NoError(t, err)

	// Get components via API
	resp, err := apiClient.GetComponentsWithResponse(context.Background())
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.NotNil(t, resp.JSON200)

	components := *resp.JSON200
	require.Len(t, components, 4, "Should have synced 4 components from testdata")

	// Verify expected components exist
	componentNames := make([]string, len(components))
	for i, comp := range components {
		componentNames[i] = *comp.Name
	}

	expectedComponents := []string{
		"auth-service",
		"api-gateway",
		"user-service",
		"platform-infrastructure",
	}

	for _, expected := range expectedComponents {
		assert.Contains(t, componentNames, expected, "Should contain component: %s", expected)
	}
}

func TestFilesystemSyncWithBasePath(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDataPath := getTestDataPath(t)

	// Test with BasePath pointing to services only
	testConfig := TestConfig
	testConfig.Sync = sync.Config{
		Sources: []sync.SourceConfig{
			{
				Type:     "filesystem",
				Path:     testDataPath,
				BasePath: "services", // Only sync services, not platform
				Interval: time.Second,
			},
		},
	}

	// Start server with sync enabled
	stop, err := server.Start(testConfig)
	require.NoError(t, err)
	defer stop()

	// Wait for server to start and sync
	time.Sleep(3 * time.Second)

	// Create API client
	apiClient, err := client.NewClientWithResponses("http://localhost:8080/api")
	require.NoError(t, err)

	// Get components via API
	resp, err := apiClient.GetComponentsWithResponse(context.Background())
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.NotNil(t, resp.JSON200)

	components := *resp.JSON200
	require.Len(t, components, 3, "Should have synced 3 service components only")

	// Verify only service components exist (no platform components)
	componentNames := make([]string, len(components))
	for i, comp := range components {
		componentNames[i] = *comp.Name
	}

	expectedServices := []string{
		"auth-service",
		"api-gateway",
		"user-service",
	}

	for _, expected := range expectedServices {
		assert.Contains(t, componentNames, expected, "Should contain service: %s", expected)
	}

	// Verify platform component is NOT present
	assert.NotContains(t, componentNames, "platform-infrastructure",
		"Should not contain platform component when BasePath=services")
}

func TestGitSyncIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	skipIfRepositoryNotAccessible(t)

	// Create config with git source pointing to testdata in repository
	testConfig := TestConfig
	testConfig.Sync = sync.Config{
		Sources: []sync.SourceConfig{
			{
				Type:     "git",
				URL:      getTestRepositoryURL(),
				Branch:   "main",
				BasePath: "backend/tests/testdata", // Point to moved testdata location
				Interval: time.Second,
			},
		},
	}

	// Start server with sync enabled
	stop, err := server.Start(testConfig)
	require.NoError(t, err)
	defer stop()

	// Wait for server to start and git sync to complete (takes longer than filesystem)
	time.Sleep(10 * time.Second)

	// Create API client
	apiClient, err := client.NewClientWithResponses("http://localhost:8080/api")
	require.NoError(t, err)

	// Get components via API
	resp, err := apiClient.GetComponentsWithResponse(context.Background())
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.NotNil(t, resp.JSON200)

	components := *resp.JSON200
	require.Len(t, components, 4, "Should have synced 4 components from git repository")

	// Verify expected components exist
	componentNames := make([]string, len(components))
	for i, comp := range components {
		componentNames[i] = *comp.Name
	}

	expectedComponents := []string{
		"auth-service",
		"api-gateway",
		"user-service",
		"platform-infrastructure",
	}

	for _, expected := range expectedComponents {
		assert.Contains(t, componentNames, expected, "Should contain component: %s", expected)
	}
}

func TestMixedSourcesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	skipIfRepositoryNotAccessible(t)

	testDataPath := getTestDataPath(t)

	// Create config with both filesystem and git sources
	testConfig := TestConfig
	testConfig.Sync = sync.Config{
		Sources: []sync.SourceConfig{
			// Filesystem source for services only
			{
				Type:     "filesystem",
				Path:     testDataPath,
				BasePath: "services",
				Interval: time.Second,
			},
			// Git source for platform only
			{
				Type:     "git",
				URL:      getTestRepositoryURL(),
				Branch:   "main",
				BasePath: "backend/tests/testdata/platform",
				Interval: time.Second,
			},
		},
	}

	// Start server with sync enabled
	stop, err := server.Start(testConfig)
	require.NoError(t, err)
	defer stop()

	// Wait for both syncs to complete
	time.Sleep(12 * time.Second)

	// Create API client
	apiClient, err := client.NewClientWithResponses("http://localhost:8080/api")
	require.NoError(t, err)

	// Get components via API
	resp, err := apiClient.GetComponentsWithResponse(context.Background())
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.NotNil(t, resp.JSON200)

	components := *resp.JSON200
	require.Len(t, components, 4, "Should have synced components from both sources")

	// Verify all components exist (services from filesystem, platform from git)
	componentNames := make([]string, len(components))
	for i, comp := range components {
		componentNames[i] = *comp.Name
	}

	expectedComponents := []string{
		"auth-service",            // From filesystem
		"api-gateway",             // From filesystem
		"user-service",            // From filesystem
		"platform-infrastructure", // From git
	}

	for _, expected := range expectedComponents {
		assert.Contains(t, componentNames, expected, "Should contain component: %s", expected)
	}
}

func TestSyncWithNoSources(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create config with no sync sources
	testConfig := TestConfig
	testConfig.Sync = sync.Config{
		Sources: []sync.SourceConfig{}, // Empty sources
	}

	// Start server - should start successfully but log warning about no sources
	stop, err := server.Start(testConfig)
	require.NoError(t, err)
	defer stop()

	// Wait for server to start
	time.Sleep(2 * time.Second)

	// Create API client
	apiClient, err := client.NewClientWithResponses("http://localhost:8080/api")
	require.NoError(t, err)

	// Get components via API - should be empty since no sync occurred
	resp, err := apiClient.GetComponentsWithResponse(context.Background())
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.NotNil(t, resp.JSON200)

	components := *resp.JSON200
	require.Len(t, components, 0, "Should have no components when no sources configured")
}

func TestSyncErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create config with invalid filesystem path
	testConfig := TestConfig
	testConfig.Sync = sync.Config{
		Sources: []sync.SourceConfig{
			{
				Type:     "filesystem",
				Path:     "/non/existent/path",
				Interval: time.Second,
			},
		},
	}

	// Start server - should start successfully even with invalid source
	stop, err := server.Start(testConfig)
	require.NoError(t, err)
	defer stop()

	// Wait for server to start and sync attempts
	time.Sleep(3 * time.Second)

	// Create API client
	apiClient, err := client.NewClientWithResponses("http://localhost:8080/api")
	require.NoError(t, err)

	// Get components via API - should be empty due to sync failures
	resp, err := apiClient.GetComponentsWithResponse(context.Background())
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.NotNil(t, resp.JSON200)

	components := *resp.JSON200
	require.Len(t, components, 0, "Should have no components when sync source is invalid")
}

func TestSyncPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDataPath := getTestDataPath(t)

	// Test filesystem sync performance
	testConfig := TestConfig
	testConfig.Sync = sync.Config{
		Sources: []sync.SourceConfig{
			{
				Type:     "filesystem",
				Path:     testDataPath,
				Interval: 100 * time.Millisecond, // Very fast sync
			},
		},
	}

	start := time.Now()

	// Start server with sync enabled
	stop, err := server.Start(testConfig)
	require.NoError(t, err)
	defer stop()

	// Wait for multiple sync cycles
	time.Sleep(2 * time.Second)

	elapsed := time.Since(start)

	// Create API client and verify components exist
	apiClient, err := client.NewClientWithResponses("http://localhost:8080/api")
	require.NoError(t, err)

	resp, err := apiClient.GetComponentsWithResponse(context.Background())
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.NotNil(t, resp.JSON200)

	components := *resp.JSON200
	require.Len(t, components, 4, "Should have synced all components")

	t.Logf("Filesystem sync performance: %d components synced in %v with 100ms intervals",
		len(components), elapsed)

	// Filesystem sync should be very fast
	assert.Less(t, elapsed, 5*time.Second, "Filesystem sync should complete quickly")
}
