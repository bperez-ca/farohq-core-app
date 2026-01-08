package usecases

import (
	"context"

	"farohq-core-app/internal/domains/brand/domain/model"
	"farohq-core-app/internal/domains/brand/domain/ports/inbound"
	"farohq-core-app/internal/domains/brand/domain/ports/outbound"

	"github.com/google/uuid"
)

// CreateBrand implements the CreateBrand inbound port
type CreateBrand struct {
	brandRepo outbound.BrandRepository
}

// NewCreateBrand creates a new CreateBrand use case
func NewCreateBrand(brandRepo outbound.BrandRepository) inbound.CreateBrand {
	return &CreateBrand{
		brandRepo: brandRepo,
	}
}

// Execute executes the use case
func (uc *CreateBrand) Execute(ctx context.Context, req *inbound.CreateBrandRequest) (*inbound.CreateBrandResponse, error) {
	agencyID, err := uuid.Parse(req.AgencyID)
	if err != nil {
		return nil, err
	}

	branding := model.NewBranding(
		agencyID,
		req.Domain,
		req.LogoURL,
		req.FaviconURL,
		req.PrimaryColor,
		req.SecondaryColor,
		req.ThemeJSON,
	)

	if err := uc.brandRepo.Save(ctx, branding); err != nil {
		return nil, err
	}

	return &inbound.CreateBrandResponse{
		Branding: branding,
	}, nil
}

