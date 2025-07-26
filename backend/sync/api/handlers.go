package api

import (
	"encoding/json"
	"net/http"

	"github.com/doron-cohen/argus/backend/sync"
)

type SyncAPIServer struct {
	Service *sync.Service
}

func NewSyncAPIServer(service *sync.Service) ServerInterface {
	return &SyncAPIServer{Service: service}
}

func (s *SyncAPIServer) GetSyncSources(w http.ResponseWriter, r *http.Request) {
	// Get all sources from the service
	sources := s.Service.GetSources()

	var apiSources []SyncSource
	for i, source := range sources {
		apiSource := s.convertToAPISource(source, int64(i))
		apiSources = append(apiSources, apiSource)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(apiSources); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *SyncAPIServer) GetSyncSource(w http.ResponseWriter, r *http.Request, id int) {
	// Get source by index
	source, err := s.Service.GetSourceByIndex(id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "Source not found", "SOURCE_NOT_FOUND")
		return
	}

	apiSource := s.convertToAPISource(source, int64(id))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(apiSource); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *SyncAPIServer) GetSyncSourceStatus(w http.ResponseWriter, r *http.Request, id int) {
	// Get status for the source
	status, err := s.Service.GetSourceStatus(id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "Source not found", "SOURCE_NOT_FOUND")
		return
	}

	apiStatus := s.convertToAPIStatus(status, int64(id))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(apiStatus); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *SyncAPIServer) TriggerSyncSource(w http.ResponseWriter, r *http.Request, id int) {
	// Trigger sync for the source
	err := s.Service.TriggerSync(id)
	if err != nil {
		if err == sync.ErrSourceNotFound {
			s.writeError(w, http.StatusNotFound, "Source not found", "SOURCE_NOT_FOUND")
			return
		}
		if err == sync.ErrSyncAlreadyRunning {
			s.writeError(w, http.StatusConflict, "Sync already running for this source", "SYNC_ALREADY_RUNNING")
			return
		}
		s.writeError(w, http.StatusInternalServerError, "Failed to trigger sync", "INTERNAL_ERROR")
		return
	}

	response := SyncTriggerResponse{
		Message:   stringPtr("Sync triggered successfully"),
		SourceId:  &id,
		Triggered: boolPtr(true),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// Helper methods

func (s *SyncAPIServer) convertToAPISource(source sync.SourceConfig, id int64) SyncSource {
	apiSource := SyncSource{
		Id: intPtr(int(id)),
	}

	// Set type and config based on source type
	cfg := source.GetConfig()
	if cfg != nil {
		apiSource.Interval = stringPtr(cfg.GetInterval().String())

		switch cfg.GetSourceType() {
		case "git":
			gitConfig := cfg.(*sync.GitSourceConfig)
			apiSource.Type = (*SyncSourceType)(stringPtr("git"))
			gitAPIConfig := GitSourceConfig{
				Url:      stringPtr(gitConfig.URL),
				Branch:   stringPtr(gitConfig.Branch),
				BasePath: stringPtr(gitConfig.BasePath),
			}
			apiSource.Config = &SyncSource_Config{}
			if err := apiSource.Config.FromGitSourceConfig(gitAPIConfig); err != nil {
				// Log error but continue - this is a conversion issue
				// The source config is already validated, so this shouldn't fail
				return apiSource
			}

		case "filesystem":
			fsConfig := cfg.(*sync.FilesystemSourceConfig)
			apiSource.Type = (*SyncSourceType)(stringPtr("filesystem"))
			fsAPIConfig := FilesystemSourceConfig{
				Path: stringPtr(fsConfig.Path),
			}
			apiSource.Config = &SyncSource_Config{}
			if err := apiSource.Config.FromFilesystemSourceConfig(fsAPIConfig); err != nil {
				// Log error but continue - this is a conversion issue
				// The source config is already validated, so this shouldn't fail
				return apiSource
			}
		}
	}

	return apiSource
}

func (s *SyncAPIServer) convertToAPIStatus(status *sync.SourceStatus, id int64) SyncStatus {
	apiStatus := SyncStatus{
		SourceId: intPtr(int(id)),
	}

	if status != nil {
		// Convert status enum
		var statusEnum SyncStatusStatus
		switch status.Status {
		case sync.StatusIdle:
			statusEnum = Idle
		case sync.StatusRunning:
			statusEnum = Running
		case sync.StatusCompleted:
			statusEnum = Completed
		case sync.StatusFailed:
			statusEnum = Failed
		default:
			statusEnum = Idle
		}
		apiStatus.Status = &statusEnum

		// Set other fields
		apiStatus.LastSync = status.LastSync
		apiStatus.LastError = status.LastError
		apiStatus.ComponentsCount = &status.ComponentsCount
		if status.Duration > 0 {
			duration := status.Duration.String()
			apiStatus.Duration = &duration
		}
	} else {
		// Default status for unknown sources
		idle := Idle
		apiStatus.Status = &idle
	}

	return apiStatus
}

func (s *SyncAPIServer) writeError(w http.ResponseWriter, statusCode int, message, code string) {
	error := Error{
		Message: &message,
		Code:    &code,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(error); err != nil {
		http.Error(w, "failed to encode error response", http.StatusInternalServerError)
		return
	}
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}
