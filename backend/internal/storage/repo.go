package storage

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ErrComponentNotFound is returned when a component is not found
var ErrComponentNotFound = errors.New("component not found")

// ErrCheckNotFound is returned when a check is not found
var ErrCheckNotFound = errors.New("check not found")

type Repository struct {
	DB *gorm.DB
}

func ConnectAndMigrate(ctx context.Context, dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Migrate all tables
	if err := db.WithContext(ctx).AutoMigrate(&Component{}, &Check{}, &CheckReport{}); err != nil {
		return nil, err
	}

	return &Repository{DB: db}, nil
}

func (r *Repository) Migrate(ctx context.Context) error {
	// Migrate all tables
	if err := r.DB.WithContext(ctx).AutoMigrate(&Component{}, &Check{}, &CheckReport{}); err != nil {
		return err
	}

	return nil
}

// Component methods
func (r *Repository) GetComponents(ctx context.Context) ([]Component, error) {
	var components []Component
	err := r.DB.WithContext(ctx).Find(&components).Error
	return components, err
}

// GetComponentByID returns a component by its unique identifier
func (r *Repository) GetComponentByID(ctx context.Context, componentID string) (*Component, error) {
	var component Component
	err := r.DB.WithContext(ctx).Where("component_id = ?", componentID).First(&component).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrComponentNotFound
		}
		return nil, err
	}
	return &component, nil
}

// GetComponentByName returns a component by name (for backward compatibility)
func (r *Repository) GetComponentByName(ctx context.Context, name string) (*Component, error) {
	var component Component
	err := r.DB.WithContext(ctx).Where("name = ?", name).First(&component).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrComponentNotFound
		}
		return nil, err
	}
	return &component, nil
}

// GetComponentsByTeam returns all components owned by a specific team
func (r *Repository) GetComponentsByTeam(ctx context.Context, team string) ([]Component, error) {
	var components []Component
	err := r.DB.WithContext(ctx).Where("team = ?", team).Find(&components).Error
	if err != nil {
		return nil, err
	}
	return components, nil
}

// CreateComponent creates a new component
func (r *Repository) CreateComponent(ctx context.Context, component Component) error {
	return r.DB.WithContext(ctx).Create(&component).Error
}

// Check methods - only what's needed for handlers
func (r *Repository) GetCheckBySlug(ctx context.Context, slug string) (*Check, error) {
	var check Check
	err := r.DB.WithContext(ctx).Where("slug = ?", slug).First(&check).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCheckNotFound
		}
		return nil, err
	}
	return &check, nil
}

func (r *Repository) CreateCheck(ctx context.Context, check Check) error {
	return r.DB.WithContext(ctx).Create(&check).Error
}

// GetOrCreateCheckBySlug auto-creates a check if it doesn't exist, returns CheckID
func (r *Repository) GetOrCreateCheckBySlug(ctx context.Context, slug string, name *string, description *string) (uuid.UUID, error) {
	// First try to get existing check
	check, err := r.GetCheckBySlug(ctx, slug)
	if err == nil {
		// Check exists, return its ID
		return check.ID, nil
	}
	if !errors.Is(err, ErrCheckNotFound) {
		// Some other error occurred
		return uuid.Nil, err
	}

	// Check doesn't exist, create it with provided values or defaults
	checkName := slug // Default name is slug
	if name != nil && *name != "" {
		checkName = *name
	}

	checkDescription := "Auto-created check for slug: " + slug // Default description
	if description != nil && *description != "" {
		checkDescription = *description
	}

	newCheck := Check{
		Slug:        slug,
		Name:        checkName,
		Description: checkDescription,
	}

	err = r.CreateCheck(ctx, newCheck)
	if err != nil {
		return uuid.Nil, err
	}

	// Get the created check to return its ID
	createdCheck, err := r.GetCheckBySlug(ctx, slug)
	if err != nil {
		return uuid.Nil, err
	}

	return createdCheck.ID, nil
}

// CreateCheckReportInput represents the input data for creating a check report
type CreateCheckReportInput struct {
	ComponentID      string
	CheckSlug        string
	CheckName        *string
	CheckDescription *string
	Status           CheckStatus
	Timestamp        time.Time
	Details          JSONB
	Metadata         JSONB
}

