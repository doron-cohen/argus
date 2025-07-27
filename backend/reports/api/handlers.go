package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/doron-cohen/argus/backend/internal/storage"
	"github.com/doron-cohen/argus/backend/reports"
)

type ReportsServer struct {
	Service *reports.Service
}

func NewReportsServer(repo *storage.Repository) ServerInterface {
	return &ReportsServer{Service: reports.NewService(repo)}
}

func (s *ReportsServer) SubmitReport(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var submission ReportSubmission
	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
		s.writeValidationError(w, "Invalid JSON format", "INVALID_JSON", map[string]interface{}{
			"reason": "Request body is not valid JSON",
		})
		return
	}

	// Convert API submission to service input
	input := reports.SubmitReportInput{
		ComponentID:      submission.ComponentId,
		CheckSlug:        submission.Check.Slug,
		CheckName:        submission.Check.Name,
		CheckDescription: submission.Check.Description,
		Status:           string(submission.Status),
		Timestamp:        submission.Timestamp,
		Details:          submission.Details,
		Metadata:         submission.Metadata,
	}

	// Submit report via service layer
	result, err := s.Service.SubmitReport(r.Context(), input)
	if err != nil {
		// Handle specific error types
		if strings.Contains(err.Error(), "component not found") {
			s.writeValidationError(w, "Component not found", "COMPONENT_NOT_FOUND", map[string]interface{}{
				"component_id": submission.ComponentId,
			})
			return
		}
		if strings.Contains(err.Error(), "validation error") {
			s.writeValidationError(w, err.Error(), "VALIDATION_ERROR", map[string]interface{}{
				"reason": err.Error(),
			})
			return
		}
		http.Error(w, "Failed to store report", http.StatusInternalServerError)
		return
	}

	// Return success response
	response := ReportSubmissionResponse{
		Message:   &[]string{"Report submitted successfully"}[0],
		ReportId:  &result.ReportID,
		Timestamp: &result.Timestamp,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *ReportsServer) writeValidationError(w http.ResponseWriter, message, code string, details map[string]interface{}) {
	errorResponse := Error{
		Error:   &message,
		Code:    &code,
		Details: &details,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		return
	}
}
