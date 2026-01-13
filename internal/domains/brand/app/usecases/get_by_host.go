package usecases

import (
	"context"
	"strings"

	"farohq-core-app/internal/domains/brand/domain"
	"farohq-core-app/internal/domains/brand/domain/model"
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
	// Extract host domain (remove port if present)
	hostParts := strings.Split(req.Host, ":")
	hostDomain := hostParts[0]
	hostDomain = strings.ToLower(hostDomain)

	var branding *model.Branding
	var err error

	// Determine if host is subdomain or custom domain
	if strings.HasSuffix(hostDomain, ".portal.farohq.com") {
		// Subdomain resolution: {slug}.portal.farohq.com
		// Extract full subdomain (including .portal.farohq.com)
		subdomain := hostDomain
		branding, err = uc.brandRepo.FindBySubdomain(ctx, subdomain)
		if err != nil {
			return nil, domain.ErrBrandingNotFound
		}
	} else {
		// Custom domain resolution: portal.agency.com
		// Query by domain field
		branding, err = uc.brandRepo.FindByDomain(ctx, hostDomain)
		if err != nil {
			return nil, domain.ErrBrandingNotFound
		}
	}

	return &inbound.GetByHostResponse{
		Branding: branding,
	}, nil
}