// CreateCheckReportFromSubmission creates a check report from API submission data
func (r *Repository) CreateCheckReportFromSubmission(ctx context.Context, input CreateCheckReportInput) error {
	// Use transaction to ensure atomicity
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Verify component exists and get its UUID using the transaction
		component, err := r.getComponentInTransaction(ctx, tx, input.ComponentID)
		if err != nil {
			return err
		}

		// Get or create check by slug with provided name and description using the transaction
		checkID, err := r.getOrCreateCheckInTransaction(ctx, tx, input)
		if err != nil {
			return err
		}

		// Create the report
		report := CheckReport{
			CheckID:     checkID,
			ComponentID: component.ID,
			Status:      input.Status,
			Timestamp:   input.Timestamp,
			Details:     input.Details,
			Metadata:    input.Metadata,
		}

		return tx.Create(&report).Error
	})
}

// getComponentInTransaction gets a component within a transaction
func (r *Repository) getComponentInTransaction(ctx context.Context, tx *gorm.DB, componentID string) (*Component, error) {
	var component Component
	err := tx.WithContext(ctx).Where("component_id = ?", componentID).First(&component).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrComponentNotFound
		}
		return nil, err
	}
	return &component, nil
}

// getOrCreateCheckInTransaction gets or creates a check within a transaction
func (r *Repository) getOrCreateCheckInTransaction(ctx context.Context, tx *gorm.DB, input CreateCheckReportInput) (uuid.UUID, error) {
	var check Check
	err := tx.WithContext(ctx).Where("slug = ?", input.CheckSlug).First(&check).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return uuid.Nil, err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.createCheckInTransaction(ctx, tx, input)
	}

	return check.ID, nil
}

// createCheckInTransaction creates a new check within a transaction
func (r *Repository) createCheckInTransaction(ctx context.Context, tx *gorm.DB, input CreateCheckReportInput) (uuid.UUID, error) {
	checkName := input.CheckSlug // Default name is slug
	if input.CheckName != nil && *input.CheckName != "" {
		checkName = *input.CheckName
	}

	checkDescription := "Auto-created check for slug: " + input.CheckSlug // Default description
	if input.CheckDescription != nil && *input.CheckDescription != "" {
		checkDescription = *input.CheckDescription
	}

	newCheck := Check{
		Slug:        input.CheckSlug,
		Name:        checkName,
		Description: checkDescription,
	}

	if err := tx.WithContext(ctx).Create(&newCheck).Error; err != nil {
		return uuid.Nil, err
	}

	return newCheck.ID, nil
}

// GetCheckReportsForComponent retrieves all check reports for a specific component
func (r *Repository) GetCheckReportsForComponent(ctx context.Context, componentID string) ([]CheckReport, error) {
	// First verify the component exists
	component, err := r.GetComponentByID(ctx, componentID)
	if err != nil {
		return nil, err
	}

	var reports []CheckReport
	err = r.DB.WithContext(ctx).
		Preload("Check").
		Where("component_id = ?", component.ID).
		Order("timestamp DESC").
		Find(&reports).Error

	return reports, err
}

// GetLatestCheckReportsForComponent retrieves the latest report for each check type for a specific component
func (r *Repository) GetLatestCheckReportsForComponent(ctx context.Context, componentID string) ([]CheckReport, error) {
	// First verify the component exists
	component, err := r.GetComponentByID(ctx, componentID)
	if err != nil {
		return nil, err
	}

	var reports []CheckReport
	err = r.DB.WithContext(ctx).
		Preload("Check").
		Where("component_id = ?", component.ID).
		Where("id IN (?)",
			r.DB.Table("check_reports").
				Select("DISTINCT ON (check_id) id").
				Where("component_id = ?", component.ID).
				Order("check_id, timestamp DESC")).
		Order("timestamp DESC").
		Find(&reports).Error

	return reports, err
}

// GetCheckReportsByStatus retrieves check reports filtered by status
func (r *Repository) GetCheckReportsByStatus(ctx context.Context, status CheckStatus) ([]CheckReport, error) {
	var reports []CheckReport
	err := r.DB.WithContext(ctx).
		Preload("Check").
		Preload("Component").
		Where("status = ?", status).
		Order("timestamp DESC").
		Find(&reports).Error

	return reports, err
}

// GetCheckReportsByCheckSlug retrieves all reports for a specific check type
func (r *Repository) GetCheckReportsByCheckSlug(ctx context.Context, checkSlug string) ([]CheckReport, error) {
	var reports []CheckReport
	err := r.DB.WithContext(ctx).
		Preload("Check").
		Preload("Component").
		Joins("JOIN checks ON checks.id = check_reports.check_id").
		Where("checks.slug = ?", checkSlug).
		Order("check_reports.timestamp DESC").
		Find(&reports).Error

	return reports, err
}

// HealthCheck implements the health.Checker interface
func (r *Repository) HealthCheck(ctx context.Context) error {
	return r.DB.WithContext(ctx).Raw("SELECT 1").Error
}

func (r *Repository) Name() string {
	return "database"
}
