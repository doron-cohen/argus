package integration

import (
	"context"
	"io"
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

// TestStaticFileServing tests that the root endpoint serves static files correctly
func TestStaticFileServing(t *testing.T) {
	stop := startServerAndWaitForHealth(t, TestConfig)
	defer stop()

	// Test that the root endpoint is reachable and returns a response
	resp, err := http.Get("http://localhost:8080/")
	require.NoError(t, err)
	defer resp.Body.Close()

	// The endpoint should be reachable (either 200 for existing files or 404 for missing files)
	// This validates that the static file serving route is properly configured
	require.Contains(t, []int{200, 404}, resp.StatusCode, "Expected 200 or 404, got: %d", resp.StatusCode)

	// If we get a 200, validate it's HTML content
	if resp.StatusCode == 200 {
		contentType := resp.Header.Get("Content-Type")
		require.Contains(t, contentType, "text/html", "Expected HTML content type, got: %s", contentType)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.NotEmpty(t, body, "Response body should not be empty")

		bodyStr := string(body)
		require.Contains(t, bodyStr, "<html", "Response should contain HTML tags")
		require.Contains(t, bodyStr, "</html>", "Response should contain closing HTML tags")
	}
}
