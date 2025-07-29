package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	syncclient "github.com/doron-cohen/argus/backend/sync/api/client"
)

// waitForSyncCompletion waits for sync to complete using the sync client
func waitForSyncCompletion(t *testing.T, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create sync client
	syncClient, err := syncclient.NewClientWithResponses("http://localhost:8080/api/sync/v1")
	require.NoError(t, err)

	// Poll for sync completion
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			require.Fail(t, "Timeout waiting for sync completion")
			return
		case <-ticker.C:
			if isSyncCompleted(t, syncClient) {
				return // Sync is complete
			}
		}
	}
}

// isSyncCompleted checks if sync is completed by examining sync sources and their status
func isSyncCompleted(t *testing.T, syncClient *syncclient.ClientWithResponses) bool {
	ctx := context.Background()

	// Check if there are any sources
	resp, err := syncClient.GetSyncSourcesWithResponse(ctx)
	if err != nil || resp.StatusCode() != 200 || resp.JSON200 == nil || len(*resp.JSON200) == 0 {
		return false // No sources configured or error, consider sync complete
	}

	// Check status of all sources
	for i, source := range *resp.JSON200 {
		if source.Id == nil {
			continue
		}

		if !isSourceCompleted(t, syncClient, *source.Id, i) {
			return false
		}
	}

	// All sources are completed
	return true
}

// isSourceCompleted checks if a specific source is completed
func isSourceCompleted(t *testing.T, syncClient *syncclient.ClientWithResponses, sourceID int, index int) bool {
	ctx := context.Background()

	statusResp, err := syncClient.GetSyncSourceStatusWithResponse(ctx, sourceID)
	if err != nil || statusResp.StatusCode() != 200 || statusResp.JSON200 == nil {
		return false
	}

	status := statusResp.JSON200
	if status.Status == nil {
		return false
	}

	// Check if sync is completed or idle
	if *status.Status != "completed" && *status.Status != "idle" {
		return false
	}

	// If this is the first source and it has components, consider sync complete
	if index == 0 && status.ComponentsCount != nil && *status.ComponentsCount > 0 {
		return true
	}

	return true
}
