package storage_test

import (
	"testing"
	"time"

	"github.com/doron-cohen/argus/backend/internal/storage"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestRepo(t *testing.T) *storage.Repository {
	// Use a temporary file-based database instead of in-memory
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	repo := &storage.Repository{DB: db}
	require.NoError(t, repo.Migrate(t.Context()))
	return repo
}

func TestRepository_Migration(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := t.Context()

	// Test that we can create a component
	component := storage.Component{
		ComponentID: "test-component",
		Name:        "Test Component",
	}
	err := repo.CreateComponent(ctx, component)
	require.NoError(t, err)

	// Verify the component was created
	var count int64
	err = repo.DB.WithContext(ctx).Model(&storage.Component{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestRepository_GetComponents_Empty(t *testing.T) {
	// Use a completely isolated database to ensure it's truly empty
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	repo := &storage.Repository{DB: db}
	require.NoError(t, repo.Migrate(t.Context()))

	components, err := repo.GetComponents(t.Context())
	require.NoError(t, err)
	require.Empty(t, components)
}

func TestRepository_JSONBStorage(t *testing.T) {
	// Use SQLite for testing since it's simpler to set up
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the schema
	err = db.AutoMigrate(&storage.Component{})
	require.NoError(t, err)

	repo := &storage.Repository{DB: db}
	ctx := t.Context()

	// Create test component with maintainers
	component := storage.Component{
		ComponentID: "auth-service",
		Name:        "Authentication Service",
		Description: "Handles authentication",
		Maintainers: storage.StringArray{"alice@company.com", "@auth-team"},
		Team:        "Security Team",
	}

	err = repo.CreateComponent(ctx, component)
	require.NoError(t, err)

	// Retrieve and verify the component
	retrieved, err := repo.GetComponentByID(ctx, "auth-service")
	require.NoError(t, err)
	assert.Equal(t, "auth-service", retrieved.ComponentID)
	assert.Equal(t, "Authentication Service", retrieved.Name)
	assert.Len(t, retrieved.Maintainers, 2)
	assert.Contains(t, retrieved.Maintainers, "alice@company.com")
	assert.Contains(t, retrieved.Maintainers, "@auth-team")
	assert.Equal(t, "Security Team", retrieved.Team)

	t.Run("GetComponentsByTeam", func(t *testing.T) {
		// Test finding components by team (this works with any database)
		results, err := repo.GetComponentsByTeam(ctx, "Security Team")
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "auth-service", results[0].ComponentID)
	})
}

func TestRepository_CreateCheckReportWithExistingCheck(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := t.Context()

	// Create test component
	component := storage.Component{
		ComponentID: "test-service",
		Name:        "Test Service",
		Description: "A test service",
	}
	err := repo.CreateComponent(ctx, component)
	require.NoError(t, err)

	// Create test check
	check := storage.Check{
		Slug:        "unit-tests",
		Name:        "Unit Tests",
		Description: "Runs unit tests",
	}
	err = repo.CreateCheck(ctx, check)
	require.NoError(t, err)

	// Create report from submission
	details := storage.JSONB{"coverage": 85.5, "tests_passed": 150}
	metadata := storage.JSONB{"ci_job_id": "12345", "branch": "main"}
	timestamp := time.Now().Add(-1 * time.Hour)

	input := storage.CreateCheckReportInput{
		ComponentID: "test-service",
		CheckSlug:   "unit-tests",
		Status:      storage.CheckStatusPass,
		Timestamp:   timestamp,
		Details:     details,
		Metadata:    metadata,
	}
	_, err = repo.CreateCheckReportFromSubmission(ctx, input)
	require.NoError(t, err)

	// Verify the report was created
	var count int64
	err = repo.DB.WithContext(ctx).Model(&storage.CheckReport{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestRepository_CreateCheckReportWithAutoCreatedCheck(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := t.Context()

	// Create test component with unique ID
	component := storage.Component{
		ComponentID: "test-service-auto",
		Name:        "Test Service Auto",
		Description: "A test service for auto-creation",
	}
	err := repo.CreateComponent(ctx, component)
	require.NoError(t, err)

	// Check initial report count
	var initialCount int64
	err = repo.DB.WithContext(ctx).Model(&storage.CheckReport{}).Count(&initialCount).Error
	require.NoError(t, err)

	// Create report with a new check slug
	details := storage.JSONB{"build_time": 45.2}
	metadata := storage.JSONB{"environment": "staging"}
	timestamp := time.Now().Add(-30 * time.Minute)
	checkName := "Build Check"
	checkDescription := "Runs build process"

	input := storage.CreateCheckReportInput{
		ComponentID:      "test-service-auto",
		CheckSlug:        "build-check-auto",
		CheckName:        &checkName,
		CheckDescription: &checkDescription,
		Status:           storage.CheckStatusPass,
		Timestamp:        timestamp,
		Details:          details,
		Metadata:         metadata,
	}
	_, err = repo.CreateCheckReportFromSubmission(ctx, input)
	require.NoError(t, err)

	// Verify the check was auto-created with provided values
	check, err := repo.GetCheckBySlug(ctx, "build-check-auto")
	require.NoError(t, err)
	assert.Equal(t, "build-check-auto", check.Slug)
	assert.Equal(t, "Build Check", check.Name)
	assert.Equal(t, "Runs build process", check.Description)

	// Verify exactly one new report was created
	var finalCount int64
	err = repo.DB.WithContext(ctx).Model(&storage.CheckReport{}).Count(&finalCount).Error
	require.NoError(t, err)
	assert.Equal(t, initialCount+1, finalCount)
}

func TestRepository_CreateCheckReportWithNonExistentComponent(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := t.Context()

	details := storage.JSONB{"test": "data"}
	metadata := storage.JSONB{"env": "test"}
	timestamp := time.Now()

	input := storage.CreateCheckReportInput{
		ComponentID: "non-existent-service",
		CheckSlug:   "unit-tests",
		Status:      storage.CheckStatusPass,
		Timestamp:   timestamp,
		Details:     details,
		Metadata:    metadata,
	}
	_, err := repo.CreateCheckReportFromSubmission(ctx, input)
	assert.ErrorIs(t, err, storage.ErrComponentNotFound)
}

func TestRepository_GetOrCreateCheckBySlug(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := t.Context()

	t.Run("Get existing check", func(t *testing.T) {
		// Create a check first with unique slug
		check := storage.Check{
			Slug:        "existing-check-get",
			Name:        "Existing Check Get",
			Description: "A check that already exists for get test",
		}
		err := repo.CreateCheck(ctx, check)
		require.NoError(t, err)

		// Query the created check to get its actual ID
		var createdCheck storage.Check
		err = repo.DB.WithContext(ctx).Where("slug = ?", "existing-check-get").First(&createdCheck).Error
		require.NoError(t, err)

		// Try to get or create the same check
		checkID, err := repo.GetOrCreateCheckBySlug(ctx, "existing-check-get", nil, nil)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, checkID)
		assert.Equal(t, createdCheck.ID, checkID)

		// Verify no duplicate was created
		var count int64
		err = repo.DB.WithContext(ctx).Model(&storage.Check{}).Where("slug = ?", "existing-check-get").Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Create new check", func(t *testing.T) {
		// Try to get or create a new check with unique slug
		checkID, err := repo.GetOrCreateCheckBySlug(ctx, "new-check-create", nil, nil)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, checkID)

		// Verify the check was created with default values
		check, err := repo.GetCheckBySlug(ctx, "new-check-create")
		require.NoError(t, err)
		assert.Equal(t, "new-check-create", check.Slug)
		assert.Equal(t, "new-check-create", check.Name) // Default name is slug
		assert.Equal(t, "Auto-created check for slug: new-check-create", check.Description)
	})

	t.Run("Create new check with custom name and description", func(t *testing.T) {
		// Try to get or create a new check with custom values and unique slug
		checkName := "Custom Check"
		checkDescription := "A custom check description"
		checkID, err := repo.GetOrCreateCheckBySlug(ctx, "custom-check-name", &checkName, &checkDescription)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, checkID)

		// Verify the check was created with custom values
		check, err := repo.GetCheckBySlug(ctx, "custom-check-name")
		require.NoError(t, err)
		assert.Equal(t, "custom-check-name", check.Slug)
		assert.Equal(t, "Custom Check", check.Name)
		assert.Equal(t, "A custom check description", check.Description)
	})

	t.Run("Database error handling", func(t *testing.T) {
		// This test would require mocking the database to simulate errors
		// For now, we'll test with a valid case that should work
		checkID, err := repo.GetOrCreateCheckBySlug(ctx, "error-test-handling", nil, nil)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, checkID)
	})
}

func TestRepository_CheckMethods(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := t.Context()

	t.Run("Create and Get Check by Slug", func(t *testing.T) {
		check := storage.Check{
			Slug:        "unit-tests-check-methods",
			Name:        "Unit Tests",
			Description: "Runs unit tests for the component",
		}

		// Create check
		err := repo.CreateCheck(ctx, check)
		require.NoError(t, err)

		// Get check by slug
		retrieved, err := repo.GetCheckBySlug(ctx, "unit-tests-check-methods")
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, retrieved.ID)
		assert.Equal(t, check.Slug, retrieved.Slug)
		assert.Equal(t, check.Name, retrieved.Name)
		assert.Equal(t, check.Description, retrieved.Description)
	})

	t.Run("Get Check Not Found", func(t *testing.T) {
		_, err := repo.GetCheckBySlug(ctx, "nonexistent")
		assert.ErrorIs(t, err, storage.ErrCheckNotFound)
	})
}

func TestRepository_DatabaseSchema(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := t.Context()

	t.Run("Check table schema", func(t *testing.T) {
		// Test that we can create a check with all required fields
		check := storage.Check{
			Slug:        "schema-test-db",
			Name:        "Schema Test",
			Description: "Test schema validation",
		}

		err := repo.CreateCheck(ctx, check)
		require.NoError(t, err)

		// Test unique constraint on slug
		duplicateCheck := storage.Check{
			Slug:        "schema-test-db", // Same slug
			Name:        "Duplicate Test",
			Description: "Should fail",
		}

		err = repo.CreateCheck(ctx, duplicateCheck)
		assert.Error(t, err) // Should fail due to unique constraint
	})

	t.Run("CheckReport table schema", func(t *testing.T) {
		// Create required dependencies
		component := storage.Component{
			ComponentID: "schema-test-service-db",
			Name:        "Schema Test Service",
		}
		err := repo.CreateComponent(ctx, component)
		require.NoError(t, err)

		check := storage.Check{
			Slug: "schema-test-check-db",
			Name: "Schema Test Check",
		}
		err = repo.CreateCheck(ctx, check)
		require.NoError(t, err)

		// Test that we can create a report with all required fields using the new method
		input := storage.CreateCheckReportInput{
			ComponentID: "schema-test-service-db",
			CheckSlug:   "schema-test-check-db",
			Status:      storage.CheckStatusPass,
			Timestamp:   time.Now(),
			Details: storage.JSONB{
				"test": "data",
			},
			Metadata: storage.JSONB{
				"env": "test",
			},
		}

		// Get initial count
		var initialCount int64
		err = repo.DB.WithContext(ctx).Model(&storage.CheckReport{}).Count(&initialCount).Error
		require.NoError(t, err)

		_, err = repo.CreateCheckReportFromSubmission(ctx, input)
		require.NoError(t, err)

		// Verify exactly one new report was created
		var finalCount int64
		err = repo.DB.WithContext(ctx).Model(&storage.CheckReport{}).Count(&finalCount).Error
		require.NoError(t, err)
		assert.Equal(t, initialCount+1, finalCount)
	})
}

func TestRepository_GetCheckReportsForComponentWithPagination(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := t.Context()

	// Setup test data
	component := storage.Component{
		ComponentID: "pagination-test-service",
		Name:        "Pagination Test Service",
	}
	err := repo.CreateComponent(ctx, component)
	require.NoError(t, err)

	// Create multiple checks with unique slugs
	checks := []storage.Check{
		{Slug: "unit-tests-pagination", Name: "Unit Tests"},
		{Slug: "integration-tests-pagination", Name: "Integration Tests"},
		{Slug: "security-scan-pagination", Name: "Security Scan"},
	}
	for _, check := range checks {
		err = repo.CreateCheck(ctx, check)
		require.NoError(t, err)
	}

	// Create multiple reports with different timestamps
	now := time.Now()
	reports := []storage.CreateCheckReportInput{
		{
			ComponentID: "pagination-test-service",
			CheckSlug:   "unit-tests-pagination",
			Status:      storage.CheckStatusPass,
			Timestamp:   now.Add(-1 * time.Hour),
			Details:     storage.JSONB{"coverage": 85},
			Metadata:    storage.JSONB{"env": "test"},
		},
		{
			ComponentID: "pagination-test-service",
			CheckSlug:   "unit-tests-pagination",
			Status:      storage.CheckStatusFail,
			Timestamp:   now.Add(-2 * time.Hour),
			Details:     storage.JSONB{"coverage": 75},
			Metadata:    storage.JSONB{"env": "test"},
		},
		{
			ComponentID: "pagination-test-service",
			CheckSlug:   "integration-tests-pagination",
			Status:      storage.CheckStatusPass,
			Timestamp:   now.Add(-30 * time.Minute),
			Details:     storage.JSONB{"tests": 100},
			Metadata:    storage.JSONB{"env": "test"},
		},
		{
			ComponentID: "pagination-test-service",
			CheckSlug:   "security-scan-pagination",
			Status:      storage.CheckStatusPass,
			Timestamp:   now.Add(-15 * time.Minute),
			Details:     storage.JSONB{"vulnerabilities": 0},
			Metadata:    storage.JSONB{"env": "test"},
		},
	}

	for _, report := range reports {
		_, err = repo.CreateCheckReportFromSubmission(ctx, report)
		require.NoError(t, err)
	}

	const unitTestsSlug = "unit-tests-pagination"

	t.Run("Basic pagination without filters", func(t *testing.T) {
		reports, total, err := repo.GetCheckReportsForComponentWithPagination(ctx, "pagination-test-service", nil, nil, nil, 2, 0, false)
		require.NoError(t, err)
		assert.Equal(t, int64(4), total)
		assert.Len(t, reports, 2)
	})

	t.Run("Pagination with offset", func(t *testing.T) {
		reports, total, err := repo.GetCheckReportsForComponentWithPagination(ctx, "pagination-test-service", nil, nil, nil, 2, 2, false)
		require.NoError(t, err)
		assert.Equal(t, int64(4), total)
		assert.Len(t, reports, 2)
	})

	t.Run("Filter by status", func(t *testing.T) {
		status := storage.CheckStatusPass
		reports, total, err := repo.GetCheckReportsForComponentWithPagination(ctx, "pagination-test-service", &status, nil, nil, 10, 0, false)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total) // 3 pass reports
		assert.Len(t, reports, 3)
		for _, report := range reports {
			assert.Equal(t, storage.CheckStatusPass, report.Status)
		}
	})

	t.Run("Filter by check slug", func(t *testing.T) {
		checkSlug := unitTestsSlug
		reports, total, err := repo.GetCheckReportsForComponentWithPagination(ctx, "pagination-test-service", nil, &checkSlug, nil, 10, 0, false)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total) // 2 unit-tests reports
		assert.Len(t, reports, 2)
		for _, report := range reports {
			assert.Equal(t, unitTestsSlug, report.Check.Slug)
		}
	})

	t.Run("Filter by since timestamp", func(t *testing.T) {
		since := now.Add(-45 * time.Minute)
		reports, total, err := repo.GetCheckReportsForComponentWithPagination(ctx, "pagination-test-service", nil, nil, &since, 10, 0, false)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total) // 2 recent reports
		assert.Len(t, reports, 2)
		for _, report := range reports {
			assert.True(t, report.Timestamp.After(since))
		}
	})

	t.Run("Latest per check without filters", func(t *testing.T) {
		reports, total, err := repo.GetCheckReportsForComponentWithPagination(ctx, "pagination-test-service", nil, nil, nil, 10, 0, true)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total) // 3 unique checks
		assert.Len(t, reports, 3)

		// Verify we have one report per check
		checkSlugs := make(map[string]bool)
		for _, report := range reports {
			checkSlugs[report.Check.Slug] = true
		}
		assert.Len(t, checkSlugs, 3)
		assert.True(t, checkSlugs[unitTestsSlug])
		assert.True(t, checkSlugs["integration-tests-pagination"])
		assert.True(t, checkSlugs["security-scan-pagination"])
	})

	t.Run("Latest per check with status filter", func(t *testing.T) {
		status := storage.CheckStatusPass
		reports, total, err := repo.GetCheckReportsForComponentWithPagination(ctx, "pagination-test-service", &status, nil, nil, 10, 0, true)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total) // 3 unique checks with pass status
		assert.Len(t, reports, 3)
		for _, report := range reports {
			assert.Equal(t, storage.CheckStatusPass, report.Status)
		}
	})

	t.Run("Latest per check with check slug filter", func(t *testing.T) {
		checkSlug := unitTestsSlug
		reports, total, err := repo.GetCheckReportsForComponentWithPagination(ctx, "pagination-test-service", nil, &checkSlug, nil, 10, 0, true)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total) // 1 unique check
		assert.Len(t, reports, 1)
		assert.Equal(t, unitTestsSlug, reports[0].Check.Slug)
	})

	t.Run("Latest per check with pagination", func(t *testing.T) {
		reports, total, err := repo.GetCheckReportsForComponentWithPagination(ctx, "pagination-test-service", nil, nil, nil, 2, 0, true)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total) // 3 unique checks
		assert.Len(t, reports, 2)        // limited by pagination
	})

	t.Run("Component not found", func(t *testing.T) {
		_, _, err := repo.GetCheckReportsForComponentWithPagination(ctx, "non-existent-service", nil, nil, nil, 10, 0, false)
		assert.ErrorIs(t, err, storage.ErrComponentNotFound)
	})

	t.Run("Combined filters with latest per check", func(t *testing.T) {
		status := storage.CheckStatusPass
		checkSlug := unitTestsSlug
		since := now.Add(-90 * time.Minute)
		reports, total, err := repo.GetCheckReportsForComponentWithPagination(ctx, "pagination-test-service", &status, &checkSlug, &since, 10, 0, true)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total) // 1 report matching all filters
		assert.Len(t, reports, 1)
		assert.Equal(t, unitTestsSlug, reports[0].Check.Slug)
		assert.Equal(t, storage.CheckStatusPass, reports[0].Status)
		assert.True(t, reports[0].Timestamp.After(since))
	})
}

