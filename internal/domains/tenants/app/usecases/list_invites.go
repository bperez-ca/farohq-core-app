package usecases

import (
	"context"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/model"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"

	"github.com/google/uuid"
)

// ListInvites handles the use case of listing all invites for a tenant
type ListInvites struct {
	inviteRepo outbound.InviteRepository
	tenantRepo outbound.TenantRepository
}

// NewListInvites creates a new ListInvites use case
func NewListInvites(
	inviteRepo outbound.InviteRepository,
	tenantRepo outbound.TenantRepository,
) *ListInvites {
	return &ListInvites{
		inviteRepo: inviteRepo,
		tenantRepo: tenantRepo,
	}
}

// ListInvitesRequest represents the request to list invites
type ListInvitesRequest struct {
	TenantID uuid.UUID
}

// ListInvitesResponse represents the response from listing invites
type ListInvitesResponse struct {
	Invites []*model.Invite
}

// Execute executes the use case
func (uc *ListInvites) Execute(ctx context.Context, req *ListInvitesRequest) (*ListInvitesResponse, error) {
	// Verify tenant exists
	_, err := uc.tenantRepo.FindByID(ctx, req.TenantID)
	if err != nil {
		return nil, domain.ErrTenantNotFound
	}

	// Find all invites for the tenant
	invites, err := uc.inviteRepo.FindByTenantID(ctx, req.TenantID)
	if err != nil {
		return nil, err
	}

	return &ListInvitesResponse{
		Invites: invites,
	}, nil
}
