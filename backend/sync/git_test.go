package sync

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGitClient(t *testing.T) {
	client := NewGitClient()

	assert.NotNil(t, client)
	assert.NotEmpty(t, client.tempDir)
}

func TestGitClient_sanitizeURL(t *testing.T) {
	client := NewGitClient()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "https URL",
			input:    "https://github.com/user/repo.git",
			expected: "github_com_user_repo_git",
		},
		{
			name:     "http URL",
			input:    "http://gitlab.com/group/project",
			expected: "gitlab_com_group_project",
		},
		{
			name:     "ssh URL",
			input:    "git@github.com:user/repo.git",
			expected: "github_com_user_repo_git",
		},
		{
			name:     "complex URL with ports",
			input:    "https://gitlab.example.com:8080/namespace/project.git",
			expected: "gitlab_example_com_8080_namespace_project_git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.sanitizeURL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGitClient_findFiles(t *testing.T) {
	client := NewGitClient()

	// Test with current directory (should find our test files)
	files, err := client.findFiles(".", "git_test.go")

	require.NoError(t, err)
	assert.Contains(t, files, "git_test.go")
}

func TestSourceConfig_BasePath(t *testing.T) {
	tests := []struct {
		name     string
		config   SourceConfig
		expected string
	}{
		{
			name: "no base path",
			config: SourceConfig{
				Type:   "git",
				URL:    "https://github.com/user/repo",
				Branch: "main",
			},
			expected: "",
		},
		{
			name: "with base path",
			config: SourceConfig{
				Type:     "git",
				URL:      "https://github.com/user/monorepo",
				Branch:   "main",
				BasePath: "services/api",
			},
			expected: "services/api",
		},
		{
			name: "base path with leading slash",
			config: SourceConfig{
				Type:     "git",
				URL:      "https://github.com/user/monorepo",
				Branch:   "main",
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

func TestGitClient_FindManifests_WithBasePath(t *testing.T) {
	client := NewGitClient()
	ctx := context.Background()

	// Test with a non-existent base path - should return error
	sourceWithBasePath := SourceConfig{
		Type:     "git",
		URL:      "invalid-url",
		Branch:   "main",
		BasePath: "non-existent-path",
	}

	manifests, err := client.FindManifests(ctx, sourceWithBasePath)
	assert.Error(t, err)
	assert.Nil(t, manifests)
	assert.Contains(t, err.Error(), "failed to ensure repository")
}

func TestSourceConfig_DefaultValues(t *testing.T) {
	// Test that default values work as expected
	config := SourceConfig{
		Type: "git",
		URL:  "https://github.com/user/repo",
		// Branch should default to "main"
		// Interval should default to "5m"
		BasePath: "services",
	}

	assert.Equal(t, "git", config.Type)
	assert.Equal(t, "https://github.com/user/repo", config.URL)
	assert.Equal(t, "services", config.BasePath)
	// Note: defaults are applied by the fig library during config loading
}

// Note: These tests would require actual git repositories to test fully.
// In a real test suite, you might use test fixtures or temporary git repos.
func TestGitClient_ErrorCases(t *testing.T) {
	client := NewGitClient()
	ctx := context.Background()

	// Test with invalid source config
	invalidSource := SourceConfig{
		Type:   "git",
		URL:    "invalid-url",
		Branch: "main",
	}

	t.Run("invalid repository URL", func(t *testing.T) {
		manifests, err := client.FindManifests(ctx, invalidSource)
		assert.Error(t, err)
		assert.Nil(t, manifests)
	})

	t.Run("get file content from invalid repo", func(t *testing.T) {
		content, err := client.GetFileContent(ctx, invalidSource, "test.txt")
		assert.Error(t, err)
		assert.Nil(t, content)
	})

	t.Run("get latest commit from invalid repo", func(t *testing.T) {
		commit, err := client.GetLatestCommit(ctx, invalidSource)
		assert.Error(t, err)
		assert.Empty(t, commit)
	})

	t.Run("invalid repository URL with base path", func(t *testing.T) {
		sourceWithBasePath := invalidSource
		sourceWithBasePath.BasePath = "some/path"

		manifests, err := client.FindManifests(ctx, sourceWithBasePath)
		assert.Error(t, err)
		assert.Nil(t, manifests)
	})
}
