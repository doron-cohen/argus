package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/doron-cohen/argus/backend/internal/server"
	reportsclient "github.com/doron-cohen/argus/backend/reports/api/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReportsAPIEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Start server
	stop, err := server.Start(TestConfig)
	require.NoError(t, err)
	defer stop()

	// Wait for server to start
	time.Sleep(1 * time.Second)

	// Create reports API client
	reportsClient, err := reportsclient.NewClientWithResponses("http://localhost:8080/reports")
	require.NoError(t, err)

	t.Run("SubmitValidReport", func(t *testing.T) {
		// Submit a valid report
		report := reportsclient.ReportSubmission{
			Check: reportsclient.Check{
				Slug:        "unit-tests",
				Name:        &[]string{"Unit Tests"}[0],
				Description: &[]string{"Runs unit tests for the component"}[0],
			},
			ComponentId: "auth-service",
			Status:      reportsclient.ReportSubmissionStatusPass,
			Timestamp:   time.Now().Add(-1 * time.Hour), // 1 hour ago
			Details: &map[string]interface{}{
				"coverage_percentage": 85.5,
				"tests_passed":        150,
				"tests_failed":        0,
				"duration_seconds":    45,
			},
			Metadata: &map[string]interface{}{
				"ci_job_id":   "12345",
				"environment": "staging",
				"branch":      "main",
				"commit_sha":  "abc123",
			},
		}

		resp, err := reportsClient.SubmitReportWithResponse(context.Background(), report)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
		require.NotNil(t, resp.JSON200)

		response := *resp.JSON200
		assert.Equal(t, "Report submitted successfully", *response.Message)
		assert.NotNil(t, response.ReportId)
		assert.NotNil(t, response.Timestamp)
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

				resp, err := reportsClient.SubmitReportWithResponse(context.Background(), report)
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

				resp, err := reportsClient.SubmitReportWithResponse(context.Background(), report)
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

		resp, err := reportsClient.SubmitReportWithResponse(context.Background(), report)
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

		resp, err := reportsClient.SubmitReportWithResponse(context.Background(), report)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode())
	})
}

func TestReportsAPIValidationErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Start server
	stop, err := server.Start(TestConfig)
	require.NoError(t, err)
	defer stop()

	// Wait for server to start
	time.Sleep(1 * time.Second)

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

func TestReportsAPIWithDirectHTTP(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Start server
	stop, err := server.Start(TestConfig)
	require.NoError(t, err)
	defer stop()

	// Wait for server to start
	time.Sleep(1 * time.Second)

	t.Run("InvalidJSON", func(t *testing.T) {
		// Test with invalid JSON
		resp, err := http.Post("http://localhost:8080/reports/reports", "application/json", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

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
		defer resp.Body.Close()

		// Should still work as the handler checks content type
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
