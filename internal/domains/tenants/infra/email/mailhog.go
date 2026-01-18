package email

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"

	"github.com/rs/zerolog"
)

// MailhogEmailService implements EmailService using Mailhog SMTP server for local development
type MailhogEmailService struct {
	smtpHost string
	smtpPort string
	logger   zerolog.Logger
}

// NewMailhogEmailService creates a new Mailhog email service
// Note: Mailhog uses SMTP port 1025 (not the HTTP API port 8025)
func NewMailhogEmailService(host, port string, logger zerolog.Logger) outbound.EmailService {
	// Mailhog SMTP is on port 1025, but we accept port from config for flexibility
	// If port is 8025 (web UI), use 1025 for SMTP
	smtpPort := port
	if port == "8025" {
		smtpPort = "1025"
	}
	return &MailhogEmailService{
		smtpHost: host,
		smtpPort: smtpPort,
		logger:   logger,
	}
}

// SendInviteEmail sends an invitation email via Mailhog SMTP with branding support
func (s *MailhogEmailService) SendInviteEmail(ctx context.Context, emailCtx *outbound.InviteEmailContext) error {
	// Build email data from context
	tierStr := "starter"
	if emailCtx.Tier != nil {
		tierStr = string(*emailCtx.Tier)
	}

	data := InviteEmailData{
		InviteEmail:      emailCtx.Invite.Email(),
		RoleName:         string(emailCtx.Invite.Role()),
		InviteURL:        emailCtx.AcceptURL,
		ExpiresAt:        emailCtx.Invite.ExpiresAt(),
		AgencyName:       emailCtx.AgencyName,
		Tier:             tierStr,
		LogoURL:          emailCtx.LogoURL,
		PrimaryColor:     emailCtx.PrimaryColor,
		SecondaryColor:   emailCtx.SecondaryColor,
		HidePoweredBy:    emailCtx.HidePoweredBy,
		InviteeFirstName: emailCtx.InviteeFirstName,
		InviterName:      emailCtx.InviterName,
		InviterEmail:     emailCtx.InviterEmail,
	}

	// Build subject and from name
	subject := BuildInviteEmailSubject(data)
	fromName := BuildInviteEmailFromName(data)

	// Build HTML email body
	htmlBody, err := BuildInviteEmailHTML(data)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to build HTML email")
		return fmt.Errorf("failed to build HTML email: %w", err)
	}

	// Build email message (RFC 5322 format)
	from := fmt.Sprintf("%s <noreply@localhost>", fromName)
	to := emailCtx.Invite.Email()

	// Email headers
	headers := []string{
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		fmt.Sprintf("X-Invite-ID: %s", emailCtx.Invite.ID().String()),
		"",
		htmlBody,
	}

	message := []byte(strings.Join(headers, "\r\n"))

	// Send email via SMTP
	// Mailhog doesn't require authentication
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	auth := smtp.PlainAuth("", "", "", s.smtpHost)

	err = smtp.SendMail(addr, auth, "noreply@localhost", []string{to}, message)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("to", emailCtx.Invite.Email()).
			Str("smtp_addr", addr).
			Msg("Failed to send email via Mailhog SMTP")
		return fmt.Errorf("failed to send email via SMTP: %w", err)
	}

	s.logger.Info().
		Str("to", emailCtx.Invite.Email()).
		Str("invite_id", emailCtx.Invite.ID().String()).
		Str("smtp_addr", addr).
		Str("mailhog_ui", fmt.Sprintf("http://%s:8025", s.smtpHost)).
		Str("agency", emailCtx.AgencyName).
		Bool("white_label", emailCtx.HidePoweredBy).
		Msg("Invite email sent successfully via Mailhog SMTP")

	return nil
}
