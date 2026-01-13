package usecases

import (
	"context"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"
)

// ValidateSlug handles the use case of validating slug availability
type ValidateSlug struct {
	tenantRepo outbound.TenantRepository
}

// NewValidateSlug creates a new ValidateSlug use case
func NewValidateSlug(tenantRepo outbound.TenantRepository) *ValidateSlug {
	return &ValidateSlug{
		tenantRepo: tenantRepo,
	}
}

// ValidateSlugRequest represents the request to validate a slug
type ValidateSlugRequest struct {
	Slug string
}

// ValidateSlugResponse represents the response from validating a slug
type ValidateSlugResponse struct {
	Available bool
	Slug      string
}

// Execute executes the use case
func (uc *ValidateSlug) Execute(ctx context.Context, req *ValidateSlugRequest) (*ValidateSlugResponse, error) {
	// Check if tenant with slug already exists
	existing, err := uc.tenantRepo.FindBySlug(ctx, req.Slug)
	if err == nil && existing != nil {
		// Tenant exists, slug is not available
		return &ValidateSlugResponse{
			Available: false,
			Slug:      req.Slug,
		}, nil
	}

	// If error is ErrTenantNotFound, slug is available
	if err == domain.ErrTenantNotFound {
		return &ValidateSlugResponse{
			Available: true,
			Slug:      req.Slug,
		}, nil
	}

	// For other errors, return error (don't assume available)
	if err != nil {
		return nil, err
	}

	// No error and no tenant found (shouldn't happen but handle gracefully)
	return &ValidateSlugResponse{
		Available: true,
		Slug:      req.Slug,
	}, nil
}
