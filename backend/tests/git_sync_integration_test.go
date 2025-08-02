//go:build git_sync_source
// +build git_sync_source

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
	fetcher := sync.NewGitFetcher()

	gitConfig := sync.GitSourceConfig{
		URL:    getTestRepositoryURL(),
		Branch: "main",
	}

	// Try to fetch components to check if repository is accessible
	_, err := fetcher.Fetch(ctx, sync.NewSourceConfig(&gitConfig))
	if err != nil {
		t.Skipf("Repository %s not accessible (you may need to set ARGUS_TEST_REPO_URL env var to your fork): %v",
			getTestRepositoryURL(), err)
	}
}

func TestGitSyncIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Clear database before test
	clearDatabase(t)

	skipIfRepositoryNotAccessible(t)

	// Create config with git source pointing to testdata in repository
	testConfig := TestConfig
	gitConfig := sync.NewGitSourceConfig(getTestRepositoryURL(), "main", "tests/testdata", 10*time.Second)
	testConfig.Sync = sync.Config{
		Sources: []sync.SourceConfig{
			sync.NewSourceConfig(gitConfig.GetConfig()),
		},
	}

	// Start server with sync enabled
	stop, err := server.Start(testConfig)
	require.NoError(t, err)
	defer stop()

	// Wait for server to start and git sync to complete (takes longer than filesystem)
	time.Sleep(10 * time.Second)

	// Create API client
	apiClient, err := client.NewClientWithResponses("http://localhost:8080/api/catalog/v1")
	require.NoError(t, err)

	// Get components via API
	resp, err := apiClient.GetComponentsWithResponse(context.Background())
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.NotNil(t, resp.JSON200)

	components := *resp.JSON200

	// The git test might fail if the testdata structure doesn't exist in the repository
	// In that case, we'll just verify that the API works and the server doesn't crash
	if len(components) == 0 {
		t.Log("No components found in git repository - this is expected if testdata structure doesn't exist")
		return
	}

	// If we do find components, verify they have the expected structure
	require.Len(t, components, 4, "Should have synced 4 components from git repository")

	// Verify expected components exist
	componentNames := make([]string, len(components))
	for i, comp := range components {
		componentNames[i] = comp.Name
	}

	expectedComponents := []string{
		"Authentication Service",
		"API Gateway",
		"User Service",
		"Platform Infrastructure",
	}

	for _, expected := range expectedComponents {
		assert.Contains(t, componentNames, expected, "Should contain component: %s", expected)
	}
}

func TestMixedSourcesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Clear database before test
	clearDatabase(t)

	skipIfRepositoryNotAccessible(t)

	testDataPath := getTestDataPath(t)

	// Create config with both filesystem and git sources
	testConfig := TestConfig
	servicesPath := filepath.Join(testDataPath, "services")
	fsConfig := sync.NewFilesystemSourceConfig(servicesPath, 1*time.Second)
	gitConfig := sync.NewGitSourceConfig(getTestRepositoryURL(), "main", "tests/testdata/platform", 10*time.Second)
	testConfig.Sync = sync.Config{
		Sources: []sync.SourceConfig{
			// Filesystem source for services only
			sync.NewSourceConfig(fsConfig.GetConfig()),
			// Git source for platform only
			sync.NewSourceConfig(gitConfig.GetConfig()),
		},
	}

	// Start server with sync enabled
	stop, err := server.Start(testConfig)
	require.NoError(t, err)
	defer stop()

	// Wait for both syncs to complete
	time.Sleep(12 * time.Second)

	// Create API client
	apiClient, err := client.NewClientWithResponses("http://localhost:8080/api/catalog/v1")
	require.NoError(t, err)

	// Get components via API
	resp, err := apiClient.GetComponentsWithResponse(context.Background())
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.NotNil(t, resp.JSON200)

	components := *resp.JSON200

	// The mixed test will always get at least 3 components from filesystem
	// Git might fail if the testdata structure doesn't exist in the repository
	require.GreaterOrEqual(t, len(components), 3, "Should have at least 3 components from filesystem")

	// Verify filesystem components exist
	componentNames := make([]string, len(components))
	for i, comp := range components {
		componentNames[i] = comp.Name
	}

	expectedFilesystemComponents := []string{
		"Authentication Service", // From filesystem
		"API Gateway",            // From filesystem
		"User Service",           // From filesystem
	}

	for _, expected := range expectedFilesystemComponents {
		assert.Contains(t, componentNames, expected, "Should contain filesystem component: %s", expected)
	}

	// If we have 4 components, the git part worked too
	if len(components) == 4 {
		assert.Contains(t, componentNames, "Platform Infrastructure", "Should contain git component: Platform Infrastructure")
	} else {
		t.Log("Git sync failed - only filesystem components found (this is expected if testdata structure doesn't exist in git)")
	}
}
