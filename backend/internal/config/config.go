package config

import (
	"os"

	"github.com/doron-cohen/argus/backend/internal/storage"
	"github.com/doron-cohen/argus/backend/sync"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Storage storage.Config `yaml:"storage"`
	Sync    sync.Config    `yaml:"sync"`
}

func LoadConfig() (Config, error) {
	var cfg Config

	// Determine config file path
	configPath := "config.yaml"
	if envPath := os.Getenv("ARGUS_CONFIG_PATH"); envPath != "" {
		configPath = envPath
	}

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return cfg, err
	}

	// Parse the YAML
	err = yaml.Unmarshal(data, &cfg)
	return cfg, err
}
