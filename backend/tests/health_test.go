package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/doron-cohen/argus/backend/internal/health"
	"github.com/doron-cohen/argus/backend/internal/server"
	"github.com/stretchr/testify/require"
)

func TestHealthzIntegration(t *testing.T) {
	stop, err := server.Start(TestConfig)
	require.NoError(t, err)
	defer stop()

	// Wait briefly for the server to start
	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get("http://localhost:8080/healthz")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var healthResponse health.HealthResponse
	err = json.Unmarshal(body, &healthResponse)
	require.NoError(t, err)

	require.Equal(t, "healthy", healthResponse.Status)
	require.NotEmpty(t, healthResponse.Checks)
	require.Equal(t, "healthy", healthResponse.Checks["database"])
	require.NotEmpty(t, healthResponse.Timestamp)

	// Verify timestamp is in RFC3339 format
	_, err = time.Parse(time.RFC3339, healthResponse.Timestamp)
	require.NoError(t, err)
}
