package usecases

import (
	"context"

	"farohq-core-app/internal/domains/brand/domain"
	"farohq-core-app/internal/domains/brand/domain/ports/inbound"
	"farohq-core-app/internal/domains/brand/domain/ports/outbound"

	"github.com/google/uuid"
)

// GetBrand implements the GetBrand inbound port
type GetBrand struct {
	brandRepo outbound.BrandRepository
}

// NewGetBrand creates a new GetBrand use case
func NewGetBrand(brandRepo outbound.BrandRepository) inbound.GetBrand {
	return &GetBrand{
		brandRepo: brandRepo,
	}
}

// Execute executes the use case
func (uc *GetBrand) Execute(ctx context.Context, req *inbound.GetBrandRequest) (*inbound.GetBrandResponse, error) {
	brandID, err := uuid.Parse(req.BrandID)
	if err != nil {
		return nil, domain.ErrBrandingNotFound
	}

	branding, err := uc.brandRepo.FindByAgencyID(ctx, brandID)
	if err != nil {
		return nil, domain.ErrBrandingNotFound
	}

	return &inbound.GetBrandResponse{
		Branding: branding,
	}, nil
}

