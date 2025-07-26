package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/doron-cohen/argus/backend/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONB(t *testing.T) {
	t.Run("Value and Scan", func(t *testing.T) {
		original := storage.JSONB{
			"key1": "value1",
			"key2": float64(42), // JSON unmarshaling converts numbers to float64
			"key3": true,
			"key4": map[string]interface{}{
				"nested": "value",
			},
		}

		// Test Value()
		value, err := original.Value()
		require.NoError(t, err)
		// Value should be JSON bytes
		assert.IsType(t, []byte{}, value)

		// Test Scan()
		var scanned storage.JSONB
		err = scanned.Scan(value)
		require.NoError(t, err)
		assert.Equal(t, original, scanned)
	})

	t.Run("Empty JSONB", func(t *testing.T) {
		original := storage.JSONB{}

		value, err := original.Value()
		require.NoError(t, err)
		assert.IsType(t, []byte{}, value)

		var scanned storage.JSONB
		err = scanned.Scan(value)
		require.NoError(t, err)
		assert.Equal(t, original, scanned)
	})

	t.Run("Nil JSONB", func(t *testing.T) {
		var original storage.JSONB

		value, err := original.Value()
		require.NoError(t, err)
		assert.Nil(t, value)

		var scanned storage.JSONB
		err = scanned.Scan(nil)
		require.NoError(t, err)
		assert.Nil(t, scanned)
	})

	t.Run("JSONB methods", func(t *testing.T) {
		jsonb := storage.JSONB{
			"string": "value",
			"number": 42,
			"bool":   true,
		}

		// Test Get
		value, exists := jsonb.Get("string")
		assert.True(t, exists)
		assert.Equal(t, "value", value)

		_, exists = jsonb.Get("nonexistent")
		assert.False(t, exists)

		// Test Set
		jsonb.Set("new_key", "new_value")
		value, exists = jsonb.Get("new_key")
		assert.True(t, exists)
		assert.Equal(t, "new_value", value)

		// Test Delete
		jsonb.Delete("string")
		_, exists = jsonb.Get("string")
		assert.False(t, exists)

		// Test Keys
		keys := jsonb.Keys()
		assert.Len(t, keys, 3)
		assert.Contains(t, keys, "number")
		assert.Contains(t, keys, "bool")
		assert.Contains(t, keys, "new_key")

		// Test Has
		assert.True(t, jsonb.Has("number"))
		assert.False(t, jsonb.Has("string"))
	})
}

func TestCheckModel(t *testing.T) {
	t.Run("BeforeCreate generates UUID", func(t *testing.T) {
		check := storage.Check{
			Slug:        "test-check",
			Name:        "Test Check",
			Description: "A test check",
		}

		// UUID should be nil initially
		assert.Equal(t, uuid.Nil, check.ID)

		// Simulate BeforeCreate
		err := check.BeforeCreate(nil)
		require.NoError(t, err)

		// UUID should be generated
		assert.NotEqual(t, uuid.Nil, check.ID)
	})

	t.Run("CheckStatus constants", func(t *testing.T) {
		assert.Equal(t, storage.CheckStatus("pass"), storage.CheckStatusPass)
		assert.Equal(t, storage.CheckStatus("fail"), storage.CheckStatusFail)
		assert.Equal(t, storage.CheckStatus("disabled"), storage.CheckStatusDisabled)
		assert.Equal(t, storage.CheckStatus("skipped"), storage.CheckStatusSkipped)
		assert.Equal(t, storage.CheckStatus("unknown"), storage.CheckStatusUnknown)
		assert.Equal(t, storage.CheckStatus("error"), storage.CheckStatusError)
		assert.Equal(t, storage.CheckStatus("completed"), storage.CheckStatusCompleted)
	})
}

func TestCheckReportModel(t *testing.T) {
	t.Run("BeforeCreate generates UUID", func(t *testing.T) {
		report := storage.CheckReport{
			CheckID:     uuid.New(),
			ComponentID: uuid.New(),
			Status:      storage.CheckStatusPass,
			Timestamp:   time.Now(),
		}

		// UUID should be nil initially
		assert.Equal(t, uuid.Nil, report.ID)

		// Simulate BeforeCreate
		err := report.BeforeCreate(nil)
		require.NoError(t, err)

		// UUID should be generated
		assert.NotEqual(t, uuid.Nil, report.ID)
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
