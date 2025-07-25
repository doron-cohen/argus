package sync

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestSourceConfig_GitConfig(t *testing.T) {
	tests := []struct {
		name        string
		yamlSource  string
		expectError bool
		expected    GitSourceConfig
	}{
		{
			name: "valid git config",
			yamlSource: `type: git
url: https://github.com/user/repo`,
			expectError: false,
			expected: GitSourceConfig{
				Type: "git",
				URL:  "https://github.com/user/repo",
			},
		},
		{
			name: "git config with custom branch",
			yamlSource: `type: git
url: https://github.com/user/repo
branch: develop`,
			expectError: false,
			expected: GitSourceConfig{
				Type:   "git",
				URL:    "https://github.com/user/repo",
				Branch: "develop",
			},
		},
		{
			name: "git config with valid interval",
			yamlSource: `type: git
url: https://github.com/user/repo
interval: 10s`,
			expectError: false,
			expected: GitSourceConfig{
				Type:     "git",
				URL:      "https://github.com/user/repo",
				Interval: 10 * time.Second,
			},
		},
		{
			name: "git config with interval too low",
			yamlSource: `type: git
url: https://github.com/user/repo
interval: 500ms`,
			expectError: true,
		},
		{
			name: "git config with interval way too low",
			yamlSource: `type: git
url: https://github.com/user/repo
interval: 100ms`,
			expectError: true,
		},
		{
			name: "wrong type",
			yamlSource: `type: filesystem
path: /some/path`,
			expectError: true,
		},
		{
			name:        "missing URL",
			yamlSource:  `type: git`,
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
				_, ok := cfg.(*GitSourceConfig)
				assert.False(t, ok, "Expected type assertion to fail for wrong type")
				return
			}
			assert.NoError(t, err)
			cfg := source.GetConfig()
			gitConfig, ok := cfg.(*GitSourceConfig)
			assert.True(t, ok)
			assert.Equal(t, tt.expected.Type, gitConfig.Type)
			assert.Equal(t, tt.expected.URL, gitConfig.URL)
			if tt.expected.Branch != "" {
				assert.Equal(t, tt.expected.Branch, gitConfig.Branch)
			}
			if tt.expected.Interval > 0 {
				assert.Equal(t, tt.expected.Interval, gitConfig.Interval)
			}
		})
	}
}

func TestGitSourceConfig_BasePath(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		expected string
	}{
		{
			name:     "no base path",
			basePath: "",
			expected: "",
		},
		{
			name:     "with base path",
			basePath: "services",
			expected: "services",
		},
		{
			name:     "base path with leading slash",
			basePath: "/services",
			expected: "/services",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GitSourceConfig{
				BasePath: tt.basePath,
			}
			assert.Equal(t, tt.expected, config.GetBasePath())
		})
	}
}

func TestGitFetcher_sanitizeURL(t *testing.T) {
	fetcher := &GitFetcher{}

	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "https URL",
			url:      "https://github.com/user/repo",
			expected: "github_com_user_repo",
		},
		{
			name:     "http URL",
			url:      "http://github.com/user/repo",
			expected: "github_com_user_repo",
		},
		{
			name:     "ssh URL",
			url:      "git@github.com:user/repo.git",
			expected: "github_com_user_repo_git",
		},
		{
			name:     "complex URL with ports",
			url:      "https://git.example.com:8443/user/repo",
			expected: "git_example_com_8443_user_repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fetcher.sanitizeURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewFetcher_GitType(t *testing.T) {
	fetcher, err := NewFetcher("git")

	require.NoError(t, err)
	assert.NotNil(t, fetcher)
	assert.IsType(t, &GitFetcher{}, fetcher)
}
