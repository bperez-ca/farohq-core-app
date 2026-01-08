package outbound

import (
	"context"

	"farohq-core-app/internal/domains/tenants/domain/model"

	"github.com/google/uuid"
)

// TenantRepository defines the interface for tenant data access
type TenantRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.Tenant, error)
	FindBySlug(ctx context.Context, slug string) (*model.Tenant, error)
	Save(ctx context.Context, tenant *model.Tenant) error
	Update(ctx context.Context, tenant *model.Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error
}

