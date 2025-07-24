package sync

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFilesystemClient(t *testing.T) {
	client := NewFilesystemClient()

	assert.NotNil(t, client)
	assert.NotNil(t, client.manifestClient)
}

func TestFilesystemSourceConfig_BasePath(t *testing.T) {
	tests := []struct {
		name     string
		config   FilesystemSourceConfig
		expected string
	}{
		{
			name: "no base path",
			config: FilesystemSourceConfig{
				Path: "/some/path",
			},
			expected: "",
		},
		{
			name: "with base path",
			config: FilesystemSourceConfig{
				Path:     "/some/monorepo",
				BasePath: "services/api",
			},
			expected: "services/api",
		},
		{
			name: "base path with leading slash",
			config: FilesystemSourceConfig{
				Path:     "/some/monorepo",
				BasePath: "/microservices/auth",
			},
			expected: "/microservices/auth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.config.BasePath)
		})
	}
}

func TestSourceConfig_FilesystemConfig(t *testing.T) {
	tests := []struct {
		name        string
		source      SourceConfig
		expectError bool
		expected    FilesystemSourceConfig
	}{
		{
			name: "valid filesystem config",
			source: SourceConfig{
				Type: "filesystem",
				Path: "/some/path",
			},
			expectError: false,
			expected: FilesystemSourceConfig{
				Path: "/some/path",
			},
		},
		{
			name: "filesystem config with base path",
			source: SourceConfig{
				Type:     "filesystem",
				Path:     "/some/path",
				BasePath: "services",
			},
			expectError: false,
			expected: FilesystemSourceConfig{
				Path:     "/some/path",
				BasePath: "services",
			},
		},
		{
			name: "wrong type",
			source: SourceConfig{
				Type: "git",
				URL:  "https://github.com/user/repo",
			},
			expectError: true,
		},
		{
			name: "missing path",
			source: SourceConfig{
				Type: "filesystem",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsConfig, err := tt.source.FilesystemConfig()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, fsConfig)
			}
		})
	}
}

func TestFilesystemClient_WithTestData(t *testing.T) {
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
	authManifest := `name: "auth-service"`
	apiManifest := `name: "api-gateway"`
	infraManifest := `name: "platform-infrastructure"`

	require.NoError(t, os.WriteFile(filepath.Join(authDir, "manifest.yaml"), []byte(authManifest), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(apiDir, "manifest.yml"), []byte(apiManifest), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(infraDir, "manifest.yaml"), []byte(infraManifest), 0644))

	client := NewFilesystemClient()
	ctx := context.Background()

	t.Run("find manifests in entire directory", func(t *testing.T) {
		config := FilesystemSourceConfig{
			Path: tempDir,
		}

		manifests, err := client.FindManifests(ctx, config)
		require.NoError(t, err)

		assert.Len(t, manifests, 3, "Should find 3 manifest files")
		assert.Contains(t, manifests, filepath.Join("services", "auth", "manifest.yaml"))
		assert.Contains(t, manifests, filepath.Join("services", "api", "manifest.yml"))
		assert.Contains(t, manifests, filepath.Join("platform", "infrastructure", "manifest.yaml"))
	})

	t.Run("find manifests with base path - services only", func(t *testing.T) {
		config := FilesystemSourceConfig{
			Path:     tempDir,
			BasePath: "services",
		}

		manifests, err := client.FindManifests(ctx, config)
		require.NoError(t, err)

		assert.Len(t, manifests, 2, "Should find 2 manifest files in services")
		assert.Contains(t, manifests, filepath.Join("services", "auth", "manifest.yaml"))
		assert.Contains(t, manifests, filepath.Join("services", "api", "manifest.yml"))
		assert.NotContains(t, manifests, filepath.Join("platform", "infrastructure", "manifest.yaml"))
	})

	t.Run("find manifests with base path - platform only", func(t *testing.T) {
		config := FilesystemSourceConfig{
			Path:     tempDir,
			BasePath: "platform",
		}

		manifests, err := client.FindManifests(ctx, config)
		require.NoError(t, err)

		assert.Len(t, manifests, 1, "Should find 1 manifest file in platform")
		assert.Contains(t, manifests, filepath.Join("platform", "infrastructure", "manifest.yaml"))
	})

	t.Run("read file content", func(t *testing.T) {
		config := FilesystemSourceConfig{
			Path: tempDir,
		}

		content, err := client.GetFileContent(ctx, config, filepath.Join("services", "auth", "manifest.yaml"))
		require.NoError(t, err)

		assert.Equal(t, authManifest, string(content))
	})

	t.Run("get last modified", func(t *testing.T) {
		config := FilesystemSourceConfig{
			Path: tempDir,
		}

		lastModified, err := client.GetLastModified(ctx, config)
		require.NoError(t, err)

		assert.Contains(t, lastModified, "filesystem:")
		assert.Contains(t, lastModified, tempDir)
	})
}

