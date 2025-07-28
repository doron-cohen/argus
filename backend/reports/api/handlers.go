package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
		s.sendErrorResponse(w, "Invalid JSON format", "VALIDATION_ERROR", http.StatusBadRequest)
		return
	}

	// Validate using OpenAPI spec constraints
	if err := validateReportSubmission(submission); err != nil {
		s.sendErrorResponse(w, err.Error(), "VALIDATION_ERROR", http.StatusBadRequest)
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
	reportID, err := s.Repo.CreateCheckReportFromSubmission(ctx, input)
	if err != nil {
		if err == storage.ErrComponentNotFound {
			s.sendErrorResponse(w, "Component not found", "NOT_FOUND", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("failed to create report: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	response := client.ReportSubmissionResponse{
		Message:   utils.ToPointer("Report submitted successfully"),
		ReportId:  utils.ToPointer(reportID.String()),
		Timestamp: utils.ToPointer(time.Now()),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// sendErrorResponse sends a JSON error response
func (s *APIServer) sendErrorResponse(w http.ResponseWriter, message, code string, statusCode int) {
	errorResponse := client.Error{
		Error: utils.ToPointer(message),
		Code:  utils.ToPointer(code),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}

// validateReportSubmission validates a report submission against OpenAPI spec constraints
func validateReportSubmission(submission client.ReportSubmission) error {
	// Validate required fields (OpenAPI spec already enforces this via struct tags)
	if submission.Check.Slug == "" {
		return fmt.Errorf("check slug is required")
	}
	if submission.ComponentId == "" {
		return fmt.Errorf("component ID is required")
	}
	if submission.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}

	// Validate timestamp is not in the future
	if submission.Timestamp.After(time.Now()) {
		return fmt.Errorf("timestamp cannot be in the future")
	}

	// Validate check slug format using existing utility
	if !utils.IsValidSlug(submission.Check.Slug) {
		return fmt.Errorf("check slug can only contain alphanumeric characters, hyphens, and underscores")
	}

	// Validate component ID format (no leading/trailing whitespace)
	if strings.TrimSpace(submission.ComponentId) != submission.ComponentId {
		return fmt.Errorf("component ID cannot have leading or trailing whitespace")
	}

	// Validate status is one of the allowed values (OpenAPI enum already enforces this)
	switch submission.Status {
	case client.ReportSubmissionStatusPass,
		client.ReportSubmissionStatusFail,
		client.ReportSubmissionStatusDisabled,
		client.ReportSubmissionStatusSkipped,
		client.ReportSubmissionStatusUnknown,
		client.ReportSubmissionStatusError,
		client.ReportSubmissionStatusCompleted:
		// Valid status
	default:
		return fmt.Errorf("status must be one of: pass, fail, disabled, skipped, unknown, error, completed")
	}

	return nil
}
