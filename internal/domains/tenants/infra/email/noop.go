package email

import (
	"context"

	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"

	"github.com/rs/zerolog"
)

// NoopEmailService is a no-op email service that only logs (for testing or fallback)
type NoopEmailService struct {
	logger zerolog.Logger
}

// NewNoopEmailService creates a new no-op email service
func NewNoopEmailService(logger zerolog.Logger) outbound.EmailService {
	return &NoopEmailService{
		logger: logger,
	}
}

// SendInviteEmail logs the email send attempt but doesn't actually send
func (s *NoopEmailService) SendInviteEmail(ctx context.Context, emailCtx *outbound.InviteEmailContext) error {
	s.logger.Info().
		Str("to", emailCtx.Invite.Email()).
		Str("invite_id", emailCtx.Invite.ID().String()).
		Str("accept_url", emailCtx.AcceptURL).
		Str("agency", emailCtx.AgencyName).
		Bool("white_label", emailCtx.HidePoweredBy).
		Msg("No-op email service: would send invite email (email sending disabled)")
	return nil
}
