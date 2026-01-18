package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"

	"github.com/rs/zerolog"
)

// PostmarkEmailService implements EmailService using Postmark API
type PostmarkEmailService struct {
	apiToken  string
	fromEmail string
	client    *http.Client
	logger    zerolog.Logger
}

// NewPostmarkEmailService creates a new Postmark email service
func NewPostmarkEmailService(apiToken, fromEmail string, logger zerolog.Logger) outbound.EmailService {
	return &PostmarkEmailService{
		apiToken:  apiToken,
		fromEmail: fromEmail,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// SendInviteEmail sends an invitation email via Postmark with branding support
func (s *PostmarkEmailService) SendInviteEmail(ctx context.Context, emailCtx *outbound.InviteEmailContext) error {
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

	// Build HTML and text email bodies
	htmlBody, err := BuildInviteEmailHTML(data)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to build HTML email")
		return fmt.Errorf("failed to build HTML email: %w", err)
	}

	textBody := BuildInviteEmailText(data)

	// Postmark API request payload
	payload := map[string]interface{}{
		"From":          fmt.Sprintf("%s <%s>", fromName, s.fromEmail),
		"To":            emailCtx.Invite.Email(),
		"Subject":       subject,
		"HtmlBody":      htmlBody,
		"TextBody":      textBody,
		"MessageStream": "outbound",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to marshal Postmark email payload")
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.postmarkapp.com/email", bytes.NewBuffer(jsonData))
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create Postmark email request")
		return fmt.Errorf("failed to create email request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Postmark-Server-Token", s.apiToken)

	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error().Err(err).Str("to", emailCtx.Invite.Email()).Msg("Failed to send email via Postmark")
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResp)
		s.logger.Error().
			Int("status_code", resp.StatusCode).
			Interface("error", errorResp).
			Str("to", emailCtx.Invite.Email()).
			Msg("Postmark API returned error")
		return fmt.Errorf("postmark API error: status %d", resp.StatusCode)
	}

	s.logger.Info().
		Str("to", emailCtx.Invite.Email()).
		Str("invite_id", emailCtx.Invite.ID().String()).
		Str("agency", emailCtx.AgencyName).
		Bool("white_label", emailCtx.HidePoweredBy).
		Msg("Invite email sent successfully via Postmark")

	return nil
}
