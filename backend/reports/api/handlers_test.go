package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/doron-cohen/argus/backend/internal/storage"
	"github.com/doron-cohen/argus/backend/internal/utils"
	reportsclient "github.com/doron-cohen/argus/backend/reports/api/client"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type MockRepository struct {
	*storage.Repository
}

func NewMockRepository(t *testing.T) *MockRepository {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	repo := &storage.Repository{DB: db}
	// Run migrations to create tables
	if err := repo.Migrate(context.Background()); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return &MockRepository{Repository: repo}
}

func TestSubmitReport_Success(t *testing.T) {
	// Create a test database using SQLite
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	// Create repository and migrate
	repo := &storage.Repository{DB: db}
	if err := repo.Migrate(context.Background()); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Create API server
	server := NewAPIServer(repo)

	// Create test component first
	component := storage.Component{
		ComponentID: "auth-service",
		Name:        "Auth Service",
		Description: "Authentication service",
	}
	if err := repo.CreateComponent(context.Background(), component); err != nil {
		t.Fatalf("Failed to create test component: %v", err)
	}

	// Create test request
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

	// Create HTTP request
	body, err := json.Marshal(report)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/reports", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call the handler
	server.SubmitReport(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response reportsclient.ReportSubmissionResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Report submitted successfully", *response.Message)
	assert.NotNil(t, response.ReportId)
	assert.NotNil(t, response.Timestamp)
}

func TestSubmitReport_MissingRequiredFields(t *testing.T) {
	mockRepo := NewMockRepository(t)
	server := NewAPIServer(mockRepo.Repository)

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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.report)
			req := httptest.NewRequest("POST", "/reports", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.SubmitReport(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestSubmitReport_InvalidJSON(t *testing.T) {
	mockRepo := NewMockRepository(t)
	server := NewAPIServer(mockRepo.Repository)

	// Test with invalid JSON
	req := httptest.NewRequest("POST", "/reports", bytes.NewBufferString(`{"invalid": json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.SubmitReport(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubmitReport_ValidStatuses(t *testing.T) {
	mockRepo := NewMockRepository(t)
	server := NewAPIServer(mockRepo.Repository)

	// Create a test component first
	component := storage.Component{
		ComponentID: "auth-service-valid-statuses",
		Name:        "Auth Service",
		Description: "Authentication service",
	}
	if err := mockRepo.CreateComponent(context.Background(), component); err != nil {
		t.Fatalf("Failed to create test component: %v", err)
	}

	validStatuses := []reportsclient.ReportSubmissionStatus{
		reportsclient.ReportSubmissionStatusPass,
		reportsclient.ReportSubmissionStatusFail,
		reportsclient.ReportSubmissionStatusDisabled,
		reportsclient.ReportSubmissionStatusSkipped,
		reportsclient.ReportSubmissionStatusUnknown,
		reportsclient.ReportSubmissionStatusError,
		reportsclient.ReportSubmissionStatusCompleted,
	}

	for _, status := range validStatuses {
		t.Run(string(status), func(t *testing.T) {
			report := reportsclient.ReportSubmission{
				Check: reportsclient.Check{
					Slug:        "unit-tests",
					Name:        utils.ToPointer("Unit Tests"),
					Description: utils.ToPointer("Runs unit tests"),
				},
				ComponentId: "auth-service-valid-statuses",
				Status:      status,
				Timestamp:   time.Now(),
				Details: &map[string]interface{}{
					"coverage_percentage": 85,
				},
				Metadata: &map[string]interface{}{
					"ci_job_id": "job-123",
				},
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

func TestSubmitReport_ValidationErrors(t *testing.T) {
	mockRepo := NewMockRepository(t)
	server := NewAPIServer(mockRepo.Repository)

	// Create a test component for valid cases
	component := storage.Component{
		ComponentID: "auth-service-validation",
		Name:        "Auth Service",
		Description: "Authentication service",
	}
	if err := mockRepo.CreateComponent(context.Background(), component); err != nil {
		t.Fatalf("Failed to create test component: %v", err)
	}

	testCases := []struct {
		name           string
		report         reportsclient.ReportSubmission
		expectedStatus int
		expectedError  string
	}{
		{
			name: "empty_check_slug",
			report: reportsclient.ReportSubmission{
				Check: reportsclient.Check{
					Slug: "",
				},
				ComponentId: "auth-service-validation",
				Status:      reportsclient.ReportSubmissionStatusPass,
				Timestamp:   time.Now(),
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "check slug is required",
		},
		{
			name: "empty_component_id",
			report: reportsclient.ReportSubmission{
				Check: reportsclient.Check{
					Slug: "unit-tests",
				},
				ComponentId: "",
				Status:      reportsclient.ReportSubmissionStatusPass,
				Timestamp:   time.Now(),
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "component ID is required",
		},
		{
			name: "zero_timestamp",
			report: reportsclient.ReportSubmission{
				Check: reportsclient.Check{
					Slug: "unit-tests",
				},
				ComponentId: "auth-service-validation",
				Status:      reportsclient.ReportSubmissionStatusPass,
				Timestamp:   time.Time{}, // zero time
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "timestamp is required",
		},
		{
			name: "future_timestamp",
			report: reportsclient.ReportSubmission{
				Check: reportsclient.Check{
					Slug: "unit-tests",
				},
				ComponentId: "auth-service-validation",
				Status:      reportsclient.ReportSubmissionStatusPass,
				Timestamp:   time.Now().Add(1 * time.Hour), // future time
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "timestamp cannot be in the future",
		},
		{
			name: "check_slug_with_spaces",
			report: reportsclient.ReportSubmission{
				Check: reportsclient.Check{
					Slug: "unit tests",
				},
				ComponentId: "auth-service-validation",
				Status:      reportsclient.ReportSubmissionStatusPass,
				Timestamp:   time.Now(),
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "check slug can only contain alphanumeric characters, hyphens, and underscores",
		},
		{
			name: "check_slug_with_invalid_chars",
			report: reportsclient.ReportSubmission{
				Check: reportsclient.Check{
					Slug: "unit-tests@",
				},
				ComponentId: "auth-service-validation",
				Status:      reportsclient.ReportSubmissionStatusPass,
				Timestamp:   time.Now(),
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "check slug can only contain alphanumeric characters, hyphens, and underscores",
		},
		{
			name: "component_id_with_whitespace",
			report: reportsclient.ReportSubmission{
				Check: reportsclient.Check{
					Slug: "unit-tests",
				},
				ComponentId: " auth-service-validation ",
				Status:      reportsclient.ReportSubmissionStatusPass,
				Timestamp:   time.Now(),
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "component ID cannot have leading or trailing whitespace",
		},
		{
			name: "invalid_status",
			report: reportsclient.ReportSubmission{
				Check: reportsclient.Check{
					Slug: "unit-tests",
				},
				ComponentId: "auth-service-validation",
				Status:      "invalid-status",
				Timestamp:   time.Now(),
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "status must be one of: pass, fail, disabled, skipped, unknown, error, completed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.report)
			req := httptest.NewRequest("POST", "/reports", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.SubmitReport(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			// Check error response format
			var errorResp reportsclient.Error
			err := json.Unmarshal(w.Body.Bytes(), &errorResp)
			require.NoError(t, err)
			assert.Equal(t, "VALIDATION_ERROR", *errorResp.Code)
			assert.Contains(t, *errorResp.Error, tc.expectedError)
		})
	}
}

func TestSubmitReport_ComponentNotFound(t *testing.T) {
	mockRepo := NewMockRepository(t)
	server := NewAPIServer(mockRepo.Repository)

	// Test with non-existent component
	report := reportsclient.ReportSubmission{
		Check: reportsclient.Check{
			Slug:        "unit-tests",
			Name:        utils.ToPointer("Unit Tests"),
			Description: utils.ToPointer("Runs unit tests"),
		},
		ComponentId: "non-existent-component",
		Status:      reportsclient.ReportSubmissionStatusPass,
		Timestamp:   time.Now(),
	}

	body, _ := json.Marshal(report)
	req := httptest.NewRequest("POST", "/reports", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.SubmitReport(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	// Check error response format
	var errorResp reportsclient.Error
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "NOT_FOUND", *errorResp.Code)
	assert.Equal(t, "Component not found", *errorResp.Error)
}
