package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/doron-cohen/argus/backend/api/client"
	"github.com/doron-cohen/argus/backend/internal/server"
	"github.com/doron-cohen/argus/backend/internal/utils"
	reportsclient "github.com/doron-cohen/argus/backend/reports/api/client"
	"github.com/doron-cohen/argus/backend/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateComponentReports is a helper function to create test reports for a component
func generateComponentReports(t *testing.T, reportsClient *reportsclient.ClientWithResponses, componentID string, checkSlug string, status reportsclient.ReportSubmissionStatus, count int) {
	for i := 0; i < count; i++ {
		report := reportsclient.ReportSubmission{
			Check: reportsclient.Check{
				Slug:        checkSlug,
				Name:        utils.ToPointer(fmt.Sprintf("%s Check", checkSlug)),
				Description: utils.ToPointer(fmt.Sprintf("Runs %s for %s", checkSlug, componentID)),
			},
			ComponentId: componentID,
			Status:      status,
			Timestamp:   time.Now().Add(-time.Duration(i) * time.Hour),
			Details: &map[string]interface{}{
				"coverage_percentage": 80 + i,
				"tests_passed":        100 + i,
				"tests_failed":        i,
				"duration_seconds":    30 + i,
			},
			Metadata: &map[string]interface{}{
				"ci_job_id":   fmt.Sprintf("job-%s-%d", checkSlug, i),
				"environment": "staging",
				"branch":      "main",
				"commit_sha":  fmt.Sprintf("abc%s%d", checkSlug, i),
			},
		}

		resp, err := reportsClient.SubmitReportWithResponse(context.Background(), report)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
	}
}

// generateComponentReportsWithTimestamps creates reports with specific timestamps
func generateComponentReportsWithTimestamps(t *testing.T, reportsClient *reportsclient.ClientWithResponses, componentID string, checkSlug string, status reportsclient.ReportSubmissionStatus, timestamps []time.Time) {
	for i, timestamp := range timestamps {
		report := reportsclient.ReportSubmission{
			Check: reportsclient.Check{
				Slug:        checkSlug,
				Name:        utils.ToPointer(fmt.Sprintf("%s Check", checkSlug)),
				Description: utils.ToPointer(fmt.Sprintf("Runs %s for %s", checkSlug, componentID)),
			},
			ComponentId: componentID,
			Status:      status,
			Timestamp:   timestamp,
			Details: &map[string]interface{}{
				"coverage_percentage": 80 + i,
				"tests_passed":        100 + i,
				"tests_failed":        i,
				"duration_seconds":    30 + i,
			},
			Metadata: &map[string]interface{}{
				"ci_job_id":   fmt.Sprintf("job-%s-%d", checkSlug, i),
				"environment": "staging",
				"branch":      "main",
				"commit_sha":  fmt.Sprintf("abc%s%d", checkSlug, i),
			},
		}

		resp, err := reportsClient.SubmitReportWithResponse(context.Background(), report)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
	}
}

// setupComponentReportsTest creates the test environment and returns the API clients
func setupComponentReportsTest(t *testing.T) (*client.ClientWithResponses, *reportsclient.ClientWithResponses) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Clear database before test
	clearDatabase(t)

	// Create config with filesystem source pointing to testdata
	testConfig := TestConfig
	fsConfig := sync.NewFilesystemSourceConfig(getTestDataPath(t), 1*time.Second)
	testConfig.Sync = sync.Config{
		Sources: []sync.SourceConfig{
			sync.NewSourceConfig(fsConfig.GetConfig()),
		},
	}

	// Start server with sync enabled
	stop, err := server.Start(testConfig)
	require.NoError(t, err)
	t.Cleanup(stop)

	// Wait for server to start and sync to complete
	waitForSyncCompletion(t, 30*time.Second)

	// Create API client
	apiClient, err := client.NewClientWithResponses("http://localhost:8080/api/catalog/v1")
	require.NoError(t, err)

	// Create reports client
	reportsClient, err := reportsclient.NewClientWithResponses("http://localhost:8080/api/reports/v1")
	require.NoError(t, err)

	return apiClient, reportsClient
}

