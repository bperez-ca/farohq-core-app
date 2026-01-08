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

// ClientMemberRepository implements the outbound.ClientMemberRepository interface
type ClientMemberRepository struct {
	db *pgxpool.Pool
}

// NewClientMemberRepository creates a new PostgreSQL client member repository
func NewClientMemberRepository(db *pgxpool.Pool) outbound.ClientMemberRepository {
	return &ClientMemberRepository{
		db: db,
	}
}

// Save saves or updates a client member
func (r *ClientMemberRepository) Save(ctx context.Context, member *model.ClientMember) error {
	query := `
		INSERT INTO client_members (id, client_id, user_id, role, location_id, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (client_id, user_id, location_id) 
		DO UPDATE SET
			role = EXCLUDED.role,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
	`

	_, err := r.db.Exec(ctx, query,
		member.ID(),
		member.ClientID(),
		member.UserID(),
		string(member.Role()),
		member.LocationID(),
		member.CreatedAt(),
		member.UpdatedAt(),
		member.DeletedAt(),
	)

	return err
}

// FindByID finds a client member by ID
func (r *ClientMemberRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.ClientMember, error) {
	query := `
		SELECT id, client_id, user_id, role, location_id, created_at, updated_at, deleted_at
		FROM client_members
		WHERE id = $1 AND deleted_at IS NULL
	`

	var (
		dbID       uuid.UUID
		clientID   uuid.UUID
		userID     uuid.UUID
		role       string
		locationID *uuid.UUID
		createdAt  time.Time
		updatedAt  time.Time
		deletedAt  *time.Time
	)

	err := r.db.QueryRow(ctx, query, id).Scan(
		&dbID,
		&clientID,
		&userID,
		&role,
		&locationID,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrClientMemberNotFound
		}
		return nil, err
	}

	return r.mapToDomainMember(dbID, clientID, userID, role, locationID, createdAt, updatedAt, deletedAt), nil
}

// FindByClientAndUser finds a client member by client ID and user ID
func (r *ClientMemberRepository) FindByClientAndUser(ctx context.Context, clientID, userID uuid.UUID, locationID *uuid.UUID) (*model.ClientMember, error) {
	var query string
	var row pgx.Row

	if locationID != nil {
		query = `
			SELECT id, client_id, user_id, role, location_id, created_at, updated_at, deleted_at
			FROM client_members
			WHERE client_id = $1 AND user_id = $2 AND location_id = $3 AND deleted_at IS NULL
		`
		row = r.db.QueryRow(ctx, query, clientID, userID, locationID)
	} else {
		query = `
			SELECT id, client_id, user_id, role, location_id, created_at, updated_at, deleted_at
			FROM client_members
			WHERE client_id = $1 AND user_id = $2 AND location_id IS NULL AND deleted_at IS NULL
		`
		row = r.db.QueryRow(ctx, query, clientID, userID)
	}

	var (
		id         uuid.UUID
		dbClientID uuid.UUID
		dbUserID   uuid.UUID
		role       string
		dbLocationID *uuid.UUID
		createdAt  time.Time
		updatedAt  time.Time
		deletedAt  *time.Time
	)

	err := row.Scan(
		&id,
		&dbClientID,
		&dbUserID,
		&role,
		&dbLocationID,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrClientMemberNotFound
		}
		return nil, err
	}

	return r.mapToDomainMember(id, dbClientID, dbUserID, role, dbLocationID, createdAt, updatedAt, deletedAt), nil
}

// ListByClient lists all members for a client
func (r *ClientMemberRepository) ListByClient(ctx context.Context, clientID uuid.UUID, locationID *uuid.UUID) ([]*model.ClientMember, error) {
	var query string
	var rows pgx.Rows
	var err error

	if locationID != nil {
		query = `
			SELECT id, client_id, user_id, role, location_id, created_at, updated_at, deleted_at
			FROM client_members
			WHERE client_id = $1 AND location_id = $2 AND deleted_at IS NULL
			ORDER BY created_at ASC
		`
		rows, err = r.db.Query(ctx, query, clientID, locationID)
	} else {
		query = `
			SELECT id, client_id, user_id, role, location_id, created_at, updated_at, deleted_at
			FROM client_members
			WHERE client_id = $1 AND deleted_at IS NULL
			ORDER BY created_at ASC
		`
		rows, err = r.db.Query(ctx, query, clientID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*model.ClientMember
	for rows.Next() {
		var (
			id         uuid.UUID
			dbClientID uuid.UUID
			userID     uuid.UUID
			role       string
			dbLocationID *uuid.UUID
			createdAt  time.Time
			updatedAt  time.Time
			deletedAt  *time.Time
		)

		if err := rows.Scan(&id, &dbClientID, &userID, &role, &dbLocationID, &createdAt, &updatedAt, &deletedAt); err != nil {
			return nil, err
		}

		members = append(members, r.mapToDomainMember(id, dbClientID, userID, role, dbLocationID, createdAt, updatedAt, deletedAt))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

// CountByClient counts members for a client (excluding soft-deleted)
func (r *ClientMemberRepository) CountByClient(ctx context.Context, clientID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM client_members
		WHERE client_id = $1 AND deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRow(ctx, query, clientID).Scan(&count)
	return count, err
}

// CountByClientAndLocation counts members for a client and location (excluding soft-deleted)
func (r *ClientMemberRepository) CountByClientAndLocation(ctx context.Context, clientID uuid.UUID, locationID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM client_members
		WHERE client_id = $1 AND location_id = $2 AND deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRow(ctx, query, clientID, locationID).Scan(&count)
	return count, err
}

// mapToDomainMember maps database row to domain client member
func (r *ClientMemberRepository) mapToDomainMember(id, clientID, userID uuid.UUID, role string, locationID *uuid.UUID, createdAt, updatedAt time.Time, deletedAt *time.Time) *model.ClientMember {
	memberRole := model.Role(role)
	return model.NewClientMemberWithID(id, clientID, userID, memberRole, locationID, createdAt, updatedAt, deletedAt)
}

