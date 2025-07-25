package sync

import (
	"context"
	"testing"
	"time"

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

func TestManifestClient_findFiles(t *testing.T) {
	manifestClient := NewManifestClient()

	// Test with current directory (should find our test files)
	files, err := manifestClient.findFiles(".", "git_test.go")

	require.NoError(t, err)
	assert.Contains(t, files, "git_test.go")
}

func TestGitSourceConfig_BasePath(t *testing.T) {
	tests := []struct {
		name     string
		config   GitSourceConfig
		expected string
	}{
		{
			name: "no base path",
			config: GitSourceConfig{
				URL:    "https://github.com/user/repo",
				Branch: "main",
			},
			expected: "",
		},
		{
			name: "with base path",
			config: GitSourceConfig{
				URL:      "https://github.com/user/monorepo",
				Branch:   "main",
				BasePath: "services/api",
			},
			expected: "services/api",
		},
		{
			name: "base path with leading slash",
			config: GitSourceConfig{
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

func TestSourceConfig_GitConfig(t *testing.T) {
	tests := []struct {
		name        string
		source      SourceConfig
		expectError bool
		expected    GitSourceConfig
	}{
		{
			name: "valid git config",
			source: SourceConfig{
				Type: "git",
				URL:  "https://github.com/user/repo",
			},
			expectError: false,
			expected: GitSourceConfig{
				URL:    "https://github.com/user/repo",
				Branch: "main", // default
			},
		},
		{
			name: "git config with custom branch",
			source: SourceConfig{
				Type:   "git",
				URL:    "https://github.com/user/repo",
				Branch: "develop",
			},
			expectError: false,
			expected: GitSourceConfig{
				URL:    "https://github.com/user/repo",
				Branch: "develop",
			},
		},
		{
			name: "git config with valid interval",
			source: SourceConfig{
				Type:     "git",
				URL:      "https://github.com/user/repo",
				Interval: 30 * time.Second,
			},
			expectError: false,
			expected: GitSourceConfig{
				URL:    "https://github.com/user/repo",
				Branch: "main",
			},
		},
		{
			name: "git config with interval too low",
			source: SourceConfig{
				Type:     "git",
				URL:      "https://github.com/user/repo",
				Interval: 5 * time.Second, // Below 10 second minimum
			},
			expectError: true,
		},
		{
			name: "git config with interval way too low",
			source: SourceConfig{
				Type:     "git",
				URL:      "https://github.com/user/repo",
				Interval: 100 * time.Millisecond,
			},
			expectError: true,
		},
		{
			name: "wrong type",
			source: SourceConfig{
				Type: "filesystem",
				Path: "/some/path",
			},
			expectError: true,
		},
		{
			name: "missing URL",
			source: SourceConfig{
				Type: "git",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gitConfig, err := tt.source.GitConfig()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, gitConfig)
			}
		})
	}
}

func TestGitClient_FindManifests_WithBasePath(t *testing.T) {
	client := NewGitClient()
	ctx := context.Background()

	// Test with a non-existent base path - should return error
	gitConfig := GitSourceConfig{
		URL:      "invalid-url",
		Branch:   "main",
		BasePath: "non-existent-path",
	}

	manifests, err := client.FindManifests(ctx, gitConfig)
	assert.Error(t, err)
	assert.Nil(t, manifests)
	assert.Contains(t, err.Error(), "failed to ensure repository")
}

// Note: These tests would require actual git repositories to test fully.
// In a real test suite, you might use test fixtures or temporary git repos.
func TestGitClient_ErrorCases(t *testing.T) {
	client := NewGitClient()
	ctx := context.Background()

	// Test with invalid git config
	invalidGitConfig := GitSourceConfig{
		URL:    "invalid-url",
		Branch: "main",
	}

	t.Run("invalid repository URL", func(t *testing.T) {
		manifests, err := client.FindManifests(ctx, invalidGitConfig)
		assert.Error(t, err)
		assert.Nil(t, manifests)
	})

	t.Run("get file content from invalid repo", func(t *testing.T) {
		content, err := client.GetFileContent(ctx, invalidGitConfig, "test.txt")
		assert.Error(t, err)
		assert.Nil(t, content)
	})

	t.Run("get latest commit from invalid repo", func(t *testing.T) {
		commit, err := client.GetLatestCommit(ctx, invalidGitConfig)
		assert.Error(t, err)
		assert.Empty(t, commit)
	})

	t.Run("invalid repository URL with base path", func(t *testing.T) {
		gitConfigWithBasePath := invalidGitConfig
		gitConfigWithBasePath.BasePath = "some/path"

		manifests, err := client.FindManifests(ctx, gitConfigWithBasePath)
		assert.Error(t, err)
		assert.Nil(t, manifests)
	})
}
