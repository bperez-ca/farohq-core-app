package usecases

import (
	"context"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/model"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"

	"github.com/google/uuid"
)

// RevokeInvite handles the use case of revoking an invitation
type RevokeInvite struct {
	inviteRepo outbound.InviteRepository
	tenantRepo outbound.TenantRepository
}

// NewRevokeInvite creates a new RevokeInvite use case
func NewRevokeInvite(
	inviteRepo outbound.InviteRepository,
	tenantRepo outbound.TenantRepository,
) *RevokeInvite {
	return &RevokeInvite{
		inviteRepo: inviteRepo,
		tenantRepo: tenantRepo,
	}
}

// RevokeInviteRequest represents the request to revoke an invite
type RevokeInviteRequest struct {
	InviteID uuid.UUID
	TenantID uuid.UUID
}

// RevokeInviteResponse represents the response from revoking an invite
type RevokeInviteResponse struct {
	Invite *model.Invite
}

// Execute executes the use case
func (uc *RevokeInvite) Execute(ctx context.Context, req *RevokeInviteRequest) (*RevokeInviteResponse, error) {
	// Verify tenant exists
	_, err := uc.tenantRepo.FindByID(ctx, req.TenantID)
	if err != nil {
		return nil, domain.ErrTenantNotFound
	}

	// Find the invite
	invite, err := uc.inviteRepo.FindByID(ctx, req.InviteID)
	if err != nil {
		return nil, domain.ErrInviteNotFound
	}

	// Verify invite belongs to the tenant
	if invite.TenantID() != req.TenantID {
		return nil, domain.ErrInviteNotFound
	}

	// Check if invite is already revoked
	if invite.IsRevoked() {
		return &RevokeInviteResponse{
			Invite: invite,
		}, nil
	}

	// Check if invite is already accepted
	if invite.IsAccepted() {
		return nil, domain.ErrInviteAlreadyAccepted
	}

	// Revoke the invite
	invite.Revoke()

	// Update the invite in the repository
	if err := uc.inviteRepo.Update(ctx, invite); err != nil {
		return nil, err
	}

	return &RevokeInviteResponse{
		Invite: invite,
	}, nil
}
