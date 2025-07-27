package storage_test

import (
	"testing"

	"github.com/doron-cohen/argus/backend/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
