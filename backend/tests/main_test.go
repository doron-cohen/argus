package integration

import (
	"context"
	"os"
	"testing"

	"github.com/doron-cohen/argus/backend/internal/config"
	"github.com/doron-cohen/argus/backend/internal/storage"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	pgContainer, err := postgres.Run(ctx, "postgres:16",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)
	if err != nil {
		panic(err)
	}
	host, err := pgContainer.Host(ctx)
	if err != nil {
		panic(err)
	}
	port, err := pgContainer.MappedPort(ctx, "5432")
	if err != nil {
		panic(err)
	}

	TestConfig = config.Config{
		Storage: storage.Config{
			Host:     host,
			Port:     port.Int(),
			User:     "testuser",
			Password: "testpass",
			DBName:   "testdb",
			SSLMode:  "disable",
		},
	}

	code := m.Run()
	_ = pgContainer.Terminate(ctx)
	os.Exit(code)
}
