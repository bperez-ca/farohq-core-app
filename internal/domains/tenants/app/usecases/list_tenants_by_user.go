package usecases

import (
	"context"

	"farohq-core-app/internal/domains/tenants/domain/model"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"

	"github.com/google/uuid"
)

// ListTenantsByUser handles the use case of listing tenants for a user
type ListTenantsByUser struct {
	memberRepo outbound.TenantMemberRepository
	tenantRepo outbound.TenantRepository
}

// NewListTenantsByUser creates a new ListTenantsByUser use case
func NewListTenantsByUser(memberRepo outbound.TenantMemberRepository, tenantRepo outbound.TenantRepository) *ListTenantsByUser {
	return &ListTenantsByUser{
		memberRepo: memberRepo,
		tenantRepo: tenantRepo,
	}
}

// ListTenantsByUserRequest represents the request to list tenants for a user
type ListTenantsByUserRequest struct {
	UserID uuid.UUID
}

// TenantWithRole represents a tenant with the user's role in it
type TenantWithRole struct {
	Tenant *model.Tenant
	Role   model.Role
}

// ListTenantsByUserResponse represents the response from listing tenants for a user
type ListTenantsByUserResponse struct {
	Tenants []TenantWithRole
}

// Execute executes the use case
func (uc *ListTenantsByUser) Execute(ctx context.Context, req *ListTenantsByUserRequest) (*ListTenantsByUserResponse, error) {
	// Get all tenant members for the user
	members, err := uc.memberRepo.FindByUserID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	// Fetch tenant details for each member
	tenants := make([]TenantWithRole, 0, len(members))
	for _, member := range members {
		tenant, err := uc.tenantRepo.FindByID(ctx, member.TenantID())
		if err != nil {
			// Skip if tenant not found (shouldn't happen but handle gracefully)
			continue
		}

		tenants = append(tenants, TenantWithRole{
			Tenant: tenant,
			Role:   member.Role(),
		})
	}

	return &ListTenantsByUserResponse{
		Tenants: tenants,
	}, nil
}
