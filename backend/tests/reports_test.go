package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/doron-cohen/argus/backend/internal/utils"
	reportsclient "github.com/doron-cohen/argus/backend/reports/api/client"
	"github.com/doron-cohen/argus/backend/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReportsAPIEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Clear database before test
	clearDatabase(t)

	// Set up test config with sync enabled
	testConfig := TestConfig
	fsConfig := sync.NewFilesystemSourceConfig(getTestDataPath(t), 1*time.Second)
	testConfig.Sync = sync.Config{
		Sources: []sync.SourceConfig{
			sync.NewSourceConfig(fsConfig.GetConfig()),
		},
	}

	// Start server with sync enabled
	stop := startServerAndWaitForHealth(t, testConfig)
	defer stop()

	// Wait for initial sync to complete (components need to be created)
	waitForSyncCompletion(t, 10*time.Second)

	// Create API client
	client, err := reportsclient.NewClientWithResponses("http://localhost:8080/reports")
	require.NoError(t, err)

	// Test submitting a valid report
	t.Run("SubmitValidReport", func(t *testing.T) {
		report := reportsclient.ReportSubmission{
			Check: reportsclient.Check{
				Slug:        "unit-tests",
				Name:        utils.ToPointer("Unit Tests"),
				Description: utils.ToPointer("Runs unit tests for the component"),
			},
			ComponentId: "auth-service",
			Status:      reportsclient.ReportSubmissionStatusPass,
			Timestamp:   time.Now(),
			Details: &map[string]interface{}{
				"coverage_percentage": 85,
				"tests_passed":        100,
				"tests_failed":        0,
				"duration_seconds":    30,
			},
			Metadata: &map[string]interface{}{
				"ci_job_id":   "job-123",
				"environment": "staging",
				"branch":      "main",
				"commit_sha":  "abc123",
			},
		}

		resp, err := client.SubmitReportWithResponse(context.Background(), report)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
	})

	t.Run("SubmitReportWithDifferentStatuses", func(t *testing.T) {
		statuses := []reportsclient.ReportSubmissionStatus{
			reportsclient.ReportSubmissionStatusPass,
			reportsclient.ReportSubmissionStatusFail,
			reportsclient.ReportSubmissionStatusDisabled,
			reportsclient.ReportSubmissionStatusSkipped,
			reportsclient.ReportSubmissionStatusUnknown,
			reportsclient.ReportSubmissionStatusError,
			reportsclient.ReportSubmissionStatusCompleted,
		}

		for _, status := range statuses {
			t.Run(string(status), func(t *testing.T) {
				report := reportsclient.ReportSubmission{
					Check: reportsclient.Check{
						Slug: "build",
					},
					ComponentId: "auth-service",
					Status:      status,
					Timestamp:   time.Now().Add(-30 * time.Minute),
				}

				resp, err := client.SubmitReportWithResponse(context.Background(), report)
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode())
			})
		}
	})

	t.Run("SubmitReportWithDifferentCheckSlugs", func(t *testing.T) {
		checkSlugs := []string{
			"unit-tests",
			"build",
			"linter",
			"security_scan",
			"integration-tests",
			"e2e_tests",
			"coverage",
		}

		for _, slug := range checkSlugs {
			t.Run(slug, func(t *testing.T) {
				report := reportsclient.ReportSubmission{
					Check: reportsclient.Check{
						Slug: slug,
					},
					ComponentId: "auth-service",
					Status:      reportsclient.ReportSubmissionStatusPass,
					Timestamp:   time.Now().Add(-15 * time.Minute),
				}

				resp, err := client.SubmitReportWithResponse(context.Background(), report)
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode())
			})
		}
	})

	t.Run("SubmitReportWithMinimalData", func(t *testing.T) {
		// Submit report with only required fields
		report := reportsclient.ReportSubmission{
			Check: reportsclient.Check{
				Slug: "unit-tests",
			},
			ComponentId: "auth-service",
			Status:      reportsclient.ReportSubmissionStatusPass,
			Timestamp:   time.Now().Add(-5 * time.Minute),
		}

		resp, err := client.SubmitReportWithResponse(context.Background(), report)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode())
	})

	t.Run("SubmitReportWithComplexDetails", func(t *testing.T) {
		// Submit report with complex nested details
		report := reportsclient.ReportSubmission{
			Check: reportsclient.Check{
				Slug: "integration-tests",
			},
			ComponentId: "auth-service",
			Status:      reportsclient.ReportSubmissionStatusPass,
			Timestamp:   time.Now().Add(-10 * time.Minute),
			Details: &map[string]interface{}{
				"test_suites": map[string]interface{}{
					"authentication": map[string]interface{}{
						"passed":  15,
						"failed":  0,
						"skipped": 2,
					},
					"authorization": map[string]interface{}{
						"passed":  8,
						"failed":  1,
						"skipped": 0,
					},
				},
				"coverage": map[string]interface{}{
					"lines":     85.5,
					"functions": 92.1,
					"branches":  78.3,
				},
			},
		}

		resp, err := client.SubmitReportWithResponse(context.Background(), report)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode())
	})
}

func TestReportsAPIValidationErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Start server
	stop := startServerAndWaitForHealth(t, TestConfig)
	defer stop()

	// Create reports API client
	reportsClient, err := reportsclient.NewClientWithResponses("http://localhost:8080/reports")
	require.NoError(t, err)

	t.Run("MissingRequiredFields", func(t *testing.T) {
		testCases := []struct {
			name   string
			report reportsclient.ReportSubmission
			field  string
		}{
			{
				name: "missing_check_slug",
				report: reportsclient.ReportSubmission{
					Check: reportsclient.Check{
						Slug: "", // empty slug
					},
					ComponentId: "auth-service",
					Status:      reportsclient.ReportSubmissionStatusPass,
					Timestamp:   time.Now().Add(-1 * time.Hour),
				},
				field: "check.slug",
			},
			{
				name: "missing_component_id",
				report: reportsclient.ReportSubmission{
					Check: reportsclient.Check{
						Slug: "unit-tests",
					},
					ComponentId: "",
					Status:      reportsclient.ReportSubmissionStatusPass,
					Timestamp:   time.Now().Add(-1 * time.Hour),
				},
				field: "component_id",
			},
			{
				name: "missing_timestamp",
				report: reportsclient.ReportSubmission{
					Check: reportsclient.Check{
						Slug: "unit-tests",
					},
					ComponentId: "auth-service",
					Status:      reportsclient.ReportSubmissionStatusPass,
					Timestamp:   time.Time{}, // zero time
				},
				field: "timestamp",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := reportsClient.SubmitReportWithResponse(context.Background(), tc.report)
				require.NoError(t, err)
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
				require.NotNil(t, resp.JSON400)
				assert.Contains(t, *resp.JSON400.Error, "required")
			})
		}
	})

	t.Run("InvalidCheckSlug", func(t *testing.T) {
		testCases := []struct {
			name      string
			checkSlug string
		}{
			{"empty_slug", ""},
			{"whitespace_slug", "   "},
			{"invalid_chars", "unit-tests@"},
			{"with_spaces", "unit tests"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				report := reportsclient.ReportSubmission{
					Check: reportsclient.Check{
						Slug: tc.checkSlug,
					},
					ComponentId: "auth-service",
					Status:      reportsclient.ReportSubmissionStatusPass,
					Timestamp:   time.Now().Add(-1 * time.Hour),
				}

				resp, err := reportsClient.SubmitReportWithResponse(context.Background(), report)
				require.NoError(t, err)
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
			})
		}
	})

	t.Run("InvalidComponentId", func(t *testing.T) {
		testCases := []struct {
			name        string
			componentId string
		}{
			{"empty_id", ""},
			{"whitespace_id", "   "},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				report := reportsclient.ReportSubmission{
					Check: reportsclient.Check{
						Slug: "unit-tests",
					},
					ComponentId: tc.componentId,
					Status:      reportsclient.ReportSubmissionStatusPass,
					Timestamp:   time.Now().Add(-1 * time.Hour),
				}

				resp, err := reportsClient.SubmitReportWithResponse(context.Background(), report)
				require.NoError(t, err)
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
			})
		}
	})

	t.Run("InvalidStatus", func(t *testing.T) {
		report := reportsclient.ReportSubmission{
			Check: reportsclient.Check{
				Slug: "unit-tests",
			},
			ComponentId: "auth-service",
			Status:      "invalid-status",
			Timestamp:   time.Now().Add(-1 * time.Hour),
		}

		resp, err := reportsClient.SubmitReportWithResponse(context.Background(), report)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
		require.NotNil(t, resp.JSON400)
		assert.Contains(t, *resp.JSON400.Error, "status must be one of")
	})

	t.Run("FutureTimestamp", func(t *testing.T) {
		report := reportsclient.ReportSubmission{
			Check: reportsclient.Check{
				Slug: "unit-tests",
			},
			ComponentId: "auth-service",
			Status:      reportsclient.ReportSubmissionStatusPass,
			Timestamp:   time.Now().Add(10 * time.Minute), // 10 minutes in future
		}

		resp, err := reportsClient.SubmitReportWithResponse(context.Background(), report)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
		require.NotNil(t, resp.JSON400)
		assert.Contains(t, *resp.JSON400.Error, "timestamp cannot be in the future")
	})
}

func TestReportsAPI_InvalidRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Start server
	stop := startServerAndWaitForHealth(t, TestConfig)
	defer stop()

	t.Run("InvalidJSON", func(t *testing.T) {
		// Test with invalid JSON
		resp, err := http.Post("http://localhost:8080/reports/reports", "application/json", nil)
		require.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("Failed to close response body: %v", err)
			}
		}()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errorResp struct {
			Error string `json:"error"`
		}
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Error, "Invalid JSON format")
	})

	t.Run("WrongContentType", func(t *testing.T) {
		// Test with wrong content type
		resp, err := http.Post("http://localhost:8080/reports/reports", "text/plain", nil)
		require.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("Failed to close response body: %v", err)
			}
		}()

		// Should still work as the handler checks content type
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
