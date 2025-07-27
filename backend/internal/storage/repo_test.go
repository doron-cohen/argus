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
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	repo := &storage.Repository{DB: db}
	require.NoError(t, repo.Migrate(t.Context()))
	return repo
}

func TestRepository_GetComponents_Empty(t *testing.T) {
	repo := setupTestRepo(t)
	components, err := repo.GetComponents(t.Context())
	require.NoError(t, err)
	require.Empty(t, components)
}

func TestStringArray(t *testing.T) {
	t.Run("Value and Scan", func(t *testing.T) {
		original := storage.StringArray{"alice@company.com", "@auth-team"}

		// Test Value()
		value, err := original.Value()
		require.NoError(t, err)
		// Value should be JSON bytes
		assert.IsType(t, []byte{}, value)

		// Test Scan()
		var scanned storage.StringArray
		err = scanned.Scan(value)
		require.NoError(t, err)
		assert.Equal(t, original, scanned)
	})

	t.Run("Empty array", func(t *testing.T) {
		original := storage.StringArray{}

		value, err := original.Value()
		require.NoError(t, err)
		assert.IsType(t, []byte{}, value)

		var scanned storage.StringArray
		err = scanned.Scan(value)
		require.NoError(t, err)
		assert.Equal(t, original, scanned)
	})

	t.Run("Nil array", func(t *testing.T) {
		var original storage.StringArray

		value, err := original.Value()
		require.NoError(t, err)
		assert.Nil(t, value)

		var scanned storage.StringArray
		err = scanned.Scan(nil)
		require.NoError(t, err)
		assert.Nil(t, scanned)
	})

	t.Run("Contains methods", func(t *testing.T) {
		array := storage.StringArray{"alice@company.com", "@auth-team", "bob@company.com"}

		assert.True(t, array.Contains("alice@company.com"))
		assert.True(t, array.Contains("@auth-team"))
		assert.False(t, array.Contains("charlie@company.com"))

		assert.True(t, array.ContainsAny([]string{"alice@company.com", "charlie@company.com"}))
		assert.True(t, array.ContainsAny([]string{"@auth-team"}))
		assert.False(t, array.ContainsAny([]string{"charlie@company.com", "dave@company.com"}))
	})
}

func TestRepository_JSONBStorage(t *testing.T) {
	// Use SQLite for testing since it's simpler to set up
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the schema
	err = db.AutoMigrate(&storage.Component{})
	require.NoError(t, err)

	repo := &storage.Repository{DB: db}
	ctx := context.Background()

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

func TestRepository_CheckMethods(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	t.Run("Create and Get Check by Slug", func(t *testing.T) {
		check := storage.Check{
			Slug:        "unit-tests",
			Name:        "Unit Tests",
			Description: "Runs unit tests for the component",
		}

		// Create check
		err := repo.CreateCheck(ctx, check)
		require.NoError(t, err)

		// Get check by slug
		retrieved, err := repo.GetCheckBySlug(ctx, "unit-tests")
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

func TestRepository_CheckReportMethods(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	// Create test component and check
	component := storage.Component{
		ComponentID: "test-service",
		Name:        "Test Service",
		Description: "A test service",
	}
	err := repo.CreateComponent(ctx, component)
	require.NoError(t, err)

	check := storage.Check{
		Slug:        "unit-tests",
		Name:        "Unit Tests",
		Description: "Runs unit tests",
	}
	err = repo.CreateCheck(ctx, check)
	require.NoError(t, err)

	t.Run("Create CheckReport", func(t *testing.T) {
		report := storage.CheckReport{
			CheckID:     check.ID,
			ComponentID: component.ID,
			Status:      storage.CheckStatusPass,
			Timestamp:   time.Now(),
			Details: storage.JSONB{
				"test_count": 42,
				"coverage":   85.5,
			},
			Metadata: storage.JSONB{
				"ci_job_id": "12345",
				"branch":    "main",
			},
		}

		// Create report
		err := repo.CreateCheckReport(ctx, report)
		require.NoError(t, err)

		// Verify the report was created by querying the database directly
		var count int64
		err = repo.DB.WithContext(ctx).Model(&storage.CheckReport{}).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Create CheckReport with nil JSONB", func(t *testing.T) {
		report := storage.CheckReport{
			CheckID:     check.ID,
			ComponentID: component.ID,
			Status:      storage.CheckStatusFail,
			Timestamp:   time.Now(),
			Details:     nil,
			Metadata:    nil,
		}

		// Create report
		err := repo.CreateCheckReport(ctx, report)
		require.NoError(t, err)
	})
}

func TestRepository_DatabaseSchema(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	t.Run("Check table schema", func(t *testing.T) {
		// Test that we can create a check with all required fields
		check := storage.Check{
			Slug:        "schema-test",
			Name:        "Schema Test",
			Description: "Test schema validation",
		}

		err := repo.CreateCheck(ctx, check)
		require.NoError(t, err)

		// Test unique constraint on slug
		duplicateCheck := storage.Check{
			Slug:        "schema-test", // Same slug
			Name:        "Duplicate Test",
			Description: "Should fail",
		}

		err = repo.CreateCheck(ctx, duplicateCheck)
		assert.Error(t, err) // Should fail due to unique constraint
	})

	t.Run("CheckReport table schema", func(t *testing.T) {
		// Create required dependencies
		component := storage.Component{
			ComponentID: "schema-test-service",
			Name:        "Schema Test Service",
		}
		err := repo.CreateComponent(ctx, component)
		require.NoError(t, err)

		check := storage.Check{
			Slug: "schema-test-check",
			Name: "Schema Test Check",
		}
		err = repo.CreateCheck(ctx, check)
		require.NoError(t, err)

		// Test that we can create a report with all required fields
		report := storage.CheckReport{
			CheckID:     check.ID,
			ComponentID: component.ID,
			Status:      storage.CheckStatusPass,
			Timestamp:   time.Now(),
			Details: storage.JSONB{
				"test": "data",
			},
			Metadata: storage.JSONB{
				"env": "test",
			},
		}

		err = repo.CreateCheckReport(ctx, report)
		require.NoError(t, err)
	})
}
