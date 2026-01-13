package usecases

import (
	"context"
	"errors"
	"fmt"

	"farohq-core-app/internal/domains/brand/domain"
	"farohq-core-app/internal/domains/brand/domain/model"
	"farohq-core-app/internal/domains/brand/domain/ports/inbound"
	"farohq-core-app/internal/domains/brand/domain/ports/outbound"
	"farohq-core-app/internal/domains/brand/infra/dns"
	"farohq-core-app/internal/domains/brand/infra/vercel"
	tenants_outbound "farohq-core-app/internal/domains/tenants/domain/ports/outbound"
	tenants_model "farohq-core-app/internal/domains/tenants/domain/model"

	"github.com/google/uuid"
)

// VerifyDomain implements the VerifyDomain inbound port
type VerifyDomain struct {
	brandRepo     outbound.BrandRepository
	tenantRepo    tenants_outbound.TenantRepository
	vercelService *vercel.VercelService
	dnsService    *dns.DNSService // Optional, for UX feedback only
}

// NewVerifyDomain creates a new VerifyDomain use case
func NewVerifyDomain(
	brandRepo outbound.BrandRepository,
	tenantRepo tenants_outbound.TenantRepository,
	vercelService *vercel.VercelService,
	dnsService *dns.DNSService, // Optional, can be nil
) inbound.VerifyDomain {
	return &VerifyDomain{
		brandRepo:     brandRepo,
		tenantRepo:    tenantRepo,
		vercelService: vercelService,
		dnsService:    dnsService,
	}
}

// Execute executes the use case
func (uc *VerifyDomain) Execute(ctx context.Context, req *inbound.VerifyDomainRequest) (*inbound.VerifyDomainResponse, error) {
	agencyID, err := uuid.Parse(req.BrandID)
	if err != nil {
		return nil, domain.ErrBrandingNotFound
	}

	// Tier Validation (CRITICAL - FIRST CHECK)
	tenant, err := uc.tenantRepo.FindByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	tier := tenant.Tier()
	if !tenants_model.TierSupportsCustomDomain(tier) {
		return nil, errors.New("Custom domain support is only available for Scale tier")
	}

	// Only proceed with domain verification if tier is Scale
	branding, err := uc.brandRepo.FindByAgencyID(ctx, agencyID)
	if err != nil {
		return nil, domain.ErrBrandingNotFound
	}

	// Use domain from request or from branding
	domainToVerify := req.Domain
	if domainToVerify == "" {
		domainToVerify = branding.Domain()
	}
	if domainToVerify == "" {
		return nil, errors.New("domain is required for verification")
	}

	// Primary flow (required): Use VercelService as source of truth (Scale tier only)
	// Step 1: Add domain to Vercel project (if not already added)
	dnsConfig, err := uc.vercelService.AddDomainToProject(ctx, domainToVerify)
	if err != nil {
		// Check if domain already exists (might be 409 conflict or similar)
		// Try to get existing domain status
		status, statusErr := uc.vercelService.GetDomainStatus(ctx, domainToVerify)
		if statusErr != nil {
			return nil, fmt.Errorf("failed to add domain to Vercel: %w", err)
		}
		// Domain already exists, use existing status
		dnsConfig, _ = uc.vercelService.GetExpectedDNSConfig(ctx, domainToVerify)
		if dnsConfig == nil && len(status.Config) > 0 {
			dnsConfig = &vercel.DomainConfig{
				CNAMETarget: status.Config[0].RecordValue,
				RecordType:  "CNAME",
				RecordValue: status.Config[0].RecordValue,
			}
		}
	}

	// Step 2: Verify domain via Vercel API (source of truth)
	verified, err := uc.vercelService.VerifyDomain(ctx, domainToVerify)
	if err != nil {
		return nil, fmt.Errorf("failed to verify domain via Vercel API: %w", err)
	}

	// Step 3: Get full domain status from Vercel API
	domainStatus, err := uc.vercelService.GetDomainStatus(ctx, domainToVerify)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain status from Vercel API: %w", err)
	}

	// Step 4: Update brand verification fields based on Vercel API response
	if verified {
		branding.Verify()
	}

	// Update SSL status from Vercel API
	var sslStatus *model.SSLStatus
	if domainStatus.SSL.Status != "" {
		ss := model.SSLStatus(domainStatus.SSL.Status)
		sslStatus = &ss
		branding.SetSSLStatus(sslStatus)
	}

	// Update domain if it changed
	if branding.Domain() != domainToVerify {
		branding.SetDomain(domainToVerify)
		dt := model.DomainTypeCustom
		branding.SetDomainType(&dt)
	}

	// Save updated branding
	if err := uc.brandRepo.Update(ctx, branding); err != nil {
		return nil, fmt.Errorf("failed to update branding: %w", err)
	}

	// Optional flow (UX only): Use DNSService for UI feedback
	var currentCNAME string
	if uc.dnsService != nil {
		currentCNAME, _ = uc.dnsService.LookupCNAME(ctx, domainToVerify)
		// Note: We don't use this result for verification decisions, only for UI feedback
	}

	expectedCNAME := ""
	if dnsConfig != nil {
		expectedCNAME = dnsConfig.CNAMETarget
	}

	sslStatusStr := ""
	if sslStatus != nil {
		sslStatusStr = string(*sslStatus)
	}

	return &inbound.VerifyDomainResponse{
		Branding:      branding,
		Verified:      verified,
		ExpectedCNAME: expectedCNAME,
		CurrentCNAME:  currentCNAME, // Optional, for UX feedback only
		SSLStatus:     sslStatusStr,
	}, nil
}
