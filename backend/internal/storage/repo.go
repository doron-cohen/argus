package storage

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Component struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name string    `gorm:"not null;unique"`
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
