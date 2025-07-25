package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser_Parse_Success(t *testing.T) {
	parser := NewParser()

	yamlContent := []byte(`
version: "v1"
name: "user-service"
`)

	manifest, err := parser.Parse(yamlContent)

	require.NoError(t, err)
	assert.Equal(t, "v1", manifest.Version)
	assert.Equal(t, "user-service", manifest.Name)
}

func TestParser_Parse_InvalidYAML(t *testing.T) {
	parser := NewParser()

	invalidYaml := []byte(`version: "v1" component: [invalid`) // Malformed YAML

	manifest, err := parser.Parse(invalidYaml)

	assert.Error(t, err)
	assert.Nil(t, manifest)
}

func TestParser_Parse_EmptyContent(t *testing.T) {
	parser := NewParser()

	emptyContent := []byte(``)

	manifest, err := parser.Parse(emptyContent)

	require.NoError(t, err)
	assert.Equal(t, "", manifest.Version)
	assert.Equal(t, "", manifest.Name)
}

func TestParser_Validate_Success(t *testing.T) {
	parser := NewParser()

	manifest := &Manifest{
		Version: "v1",
		Name:    "user-service",
	}

	err := parser.Validate(manifest)

	assert.NoError(t, err)
}

func TestParser_Validate_EmptyVersion(t *testing.T) {
	parser := NewParser()

	manifest := &Manifest{
		Version: "",
		Name:    "user-service",
	}

	err := parser.Validate(manifest)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "manifest version is required")
}

func TestParser_Validate_UnsupportedVersion(t *testing.T) {
	parser := NewParser()

	manifest := &Manifest{
		Version: "v2",
		Name:    "user-service",
	}

	err := parser.Validate(manifest)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported manifest version")
}

func TestParser_Validate_EmptyName(t *testing.T) {
	parser := NewParser()

	manifest := &Manifest{
		Version: "v1",
		Name:    "",
	}

	err := parser.Validate(manifest)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "component name is required")
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
			name: "valid manifest",
			content: `
version: "v1"
name: "api-gateway"
`,
			expectError: false,
		},
		{
			name: "valid manifest with full component",
			content: `
version: "v1"
id: "api-gateway-v1"
name: "api-gateway"
description: "API Gateway service"
owners:
  maintainers:
    - "alice@company.com"
  team: "Platform"
`,
			expectError: false,
		},
		{
			name: "empty component name",
			content: `
version: "v1"
name: ""
`,
			expectError: true,
			expectedMsg: "component name is required",
		},
		{
			name: "missing version",
			content: `
name: "api-gateway"
`,
			expectError: true,
			expectedMsg: "manifest version is required",
		},
		{
			name: "unsupported version",
			content: `
version: "v2"
name: "api-gateway"
`,
			expectError: true,
			expectedMsg: "unsupported manifest version",
		},
		{
			name:        "malformed yaml",
			content:     `version: "v1" component: [invalid`,
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
				assert.Equal(t, "v1", manifest.Version)
				assert.NotEmpty(t, manifest.Name)
			}
		})
	}
}

func TestComponent_GetIdentifier(t *testing.T) {
	tests := []struct {
		name      string
		component Component
		expected  string
	}{
		{
			name: "with ID",
			component: Component{
				ID:   "unique-id-123",
				Name: "service-name",
			},
			expected: "unique-id-123",
		},
		{
			name: "without ID, uses name",
			component: Component{
				Name: "service-name",
			},
			expected: "service-name",
		},
		{
			name: "empty ID, uses name",
			component: Component{
				ID:   "",
				Name: "service-name",
			},
			expected: "service-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.component.GetIdentifier()
			assert.Equal(t, tt.expected, result)
		})
	}
}
