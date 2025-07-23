package integration

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/doron-cohen/argus/backend/internal/server"
	"github.com/stretchr/testify/require"
)

func TestHealthzIntegration(t *testing.T) {
	stop, err := server.StartServer()
	require.NoError(t, err)
	defer stop()

	// Wait briefly for the server to start
	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get("http://localhost:8080/healthz")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "ok", string(body))
}
