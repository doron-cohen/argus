package storage

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CheckStatus represents the status of a check execution
type CheckStatus string

const (
	CheckStatusPass      CheckStatus = "pass"
	CheckStatusFail      CheckStatus = "fail"
	CheckStatusDisabled  CheckStatus = "disabled"
	CheckStatusSkipped   CheckStatus = "skipped"
	CheckStatusUnknown   CheckStatus = "unknown"
	CheckStatusError     CheckStatus = "error"
	CheckStatusCompleted CheckStatus = "completed"
)

// Component represents a component stored in the database.
type Component struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	ComponentID string    `gorm:"not null;uniqueIndex"` // Unique identifier from manifest
	Name        string    `gorm:"not null"`
	Description string
	Maintainers StringArray `gorm:"type:jsonb"`
	Team        string

	// Relationships
	CheckReports []CheckReport
}

func (c *Component) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID, err = uuid.NewV7()
	}
	return
}

// Check represents a quality check that can be performed on components
type Check struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Slug        string    `gorm:"not null;uniqueIndex;size:100"`
	Name        string    `gorm:"not null;size:255"`
	Description string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func (c *Check) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID, err = uuid.NewV7()
	}
	return
}

// CheckReport represents a report of a check execution on a component
type CheckReport struct {
	ID          uuid.UUID   `gorm:"type:uuid;primaryKey"`
	CheckID     uuid.UUID   `gorm:"type:uuid;not null;index:idx_check_timestamp"`
	ComponentID uuid.UUID   `gorm:"type:uuid;not null;index:idx_component_check"`
	Status      CheckStatus `gorm:"type:varchar(20);not null;index:idx_check_status"`
	Timestamp   time.Time   `gorm:"not null;index:idx_check_timestamp"`
	Details     JSONB       `gorm:"type:jsonb"`
	Metadata    JSONB       `gorm:"type:jsonb"`
	CreatedAt   time.Time   `gorm:"autoCreateTime"`
	UpdatedAt   time.Time   `gorm:"autoUpdateTime"`

	// Relationships
	Check     Check
	Component Component
}

func (cr *CheckReport) BeforeCreate(tx *gorm.DB) (err error) {
	if cr.ID == uuid.Nil {
		cr.ID, err = uuid.NewV7()
	}
	return
}
