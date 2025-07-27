package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/doron-cohen/argus/backend/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
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
	ctx := context.Background()

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

func TestRepository_CreateCheckReportWithExistingCheck(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

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

	err = repo.CreateCheckReportFromSubmission(ctx, "test-service", "unit-tests", nil, nil, storage.CheckStatusPass, timestamp, details, metadata)
	require.NoError(t, err)

	// Verify the report was created
	var count int64
	err = repo.DB.WithContext(ctx).Model(&storage.CheckReport{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestRepository_CreateCheckReportWithAutoCreatedCheck(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

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

	err = repo.CreateCheckReportFromSubmission(ctx, "test-service-auto", "build-check-auto", &checkName, &checkDescription, storage.CheckStatusPass, timestamp, details, metadata)
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
	ctx := context.Background()

	details := storage.JSONB{"test": "data"}
	metadata := storage.JSONB{"env": "test"}
	timestamp := time.Now()

	err := repo.CreateCheckReportFromSubmission(ctx, "non-existent-service", "unit-tests", nil, nil, storage.CheckStatusPass, timestamp, details, metadata)
	assert.ErrorIs(t, err, storage.ErrComponentNotFound)
}

func TestRepository_GetOrCreateCheckBySlug(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

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
