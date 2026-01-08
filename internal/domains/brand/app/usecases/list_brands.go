package usecases

import (
	"context"

	"farohq-core-app/internal/domains/brand/domain/ports/inbound"
	"farohq-core-app/internal/domains/brand/domain/ports/outbound"

	"github.com/google/uuid"
)

// ListBrands implements the ListBrands inbound port
type ListBrands struct {
	brandRepo outbound.BrandRepository
}

// NewListBrands creates a new ListBrands use case
func NewListBrands(brandRepo outbound.BrandRepository) inbound.ListBrands {
	return &ListBrands{
		brandRepo: brandRepo,
	}
}

// Execute executes the use case
func (uc *ListBrands) Execute(ctx context.Context, req *inbound.ListBrandsRequest) (*inbound.ListBrandsResponse, error) {
	agencyID, err := uuid.Parse(req.AgencyID)
	if err != nil {
		return nil, err
	}

	brands, err := uc.brandRepo.ListByAgencyID(ctx, agencyID)
	if err != nil {
		return nil, err
	}

	return &inbound.ListBrandsResponse{
		Brands: brands,
	}, nil
}

