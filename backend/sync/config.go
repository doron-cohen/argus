package sync

import (
	"fmt"
	"time"
)

// Config holds the sync module configuration
type Config struct {
	Sources []SourceConfig `fig:"sources"`
}

// SourceConfig defines a single source to sync from with type-specific configuration
type SourceConfig struct {
	Type     string        `fig:"type"`                  // "git" or "filesystem"
	Interval time.Duration `fig:"interval" default:"5m"` // How often to sync this source

	// Git-specific configuration (only used when Type="git")
	URL    string `fig:"url,omitempty"`    // Git repository URL
	Branch string `fig:"branch,omitempty"` // Git branch to sync from

	// Filesystem-specific configuration (only used when Type="filesystem")
	Path string `fig:"path,omitempty"` // Local filesystem path

	// Common configuration (used by both git and filesystem)
	BasePath string `fig:"base_path,omitempty"` // Optional: subdirectory to sync from
}

// GitConfig returns git-specific configuration, validates required fields
func (s SourceConfig) GitConfig() (GitSourceConfig, error) {
	if s.Type != "git" {
		return GitSourceConfig{}, fmt.Errorf("source type is %s, not git", s.Type)
	}
	if s.URL == "" {
		return GitSourceConfig{}, fmt.Errorf("git source requires url field")
	}

	branch := s.Branch
	if branch == "" {
		branch = "main" // Default branch
	}

	return GitSourceConfig{
		URL:      s.URL,
		Branch:   branch,
		BasePath: s.BasePath,
	}, nil
}

// FilesystemConfig returns filesystem-specific configuration, validates required fields
func (s SourceConfig) FilesystemConfig() (FilesystemSourceConfig, error) {
	if s.Type != "filesystem" {
		return FilesystemSourceConfig{}, fmt.Errorf("source type is %s, not filesystem", s.Type)
	}
	if s.Path == "" {
		return FilesystemSourceConfig{}, fmt.Errorf("filesystem source requires path field")
	}

	return FilesystemSourceConfig{
		Path:     s.Path,
		BasePath: s.BasePath,
	}, nil
}

// GitSourceConfig holds git-specific configuration
type GitSourceConfig struct {
	URL      string // Repository URL
	Branch   string // Git branch to sync from
	BasePath string // Optional: subdirectory to sync from
}

// FilesystemSourceConfig holds filesystem-specific configuration
type FilesystemSourceConfig struct {
	Path     string // Local filesystem path
	BasePath string // Optional: subdirectory to sync from
}
