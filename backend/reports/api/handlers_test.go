package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSubmitReport_ValidRequest(t *testing.T) {
	server := NewReportsServer()

	validSubmission := ReportSubmission{
		CheckSlug:   "unit-tests",
		ComponentId: "auth-service",
		Status:      ReportSubmissionStatusPass,
		Timestamp:   time.Now().Add(-1 * time.Hour), // 1 hour ago
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

	body, _ := json.Marshal(validSubmission)
	req := httptest.NewRequest("POST", "/reports", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.SubmitReport(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response ReportSubmissionResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Message != "Report submitted successfully" {
		t.Errorf("Expected success message, got %s", response.Message)
	}

	if response.ReportId == nil {
		t.Error("Expected report ID to be generated")
	}

	if response.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestSubmitReport_MissingRequiredFields(t *testing.T) {
	server := NewReportsServer()

	testCases := []struct {
		name          string
		submission    ReportSubmission
		expectedError string
	}{
		{
			name: "missing check_slug",
			submission: ReportSubmission{
				ComponentId: "auth-service",
				Status:      ReportSubmissionStatusPass,
				Timestamp:   time.Now(),
			},
			expectedError: "check_slug is required and cannot be empty",
		},
		{
			name: "missing component_id",
			submission: ReportSubmission{
				CheckSlug: "unit-tests",
				Status:    ReportSubmissionStatusPass,
				Timestamp: time.Now(),
			},
			expectedError: "component_id is required and cannot be empty",
		},
		{
			name: "missing status",
			submission: ReportSubmission{
				CheckSlug:   "unit-tests",
				ComponentId: "auth-service",
				Timestamp:   time.Now(),
			},
			expectedError: "status must be one of: pass, fail, disabled, skipped, unknown, error, completed",
		},
		{
			name: "missing timestamp",
			submission: ReportSubmission{
				CheckSlug:   "unit-tests",
				ComponentId: "auth-service",
				Status:      ReportSubmissionStatusPass,
				Timestamp:   time.Time{}, // zero time
			},
			expectedError: "timestamp is required and cannot be zero",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.submission)
			req := httptest.NewRequest("POST", "/reports", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.SubmitReport(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status 400, got %d", w.Code)
			}

			var errorResponse Error
			if err := json.NewDecoder(w.Body).Decode(&errorResponse); err != nil {
				t.Fatalf("Failed to decode error response: %v", err)
			}

			if errorResponse.Error != tc.expectedError {
				t.Errorf("Expected error '%s', got '%s'", tc.expectedError, errorResponse.Error)
			}
		})
	}
}

func TestSubmitReport_InvalidCheckSlug(t *testing.T) {
	server := NewReportsServer()

	testCases := []struct {
		name          string
		checkSlug     string
		expectedError string
	}{
		{
			name:          "empty check_slug",
			checkSlug:     "",
			expectedError: "check_slug is required and cannot be empty",
		},
		{
			name:          "whitespace only check_slug",
			checkSlug:     "   ",
			expectedError: "check_slug is required and cannot be empty",
		},
		{
			name:          "check_slug too long",
			checkSlug:     string(make([]byte, 101)), // 101 characters
			expectedError: "check_slug cannot exceed 100 characters",
		},
		{
			name:          "check_slug with invalid characters",
			checkSlug:     "unit-tests@",
			expectedError: "check_slug must contain only alphanumeric characters, hyphens, and underscores",
		},
		{
			name:          "check_slug with spaces",
			checkSlug:     "unit tests",
			expectedError: "check_slug must contain only alphanumeric characters, hyphens, and underscores",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			submission := ReportSubmission{
				CheckSlug:   tc.checkSlug,
				ComponentId: "auth-service",
				Status:      ReportSubmissionStatusPass,
				Timestamp:   time.Now().Add(-1 * time.Hour),
			}

			body, _ := json.Marshal(submission)
			req := httptest.NewRequest("POST", "/reports", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.SubmitReport(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status 400, got %d", w.Code)
			}

			var errorResponse Error
			if err := json.NewDecoder(w.Body).Decode(&errorResponse); err != nil {
				t.Fatalf("Failed to decode error response: %v", err)
			}

			if errorResponse.Error != tc.expectedError {
				t.Errorf("Expected error '%s', got '%s'", tc.expectedError, errorResponse.Error)
			}
		})
	}
}

func TestSubmitReport_InvalidComponentId(t *testing.T) {
	server := NewReportsServer()

	testCases := []struct {
		name          string
		componentId   string
		expectedError string
	}{
		{
			name:          "empty component_id",
			componentId:   "",
			expectedError: "component_id is required and cannot be empty",
		},
		{
			name:          "whitespace only component_id",
			componentId:   "   ",
			expectedError: "component_id is required and cannot be empty",
		},
		{
			name:          "component_id too long",
			componentId:   string(make([]byte, 256)), // 256 characters
			expectedError: "component_id cannot exceed 255 characters",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			submission := ReportSubmission{
				CheckSlug:   "unit-tests",
				ComponentId: tc.componentId,
				Status:      ReportSubmissionStatusPass,
				Timestamp:   time.Now().Add(-1 * time.Hour),
			}

			body, _ := json.Marshal(submission)
			req := httptest.NewRequest("POST", "/reports", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.SubmitReport(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status 400, got %d", w.Code)
			}

			var errorResponse Error
			if err := json.NewDecoder(w.Body).Decode(&errorResponse); err != nil {
				t.Fatalf("Failed to decode error response: %v", err)
			}

			if errorResponse.Error != tc.expectedError {
				t.Errorf("Expected error '%s', got '%s'", tc.expectedError, errorResponse.Error)
			}
		})
	}
}

func TestSubmitReport_InvalidStatus(t *testing.T) {
	server := NewReportsServer()

	submission := ReportSubmission{
		CheckSlug:   "unit-tests",
		ComponentId: "auth-service",
		Status:      "invalid-status",
		Timestamp:   time.Now().Add(-1 * time.Hour),
	}

	body, _ := json.Marshal(submission)
	req := httptest.NewRequest("POST", "/reports", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.SubmitReport(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var errorResponse Error
	if err := json.NewDecoder(w.Body).Decode(&errorResponse); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	expectedError := "status must be one of: pass, fail, disabled, skipped, unknown, error, completed"
	if errorResponse.Error != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, errorResponse.Error)
	}
}

func TestSubmitReport_InvalidTimestamp(t *testing.T) {
	server := NewReportsServer()

	// Test future timestamp
	submission := ReportSubmission{
		CheckSlug:   "unit-tests",
		ComponentId: "auth-service",
		Status:      ReportSubmissionStatusPass,
		Timestamp:   time.Now().Add(10 * time.Minute), // 10 minutes in future
	}

	body, _ := json.Marshal(submission)
	req := httptest.NewRequest("POST", "/reports", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.SubmitReport(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var errorResponse Error
	if err := json.NewDecoder(w.Body).Decode(&errorResponse); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	expectedError := "timestamp cannot be in the future"
	if errorResponse.Error != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, errorResponse.Error)
	}
}

func TestSubmitReport_InvalidJSON(t *testing.T) {
	server := NewReportsServer()

	// Send invalid JSON
	req := httptest.NewRequest("POST", "/reports", bytes.NewBufferString(`{"invalid": json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.SubmitReport(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var errorResponse Error
	if err := json.NewDecoder(w.Body).Decode(&errorResponse); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if errorResponse.Error != "Invalid JSON format" {
		t.Errorf("Expected 'Invalid JSON format' error, got '%s'", errorResponse.Error)
	}
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
			submission := ReportSubmission{
				CheckSlug:   "unit-tests",
				ComponentId: "auth-service",
				Status:      status,
				Timestamp:   time.Now().Add(-1 * time.Hour),
			}

			body, _ := json.Marshal(submission)
			req := httptest.NewRequest("POST", "/reports", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.SubmitReport(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200 for status %s, got %d", status, w.Code)
			}
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
		"test123",
		"TEST_CASE",
	}

	for _, slug := range validSlugs {
		t.Run(slug, func(t *testing.T) {
			submission := ReportSubmission{
				CheckSlug:   slug,
				ComponentId: "auth-service",
				Status:      ReportSubmissionStatusPass,
				Timestamp:   time.Now().Add(-1 * time.Hour),
			}

			body, _ := json.Marshal(submission)
			req := httptest.NewRequest("POST", "/reports", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.SubmitReport(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200 for slug %s, got %d", slug, w.Code)
			}
		})
	}
}
