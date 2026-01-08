package outbound

import (
	"context"

	"farohq-core-app/internal/domains/tenants/domain/model"

	"github.com/google/uuid"
)

// ClientRepository defines the interface for client persistence
type ClientRepository interface {
	// Save saves or updates a client
	Save(ctx context.Context, client *model.Client) error

	// FindByID finds a client by ID
	FindByID(ctx context.Context, id uuid.UUID) (*model.Client, error)

	// FindBySlug finds a client by slug within an agency
	FindBySlug(ctx context.Context, agencyID uuid.UUID, slug string) (*model.Client, error)

	// ListByAgency lists all clients for an agency
	ListByAgency(ctx context.Context, agencyID uuid.UUID) ([]*model.Client, error)

	// CountByAgency counts clients for an agency (excluding soft-deleted)
	CountByAgency(ctx context.Context, agencyID uuid.UUID) (int, error)

	// CountByAgencyAndTier counts clients for an agency by tier (excluding soft-deleted)
	CountByAgencyAndTier(ctx context.Context, agencyID uuid.UUID, tier model.Tier) (int, error)
}

