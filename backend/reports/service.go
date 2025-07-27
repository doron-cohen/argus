package reports

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/doron-cohen/argus/backend/internal/storage"
	"github.com/doron-cohen/argus/backend/internal/utils"
	"github.com/google/uuid"
)

// Service orchestrates the reports process
type Service struct {
	repo *storage.Repository
}

// NewService creates a new reports service
func NewService(repo *storage.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// SubmitReportInput represents the input for submitting a report
type SubmitReportInput struct {
	ComponentID      string
	CheckSlug        string
	CheckName        *string
	CheckDescription *string
	Status           string
	Timestamp        time.Time
	Details          *map[string]interface{}
	Metadata         *map[string]interface{}
}

// SubmitReportResult represents the result of submitting a report
type SubmitReportResult struct {
	ReportID  string
	Timestamp time.Time
}

// SubmitReport submits a check report with proper business logic
func (s *Service) SubmitReport(ctx context.Context, input SubmitReportInput) (*SubmitReportResult, error) {
	slog.Info("Submitting report",
		"component_id", input.ComponentID,
		"check_slug", input.CheckSlug,
		"status", input.Status)

	// Validate input
	if err := s.validateSubmitReportInput(input); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Convert to storage input
	storageInput := storage.CreateCheckReportInput{
		ComponentID:      input.ComponentID,
		CheckSlug:        input.CheckSlug,
		CheckName:        input.CheckName,
		CheckDescription: input.CheckDescription,
		Status:           storage.CheckStatus(input.Status),
		Timestamp:        input.Timestamp,
	}

	// Convert optional JSONB fields
	if input.Details != nil {
		storageInput.Details = storage.JSONB(*input.Details)
	}
	if input.Metadata != nil {
		storageInput.Metadata = storage.JSONB(*input.Metadata)
	}

	// Store the report in the database
	if err := s.repo.CreateCheckReportFromSubmission(ctx, storageInput); err != nil {
		if err == storage.ErrComponentNotFound {
			return nil, fmt.Errorf("component not found: %s", input.ComponentID)
		}
		return nil, fmt.Errorf("failed to store report: %w", err)
	}

	// Generate result
	result := &SubmitReportResult{
		ReportID:  uuid.New().String(),
		Timestamp: time.Now(),
	}

	slog.Info("Report submitted successfully",
		"component_id", input.ComponentID,
		"check_slug", input.CheckSlug,
		"report_id", result.ReportID)

	return result, nil
}

// validateSubmitReportInput validates the input for submitting a report
func (s *Service) validateSubmitReportInput(input SubmitReportInput) error {
	// Validate check slug
	if err := s.validateCheckSlug(input.CheckSlug); err != nil {
		return err
	}

	// Validate check name (optional)
	if input.CheckName != nil {
		if err := s.validateCheckName(*input.CheckName); err != nil {
			return err
		}
	}

	// Validate check description (optional)
	if input.CheckDescription != nil {
		if err := s.validateCheckDescription(*input.CheckDescription); err != nil {
			return err
		}
	}

	// Validate component ID
	if err := s.validateComponentID(input.ComponentID); err != nil {
		return err
	}

	// Validate status
	if err := s.validateStatus(input.Status); err != nil {
		return err
	}

	// Validate timestamp
	if err := s.validateTimestamp(input.Timestamp); err != nil {
		return err
	}

	// Validate optional fields
	if err := s.validateOptionalFields(input); err != nil {
		return err
	}

	return nil
}

func (s *Service) validateCheckSlug(slug string) error {
	if strings.TrimSpace(slug) == "" {
		return fmt.Errorf("check slug is required and cannot be empty")
	}
	if len(slug) > 100 {
		return fmt.Errorf("check slug cannot exceed 100 characters")
	}
	if !utils.IsValidSlug(slug) {
		return fmt.Errorf("check slug must contain only alphanumeric characters, hyphens, and underscores")
	}
	return nil
}

func (s *Service) validateCheckName(name string) error {
	if len(name) > 255 {
		return fmt.Errorf("check name cannot exceed 255 characters")
	}
	return nil
}

func (s *Service) validateCheckDescription(description string) error {
	if len(description) > 1000 {
		return fmt.Errorf("check description cannot exceed 1000 characters")
	}
	return nil
}

func (s *Service) validateComponentID(componentID string) error {
	if strings.TrimSpace(componentID) == "" {
		return fmt.Errorf("component ID is required and cannot be empty")
	}
	if len(componentID) > 255 {
		return fmt.Errorf("component ID cannot exceed 255 characters")
	}
	return nil
}

func (s *Service) validateStatus(status string) error {
	validStatuses := []string{
		"pass",
		"fail",
		"disabled",
		"skipped",
		"unknown",
		"error",
		"completed",
	}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return nil
		}
	}
	return fmt.Errorf("status must be one of: %s", strings.Join(validStatuses, ", "))
}

func (s *Service) validateTimestamp(timestamp time.Time) error {
	if timestamp.IsZero() {
		return fmt.Errorf("timestamp is required and cannot be zero")
	}

	// Check if timestamp is not in the future (with 5 minute tolerance for clock skew)
	if timestamp.After(time.Now().Add(5 * time.Minute)) {
		return fmt.Errorf("timestamp cannot be in the future")
	}
	return nil
}

func (s *Service) validateOptionalFields(input SubmitReportInput) error {
	// Validate details (if provided)
	if input.Details != nil {
		if err := utils.ValidateJSONBField(*input.Details, "details"); err != nil {
			return err
		}
	}

	// Validate metadata (if provided)
	if input.Metadata != nil {
		if err := utils.ValidateJSONBField(*input.Metadata, "metadata"); err != nil {
			return err
		}
	}

	return nil
}
