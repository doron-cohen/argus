package integration

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	catalogclient "github.com/doron-cohen/argus/backend/api/client"
	"github.com/doron-cohen/argus/backend/sync"
	syncclient "github.com/doron-cohen/argus/backend/sync/api/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSyncStatusComponentsCountBasic tests the basic functionality of componentsCount reporting
func TestSyncStatusComponentsCountBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Clear database before test
	clearDatabase(t)

	testDataPath := getTestDataPath(t)

	// Create config with filesystem source pointing to testdata
	testConfig := TestConfig
	fsConfig := sync.NewFilesystemSourceConfig(testDataPath, 1*time.Second)
	testConfig.Sync = sync.Config{
		Sources: []sync.SourceConfig{
			sync.NewSourceConfig(fsConfig.GetConfig()),
		},
	}

	// Start server with sync enabled and wait for health
	stop := startServerAndWaitForHealth(t, testConfig)
	defer stop()

	// Create API clients
	syncClient, err := syncclient.NewClientWithResponses("http://localhost:8080/api/sync/v1")
	require.NoError(t, err)
	catalogClient, err := catalogclient.NewClientWithResponses("http://localhost:8080/api/catalog/v1")
	require.NoError(t, err)

	t.Run("InitialStatusAfterServerStart", func(t *testing.T) {
		// Check sync status after server start (initial sync has already happened)
		statusResp, err := syncClient.GetSyncSourceStatusWithResponse(context.Background(), 0)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusResp.StatusCode())
		require.NotNil(t, statusResp.JSON200)

		status := *statusResp.JSON200
		require.NotNil(t, status.SourceId)
		assert.Equal(t, 0, *status.SourceId)
		require.NotNil(t, status.Status)
		assert.Contains(t, []string{"idle", "running", "completed"}, string(*status.Status))

		// After initial sync, componentsCount should reflect the actual components
		if status.ComponentsCount != nil {
			assert.Equal(t, 4, *status.ComponentsCount, "ComponentsCount should be 4 after initial sync")
		}
	})

	t.Run("ComponentsCountAfterSuccessfulSync", func(t *testing.T) {
		// Trigger manual sync
		triggerResp, err := syncClient.TriggerSyncSourceWithResponse(context.Background(), 0)
		require.NoError(t, err)
		require.Equal(t, http.StatusAccepted, triggerResp.StatusCode())

		// Wait for sync to complete
		time.Sleep(3 * time.Second)

		// Check sync status after sync
		statusResp, err := syncClient.GetSyncSourceStatusWithResponse(context.Background(), 0)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusResp.StatusCode())
		require.NotNil(t, statusResp.JSON200)

		status := *statusResp.JSON200
		require.NotNil(t, status.Status)
		assert.Equal(t, "completed", string(*status.Status), "Sync should be completed")
		require.NotNil(t, status.LastSync, "LastSync should be set")
		require.NotNil(t, status.Duration, "Duration should be set")
		assert.Nil(t, status.LastError, "LastError should be nil for successful sync")

		// Verify componentsCount is correctly set to 4
		require.NotNil(t, status.ComponentsCount, "ComponentsCount should not be nil")
		assert.Equal(t, 4, *status.ComponentsCount, "ComponentsCount should be 4 after successful sync")

		// Verify components were actually created in the database
		catalogResp, err := catalogClient.GetComponentsWithResponse(context.Background())
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, catalogResp.StatusCode())
		require.NotNil(t, catalogResp.JSON200)

		components := *catalogResp.JSON200
		assert.Len(t, components, 4, "Should have 4 components in database")
		assert.Equal(t, 4, *status.ComponentsCount, "ComponentsCount should match actual component count")
	})
}

