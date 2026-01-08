package usecases

import (
	"context"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"

	"github.com/google/uuid"
)

// RemoveMember handles the use case of removing a member from a tenant
type RemoveMember struct {
	memberRepo outbound.TenantMemberRepository
	tenantRepo outbound.TenantRepository
}

// NewRemoveMember creates a new RemoveMember use case
func NewRemoveMember(memberRepo outbound.TenantMemberRepository, tenantRepo outbound.TenantRepository) *RemoveMember {
	return &RemoveMember{
		memberRepo: memberRepo,
		tenantRepo: tenantRepo,
	}
}

// RemoveMemberRequest represents the request to remove a member
type RemoveMemberRequest struct {
	TenantID uuid.UUID
	UserID   uuid.UUID
}

// RemoveMemberResponse represents the response from removing a member
type RemoveMemberResponse struct {
	Success bool
}

// Execute executes the use case
func (uc *RemoveMember) Execute(ctx context.Context, req *RemoveMemberRequest) (*RemoveMemberResponse, error) {
	// Verify tenant exists
	_, err := uc.tenantRepo.FindByID(ctx, req.TenantID)
	if err != nil {
		return nil, domain.ErrTenantNotFound
	}

	// Find member
	member, err := uc.memberRepo.FindByTenantAndUserID(ctx, req.TenantID, req.UserID)
	if err != nil {
		return nil, domain.ErrMemberNotFound
	}

	// Delete member
	if err := uc.memberRepo.Delete(ctx, member.ID()); err != nil {
		return nil, err
	}

	return &RemoveMemberResponse{
		Success: true,
	}, nil
}

