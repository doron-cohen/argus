package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser_Parse_Success(t *testing.T) {
	parser := NewParser()

	yamlContent := []byte(`name: "user-service"`)

	manifest, err := parser.Parse(yamlContent)

	require.NoError(t, err)
	assert.Equal(t, "user-service", manifest.Name)
}

func TestParser_Parse_InvalidYAML(t *testing.T) {
	parser := NewParser()

	invalidYaml := []byte(`name: "user-service`) // Missing closing quote

	manifest, err := parser.Parse(invalidYaml)

	assert.Error(t, err)
	assert.Nil(t, manifest)
}

func TestParser_Parse_EmptyContent(t *testing.T) {
	parser := NewParser()

	emptyContent := []byte(``)

	manifest, err := parser.Parse(emptyContent)

	require.NoError(t, err)
	assert.Equal(t, "", manifest.Name)
}

func TestParser_Validate_Success(t *testing.T) {
	parser := NewParser()

	manifest := &Manifest{
		Name: "user-service",
	}

	err := parser.Validate(manifest)

	assert.NoError(t, err)
}

func TestParser_Validate_EmptyName(t *testing.T) {
	parser := NewParser()

	manifest := &Manifest{
		Name: "",
	}

	err := parser.Validate(manifest)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "manifest name is required")
}

func TestParser_ParseAndValidate_FullWorkflow(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name        string
		content     string
		expectError bool
		expectedMsg string
	}{
		{
			name:        "valid manifest",
			content:     `name: "api-gateway"`,
			expectError: false,
		},
		{
			name:        "empty name",
			content:     `name: ""`,
			expectError: true,
			expectedMsg: "manifest name is required",
		},
		{
			name:        "missing name field",
			content:     `description: "some service"`,
			expectError: true,
			expectedMsg: "manifest name is required",
		},
		{
			name:        "malformed yaml",
			content:     `name: [invalid`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse
			manifest, parseErr := parser.Parse([]byte(tt.content))

			if parseErr != nil {
				assert.True(t, tt.expectError, "Expected no parse error but got: %v", parseErr)
				return
			}

			// Validate
			validateErr := parser.Validate(manifest)

			if tt.expectError {
				assert.Error(t, validateErr)
				if tt.expectedMsg != "" {
					assert.Contains(t, validateErr.Error(), tt.expectedMsg)
				}
			} else {
				assert.NoError(t, validateErr)
				assert.NotEmpty(t, manifest.Name)
			}
		})
	}
}
