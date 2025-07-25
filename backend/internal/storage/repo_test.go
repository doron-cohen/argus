package storage_test

import (
	"context"
	"testing"

	"github.com/doron-cohen/argus/backend/internal/storage"
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
