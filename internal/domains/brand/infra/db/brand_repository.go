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
		SELECT agency_id, domain, subdomain, domain_type, website, verified_at, logo_url, favicon_url, 
		       primary_color, secondary_color, theme_json, hide_powered_by, email_domain, 
		       cloudflare_zone_id, domain_verification_token, ssl_status, updated_at
		FROM branding
		WHERE agency_id = $1
	`

	var (
		dbAgencyID              uuid.UUID
		brandDomain             string
		subdomain               string
		domainTypeStr           *string
		website                 string
		verifiedAt              *time.Time
		logoURL                 string
		faviconURL              string
		primaryColor            string
		secondaryColor          string
		themeJSONBytes          []byte
		hidePoweredBy           bool
		emailDomain             string
		cloudflareZoneID        string
		domainVerificationToken string
		sslStatusStr            *string
		updatedAt               time.Time
	)

	err := r.db.QueryRow(ctx, query, agencyID).Scan(
		&dbAgencyID,
		&brandDomain,
		&subdomain,
		&domainTypeStr,
		&website,
		&verifiedAt,
		&logoURL,
		&faviconURL,
		&primaryColor,
		&secondaryColor,
		&themeJSONBytes,
		&hidePoweredBy,
		&emailDomain,
		&cloudflareZoneID,
		&domainVerificationToken,
		&sslStatusStr,
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

	var domainType *model.DomainType
	if domainTypeStr != nil {
		dt := model.DomainType(*domainTypeStr)
		domainType = &dt
	}

	var sslStatus *model.SSLStatus
	if sslStatusStr != nil {
		ss := model.SSLStatus(*sslStatusStr)
		sslStatus = &ss
	}

	return model.NewBrandingWithID(
		dbAgencyID,
		brandDomain,
		subdomain,
		domainType,
		website,
		verifiedAt,
		logoURL,
		faviconURL,
		primaryColor,
		secondaryColor,
		themeJSON,
		hidePoweredBy,
		emailDomain,
		cloudflareZoneID,
		domainVerificationToken,
		sslStatus,
		updatedAt,
	), nil
}

// FindByDomain finds branding by domain (custom domain)
func (r *BrandRepository) FindByDomain(ctx context.Context, domainParam string) (*model.Branding, error) {
	query := `
		SELECT agency_id, domain, subdomain, domain_type, website, verified_at, logo_url, favicon_url, 
		       primary_color, secondary_color, theme_json, hide_powered_by, email_domain, 
		       cloudflare_zone_id, domain_verification_token, ssl_status, updated_at
		FROM branding
		WHERE domain = $1 AND domain_type = 'custom'
	`

	var (
		agencyID                uuid.UUID
		dbDomain                string
		subdomain               string
		domainTypeStr           *string
		website                 string
		verifiedAt              *time.Time
		logoURL                 string
		faviconURL              string
		primaryColor            string
		secondaryColor          string
		themeJSONBytes          []byte
		hidePoweredBy           bool
		emailDomain             string
		cloudflareZoneID        string
		domainVerificationToken string
		sslStatusStr            *string
		updatedAt               time.Time
	)

	err := r.db.QueryRow(ctx, query, domainParam).Scan(
		&agencyID,
		&dbDomain,
		&subdomain,
		&domainTypeStr,
		&website,
		&verifiedAt,
		&logoURL,
		&faviconURL,
		&primaryColor,
		&secondaryColor,
		&themeJSONBytes,
		&hidePoweredBy,
		&emailDomain,
		&cloudflareZoneID,
		&domainVerificationToken,
		&sslStatusStr,
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

	var domainType *model.DomainType
	if domainTypeStr != nil {
		dt := model.DomainType(*domainTypeStr)
		domainType = &dt
	}

	var sslStatus *model.SSLStatus
	if sslStatusStr != nil {
		ss := model.SSLStatus(*sslStatusStr)
		sslStatus = &ss
	}

	return model.NewBrandingWithID(
		agencyID,
		dbDomain,
		subdomain,
		domainType,
		website,
		verifiedAt,
		logoURL,
		faviconURL,
		primaryColor,
		secondaryColor,
		themeJSON,
		hidePoweredBy,
		emailDomain,
		cloudflareZoneID,
		domainVerificationToken,
		sslStatus,
		updatedAt,
	), nil
}

// FindBySubdomain finds branding by subdomain
func (r *BrandRepository) FindBySubdomain(ctx context.Context, subdomainParam string) (*model.Branding, error) {
	query := `
		SELECT agency_id, domain, subdomain, domain_type, website, verified_at, logo_url, favicon_url, 
		       primary_color, secondary_color, theme_json, hide_powered_by, email_domain, 
		       cloudflare_zone_id, domain_verification_token, ssl_status, updated_at
		FROM branding
		WHERE subdomain = $1
	`

	var (
		agencyID                uuid.UUID
		brandDomain             string
		subdomain               string
		domainTypeStr           *string
		website                 string
		verifiedAt              *time.Time
		logoURL                 string
		faviconURL              string
		primaryColor            string
		secondaryColor          string
		themeJSONBytes          []byte
		hidePoweredBy           bool
		emailDomain             string
		cloudflareZoneID        string
		domainVerificationToken string
		sslStatusStr            *string
		updatedAt               time.Time
	)

	err := r.db.QueryRow(ctx, query, subdomainParam).Scan(
		&agencyID,
		&brandDomain,
		&subdomain,
		&domainTypeStr,
		&website,
		&verifiedAt,
		&logoURL,
		&faviconURL,
		&primaryColor,
		&secondaryColor,
		&themeJSONBytes,
		&hidePoweredBy,
		&emailDomain,
		&cloudflareZoneID,
		&domainVerificationToken,
		&sslStatusStr,
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

	var domainType *model.DomainType
	if domainTypeStr != nil {
		dt := model.DomainType(*domainTypeStr)
		domainType = &dt
	}

	var sslStatus *model.SSLStatus
	if sslStatusStr != nil {
		ss := model.SSLStatus(*sslStatusStr)
		sslStatus = &ss
	}

	return model.NewBrandingWithID(
		agencyID,
		brandDomain,
		subdomain,
		domainType,
		website,
		verifiedAt,
		logoURL,
		faviconURL,
		primaryColor,
		secondaryColor,
		themeJSON,
		hidePoweredBy,
		emailDomain,
		cloudflareZoneID,
		domainVerificationToken,
		sslStatus,
		updatedAt,
	), nil
}

// CheckSubdomainExists checks if a subdomain already exists
func (r *BrandRepository) CheckSubdomainExists(ctx context.Context, subdomainParam string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM branding WHERE subdomain = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, query, subdomainParam).Scan(&exists)
	return exists, err
}

// Save saves a new branding
func (r *BrandRepository) Save(ctx context.Context, branding *model.Branding) error {
	themeJSONBytes, _ := json.Marshal(branding.ThemeJSON())

	var domainTypeStr *string
	if branding.DomainType() != nil {
		dt := string(*branding.DomainType())
		domainTypeStr = &dt
	}

	var sslStatusStr *string
	if branding.SSLStatus() != nil {
		ss := string(*branding.SSLStatus())
		sslStatusStr = &ss
	}

	query := `
		INSERT INTO branding (agency_id, domain, subdomain, domain_type, website, verified_at, logo_url, favicon_url, 
		                      primary_color, secondary_color, theme_json, hide_powered_by, email_domain, 
		                      cloudflare_zone_id, domain_verification_token, ssl_status, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		ON CONFLICT (agency_id) DO UPDATE SET
			domain = EXCLUDED.domain,
			subdomain = EXCLUDED.subdomain,
			domain_type = EXCLUDED.domain_type,
			website = EXCLUDED.website,
			verified_at = EXCLUDED.verified_at,
			logo_url = EXCLUDED.logo_url,
			favicon_url = EXCLUDED.favicon_url,
			primary_color = EXCLUDED.primary_color,
			secondary_color = EXCLUDED.secondary_color,
			theme_json = EXCLUDED.theme_json,
			hide_powered_by = EXCLUDED.hide_powered_by,
			email_domain = EXCLUDED.email_domain,
			cloudflare_zone_id = EXCLUDED.cloudflare_zone_id,
			domain_verification_token = EXCLUDED.domain_verification_token,
			ssl_status = EXCLUDED.ssl_status,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.Exec(ctx, query,
		branding.AgencyID(),
		branding.Domain(),
		branding.Subdomain(),
		domainTypeStr,
		branding.Website(),
		branding.VerifiedAt(),
		branding.LogoURL(),
		branding.FaviconURL(),
		branding.PrimaryColor(),
		branding.SecondaryColor(),
		themeJSONBytes,
		branding.HidePoweredBy(),
		branding.EmailDomain(),
		branding.CloudflareZoneID(),
		branding.DomainVerificationToken(),
		sslStatusStr,
		branding.UpdatedAt(),
	)

	return err
}

