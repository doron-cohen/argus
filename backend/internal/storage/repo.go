package storage

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// StringArray is a custom type to handle PostgreSQL JSONB arrays
// This allows for powerful querying capabilities using PostgreSQL's JSONB operators:
// - ? : Check if a string exists in the array
// - ?| : Check if any of the strings exist in the array
// - ?& : Check if all of the strings exist in the array
// Example queries:
//
//	WHERE maintainers ? 'alice@company.com'                    -- Find components maintained by alice
//	WHERE maintainers ?| '["alice@company.com", "bob@company.com"]' -- Find components maintained by either alice or bob
//	WHERE maintainers ?& '["alice@company.com", "bob@company.com"]' -- Find components maintained by both alice and bob
type StringArray []string

// Value implements the driver.Valuer interface
func (sa StringArray) Value() (driver.Value, error) {
	if sa == nil {
		return nil, nil
	}
	return json.Marshal(sa)
}

// Scan implements the sql.Scanner interface
func (sa *StringArray) Scan(value interface{}) error {
	if value == nil {
		*sa = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot scan non-json value into StringArray")
	}

	return json.Unmarshal(bytes, sa)
}

// GormDataType implements GormDataTypeInterface
func (sa StringArray) GormDataType() string {
	return "jsonb"
}

// GormDBDataType implements GormDBDataTypeInterface
func (sa StringArray) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "jsonb"
}

// Contains checks if the array contains a specific string
func (sa StringArray) Contains(value string) bool {
	for _, item := range sa {
		if item == value {
			return true
		}
	}
	return false
}

// ContainsAny checks if the array contains any of the provided strings
func (sa StringArray) ContainsAny(values []string) bool {
	for _, value := range values {
		if sa.Contains(value) {
			return true
		}
	}
	return false
}

// ErrComponentNotFound is returned when a component is not found
var ErrComponentNotFound = errors.New("component not found")

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
