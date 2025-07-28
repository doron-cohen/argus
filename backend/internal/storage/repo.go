package storage

import (
	"context"
	"errors"
	"fmt"
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

// GORM Scopes for reusable query logic

// WithComponentID scope filters by component ID
func WithComponentID(componentID uuid.UUID) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("component_id = ?", componentID)
	}
}

// WithStatus scope filters by check status
func WithStatus(status CheckStatus) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

// WithCheckSlug scope filters by check slug
func WithCheckSlug(checkSlug string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Joins("JOIN checks ON check_reports.check_id = checks.id").
			Where("checks.slug = ?", checkSlug)
	}
}

// WithSince scope filters by timestamp (since)
func WithSince(since time.Time) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("timestamp >= ?", since)
	}
}

// WithPagination scope applies pagination
func WithPagination(limit, offset int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset).Limit(limit)
	}
}

// WithOrderByTimestamp scope orders by timestamp descending
func WithOrderByTimestamp() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order("timestamp DESC")
	}
}

// WithLatestPerCheck scope applies latest per check logic
// For PostgreSQL, uses DISTINCT ON; for SQLite, uses a subquery approach
func WithLatestPerCheck() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// Check if we're using PostgreSQL by looking at the driver
		dialectorName := db.Name()
		if dialectorName == "postgres" {
			// PostgreSQL-specific DISTINCT ON approach
			return db.Distinct("check_id, id, component_id, status, timestamp, details, metadata, created_at, updated_at").
				Order("check_id, timestamp DESC")
		}

		// For SQLite and other databases, we'll handle this in the main query
		// by using a subquery to get the latest timestamp for each check_id
		subQuery := db.Session(&gorm.Session{}).
			Model(&CheckReport{}).
			Select("check_id, MAX(timestamp) as max_timestamp").
			Group("check_id")

		return db.Joins("JOIN (?) as latest ON check_reports.check_id = latest.check_id AND check_reports.timestamp = latest.max_timestamp", subQuery)
	}
}

// WithPreloads scope adds necessary preloads
func WithPreloads() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Check")
	}
}

// applyFilters applies all filters to a query
func (r *Repository) applyFilters(query *gorm.DB, status *CheckStatus, checkSlug *string, since *time.Time) *gorm.DB {
	if status != nil {
		query = query.Scopes(WithStatus(*status))
	}
	if checkSlug != nil && *checkSlug != "" {
		query = query.Scopes(WithCheckSlug(*checkSlug))
	}
	if since != nil {
		query = query.Scopes(WithSince(*since))
	}
	return query
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
func (r *Repository) CreateCheckReportFromSubmission(ctx context.Context, input CreateCheckReportInput) (uuid.UUID, error) {
	var reportID uuid.UUID

	// Use transaction to ensure atomicity
	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

		if err := tx.Create(&report).Error; err != nil {
			return err
		}

		reportID = report.ID
		return nil
	})

	return reportID, err
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

// HealthCheck implements the health.Checker interface
func (r *Repository) HealthCheck(ctx context.Context) error {
	return r.DB.WithContext(ctx).Raw("SELECT 1").Error
}

func (r *Repository) Name() string {
	return "database"
}

// GetCheckReportsForComponentWithPagination retrieves check reports for a component with database-level filtering, pagination, and latest per check
func (r *Repository) GetCheckReportsForComponentWithPagination(ctx context.Context, componentID string, status *CheckStatus, checkSlug *string, since *time.Time, limit int, offset int, latestPerCheck bool) ([]CheckReport, int64, error) {
	// First verify the component exists
	component, err := r.GetComponentByID(ctx, componentID)
	if err != nil {
		return nil, 0, err
	}

	// Get total count for pagination
	var total int64
	if latestPerCheck {
		// For latest per check, we need to count the number of unique checks
		countQuery := r.DB.WithContext(ctx).
			Model(&CheckReport{}).
			Select("COUNT(DISTINCT check_id)").
			Where("component_id = ?", component.ID)

		// Apply filters to count query
		countQuery = r.applyFilters(countQuery, status, checkSlug, since)

		err = countQuery.Scan(&total).Error
		if err != nil {
			return nil, 0, fmt.Errorf("count query failed: %w", err)
		}
	} else {
		// Build base query for counting
		countQuery := r.DB.WithContext(ctx).Model(&CheckReport{}).
			Scopes(WithComponentID(component.ID))

		// Apply filters to count query
		countQuery = r.applyFilters(countQuery, status, checkSlug, since)

		err = countQuery.Count(&total).Error
		if err != nil {
			return nil, 0, fmt.Errorf("count query failed: %w", err)
		}
	}

	// Build query for fetching data
	query := r.DB.WithContext(ctx).
		Scopes(WithComponentID(component.ID), WithPreloads())

	// Apply filters
	query = r.applyFilters(query, status, checkSlug, since)

	// Handle latest per check logic
	if latestPerCheck {
		return r.getLatestPerCheckReports(ctx, query, *component, status, checkSlug, since, limit, offset)
	}

	// Apply pagination and ordering
	query = query.Scopes(WithPagination(limit, offset), WithOrderByTimestamp())

	var reports []CheckReport
	err = query.Find(&reports).Error
	if err != nil {
		return nil, 0, fmt.Errorf("find query failed: %w", err)
	}

	return reports, total, err
}

