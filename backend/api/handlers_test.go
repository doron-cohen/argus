package api

import (
	"context"
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
	// Create a test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	repo := &storage.Repository{DB: db}
	// Run migrations to create tables
	if err := repo.Migrate(context.Background()); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	server := &APIServer{Repo: repo}

	t.Run("successful retrieval", func(t *testing.T) {
		// Create a test component
		component := storage.Component{
			ComponentID: "test-component",
			Name:        "Test Component",
		}
		if err := repo.DB.Create(&component).Error; err != nil {
			t.Fatalf("Failed to create test component: %v", err)
		}

		// Create a test check
		check := storage.Check{
			Slug:        "unit-tests",
			Name:        "Unit Tests",
			Description: "Runs unit tests",
		}
		if err := repo.DB.Create(&check).Error; err != nil {
			t.Fatalf("Failed to create test check: %v", err)
		}

		// Create a test report
		report := storage.CheckReport{
			CheckID:     check.ID,
			ComponentID: component.ID,
			Status:      storage.CheckStatusPass,
			Timestamp:   time.Now(),
			Details:     storage.JSONB{"coverage": 85.5},
			Metadata:    storage.JSONB{"ci_job": "123"},
		}
		if err := repo.DB.Create(&report).Error; err != nil {
			t.Fatalf("Failed to create test report: %v", err)
		}

		// Create request
		req := httptest.NewRequest("GET", "/components/test-component/reports", nil)
		w := httptest.NewRecorder()

		// Call handler
		server.GetComponentReports(w, req, "test-component", GetComponentReportsParams{})

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response ComponentReportsResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		assert.Len(t, response.Reports, 1)
		assert.Equal(t, report.ID.String(), response.Reports[0].Id)
		assert.Equal(t, "unit-tests", response.Reports[0].CheckSlug)
		assert.Equal(t, CheckReportStatusPass, response.Reports[0].Status)
	})

	t.Run("component not found", func(t *testing.T) {
		// Create request
		req := httptest.NewRequest("GET", "/components/non-existent/reports", nil)
		w := httptest.NewRecorder()

		// Call handler
		server.GetComponentReports(w, req, "non-existent", GetComponentReportsParams{})

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)

		var errorResponse Error
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		require.NoError(t, err)
		assert.Equal(t, "Component not found", errorResponse.Error)
	})

	t.Run("latest per check", func(t *testing.T) {
		// Create a test component
		component := storage.Component{
			ComponentID: "test-component-latest",
			Name:        "Test Component Latest",
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

		// Create request with latest_per_check=true
		req := httptest.NewRequest("GET", "/components/test-component-latest/reports?latest_per_check=true", nil)
		w := httptest.NewRecorder()

		// Call handler
		latestPerCheck := true
		server.GetComponentReports(w, req, "test-component-latest", GetComponentReportsParams{
			LatestPerCheck: &latestPerCheck,
		})

		// Assert response
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
			if response.Reports[i].CheckSlug == "unit-tests-latest" {
				unitTestReport = &response.Reports[i]
			} else if response.Reports[i].CheckSlug == "integration-tests-latest" {
				integrationTestReport = &response.Reports[i]
			}
		}

		require.NotNil(t, unitTestReport, "Should have unit-tests-latest report")
		require.NotNil(t, integrationTestReport, "Should have integration-tests-latest report")

		// Verify the unit-tests report is the newer one (should have fail status)
		assert.Equal(t, CheckReportStatusFail, unitTestReport.Status)
		assert.Equal(t, newReport.ID.String(), unitTestReport.Id)

		// Verify the integration-tests report
		assert.Equal(t, CheckReportStatusPass, integrationTestReport.Status)
		assert.Equal(t, integrationReport.ID.String(), integrationTestReport.Id)
	})
}
