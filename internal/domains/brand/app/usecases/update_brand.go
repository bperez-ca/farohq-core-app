package usecases

import (
	"context"
	"errors"
	"fmt"

	"farohq-core-app/internal/domains/brand/domain"
	"farohq-core-app/internal/domains/brand/domain/model"
	"farohq-core-app/internal/domains/brand/domain/ports/inbound"
	"farohq-core-app/internal/domains/brand/domain/ports/outbound"
	tenants_outbound "farohq-core-app/internal/domains/tenants/domain/ports/outbound"
	tenants_model "farohq-core-app/internal/domains/tenants/domain/model"

	"github.com/google/uuid"
)

// UpdateBrand implements the UpdateBrand inbound port
type UpdateBrand struct {
	brandRepo  outbound.BrandRepository
	tenantRepo tenants_outbound.TenantRepository
}

// NewUpdateBrand creates a new UpdateBrand use case
func NewUpdateBrand(brandRepo outbound.BrandRepository, tenantRepo tenants_outbound.TenantRepository) inbound.UpdateBrand {
	return &UpdateBrand{
		brandRepo:  brandRepo,
		tenantRepo: tenantRepo,
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

	// Get tenant/agency to check tier for validation
	tenant, err := uc.tenantRepo.FindByID(ctx, branding.AgencyID())
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	tier := tenant.Tier()

	// Update website (always allowed, can be updated at any time)
	if req.Website != nil {
		branding.SetWebsite(*req.Website)
	}

	// Update hide_powered_by with tier validation (Growth+ tiers only)
	if req.HidePoweredBy != nil {
		if !tenants_model.TierCanHidePoweredBy(tier) {
			return nil, errors.New("Hide 'Powered by Faro' badge is only available for Growth+ tiers")
		}
		branding.SetHidePoweredBy(*req.HidePoweredBy)
	}

	// Update domain with tier validation (Scale tier only)
	if req.Domain != nil {
		if !tenants_model.TierSupportsCustomDomain(tier) {
			return nil, errors.New("Custom domain support is only available for Scale tier")
		}

		// Scale tier: Allow custom domain configuration
		if *req.Domain != "" {
			// Set custom domain
			branding.SetDomain(*req.Domain)
			dt := model.DomainTypeCustom
			branding.SetDomainType(&dt)
			// Reset verification status when domain changes
			// Note: verified_at will be set by domain verification use case
			// Note: ssl_status will be set to 'pending' by domain verification use case
		} else {
			// Clear custom domain (use subdomain as fallback)
			branding.SetDomain("")
			dt := model.DomainTypeSubdomain
			branding.SetDomainType(&dt)
			// Reset verification status
			// Note: ssl_status will be cleared (set to NULL)
		}
	}

	// Update other fields if provided
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

