package usecases

import (
	"context"
	"errors"
	"fmt"

	"farohq-core-app/internal/domains/brand/domain"
	"farohq-core-app/internal/domains/brand/domain/model"
	"farohq-core-app/internal/domains/brand/domain/ports/inbound"
	"farohq-core-app/internal/domains/brand/domain/ports/outbound"
	"farohq-core-app/internal/domains/brand/infra/vercel"
	tenants_outbound "farohq-core-app/internal/domains/tenants/domain/ports/outbound"
	tenants_model "farohq-core-app/internal/domains/tenants/domain/model"

	"github.com/google/uuid"
)

// GetDomainStatus implements the GetDomainStatus inbound port
type GetDomainStatus struct {
	brandRepo     outbound.BrandRepository
	tenantRepo    tenants_outbound.TenantRepository
	vercelService *vercel.VercelService
}

// NewGetDomainStatus creates a new GetDomainStatus use case
func NewGetDomainStatus(
	brandRepo outbound.BrandRepository,
	tenantRepo tenants_outbound.TenantRepository,
	vercelService *vercel.VercelService,
) inbound.GetDomainStatus {
	return &GetDomainStatus{
		brandRepo:     brandRepo,
		tenantRepo:    tenantRepo,
		vercelService: vercelService,
	}
}

// Execute executes the use case
func (uc *GetDomainStatus) Execute(ctx context.Context, req *inbound.GetDomainStatusRequest) (*inbound.GetDomainStatusResponse, error) {
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
		return nil, errors.New("no custom domain configured")
	}

	// Poll Vercel API for full domain status (verification, SSL readiness)
	domainStatus, err := uc.vercelService.GetDomainStatus(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain status from Vercel API: %w", err)
	}

	// Update brand record with latest status from Vercel API
	if domainStatus.Verified && branding.VerifiedAt() == nil {
		branding.Verify()
	}

	var sslStatus *model.SSLStatus
	if domainStatus.SSL.Status != "" {
		ss := model.SSLStatus(domainStatus.SSL.Status)
		sslStatus = &ss
		branding.SetSSLStatus(sslStatus)
	}

	// Save updated status
	if err := uc.brandRepo.Update(ctx, branding); err != nil {
		return nil, fmt.Errorf("failed to update branding: %w", err)
	}

	// Get expected CNAME target from Vercel API
	expectedCNAME := ""
	dnsConfig, err := uc.vercelService.GetExpectedDNSConfig(ctx, domain)
	if err == nil && dnsConfig != nil {
		expectedCNAME = dnsConfig.CNAMETarget
	}

	sslStatusStr := ""
	if sslStatus != nil {
		sslStatusStr = string(*sslStatus)
	}

	return &inbound.GetDomainStatusResponse{
		Branding:      branding,
		Verified:      domainStatus.Verified,
		ExpectedCNAME: expectedCNAME,
		SSLStatus:     sslStatusStr,
	}, nil
}