// TestSyncStatusComponentsCountEdgeCases tests various edge cases for componentsCount reporting
func TestSyncStatusComponentsCountEdgeCases(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("ComponentsCountWithEmptySource", func(t *testing.T) {
		// Clear database before test
		clearDatabase(t)

		// Create temporary empty directory
		emptyDir, err := os.MkdirTemp("", "empty-sync-test")
		require.NoError(t, err)
		defer os.RemoveAll(emptyDir)

		// Create config with filesystem source pointing to empty directory
		testConfig := TestConfig
		fsConfig := sync.NewFilesystemSourceConfig(emptyDir, 1*time.Second)
		testConfig.Sync = sync.Config{
			Sources: []sync.SourceConfig{
				sync.NewSourceConfig(fsConfig.GetConfig()),
			},
		}

		// Start server with sync enabled and wait for health
		stop := startServerAndWaitForHealth(t, testConfig)
		defer stop()

		// Create sync API client
		syncClient, err := syncclient.NewClientWithResponses("http://localhost:8080/api/sync/v1")
		require.NoError(t, err)

		// Trigger sync
		triggerResp, err := syncClient.TriggerSyncSourceWithResponse(context.Background(), 0)
		require.NoError(t, err)
		require.Equal(t, http.StatusAccepted, triggerResp.StatusCode())

		// Wait for sync to complete
		time.Sleep(3 * time.Second)

		// Check sync status
		statusResp, err := syncClient.GetSyncSourceStatusWithResponse(context.Background(), 0)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusResp.StatusCode())
		require.NotNil(t, statusResp.JSON200)

		status := *statusResp.JSON200
		require.NotNil(t, status.Status)
		assert.Equal(t, "completed", string(*status.Status), "Sync should be completed")
		require.NotNil(t, status.ComponentsCount, "ComponentsCount should not be nil")
		assert.Equal(t, 0, *status.ComponentsCount, "ComponentsCount should be 0 for empty source")
	})

	t.Run("ComponentsCountWithInvalidSource", func(t *testing.T) {
		// Clear database before test
		clearDatabase(t)

		// Create config with non-existent filesystem path
		testConfig := TestConfig
		fsConfig := sync.NewFilesystemSourceConfig("/non/existent/path", 1*time.Second)
		testConfig.Sync = sync.Config{
			Sources: []sync.SourceConfig{
				sync.NewSourceConfig(fsConfig.GetConfig()),
			},
		}

		// Start server with sync enabled and wait for health
		stop := startServerAndWaitForHealth(t, testConfig)
		defer stop()

		// Create sync API client
		syncClient, err := syncclient.NewClientWithResponses("http://localhost:8080/api/sync/v1")
		require.NoError(t, err)

		// Trigger sync
		triggerResp, err := syncClient.TriggerSyncSourceWithResponse(context.Background(), 0)
		require.NoError(t, err)
		require.Equal(t, http.StatusAccepted, triggerResp.StatusCode())

		// Wait for sync to complete
		time.Sleep(3 * time.Second)

		// Check sync status
		statusResp, err := syncClient.GetSyncSourceStatusWithResponse(context.Background(), 0)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusResp.StatusCode())
		require.NotNil(t, statusResp.JSON200)

		status := *statusResp.JSON200
		require.NotNil(t, status.Status)
		assert.Equal(t, "failed", string(*status.Status), "Sync should be failed")
		require.NotNil(t, status.LastError, "LastError should be set for failed sync")
		require.NotNil(t, status.ComponentsCount, "ComponentsCount should not be nil")
		assert.Equal(t, 0, *status.ComponentsCount, "ComponentsCount should be 0 for failed sync")
	})

	t.Run("ComponentsCountWithPartialSource", func(t *testing.T) {
		// Clear database before test
		clearDatabase(t)

		testDataPath := getTestDataPath(t)
		servicesPath := filepath.Join(testDataPath, "services")

		// Create config with filesystem source pointing to services subdirectory only
		testConfig := TestConfig
		fsConfig := sync.NewFilesystemSourceConfig(servicesPath, 1*time.Second)
		testConfig.Sync = sync.Config{
			Sources: []sync.SourceConfig{
				sync.NewSourceConfig(fsConfig.GetConfig()),
			},
		}

		// Start server with sync enabled and wait for health
		stop := startServerAndWaitForHealth(t, testConfig)
		defer stop()

		// Create API clients
		syncClient, err := syncclient.NewClientWithResponses("http://localhost:8080/api/sync/v1")
		require.NoError(t, err)
		catalogClient, err := catalogclient.NewClientWithResponses("http://localhost:8080/api/catalog/v1")
		require.NoError(t, err)

		// Trigger sync
		triggerResp, err := syncClient.TriggerSyncSourceWithResponse(context.Background(), 0)
		require.NoError(t, err)
		require.Equal(t, http.StatusAccepted, triggerResp.StatusCode())

		// Wait for sync to complete
		time.Sleep(3 * time.Second)

		// Check sync status
		statusResp, err := syncClient.GetSyncSourceStatusWithResponse(context.Background(), 0)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusResp.StatusCode())
		require.NotNil(t, statusResp.JSON200)

		status := *statusResp.JSON200
		require.NotNil(t, status.Status)
		assert.Equal(t, "completed", string(*status.Status), "Sync should be completed")
		require.NotNil(t, status.ComponentsCount, "ComponentsCount should not be nil")
		assert.Equal(t, 3, *status.ComponentsCount, "ComponentsCount should be 3 for services subdirectory")

		// Verify components were actually created in the database
		catalogResp, err := catalogClient.GetComponentsWithResponse(context.Background())
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, catalogResp.StatusCode())
		require.NotNil(t, catalogResp.JSON200)

		components := *catalogResp.JSON200
		assert.Len(t, components, 3, "Should have 3 components in database")
		assert.Equal(t, 3, *status.ComponentsCount, "ComponentsCount should match actual component count")
	})
}

