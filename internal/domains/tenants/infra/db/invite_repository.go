package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/model"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
		SELECT id, tenant_id, email, role, token, expires_at, accepted_at, revoked_at, created_at, created_by
		FROM tenant_invites
		WHERE id = $1 AND deleted_at IS NULL
	`

	var (
		dbID       uuid.UUID
		tenantID   uuid.UUID
		email      string
		role       string
		token      string
		expiresAt  time.Time
		acceptedAt *time.Time
		revokedAt  *time.Time
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
		&revokedAt,
		&createdAt,
		&createdBy,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrInviteNotFound
		}
		return nil, err
	}

	return r.mapToDomainInvite(dbID, tenantID, email, role, token, expiresAt, acceptedAt, revokedAt, createdAt, createdBy), nil
}

// FindByToken finds an invite by token
func (r *InviteRepository) FindByToken(ctx context.Context, token string) (*model.Invite, error) {
	query := `
		SELECT id, tenant_id, email, role, token, expires_at, accepted_at, revoked_at, created_at, created_by
		FROM tenant_invites
		WHERE token = $1 AND deleted_at IS NULL
	`

	var (
		id         uuid.UUID
		tenantID   uuid.UUID
		email      string
		role       string
		dbToken    string
		expiresAt  time.Time
		acceptedAt *time.Time
		revokedAt  *time.Time
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
		&revokedAt,
		&createdAt,
		&createdBy,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrInviteNotFound
		}
		return nil, err
	}

	return r.mapToDomainInvite(id, tenantID, email, role, dbToken, expiresAt, acceptedAt, revokedAt, createdAt, createdBy), nil
}

// FindByTenantID finds all invites for a tenant
func (r *InviteRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*model.Invite, error) {
	query := `
		SELECT id, tenant_id, email, role, token, expires_at, accepted_at, revoked_at, created_at, created_by
		FROM tenant_invites
		WHERE tenant_id = $1 AND deleted_at IS NULL
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
			revokedAt  *time.Time
			createdAt  time.Time
			createdBy  uuid.UUID
		)

		if err := rows.Scan(&id, &dbTenantID, &email, &role, &token, &expiresAt, &acceptedAt, &revokedAt, &createdAt, &createdBy); err != nil {
			return nil, err
		}

		invites = append(invites, r.mapToDomainInvite(id, dbTenantID, email, role, token, expiresAt, acceptedAt, revokedAt, createdAt, createdBy))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return invites, nil
}

