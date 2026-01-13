package usecases

import (
	"context"
	"errors"
	"fmt"

	"farohq-core-app/internal/domains/brand/domain"
	"farohq-core-app/internal/domains/brand/domain/ports/inbound"
	"farohq-core-app/internal/domains/brand/domain/ports/outbound"
	"farohq-core-app/internal/domains/brand/infra/vercel"
	tenants_outbound "farohq-core-app/internal/domains/tenants/domain/ports/outbound"
	tenants_model "farohq-core-app/internal/domains/tenants/domain/model"

	"github.com/google/uuid"
)

// GetDomainInstructions implements the GetDomainInstructions inbound port
type GetDomainInstructions struct {
	brandRepo     outbound.BrandRepository
	tenantRepo    tenants_outbound.TenantRepository
	vercelService *vercel.VercelService
}

// NewGetDomainInstructions creates a new GetDomainInstructions use case
func NewGetDomainInstructions(
	brandRepo outbound.BrandRepository,
	tenantRepo tenants_outbound.TenantRepository,
	vercelService *vercel.VercelService,
) inbound.GetDomainInstructions {
	return &GetDomainInstructions{
		brandRepo:     brandRepo,
		tenantRepo:    tenantRepo,
		vercelService: vercelService,
	}
}

// Execute executes the use case
func (uc *GetDomainInstructions) Execute(ctx context.Context, req *inbound.GetDomainInstructionsRequest) (*inbound.GetDomainInstructionsResponse, error) {
	agencyID, err := uuid.Parse(req.BrandID)
	if err != nil {
		return nil, domain.ErrBrandingNotFound
	}

	// Tier validation: Check if tenant tier is Scale
	tenant, err := uc.tenantRepo.FindByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	tier := tenant.Tier()
	if !tenants_model.TierSupportsCustomDomain(tier) {
		return nil, errors.New("Custom domain support is only available for Scale tier")
	}

	branding, err := uc.brandRepo.FindByAgencyID(ctx, agencyID)
	if err != nil {
		return nil, domain.ErrBrandingNotFound
	}

	domain := branding.Domain()
	if domain == "" {
		// Try to use website if available
		website := branding.Website()
		if website == "" {
			return nil, errors.New("no custom domain configured and no website available")
		}
		// Extract domain from website URL
		domain = website
	}

	// Fetch expected DNS configuration from Vercel API (CNAME target and other records)
	// Do NOT hardcode CNAME target - always fetch from Vercel API (value may vary)
	dnsConfig, err := uc.vercelService.GetExpectedDNSConfig(ctx, domain)
	if err != nil {
		// If domain not added yet, try to add it to get expected config
		dnsConfig, err = uc.vercelService.AddDomainToProject(ctx, domain)
		if err != nil {
			return nil, fmt.Errorf("failed to get DNS configuration from Vercel API: %w", err)
		}
	}

	cnameTarget := ""
	if dnsConfig != nil {
		cnameTarget = dnsConfig.CNAMETarget
	}

	if cnameTarget == "" {
		return nil, errors.New("no CNAME target found in Vercel API response")
	}

	// Generate human-readable instructions
	instructions := fmt.Sprintf(
		"Create a CNAME record in your DNS provider:\n\n"+
			"Record Type: CNAME\n"+
			"Name/Host: %s (or @ for root domain)\n"+
			"Value/Target: %s\n\n"+
			"After creating the CNAME record, DNS propagation can take up to 48 hours globally. "+
			"Click 'Verify Domain' once you've created the record.",
		domain,
		cnameTarget,
	)

	return &inbound.GetDomainInstructionsResponse{
		Domain:       domain,
		CNAMETarget:  cnameTarget,
		Instructions: instructions,
	}, nil
}
