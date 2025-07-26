package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubmitReport_ValidRequest(t *testing.T) {
	server := NewReportsServer()

	validReport := ReportSubmission{
		Check: Check{
			Slug:        "unit-tests",
			Name:        &[]string{"Unit Tests"}[0],
			Description: &[]string{"Runs unit tests for the component"}[0],
		},
		ComponentId: "auth-service",
		Status:      ReportSubmissionStatusPass,
		Timestamp:   time.Now().Add(-1 * time.Hour),
		Details: &map[string]interface{}{
			"coverage_percentage": 85.5,
			"tests_passed":        150,
			"tests_failed":        0,
		},
		Metadata: &map[string]interface{}{
			"ci_job_id":   "12345",
			"environment": "staging",
		},
	}

	body, _ := json.Marshal(validReport)
	req := httptest.NewRequest("POST", "/reports", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.SubmitReport(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ReportSubmissionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Report submitted successfully", *response.Message)
	assert.NotNil(t, response.ReportId)
	assert.NotNil(t, response.Timestamp)
}

func TestSubmitReport_MissingRequiredFields(t *testing.T) {
	server := NewReportsServer()

	testCases := []struct {
		name   string
		report ReportSubmission
		field  string
	}{
		{
			name: "missing_check_slug",
			report: ReportSubmission{
				Check: Check{
					Slug: "", // empty slug
				},
				ComponentId: "auth-service",
				Status:      ReportSubmissionStatusPass,
				Timestamp:   time.Now().Add(-1 * time.Hour),
			},
			field: "check.slug",
		},
		{
			name: "missing_component_id",
			report: ReportSubmission{
				Check: Check{
					Slug: "unit-tests",
				},
				ComponentId: "",
				Status:      ReportSubmissionStatusPass,
				Timestamp:   time.Now().Add(-1 * time.Hour),
			},
			field: "component_id",
		},
		{
			name: "missing_timestamp",
			report: ReportSubmission{
				Check: Check{
					Slug: "unit-tests",
				},
				ComponentId: "auth-service",
				Status:      ReportSubmissionStatusPass,
				Timestamp:   time.Time{}, // zero time
			},
			field: "timestamp",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.report)
			req := httptest.NewRequest("POST", "/reports", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.SubmitReport(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var errorResponse Error
			err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
			require.NoError(t, err)
			assert.Contains(t, *errorResponse.Error, "required")
		})
	}
}

func TestSubmitReport_InvalidCheckSlug(t *testing.T) {
	server := NewReportsServer()

	testCases := []struct {
		name      string
		checkSlug string
	}{
		{"whitespace_slug", "   "},
		{"invalid_chars", "unit-tests@"},
		{"with_spaces", "unit tests"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			report := ReportSubmission{
				Check: Check{
					Slug: tc.checkSlug,
				},
				ComponentId: "auth-service",
				Status:      ReportSubmissionStatusPass,
				Timestamp:   time.Now().Add(-1 * time.Hour),
			}

			body, _ := json.Marshal(report)
			req := httptest.NewRequest("POST", "/reports", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.SubmitReport(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestSubmitReport_InvalidComponentId(t *testing.T) {
	server := NewReportsServer()

	testCases := []struct {
		name        string
		componentId string
	}{
		{"empty_id", ""},
		{"whitespace_id", "   "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			report := ReportSubmission{
				Check: Check{
					Slug: "unit-tests",
				},
				ComponentId: tc.componentId,
				Status:      ReportSubmissionStatusPass,
				Timestamp:   time.Now().Add(-1 * time.Hour),
			}

			body, _ := json.Marshal(report)
			req := httptest.NewRequest("POST", "/reports", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.SubmitReport(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestSubmitReport_InvalidStatus(t *testing.T) {
	server := NewReportsServer()

	report := ReportSubmission{
		Check: Check{
			Slug: "unit-tests",
		},
		ComponentId: "auth-service",
		Status:      "invalid-status",
		Timestamp:   time.Now().Add(-1 * time.Hour),
	}

	body, _ := json.Marshal(report)
	req := httptest.NewRequest("POST", "/reports", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.SubmitReport(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResponse Error
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Contains(t, *errorResponse.Error, "status must be one of")
}

func TestSubmitReport_InvalidTimestamp(t *testing.T) {
	server := NewReportsServer()

	report := ReportSubmission{
		Check: Check{
			Slug: "unit-tests",
		},
		ComponentId: "auth-service",
		Status:      ReportSubmissionStatusPass,
		Timestamp:   time.Now().Add(10 * time.Minute), // 10 minutes in future
	}

	body, _ := json.Marshal(report)
	req := httptest.NewRequest("POST", "/reports", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.SubmitReport(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResponse Error
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Contains(t, *errorResponse.Error, "timestamp cannot be in the future")
}

func TestSubmitReport_InvalidJSON(t *testing.T) {
	server := NewReportsServer()

	req := httptest.NewRequest("POST", "/reports", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.SubmitReport(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResponse Error
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Contains(t, *errorResponse.Error, "Invalid JSON format")
}

func TestSubmitReport_ValidStatuses(t *testing.T) {
	server := NewReportsServer()

	validStatuses := []ReportSubmissionStatus{
		ReportSubmissionStatusPass,
		ReportSubmissionStatusFail,
		ReportSubmissionStatusDisabled,
		ReportSubmissionStatusSkipped,
		ReportSubmissionStatusUnknown,
		ReportSubmissionStatusError,
		ReportSubmissionStatusCompleted,
	}

	for _, status := range validStatuses {
		t.Run(string(status), func(t *testing.T) {
			report := ReportSubmission{
				Check: Check{
					Slug: "unit-tests",
				},
				ComponentId: "auth-service",
				Status:      status,
				Timestamp:   time.Now().Add(-1 * time.Hour),
			}

			body, _ := json.Marshal(report)
			req := httptest.NewRequest("POST", "/reports", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.SubmitReport(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestSubmitReport_ValidSlugs(t *testing.T) {
	server := NewReportsServer()

	validSlugs := []string{
		"unit-tests",
		"build",
		"linter",
		"security_scan",
		"integration-tests",
		"e2e_tests",
		"coverage",
	}

	for _, slug := range validSlugs {
		t.Run(slug, func(t *testing.T) {
			report := ReportSubmission{
				Check: Check{
					Slug: slug,
				},
				ComponentId: "auth-service",
				Status:      ReportSubmissionStatusPass,
				Timestamp:   time.Now().Add(-1 * time.Hour),
			}

			body, _ := json.Marshal(report)
			req := httptest.NewRequest("POST", "/reports", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.SubmitReport(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}
