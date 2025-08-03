package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/doron-cohen/argus/backend/internal/storage"
	"github.com/doron-cohen/argus/backend/sync"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Storage storage.Config `yaml:"storage"`
	Sync    sync.Config    `yaml:"sync"`
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() Config {
	return Config{
		Storage: storage.Config{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			DBName:   "argus",
			SSLMode:  "disable",
		},
		Sync: sync.Config{
			Sources: []sync.SourceConfig{},
		},
	}
}

// LoadConfig loads configuration with the following priority:
// 1. Environment variables (highest priority)
// 2. Config file values (if file exists)
// 3. Default values (lowest priority)
func LoadConfig() (Config, error) {
	// Start with defaults
	cfg := DefaultConfig()

	// Determine config file path
	configPath := "config.yaml"
	if envPath := os.Getenv("ARGUS_CONFIG_PATH"); envPath != "" {
		configPath = envPath
	}

	// Try to load config file (optional)
	if data, err := os.ReadFile(configPath); err == nil {
		// Parse the YAML
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("failed to parse config file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		// Only return error if it's not a "file not found" error
		return cfg, fmt.Errorf("failed to read config file: %w", err)
	}

	// Override with environment variables
	cfg = overrideWithEnvironment(cfg)

	return cfg, nil
}

// overrideWithEnvironment overrides config values with environment variables
func overrideWithEnvironment(cfg Config) Config {
	// Storage configuration
	if val := os.Getenv("ARGUS_STORAGE_HOST"); val != "" {
		cfg.Storage.Host = val
	}
	if val := os.Getenv("ARGUS_STORAGE_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			cfg.Storage.Port = port
		}
	}
	if val := os.Getenv("ARGUS_STORAGE_USER"); val != "" {
		cfg.Storage.User = val
	}
	if val := os.Getenv("ARGUS_STORAGE_PASSWORD"); val != "" {
		cfg.Storage.Password = val
	}
	if val := os.Getenv("ARGUS_STORAGE_DBNAME"); val != "" {
		cfg.Storage.DBName = val
	}
	if val := os.Getenv("ARGUS_STORAGE_SSLMODE"); val != "" {
		cfg.Storage.SSLMode = val
	}

	// Note: Sync sources are not overridden by environment variables
	// as they require complex configuration that's better handled via config files

	return cfg
}

// GetEnvironmentVariables returns a map of all ARGUS_ environment variables
// This is useful for debugging and documentation
func GetEnvironmentVariables() map[string]string {
	envVars := make(map[string]string)
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "ARGUS_") {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				envVars[parts[0]] = parts[1]
			}
		}
	}
	return envVars
}
