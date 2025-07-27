package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/doron-cohen/argus/backend/api/client"
	"github.com/doron-cohen/argus/backend/internal/server"
	reportsclient "github.com/doron-cohen/argus/backend/reports/api/client"
	"github.com/doron-cohen/argus/backend/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponentReportsAPIEndpoints(t *testing.T) {
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

	// Start server with sync enabled
	stop, err := server.Start(testConfig)
	require.NoError(t, err)
	defer stop()

	// Wait for server to start and initial sync to complete
	time.Sleep(3 * time.Second)

	// Create API clients
	apiClient, err := client.NewClientWithResponses("http://localhost:8080/api")
	require.NoError(t, err)

	reportsClient, err := reportsclient.NewClientWithResponses("http://localhost:8080/reports")
	require.NoError(t, err)

	// Setup test data
	t.Run("SetupTestData", func(t *testing.T) {
		// Components should already exist from sync (auth-service, user-service, etc.)
		// Now submit reports for these components
		components := []string{"auth-service", "user-service"}
		checks := []string{"unit-tests", "integration-tests", "security-scan", "performance-tests"}

		// Create reports with different timestamps and statuses
		for i, component := range components {
			for j, check := range checks {
				// Create multiple reports for each check with different timestamps
				for k := 0; k < 3; k++ {
					status := reportsclient.ReportSubmissionStatusPass
					if k == 1 {
						status = reportsclient.ReportSubmissionStatusFail
					} else if k == 2 {
						status = reportsclient.ReportSubmissionStatusPass
					}

					report := reportsclient.ReportSubmission{
						Check: reportsclient.Check{
							Slug:        check,
							Name:        &[]string{fmt.Sprintf("%s Check", check)}[0],
							Description: &[]string{fmt.Sprintf("Runs %s for %s", check, component)}[0],
						},
						ComponentId: component,
						Status:      status,
						Timestamp:   time.Now().Add(-time.Duration(i+j+k) * time.Hour),
						Details: &map[string]interface{}{
							"coverage_percentage": 80 + (i * 5) + (j * 2) + k,
							"tests_passed":        100 + (i * 10) + (j * 5) + k,
							"tests_failed":        k,
							"duration_seconds":    30 + (i * 5) + (j * 3) + k,
						},
						Metadata: &map[string]interface{}{
							"ci_job_id":   fmt.Sprintf("job-%d-%d-%d", i, j, k),
							"environment": "staging",
							"branch":      "main",
							"commit_sha":  fmt.Sprintf("abc%d%d%d", i, j, k),
						},
					}

					resp, err := reportsClient.SubmitReportWithResponse(context.Background(), report)
					require.NoError(t, err)
					require.Equal(t, http.StatusOK, resp.StatusCode())
				}
			}
		}

		// Wait a bit for data to be processed
		time.Sleep(500 * time.Millisecond)
	})

	t.Run("GetComponentReportsBasic", func(t *testing.T) {
		// Test basic retrieval without filters
		resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
		require.NotNil(t, resp.JSON200)

		response := *resp.JSON200
		require.NotNil(t, response.Reports)
		require.NotNil(t, response.Pagination)

		// Should have reports for all 4 checks (3 reports each = 12 total)
		assert.Len(t, response.Reports, 12)
		assert.Equal(t, 12, response.Pagination.Total)
		assert.Equal(t, 50, response.Pagination.Limit) // default limit
		assert.Equal(t, 0, response.Pagination.Offset)
		assert.False(t, response.Pagination.HasMore)
	})

	t.Run("GetComponentReportsWithStatusFilter", func(t *testing.T) {
		// Test filtering by status
		status := client.GetComponentReportsParamsStatusPass
		resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{
			Status: &status,
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())

		response := *resp.JSON200
		// Should have 8 reports with pass status (2 per check * 4 checks)
		assert.Len(t, response.Reports, 8)
		assert.Equal(t, 8, response.Pagination.Total)

		// Verify all reports have pass status
		for _, report := range response.Reports {
			assert.Equal(t, client.CheckReportStatusPass, report.Status)
		}
	})

	t.Run("GetComponentReportsWithCheckSlugFilter", func(t *testing.T) {
		// Test filtering by check slug
		checkSlug := "unit-tests"
		resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{
			CheckSlug: &checkSlug,
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())

		response := *resp.JSON200
		// Should have 3 reports for unit-tests
		assert.Len(t, response.Reports, 3)
		assert.Equal(t, 3, response.Pagination.Total)

		// Verify all reports are for unit-tests
		for _, report := range response.Reports {
			assert.Equal(t, "unit-tests", report.CheckSlug)
		}
	})

	t.Run("GetComponentReportsWithSinceFilter", func(t *testing.T) {
		// Test filtering by since timestamp
		since := time.Now().Add(-2 * time.Hour)
		resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{
			Since: &since,
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())

		response := *resp.JSON200
		// Should have fewer reports (only recent ones)
		assert.True(t, len(response.Reports) < 12)
		assert.True(t, response.Pagination.Total < 12)

		// Verify all reports are after the since timestamp
		for _, report := range response.Reports {
			assert.True(t, report.Timestamp.After(since))
		}
	})

	t.Run("GetComponentReportsWithPagination", func(t *testing.T) {
		// Test pagination
		limit := 5
		offset := 0
		resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{
			Limit:  &limit,
			Offset: &offset,
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())

		response := *resp.JSON200
		// Should have 5 reports (limit)
		assert.Len(t, response.Reports, 5)
		assert.Equal(t, 12, response.Pagination.Total)
		assert.Equal(t, 5, response.Pagination.Limit)
		assert.Equal(t, 0, response.Pagination.Offset)
		assert.True(t, response.Pagination.HasMore)

		// Test second page
		offset = 5
		resp2, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{
			Limit:  &limit,
			Offset: &offset,
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp2.StatusCode())

		response2 := *resp2.JSON200
		assert.Len(t, response2.Reports, 5)
		assert.Equal(t, 5, response2.Pagination.Offset)
		assert.True(t, response2.Pagination.HasMore)

		// Test third page (should have remaining reports)
		offset = 10
		resp3, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{
			Limit:  &limit,
			Offset: &offset,
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp3.StatusCode())

		response3 := *resp3.JSON200
		assert.Len(t, response3.Reports, 2) // 12 total - 10 offset = 2 remaining
		assert.Equal(t, 10, response3.Pagination.Offset)
		assert.False(t, response3.Pagination.HasMore)
	})

	t.Run("GetComponentReportsWithLatestPerCheck", func(t *testing.T) {
		// Test latest_per_check functionality
		latestPerCheck := true
		resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{
			LatestPerCheck: &latestPerCheck,
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())

		response := *resp.JSON200
		// Should have exactly 4 reports (one per check type)
		assert.Len(t, response.Reports, 4)
		assert.Equal(t, 4, response.Pagination.Total)

		// Verify we have one report for each check type
		checkSlugs := make(map[string]bool)
		for _, report := range response.Reports {
			checkSlugs[report.CheckSlug] = true
		}
		assert.Len(t, checkSlugs, 4)
		assert.True(t, checkSlugs["unit-tests"])
		assert.True(t, checkSlugs["integration-tests"])
		assert.True(t, checkSlugs["security-scan"])
		assert.True(t, checkSlugs["performance-tests"])
	})

	t.Run("GetComponentReportsWithCombinedFilters", func(t *testing.T) {
		// Test combining multiple filters
		status := client.GetComponentReportsParamsStatusPass
		checkSlug := "unit-tests"
		limit := 2
		latestPerCheck := true

		resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{
			Status:         &status,
			CheckSlug:      &checkSlug,
			Limit:          &limit,
			LatestPerCheck: &latestPerCheck,
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())

		response := *resp.JSON200
		// Should have 1 report (latest pass for unit-tests)
		assert.Len(t, response.Reports, 1)
		assert.Equal(t, 1, response.Pagination.Total)

		report := response.Reports[0]
		assert.Equal(t, "unit-tests", report.CheckSlug)
		assert.Equal(t, client.CheckReportStatusPass, report.Status)
	})

	t.Run("GetComponentReportsComponentNotFound", func(t *testing.T) {
		// Test non-existent component
		resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), "non-existent-component", &client.GetComponentReportsParams{})
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode())
	})

	t.Run("GetComponentReportsForDifferentComponents", func(t *testing.T) {
		// Test that different components have their own reports
		latestPerCheck := true

		// Get reports for auth-service
		resp1, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{
			LatestPerCheck: &latestPerCheck,
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp1.StatusCode())

		response1 := *resp1.JSON200
		assert.Len(t, response1.Reports, 4)

		// Get reports for user-service
		resp2, err := apiClient.GetComponentReportsWithResponse(context.Background(), "user-service", &client.GetComponentReportsParams{
			LatestPerCheck: &latestPerCheck,
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp2.StatusCode())

		response2 := *resp2.JSON200
		assert.Len(t, response2.Reports, 4)

		// Verify they have different report IDs (different components)
		authReportIds := make(map[string]bool)
		for _, report := range response1.Reports {
			authReportIds[report.Id] = true
		}

		userReportIds := make(map[string]bool)
		for _, report := range response2.Reports {
			userReportIds[report.Id] = true
		}

		// Should have no overlapping report IDs
		for id := range authReportIds {
			assert.False(t, userReportIds[id], "Report IDs should not overlap between components")
		}
	})

	t.Run("GetComponentReportsWithMaxLimit", func(t *testing.T) {
		// Test maximum limit
		limit := 100
		resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{
			Limit: &limit,
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())

		response := *resp.JSON200
		assert.Len(t, response.Reports, 12) // All reports should be returned
		assert.Equal(t, 100, response.Pagination.Limit)
	})

	t.Run("GetComponentReportsWithInvalidLimit", func(t *testing.T) {
		// Test invalid limit (should use default)
		limit := 150 // Over maximum
		resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{
			Limit: &limit,
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())

		response := *resp.JSON200
		// Should use default limit of 50
		assert.Len(t, response.Reports, 12) // All reports fit within 50
		assert.Equal(t, 50, response.Pagination.Limit)
	})
}