// setupTestDataWithExactCounts creates test data with predictable, exact counts
func setupTestDataWithExactCounts(t *testing.T, reportsClient *reportsclient.ClientWithResponses) {
	// Generate exactly 3 reports for each component-check combination
	// This gives us predictable counts for testing
	components := []string{"auth-service", "user-service"}
	checks := []string{"unit-tests", "integration-tests", "security-scan", "performance-tests"}

	for _, component := range components {
		for _, check := range checks {
			// Generate exactly 3 reports per check (2 pass, 1 fail)
			// Create reports with different timestamps: 1 recent, 1 old, 1 very old
			generateComponentReportsWithTimestamps(t, reportsClient, component, check, reportsclient.ReportSubmissionStatusPass, []time.Time{
				time.Now().Add(-1 * time.Hour),  // Recent
				time.Now().Add(-5 * time.Hour),  // Old
				time.Now().Add(-10 * time.Hour), // Very old
			})
			generateComponentReportsWithTimestamps(t, reportsClient, component, check, reportsclient.ReportSubmissionStatusFail, []time.Time{
				time.Now().Add(-2 * time.Hour),  // Recent
				time.Now().Add(-6 * time.Hour),  // Old
				time.Now().Add(-12 * time.Hour), // Very old
			})
		}
	}

	// Wait a bit for the API server to process the data
	time.Sleep(500 * time.Millisecond)
}

func TestComponentReportsBasic(t *testing.T) {
	apiClient, reportsClient := setupComponentReportsTest(t)
	setupTestDataWithExactCounts(t, reportsClient)

	// Test basic retrieval without filters
	resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	response := resp.JSON200
	require.NotNil(t, response)
	require.NotNil(t, response.Reports)

	// We should have exactly 24 reports for auth-service (4 checks × 3 reports each)
	require.Equal(t, 24, len(response.Reports))
	require.Equal(t, 24, response.Pagination.Total)
	assert.False(t, response.Pagination.HasMore)
}

