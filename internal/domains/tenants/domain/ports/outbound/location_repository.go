package outbound

import (
	"context"

	"farohq-core-app/internal/domains/tenants/domain/model"

	"github.com/google/uuid"
)

// LocationRepository defines the interface for location persistence
type LocationRepository interface {
	// Save saves or updates a location
	Save(ctx context.Context, location *model.Location) error

	// FindByID finds a location by ID
	FindByID(ctx context.Context, id uuid.UUID) (*model.Location, error)

	// ListByClient lists all locations for a client
	ListByClient(ctx context.Context, clientID uuid.UUID) ([]*model.Location, error)

	// CountByClient counts locations for a client (excluding soft-deleted)
	CountByClient(ctx context.Context, clientID uuid.UUID) (int, error)
}

