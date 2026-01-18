package usecases

import (
	"context"
	"strings"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/model"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"

	"github.com/google/uuid"
)

// OnboardTenant handles the use case of onboarding a new tenant (creating tenant + adding user as owner)
type OnboardTenant struct {
	tenantRepo  outbound.TenantRepository
	memberRepo  outbound.TenantMemberRepository
}

// NewOnboardTenant creates a new OnboardTenant use case
func NewOnboardTenant(tenantRepo outbound.TenantRepository, memberRepo outbound.TenantMemberRepository) *OnboardTenant {
	return &OnboardTenant{
		tenantRepo: tenantRepo,
		memberRepo: memberRepo,
	}
}

// OnboardTenantRequest represents the request to onboard a tenant
type OnboardTenantRequest struct {
	Name            string
	Slug            string
	Tier            *model.Tier
	AgencySeatLimit int
	UserID          uuid.UUID // User to add as owner
}

// OnboardTenantResponse represents the response from onboarding a tenant
type OnboardTenantResponse struct {
	Tenant *model.Tenant
}

// Execute executes the use case
func (uc *OnboardTenant) Execute(ctx context.Context, req *OnboardTenantRequest) (*OnboardTenantResponse, error) {
	// Validate input
	if strings.TrimSpace(req.Name) == "" {
		return nil, domain.ErrInvalidTenantName
	}

	if strings.TrimSpace(req.Slug) == "" {
		return nil, domain.ErrInvalidTenantSlug
	}

	// Normalize slug
	slug := strings.ToLower(strings.TrimSpace(req.Slug))
	slug = strings.ReplaceAll(slug, " ", "-")

	// Check if tenant with slug already exists
	existing, err := uc.tenantRepo.FindBySlug(ctx, slug)
	if err == nil && existing != nil {
		return nil, domain.ErrTenantAlreadyExists
	}

	// Create new tenant (inviteExpiryHours defaults to 24 hours if nil)
	tenant := model.NewTenant(strings.TrimSpace(req.Name), slug, req.Tier, req.AgencySeatLimit, nil)

	// Save tenant
	if err := uc.tenantRepo.Save(ctx, tenant); err != nil {
		return nil, err
	}

	// Create tenant member with owner role
	member := model.NewTenantMember(tenant.ID(), req.UserID, model.RoleOwner)
	if err := uc.memberRepo.Save(ctx, member); err != nil {
		// If member creation fails, we should rollback tenant creation
		// For now, log error but continue (tenant is created, user can be added later)
		// In production, consider using transactions
		return nil, err
	}

	return &OnboardTenantResponse{
		Tenant: tenant,
	}, nil
}
