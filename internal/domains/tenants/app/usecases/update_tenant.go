package usecases

import (
	"context"
	"strings"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/model"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"

	"github.com/google/uuid"
)

// UpdateTenant handles the use case of updating a tenant
type UpdateTenant struct {
	tenantRepo outbound.TenantRepository
}

// NewUpdateTenant creates a new UpdateTenant use case
func NewUpdateTenant(tenantRepo outbound.TenantRepository) *UpdateTenant {
	return &UpdateTenant{
		tenantRepo: tenantRepo,
	}
}

// UpdateTenantRequest represents the request to update a tenant
type UpdateTenantRequest struct {
	TenantID        uuid.UUID
	Name            *string
	Slug            *string
	Status          *model.TenantStatus
	Tier            *model.Tier
	AgencySeatLimit *int
}

// UpdateTenantResponse represents the response from updating a tenant
type UpdateTenantResponse struct {
	Tenant *model.Tenant
}

// Execute executes the use case
func (uc *UpdateTenant) Execute(ctx context.Context, req *UpdateTenantRequest) (*UpdateTenantResponse, error) {
	// Get existing tenant
	tenant, err := uc.tenantRepo.FindByID(ctx, req.TenantID)
	if err != nil {
		return nil, domain.ErrTenantNotFound
	}

	// Update fields if provided
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, domain.ErrInvalidTenantName
		}
		tenant.SetName(name)
	}

	if req.Slug != nil {
		slug := strings.ToLower(strings.TrimSpace(*req.Slug))
		slug = strings.ReplaceAll(slug, " ", "-")
		if slug == "" {
			return nil, domain.ErrInvalidTenantSlug
		}
		// Check if slug is already taken by another tenant
		existing, err := uc.tenantRepo.FindBySlug(ctx, slug)
		if err == nil && existing != nil && existing.ID() != tenant.ID() {
			return nil, domain.ErrTenantAlreadyExists
		}
		tenant.SetSlug(slug)
	}

	if req.Status != nil {
		tenant.SetStatus(*req.Status)
	}

	if req.Tier != nil {
		if !model.IsValidTier(*req.Tier) {
			return nil, domain.ErrInvalidRole // TODO: create ErrInvalidTier
		}
		tenant.SetTier(req.Tier)
	}

	if req.AgencySeatLimit != nil {
		tenant.SetAgencySeatLimit(*req.AgencySeatLimit)
	}

	// Save updated tenant
	if err := uc.tenantRepo.Update(ctx, tenant); err != nil {
		return nil, err
	}

	return &UpdateTenantResponse{
		Tenant: tenant,
	}, nil
}

