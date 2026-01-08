package usecases

import (
	"context"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/model"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"
	"farohq-core-app/internal/domains/tenants/domain/services"

	"github.com/google/uuid"
)

// AddClientMember handles the use case of adding a member to a client
type AddClientMember struct {
	clientMemberRepo outbound.ClientMemberRepository
	locationRepo     outbound.LocationRepository
	seatValidator    *services.SeatValidator
}

// NewAddClientMember creates a new AddClientMember use case
func NewAddClientMember(
	clientMemberRepo outbound.ClientMemberRepository,
	locationRepo outbound.LocationRepository,
	seatValidator *services.SeatValidator,
) *AddClientMember {
	return &AddClientMember{
		clientMemberRepo: clientMemberRepo,
		locationRepo:     locationRepo,
		seatValidator:    seatValidator,
	}
}

// AddClientMemberRequest represents the request to add a client member
type AddClientMemberRequest struct {
	ClientID   uuid.UUID
	UserID     uuid.UUID
	Role       model.Role
	LocationID *uuid.UUID // nullable - for location-scoped members
}

// AddClientMemberResponse represents the response from adding a client member
type AddClientMemberResponse struct {
	Member *model.ClientMember
}

// Execute executes the use case
func (uc *AddClientMember) Execute(ctx context.Context, req *AddClientMemberRequest) (*AddClientMemberResponse, error) {
	// Validate role
	if req.Role != model.RoleOwner && req.Role != model.RoleAdmin &&
		req.Role != model.RoleStaff && req.Role != model.RoleViewer &&
		req.Role != model.RoleClientViewer {
		return nil, domain.ErrInvalidRole
	}

	// Get location count for seat calculation
	locationCount, err := uc.locationRepo.CountByClient(ctx, req.ClientID)
	if err != nil {
		return nil, err
	}

	// Get current member count
	currentMemberCount, err := uc.clientMemberRepo.CountByClient(ctx, req.ClientID)
	if err != nil {
		return nil, err
	}

	// Validate seat limits (1 base + 1 per location)
	if err := uc.seatValidator.ValidateClientSeats(locationCount, currentMemberCount, 1); err != nil {
		return nil, err
	}

	// Check if member already exists
	existing, err := uc.clientMemberRepo.FindByClientAndUser(ctx, req.ClientID, req.UserID, req.LocationID)
	if err == nil && existing != nil {
		return nil, domain.ErrMemberAlreadyExists
	}

	// Create new client member
	member := model.NewClientMember(req.ClientID, req.UserID, req.Role, req.LocationID)

	// Save member
	if err := uc.clientMemberRepo.Save(ctx, member); err != nil {
		return nil, err
	}

	return &AddClientMemberResponse{
		Member: member,
	}, nil
}

