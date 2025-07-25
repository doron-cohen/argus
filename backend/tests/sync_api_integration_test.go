package integration

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/doron-cohen/argus/backend/internal/server"
	"github.com/doron-cohen/argus/backend/sync"
	"github.com/doron-cohen/argus/backend/sync/api/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncAPIEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDataPath := getTestDataPath(t)

	// Create config with filesystem source pointing to testdata
	testConfig := TestConfig
	fsConfig := sync.NewFilesystemSourceConfig(testDataPath, "", time.Second)
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

	// Create sync API client
	syncClient, err := client.NewClientWithResponses("http://localhost:8080/sync")
	require.NoError(t, err)

	t.Run("GetSyncSources", func(t *testing.T) {
		// Get all sync sources
		resp, err := syncClient.GetSyncSourcesWithResponse(context.Background())
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
		require.NotNil(t, resp.JSON200)

		sources := *resp.JSON200
		require.Len(t, sources, 1, "Should have 1 configured source")

		// Verify source details
		source := sources[0]
		require.NotNil(t, source.Id)
		assert.Equal(t, 0, *source.Id)
		require.NotNil(t, source.Type)
		assert.Equal(t, "filesystem", string(*source.Type))
		require.NotNil(t, source.Interval)
		assert.Equal(t, "1s", *source.Interval)
	})

	t.Run("GetSyncSource", func(t *testing.T) {
		// Get specific source by ID
		resp, err := syncClient.GetSyncSourceWithResponse(context.Background(), 0)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
		require.NotNil(t, resp.JSON200)

		source := *resp.JSON200
		require.NotNil(t, source.Id)
		assert.Equal(t, 0, *source.Id)
		require.NotNil(t, source.Type)
		assert.Equal(t, "filesystem", string(*source.Type))
	})

	t.Run("GetSyncSourceNotFound", func(t *testing.T) {
		// Try to get non-existent source
		resp, err := syncClient.GetSyncSourceWithResponse(context.Background(), 999)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode())
	})

	t.Run("GetSyncSourceStatus", func(t *testing.T) {
		// Get status for source
		resp, err := syncClient.GetSyncSourceStatusWithResponse(context.Background(), 0)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
		require.NotNil(t, resp.JSON200)

		status := *resp.JSON200
		require.NotNil(t, status.SourceId)
		assert.Equal(t, 0, *status.SourceId)
		require.NotNil(t, status.Status)
		// Status could be idle, running, completed, or failed depending on timing
		assert.Contains(t, []string{"idle", "running", "completed", "failed"}, string(*status.Status))
	})

	t.Run("TriggerSyncSource", func(t *testing.T) {
		// Trigger manual sync
		resp, err := syncClient.TriggerSyncSourceWithResponse(context.Background(), 0)
		require.NoError(t, err)
		require.Equal(t, http.StatusAccepted, resp.StatusCode())
		require.NotNil(t, resp.JSON202)

		triggerResp := *resp.JSON202
		require.NotNil(t, triggerResp.Message)
		assert.Contains(t, *triggerResp.Message, "triggered successfully")
		require.NotNil(t, triggerResp.SourceId)
		assert.Equal(t, 0, *triggerResp.SourceId)
		require.NotNil(t, triggerResp.Triggered)
		assert.True(t, *triggerResp.Triggered)
	})

	t.Run("TriggerSyncSourceNotFound", func(t *testing.T) {
		// Try to trigger sync for non-existent source
		resp, err := syncClient.TriggerSyncSourceWithResponse(context.Background(), 999)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode())
	})
}

func TestSyncAPIWithNoSources(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create config with no sync sources
	testConfig := TestConfig
	testConfig.Sync = sync.Config{
		Sources: []sync.SourceConfig{}, // Empty sources
	}

	// Start server
	stop, err := server.Start(testConfig)
	require.NoError(t, err)
	defer stop()

	// Wait for server to start
	time.Sleep(2 * time.Second)

	// Create sync API client
	syncClient, err := client.NewClientWithResponses("http://localhost:8080/sync")
	require.NoError(t, err)

	t.Run("GetSyncSourcesEmpty", func(t *testing.T) {
		// Get all sync sources (should be empty)
		resp, err := syncClient.GetSyncSourcesWithResponse(context.Background())
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
		require.NotNil(t, resp.JSON200)

		sources := *resp.JSON200
		assert.Empty(t, sources, "Should have no sources configured")
	})

	t.Run("GetSyncSourceNotFound", func(t *testing.T) {
		// Try to get source when none configured
		resp, err := syncClient.GetSyncSourceWithResponse(context.Background(), 0)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode())
	})
}
