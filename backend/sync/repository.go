package sync

import (
	"context"

	"github.com/doron-cohen/argus/backend/internal/storage"
)

// Repository defines the interface for storage operations needed by the sync service
type Repository interface {
	GetComponentByName(ctx context.Context, name string) (*storage.Component, error)
	CreateComponent(ctx context.Context, component storage.Component) error
}

// Ensure storage.Repository implements our interface
var _ Repository = (*storage.Repository)(nil)
