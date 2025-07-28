package integration

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/doron-cohen/argus/backend/api/client"
	"github.com/doron-cohen/argus/backend/internal/config"
	"github.com/doron-cohen/argus/backend/internal/server"
	"github.com/doron-cohen/argus/backend/sync"
	"github.com/stretchr/testify/require"
)

var TestConfig config.Config

func TestGetComponentsIntegration(t *testing.T) {
	// Clear database before test
	clearDatabase(t)

	stop, err := server.Start(TestConfig)
	require.NoError(t, err)
	defer stop()

	// Wait briefly for the server to start
	time.Sleep(100 * time.Millisecond)

	client, err := client.NewClientWithResponses("http://localhost:8080/api/catalog/v1")
	require.NoError(t, err)

	resp, err := client.GetComponentsWithResponse(context.Background())
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.NotNil(t, resp.JSON200)
	require.Len(t, *resp.JSON200, 0)
}

func TestGetComponentByIdIntegration(t *testing.T) {
	stop, err := server.Start(TestConfig)
	require.NoError(t, err)
	defer stop()

	// Wait briefly for the server to start
	time.Sleep(100 * time.Millisecond)

	client, err := client.NewClientWithResponses("http://localhost:8080/api/catalog/v1")
	require.NoError(t, err)

	// Test getting a non-existent component
	resp, err := client.GetComponentByIdWithResponse(context.Background(), "non-existent-component")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode())
	require.NotNil(t, resp.JSON404)
	require.Equal(t, "Component not found", resp.JSON404.Error)
	require.NotNil(t, resp.JSON404.Code)
	require.Equal(t, "NOT_FOUND", *resp.JSON404.Code)
}

func TestGetComponentByIdWithSyncIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDataPath := getTestDataPath(t)

	// Create config with filesystem source pointing to testdata
	testConfig := TestConfig
	fsConfig := sync.NewFilesystemSourceConfig(testDataPath, time.Second)
	testConfig.Sync = sync.Config{
		Sources: []sync.SourceConfig{
			sync.NewSourceConfig(fsConfig.GetConfig()),
		},
	}

	// Start server with sync enabled
	stop, err := server.Start(testConfig)
	require.NoError(t, err)
	defer stop()

	// Wait for server to start and initial sync to complete
	time.Sleep(3 * time.Second)

	// Create API client
	apiClient, err := client.NewClientWithResponses("http://localhost:8080/api/catalog/v1")
	require.NoError(t, err)

	// Test getting an existing component
	resp, err := apiClient.GetComponentByIdWithResponse(context.Background(), "auth-service")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.NotNil(t, resp.JSON200)

	component := *resp.JSON200
	require.Equal(t, "Authentication Service", component.Name)
	require.NotNil(t, component.Id)
	require.Equal(t, "auth-service", *component.Id)
	require.NotNil(t, component.Description)
	require.Equal(t, "Handles user authentication and authorization including login, logout, and session management", *component.Description)
	require.NotNil(t, component.Owners)
	require.NotNil(t, component.Owners.Maintainers)
	require.Len(t, *component.Owners.Maintainers, 2)
	require.Contains(t, *component.Owners.Maintainers, "alice@company.com")
	require.Contains(t, *component.Owners.Maintainers, "@auth-team")
	require.NotNil(t, component.Owners.Team)
	require.Equal(t, "Security Team", *component.Owners.Team)
}
