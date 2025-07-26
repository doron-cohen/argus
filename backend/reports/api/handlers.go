package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ReportsServer struct{}

func NewReportsServer() ServerInterface {
	return &ReportsServer{}
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

	// For now, just return success (no storage yet)
	response := ReportSubmissionResponse{
		Message:   "Report submitted successfully",
		ReportId:  s.generateReportID(),
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *ReportsServer) validateReportSubmission(submission *ReportSubmission) error {
	// Validate check_slug
	if strings.TrimSpace(submission.CheckSlug) == "" {
		return fmt.Errorf("check_slug is required and cannot be empty")
	}
	if len(submission.CheckSlug) > 100 {
		return fmt.Errorf("check_slug cannot exceed 100 characters")
	}
	if !s.isValidSlug(submission.CheckSlug) {
		return fmt.Errorf("check_slug must contain only alphanumeric characters, hyphens, and underscores")
	}

	// Validate component_id
	if strings.TrimSpace(submission.ComponentId) == "" {
		return fmt.Errorf("component_id is required and cannot be empty")
	}
	if len(submission.ComponentId) > 255 {
		return fmt.Errorf("component_id cannot exceed 255 characters")
	}

	// Validate status
	if !s.isValidStatus(submission.Status) {
		return fmt.Errorf("status must be one of: pass, fail, disabled, skipped, unknown, error, completed")
	}

	// Validate timestamp
	if submission.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required and cannot be zero")
	}

	// Check if timestamp is not in the future (with 5 minute tolerance for clock skew)
	if submission.Timestamp.After(time.Now().Add(5 * time.Minute)) {
		return fmt.Errorf("timestamp cannot be in the future")
	}

	// Validate details (if provided)
	if submission.Details != nil {
		if err := s.validateJSONBField(*submission.Details, "details"); err != nil {
			return err
		}
	}

	// Validate metadata (if provided)
	if submission.Metadata != nil {
		if err := s.validateJSONBField(*submission.Metadata, "metadata"); err != nil {
			return err
		}
	}

	return nil
}

func (s *ReportsServer) isValidSlug(slug string) bool {
	// Allow alphanumeric characters, hyphens, and underscores
	for _, char := range slug {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return false
		}
	}
	return true
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

func (s *ReportsServer) validateJSONBField(data map[string]interface{}, fieldName string) error {
	// Check for reasonable size limit (1MB)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("%s must be a valid JSON object", fieldName)
	}

	if len(jsonData) > 1024*1024 {
		return fmt.Errorf("%s cannot exceed 1MB", fieldName)
	}

	// Check for reasonable depth (max 10 levels)
	if s.getMaxDepth(data) > 10 {
		return fmt.Errorf("%s cannot exceed 10 levels of nesting", fieldName)
	}

	return nil
}

func (s *ReportsServer) getMaxDepth(data interface{}) int {
	switch v := data.(type) {
	case map[string]interface{}:
		maxDepth := 1
		for _, value := range v {
			depth := s.getMaxDepth(value)
			if depth+1 > maxDepth {
				maxDepth = depth + 1
			}
		}
		return maxDepth
	case []interface{}:
		maxDepth := 1
		for _, value := range v {
			depth := s.getMaxDepth(value)
			if depth+1 > maxDepth {
				maxDepth = depth + 1
			}
		}
		return maxDepth
	default:
		return 0
	}
}

func (s *ReportsServer) generateReportID() *string {
	id := uuid.New().String()
	return &id
}

func (s *ReportsServer) writeValidationError(w http.ResponseWriter, message, code string, details map[string]interface{}) {
	errorResponse := Error{
		Error:   message,
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
