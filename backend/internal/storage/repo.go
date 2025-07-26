package storage

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ErrComponentNotFound is returned when a component is not found
var ErrComponentNotFound = errors.New("component not found")

// ErrCheckNotFound is returned when a check is not found
var ErrCheckNotFound = errors.New("check not found")

// Component represents a component stored in the database.
type Component struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	ComponentID string    `gorm:"not null;uniqueIndex"` // Unique identifier from manifest
	Name        string    `gorm:"not null"`
	Description string
	Maintainers StringArray `gorm:"type:jsonb"`
	Team        string
}

func (c *Component) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID, err = uuid.NewV7()
	}
	return
}

type Repository struct {
	DB *gorm.DB
}

func ConnectAndMigrate(ctx context.Context, dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.WithContext(ctx).AutoMigrate(&Component{}, &Check{}, &CheckReport{}); err != nil {
		return nil, err
	}
	return &Repository{DB: db}, nil
}

func (r *Repository) Migrate(ctx context.Context) error {
	return r.DB.WithContext(ctx).AutoMigrate(&Component{}, &Check{}, &CheckReport{})
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

// CheckReport methods - only what's needed for handlers
func (r *Repository) CreateCheckReport(ctx context.Context, report CheckReport) error {
	return r.DB.WithContext(ctx).Create(&report).Error
}

// HealthCheck implements the health.Checker interface
func (r *Repository) HealthCheck(ctx context.Context) error {
	return r.DB.WithContext(ctx).Raw("SELECT 1").Error
}

func (r *Repository) Name() string {
	return "database"
}
