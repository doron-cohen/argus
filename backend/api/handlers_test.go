package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/doron-cohen/argus/backend/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
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
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
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
