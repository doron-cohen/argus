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
}
