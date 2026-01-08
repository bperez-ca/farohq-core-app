package usecases

import (
	"context"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/model"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"

	"github.com/google/uuid"
)

// ListMembers handles the use case of listing tenant members
type ListMembers struct {
	memberRepo outbound.TenantMemberRepository
	tenantRepo outbound.TenantRepository
}

// NewListMembers creates a new ListMembers use case
func NewListMembers(memberRepo outbound.TenantMemberRepository, tenantRepo outbound.TenantRepository) *ListMembers {
	return &ListMembers{
		memberRepo: memberRepo,
		tenantRepo: tenantRepo,
	}
}

// ListMembersRequest represents the request to list members
type ListMembersRequest struct {
	TenantID uuid.UUID
}

// ListMembersResponse represents the response from listing members
type ListMembersResponse struct {
	Members []*model.TenantMember
}

// Execute executes the use case
func (uc *ListMembers) Execute(ctx context.Context, req *ListMembersRequest) (*ListMembersResponse, error) {
	// Verify tenant exists
	_, err := uc.tenantRepo.FindByID(ctx, req.TenantID)
	if err != nil {
		return nil, domain.ErrTenantNotFound
	}

	// Get members
	members, err := uc.memberRepo.FindByTenantID(ctx, req.TenantID)
	if err != nil {
		return nil, err
	}

	return &ListMembersResponse{
		Members: members,
	}, nil
}