// applyLatestPerCheckFilters applies filters consistently for latest per check logic
func (r *Repository) applyLatestPerCheckFilters(query *gorm.DB, status *CheckStatus, checkSlug *string, since *time.Time) *gorm.DB {
	filteredQuery := query

	if status != nil {
		filteredQuery = filteredQuery.Where("check_reports.status = ?", *status)
	}
	if checkSlug != nil && *checkSlug != "" {
		// Since the main query already includes checks join through WithPreloads,
		// we can directly filter on checks.slug
		filteredQuery = filteredQuery.Where("checks.slug = ?", *checkSlug)
	}
	if since != nil {
		filteredQuery = filteredQuery.Where("check_reports.timestamp >= ?", *since)
	}

	return filteredQuery
}

// getLatestPerCheckReportsPostgreSQL handles latest per check logic for PostgreSQL
func (r *Repository) getLatestPerCheckReportsPostgreSQL(ctx context.Context, query *gorm.DB, component Component, status *CheckStatus, checkSlug *string, since *time.Time, limit int, offset int) ([]CheckReport, int64, error) {
	// We need to use a subquery to get the latest report for each check
	subQuery := r.DB.WithContext(ctx).
		Model(&CheckReport{}).
		Select("DISTINCT ON (check_id) check_reports.id").
		Where("check_reports.component_id = ?", component.ID).
		Order("check_id, timestamp DESC")

	// Apply the same filters to the subquery using the shared helper
	subQuery = r.applyLatestPerCheckFilters(subQuery, status, checkSlug, since)

	// Use the subquery to filter the main query
	query = query.Where("check_reports.id IN (?)", subQuery)

	// Apply pagination and ordering
	query = query.Scopes(WithPagination(limit, offset), WithOrderByTimestamp())

	var reports []CheckReport
	err := query.Find(&reports).Error
	if err != nil {
		return nil, 0, fmt.Errorf("find query failed: %w", err)
	}

	// Count total for pagination
	countQuery := r.DB.WithContext(ctx).
		Model(&CheckReport{}).
		Select("COUNT(DISTINCT check_id)").
		Where("component_id = ?", component.ID)

	// Apply filters to count query
	countQuery = r.applyFilters(countQuery, status, checkSlug, since)

	var total int64
	err = countQuery.Scan(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("count query failed: %w", err)
	}

	return reports, total, nil
}

// getLatestPerCheckReportsSQLite handles latest per check logic for SQLite and other databases
func (r *Repository) getLatestPerCheckReportsSQLite(ctx context.Context, query *gorm.DB, component Component, status *CheckStatus, checkSlug *string, since *time.Time, limit int, offset int) ([]CheckReport, int64, error) {
	// Apply the same filters as PostgreSQL for consistency
	filteredQuery := r.applyLatestPerCheckFilters(query, status, checkSlug, since)

	// Fetch all reports and filter in Go
	// This is simpler and more reliable than complex subqueries
	var allReports []CheckReport
	err := filteredQuery.Find(&allReports).Error
	if err != nil {
		return nil, 0, fmt.Errorf("find query failed: %w", err)
	}

	// Group by check and keep only the latest
	latestByCheck := make(map[string]CheckReport)
	for _, report := range allReports {
		checkSlug := report.Check.Slug
		if existing, exists := latestByCheck[checkSlug]; !exists || report.Timestamp.After(existing.Timestamp) {
			latestByCheck[checkSlug] = report
		}
	}

	// Convert back to slice
	var latestReports []CheckReport
	for _, report := range latestByCheck {
		latestReports = append(latestReports, report)
	}

	// Apply pagination to the filtered results
	total := int64(len(latestReports))
	start := offset
	end := offset + limit
	if start >= len(latestReports) {
		return []CheckReport{}, total, nil
	}
	if end > len(latestReports) {
		end = len(latestReports)
	}

	return latestReports[start:end], total, nil
}

// getLatestPerCheckReports handles the latest per check logic for different database types
func (r *Repository) getLatestPerCheckReports(ctx context.Context, query *gorm.DB, component Component, status *CheckStatus, checkSlug *string, since *time.Time, limit int, offset int) ([]CheckReport, int64, error) {
	// Check if we're using PostgreSQL
	dialectorName := r.DB.Name()
	if dialectorName == "postgres" {
		return r.getLatestPerCheckReportsPostgreSQL(ctx, query, component, status, checkSlug, since, limit, offset)
	} else {
		return r.getLatestPerCheckReportsSQLite(ctx, query, component, status, checkSlug, since, limit, offset)
	}
}
