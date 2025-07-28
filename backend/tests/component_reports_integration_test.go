package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/doron-cohen/argus/backend/api/client"
	"github.com/doron-cohen/argus/backend/internal/server"
	"github.com/doron-cohen/argus/backend/internal/storage"
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

func TestComponentReportsAPIEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Start server
	stop, err := server.Start(TestConfig)
	require.NoError(t, err)
	defer stop()

	// Wait for server to start and be healthy
	time.Sleep(1 * time.Second)

	// Create API clients
	apiClient, err := client.NewClientWithResponses("http://localhost:8080/api/catalog/v1")
	require.NoError(t, err)

	reportsClient, err := reportsclient.NewClientWithResponses("http://localhost:8080/reports")
	require.NoError(t, err)

	// Setup test data
	setupTestData(t, reportsClient)

	// Run test scenarios
	runBasicTests(t, apiClient)
	runFilterTests(t, apiClient)
	runPaginationTests(t, apiClient)
	runLatestPerCheckTests(t, apiClient)
	runErrorTests(t, apiClient)
	runLargeDatasetTests(t, apiClient, reportsClient)
}

// setupTestData creates test components and reports
func setupTestData(t *testing.T, reportsClient *reportsclient.ClientWithResponses) {
	t.Run("SetupTestData", func(t *testing.T) {
		// Clear database before test
		clearDatabase(t)

		// Sync components from testdata
		fsConfig := sync.NewFilesystemSourceConfig(getTestDataPath(t), 1*time.Second)
		syncConfig := sync.Config{
			Sources: []sync.SourceConfig{
				sync.NewSourceConfig(fsConfig.GetConfig()),
			},
		}

		// Create a proper repository for sync
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			TestConfig.Storage.Host,
			TestConfig.Storage.Port,
			TestConfig.Storage.User,
			TestConfig.Storage.Password,
			TestConfig.Storage.DBName,
			TestConfig.Storage.SSLMode,
		)
		repo, err := storage.ConnectAndMigrate(context.Background(), dsn)
		require.NoError(t, err)
		defer func() {
			sqlDB, _ := repo.DB.DB()
			_ = sqlDB.Close()
		}()

		syncService := sync.NewService(repo, syncConfig)
		err = syncService.SyncSource(context.Background(), syncConfig.Sources[0])
		require.NoError(t, err)

		// Submit reports for different components and checks
		components := []string{"auth-service", "user-service"}
		checks := []string{"unit-tests", "integration-tests", "security-scan", "performance-tests"}

		for i, component := range components {
			for j, check := range checks {
				// Create multiple reports for each check with different timestamps
				for k := 0; k < 3; k++ {
					var status reportsclient.ReportSubmissionStatus
					switch k {
					case 1:
						status = reportsclient.ReportSubmissionStatusFail
					default:
						status = reportsclient.ReportSubmissionStatusPass
					}

					name := fmt.Sprintf("%s Check", check)
					description := fmt.Sprintf("Runs %s for %s", check, component)
					report := reportsclient.ReportSubmission{
						Check: reportsclient.Check{
							Slug:        check,
							Name:        utils.ToPointer(name),
							Description: utils.ToPointer(description),
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
}

// runBasicTests runs basic API tests
func runBasicTests(t *testing.T, apiClient *client.ClientWithResponses) {
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
}

// runFilterTests runs tests for different filter options
func runFilterTests(t *testing.T, apiClient *client.ClientWithResponses) {
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
}

// runPaginationTests runs tests for pagination functionality
func runPaginationTests(t *testing.T, apiClient *client.ClientWithResponses) {
	t.Run("GetComponentReportsWithPagination", func(t *testing.T) {
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
		assert.Equal(t, 12, response.Pagination.Total)
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
		assert.Equal(t, 12, response2.Pagination.Total)
		assert.Equal(t, 5, response2.Pagination.Limit)
		assert.Equal(t, 5, response2.Pagination.Offset)
		assert.True(t, response2.Pagination.HasMore)
	})
}

// runLatestPerCheckTests runs tests for latest_per_check functionality
func runLatestPerCheckTests(t *testing.T, apiClient *client.ClientWithResponses) {
	t.Run("GetComponentReportsWithLatestPerCheck", func(t *testing.T) {
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
	})

	t.Run("GetComponentReportsWithLatestPerCheckAndPagination", func(t *testing.T) {
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
	})

	t.Run("GetComponentReportsWithLatestPerCheckAndFilters", func(t *testing.T) {
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
	})

	t.Run("GetComponentReportsWithLatestPerCheckPaginationAndFilters", func(t *testing.T) {
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
	})
}

// runErrorTests runs tests for error scenarios
func runErrorTests(t *testing.T, apiClient *client.ClientWithResponses) {
	t.Run("GetComponentReportsComponentNotFound", func(t *testing.T) {
		// Test component not found
		resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), "non-existent-component", &client.GetComponentReportsParams{})
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode())
	})

	t.Run("GetComponentReportsForDifferentComponents", func(t *testing.T) {
		// Test getting reports for different components
		resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), "user-service", &client.GetComponentReportsParams{})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())

		response := *resp.JSON200
		// Should have reports for user-service
		assert.Len(t, response.Reports, 12)
		assert.Equal(t, 12, response.Pagination.Total)
	})

	t.Run("GetComponentReportsWithMaxLimit", func(t *testing.T) {
		// Test with maximum limit
		limit := 100
		resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{
			Limit: &limit,
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())

		response := *resp.JSON200
		assert.Equal(t, 100, response.Pagination.Limit)
	})

	t.Run("GetComponentReportsWithInvalidLimit", func(t *testing.T) {
		// Test with invalid limit (should use default)
		limit := 150 // exceeds max
		resp, err := apiClient.GetComponentReportsWithResponse(context.Background(), "auth-service", &client.GetComponentReportsParams{
			Limit: &limit,
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())

		response := *resp.JSON200
		assert.Equal(t, 50, response.Pagination.Limit) // should use default
	})
}

// runLargeDatasetTests tests the API with a large dataset (100 reports for 2 components)
func runLargeDatasetTests(t *testing.T, apiClient *client.ClientWithResponses, reportsClient *reportsclient.ClientWithResponses) {
	t.Run("LargeDatasetTests", func(t *testing.T) {
		// Clear database and setup fresh data
		clearDatabase(t)

		// Sync components
		fsConfig := sync.NewFilesystemSourceConfig(getTestDataPath(t), 1*time.Second)
		syncConfig := sync.Config{
			Sources: []sync.SourceConfig{
				sync.NewSourceConfig(fsConfig.GetConfig()),
			},
		}

		// Create repository and sync components
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			TestConfig.Storage.Host,
			TestConfig.Storage.Port,
			TestConfig.Storage.User,
			TestConfig.Storage.Password,
			TestConfig.Storage.DBName,
			TestConfig.Storage.SSLMode,
		)
		repo, err := storage.ConnectAndMigrate(context.Background(), dsn)
		require.NoError(t, err)
		defer func() {
			sqlDB, _ := repo.DB.DB()
			_ = sqlDB.Close()
		}()

		syncService := sync.NewService(repo, syncConfig)
		err = syncService.SyncSource(context.Background(), syncConfig.Sources[0])
		require.NoError(t, err)

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
		testCases := []struct {
			name          string
			componentID   string
			status        *client.GetComponentReportsParamsStatus
			checkSlug     *string
			limit         *int
			offset        *int
			expectedCount int
			expectedTotal int
		}{
			{
				name:          "AllReportsForComponent",
				componentID:   "auth-service",
				expectedCount: 50, // default limit
				expectedTotal: 100,
			},
			{
				name:          "FilterByStatus",
				componentID:   "auth-service",
				status:        utils.ToPointer(client.GetComponentReportsParamsStatusPass),
				expectedCount: 50,
				expectedTotal: 100,
			},
			{
				name:          "FilterByCheckSlug",
				componentID:   "auth-service",
				checkSlug:     utils.ToPointer("unit-tests"),
				expectedCount: 25,
				expectedTotal: 25,
			},
			{
				name:          "CombinedFilters",
				componentID:   "auth-service",
				status:        utils.ToPointer(client.GetComponentReportsParamsStatusPass),
				checkSlug:     utils.ToPointer("integration-tests"),
				expectedCount: 25,
				expectedTotal: 25,
			},
			{
				name:          "PaginationFirstPage",
				componentID:   "auth-service",
				limit:         utils.ToPointer(10),
				offset:        utils.ToPointer(0),
				expectedCount: 10,
				expectedTotal: 100,
			},
			{
				name:          "PaginationSecondPage",
				componentID:   "auth-service",
				limit:         utils.ToPointer(10),
				offset:        utils.ToPointer(10),
				expectedCount: 10,
				expectedTotal: 100,
			},
			{
				name:          "MaxLimit",
				componentID:   "auth-service",
				limit:         utils.ToPointer(100),
				expectedCount: 100,
				expectedTotal: 100,
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
				assert.Len(t, response.Reports, tc.expectedCount)
				assert.Equal(t, tc.expectedTotal, response.Pagination.Total)
			})
		}
	})
}
