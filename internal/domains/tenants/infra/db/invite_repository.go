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

// InviteRepository implements the outbound.InviteRepository interface
type InviteRepository struct {
	db *pgxpool.Pool
}

// NewInviteRepository creates a new PostgreSQL invite repository
func NewInviteRepository(db *pgxpool.Pool) outbound.InviteRepository {
	return &InviteRepository{
		db: db,
	}
}

// FindByID finds an invite by ID
func (r *InviteRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Invite, error) {
	query := `
		SELECT id, tenant_id, email, role, token, expires_at, accepted_at, created_at, created_by
		FROM tenant_invites
		WHERE id = $1
	`

	var (
		dbID       uuid.UUID
		tenantID   uuid.UUID
		email      string
		role       string
		token      string
		expiresAt  time.Time
		acceptedAt *time.Time
		createdAt  time.Time
		createdBy  uuid.UUID
	)

	err := r.db.QueryRow(ctx, query, id).Scan(
		&dbID,
		&tenantID,
		&email,
		&role,
		&token,
		&expiresAt,
		&acceptedAt,
		&createdAt,
		&createdBy,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrInviteNotFound
		}
		return nil, err
	}

	return r.mapToDomainInvite(dbID, tenantID, email, role, token, expiresAt, acceptedAt, createdAt, createdBy), nil
}

// FindByToken finds an invite by token
func (r *InviteRepository) FindByToken(ctx context.Context, token string) (*model.Invite, error) {
	query := `
		SELECT id, tenant_id, email, role, token, expires_at, accepted_at, created_at, created_by
		FROM tenant_invites
		WHERE token = $1
	`

	var (
		id         uuid.UUID
		tenantID   uuid.UUID
		email      string
		role       string
		dbToken    string
		expiresAt  time.Time
		acceptedAt *time.Time
		createdAt  time.Time
		createdBy  uuid.UUID
	)

	err := r.db.QueryRow(ctx, query, token).Scan(
		&id,
		&tenantID,
		&email,
		&role,
		&dbToken,
		&expiresAt,
		&acceptedAt,
		&createdAt,
		&createdBy,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrInviteNotFound
		}
		return nil, err
	}

	return r.mapToDomainInvite(id, tenantID, email, role, dbToken, expiresAt, acceptedAt, createdAt, createdBy), nil
}

// FindByTenantID finds all invites for a tenant
func (r *InviteRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*model.Invite, error) {
	query := `
		SELECT id, tenant_id, email, role, token, expires_at, accepted_at, created_at, created_by
		FROM tenant_invites
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []*model.Invite
	for rows.Next() {
		var (
			id         uuid.UUID
			dbTenantID uuid.UUID
			email      string
			role       string
			token      string
			expiresAt  time.Time
			acceptedAt *time.Time
			createdAt  time.Time
			createdBy  uuid.UUID
		)

		if err := rows.Scan(&id, &dbTenantID, &email, &role, &token, &expiresAt, &acceptedAt, &createdAt, &createdBy); err != nil {
			return nil, err
		}

		invites = append(invites, r.mapToDomainInvite(id, dbTenantID, email, role, token, expiresAt, acceptedAt, createdAt, createdBy))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return invites, nil
}

// FindByEmail finds an invite by email and tenant ID
func (r *InviteRepository) FindByEmail(ctx context.Context, email string, tenantID uuid.UUID) (*model.Invite, error) {
	query := `
		SELECT id, tenant_id, email, role, token, expires_at, accepted_at, created_at, created_by
		FROM tenant_invites
		WHERE tenant_id = $1 AND email = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	var (
		id         uuid.UUID
		dbTenantID uuid.UUID
		dbEmail    string
		role       string
		token      string
		expiresAt  time.Time
		acceptedAt *time.Time
		createdAt  time.Time
		createdBy  uuid.UUID
	)

	err := r.db.QueryRow(ctx, query, tenantID, email).Scan(
		&id,
		&dbTenantID,
		&dbEmail,
		&role,
		&token,
		&expiresAt,
		&acceptedAt,
		&createdAt,
		&createdBy,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrInviteNotFound
		}
		return nil, err
	}

	return r.mapToDomainInvite(id, dbTenantID, dbEmail, role, token, expiresAt, acceptedAt, createdAt, createdBy), nil
}

// Save saves a new invite
func (r *InviteRepository) Save(ctx context.Context, invite *model.Invite) error {
	query := `
		INSERT INTO tenant_invites (id, tenant_id, email, role, token, expires_at, accepted_at, created_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(ctx, query,
		invite.ID(),
		invite.TenantID(),
		invite.Email(),
		string(invite.Role()),
		invite.Token(),
		invite.ExpiresAt(),
		invite.AcceptedAt(),
		invite.CreatedAt(),
		invite.CreatedBy(),
	)

	return err
}

// Update updates an existing invite
func (r *InviteRepository) Update(ctx context.Context, invite *model.Invite) error {
	query := `
		UPDATE tenant_invites SET
			accepted_at = $2
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query,
		invite.ID(),
		invite.AcceptedAt(),
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrInviteNotFound
	}

	return nil
}

// Delete deletes an invite by ID
func (r *InviteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tenant_invites WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrInviteNotFound
	}

	return nil
}

// mapToDomainInvite maps database row to domain invite
func (r *InviteRepository) mapToDomainInvite(id, tenantID uuid.UUID, email, role, token string, expiresAt time.Time, acceptedAt *time.Time, createdAt time.Time, createdBy uuid.UUID) *model.Invite {
	inviteRole := model.Role(role)
	return model.NewInviteWithID(id, tenantID, email, inviteRole, token, expiresAt, acceptedAt, createdAt, createdBy)
}

