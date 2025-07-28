package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/doron-cohen/argus/backend/internal/storage"
	"github.com/doron-cohen/argus/backend/internal/utils"
	"github.com/doron-cohen/argus/backend/reports/api/client"
)

// APIServer implements the ReportsAPI interface
type APIServer struct {
	Repo *storage.Repository
}

// NewAPIServer creates a new API server
func NewAPIServer(repo *storage.Repository) ServerInterface {
	return &APIServer{Repo: repo}
}

// convertToStorageStatus converts API status to storage status
func convertToStorageStatus(status client.ReportSubmissionStatus) storage.CheckStatus {
	switch status {
	case client.ReportSubmissionStatusPass:
		return storage.CheckStatusPass
	case client.ReportSubmissionStatusFail:
		return storage.CheckStatusFail
	case client.ReportSubmissionStatusDisabled:
		return storage.CheckStatusDisabled
	case client.ReportSubmissionStatusSkipped:
		return storage.CheckStatusSkipped
	case client.ReportSubmissionStatusUnknown:
		return storage.CheckStatusUnknown
	case client.ReportSubmissionStatusError:
		return storage.CheckStatusError
	case client.ReportSubmissionStatusCompleted:
		return storage.CheckStatusCompleted
	default:
		return storage.CheckStatusUnknown
	}
}

// SubmitReport handles report submission
func (s *APIServer) SubmitReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var submission client.ReportSubmission
	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if submission.Check.Slug == "" {
		http.Error(w, "check slug is required", http.StatusBadRequest)
		return
	}
	if submission.ComponentId == "" {
		http.Error(w, "component ID is required", http.StatusBadRequest)
		return
	}

	// Convert API submission to storage input
	var details storage.JSONB
	if submission.Details != nil {
		details = storage.JSONB(*submission.Details)
	}

	var metadata storage.JSONB
	if submission.Metadata != nil {
		metadata = storage.JSONB(*submission.Metadata)
	}

	input := storage.CreateCheckReportInput{
		ComponentID:      submission.ComponentId,
		CheckSlug:        submission.Check.Slug,
		CheckName:        submission.Check.Name,
		CheckDescription: submission.Check.Description,
		Status:           convertToStorageStatus(submission.Status),
		Timestamp:        submission.Timestamp,
		Details:          details,
		Metadata:         metadata,
	}

	// Create the report
	if err := s.Repo.CreateCheckReportFromSubmission(ctx, input); err != nil {
		http.Error(w, fmt.Sprintf("failed to create report: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	response := client.ReportSubmissionResponse{
		Message:   utils.ToPointer("Report submitted successfully"),
		ReportId:  utils.ToPointer("report-id"), // TODO: return actual report ID
		Timestamp: utils.ToPointer(time.Now()),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
