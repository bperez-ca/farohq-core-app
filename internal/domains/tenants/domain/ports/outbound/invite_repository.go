package outbound

import (
	"context"

	"farohq-core-app/internal/domains/tenants/domain/model"

	"github.com/google/uuid"
)

// InviteRepository defines the interface for invite data access
type InviteRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.Invite, error)
	FindByToken(ctx context.Context, token string) (*model.Invite, error)
	FindByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*model.Invite, error)
	FindByEmail(ctx context.Context, email string, tenantID uuid.UUID) (*model.Invite, error)
	Save(ctx context.Context, invite *model.Invite) error
	Update(ctx context.Context, invite *model.Invite) error
	Delete(ctx context.Context, id uuid.UUID) error
}

