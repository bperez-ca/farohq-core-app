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

// TenantMemberRepository implements the outbound.TenantMemberRepository interface
type TenantMemberRepository struct {
	db *pgxpool.Pool
}

// NewTenantMemberRepository creates a new PostgreSQL tenant member repository
func NewTenantMemberRepository(db *pgxpool.Pool) outbound.TenantMemberRepository {
	return &TenantMemberRepository{
		db: db,
	}
}

// FindByID finds a tenant member by ID
func (r *TenantMemberRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.TenantMember, error) {
	query := `
		SELECT id, tenant_id, user_id, role, client_id, created_at, updated_at, deleted_at
		FROM tenant_members
		WHERE id = $1 AND deleted_at IS NULL
	`

	var (
		dbID      uuid.UUID
		tenantID  uuid.UUID
		userID    uuid.UUID
		role      string
		clientID  *uuid.UUID
		createdAt time.Time
		updatedAt time.Time
		deletedAt *time.Time
	)

	err := r.db.QueryRow(ctx, query, id).Scan(
		&dbID,
		&tenantID,
		&userID,
		&role,
		&clientID,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrMemberNotFound
		}
		return nil, err
	}

	return r.mapToDomainMember(dbID, tenantID, userID, role, clientID, createdAt, updatedAt, deletedAt), nil
}

// FindByTenantID finds all members for a tenant
func (r *TenantMemberRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*model.TenantMember, error) {
	query := `
		SELECT id, tenant_id, user_id, role, client_id, created_at, updated_at, deleted_at
		FROM tenant_members
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*model.TenantMember
	for rows.Next() {
		var (
			id         uuid.UUID
			dbTenantID uuid.UUID
			userID     uuid.UUID
			role       string
			clientID   *uuid.UUID
			createdAt  time.Time
			updatedAt  time.Time
			deletedAt  *time.Time
		)

		if err := rows.Scan(&id, &dbTenantID, &userID, &role, &clientID, &createdAt, &updatedAt, &deletedAt); err != nil {
			return nil, err
		}

		members = append(members, r.mapToDomainMember(id, dbTenantID, userID, role, clientID, createdAt, updatedAt, deletedAt))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

// FindByTenantAndUserID finds a member by tenant and user ID
func (r *TenantMemberRepository) FindByTenantAndUserID(ctx context.Context, tenantID, userID uuid.UUID) (*model.TenantMember, error) {
	query := `
		SELECT id, tenant_id, user_id, role, client_id, created_at, updated_at, deleted_at
		FROM tenant_members
		WHERE tenant_id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	var (
		id         uuid.UUID
		dbTenantID uuid.UUID
		dbUserID   uuid.UUID
		role       string
		clientID   *uuid.UUID
		createdAt  time.Time
		updatedAt  time.Time
		deletedAt  *time.Time
	)

	err := r.db.QueryRow(ctx, query, tenantID, userID).Scan(
		&id,
		&dbTenantID,
		&dbUserID,
		&role,
		&clientID,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrMemberNotFound
		}
		return nil, err
	}

	return r.mapToDomainMember(id, dbTenantID, dbUserID, role, clientID, createdAt, updatedAt, deletedAt), nil
}

// Save saves a new tenant member
func (r *TenantMemberRepository) Save(ctx context.Context, member *model.TenantMember) error {
	query := `
		INSERT INTO tenant_members (id, tenant_id, user_id, role, client_id, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (tenant_id, user_id) 
		DO UPDATE SET
			role = EXCLUDED.role,
			client_id = EXCLUDED.client_id,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
	`

	_, err := r.db.Exec(ctx, query,
		member.ID(),
		member.TenantID(),
		member.UserID(),
		string(member.Role()),
		member.ClientID(),
		member.CreatedAt(),
		member.UpdatedAt(),
		member.DeletedAt(),
	)

	return err
}

// Update updates an existing tenant member
func (r *TenantMemberRepository) Update(ctx context.Context, member *model.TenantMember) error {
	query := `
		UPDATE tenant_members SET
			role = $2,
			client_id = $3,
			updated_at = $4,
			deleted_at = $5
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query,
		member.ID(),
		string(member.Role()),
		member.ClientID(),
		member.UpdatedAt(),
		member.DeletedAt(),
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrMemberNotFound
	}

	return nil
}

// Delete deletes a tenant member by ID (soft delete)
func (r *TenantMemberRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE tenant_members SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrMemberNotFound
	}

	return nil
}

// DeleteByTenantAndUserID deletes a tenant member by tenant and user ID (soft delete)
func (r *TenantMemberRepository) DeleteByTenantAndUserID(ctx context.Context, tenantID, userID uuid.UUID) error {
	query := `UPDATE tenant_members SET deleted_at = NOW() WHERE tenant_id = $1 AND user_id = $2 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, tenantID, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrMemberNotFound
	}

	return nil
}

// CountByTenantID counts active members for a tenant
func (r *TenantMemberRepository) CountByTenantID(ctx context.Context, tenantID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM tenant_members
		WHERE tenant_id = $1 AND deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRow(ctx, query, tenantID).Scan(&count)
	return count, err
}

// mapToDomainMember maps database row to domain tenant member
func (r *TenantMemberRepository) mapToDomainMember(id, tenantID, userID uuid.UUID, role string, clientID *uuid.UUID, createdAt, updatedAt time.Time, deletedAt *time.Time) *model.TenantMember {
	memberRole := model.Role(role)
	return model.NewTenantMemberWithID(id, tenantID, userID, memberRole, clientID, createdAt, updatedAt, deletedAt)
}

