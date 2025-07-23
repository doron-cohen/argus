package storage

import (
	"fmt"
)

type Config struct {
	Host     string `fig:"host" default:"localhost"`
	Port     int    `fig:"port" default:"5432"`
	User     string `fig:"user" default:"postgres"`
	Password string `fig:"password" default:"postgres"`
	DBName   string `fig:"dbname" default:"postgres"`
	SSLMode  string `fig:"sslmode" default:"disable"`
}

func (c Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}