// TestSyncStatusComponentsCountMultipleSyncs tests componentsCount behavior across multiple sync operations
func TestSyncStatusComponentsCountMultipleSyncs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Clear database before test
	clearDatabase(t)

	testDataPath := getTestDataPath(t)

	// Create config with filesystem source pointing to testdata
	testConfig := TestConfig
	fsConfig := sync.NewFilesystemSourceConfig(testDataPath, 1*time.Second)
	testConfig.Sync = sync.Config{
		Sources: []sync.SourceConfig{
			sync.NewSourceConfig(fsConfig.GetConfig()),
		},
	}

	// Start server with sync enabled and wait for health
	stop := startServerAndWaitForHealth(t, testConfig)
	defer stop()

	// Create sync API client
	syncClient, err := syncclient.NewClientWithResponses("http://localhost:8080/api/sync/v1")
	require.NoError(t, err)

	t.Run("ComponentsCountPersistsAcrossMultipleSyncs", func(t *testing.T) {
		// First sync
		triggerResp1, err := syncClient.TriggerSyncSourceWithResponse(context.Background(), 0)
		require.NoError(t, err)
		require.Equal(t, http.StatusAccepted, triggerResp1.StatusCode())

		// Wait for sync to complete
		time.Sleep(3 * time.Second)

		// Check first sync status
		statusResp1, err := syncClient.GetSyncSourceStatusWithResponse(context.Background(), 0)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusResp1.StatusCode())
		require.NotNil(t, statusResp1.JSON200)

		status1 := *statusResp1.JSON200
		require.NotNil(t, status1.ComponentsCount)
		assert.Equal(t, 4, *status1.ComponentsCount, "First sync should report 4 components")

		// Second sync (should still report 4 since components already exist)
		triggerResp2, err := syncClient.TriggerSyncSourceWithResponse(context.Background(), 0)
		require.NoError(t, err)
		require.Equal(t, http.StatusAccepted, triggerResp2.StatusCode())

		// Wait for sync to complete
		time.Sleep(3 * time.Second)

		// Check second sync status
		statusResp2, err := syncClient.GetSyncSourceStatusWithResponse(context.Background(), 0)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusResp2.StatusCode())
		require.NotNil(t, statusResp2.JSON200)

		status2 := *statusResp2.JSON200
		require.NotNil(t, status2.ComponentsCount)
		assert.Equal(t, 4, *status2.ComponentsCount, "Second sync should still report 4 components")
		assert.True(t, status2.LastSync.After(*status1.LastSync), "Second sync should have later timestamp")
	})

	t.Run("ComponentsCountAfterTriggeredSync", func(t *testing.T) {
		// Trigger sync
		triggerResp, err := syncClient.TriggerSyncSourceWithResponse(context.Background(), 0)
		require.NoError(t, err)
		require.Equal(t, http.StatusAccepted, triggerResp.StatusCode())

		// Wait for sync to complete
		time.Sleep(3 * time.Second)

		// Check status after sync completion
		statusResp, err := syncClient.GetSyncSourceStatusWithResponse(context.Background(), 0)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusResp.StatusCode())
		require.NotNil(t, statusResp.JSON200)

		status := *statusResp.JSON200
		require.NotNil(t, status.Status)
		assert.Equal(t, "completed", string(*status.Status), "Sync should be completed")
		require.NotNil(t, status.ComponentsCount)
		assert.Equal(t, 4, *status.ComponentsCount, "ComponentsCount should be 4 after triggered sync")
	})
}

