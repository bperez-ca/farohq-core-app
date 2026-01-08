package usecases

import (
	"context"

	"farohq-core-app/internal/domains/tenants/domain/model"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"

	"github.com/google/uuid"
)

// ListClientMembers handles the use case of listing members for a client
type ListClientMembers struct {
	clientMemberRepo outbound.ClientMemberRepository
}

// NewListClientMembers creates a new ListClientMembers use case
func NewListClientMembers(clientMemberRepo outbound.ClientMemberRepository) *ListClientMembers {
	return &ListClientMembers{
		clientMemberRepo: clientMemberRepo,
	}
}

// ListClientMembersRequest represents the request to list client members
type ListClientMembersRequest struct {
	ClientID   uuid.UUID
	LocationID *uuid.UUID // optional filter
}

// ListClientMembersResponse represents the response from listing client members
type ListClientMembersResponse struct {
	Members []*model.ClientMember
}

// Execute executes the use case
func (uc *ListClientMembers) Execute(ctx context.Context, req *ListClientMembersRequest) (*ListClientMembersResponse, error) {
	members, err := uc.clientMemberRepo.ListByClient(ctx, req.ClientID, req.LocationID)
	if err != nil {
		return nil, err
	}

	return &ListClientMembersResponse{
		Members: members,
	}, nil
}

