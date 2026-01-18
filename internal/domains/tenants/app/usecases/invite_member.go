package usecases

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/model"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"
	"farohq-core-app/internal/domains/tenants/domain/services"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// BrandRepository interface for fetching branding (to avoid circular dependency)
type BrandRepository interface {
	FindByAgencyID(ctx context.Context, agencyID uuid.UUID) (BrandingInfo, error)
}

// BrandingInfo interface for accessing branding fields
type BrandingInfo interface {
	LogoURL() string
	PrimaryColor() string
	SecondaryColor() string
	HidePoweredBy() bool
}

// UserRepository interface for fetching user information
type UserRepository interface {
	FindByID(ctx context.Context, userID uuid.UUID) (UserInfo, error)
}

// UserInfo interface for accessing user fields
type UserInfo interface {
	FullName() string
	FirstName() string
	LastName() string
	Email() string
}

// InviteMember handles the use case of inviting a member to join a tenant
type InviteMember struct {
	inviteRepo    outbound.InviteRepository
	memberRepo    outbound.TenantMemberRepository
	tenantRepo    outbound.TenantRepository
	brandRepo     BrandRepository
	userRepo      UserRepository
	emailService  outbound.EmailService
	seatValidator *services.SeatValidator
	tokenExpiry   time.Duration
	webURL        string
}

// NewInviteMember creates a new InviteMember use case
func NewInviteMember(
	inviteRepo outbound.InviteRepository,
	memberRepo outbound.TenantMemberRepository,
	tenantRepo outbound.TenantRepository,
	brandRepo BrandRepository,
	userRepo UserRepository,
	emailService outbound.EmailService,
	seatValidator *services.SeatValidator,
	tokenExpiry time.Duration,
	webURL string,
) *InviteMember {
	return &InviteMember{
		inviteRepo:    inviteRepo,
		memberRepo:    memberRepo,
		tenantRepo:    tenantRepo,
		brandRepo:     brandRepo,
		userRepo:      userRepo,
		emailService:  emailService,
		seatValidator: seatValidator,
		tokenExpiry:   tokenExpiry,
		webURL:        webURL,
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
	if err == nil && existingInvite != nil && !existingInvite.IsAccepted() && !existingInvite.IsExpired() && !existingInvite.IsRevoked() {
		return nil, domain.ErrInviteAlreadyAccepted // Or a more specific error for pending invite
	}

	// Generate secure token
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	// Use tenant-specific expiry duration (defaults to 24 hours if not configured)
	// Max is enforced at the tenant level (72 hours = 3 days)
	expiryDuration := tenant.InviteExpiryDuration()

	// Create invite with tenant-specific expiration
	invite := model.NewInvite(req.TenantID, email, req.Role, token, req.CreatedBy, expiryDuration)

	// Save invite
	if err := uc.inviteRepo.Save(ctx, invite); err != nil {
		return nil, err
	}

	// Send invite email asynchronously (non-blocking)
	go func() {
		acceptURL := fmt.Sprintf("%s/invites/accept/%s", uc.webURL, invite.Token())

		// Build email context with branding and user information
		emailCtx := uc.buildEmailContext(context.Background(), tenant, invite, acceptURL, req.CreatedBy)

		if err := uc.emailService.SendInviteEmail(context.Background(), emailCtx); err != nil {
			// Log error but don't fail the invite creation
			log.Error().
				Err(err).
				Str("invite_id", invite.ID().String()).
				Str("email", invite.Email()).
				Str("accept_url", acceptURL).
				Msg("Failed to send invite email asynchronously")
		} else {
			log.Info().
				Str("invite_id", invite.ID().String()).
				Str("email", invite.Email()).
				Msg("Invite email sent successfully")
		}
	}()

	return &InviteMemberResponse{
		Invite: invite,
	}, nil
}

// buildEmailContext builds the email context with branding and user information
func (uc *InviteMember) buildEmailContext(ctx context.Context, tenant *model.Tenant, invite *model.Invite, acceptURL string, createdBy uuid.UUID) *outbound.InviteEmailContext {
	emailCtx := &outbound.InviteEmailContext{
		Invite:     invite,
		AcceptURL:  acceptURL,
		AgencyName: tenant.Name(),
		Tier:       tenant.Tier(),
	}

	// Fetch branding information (optional - may not exist)
	if uc.brandRepo != nil {
		branding, err := uc.brandRepo.FindByAgencyID(ctx, tenant.ID())
		if err == nil && branding != nil {
			emailCtx.LogoURL = branding.LogoURL()
			emailCtx.PrimaryColor = branding.PrimaryColor()
			emailCtx.SecondaryColor = branding.SecondaryColor()
			hidePoweredBy := branding.HidePoweredBy()

			// Apply tier-based rules: only Growth+ can hide powered by
			if tenant.Tier() != nil && !model.TierCanHidePoweredBy(tenant.Tier()) {
				hidePoweredBy = false
			}
			emailCtx.HidePoweredBy = hidePoweredBy
		}
	}

	// Fetch inviter user information (optional)
	if uc.userRepo != nil {
		user, err := uc.userRepo.FindByID(ctx, createdBy)
		if err == nil && user != nil {
			inviterName := user.FullName()
			if inviterName == "" {
				// Fallback to first name + last name
				firstName := user.FirstName()
				lastName := user.LastName()
				if firstName != "" && lastName != "" {
					inviterName = firstName + " " + lastName
				} else if firstName != "" {
					inviterName = firstName
				}
			}
			emailCtx.InviterName = inviterName
			emailCtx.InviterEmail = user.Email()
		}
	}

	// Extract invitee first name from email as fallback
	emailCtx.InviteeFirstName = uc.extractFirstNameFromEmail(invite.Email())

	return emailCtx
}

// extractFirstNameFromEmail extracts first name from email address
func (uc *InviteMember) extractFirstNameFromEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) > 0 {
		localPart := parts[0]
		nameParts := strings.Split(localPart, ".")
		if len(nameParts) > 0 {
			return strings.Title(nameParts[0])
		}
		return strings.Title(localPart)
	}
	return ""
}

// generateToken generates a secure random token
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