// Update updates an existing branding
func (r *BrandRepository) Update(ctx context.Context, branding *model.Branding) error {
	themeJSONBytes, _ := json.Marshal(branding.ThemeJSON())

	var domainTypeStr *string
	if branding.DomainType() != nil {
		dt := string(*branding.DomainType())
		domainTypeStr = &dt
	}

	var sslStatusStr *string
	if branding.SSLStatus() != nil {
		ss := string(*branding.SSLStatus())
		sslStatusStr = &ss
	}

	query := `
		UPDATE branding SET
			domain = $2,
			subdomain = $3,
			domain_type = $4,
			website = $5,
			verified_at = $6,
			logo_url = $7,
			favicon_url = $8,
			primary_color = $9,
			secondary_color = $10,
			theme_json = $11,
			hide_powered_by = $12,
			email_domain = $13,
			cloudflare_zone_id = $14,
			domain_verification_token = $15,
			ssl_status = $16,
			updated_at = $17
		WHERE agency_id = $1
	`

	result, err := r.db.Exec(ctx, query,
		branding.AgencyID(),
		branding.Domain(),
		branding.Subdomain(),
		domainTypeStr,
		branding.Website(),
		branding.VerifiedAt(),
		branding.LogoURL(),
		branding.FaviconURL(),
		branding.PrimaryColor(),
		branding.SecondaryColor(),
		themeJSONBytes,
		branding.HidePoweredBy(),
		branding.EmailDomain(),
		branding.CloudflareZoneID(),
		branding.DomainVerificationToken(),
		sslStatusStr,
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
		SELECT agency_id, domain, subdomain, domain_type, website, verified_at, logo_url, favicon_url, 
		       primary_color, secondary_color, theme_json, hide_powered_by, email_domain, 
		       cloudflare_zone_id, domain_verification_token, ssl_status, updated_at
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
			agencyID                uuid.UUID
			brandDomain             string
			subdomain               string
			domainTypeStr           *string
			website                 string
			verifiedAt              *time.Time
			logoURL                 string
			faviconURL              string
			primaryColor            string
			secondaryColor          string
			themeJSONBytes          []byte
			hidePoweredBy           bool
			emailDomain             string
			cloudflareZoneID        string
			domainVerificationToken string
			sslStatusStr            *string
			updatedAt               time.Time
		)

		if err := rows.Scan(
			&agencyID,
			&brandDomain,
			&subdomain,
			&domainTypeStr,
			&website,
			&verifiedAt,
			&logoURL,
			&faviconURL,
			&primaryColor,
			&secondaryColor,
			&themeJSONBytes,
			&hidePoweredBy,
			&emailDomain,
			&cloudflareZoneID,
			&domainVerificationToken,
			&sslStatusStr,
			&updatedAt,
		); err != nil {
			return nil, err
		}

		var themeJSON map[string]interface{}
		if len(themeJSONBytes) > 0 {
			json.Unmarshal(themeJSONBytes, &themeJSON)
		}

		var domainType *model.DomainType
		if domainTypeStr != nil {
			dt := model.DomainType(*domainTypeStr)
			domainType = &dt
		}

		var sslStatus *model.SSLStatus
		if sslStatusStr != nil {
			ss := model.SSLStatus(*sslStatusStr)
			sslStatus = &ss
		}

		brands = append(brands, model.NewBrandingWithID(
			agencyID,
			brandDomain,
			subdomain,
			domainType,
			website,
			verifiedAt,
			logoURL,
			faviconURL,
			primaryColor,
			secondaryColor,
			themeJSON,
			hidePoweredBy,
			emailDomain,
			cloudflareZoneID,
			domainVerificationToken,
			sslStatus,
			updatedAt,
		))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return brands, nil
}

