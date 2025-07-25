package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/doron-cohen/argus/backend/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncAPIServer_GetSyncSources(t *testing.T) {
	// Create a mock sync service
	service := &sync.Service{}

	// Create API server
	server := NewSyncAPIServer(service)

	// Create request
	req := httptest.NewRequest("GET", "/sources", nil)
	w := httptest.NewRecorder()

	// Call handler
	server.GetSyncSources(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response []SyncSource
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should return empty array when no sources configured
	assert.Empty(t, response)
}

func TestSyncAPIServer_GetSyncSource(t *testing.T) {
	// Create a mock sync service
	service := &sync.Service{}

	// Create API server
	server := NewSyncAPIServer(service)

	// Test with non-existent ID
	req := httptest.NewRequest("GET", "/sources/0", nil)
	w := httptest.NewRecorder()

	server.GetSyncSource(w, req, 0)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var errorResponse Error
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "SOURCE_NOT_FOUND", *errorResponse.Code)
}

func TestSyncAPIServer_GetSyncSourceStatus(t *testing.T) {
	// Create a mock sync service
	service := &sync.Service{}

	// Create API server
	server := NewSyncAPIServer(service)

	// Test with non-existent ID
	req := httptest.NewRequest("GET", "/sources/0/status", nil)
	w := httptest.NewRecorder()

	server.GetSyncSourceStatus(w, req, 0)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var errorResponse Error
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "SOURCE_NOT_FOUND", *errorResponse.Code)
}

func TestSyncAPIServer_TriggerSyncSource(t *testing.T) {
	// Create a mock sync service
	service := &sync.Service{}

	// Create API server
	server := NewSyncAPIServer(service)

	// Test with non-existent ID
	req := httptest.NewRequest("POST", "/sources/0/trigger", nil)
	w := httptest.NewRecorder()

	server.TriggerSyncSource(w, req, 0)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var errorResponse Error
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "SOURCE_NOT_FOUND", *errorResponse.Code)
}

func TestSyncAPIServer_convertToAPIStatus(t *testing.T) {
	server := &SyncAPIServer{}

	// Test with nil status
	apiStatus := server.convertToAPIStatus(nil, 0)
	assert.Equal(t, 0, *apiStatus.SourceId)
	assert.Equal(t, Idle, *apiStatus.Status)

	// Test with completed status
	now := time.Now()
	errorMsg := "test error"
	status := &sync.SourceStatus{
		Status:          sync.StatusCompleted,
		LastSync:        &now,
		LastError:       &errorMsg,
		ComponentsCount: 5,
		Duration:        10 * time.Second,
	}

	apiStatus = server.convertToAPIStatus(status, 1)
	assert.Equal(t, 1, *apiStatus.SourceId)
	assert.Equal(t, Completed, *apiStatus.Status)
	assert.Equal(t, &now, apiStatus.LastSync)
	assert.Equal(t, &errorMsg, apiStatus.LastError)
	assert.Equal(t, 5, *apiStatus.ComponentsCount)
	duration := "10s"
	assert.Equal(t, &duration, apiStatus.Duration)
}