// FindByEmail finds an invite by email and tenant ID
func (r *InviteRepository) FindByEmail(ctx context.Context, email string, tenantID uuid.UUID) (*model.Invite, error) {
	query := `
		SELECT id, tenant_id, email, role, token, expires_at, accepted_at, revoked_at, created_at, created_by
		FROM tenant_invites
		WHERE tenant_id = $1 AND email = $2 AND deleted_at IS NULL
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
		revokedAt  *time.Time
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
		&revokedAt,
		&createdAt,
		&createdBy,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrInviteNotFound
		}
		return nil, err
	}

	return r.mapToDomainInvite(id, dbTenantID, dbEmail, role, token, expiresAt, acceptedAt, revokedAt, createdAt, createdBy), nil
}

// FindPendingInvitesByEmail finds all pending invites for an email across all tenants
func (r *InviteRepository) FindPendingInvitesByEmail(ctx context.Context, email string) ([]*model.Invite, error) {
	// #region agent log
	logData := map[string]interface{}{
		"location":     "invite_repository.go:212",
		"message":      "FindPendingInvitesByEmail: entry",
		"raw_email":    email,
		"timestamp":    time.Now().UnixMilli(),
		"sessionId":    "debug-session",
		"runId":        "run1",
		"hypothesisId": "A",
	}
	if logBytes, err := json.Marshal(logData); err == nil {
		if f, err := os.OpenFile("/Users/bperez/Projects/farohq-core-app/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			f.Write(append(logBytes, '\n'))
			f.Close()
		}
	}
	// #endregion

	// Normalize email: trim spaces and convert to lowercase for consistent matching
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	// #region agent log
	logData = map[string]interface{}{
		"location":         "invite_repository.go:215",
		"message":          "FindPendingInvitesByEmail: normalized email",
		"normalized_email": normalizedEmail,
		"timestamp":        time.Now().UnixMilli(),
		"sessionId":        "debug-session",
		"runId":            "run1",
		"hypothesisId":     "A",
	}
	if logBytes, err := json.Marshal(logData); err == nil {
		if f, err := os.OpenFile("/Users/bperez/Projects/farohq-core-app/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			f.Write(append(logBytes, '\n'))
			f.Close()
		}
	}
	// #endregion

	// Use LOWER() on database column for case-insensitive comparison
	// This ensures matching even if emails were stored with different casing
	query := `
		SELECT id, tenant_id, email, role, token, expires_at, accepted_at, revoked_at, created_at, created_by
		FROM tenant_invites
		WHERE LOWER(TRIM(email)) = $1 
			AND accepted_at IS NULL 
			AND revoked_at IS NULL 
			AND expires_at > NOW()
			AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, normalizedEmail)

	// #region agent log
	logData = map[string]interface{}{
		"location":     "invite_repository.go:232",
		"message":      "FindPendingInvitesByEmail: after Query",
		"query_param":  normalizedEmail,
		"error":        fmt.Sprintf("%v", err),
		"timestamp":    time.Now().UnixMilli(),
		"sessionId":    "debug-session",
		"runId":        "run1",
		"hypothesisId": "D",
	}
	if logBytes, err2 := json.Marshal(logData); err2 == nil {
		os.WriteFile("/Users/bperez/Projects/farohq-core-app/.cursor/debug.log", append(logBytes, '\n'), 0644)
	}
	// #endregion

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []*model.Invite
	rowCount := 0
	now := time.Now()
	for rows.Next() {
		rowCount++
		var (
			id         uuid.UUID
			tenantID   uuid.UUID
			dbEmail    string
			role       string
			token      string
			expiresAt  time.Time
			acceptedAt *time.Time
			revokedAt  *time.Time
			createdAt  time.Time
			createdBy  uuid.UUID
		)

		if err := rows.Scan(&id, &tenantID, &dbEmail, &role, &token, &expiresAt, &acceptedAt, &revokedAt, &createdAt, &createdBy); err != nil {
			return nil, err
		}

		// #region agent log
		logData = map[string]interface{}{
			"location":               "invite_repository.go:298",
			"message":                "FindPendingInvitesByEmail: scanned row",
			"db_email":               dbEmail,
			"normalized_db_email":    strings.ToLower(strings.TrimSpace(dbEmail)),
			"normalized_query_email": normalizedEmail,
			"expires_at":             expiresAt.Format(time.RFC3339),
			"now":                    now.Format(time.RFC3339),
			"expired":                expiresAt.Before(now),
			"accepted_at":            acceptedAt,
			"revoked_at":             revokedAt,
			"timestamp":              time.Now().UnixMilli(),
			"sessionId":              "debug-session",
			"runId":                  "run1",
			"hypothesisId":           "B",
		}
		if logBytes, err := json.Marshal(logData); err == nil {
			if f, err := os.OpenFile("/Users/bperez/Projects/farohq-core-app/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				f.Write(append(logBytes, '\n'))
				f.Close()
			}
		}
		// #endregion

		invites = append(invites, r.mapToDomainInvite(id, tenantID, dbEmail, role, token, expiresAt, acceptedAt, revokedAt, createdAt, createdBy))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// #region agent log
	logData = map[string]interface{}{
		"location":      "invite_repository.go:330",
		"message":       "FindPendingInvitesByEmail: after scanning all rows",
		"row_count":     rowCount,
		"invites_count": len(invites),
		"timestamp":     time.Now().UnixMilli(),
		"sessionId":     "debug-session",
		"runId":         "run1",
		"hypothesisId":  "B",
	}
	if logBytes, err := json.Marshal(logData); err == nil {
		if f, err := os.OpenFile("/Users/bperez/Projects/farohq-core-app/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			f.Write(append(logBytes, '\n'))
			f.Close()
		}
	}
	// #endregion

	return invites, nil
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

	if err != nil {
		// Check for unique constraint violation (duplicate invite)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// Check if it's the tenant_id + email unique constraint
			if strings.Contains(pgErr.ConstraintName, "tenant_id_email") {
				return domain.ErrPendingInviteExists
			}
		}
		return err
	}

	return nil
}

// Update updates an existing invite
func (r *InviteRepository) Update(ctx context.Context, invite *model.Invite) error {
	query := `
		UPDATE tenant_invites SET
			accepted_at = $2,
			revoked_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query,
		invite.ID(),
		invite.AcceptedAt(),
		invite.RevokedAt(),
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
func (r *InviteRepository) mapToDomainInvite(id, tenantID uuid.UUID, email, role, token string, expiresAt time.Time, acceptedAt *time.Time, revokedAt *time.Time, createdAt time.Time, createdBy uuid.UUID) *model.Invite {
	inviteRole := model.Role(role)
	return model.NewInviteWithID(id, tenantID, email, inviteRole, token, expiresAt, acceptedAt, revokedAt, createdAt, createdBy)
}
