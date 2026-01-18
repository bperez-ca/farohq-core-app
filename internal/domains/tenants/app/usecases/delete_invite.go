package usecases

import (
	"context"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"

	"github.com/google/uuid"
)

// DeleteInvite handles the use case of deleting an invitation (distinct from revoke)
// Delete permanently removes the invitation from the database
// Revoke marks the invitation as revoked but keeps it in the database
type DeleteInvite struct {
	inviteRepo outbound.InviteRepository
	tenantRepo outbound.TenantRepository
}

// NewDeleteInvite creates a new DeleteInvite use case
func NewDeleteInvite(
	inviteRepo outbound.InviteRepository,
	tenantRepo outbound.TenantRepository,
) *DeleteInvite {
	return &DeleteInvite{
		inviteRepo: inviteRepo,
		tenantRepo: tenantRepo,
	}
}

// DeleteInviteRequest represents the request to delete an invite
type DeleteInviteRequest struct {
	InviteID uuid.UUID
	TenantID uuid.UUID
}

// DeleteInviteResponse represents the response from deleting an invite
type DeleteInviteResponse struct {
	Success bool
}

// Execute executes the use case
func (uc *DeleteInvite) Execute(ctx context.Context, req *DeleteInviteRequest) (*DeleteInviteResponse, error) {
	// Verify tenant exists
	_, err := uc.tenantRepo.FindByID(ctx, req.TenantID)
	if err != nil {
		return nil, domain.ErrTenantNotFound
	}

	// Find the invite to verify it exists and belongs to the tenant
	invite, err := uc.inviteRepo.FindByID(ctx, req.InviteID)
	if err != nil {
		return nil, domain.ErrInviteNotFound
	}

	// Verify invite belongs to the tenant
	if invite.TenantID() != req.TenantID {
		return nil, domain.ErrInviteNotFound
	}

	// Delete the invite permanently
	if err := uc.inviteRepo.Delete(ctx, req.InviteID); err != nil {
		return nil, err
	}

	return &DeleteInviteResponse{
		Success: true,
	}, nil
}
