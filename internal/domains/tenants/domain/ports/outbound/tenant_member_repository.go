package outbound

import (
	"context"

	"farohq-core-app/internal/domains/tenants/domain/model"

	"github.com/google/uuid"
)

// TenantMemberRepository defines the interface for tenant member data access
type TenantMemberRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.TenantMember, error)
	FindByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*model.TenantMember, error)
	FindByTenantAndUserID(ctx context.Context, tenantID, userID uuid.UUID) (*model.TenantMember, error)
	Save(ctx context.Context, member *model.TenantMember) error
	Update(ctx context.Context, member *model.TenantMember) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByTenantAndUserID(ctx context.Context, tenantID, userID uuid.UUID) error
	CountByTenantID(ctx context.Context, tenantID uuid.UUID) (int, error)
}

