package config

import (
	"os"

	"github.com/doron-cohen/argus/backend/internal/storage"
	"github.com/doron-cohen/argus/backend/sync"
	"github.com/kkyr/fig"
)

type Config struct {
	Storage storage.Config
	Sync    sync.Config
}

func LoadConfig() (Config, error) {
	var cfg Config

	// Check if config file path is provided via environment variable
	if configPath := os.Getenv("ARGUS_CONFIG_PATH"); configPath != "" {
		err := fig.Load(&cfg,
			fig.File(configPath),
			fig.UseEnv("ARGUS"),
		)
		return cfg, err
	}

	// Default behavior - look for config.yaml in current directory
	err := fig.Load(&cfg, fig.UseEnv("ARGUS"))
	return cfg, err
}
