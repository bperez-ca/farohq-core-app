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

// ClientRepository implements the outbound.ClientRepository interface
type ClientRepository struct {
	db *pgxpool.Pool
}

// NewClientRepository creates a new PostgreSQL client repository
func NewClientRepository(db *pgxpool.Pool) outbound.ClientRepository {
	return &ClientRepository{
		db: db,
	}
}

// Save saves or updates a client
func (r *ClientRepository) Save(ctx context.Context, client *model.Client) error {
	var tierStr *string
	if client.Tier() != "" {
		t := client.Tier().String()
		tierStr = &t
	}

	query := `
		INSERT INTO clients (id, agency_id, name, slug, tier, status, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (agency_id, slug) 
		DO UPDATE SET
			name = EXCLUDED.name,
			tier = EXCLUDED.tier,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
	`

	_, err := r.db.Exec(ctx, query,
		client.ID(),
		client.AgencyID(),
		client.Name(),
		client.Slug(),
		tierStr,
		string(client.Status()),
		client.CreatedAt(),
		client.UpdatedAt(),
		client.DeletedAt(),
	)

	return err
}

// FindByID finds a client by ID
func (r *ClientRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Client, error) {
	query := `
		SELECT id, agency_id, name, slug, tier, status, created_at, updated_at, deleted_at
		FROM clients
		WHERE id = $1 AND deleted_at IS NULL
	`

	var (
		dbID      uuid.UUID
		agencyID  uuid.UUID
		name      string
		slug      string
		tier      *string
		status    string
		createdAt time.Time
		updatedAt time.Time
		deletedAt *time.Time
	)

	err := r.db.QueryRow(ctx, query, id).Scan(
		&dbID,
		&agencyID,
		&name,
		&slug,
		&tier,
		&status,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrClientNotFound
		}
		return nil, err
	}

	return r.mapToDomainClient(dbID, agencyID, name, slug, tier, status, createdAt, updatedAt, deletedAt), nil
}

// FindBySlug finds a client by slug within an agency
func (r *ClientRepository) FindBySlug(ctx context.Context, agencyID uuid.UUID, slug string) (*model.Client, error) {
	query := `
		SELECT id, agency_id, name, slug, tier, status, created_at, updated_at, deleted_at
		FROM clients
		WHERE agency_id = $1 AND slug = $2 AND deleted_at IS NULL
	`

	var (
		id        uuid.UUID
		dbAgencyID uuid.UUID
		name      string
		dbSlug    string
		tier      *string
		status    string
		createdAt time.Time
		updatedAt time.Time
		deletedAt *time.Time
	)

	err := r.db.QueryRow(ctx, query, agencyID, slug).Scan(
		&id,
		&dbAgencyID,
		&name,
		&dbSlug,
		&tier,
		&status,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrClientNotFound
		}
		return nil, err
	}

	return r.mapToDomainClient(id, dbAgencyID, name, dbSlug, tier, status, createdAt, updatedAt, deletedAt), nil
}

// ListByAgency lists all clients for an agency
func (r *ClientRepository) ListByAgency(ctx context.Context, agencyID uuid.UUID) ([]*model.Client, error) {
	query := `
		SELECT id, agency_id, name, slug, tier, status, created_at, updated_at, deleted_at
		FROM clients
		WHERE agency_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, agencyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []*model.Client
	for rows.Next() {
		var (
			id        uuid.UUID
			dbAgencyID uuid.UUID
			name      string
			slug      string
			tier      *string
			status    string
			createdAt time.Time
			updatedAt time.Time
			deletedAt *time.Time
		)

		if err := rows.Scan(&id, &dbAgencyID, &name, &slug, &tier, &status, &createdAt, &updatedAt, &deletedAt); err != nil {
			return nil, err
		}

		clients = append(clients, r.mapToDomainClient(id, dbAgencyID, name, slug, tier, status, createdAt, updatedAt, deletedAt))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return clients, nil
}

// CountByAgency counts clients for an agency (excluding soft-deleted)
func (r *ClientRepository) CountByAgency(ctx context.Context, agencyID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM clients
		WHERE agency_id = $1 AND deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRow(ctx, query, agencyID).Scan(&count)
	return count, err
}

// CountByAgencyAndTier counts clients for an agency by tier (excluding soft-deleted)
func (r *ClientRepository) CountByAgencyAndTier(ctx context.Context, agencyID uuid.UUID, tier model.Tier) (int, error) {
	tierStr := tier.String()
	query := `
		SELECT COUNT(*)
		FROM clients
		WHERE agency_id = $1 AND tier = $2 AND deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRow(ctx, query, agencyID, tierStr).Scan(&count)
	return count, err
}

// mapToDomainClient maps database row to domain client
func (r *ClientRepository) mapToDomainClient(id, agencyID uuid.UUID, name, slug string, tier *string, status string, createdAt, updatedAt time.Time, deletedAt *time.Time) *model.Client {
	clientStatus := model.ClientStatus(status)
	var domainTier model.Tier
	if tier != nil {
		domainTier = model.Tier(*tier)
	}
	return model.NewClientWithID(id, agencyID, name, slug, domainTier, clientStatus, createdAt, updatedAt, deletedAt)
}

