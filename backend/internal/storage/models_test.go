package storage_test

import (
	"testing"
	"time"

	"github.com/doron-cohen/argus/backend/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponentModel(t *testing.T) {
	t.Run("BeforeCreate generates UUID", func(t *testing.T) {
		component := storage.Component{
			ComponentID: "test-service",
			Name:        "Test Service",
			Description: "A test service",
			Maintainers: storage.StringArray{"alice", "bob"},
			Team:        "platform",
		}

		// UUID should be nil initially
		assert.Equal(t, uuid.Nil, component.ID)

		// Simulate BeforeCreate
		err := component.BeforeCreate(nil)
		require.NoError(t, err)

		// UUID should be generated
		assert.NotEqual(t, uuid.Nil, component.ID)
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