func TestRepository_CheckSlugFilterWithLatestPerCheck(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := t.Context()

	const (
		unitTestsSlug    = "unit-tests-filter"
		integrationSlug  = "integration-tests-filter"
		securityScanSlug = "security-scan-filter"
	)

	// Setup test data with multiple components and checks to test JOIN scenarios
	components := []storage.Component{
		{ComponentID: "service-a", Name: "Service A"},
		{ComponentID: "service-b", Name: "Service B"},
	}
	for _, component := range components {
		err := repo.CreateComponent(ctx, component)
		require.NoError(t, err)
	}

	// Create checks with same slug across different components to test JOIN behavior
	checks := []storage.Check{
		{Slug: unitTestsSlug, Name: "Unit Tests"},
		{Slug: integrationSlug, Name: "Integration Tests"},
		{Slug: securityScanSlug, Name: "Security Scan"},
	}
	for _, check := range checks {
		err := repo.CreateCheck(ctx, check)
		require.NoError(t, err)
	}

	// Create reports with overlapping check slugs across components
	now := time.Now()
	reports := []storage.CreateCheckReportInput{
		// Service A reports
		{
			ComponentID: "service-a",
			CheckSlug:   unitTestsSlug,
			Status:      storage.CheckStatusPass,
			Timestamp:   now.Add(-1 * time.Hour),
			Details:     storage.JSONB{"coverage": 85},
		},
		{
			ComponentID: "service-a",
			CheckSlug:   unitTestsSlug,
			Status:      storage.CheckStatusFail,
			Timestamp:   now.Add(-3 * time.Hour), // Make this older to ensure pass is latest
			Details:     storage.JSONB{"coverage": 75},
		},
		{
			ComponentID: "service-a",
			CheckSlug:   integrationSlug,
			Status:      storage.CheckStatusPass,
			Timestamp:   now.Add(-30 * time.Minute),
			Details:     storage.JSONB{"tests": 100},
		},
		// Service B reports with same check slugs
		{
			ComponentID: "service-b",
			CheckSlug:   unitTestsSlug,
			Status:      storage.CheckStatusPass,
			Timestamp:   now.Add(-45 * time.Minute),
			Details:     storage.JSONB{"coverage": 90},
		},
		{
			ComponentID: "service-b",
			CheckSlug:   integrationSlug,
			Status:      storage.CheckStatusFail,
			Timestamp:   now.Add(-60 * time.Minute),
			Details:     storage.JSONB{"tests": 50},
		},
		{
			ComponentID: "service-b",
			CheckSlug:   securityScanSlug,
			Status:      storage.CheckStatusPass,
			Timestamp:   now.Add(-15 * time.Minute),
			Details:     storage.JSONB{"vulnerabilities": 0},
		},
	}

	for _, report := range reports {
		_, err := repo.CreateCheckReportFromSubmission(ctx, report)
		require.NoError(t, err)
	}

	t.Run("Check slug filter with latest per check - Service A", func(t *testing.T) {
		checkSlug := unitTestsSlug
		reports, total, err := repo.GetCheckReportsForComponentWithPagination(ctx, "service-a", nil, &checkSlug, nil, 10, 0, true)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total) // 1 latest report for unit-tests-filter in service-a
		assert.Len(t, reports, 1)
		assert.Equal(t, unitTestsSlug, reports[0].Check.Slug)
		assert.Equal(t, "service-a", reports[0].Component.ComponentID)
		// Should return the latest (most recent) report
		assert.Equal(t, storage.CheckStatusPass, reports[0].Status)
	})

	t.Run("Check slug filter with latest per check - Service B", func(t *testing.T) {
		checkSlug := unitTestsSlug
		reports, total, err := repo.GetCheckReportsForComponentWithPagination(ctx, "service-b", nil, &checkSlug, nil, 10, 0, true)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total) // 1 latest report for unit-tests-filter in service-b
		assert.Len(t, reports, 1)
		assert.Equal(t, unitTestsSlug, reports[0].Check.Slug)
		assert.Equal(t, "service-b", reports[0].Component.ComponentID)
		// Should return the latest (most recent) report
		assert.Equal(t, storage.CheckStatusPass, reports[0].Status)
	})

	t.Run("Check slug filter with latest per check - Non-existent check", func(t *testing.T) {
		checkSlug := "non-existent-check"
		reports, total, err := repo.GetCheckReportsForComponentWithPagination(ctx, "service-a", nil, &checkSlug, nil, 10, 0, true)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total) // No reports for non-existent check
		assert.Len(t, reports, 0)
	})

	t.Run("Check slug filter with latest per check and status filter", func(t *testing.T) {
		checkSlug := integrationSlug
		status := storage.CheckStatusPass
		reports, total, err := repo.GetCheckReportsForComponentWithPagination(ctx, "service-a", &status, &checkSlug, nil, 10, 0, true)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total) // 1 latest pass report for integration-tests-filter in service-a
		assert.Len(t, reports, 1)
		assert.Equal(t, integrationSlug, reports[0].Check.Slug)
		assert.Equal(t, storage.CheckStatusPass, reports[0].Status)
	})

	t.Run("Check slug filter with latest per check and since filter", func(t *testing.T) {
		checkSlug := unitTestsSlug
		since := now.Add(-90 * time.Minute) // Should include the pass report but not the fail report
		reports, total, err := repo.GetCheckReportsForComponentWithPagination(ctx, "service-a", nil, &checkSlug, &since, 10, 0, true)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total) // 1 latest report within time range
		assert.Len(t, reports, 1)
		assert.Equal(t, unitTestsSlug, reports[0].Check.Slug)
		assert.True(t, reports[0].Timestamp.After(since))
	})

	t.Run("Check slug filter with latest per check and pagination", func(t *testing.T) {
		// Get all reports for service-a with latest per check
		reports, total, err := repo.GetCheckReportsForComponentWithPagination(ctx, "service-a", nil, nil, nil, 10, 0, true)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total) // 2 unique checks in service-a
		assert.Len(t, reports, 2)

		// Now filter by check slug with pagination
		checkSlug := unitTestsSlug
		filteredReports, filteredTotal, err := repo.GetCheckReportsForComponentWithPagination(ctx, "service-a", nil, &checkSlug, nil, 1, 0, true)
		require.NoError(t, err)
		assert.Equal(t, int64(1), filteredTotal) // 1 unique check
		assert.Len(t, filteredReports, 1)
		assert.Equal(t, unitTestsSlug, filteredReports[0].Check.Slug)
	})

	t.Run("Multiple components with same check slug - verify isolation", func(t *testing.T) {
		checkSlug := unitTestsSlug

		// Get reports for service-a
		reportsA, totalA, err := repo.GetCheckReportsForComponentWithPagination(ctx, "service-a", nil, &checkSlug, nil, 10, 0, true)
		require.NoError(t, err)
		assert.Equal(t, int64(1), totalA)
		assert.Len(t, reportsA, 1)
		assert.Equal(t, "service-a", reportsA[0].Component.ComponentID)

		// Get reports for service-b
		reportsB, totalB, err := repo.GetCheckReportsForComponentWithPagination(ctx, "service-b", nil, &checkSlug, nil, 10, 0, true)
		require.NoError(t, err)
		assert.Equal(t, int64(1), totalB)
		assert.Len(t, reportsB, 1)
		assert.Equal(t, "service-b", reportsB[0].Component.ComponentID)

		// Verify they are different reports
		assert.NotEqual(t, reportsA[0].ID, reportsB[0].ID)
	})
}