// TestSyncStatusComponentsCountConcurrent tests componentsCount behavior with concurrent sync operations
func TestSyncStatusComponentsCountConcurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Clear database before test
	clearDatabase(t)

	testDataPath := getTestDataPath(t)

	// Create config with filesystem source pointing to testdata
	testConfig := TestConfig
	fsConfig := sync.NewFilesystemSourceConfig(testDataPath, 1*time.Second)
	testConfig.Sync = sync.Config{
		Sources: []sync.SourceConfig{
			sync.NewSourceConfig(fsConfig.GetConfig()),
		},
	}

	// Start server with sync enabled and wait for health
	stop := startServerAndWaitForHealth(t, testConfig)
	defer stop()

	// Create sync API client
	syncClient, err := syncclient.NewClientWithResponses("http://localhost:8080/api/sync/v1")
	require.NoError(t, err)

	t.Run("ConcurrentSyncTriggers", func(t *testing.T) {
		// Trigger first sync
		triggerResp1, err := syncClient.TriggerSyncSourceWithResponse(context.Background(), 0)
		require.NoError(t, err)
		require.Equal(t, http.StatusAccepted, triggerResp1.StatusCode())

		// Immediately trigger second sync (should be rejected)
		triggerResp2, err := syncClient.TriggerSyncSourceWithResponse(context.Background(), 0)
		require.NoError(t, err)
		assert.Equal(t, http.StatusConflict, triggerResp2.StatusCode(), "Second sync should be rejected")

		// Wait for first sync to complete
		time.Sleep(3 * time.Second)

		// Check final status
		statusResp, err := syncClient.GetSyncSourceStatusWithResponse(context.Background(), 0)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusResp.StatusCode())
		require.NotNil(t, statusResp.JSON200)

		status := *statusResp.JSON200
		require.NotNil(t, status.Status)
		assert.Equal(t, "completed", string(*status.Status), "Sync should be completed")
		require.NotNil(t, status.ComponentsCount)
		assert.Equal(t, 4, *status.ComponentsCount, "ComponentsCount should be 4 after successful sync")
	})
}

// TestSyncStatusComponentsCountErrorHandling tests componentsCount behavior when sync encounters errors
func TestSyncStatusComponentsCountErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("ComponentsCountAfterSyncError", func(t *testing.T) {
		// Clear database before test
		clearDatabase(t)

		// Create config with invalid source
		testConfig := TestConfig
		fsConfig := sync.NewFilesystemSourceConfig("/non/existent/path", 1*time.Second)
		testConfig.Sync = sync.Config{
			Sources: []sync.SourceConfig{
				sync.NewSourceConfig(fsConfig.GetConfig()),
			},
		}

		// Start server with sync enabled and wait for health
		stop := startServerAndWaitForHealth(t, testConfig)
		defer stop()

		// Create sync API client
		syncClient, err := syncclient.NewClientWithResponses("http://localhost:8080/api/sync/v1")
		require.NoError(t, err)

		// Trigger sync
		triggerResp, err := syncClient.TriggerSyncSourceWithResponse(context.Background(), 0)
		require.NoError(t, err)
		require.Equal(t, http.StatusAccepted, triggerResp.StatusCode())

		// Wait for sync to complete
		time.Sleep(3 * time.Second)

		// Check sync status
		statusResp, err := syncClient.GetSyncSourceStatusWithResponse(context.Background(), 0)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusResp.StatusCode())
		require.NotNil(t, statusResp.JSON200)

		status := *statusResp.JSON200
		require.NotNil(t, status.Status)
		assert.Equal(t, "failed", string(*status.Status), "Sync should be failed")
		require.NotNil(t, status.LastError, "LastError should be set")
		require.NotNil(t, status.ComponentsCount)
		assert.Equal(t, 0, *status.ComponentsCount, "ComponentsCount should be 0 for failed sync")
	})
}

