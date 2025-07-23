package integration

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/doron-cohen/argus/backend/api/client"
	"github.com/doron-cohen/argus/backend/internal/config"
	"github.com/doron-cohen/argus/backend/internal/server"
	"github.com/stretchr/testify/require"
)

var TestConfig config.Config

func TestGetComponentsIntegration(t *testing.T) {
	stop, err := server.Start(TestConfig)
	require.NoError(t, err)
	defer stop()

	// Wait briefly for the server to start
	time.Sleep(100 * time.Millisecond)

	client, err := client.NewClientWithResponses("http://localhost:8080/api")
	require.NoError(t, err)

	resp, err := client.GetComponentsWithResponse(context.Background())
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.NotNil(t, resp.JSON200)
	require.Len(t, *resp.JSON200, 0)
}
