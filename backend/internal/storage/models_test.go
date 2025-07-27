package storage_test

import (
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
