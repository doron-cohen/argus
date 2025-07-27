package storage_test

import (
	"testing"

	"github.com/doron-cohen/argus/backend/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestJSONB(t *testing.T) {
	t.Run("Value and Scan", func(t *testing.T) {
		jsonb := storage.JSONB{
			"key1": "value1",
			"key2": float64(42),
			"key3": true,
		}

		// Test Value() method
		value, err := jsonb.Value()
		assert.NoError(t, err)
		assert.NotNil(t, value)

		// Test Scan() method
		var newJSONB storage.JSONB
		err = newJSONB.Scan(value)
		assert.NoError(t, err)
		assert.Equal(t, jsonb, newJSONB)
	})

	t.Run("Empty JSONB", func(t *testing.T) {
		jsonb := storage.JSONB{}

		value, err := jsonb.Value()
		assert.NoError(t, err)
		assert.NotNil(t, value)

		var newJSONB storage.JSONB
		err = newJSONB.Scan(value)
		assert.NoError(t, err)
		assert.Equal(t, jsonb, newJSONB)
	})

	t.Run("Nil JSONB", func(t *testing.T) {
		var jsonb storage.JSONB

		value, err := jsonb.Value()
		assert.NoError(t, err)
		assert.Nil(t, value)

		var newJSONB storage.JSONB
		err = newJSONB.Scan(nil)
		assert.NoError(t, err)
		assert.Nil(t, newJSONB)
	})

	t.Run("JSONB methods", func(t *testing.T) {
		jsonb := storage.JSONB{
			"string": "value",
			"number": float64(123),
			"bool":   true,
		}

		// Test Get method
		value, exists := jsonb.Get("string")
		assert.True(t, exists)
		assert.Equal(t, "value", value)

		value, exists = jsonb.Get("number")
		assert.True(t, exists)
		assert.Equal(t, float64(123), value)

		value, exists = jsonb.Get("bool")
		assert.True(t, exists)
		assert.Equal(t, true, value)

		value, exists = jsonb.Get("nonexistent")
		assert.False(t, exists)
		assert.Nil(t, value)

		// Test Set method
		jsonb.Set("new_key", "new_value")
		value, exists = jsonb.Get("new_key")
		assert.True(t, exists)
		assert.Equal(t, "new_value", value)

		// Test Delete method
		jsonb.Delete("string")
		value, exists = jsonb.Get("string")
		assert.False(t, exists)
		assert.Nil(t, value)

		// Test Keys method
		keys := jsonb.Keys()
		assert.Contains(t, keys, "number")
		assert.Contains(t, keys, "bool")
		assert.Contains(t, keys, "new_key")
		assert.NotContains(t, keys, "string")

		// Test Has method
		assert.True(t, jsonb.Has("number"))
		assert.True(t, jsonb.Has("bool"))
		assert.True(t, jsonb.Has("new_key"))
		assert.False(t, jsonb.Has("string"))
	})
}
