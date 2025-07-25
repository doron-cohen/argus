package sync

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	// Minimum sync intervals to prevent system overload
	MinFilesystemInterval = time.Second      // 1 second minimum for filesystem sources
	MinGitInterval        = 10 * time.Second // 10 seconds minimum for git sources
)

// Config holds the sync module configuration
type Config struct {
	Sources []SourceConfig `fig:"sources"`
}

// SourceTypeConfig is a regular interface for different source types
type SourceTypeConfig interface {
	Validate() error
	GetInterval() time.Duration
	GetBasePath() string
	GetSourceType() string
}

// SourceConfigConstraint is a type constraint for compile-time type safety
type SourceConfigConstraint interface {
	*GitSourceConfig | *FilesystemSourceConfig
	SourceTypeConfig
}

// TypedSourceConfig provides compile-time type safety for individual source configs
type TypedSourceConfig[T SourceConfigConstraint] struct {
	Config T `fig:",inline" yaml:",inline"`
}

func (t *TypedSourceConfig[T]) GetConfig() T {
	return t.Config
}

func (t *TypedSourceConfig[T]) UnmarshalYAML(node *yaml.Node) error {
	// Unmarshal directly into the generic config
	if err := node.Decode(&t.Config); err != nil {
		return fmt.Errorf("failed to decode source config: %w", err)
	}

	// Validate the configuration
	if err := t.Config.Validate(); err != nil {
		return fmt.Errorf("invalid source config: %w", err)
	}

	return nil
}

// SourceConfig wraps any valid source type configuration
// This is needed for heterogeneous collections since we can't have []TypedSourceConfig[T] with mixed types
type SourceConfig struct {
	config SourceTypeConfig
}

// GetConfig returns the underlying type-specific configuration
func (s *SourceConfig) GetConfig() SourceTypeConfig {
	return s.config
}

// UnmarshalYAML implements custom YAML unmarshaling for SourceConfig
func (s *SourceConfig) UnmarshalYAML(node *yaml.Node) error {
	// First, decode just enough to determine the type
	var typeInfo struct {
		Type string `yaml:"type"`
	}

	if err := node.Decode(&typeInfo); err != nil {
		return fmt.Errorf("failed to decode source type: %w", err)
	}

	// Create the appropriate config type based on the "type" field
	var config SourceTypeConfig
	switch typeInfo.Type {
	case "git":
		config = &GitSourceConfig{}
	case "filesystem":
		config = &FilesystemSourceConfig{}
	default:
		return fmt.Errorf("unknown source type: %s", typeInfo.Type)
	}

	// Unmarshal the full configuration into the specific type
	if err := node.Decode(config); err != nil {
		return fmt.Errorf("failed to decode %s source config: %w", typeInfo.Type, err)
	}

	// Validate the configuration
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid %s source config: %w", typeInfo.Type, err)
	}

	s.config = config
	return nil
}

// MarshalYAML implements custom YAML marshaling for SourceConfig
func (s *SourceConfig) MarshalYAML() (interface{}, error) {
	return s.config, nil
}

// NewSourceConfig creates a new SourceConfig from a SourceTypeConfig
func NewSourceConfig(config SourceTypeConfig) SourceConfig {
	return SourceConfig{config: config}
}

// Factory functions for type-safe creation
func NewGitSourceConfig(url, branch, basePath string, interval time.Duration) TypedSourceConfig[*GitSourceConfig] {
	return TypedSourceConfig[*GitSourceConfig]{
		Config: &GitSourceConfig{
			Type:     "git",
			URL:      url,
			Branch:   branch,
			BasePath: basePath,
			Interval: interval,
		},
	}
}

func NewFilesystemSourceConfig(path, basePath string, interval time.Duration) TypedSourceConfig[*FilesystemSourceConfig] {
	return TypedSourceConfig[*FilesystemSourceConfig]{
		Config: &FilesystemSourceConfig{
			Type:     "filesystem",
			Path:     path,
			BasePath: basePath,
			Interval: interval,
		},
	}
}
