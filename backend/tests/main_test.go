package integration

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/doron-cohen/argus/backend/internal/config"
	"github.com/doron-cohen/argus/backend/internal/server"
	"github.com/doron-cohen/argus/backend/internal/storage"
	"github.com/stretchr/testify/require"
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

// startServerAndWaitForHealth starts the server and waits for health endpoint to return 200
func startServerAndWaitForHealth(t *testing.T, cfg config.Config) func() {
	t.Helper()

	stop, err := server.Start(cfg)
	require.NoError(t, err)

	// Wait for server to be ready
	maxWait := 10 * time.Second
	startTime := time.Now()

	for time.Since(startTime) < maxWait {
		// Check if server is ready
		resp, err := http.Get("http://localhost:8080/healthz")
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if err := resp.Body.Close(); err != nil {
			t.Logf("Failed to close response body: %v", err)
		}
		if resp.StatusCode == http.StatusOK {
			return stop
		}
		time.Sleep(100 * time.Millisecond)
	}

	// If we get here, server didn't become ready in time
	stop()
	t.Fatal("Server did not become ready in time")
	return nil
}
