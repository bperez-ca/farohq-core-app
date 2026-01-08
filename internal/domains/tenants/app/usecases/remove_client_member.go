package usecases

import (
	"context"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"

	"github.com/google/uuid"
)

// RemoveClientMember handles the use case of removing a member from a client (soft delete)
type RemoveClientMember struct {
	clientMemberRepo outbound.ClientMemberRepository
}

// NewRemoveClientMember creates a new RemoveClientMember use case
func NewRemoveClientMember(clientMemberRepo outbound.ClientMemberRepository) *RemoveClientMember {
	return &RemoveClientMember{
		clientMemberRepo: clientMemberRepo,
	}
}

// RemoveClientMemberRequest represents the request to remove a client member
type RemoveClientMemberRequest struct {
	MemberID uuid.UUID
}

// RemoveClientMemberResponse represents the response from removing a client member
type RemoveClientMemberResponse struct {
	Success bool
}

// Execute executes the use case
func (uc *RemoveClientMember) Execute(ctx context.Context, req *RemoveClientMemberRequest) (*RemoveClientMemberResponse, error) {
	// Get existing member
	member, err := uc.clientMemberRepo.FindByID(ctx, req.MemberID)
	if err != nil {
		return nil, domain.ErrClientMemberNotFound
	}

	// Soft delete
	member.Delete()

	// Save updated member
	if err := uc.clientMemberRepo.Save(ctx, member); err != nil {
		return nil, err
	}

	return &RemoveClientMemberResponse{
		Success: true,
	}, nil
}