func TestFilesystemClient_ErrorCases(t *testing.T) {
	client := NewFilesystemClient()
	ctx := context.Background()

	t.Run("non-existent path", func(t *testing.T) {
		config := FilesystemSourceConfig{
			Path: "/non/existent/path",
		}

		manifests, err := client.FindManifests(ctx, config)
		assert.Error(t, err)
		assert.Nil(t, manifests)
		// The error could be about path resolution or file walking failure
		assert.True(t,
			err.Error() == "filesystem path /non/existent/path does not exist" ||
				strings.Contains(err.Error(), "no such file or directory") ||
				strings.Contains(err.Error(), "does not exist"),
			"Error should indicate path doesn't exist, got: %s", err.Error())
	})

	t.Run("non-existent base path", func(t *testing.T) {
		tempDir := t.TempDir()

		config := FilesystemSourceConfig{
			Path:     tempDir,
			BasePath: "non-existent",
		}

		manifests, err := client.FindManifests(ctx, config)
		assert.Error(t, err)
		assert.Nil(t, manifests)
		assert.Contains(t, err.Error(), "base path non-existent does not exist")
	})

	t.Run("read non-existent file", func(t *testing.T) {
		tempDir := t.TempDir()

		config := FilesystemSourceConfig{
			Path: tempDir,
		}

		content, err := client.GetFileContent(ctx, config, "non-existent.yaml")
		assert.Error(t, err)
		assert.Nil(t, content)
	})
}

func TestFilesystemFetcher(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()

	// Create test directory structure
	servicesDir := filepath.Join(tempDir, "services")
	authDir := filepath.Join(servicesDir, "auth")
	require.NoError(t, os.MkdirAll(authDir, 0755))

	// Create test manifest file
	authManifest := `name: "auth-service"`
	require.NoError(t, os.WriteFile(filepath.Join(authDir, "manifest.yaml"), []byte(authManifest), 0644))

	fetcher := NewFilesystemFetcher()
	ctx := context.Background()

	t.Run("fetch components successfully", func(t *testing.T) {
		source := SourceConfig{
			Type: "filesystem",
			Path: tempDir,
		}

		components, err := fetcher.Fetch(ctx, source)
		require.NoError(t, err)

		assert.Len(t, components, 1, "Should find 1 component")
		assert.Equal(t, "auth-service", components[0].Name)
	})

	t.Run("invalid filesystem config", func(t *testing.T) {
		source := SourceConfig{
			Type: "git", // Wrong type
			URL:  "https://github.com/user/repo",
		}

		components, err := fetcher.Fetch(ctx, source)
		assert.Error(t, err)
		assert.Nil(t, components)
		assert.Contains(t, err.Error(), "invalid filesystem configuration")
	})
}

func TestNewFetcher_FilesystemType(t *testing.T) {
	fetcher, err := NewFetcher("filesystem")

	require.NoError(t, err)
	assert.NotNil(t, fetcher)
	assert.IsType(t, &FilesystemFetcher{}, fetcher)
}

func TestManifestClient_SharedFunctionality(t *testing.T) {
	manifestClient := NewManifestClient()
	tempDir := t.TempDir()

	// Create test files
	authDir := filepath.Join(tempDir, "auth")
	require.NoError(t, os.MkdirAll(authDir, 0755))

	authContent := `name: "auth-service"`
	authFile := filepath.Join(authDir, "manifest.yaml")
	require.NoError(t, os.WriteFile(authFile, []byte(authContent), 0644))

	t.Run("find manifests", func(t *testing.T) {
		manifests, err := manifestClient.FindManifests(tempDir, "")
		require.NoError(t, err)

		assert.Len(t, manifests, 1)
		assert.Contains(t, manifests, filepath.Join("auth", "manifest.yaml"))
	})

	t.Run("find manifests with base path", func(t *testing.T) {
		// Create subdirectory
		subDir := filepath.Join(tempDir, "services")
		subAuthDir := filepath.Join(subDir, "auth")
		require.NoError(t, os.MkdirAll(subAuthDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(subAuthDir, "manifest.yml"), []byte(authContent), 0644))

		manifests, err := manifestClient.FindManifests(tempDir, "services")
		require.NoError(t, err)

		assert.Len(t, manifests, 1)
		assert.Contains(t, manifests, filepath.Join("services", "auth", "manifest.yml"))
	})

	t.Run("get file content", func(t *testing.T) {
		content, err := manifestClient.GetFileContent(tempDir, filepath.Join("auth", "manifest.yaml"))
		require.NoError(t, err)

		assert.Equal(t, authContent, string(content))
	})

	t.Run("validate base path", func(t *testing.T) {
		// Valid base path
		err := manifestClient.ValidateBasePath(tempDir, "auth")
		assert.NoError(t, err)

		// Invalid base path
		err = manifestClient.ValidateBasePath(tempDir, "non-existent")
		assert.Error(t, err)

		// Empty base path (should be valid)
		err = manifestClient.ValidateBasePath(tempDir, "")
		assert.NoError(t, err)
	})
}