// TestSyncStatusComponentsCountScheduledSync tests componentsCount behavior with scheduled sync (no manual trigger)
func TestSyncStatusComponentsCountScheduledSync(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Clear database before test
	clearDatabase(t)

	testDataPath := getTestDataPath(t)

	// Create config with filesystem source pointing to testdata with short interval
	testConfig := TestConfig
	fsConfig := sync.NewFilesystemSourceConfig(testDataPath, 2*time.Second)
	testConfig.Sync = sync.Config{
		Sources: []sync.SourceConfig{
			sync.NewSourceConfig(fsConfig.GetConfig()),
		},
	}

	// Start server with sync enabled and wait for health
	stop := startServerAndWaitForHealth(t, testConfig)
	defer stop()

	// Create API clients
	syncClient, err := syncclient.NewClientWithResponses("http://localhost:8080/api/sync/v1")
	require.NoError(t, err)
	catalogClient, err := catalogclient.NewClientWithResponses("http://localhost:8080/api/catalog/v1")
	require.NoError(t, err)

	t.Run("ComponentsCountAfterScheduledSync", func(t *testing.T) {
		// Wait for initial sync to complete (it happens immediately when server starts)
		time.Sleep(3 * time.Second)

		// Check status after initial sync
		initialStatusResp, err := syncClient.GetSyncSourceStatusWithResponse(context.Background(), 0)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, initialStatusResp.StatusCode())
		require.NotNil(t, initialStatusResp.JSON200)

		initialStatus := *initialStatusResp.JSON200
		require.NotNil(t, initialStatus.Status)
		assert.Equal(t, "completed", string(*initialStatus.Status), "Initial sync should be completed")
		require.NotNil(t, initialStatus.ComponentsCount)
		assert.Equal(t, 4, *initialStatus.ComponentsCount, "Initial sync should report 4 components")

		// Wait for next scheduled sync to complete
		time.Sleep(3 * time.Second)

		// Check sync status after scheduled sync
		statusResp, err := syncClient.GetSyncSourceStatusWithResponse(context.Background(), 0)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusResp.StatusCode())
		require.NotNil(t, statusResp.JSON200)

		status := *statusResp.JSON200
		require.NotNil(t, status.Status)
		assert.Equal(t, "completed", string(*status.Status), "Sync should be completed")
		require.NotNil(t, status.LastSync, "LastSync should be set")
		require.NotNil(t, status.Duration, "Duration should be set")
		assert.Nil(t, status.LastError, "LastError should be nil for successful sync")

		// Verify componentsCount is correctly set to 4
		require.NotNil(t, status.ComponentsCount, "ComponentsCount should not be nil")
		assert.Equal(t, 4, *status.ComponentsCount, "ComponentsCount should be 4 after scheduled sync")

		// Verify components were actually created in the database
		catalogResp, err := catalogClient.GetComponentsWithResponse(context.Background())
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, catalogResp.StatusCode())
		require.NotNil(t, catalogResp.JSON200)

		components := *catalogResp.JSON200
		assert.Len(t, components, 4, "Should have 4 components in database")
		assert.Equal(t, 4, *status.ComponentsCount, "ComponentsCount should match actual component count")

		// Verify that the last sync time is more recent than the initial check
		if initialStatus.LastSync != nil {
			assert.True(t, status.LastSync.After(*initialStatus.LastSync), "LastSync should be updated after scheduled sync")
		}
	})

	t.Run("ComponentsCountPersistsAcrossScheduledSyncs", func(t *testing.T) {
		// Get current status
		statusResp1, err := syncClient.GetSyncSourceStatusWithResponse(context.Background(), 0)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusResp1.StatusCode())
		require.NotNil(t, statusResp1.JSON200)

		status1 := *statusResp1.JSON200
		require.NotNil(t, status1.ComponentsCount)
		assert.Equal(t, 4, *status1.ComponentsCount, "First scheduled sync should report 4 components")

		// Wait for next scheduled sync
		time.Sleep(3 * time.Second)

		// Check status after second scheduled sync
		statusResp2, err := syncClient.GetSyncSourceStatusWithResponse(context.Background(), 0)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusResp2.StatusCode())
		require.NotNil(t, statusResp2.JSON200)

		status2 := *statusResp2.JSON200
		require.NotNil(t, status2.ComponentsCount)
		assert.Equal(t, 4, *status2.ComponentsCount, "Second scheduled sync should still report 4 components")
		assert.True(t, status2.LastSync.After(*status1.LastSync), "Second scheduled sync should have later timestamp")
	})
}

// TestSyncStatusComponentsCountNoSources tests componentsCount behavior when no sources are configured
func TestSyncStatusComponentsCountNoSources(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Clear database before test
	clearDatabase(t)

	// Create config with no sync sources
	testConfig := TestConfig
	testConfig.Sync = sync.Config{
		Sources: []sync.SourceConfig{}, // Empty sources
	}

	// Start server and wait for health
	stop := startServerAndWaitForHealth(t, testConfig)
	defer stop()

	// Create sync API client
	syncClient, err := syncclient.NewClientWithResponses("http://localhost:8080/api/sync/v1")
	require.NoError(t, err)

	t.Run("NoSourcesConfigured", func(t *testing.T) {
		// Try to get status for non-existent source
		statusResp, err := syncClient.GetSyncSourceStatusWithResponse(context.Background(), 0)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, statusResp.StatusCode(), "Should return 404 for non-existent source")

		// Try to trigger sync for non-existent source
		triggerResp, err := syncClient.TriggerSyncSourceWithResponse(context.Background(), 0)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, triggerResp.StatusCode(), "Should return 404 for non-existent source")
	})
}
