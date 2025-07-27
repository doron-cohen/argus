package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/doron-cohen/argus/backend/internal/storage"
	"github.com/doron-cohen/argus/backend/internal/utils"
	"github.com/google/uuid"
)

type ReportsServer struct {
	Repo *storage.Repository
}

func NewReportsServer(repo *storage.Repository) ServerInterface {
	return &ReportsServer{Repo: repo}
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

	// Validate required fields
	if err := s.validateReportSubmission(&submission); err != nil {
		s.writeValidationError(w, err.Error(), "VALIDATION_ERROR", map[string]interface{}{
			"reason": err.Error(),
		})
		return
	}

	// Convert API submission to storage input
	input := storage.CreateCheckReportInput{
		ComponentID:      submission.ComponentId,
		CheckSlug:        submission.Check.Slug,
		CheckName:        submission.Check.Name,
		CheckDescription: submission.Check.Description,
		Status:           storage.CheckStatus(submission.Status),
		Timestamp:        submission.Timestamp,
	}

	// Convert optional JSONB fields
	if submission.Details != nil {
		input.Details = storage.JSONB(*submission.Details)
	}
	if submission.Metadata != nil {
		input.Metadata = storage.JSONB(*submission.Metadata)
	}

	// Store the report in the database
	if err := s.Repo.CreateCheckReportFromSubmission(r.Context(), input); err != nil {
		if err == storage.ErrComponentNotFound {
			s.writeValidationError(w, "Component not found", "COMPONENT_NOT_FOUND", map[string]interface{}{
				"component_id": submission.ComponentId,
			})
			return
		}
		http.Error(w, "Failed to store report", http.StatusInternalServerError)
		return
	}

	// Return success response
	response := ReportSubmissionResponse{
		Message:   &[]string{"Report submitted successfully"}[0],
		ReportId:  s.generateReportID(),
		Timestamp: &[]time.Time{time.Now()}[0],
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *ReportsServer) validateReportSubmission(submission *ReportSubmission) error {
	if err := s.validateCheck(submission.Check); err != nil {
		return err
	}

	if err := s.validateComponentId(submission.ComponentId); err != nil {
		return err
	}

	if err := s.validateStatus(submission.Status); err != nil {
		return err
	}

	if err := s.validateTimestamp(submission.Timestamp); err != nil {
		return err
	}

	if err := s.validateOptionalFields(submission); err != nil {
		return err
	}

	return nil
}

func (s *ReportsServer) validateCheck(check Check) error {
	// Validate check.slug
	if strings.TrimSpace(check.Slug) == "" {
		return fmt.Errorf("check.slug is required and cannot be empty")
	}
	if len(check.Slug) > 100 {
		return fmt.Errorf("check.slug cannot exceed 100 characters")
	}
	if !utils.IsValidSlug(check.Slug) {
		return fmt.Errorf("check.slug must contain only alphanumeric characters, hyphens, and underscores")
	}

	// Validate check.name (optional)
	if check.Name != nil && len(*check.Name) > 255 {
		return fmt.Errorf("check.name cannot exceed 255 characters")
	}

	// Validate check.description (optional)
	if check.Description != nil && len(*check.Description) > 1000 {
		return fmt.Errorf("check.description cannot exceed 1000 characters")
	}

	return nil
}

func (s *ReportsServer) validateComponentId(componentId string) error {
	if strings.TrimSpace(componentId) == "" {
		return fmt.Errorf("component_id is required and cannot be empty")
	}
	if len(componentId) > 255 {
		return fmt.Errorf("component_id cannot exceed 255 characters")
	}
	return nil
}

func (s *ReportsServer) validateStatus(status ReportSubmissionStatus) error {
	if !s.isValidStatus(status) {
		return fmt.Errorf("status must be one of: pass, fail, disabled, skipped, unknown, error, completed")
	}
	return nil
}

func (s *ReportsServer) validateTimestamp(timestamp time.Time) error {
	if timestamp.IsZero() {
		return fmt.Errorf("timestamp is required and cannot be zero")
	}

	// Check if timestamp is not in the future (with 5 minute tolerance for clock skew)
	if timestamp.After(time.Now().Add(5 * time.Minute)) {
		return fmt.Errorf("timestamp cannot be in the future")
	}
	return nil
}

func (s *ReportsServer) validateOptionalFields(submission *ReportSubmission) error {
	// Validate details (if provided)
	if submission.Details != nil {
		if err := utils.ValidateJSONBField(*submission.Details, "details"); err != nil {
			return err
		}
	}

	// Validate metadata (if provided)
	if submission.Metadata != nil {
		if err := utils.ValidateJSONBField(*submission.Metadata, "metadata"); err != nil {
			return err
		}
	}

	return nil
}

func (s *ReportsServer) isValidStatus(status ReportSubmissionStatus) bool {
	validStatuses := []ReportSubmissionStatus{
		ReportSubmissionStatusPass,
		ReportSubmissionStatusFail,
		ReportSubmissionStatusDisabled,
		ReportSubmissionStatusSkipped,
		ReportSubmissionStatusUnknown,
		ReportSubmissionStatusError,
		ReportSubmissionStatusCompleted,
	}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

func (s *ReportsServer) generateReportID() *string {
	id := uuid.New().String()
	return &id
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
