package usecases

import (
	"context"

	"farohq-core-app/internal/domains/brand/domain"
	"farohq-core-app/internal/domains/brand/domain/ports/inbound"
	"farohq-core-app/internal/domains/brand/domain/ports/outbound"

	"github.com/google/uuid"
)

// UpdateBrand implements the UpdateBrand inbound port
type UpdateBrand struct {
	brandRepo outbound.BrandRepository
}

// NewUpdateBrand creates a new UpdateBrand use case
func NewUpdateBrand(brandRepo outbound.BrandRepository) inbound.UpdateBrand {
	return &UpdateBrand{
		brandRepo: brandRepo,
	}
}

// Execute executes the use case
func (uc *UpdateBrand) Execute(ctx context.Context, req *inbound.UpdateBrandRequest) (*inbound.UpdateBrandResponse, error) {
	brandID, err := uuid.Parse(req.BrandID)
	if err != nil {
		return nil, domain.ErrBrandingNotFound
	}

	branding, err := uc.brandRepo.FindByAgencyID(ctx, brandID)
	if err != nil {
		return nil, domain.ErrBrandingNotFound
	}

	// Update fields if provided
	if req.Domain != nil {
		branding.SetDomain(*req.Domain)
	}
	if req.LogoURL != nil {
		branding.SetLogoURL(*req.LogoURL)
	}
	if req.FaviconURL != nil {
		branding.SetFaviconURL(*req.FaviconURL)
	}
	if req.PrimaryColor != nil {
		branding.SetPrimaryColor(*req.PrimaryColor)
	}
	if req.SecondaryColor != nil {
		branding.SetSecondaryColor(*req.SecondaryColor)
	}
	if req.ThemeJSON != nil {
		branding.SetThemeJSON(*req.ThemeJSON)
	}

	if err := uc.brandRepo.Update(ctx, branding); err != nil {
		return nil, err
	}

	return &inbound.UpdateBrandResponse{
		Branding: branding,
	}, nil
}

