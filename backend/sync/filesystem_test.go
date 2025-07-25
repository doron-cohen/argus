package sync

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestSourceConfig_FilesystemConfig(t *testing.T) {
	tests := []struct {
		name        string
		yamlSource  string
		expectError bool
		expected    FilesystemSourceConfig
	}{
		{
			name: "valid filesystem config",
			yamlSource: `type: filesystem
path: /some/path`,
			expectError: false,
			expected: FilesystemSourceConfig{
				Type: "filesystem",
				Path: "/some/path",
			},
		},
		{
			name: "filesystem config with valid interval",
			yamlSource: `type: filesystem
path: /some/path
interval: 5s`,
			expectError: false,
			expected: FilesystemSourceConfig{
				Type:     "filesystem",
				Path:     "/some/path",
				Interval: 5 * time.Second,
			},
		},
		{
			name: "filesystem config with interval too low",
			yamlSource: `type: filesystem
path: /some/path
interval: 500ms`,
			expectError: true,
		},
		{
			name: "filesystem config with interval way too low",
			yamlSource: `type: filesystem
path: /some/path
interval: 100ms`,
			expectError: true,
		},
		{
			name: "wrong type",
			yamlSource: `type: git
url: https://github.com/user/repo`,
			expectError: true,
		},
		{
			name:        "missing path",
			yamlSource:  `type: filesystem`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var source SourceConfig
			err := yaml.Unmarshal([]byte(tt.yamlSource), &source)
			if tt.expectError {
				if err != nil {
					assert.Error(t, err)
					return
				}
				// For "wrong type" cases, check if type assertion fails
				cfg := source.GetConfig()
				_, ok := cfg.(*FilesystemSourceConfig)
				assert.False(t, ok, "Expected type assertion to fail for wrong type")
				return
			}
			assert.NoError(t, err)
			cfg := source.GetConfig()
			fsConfig, ok := cfg.(*FilesystemSourceConfig)
			assert.True(t, ok)
			assert.Equal(t, tt.expected.Type, fsConfig.Type)
			assert.Equal(t, tt.expected.Path, fsConfig.Path)
			if tt.expected.Interval > 0 {
				assert.Equal(t, tt.expected.Interval, fsConfig.Interval)
			}
		})
	}
}

func TestLoadManifests(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()

	// Create test directory structure
	servicesDir := filepath.Join(tempDir, "services")
	authDir := filepath.Join(servicesDir, "auth")
	apiDir := filepath.Join(servicesDir, "api")
	platformDir := filepath.Join(tempDir, "platform")
	infraDir := filepath.Join(platformDir, "infrastructure")

	require.NoError(t, os.MkdirAll(authDir, 0755))
	require.NoError(t, os.MkdirAll(apiDir, 0755))
	require.NoError(t, os.MkdirAll(infraDir, 0755))

	// Create test manifest files
	authManifest := `version: "v1"
name: "auth-service"`
	apiManifest := `version: "v1"
name: "api-gateway"`
	infraManifest := `version: "v1"
name: "platform-infrastructure"`

	require.NoError(t, os.WriteFile(filepath.Join(authDir, "manifest.yaml"), []byte(authManifest), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(apiDir, "manifest.yml"), []byte(apiManifest), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(infraDir, "manifest.yaml"), []byte(infraManifest), 0644))

	ctx := context.Background()

	t.Run("load manifests from root directory", func(t *testing.T) {
		manifests, err := LoadManifests(ctx, tempDir)
		require.NoError(t, err)

		// Should find all 3 manifest files
		assert.Len(t, manifests, 3)
		assert.Contains(t, manifests, filepath.Join("services", "auth", "manifest.yaml"))
		assert.Contains(t, manifests, filepath.Join("services", "api", "manifest.yml"))
		assert.Contains(t, manifests, filepath.Join("platform", "infrastructure", "manifest.yaml"))

		// Check that manifests are parsed correctly
		authManifest := manifests[filepath.Join("services", "auth", "manifest.yaml")]
		assert.Equal(t, "auth-service", authManifest.Content.Name)
		assert.Equal(t, "v1", authManifest.Content.Version)

		apiManifest := manifests[filepath.Join("services", "api", "manifest.yml")]
		assert.Equal(t, "api-gateway", apiManifest.Content.Name)
		assert.Equal(t, "v1", apiManifest.Content.Version)

		infraManifest := manifests[filepath.Join("platform", "infrastructure", "manifest.yaml")]
		assert.Equal(t, "platform-infrastructure", infraManifest.Content.Name)
		assert.Equal(t, "v1", infraManifest.Content.Version)
	})

	t.Run("load manifests from subdirectory", func(t *testing.T) {
		manifests, err := LoadManifests(ctx, servicesDir)
		require.NoError(t, err)

		// Should find 2 manifest files in services directory
		assert.Len(t, manifests, 2)
		assert.Contains(t, manifests, filepath.Join("auth", "manifest.yaml"))
		assert.Contains(t, manifests, filepath.Join("api", "manifest.yml"))
		assert.NotContains(t, manifests, filepath.Join("platform", "infrastructure", "manifest.yaml"))
	})

	t.Run("non-existent directory", func(t *testing.T) {
		manifests, err := LoadManifests(ctx, filepath.Join(tempDir, "non-existent"))
		assert.Error(t, err)
		assert.Nil(t, manifests)
		assert.Contains(t, err.Error(), "does not exist")
	})
}

func TestFilesystemFetcher(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()

	// Create test directory structure
	authDir := filepath.Join(tempDir, "auth")
	require.NoError(t, os.MkdirAll(authDir, 0755))

	// Create test manifest file
	authManifest := `version: "v1"
name: "auth-service"`
	require.NoError(t, os.WriteFile(filepath.Join(authDir, "manifest.yaml"), []byte(authManifest), 0644))

	fetcher := NewFilesystemFetcher()
	ctx := context.Background()

	t.Run("fetch components successfully", func(t *testing.T) {
		yamlSource := "type: filesystem\npath: " + tempDir
		var source SourceConfig
		err := yaml.Unmarshal([]byte(yamlSource), &source)
		require.NoError(t, err)

		components, err := fetcher.Fetch(ctx, source)
		require.NoError(t, err)

		assert.Len(t, components, 1, "Should find 1 component")
		assert.Equal(t, "auth-service", components[0].Name)
	})

	t.Run("invalid filesystem config", func(t *testing.T) {
		yamlSource := "type: git\nurl: https://github.com/user/repo"
		var source SourceConfig
		err := yaml.Unmarshal([]byte(yamlSource), &source)
		require.NoError(t, err)

		components, err := fetcher.Fetch(ctx, source)
		assert.Error(t, err)
		assert.Nil(t, components)
		assert.Contains(t, err.Error(), "source is not a filesystem config")
	})
}

func TestNewFetcher_FilesystemType(t *testing.T) {
	fetcher, err := NewFetcher("filesystem")

	require.NoError(t, err)
	assert.NotNil(t, fetcher)
	assert.IsType(t, &FilesystemFetcher{}, fetcher)
}
