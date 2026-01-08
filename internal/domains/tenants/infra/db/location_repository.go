package db

import (
	"context"
	"encoding/json"
	"time"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/model"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LocationRepository implements the outbound.LocationRepository interface
type LocationRepository struct {
	db *pgxpool.Pool
}

// NewLocationRepository creates a new PostgreSQL location repository
func NewLocationRepository(db *pgxpool.Pool) outbound.LocationRepository {
	return &LocationRepository{
		db: db,
	}
}

// Save saves or updates a location
func (r *LocationRepository) Save(ctx context.Context, location *model.Location) error {
	addressJSON, _ := json.Marshal(location.Address())
	businessHoursJSON, _ := json.Marshal(location.BusinessHours())

	query := `
		INSERT INTO locations (id, client_id, name, address, phone, business_hours, categories, is_active, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) 
		DO UPDATE SET
			name = EXCLUDED.name,
			address = EXCLUDED.address,
			phone = EXCLUDED.phone,
			business_hours = EXCLUDED.business_hours,
			categories = EXCLUDED.categories,
			is_active = EXCLUDED.is_active,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
	`

	_, err := r.db.Exec(ctx, query,
		location.ID(),
		location.ClientID(),
		location.Name(),
		addressJSON,
		location.Phone(),
		businessHoursJSON,
		location.Categories(),
		location.IsActive(),
		location.CreatedAt(),
		location.UpdatedAt(),
		location.DeletedAt(),
	)

	return err
}

// FindByID finds a location by ID
func (r *LocationRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Location, error) {
	query := `
		SELECT id, client_id, name, address, phone, business_hours, categories, is_active, created_at, updated_at, deleted_at
		FROM locations
		WHERE id = $1 AND deleted_at IS NULL
	`

	var (
		dbID            uuid.UUID
		clientID        uuid.UUID
		name            string
		addressJSON     []byte
		phone           string
		businessHoursJSON []byte
		categories      []string
		isActive        bool
		createdAt       time.Time
		updatedAt       time.Time
		deletedAt       *time.Time
	)

	err := r.db.QueryRow(ctx, query, id).Scan(
		&dbID,
		&clientID,
		&name,
		&addressJSON,
		&phone,
		&businessHoursJSON,
		&categories,
		&isActive,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrLocationNotFound
		}
		return nil, err
	}

	var address map[string]interface{}
	if len(addressJSON) > 0 {
		json.Unmarshal(addressJSON, &address)
	}

	var businessHours map[string]interface{}
	if len(businessHoursJSON) > 0 {
		json.Unmarshal(businessHoursJSON, &businessHours)
	}

	return r.mapToDomainLocation(dbID, clientID, name, phone, address, businessHours, categories, isActive, createdAt, updatedAt, deletedAt), nil
}

// ListByClient lists all locations for a client
func (r *LocationRepository) ListByClient(ctx context.Context, clientID uuid.UUID) ([]*model.Location, error) {
	query := `
		SELECT id, client_id, name, address, phone, business_hours, categories, is_active, created_at, updated_at, deleted_at
		FROM locations
		WHERE client_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []*model.Location
	for rows.Next() {
		var (
			id              uuid.UUID
			dbClientID      uuid.UUID
			name            string
			addressJSON     []byte
			phone           string
			businessHoursJSON []byte
			categories      []string
			isActive        bool
			createdAt       time.Time
			updatedAt       time.Time
			deletedAt       *time.Time
		)

		if err := rows.Scan(&id, &dbClientID, &name, &addressJSON, &phone, &businessHoursJSON, &categories, &isActive, &createdAt, &updatedAt, &deletedAt); err != nil {
			return nil, err
		}

		var address map[string]interface{}
		if len(addressJSON) > 0 {
			json.Unmarshal(addressJSON, &address)
		}

		var businessHours map[string]interface{}
		if len(businessHoursJSON) > 0 {
			json.Unmarshal(businessHoursJSON, &businessHours)
		}

		locations = append(locations, r.mapToDomainLocation(id, dbClientID, name, phone, address, businessHours, categories, isActive, createdAt, updatedAt, deletedAt))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return locations, nil
}

// CountByClient counts locations for a client (excluding soft-deleted)
func (r *LocationRepository) CountByClient(ctx context.Context, clientID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM locations
		WHERE client_id = $1 AND deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRow(ctx, query, clientID).Scan(&count)
	return count, err
}

// mapToDomainLocation maps database row to domain location
func (r *LocationRepository) mapToDomainLocation(id, clientID uuid.UUID, name, phone string, address, businessHours map[string]interface{}, categories []string, isActive bool, createdAt, updatedAt time.Time, deletedAt *time.Time) *model.Location {
	return model.NewLocationWithID(id, clientID, name, phone, address, businessHours, categories, isActive, createdAt, updatedAt, deletedAt)
}

