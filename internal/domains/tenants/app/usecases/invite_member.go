package usecases

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/model"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"
	"farohq-core-app/internal/domains/tenants/domain/services"

	"github.com/google/uuid"
)

// InviteMember handles the use case of inviting a member to join a tenant
type InviteMember struct {
	inviteRepo    outbound.InviteRepository
	memberRepo    outbound.TenantMemberRepository
	tenantRepo    outbound.TenantRepository
	seatValidator *services.SeatValidator
	tokenExpiry   time.Duration
}

// NewInviteMember creates a new InviteMember use case
func NewInviteMember(
	inviteRepo outbound.InviteRepository,
	memberRepo outbound.TenantMemberRepository,
	tenantRepo outbound.TenantRepository,
	seatValidator *services.SeatValidator,
	tokenExpiry time.Duration,
) *InviteMember {
	return &InviteMember{
		inviteRepo:    inviteRepo,
		memberRepo:    memberRepo,
		tenantRepo:    tenantRepo,
		seatValidator: seatValidator,
		tokenExpiry:   tokenExpiry,
	}
}

// InviteMemberRequest represents the request to invite a member
type InviteMemberRequest struct {
	TenantID  uuid.UUID
	Email     string
	Role      model.Role
	CreatedBy uuid.UUID
}

// InviteMemberResponse represents the response from inviting a member
type InviteMemberResponse struct {
	Invite *model.Invite
}

// Execute executes the use case
func (uc *InviteMember) Execute(ctx context.Context, req *InviteMemberRequest) (*InviteMemberResponse, error) {
	// Validate input
	if strings.TrimSpace(req.Email) == "" {
		return nil, domain.ErrInvalidEmail
	}

	// Normalize email
	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Validate role
	if req.Role != model.RoleOwner && req.Role != model.RoleAdmin &&
		req.Role != model.RoleStaff && req.Role != model.RoleViewer &&
		req.Role != model.RoleClientViewer {
		return nil, domain.ErrInvalidRole
	}

	// Verify tenant exists
	tenant, err := uc.tenantRepo.FindByID(ctx, req.TenantID)
	if err != nil {
		return nil, domain.ErrTenantNotFound
	}

	// Validate agency seat limits if set
	seatLimit := tenant.AgencySeatLimit()
	if seatLimit > 0 {
		// Count current members (excluding soft-deleted)
		members, err := uc.memberRepo.FindByTenantID(ctx, req.TenantID)
		if err != nil {
			return nil, err
		}
		currentCount := len(members)

		// Validate seat limit
		if err := uc.seatValidator.ValidateAgencySeats(seatLimit, currentCount, 1); err != nil {
			return nil, err
		}
	}

	// Check if there's already a pending invite for this email
	existingInvite, err := uc.inviteRepo.FindByEmail(ctx, email, req.TenantID)
	if err == nil && existingInvite != nil && !existingInvite.IsAccepted() && !existingInvite.IsExpired() {
		return nil, domain.ErrInviteAlreadyAccepted // Or a more specific error for pending invite
	}

	// Generate secure token
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	// Create invite
	invite := model.NewInvite(req.TenantID, email, req.Role, token, req.CreatedBy, uc.tokenExpiry)

	// Save invite
	if err := uc.inviteRepo.Save(ctx, invite); err != nil {
		return nil, err
	}

	return &InviteMemberResponse{
		Invite: invite,
	}, nil
}

// generateToken generates a secure random token
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

