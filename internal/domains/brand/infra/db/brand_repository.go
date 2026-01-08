package db

import (
	"context"
	"encoding/json"
	"time"

	"farohq-core-app/internal/domains/brand/domain"
	"farohq-core-app/internal/domains/brand/domain/model"
	"farohq-core-app/internal/domains/brand/domain/ports/outbound"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BrandRepository implements the outbound.BrandRepository interface
type BrandRepository struct {
	db *pgxpool.Pool
}

// NewBrandRepository creates a new PostgreSQL brand repository
func NewBrandRepository(db *pgxpool.Pool) outbound.BrandRepository {
	return &BrandRepository{
		db: db,
	}
}

// FindByAgencyID finds branding by agency ID
func (r *BrandRepository) FindByAgencyID(ctx context.Context, agencyID uuid.UUID) (*model.Branding, error) {
	query := `
		SELECT agency_id, domain, verified_at, logo_url, favicon_url, primary_color, secondary_color, theme_json, updated_at
		FROM branding
		WHERE agency_id = $1
	`

	var (
		dbAgencyID      uuid.UUID
		brandDomain     string
		verifiedAt      *time.Time
		logoURL         string
		faviconURL      string
		primaryColor    string
		secondaryColor  string
		themeJSONBytes  []byte
		updatedAt       time.Time
	)

	err := r.db.QueryRow(ctx, query, agencyID).Scan(
		&dbAgencyID,
		&brandDomain,
		&verifiedAt,
		&logoURL,
		&faviconURL,
		&primaryColor,
		&secondaryColor,
		&themeJSONBytes,
		&updatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrBrandingNotFound
		}
		return nil, err
	}

	var themeJSON map[string]interface{}
	if len(themeJSONBytes) > 0 {
		json.Unmarshal(themeJSONBytes, &themeJSON)
	}

	return model.NewBrandingWithID(dbAgencyID, brandDomain, verifiedAt, logoURL, faviconURL, primaryColor, secondaryColor, themeJSON, updatedAt), nil
}

// FindByDomain finds branding by domain
func (r *BrandRepository) FindByDomain(ctx context.Context, domainParam string) (*model.Branding, error) {
	query := `
		SELECT agency_id, domain, verified_at, logo_url, favicon_url, primary_color, secondary_color, theme_json, updated_at
		FROM branding
		WHERE domain = $1
	`

	var (
		agencyID        uuid.UUID
		dbDomain        string
		verifiedAt      *time.Time
		logoURL         string
		faviconURL      string
		primaryColor    string
		secondaryColor  string
		themeJSONBytes  []byte
		updatedAt       time.Time
	)

	err := r.db.QueryRow(ctx, query, domainParam).Scan(
		&agencyID,
		&dbDomain,
		&verifiedAt,
		&logoURL,
		&faviconURL,
		&primaryColor,
		&secondaryColor,
		&themeJSONBytes,
		&updatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrBrandingNotFound
		}
		return nil, err
	}

	var themeJSON map[string]interface{}
	if len(themeJSONBytes) > 0 {
		json.Unmarshal(themeJSONBytes, &themeJSON)
	}

	return model.NewBrandingWithID(agencyID, dbDomain, verifiedAt, logoURL, faviconURL, primaryColor, secondaryColor, themeJSON, updatedAt), nil
}

// Save saves a new branding
func (r *BrandRepository) Save(ctx context.Context, branding *model.Branding) error {
	themeJSONBytes, _ := json.Marshal(branding.ThemeJSON())

	query := `
		INSERT INTO branding (agency_id, domain, verified_at, logo_url, favicon_url, primary_color, secondary_color, theme_json, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (agency_id) DO UPDATE SET
			domain = EXCLUDED.domain,
			logo_url = EXCLUDED.logo_url,
			favicon_url = EXCLUDED.favicon_url,
			primary_color = EXCLUDED.primary_color,
			secondary_color = EXCLUDED.secondary_color,
			theme_json = EXCLUDED.theme_json,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.Exec(ctx, query,
		branding.AgencyID(),
		branding.Domain(),
		branding.VerifiedAt(),
		branding.LogoURL(),
		branding.FaviconURL(),
		branding.PrimaryColor(),
		branding.SecondaryColor(),
		themeJSONBytes,
		branding.UpdatedAt(),
	)

	return err
}

// Update updates an existing branding
func (r *BrandRepository) Update(ctx context.Context, branding *model.Branding) error {
	themeJSONBytes, _ := json.Marshal(branding.ThemeJSON())

	query := `
		UPDATE branding SET
			domain = $2,
			logo_url = $3,
			favicon_url = $4,
			primary_color = $5,
			secondary_color = $6,
			theme_json = $7,
			updated_at = $8
		WHERE agency_id = $1
	`

	result, err := r.db.Exec(ctx, query,
		branding.AgencyID(),
		branding.Domain(),
		branding.LogoURL(),
		branding.FaviconURL(),
		branding.PrimaryColor(),
		branding.SecondaryColor(),
		themeJSONBytes,
		branding.UpdatedAt(),
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrBrandingNotFound
	}

	return nil
}

// Delete deletes a branding
func (r *BrandRepository) Delete(ctx context.Context, agencyID uuid.UUID) error {
	query := `DELETE FROM branding WHERE agency_id = $1`

	result, err := r.db.Exec(ctx, query, agencyID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrBrandingNotFound
	}

	return nil
}

// ListByAgencyID lists all brands for an agency
func (r *BrandRepository) ListByAgencyID(ctx context.Context, agencyID uuid.UUID) ([]*model.Branding, error) {
	query := `
		SELECT agency_id, domain, verified_at, logo_url, favicon_url, primary_color, secondary_color, theme_json, updated_at
		FROM branding
		WHERE agency_id = $1
	`

	rows, err := r.db.Query(ctx, query, agencyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var brands []*model.Branding
	for rows.Next() {
		var (
			agencyID        uuid.UUID
			domain          string
			verifiedAt      *time.Time
			logoURL         string
			faviconURL      string
			primaryColor    string
			secondaryColor  string
			themeJSONBytes  []byte
			updatedAt       time.Time
		)

		if err := rows.Scan(&agencyID, &domain, &verifiedAt, &logoURL, &faviconURL, &primaryColor, &secondaryColor, &themeJSONBytes, &updatedAt); err != nil {
			return nil, err
		}

		var themeJSON map[string]interface{}
		if len(themeJSONBytes) > 0 {
			json.Unmarshal(themeJSONBytes, &themeJSON)
		}

		brands = append(brands, model.NewBrandingWithID(agencyID, domain, verifiedAt, logoURL, faviconURL, primaryColor, secondaryColor, themeJSON, updatedAt))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return brands, nil
}