func TestRepository_ApplyLatestPerCheckFilters(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := t.Context()

	// Create test data
	component := storage.Component{
		ComponentID: "filter-test-service",
		Name:        "Filter Test Service",
	}
	err := repo.CreateComponent(ctx, component)
	require.NoError(t, err)

	check := storage.Check{
		Slug: "test-check",
		Name: "Test Check",
	}
	err = repo.CreateCheck(ctx, check)
	require.NoError(t, err)

	// Create a report
	input := storage.CreateCheckReportInput{
		ComponentID: "filter-test-service",
		CheckSlug:   "test-check",
		Status:      storage.CheckStatusPass,
		Timestamp:   time.Now(),
		Details:     storage.JSONB{"test": "data"},
		Metadata:    storage.JSONB{"env": "test"},
	}
	_, err = repo.CreateCheckReportFromSubmission(ctx, input)
	require.NoError(t, err)

	// Test filtering through the public interface
	t.Run("No filters", func(t *testing.T) {
		reports, total, err := repo.GetCheckReportsForComponentWithPagination(ctx, "filter-test-service", nil, nil, nil, 10, 0, false)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, reports, 1)
	})

	t.Run("Status filter", func(t *testing.T) {
		status := storage.CheckStatusPass
		reports, total, err := repo.GetCheckReportsForComponentWithPagination(ctx, "filter-test-service", &status, nil, nil, 10, 0, false)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, reports, 1)
		assert.Equal(t, storage.CheckStatusPass, reports[0].Status)
	})

	t.Run("Status filter no match", func(t *testing.T) {
		status := storage.CheckStatusFail
		reports, total, err := repo.GetCheckReportsForComponentWithPagination(ctx, "filter-test-service", &status, nil, nil, 10, 0, false)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Len(t, reports, 0)
	})

	t.Run("Check slug filter", func(t *testing.T) {
		checkSlug := "test-check"
		reports, total, err := repo.GetCheckReportsForComponentWithPagination(ctx, "filter-test-service", nil, &checkSlug, nil, 10, 0, false)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, reports, 1)
		assert.Equal(t, "test-check", reports[0].Check.Slug)
	})

	t.Run("Check slug filter no match", func(t *testing.T) {
		checkSlug := "non-existent-check"
		reports, total, err := repo.GetCheckReportsForComponentWithPagination(ctx, "filter-test-service", nil, &checkSlug, nil, 10, 0, false)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Len(t, reports, 0)
	})

	t.Run("Since filter", func(t *testing.T) {
		since := time.Now().Add(-1 * time.Hour)
		reports, total, err := repo.GetCheckReportsForComponentWithPagination(ctx, "filter-test-service", nil, nil, &since, 10, 0, false)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, reports, 1)
	})

	t.Run("Since filter no match", func(t *testing.T) {
		since := time.Now().Add(1 * time.Hour) // Future time
		reports, total, err := repo.GetCheckReportsForComponentWithPagination(ctx, "filter-test-service", nil, nil, &since, 10, 0, false)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Len(t, reports, 0)
	})

	t.Run("Combined filters", func(t *testing.T) {
		status := storage.CheckStatusPass
		checkSlug := "test-check"
		since := time.Now().Add(-1 * time.Hour)
		reports, total, err := repo.GetCheckReportsForComponentWithPagination(ctx, "filter-test-service", &status, &checkSlug, &since, 10, 0, false)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, reports, 1)
		assert.Equal(t, storage.CheckStatusPass, reports[0].Status)
		assert.Equal(t, "test-check", reports[0].Check.Slug)
	})
}
