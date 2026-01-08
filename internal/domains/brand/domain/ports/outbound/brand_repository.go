package outbound

import (
	"context"

	"farohq-core-app/internal/domains/brand/domain/model"

	"github.com/google/uuid"
)

// BrandRepository defines the interface for brand data access
type BrandRepository interface {
	FindByAgencyID(ctx context.Context, agencyID uuid.UUID) (*model.Branding, error)
	FindByDomain(ctx context.Context, domain string) (*model.Branding, error)
	Save(ctx context.Context, branding *model.Branding) error
	Update(ctx context.Context, branding *model.Branding) error
	Delete(ctx context.Context, agencyID uuid.UUID) error
	ListByAgencyID(ctx context.Context, agencyID uuid.UUID) ([]*model.Branding, error)
}

