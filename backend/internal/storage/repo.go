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

// Component represents a component stored in the database.
type Component struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	ComponentID string    `gorm:"not null;uniqueIndex"` // Unique identifier from manifest
	Name        string    `gorm:"not null"`
	Description string
	Maintainers []string `gorm:"type:text[]"`
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
	if err := db.WithContext(ctx).AutoMigrate(&Component{}); err != nil {
		return nil, err
	}
	return &Repository{DB: db}, nil
}

func (r *Repository) Migrate(ctx context.Context) error {
	return r.DB.WithContext(ctx).AutoMigrate(&Component{})
}

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

// CreateComponent creates a new component
func (r *Repository) CreateComponent(ctx context.Context, component Component) error {
	return r.DB.WithContext(ctx).Create(&component).Error
}
