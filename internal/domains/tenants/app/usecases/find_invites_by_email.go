package usecases

import (
	"context"
	"strings"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/model"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"
)

// FindInvitesByEmail handles the use case of finding pending invites by email
type FindInvitesByEmail struct {
	inviteRepo outbound.InviteRepository
}

// NewFindInvitesByEmail creates a new FindInvitesByEmail use case
func NewFindInvitesByEmail(inviteRepo outbound.InviteRepository) *FindInvitesByEmail {
	return &FindInvitesByEmail{
		inviteRepo: inviteRepo,
	}
}

// FindInvitesByEmailRequest represents the request to find invites by email
type FindInvitesByEmailRequest struct {
	Email string
}

// FindInvitesByEmailResponse represents the response from finding invites by email
type FindInvitesByEmailResponse struct {
	Invites []*model.Invite
}

// Execute executes the use case
func (uc *FindInvitesByEmail) Execute(ctx context.Context, req *FindInvitesByEmailRequest) (*FindInvitesByEmailResponse, error) {
	// Validate input
	if strings.TrimSpace(req.Email) == "" {
		return nil, domain.ErrInvalidEmail
	}

	// Normalize email
	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Find pending invites by email (across all tenants)
	invites, err := uc.inviteRepo.FindPendingInvitesByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return &FindInvitesByEmailResponse{
		Invites: invites,
	}, nil
}