func TestComponentReportsWithStatusFilter(t *testing.T) {
	apiClient, reportsClient := setupComponentReportsTest(t)
	setupTestDataWithExactCounts(t, reportsClient)

	// Test filtering by status
	resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{
		Status: utils.ToPointer(client.GetComponentReportsParamsStatusPass),
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	response := resp.JSON200
	require.NotNil(t, response)

	// We should have exactly 12 pass reports for auth-service (4 checks × 3 pass reports each)
	require.Equal(t, 12, len(response.Reports))
	require.Equal(t, 12, response.Pagination.Total)

	// Verify all reports have pass status
	for _, report := range response.Reports {
		require.Equal(t, client.CheckReportStatusPass, report.Status)
	}
}

func TestComponentReportsWithCheckFilter(t *testing.T) {
	apiClient, reportsClient := setupComponentReportsTest(t)
	setupTestDataWithExactCounts(t, reportsClient)

	// Test filtering by check slug
	resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{
		CheckSlug: utils.ToPointer("unit-tests"),
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	response := resp.JSON200
	require.NotNil(t, response)

	// We should have exactly 6 reports for unit-tests (2 components × 3 reports each)
	require.Equal(t, 6, len(response.Reports))
	require.Equal(t, 6, response.Pagination.Total)

	// Verify all reports are for unit-tests
	for _, report := range response.Reports {
		require.Equal(t, "unit-tests", report.CheckSlug)
	}
}

func TestComponentReportsWithStatusAndCheckFilter(t *testing.T) {
	apiClient, reportsClient := setupComponentReportsTest(t)
	setupTestDataWithExactCounts(t, reportsClient)

	// Test filtering by both status and check
	resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{
		Status:    utils.ToPointer(client.GetComponentReportsParamsStatusPass),
		CheckSlug: utils.ToPointer("unit-tests"),
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	response := resp.JSON200
	require.NotNil(t, response)

	// We should have exactly 3 pass reports for unit-tests (1 component × 3 pass reports)
	require.Equal(t, 3, len(response.Reports))
	require.Equal(t, 3, response.Pagination.Total)

	// Verify all reports are pass status and unit-tests
	for _, report := range response.Reports {
		require.Equal(t, client.CheckReportStatusPass, report.Status)
		require.Equal(t, "unit-tests", report.CheckSlug)
	}
}

func TestComponentReportsWithSinceFilter(t *testing.T) {
	apiClient, reportsClient := setupComponentReportsTest(t)
	setupTestDataWithExactCounts(t, reportsClient)

	// Test filtering by since timestamp
	since := time.Now().Add(-2 * time.Hour)
	resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{
		Since: &since,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	response := resp.JSON200
	require.NotNil(t, response)
	// Should have fewer reports (only recent ones)
	require.Less(t, len(response.Reports), 12)
	require.Less(t, response.Pagination.Total, 12)

	// Verify all reports are after the since timestamp
	for _, report := range response.Reports {
		require.True(t, report.Timestamp.After(since))
	}
}

func TestComponentReportsWithPagination(t *testing.T) {
	apiClient, reportsClient := setupComponentReportsTest(t)
	setupTestDataWithExactCounts(t, reportsClient)

	// Test pagination with limit=5
	limit := 5
	resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{
		Limit: &limit,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	response := *resp.JSON200
	// Should have exactly 5 reports
	assert.Len(t, response.Reports, 5)
	assert.Equal(t, 24, response.Pagination.Total)
	assert.Equal(t, 5, response.Pagination.Limit)
	assert.Equal(t, 0, response.Pagination.Offset)
	assert.True(t, response.Pagination.HasMore)

	// Test second page
	offset := 5
	resp2, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{
		Limit:  &limit,
		Offset: &offset,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp2.StatusCode())

	response2 := *resp2.JSON200
	// Should have 5 more reports
	assert.Len(t, response2.Reports, 5)
	assert.Equal(t, 24, response2.Pagination.Total)
	assert.Equal(t, 5, response2.Pagination.Limit)
	assert.Equal(t, 5, response2.Pagination.Offset)
	assert.True(t, response2.Pagination.HasMore)
}

func TestComponentReportsWithLatestPerCheck(t *testing.T) {
	apiClient, reportsClient := setupComponentReportsTest(t)
	setupTestDataWithExactCounts(t, reportsClient)

	// Test latest_per_check=true
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
}

func TestComponentReportsWithLatestPerCheckAndPagination(t *testing.T) {
	apiClient, reportsClient := setupComponentReportsTest(t)
	setupTestDataWithExactCounts(t, reportsClient)

	// Test latest_per_check=true with pagination
	latestPerCheck := true
	limit := 2
	resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{
		LatestPerCheck: &latestPerCheck,
		Limit:          &limit,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	response := *resp.JSON200
	// Should have exactly 2 reports (limited by pagination)
	assert.Len(t, response.Reports, 2)
	assert.Equal(t, 4, response.Pagination.Total) // total should be 4 (one per check)
	assert.Equal(t, 2, response.Pagination.Limit)
	assert.Equal(t, 0, response.Pagination.Offset)
	assert.True(t, response.Pagination.HasMore)

	// Test second page
	offset := 2
	resp2, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{
		LatestPerCheck: &latestPerCheck,
		Limit:          &limit,
		Offset:         &offset,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp2.StatusCode())

	response2 := *resp2.JSON200
	// Should have 2 more reports
	assert.Len(t, response2.Reports, 2)
	assert.Equal(t, 4, response2.Pagination.Total)
	assert.Equal(t, 2, response2.Pagination.Limit)
	assert.Equal(t, 2, response2.Pagination.Offset)
	assert.False(t, response2.Pagination.HasMore)
}

func TestComponentReportsWithLatestPerCheckAndFilters(t *testing.T) {
	apiClient, reportsClient := setupComponentReportsTest(t)
	setupTestDataWithExactCounts(t, reportsClient)

	// Test latest_per_check=true with status filter
	latestPerCheck := true
	status := client.GetComponentReportsParamsStatusPass
	resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{
		LatestPerCheck: &latestPerCheck,
		Status:         &status,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	response := *resp.JSON200
	// Should have exactly 4 reports (one per check type, all with pass status)
	assert.Len(t, response.Reports, 4)
	assert.Equal(t, 4, response.Pagination.Total)

	// Verify all reports have pass status
	for _, report := range response.Reports {
		assert.Equal(t, client.CheckReportStatusPass, report.Status)
	}

	// Test with check slug filter
	checkSlug := "unit-tests"
	resp2, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{
		LatestPerCheck: &latestPerCheck,
		CheckSlug:      &checkSlug,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp2.StatusCode())

	response2 := *resp2.JSON200
	// Should have exactly 1 report (latest for unit-tests only)
	assert.Len(t, response2.Reports, 1)
	assert.Equal(t, 1, response2.Pagination.Total)

	// Verify it's for unit-tests
	assert.Equal(t, "unit-tests", response2.Reports[0].CheckSlug)
}

func TestComponentReportsWithLatestPerCheckPaginationAndFilters(t *testing.T) {
	apiClient, reportsClient := setupComponentReportsTest(t)
	setupTestDataWithExactCounts(t, reportsClient)

	// Test latest_per_check=true with pagination and filters combined
	latestPerCheck := true
	status := client.GetComponentReportsParamsStatusPass
	limit := 2
	resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{
		LatestPerCheck: &latestPerCheck,
		Status:         &status,
		Limit:          &limit,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	response := *resp.JSON200
	// Should have exactly 2 reports (limited by pagination)
	assert.Len(t, response.Reports, 2)
	assert.Equal(t, 4, response.Pagination.Total) // total should be 4 (one per check with pass status)
	assert.Equal(t, 2, response.Pagination.Limit)
	assert.Equal(t, 0, response.Pagination.Offset)
	assert.True(t, response.Pagination.HasMore)

	// Verify all reports have pass status
	for _, report := range response.Reports {
		assert.Equal(t, client.CheckReportStatusPass, report.Status)
	}
}

func TestComponentReportsComponentNotFound(t *testing.T) {
	apiClient, _ := setupComponentReportsTest(t)

	// Test component not found
	resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), "non-existent-component", &client.GetComponentReportsParams{})
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode())
}

func TestComponentReportsForDifferentComponents(t *testing.T) {
	apiClient, reportsClient := setupComponentReportsTest(t)
	setupTestDataWithExactCounts(t, reportsClient)

	// Test getting reports for different components
	resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), "user-service", &client.GetComponentReportsParams{})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	response := *resp.JSON200
	// Should have reports for user-service
	assert.Len(t, response.Reports, 24)
	assert.Equal(t, 24, response.Pagination.Total)
}

func TestComponentReportsWithMaxLimit(t *testing.T) {
	apiClient, reportsClient := setupComponentReportsTest(t)
	setupTestDataWithExactCounts(t, reportsClient)

	// Test with maximum limit
	limit := 100
	resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{
		Limit: &limit,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	response := *resp.JSON200
	assert.Equal(t, 100, response.Pagination.Limit)
}

func TestComponentReportsWithInvalidLimit(t *testing.T) {
	apiClient, reportsClient := setupComponentReportsTest(t)
	setupTestDataWithExactCounts(t, reportsClient)

	// Test with invalid limit (should use default)
	limit := 150 // exceeds max
	resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{
		Limit: &limit,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	response := *resp.JSON200
	assert.Equal(t, 50, response.Pagination.Limit) // should use default
}

// TestLargeDataset tests the API with a large dataset (100 reports for 2 components)
func TestLargeDataset(t *testing.T) {
	apiClient, reportsClient := setupComponentReportsTest(t)

	// Generate 100 reports for 2 components with different check types
	components := []string{"auth-service", "user-service"}
	checks := []string{"unit-tests", "integration-tests", "security-scan", "performance-tests"}

	for _, component := range components {
		for _, check := range checks {
			// Generate 25 reports per check (4 checks * 25 = 100 total per component)
			generateComponentReports(t, reportsClient, component, check, reportsclient.ReportSubmissionStatusPass, 25)
		}
	}

	// Test various filter combinations
	// Note: We'll check that we get the expected number of reports but not hardcode exact counts
	testCases := []struct {
		name        string
		componentID string
		status      *client.GetComponentReportsParamsStatus
		checkSlug   *string
		limit       *int
		offset      *int
		minExpected int
		maxExpected int
	}{
		{
			name:        "AllReportsForComponent",
			componentID: "auth-service",
			minExpected: 50,  // default limit
			maxExpected: 150, // reasonable upper bound
		},
		{
			name:        "FilterByStatus",
			componentID: "auth-service",
			status:      utils.ToPointer(client.GetComponentReportsParamsStatusPass),
			minExpected: 50,
			maxExpected: 150,
		},
		{
			name:        "FilterByCheckSlug",
			componentID: "auth-service",
			checkSlug:   utils.ToPointer("unit-tests"),
			minExpected: 25,
			maxExpected: 50,
		},
		{
			name:        "CombinedFilters",
			componentID: "auth-service",
			status:      utils.ToPointer(client.GetComponentReportsParamsStatusPass),
			checkSlug:   utils.ToPointer("integration-tests"),
			minExpected: 25,
			maxExpected: 50,
		},
		{
			name:        "PaginationFirstPage",
			componentID: "auth-service",
			limit:       utils.ToPointer(10),
			offset:      utils.ToPointer(0),
			minExpected: 10,
			maxExpected: 10,
		},
		{
			name:        "PaginationSecondPage",
			componentID: "auth-service",
			limit:       utils.ToPointer(10),
			offset:      utils.ToPointer(10),
			minExpected: 10,
			maxExpected: 10,
		},
		{
			name:        "MaxLimit",
			componentID: "auth-service",
			limit:       utils.ToPointer(100),
			minExpected: 100,
			maxExpected: 150,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), tc.componentID, &client.GetComponentReportsParams{
				Status:    tc.status,
				CheckSlug: tc.checkSlug,
				Limit:     tc.limit,
				Offset:    tc.offset,
			})
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode())

			response := resp.JSON200
			require.NotNil(t, response)

			// Check that we got a reasonable number of reports on this page
			assert.GreaterOrEqual(t, len(response.Reports), tc.minExpected)
			assert.LessOrEqual(t, len(response.Reports), tc.maxExpected)

			// For pagination tests, check that total count is much larger than page count
			if tc.limit != nil && *tc.limit < 50 {
				// This is a pagination test - total should be much larger than page size
				assert.Greater(t, response.Pagination.Total, len(response.Reports))
				assert.GreaterOrEqual(t, response.Pagination.Total, 100) // Should have many reports
			} else {
				// This is not a pagination test - total should be reasonable
				assert.GreaterOrEqual(t, response.Pagination.Total, tc.minExpected)
				assert.LessOrEqual(t, response.Pagination.Total, tc.maxExpected)
			}
		})
	}
}
