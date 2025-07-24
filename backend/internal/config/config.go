package config

import (
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
	err := fig.Load(&cfg, fig.UseEnv("ARGUS"))
	return cfg, err
}
