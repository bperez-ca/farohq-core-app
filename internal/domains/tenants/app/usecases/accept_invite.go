package usecases

import (
	"context"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/model"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"

	"github.com/google/uuid"
)

// AcceptInvite handles the use case of accepting an invitation
type AcceptInvite struct {
	inviteRepo outbound.InviteRepository
	memberRepo outbound.TenantMemberRepository
}

// NewAcceptInvite creates a new AcceptInvite use case
func NewAcceptInvite(inviteRepo outbound.InviteRepository, memberRepo outbound.TenantMemberRepository) *AcceptInvite {
	return &AcceptInvite{
		inviteRepo: inviteRepo,
		memberRepo: memberRepo,
	}
}

// AcceptInviteRequest represents the request to accept an invite
type AcceptInviteRequest struct {
	Token  string
	UserID uuid.UUID
}

// AcceptInviteResponse represents the response from accepting an invite
type AcceptInviteResponse struct {
	Member *model.TenantMember
}

// Execute executes the use case
func (uc *AcceptInvite) Execute(ctx context.Context, req *AcceptInviteRequest) (*AcceptInviteResponse, error) {
	// Find invite by token
	invite, err := uc.inviteRepo.FindByToken(ctx, req.Token)
	if err != nil {
		return nil, domain.ErrInviteNotFound
	}

	// Check if invite is already accepted
	if invite.IsAccepted() {
		return nil, domain.ErrInviteAlreadyAccepted
	}

	// Check if invite has expired
	if invite.IsExpired() {
		return nil, domain.ErrInviteExpired
	}

	// Check if user is already a member
	existingMember, err := uc.memberRepo.FindByTenantAndUserID(ctx, invite.TenantID(), req.UserID)
	if err == nil && existingMember != nil {
		return &AcceptInviteResponse{
			Member: existingMember,
		}, nil
	}

	// Create tenant member
	member := model.NewTenantMember(invite.TenantID(), req.UserID, invite.Role())

	// Save member
	if err := uc.memberRepo.Save(ctx, member); err != nil {
		return nil, err
	}

	// Mark invite as accepted
	invite.Accept()
	if err := uc.inviteRepo.Update(ctx, invite); err != nil {
		return nil, err
	}

	return &AcceptInviteResponse{
		Member: member,
	}, nil
}

