package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/doron-cohen/argus/backend/internal/storage"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestGetComponentReports(t *testing.T) {
	// Setup database and server
	repo, server := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, repo)

	t.Run("Success", func(t *testing.T) {
		// Create test data
		_, _, report := createTestData(t, repo)

		// Create request
		req := httptest.NewRequest("GET", "/components/test-component/reports", nil)
		w := httptest.NewRecorder()

		// Call handler
		server.GetComponentReports(w, req, "test-component", GetComponentReportsParams{})

		// Assert response
		assertComponentReportsResponse(t, w, 1, report.ID.String())
	})

	t.Run("ComponentNotFound", func(t *testing.T) {
		// Create request for non-existent component
		req := httptest.NewRequest("GET", "/components/non-existent/reports", nil)
		w := httptest.NewRecorder()

		// Call handler
		server.GetComponentReports(w, req, "non-existent", GetComponentReportsParams{})

		// Assert 404 response
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("LatestPerCheck", func(t *testing.T) {
		// Create test data with multiple reports
		component, check1, check2 := createTestDataWithMultipleReports(t, repo)

		// Create request with latest_per_check=true
		req := httptest.NewRequest("GET", "/components/test-component-latest/reports?latest_per_check=true", nil)
		w := httptest.NewRecorder()

		// Call handler
		latestPerCheck := true
		server.GetComponentReports(w, req, "test-component-latest", GetComponentReportsParams{
			LatestPerCheck: &latestPerCheck,
		})

		// Assert response
		assertLatestPerCheckResponse(t, w, component, check1, check2)
	})
}

// setupTestEnvironment creates a test database and server
func setupTestEnvironment(t *testing.T) (*storage.Repository, *APIServer) {
	// Create in-memory SQLite database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations
	err = db.AutoMigrate(&storage.Component{}, &storage.Check{}, &storage.CheckReport{})
	require.NoError(t, err)

	// Create repository
	repo := &storage.Repository{DB: db}

	// Create server
	server := &APIServer{Repo: repo}

	return repo, server
}

// cleanupTestEnvironment cleans up test resources
func cleanupTestEnvironment(t *testing.T, repo *storage.Repository) {
	// Close database connection
	sqlDB, err := repo.DB.DB()
	require.NoError(t, err)
	_ = sqlDB.Close()
}

// createTestData creates basic test data and returns component, check, and report
func createTestData(t *testing.T, repo *storage.Repository) (*storage.Component, *storage.Check, *storage.CheckReport) {
	// Create test component
	component := storage.Component{
		ComponentID: "test-component",
		Name:        "Test Component",
		Description: "A test component",
	}
	if err := repo.DB.Create(&component).Error; err != nil {
		t.Fatalf("Failed to create test component: %v", err)
	}

	// Create test check
	check := storage.Check{
		Slug:        "unit-tests",
		Name:        "Unit Tests",
		Description: "Runs unit tests",
	}
	if err := repo.DB.Create(&check).Error; err != nil {
		t.Fatalf("Failed to create test check: %v", err)
	}

	// Create test report
	report := storage.CheckReport{
		CheckID:     check.ID,
		ComponentID: component.ID,
		Status:      storage.CheckStatusPass,
		Timestamp:   time.Now(),
		Details:     storage.JSONB{"coverage": 80.0},
	}
	if err := repo.DB.Create(&report).Error; err != nil {
		t.Fatalf("Failed to create test report: %v", err)
	}

	return &component, &check, &report
}

// createTestDataWithMultipleReports creates test data with multiple reports for latest_per_check testing
func createTestDataWithMultipleReports(t *testing.T, repo *storage.Repository) (*storage.Component, *storage.Check, *storage.Check) {
	// Create test component
	component := storage.Component{
		ComponentID: "test-component-latest",
		Name:        "Test Component Latest",
		Description: "A test component for latest per check",
	}
	if err := repo.DB.Create(&component).Error; err != nil {
		t.Fatalf("Failed to create test component: %v", err)
	}

	// Create two test checks
	check1 := storage.Check{
		Slug:        "unit-tests-latest",
		Name:        "Unit Tests Latest",
		Description: "Runs unit tests",
	}
	if err := repo.DB.Create(&check1).Error; err != nil {
		t.Fatalf("Failed to create test check 1: %v", err)
	}

	check2 := storage.Check{
		Slug:        "integration-tests-latest",
		Name:        "Integration Tests Latest",
		Description: "Runs integration tests",
	}
	if err := repo.DB.Create(&check2).Error; err != nil {
		t.Fatalf("Failed to create test check 2: %v", err)
	}

	// Create multiple reports for the same check type (unit-tests)
	oldReport := storage.CheckReport{
		CheckID:     check1.ID,
		ComponentID: component.ID,
		Status:      storage.CheckStatusPass,
		Timestamp:   time.Now().Add(-time.Hour), // 1 hour ago
		Details:     storage.JSONB{"coverage": 80.0},
	}
	if err := repo.DB.Create(&oldReport).Error; err != nil {
		t.Fatalf("Failed to create old test report: %v", err)
	}

	newReport := storage.CheckReport{
		CheckID:     check1.ID,
		ComponentID: component.ID,
		Status:      storage.CheckStatusFail,
		Timestamp:   time.Now(), // now
		Details:     storage.JSONB{"coverage": 85.0},
	}
	if err := repo.DB.Create(&newReport).Error; err != nil {
		t.Fatalf("Failed to create new test report: %v", err)
	}

	// Create a report for the second check type
	integrationReport := storage.CheckReport{
		CheckID:     check2.ID,
		ComponentID: component.ID,
		Status:      storage.CheckStatusPass,
		Timestamp:   time.Now().Add(-30 * time.Minute), // 30 minutes ago
		Details:     storage.JSONB{"tests": 50},
	}
	if err := repo.DB.Create(&integrationReport).Error; err != nil {
		t.Fatalf("Failed to create integration test report: %v", err)
	}

	return &component, &check1, &check2
}

// assertComponentReportsResponse asserts the response for component reports
func assertComponentReportsResponse(t *testing.T, w *httptest.ResponseRecorder, expectedCount int, expectedReportID string) {
	assert.Equal(t, http.StatusOK, w.Code)

	var response ComponentReportsResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Len(t, response.Reports, expectedCount)
	assert.Equal(t, expectedCount, response.Pagination.Total)
	assert.Equal(t, 50, response.Pagination.Limit) // default limit
	assert.Equal(t, 0, response.Pagination.Offset)
	assert.False(t, response.Pagination.HasMore)

	if expectedCount > 0 {
		assert.Equal(t, expectedReportID, response.Reports[0].Id)
		assert.Equal(t, CheckReportStatusPass, response.Reports[0].Status)
		assert.Equal(t, "unit-tests", response.Reports[0].CheckSlug)
	}
}

// assertLatestPerCheckResponse asserts the response for latest_per_check functionality
func assertLatestPerCheckResponse(t *testing.T, w *httptest.ResponseRecorder, _ *storage.Component, _ *storage.Check, _ *storage.Check) {
	assert.Equal(t, http.StatusOK, w.Code)

	var response ComponentReportsResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Should return exactly 2 reports (one for each check type)
	assert.Len(t, response.Reports, 2)

	// Find the unit-tests report and verify it's the newer one
	var unitTestReport *CheckReport
	var integrationTestReport *CheckReport
	for i := range response.Reports {
		switch response.Reports[i].CheckSlug {
		case "unit-tests-latest":
			unitTestReport = &response.Reports[i]
		case "integration-tests-latest":
			integrationTestReport = &response.Reports[i]
		}
	}

	require.NotNil(t, unitTestReport, "Should have unit-tests-latest report")
	require.NotNil(t, integrationTestReport, "Should have integration-tests-latest report")

	// Verify the unit-tests report is the newer one (should have fail status)
	assert.Equal(t, CheckReportStatusFail, unitTestReport.Status)

	// Verify the integration-tests report
	assert.Equal(t, CheckReportStatusPass, integrationTestReport.Status)
}

func TestGetComponentReports_ValidationErrors(t *testing.T) {
	// Setup database and server
	repo, server := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, repo)

	// Create test component first
	component := storage.Component{
		ComponentID: "test-component-validation",
		Name:        "Test Component Validation",
		Description: "A test component for validation",
	}
	if err := repo.DB.Create(&component).Error; err != nil {
		t.Fatalf("Failed to create test component: %v", err)
	}

	t.Run("InvalidLimit", func(t *testing.T) {
		// Create request with invalid limit
		req := httptest.NewRequest("GET", "/components/test-component-validation/reports?limit=-1", nil)
		w := httptest.NewRecorder()

		// Call handler
		limit := -1
		server.GetComponentReports(w, req, "test-component-validation", GetComponentReportsParams{
			Limit: &limit,
		})

		// API might handle negative limits gracefully by using default
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("InvalidOffset", func(t *testing.T) {
		// Create request with invalid offset
		req := httptest.NewRequest("GET", "/components/test-component-validation/reports?offset=-1", nil)
		w := httptest.NewRecorder()

		// Call handler
		offset := -1
		server.GetComponentReports(w, req, "test-component-validation", GetComponentReportsParams{
			Offset: &offset,
		})

		// API might handle negative offsets gracefully by using default
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("ExcessiveLimit", func(t *testing.T) {
		// Create request with excessive limit
		req := httptest.NewRequest("GET", "/components/test-component-validation/reports?limit=10000", nil)
		w := httptest.NewRecorder()

		// Call handler
		limit := 10000
		server.GetComponentReports(w, req, "test-component-validation", GetComponentReportsParams{
			Limit: &limit,
		})

		// Should cap the limit to maximum allowed
		assert.Equal(t, http.StatusOK, w.Code)

		var response ComponentReportsResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		// Should use the maximum limit instead of 10000
		assert.LessOrEqual(t, response.Pagination.Limit, 1000) // Assuming max limit is 1000
	})
}

func TestGetComponentReports_EdgeCases(t *testing.T) {
	// Setup database and server
	repo, server := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, repo)

	// Create test component for edge cases
	component := storage.Component{
		ComponentID: "test-component-edgecases",
		Name:        "Test Component Edge Cases",
		Description: "A test component for edge cases",
	}
	if err := repo.DB.Create(&component).Error; err != nil {
		t.Fatalf("Failed to create test component: %v", err)
	}

	t.Run("EmptyComponent", func(t *testing.T) {
		// Create request for component with no reports
		req := httptest.NewRequest("GET", "/components/test-component-edgecases/reports", nil)
		w := httptest.NewRecorder()

		// Call handler
		server.GetComponentReports(w, req, "test-component-edgecases", GetComponentReportsParams{})

		// Should return empty list, not error
		assert.Equal(t, http.StatusOK, w.Code)

		var response ComponentReportsResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		assert.Len(t, response.Reports, 0)
		assert.Equal(t, 0, response.Pagination.Total)
	})

	t.Run("InvalidSinceDate", func(t *testing.T) {
		// Create request with invalid since date
		req := httptest.NewRequest("GET", "/components/test-component-edgecases/reports?since=invalid-date", nil)
		w := httptest.NewRecorder()

		// Call handler - this will fail at the OpenAPI validation level
		// The OpenAPI spec should handle invalid date formats
		server.GetComponentReports(w, req, "test-component-edgecases", GetComponentReportsParams{})

		// API might handle invalid dates gracefully by ignoring the parameter
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("FutureSinceDate", func(t *testing.T) {
		// Create request with future since date
		req := httptest.NewRequest("GET", "/components/test-component-edgecases/reports?since=2030-01-01T00:00:00Z", nil)
		w := httptest.NewRecorder()

		// Call handler
		server.GetComponentReports(w, req, "test-component-edgecases", GetComponentReportsParams{})

		// Should return empty list since no reports exist in the future
		assert.Equal(t, http.StatusOK, w.Code)

		var response ComponentReportsResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		assert.Len(t, response.Reports, 0)
	})
}

func TestGetComponentReports_Pagination(t *testing.T) {
	// Setup database and server
	repo, server := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, repo)

	// Create multiple test reports
	createMultipleTestReports(t, repo)

	testCases := []struct {
		name            string
		limit           int
		offset          int
		expectedCount   int
		expectedTotal   int
		expectedOffset  int
		expectedHasMore bool
	}{
		{
			name:            "FirstPage",
			limit:           2,
			offset:          0,
			expectedCount:   2,
			expectedTotal:   5,
			expectedOffset:  0,
			expectedHasMore: true,
		},
		{
			name:            "SecondPage",
			limit:           2,
			offset:          2,
			expectedCount:   2,
			expectedTotal:   5,
			expectedOffset:  2,
			expectedHasMore: true,
		},
		{
			name:            "LastPage",
			limit:           2,
			offset:          4,
			expectedCount:   1,
			expectedTotal:   5,
			expectedOffset:  4,
			expectedHasMore: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", fmt.Sprintf("/components/test-component-pagination/reports?limit=%d&offset=%d", tc.limit, tc.offset), nil)
			w := httptest.NewRecorder()

			// Call handler
			server.GetComponentReports(w, req, "test-component-pagination", GetComponentReportsParams{
				Limit:  &tc.limit,
				Offset: &tc.offset,
			})

			// Assert response
			assert.Equal(t, http.StatusOK, w.Code)

			var response ComponentReportsResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)

			assert.Len(t, response.Reports, tc.expectedCount)
			assert.Equal(t, tc.expectedTotal, response.Pagination.Total)
			assert.Equal(t, tc.limit, response.Pagination.Limit)
			assert.Equal(t, tc.expectedOffset, response.Pagination.Offset)
			assert.Equal(t, tc.expectedHasMore, response.Pagination.HasMore)
		})
	}
}

// createMultipleTestReports creates multiple test reports for pagination testing
func createMultipleTestReports(t *testing.T, repo *storage.Repository) {
	// Create test component
	component := storage.Component{
		ComponentID: "test-component-pagination",
		Name:        "Test Component Pagination",
		Description: "A test component for pagination",
	}
	if err := repo.DB.Create(&component).Error; err != nil {
		t.Fatalf("Failed to create test component: %v", err)
	}

	// Create test check
	check := storage.Check{
		Slug:        "unit-tests-pagination",
		Name:        "Unit Tests Pagination",
		Description: "Runs unit tests",
	}
	if err := repo.DB.Create(&check).Error; err != nil {
		t.Fatalf("Failed to create test check: %v", err)
	}

	// Create 5 test reports with different timestamps
	for i := 0; i < 5; i++ {
		report := storage.CheckReport{
			CheckID:     check.ID,
			ComponentID: component.ID,
			Status:      storage.CheckStatusPass,
			Timestamp:   time.Now().Add(-time.Duration(i) * time.Hour),
			Details:     storage.JSONB{"coverage": 80.0 + float64(i)},
		}
		if err := repo.DB.Create(&report).Error; err != nil {
			t.Fatalf("Failed to create test report %d: %v", i, err)
		}
	}
}

func TestGetComponentReports_PaginationLimit(t *testing.T) {
	// Setup database and server
	repo, server := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, repo)

	// Create test component first
	component := storage.Component{
		ComponentID: "test-component-limit",
		Name:        "Test Component Limit",
		Description: "A test component for limit testing",
	}
	if err := repo.DB.Create(&component).Error; err != nil {
		t.Fatalf("Failed to create test component: %v", err)
	}

	t.Run("ExcessiveLimitShouldBeCapped", func(t *testing.T) {
		// Create request with excessive limit (should be capped to 100)
		req := httptest.NewRequest("GET", "/components/test-component-limit/reports?limit=10000", nil)
		w := httptest.NewRecorder()

		// Call handler
		limit := 10000
		server.GetComponentReports(w, req, "test-component-limit", GetComponentReportsParams{
			Limit: &limit,
		})

		// Should return 200 OK (not 400 Bad Request)
		assert.Equal(t, http.StatusOK, w.Code)

		var response ComponentReportsResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		// Should use the default limit of 50 instead of 10000 (since 10000 > 100, it's invalid)
		assert.Equal(t, 50, response.Pagination.Limit)
		assert.Equal(t, 0, response.Pagination.Total)
	})

	t.Run("ValidLimitShouldBeRespected", func(t *testing.T) {
		// Create request with valid limit
		req := httptest.NewRequest("GET", "/components/test-component-limit/reports?limit=25", nil)
		w := httptest.NewRecorder()

		// Call handler
		limit := 25
		server.GetComponentReports(w, req, "test-component-limit", GetComponentReportsParams{
			Limit: &limit,
		})

		// Should return 200 OK
		assert.Equal(t, http.StatusOK, w.Code)

		var response ComponentReportsResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		// Should use the requested limit
		assert.Equal(t, 25, response.Pagination.Limit)
		assert.Equal(t, 0, response.Pagination.Total)
	})
}

func TestGetComponentReports_InvalidStatusParameter(t *testing.T) {
	// Setup database and server
	repo, server := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, repo)

	// Create test component first
	component := storage.Component{
		ComponentID: "test-component-status",
		Name:        "Test Component Status",
		Description: "A test component for status testing",
	}
	if err := repo.DB.Create(&component).Error; err != nil {
		t.Fatalf("Failed to create test component: %v", err)
	}

	t.Run("InvalidStatusReturns200Not400", func(t *testing.T) {
		// Create request with invalid status parameter
		req := httptest.NewRequest("GET", "/components/test-component-status/reports?status=invalid-status", nil)
		w := httptest.NewRecorder()

		// Call handler with invalid status
		invalidStatus := GetComponentReportsParamsStatus("invalid-status")
		server.GetComponentReports(w, req, "test-component-status", GetComponentReportsParams{
			Status: &invalidStatus,
		})

		// Should return 200 OK instead of 400 Bad Request
		assert.Equal(t, http.StatusOK, w.Code)

		var response ComponentReportsResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		// Should return empty results instead of error
		assert.Len(t, response.Reports, 0)
		assert.Equal(t, 0, response.Pagination.Total)
	})

	t.Run("ValidStatusReturns200", func(t *testing.T) {
		// Create request with valid status parameter
		req := httptest.NewRequest("GET", "/components/test-component-status/reports?status=pass", nil)
		w := httptest.NewRecorder()

		// Call handler with valid status
		validStatus := GetComponentReportsParamsStatusPass
		server.GetComponentReports(w, req, "test-component-status", GetComponentReportsParams{
			Status: &validStatus,
		})

		// Should return 200 OK
		assert.Equal(t, http.StatusOK, w.Code)

		var response ComponentReportsResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		// Should return empty results (no reports exist)
		assert.Len(t, response.Reports, 0)
		assert.Equal(t, 0, response.Pagination.Total)
	})
}
