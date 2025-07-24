package sync

import "time"

// SourceConfig defines a single source to sync from
type SourceConfig struct {
	Type     string        `fig:"type"`                  // "git" for now
	URL      string        `fig:"url"`                   // Repository URL
	Branch   string        `fig:"branch" default:"main"` // Git branch to sync from
	Interval time.Duration `fig:"interval" default:"5m"` // How often to sync this source
	BasePath string        `fig:"base_path"`             // Optional: subdirectory to sync from (saves bandwidth)
}

// Config holds the sync module configuration
type Config struct {
	Sources []SourceConfig `fig:"sources"`
}
