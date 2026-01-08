package usecases

import (
	"context"

	"farohq-core-app/internal/domains/brand/domain"
	"farohq-core-app/internal/domains/brand/domain/ports/inbound"
	"farohq-core-app/internal/domains/brand/domain/ports/outbound"

	"github.com/google/uuid"
)

// DeleteBrand implements the DeleteBrand inbound port
type DeleteBrand struct {
	brandRepo outbound.BrandRepository
}

// NewDeleteBrand creates a new DeleteBrand use case
func NewDeleteBrand(brandRepo outbound.BrandRepository) inbound.DeleteBrand {
	return &DeleteBrand{
		brandRepo: brandRepo,
	}
}

// Execute executes the use case
func (uc *DeleteBrand) Execute(ctx context.Context, req *inbound.DeleteBrandRequest) (*inbound.DeleteBrandResponse, error) {
	brandID, err := uuid.Parse(req.BrandID)
	if err != nil {
		return nil, domain.ErrBrandingNotFound
	}

	if err := uc.brandRepo.Delete(ctx, brandID); err != nil {
		return nil, err
	}

	return &inbound.DeleteBrandResponse{
		Success: true,
	}, nil
}

