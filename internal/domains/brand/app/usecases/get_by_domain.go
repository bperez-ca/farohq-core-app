package usecases

import (
	"context"

	"farohq-core-app/internal/domains/brand/domain"
	"farohq-core-app/internal/domains/brand/domain/ports/inbound"
	"farohq-core-app/internal/domains/brand/domain/ports/outbound"
)

// GetByDomain implements the GetByDomain inbound port
type GetByDomain struct {
	brandRepo outbound.BrandRepository
}

// NewGetByDomain creates a new GetByDomain use case
func NewGetByDomain(brandRepo outbound.BrandRepository) inbound.GetByDomain {
	return &GetByDomain{
		brandRepo: brandRepo,
	}
}

// Execute executes the use case
func (uc *GetByDomain) Execute(ctx context.Context, req *inbound.GetByDomainRequest) (*inbound.GetByDomainResponse, error) {
	branding, err := uc.brandRepo.FindByDomain(ctx, req.Domain)
	if err != nil {
		return nil, domain.ErrBrandingNotFound
	}

	return &inbound.GetByDomainResponse{
		Branding: branding,
	}, nil
}

