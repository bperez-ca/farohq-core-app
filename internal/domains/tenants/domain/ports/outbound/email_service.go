package outbound

import (
	"context"

	"farohq-core-app/internal/domains/tenants/domain/model"
)

// InviteEmailContext contains all context needed for sending invitation emails
type InviteEmailContext struct {
	// Invite information
	Invite    *model.Invite
	AcceptURL string

	// Agency/Tenant information
	AgencyName string
	Tier       *model.Tier

	// Branding information (optional - may be nil if no branding configured)
	LogoURL        string
	PrimaryColor   string
	SecondaryColor string
	HidePoweredBy  bool

	// User information (optional - may be empty if not available)
	InviteeFirstName string // Extracted from email if not available
	InviterName      string // Name of person who sent invite
	InviterEmail     string // Email of person who sent invite
}

// EmailService defines the interface for sending emails
type EmailService interface {
	// SendInviteEmail sends an invitation email to the invitee with branding support
	// ctx: request context
	// emailCtx: context containing invite, branding, and user information
	SendInviteEmail(ctx context.Context, emailCtx *InviteEmailContext) error
}
