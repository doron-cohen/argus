package storage_test

import (
	"testing"

	"github.com/doron-cohen/argus/backend/internal/storage"
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
