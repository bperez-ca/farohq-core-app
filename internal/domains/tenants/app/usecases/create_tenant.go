package usecases

import (
	"context"
	"strings"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/model"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"
)

// CreateTenant handles the use case of creating a new tenant
type CreateTenant struct {
	tenantRepo outbound.TenantRepository
}

// NewCreateTenant creates a new CreateTenant use case
func NewCreateTenant(tenantRepo outbound.TenantRepository) *CreateTenant {
	return &CreateTenant{
		tenantRepo: tenantRepo,
	}
}

// CreateTenantRequest represents the request to create a tenant
type CreateTenantRequest struct {
	Name            string
	Slug            string
	Tier            *model.Tier
	AgencySeatLimit int
}

// CreateTenantResponse represents the response from creating a tenant
type CreateTenantResponse struct {
	Tenant *model.Tenant
}

// Execute executes the use case
func (uc *CreateTenant) Execute(ctx context.Context, req *CreateTenantRequest) (*CreateTenantResponse, error) {
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

	// Create new tenant
	tenant := model.NewTenant(strings.TrimSpace(req.Name), slug, req.Tier, req.AgencySeatLimit)

	// Save tenant
	if err := uc.tenantRepo.Save(ctx, tenant); err != nil {
		return nil, err
	}

	return &CreateTenantResponse{
		Tenant: tenant,
	}, nil
}

