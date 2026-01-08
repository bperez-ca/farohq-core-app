package usecases

import (
	"context"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/model"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"

	"github.com/google/uuid"
)

// GetTenant handles the use case of getting a tenant by ID
type GetTenant struct {
	tenantRepo outbound.TenantRepository
}

// NewGetTenant creates a new GetTenant use case
func NewGetTenant(tenantRepo outbound.TenantRepository) *GetTenant {
	return &GetTenant{
		tenantRepo: tenantRepo,
	}
}

// GetTenantRequest represents the request to get a tenant
type GetTenantRequest struct {
	TenantID uuid.UUID
}

// GetTenantResponse represents the response from getting a tenant
type GetTenantResponse struct {
	Tenant *model.Tenant
}

// Execute executes the use case
func (uc *GetTenant) Execute(ctx context.Context, req *GetTenantRequest) (*GetTenantResponse, error) {
	tenant, err := uc.tenantRepo.FindByID(ctx, req.TenantID)
	if err != nil {
		return nil, domain.ErrTenantNotFound
	}

	return &GetTenantResponse{
		Tenant: tenant,
	}, nil
}

