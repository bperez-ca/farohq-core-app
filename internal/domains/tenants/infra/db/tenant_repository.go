package db

import (
	"context"
	"time"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/model"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TenantRepository implements the outbound.TenantRepository interface
type TenantRepository struct {
	db *pgxpool.Pool
}

// NewTenantRepository creates a new PostgreSQL tenant repository
func NewTenantRepository(db *pgxpool.Pool) outbound.TenantRepository {
	return &TenantRepository{
		db: db,
	}
}

// FindByID finds a tenant by ID (from agencies table)
func (r *TenantRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Tenant, error) {
	query := `
		SELECT id, name, slug, status, tier, agency_seat_limit, invite_expiry_hours, created_at, updated_at, deleted_at
		FROM agencies
		WHERE id = $1 AND deleted_at IS NULL
	`

	var (
		dbID             uuid.UUID
		name             string
		slug             string
		status           string
		tier             *string
		agencySeatLimit  int
		inviteExpiryHours *int
		createdAt        time.Time
		updatedAt        time.Time
		deletedAt        *time.Time
	)

	err := r.db.QueryRow(ctx, query, id).Scan(
		&dbID,
		&name,
		&slug,
		&status,
		&tier,
		&agencySeatLimit,
		&inviteExpiryHours,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrTenantNotFound
		}
		return nil, err
	}

	return r.mapToDomainTenant(dbID, name, slug, status, tier, agencySeatLimit, inviteExpiryHours, createdAt, updatedAt, deletedAt), nil
}

// FindBySlug finds a tenant by slug (from agencies table)
func (r *TenantRepository) FindBySlug(ctx context.Context, slug string) (*model.Tenant, error) {
	query := `
		SELECT id, name, slug, status, tier, agency_seat_limit, invite_expiry_hours, created_at, updated_at, deleted_at
		FROM agencies
		WHERE slug = $1 AND deleted_at IS NULL
	`

	var (
		id               uuid.UUID
		name             string
		dbSlug           string
		status           string
		tier             *string
		agencySeatLimit  int
		inviteExpiryHours *int
		createdAt        time.Time
		updatedAt        time.Time
		deletedAt        *time.Time
	)

	err := r.db.QueryRow(ctx, query, slug).Scan(
		&id,
		&name,
		&dbSlug,
		&status,
		&tier,
		&agencySeatLimit,
		&inviteExpiryHours,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrTenantNotFound
		}
		return nil, err
	}

	return r.mapToDomainTenant(id, name, dbSlug, status, tier, agencySeatLimit, inviteExpiryHours, createdAt, updatedAt, deletedAt), nil
}

// Save saves a new tenant (inserts into agencies table)
func (r *TenantRepository) Save(ctx context.Context, tenant *model.Tenant) error {
	var tierStr *string
	if tenant.Tier() != nil {
		t := tenant.Tier().String()
		tierStr = &t
	}

	query := `
		INSERT INTO agencies (id, name, slug, status, tier, agency_seat_limit, invite_expiry_hours, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Exec(ctx, query,
		tenant.ID(),
		tenant.Name(),
		tenant.Slug(),
		string(tenant.Status()),
		tierStr,
		tenant.AgencySeatLimit(),
		tenant.InviteExpiryHours(),
		tenant.CreatedAt(),
		tenant.UpdatedAt(),
		tenant.DeletedAt(),
	)

	return err
}

// Update updates an existing tenant (updates agencies table)
func (r *TenantRepository) Update(ctx context.Context, tenant *model.Tenant) error {
	var tierStr *string
	if tenant.Tier() != nil {
		t := tenant.Tier().String()
		tierStr = &t
	}

	query := `
		UPDATE agencies SET
			name = $2,
			slug = $3,
			status = $4,
			tier = $5,
			agency_seat_limit = $6,
			invite_expiry_hours = $7,
			updated_at = $8,
			deleted_at = $9
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query,
		tenant.ID(),
		tenant.Name(),
		tenant.Slug(),
		string(tenant.Status()),
		tierStr,
		tenant.AgencySeatLimit(),
		tenant.InviteExpiryHours(),
		tenant.UpdatedAt(),
		tenant.DeletedAt(),
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrTenantNotFound
	}

	return nil
}

// Delete deletes a tenant (soft delete)
func (r *TenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE agencies SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrTenantNotFound
	}

	return nil
}

// mapToDomainTenant maps database row to domain tenant
func (r *TenantRepository) mapToDomainTenant(id uuid.UUID, name, slug, status string, tier *string, agencySeatLimit int, inviteExpiryHours *int, createdAt, updatedAt time.Time, deletedAt *time.Time) *model.Tenant {
	tenantStatus := model.TenantStatus(status)
	var domainTier *model.Tier
	if tier != nil {
		t := model.Tier(*tier)
		domainTier = &t
	}
	return model.NewTenantWithID(id, name, slug, tenantStatus, domainTier, agencySeatLimit, inviteExpiryHours, createdAt, updatedAt, deletedAt)
}
