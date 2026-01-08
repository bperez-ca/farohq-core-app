package usecases

import (
	"context"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"

	"github.com/google/uuid"
)

// ListRoles handles the use case of listing available roles for a tenant
type ListRoles struct {
	tenantRepo outbound.TenantRepository
}

// NewListRoles creates a new ListRoles use case
func NewListRoles(tenantRepo outbound.TenantRepository) *ListRoles {
	return &ListRoles{
		tenantRepo: tenantRepo,
	}
}

// ListRolesRequest represents the request to list roles
type ListRolesRequest struct {
	TenantID uuid.UUID
}

// RoleInfo represents information about a role
type RoleInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// ListRolesResponse represents the response from listing roles
type ListRolesResponse struct {
	Roles []RoleInfo
}

// Execute executes the use case
func (uc *ListRoles) Execute(ctx context.Context, req *ListRolesRequest) (*ListRolesResponse, error) {
	// Verify tenant exists
	_, err := uc.tenantRepo.FindByID(ctx, req.TenantID)
	if err != nil {
		return nil, domain.ErrTenantNotFound
	}

	// Return available roles with descriptions
	roles := []RoleInfo{
		{
			Name:        "owner",
			Description: "Full access to tenant settings and members",
			Permissions: []string{"manage_tenant", "manage_members", "manage_branding", "view_all"},
		},
		{
			Name:        "admin",
			Description: "Can manage members and most tenant settings",
			Permissions: []string{"manage_members", "manage_branding", "view_all"},
		},
		{
			Name:        "staff",
			Description: "Can manage content and view tenant data",
			Permissions: []string{"manage_content", "view_all"},
		},
		{
			Name:        "viewer",
			Description: "Read-only access to tenant data",
			Permissions: []string{"view_all"},
		},
	}

	return &ListRolesResponse{
		Roles: roles,
	}, nil
}

