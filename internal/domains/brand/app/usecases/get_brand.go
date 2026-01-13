package usecases

import (
	"context"

	"farohq-core-app/internal/domains/brand/domain"
	"farohq-core-app/internal/domains/brand/domain/model"
	"farohq-core-app/internal/domains/brand/domain/ports/inbound"
	"farohq-core-app/internal/domains/brand/domain/ports/outbound"
	tenants_outbound "farohq-core-app/internal/domains/tenants/domain/ports/outbound"
	tenants_model "farohq-core-app/internal/domains/tenants/domain/model"

	"github.com/google/uuid"
)

// GetBrand implements the GetBrand inbound port
type GetBrand struct {
	brandRepo  outbound.BrandRepository
	tenantRepo tenants_outbound.TenantRepository
}

// NewGetBrand creates a new GetBrand use case
func NewGetBrand(brandRepo outbound.BrandRepository, tenantRepo tenants_outbound.TenantRepository) inbound.GetBrand {
	return &GetBrand{
		brandRepo:  brandRepo,
		tenantRepo: tenantRepo,
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

	// Join with agencies/tenants table to get tier information
	tenant, err := uc.tenantRepo.FindByID(ctx, branding.AgencyID())
	if err != nil {
		// If tenant not found, continue without tier info (shouldn't happen, but handle gracefully)
		return &inbound.GetBrandResponse{
			Branding: branding,
		}, nil
	}

	tier := tenant.Tier()

	// Apply tier-based rules for "Powered by Faro" badge (Growth+ tiers only)
	// If tier doesn't allow hiding badge, ensure hide_powered_by is false
	if !tenants_model.TierCanHidePoweredBy(tier) && branding.HidePoweredBy() {
		// Reset hide_powered_by if tier doesn't allow it (even if stored as true)
		branding.SetHidePoweredBy(false)
	}

	// Apply tier-based rules for custom domain support (Scale tier only)
	// If tier doesn't support custom domains, ensure domain_type is 'subdomain' and domain is empty
	if !tenants_model.TierSupportsCustomDomain(tier) {
		// Lower tiers should use subdomain, not custom domain
		if branding.DomainType() != nil && *branding.DomainType() == model.DomainTypeCustom {
			// Reset to subdomain for lower tiers
			dt := model.DomainTypeSubdomain
			branding.SetDomainType(&dt)
			branding.SetDomain("") // Clear custom domain for lower tiers
		}
	}

	return &inbound.GetBrandResponse{
		Branding: branding,
	}, nil
}

