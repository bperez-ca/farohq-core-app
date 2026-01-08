package usecases

import (
	"context"
	"strings"

	"farohq-core-app/internal/domains/brand/domain"
	"farohq-core-app/internal/domains/brand/domain/ports/inbound"
	"farohq-core-app/internal/domains/brand/domain/ports/outbound"
	tenants_outbound "farohq-core-app/internal/domains/tenants/domain/ports/outbound"
)

// GetByHost implements the GetByHost inbound port
type GetByHost struct {
	brandRepo  outbound.BrandRepository
	tenantRepo tenants_outbound.TenantRepository
}

// NewGetByHost creates a new GetByHost use case
func NewGetByHost(brandRepo outbound.BrandRepository, tenantRepo tenants_outbound.TenantRepository) inbound.GetByHost {
	return &GetByHost{
		brandRepo:  brandRepo,
		tenantRepo: tenantRepo,
	}
}

// Execute executes the use case
func (uc *GetByHost) Execute(ctx context.Context, req *inbound.GetByHostRequest) (*inbound.GetByHostResponse, error) {
	// Extract slug from host (remove port and extract subdomain)
	hostParts := strings.Split(req.Host, ":")
	hostDomain := hostParts[0]

	// Extract subdomain (slug)
	domainParts := strings.Split(hostDomain, ".")
	var slug string
	if len(domainParts) > 2 {
		slug = domainParts[0]
	} else {
		// Default to main domain
		slug = "main"
	}

	// Query database for agency by slug, then get branding
	tenant, err := uc.tenantRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, domain.ErrBrandingNotFound
	}

	branding, err := uc.brandRepo.FindByAgencyID(ctx, tenant.ID())
	if err != nil {
		return nil, domain.ErrBrandingNotFound
	}

	return &inbound.GetByHostResponse{
		Branding: branding,
	}, nil
}

