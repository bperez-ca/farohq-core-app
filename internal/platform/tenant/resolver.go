package tenant

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

// Tenant resolution errors
var (
	ErrTenantNotFound      = errors.New("tenant not found")
	ErrTenantAccessDenied  = errors.New("user does not have access to tenant")
	ErrNoAccessibleTenants = errors.New("user has no accessible tenants")
	ErrInvalidTenantID     = errors.New("invalid tenant ID format")
)

// TenantSource represents where the tenant was resolved from
type TenantSource string

const (
	TenantSourceDomain   TenantSource = "domain"
	TenantSourceHeader   TenantSource = "x-tenant-id"
	TenantSourceURL      TenantSource = "url_param"
	TenantSourceToken    TenantSource = "user_tenants"
	TenantSourceFallback TenantSource = "fallback"
)

// TenantResolutionResult represents the result of tenant resolution
type TenantResolutionResult struct {
	TenantID     string
	Source       TenantSource
	Validated    bool
	FallbackUsed bool
}

// Resolver handles tenant resolution logic
type Resolver struct {
	db          *pgxpool.Pool
	logger      zerolog.Logger
	tenantCache *TenantCache // Optional cache for user tenant IDs
}

// NewResolver creates a new tenant resolver
func NewResolver(db *pgxpool.Pool, logger zerolog.Logger) *Resolver {
	return &Resolver{
		db:     db,
		logger: logger,
	}
}

// SetCache sets the tenant cache for the resolver
func (tr *Resolver) SetCache(cache *TenantCache) {
	tr.tenantCache = cache
}

// ResolveTenant resolves tenant from request context
func (tr *Resolver) ResolveTenant(ctx context.Context, host string) (string, error) {
	// Extract domain from host (remove port if present)
	domain := strings.Split(host, ":")[0]

	// Query database for tenant by domain
	var tenantID string
	query := `SELECT agency_id::text FROM branding WHERE domain = $1 AND deleted_at IS NULL`

	err := tr.db.QueryRow(ctx, query, domain).Scan(&tenantID)
	if err != nil {
		tr.logger.Debug().
			Str("domain", domain).
			Err(err).
			Msg("Failed to resolve tenant by domain")

		// Try to resolve from URL path (e.g., /api/v1/tenants/{id}/...)
		// This is a fallback for API-based tenant resolution
		return "", err
	}

	return tenantID, nil
}

// ResolveClient resolves client from request context (optional)
func (tr *Resolver) ResolveClient(ctx context.Context, clientID, tenantID string) (string, error) {
	if clientID == "" {
		return "", nil
	}

	// Validate client belongs to tenant
	var validClientID string
	query := `SELECT id::text FROM clients WHERE id = $1 AND agency_id = $2 AND deleted_at IS NULL`

	err := tr.db.QueryRow(ctx, query, clientID, tenantID).Scan(&validClientID)
	if err != nil {
		tr.logger.Debug().
			Str("client_id", clientID).
			Str("tenant_id", tenantID).
			Err(err).
			Msg("Failed to resolve client")
		return "", err
	}

	return validClientID, nil
}

// SetTenantContext sets tenant context in the request
func (tr *Resolver) SetTenantContext(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, "tenant_id", tenantID)
}

// SetClientContext sets client context in the request
func (tr *Resolver) SetClientContext(ctx context.Context, clientID string) context.Context {
	return context.WithValue(ctx, "client_id", clientID)
}

// GetTenantFromContext gets tenant ID from context
func GetTenantFromContext(ctx context.Context) (string, bool) {
	tenantID, ok := ctx.Value("tenant_id").(string)
	return tenantID, ok
}

// GetTenantUUIDFromContext gets tenant ID as UUID from context
func GetTenantUUIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	tenantID, ok := ctx.Value("tenant_id").(string)
	if !ok {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(tenantID)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

// GetClientFromContext gets client ID from context
func GetClientFromContext(ctx context.Context) (string, bool) {
	clientID, ok := ctx.Value("client_id").(string)
	return clientID, ok
}

// GetClientUUIDFromContext gets client ID as UUID from context
func GetClientUUIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	clientID, ok := ctx.Value("client_id").(string)
	if !ok {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(clientID)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

// RequireTenant ensures tenant context exists, returns error if not found
func RequireTenant(ctx context.Context) (string, error) {
	tenantID, ok := GetTenantFromContext(ctx)
	if !ok {
		return "", fmt.Errorf("tenant context required")
	}
	return tenantID, nil
}
// GetUserTenantIDs returns all tenant IDs the user has access to
// Uses cache if available, otherwise queries database
func (tr *Resolver) GetUserTenantIDs(ctx context.Context, userID uuid.UUID) ([]string, error) {
	// Try cache first if available
	if tr.tenantCache != nil {
		return tr.tenantCache.GetWithResolver(ctx, userID, tr)
	}

	// Fallback to direct database query
	return tr.getUserTenantIDsFromDB(ctx, userID)
}

// getUserTenantIDsFromDB queries the database for user tenant IDs
func (tr *Resolver) getUserTenantIDsFromDB(ctx context.Context, userID uuid.UUID) ([]string, error) {
	// Use DISTINCT ON to get unique tenant_ids ordered by created_at
	// DISTINCT ON requires the first ORDER BY column to match the DISTINCT ON column
	query := `
		SELECT DISTINCT ON (tenant_id) tenant_id::text, created_at
		FROM tenant_members
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY tenant_id, created_at ASC
	`

	rows, err := tr.db.Query(ctx, query, userID)
	if err != nil {
		tr.logger.Debug().
			Str("user_id", userID.String()).
			Err(err).
			Msg("Failed to query user tenant IDs")
		return nil, err
	}
	defer rows.Close()

	var tenantIDs []string
	seen := make(map[string]bool) // Track seen tenant IDs to ensure uniqueness
	for rows.Next() {
		var tenantID string
		var createdAt time.Time
		if err := rows.Scan(&tenantID, &createdAt); err != nil {
			return nil, err
		}
		// Add to map to ensure uniqueness (though DISTINCT ON should handle this)
		if !seen[tenantID] {
			tenantIDs = append(tenantIDs, tenantID)
			seen[tenantID] = true
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tenantIDs, nil
}

// GetUserAccessibleTenants returns all tenant IDs the user has access to (alias for GetUserTenantIDs)
func (tr *Resolver) GetUserAccessibleTenants(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return tr.GetUserTenantIDs(ctx, userID)
}

// ValidateUserAccess validates if user has access to a specific tenant
func (tr *Resolver) ValidateUserAccess(ctx context.Context, userID uuid.UUID, tenantID string) (bool, error) {
	// Parse tenant ID to UUID
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return false, ErrInvalidTenantID
	}

	query := `
		SELECT COUNT(*)
		FROM tenant_members
		WHERE user_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`

	var count int
	err = tr.db.QueryRow(ctx, query, userID, tenantUUID).Scan(&count)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		tr.logger.Debug().
			Str("user_id", userID.String()).
			Str("tenant_id", tenantID).
			Err(err).
			Msg("Failed to validate user access")
		return false, err
	}

	return count > 0, nil
}

// ResolveTenantByDomain resolves tenant from domain/host
func (tr *Resolver) ResolveTenantByDomain(ctx context.Context, host string) (string, error) {
	// Extract domain from host (remove port if present)
	domain := strings.Split(host, ":")[0]
	domain = strings.ToLower(domain)

	// Query database for tenant by domain
	var tenantID string
	query := `SELECT agency_id::text FROM branding WHERE domain = $1 AND deleted_at IS NULL`

	err := tr.db.QueryRow(ctx, query, domain).Scan(&tenantID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", ErrTenantNotFound
		}
		tr.logger.Debug().
			Str("domain", domain).
			Err(err).
			Msg("Failed to resolve tenant by domain")
		return "", err
	}

	return tenantID, nil
}

// ExtractTenantIDFromURL extracts tenant ID from URL path patterns
func ExtractTenantIDFromURL(path string) string {
	// Pattern: /api/v1/tenants/{id}/...
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "tenants" && i+1 < len(parts) {
			// Check if next part looks like a UUID
			if tenantID := parts[i+1]; len(tenantID) == 36 {
				// Validate it's a valid UUID format
				if _, err := uuid.Parse(tenantID); err == nil {
					return tenantID
				}
			}
		}
	}
	return ""
}

// ResolveTenantWithValidation resolves tenant from multiple sources with access validation
func (tr *Resolver) ResolveTenantWithValidation(
	ctx context.Context,
	userID uuid.UUID,
	host string,
	tenantIDHeader string,
	urlPath string,
) (*TenantResolutionResult, error) {
	// First, get user's accessible tenants
	accessibleTenants, err := tr.GetUserTenantIDs(ctx, userID)
	if err != nil {
		tr.logger.Error().
			Str("user_id", userID.String()).
			Err(err).
			Msg("Failed to get user accessible tenants")
		return nil, err
	}

	if len(accessibleTenants) == 0 {
		return nil, ErrNoAccessibleTenants
	}

	var resolvedTenantID string
	var source TenantSource

	// Priority 1: Domain-based resolution
	if resolvedTenantID == "" && host != "" {
		domainTenantID, err := tr.ResolveTenantByDomain(ctx, host)
		if err == nil && domainTenantID != "" {
			resolvedTenantID = domainTenantID
			source = TenantSourceDomain
		}
	}

	// Priority 2: X-Tenant-ID header
	if resolvedTenantID == "" && tenantIDHeader != "" {
		// Validate UUID format
		if _, err := uuid.Parse(tenantIDHeader); err == nil {
			resolvedTenantID = tenantIDHeader
			source = TenantSourceHeader
		} else {
			tr.logger.Warn().
				Str("tenant_id_header", tenantIDHeader).
				Msg("Invalid tenant ID format in X-Tenant-ID header")
		}
	}

	// Priority 3: URL parameter extraction
	if resolvedTenantID == "" && urlPath != "" {
		urlTenantID := ExtractTenantIDFromURL(urlPath)
		if urlTenantID != "" {
			resolvedTenantID = urlTenantID
			source = TenantSourceURL
		}
	}

	// Validate resolved tenant against user's accessible tenants
	validated := false
	fallbackUsed := false

	if resolvedTenantID != "" {
		// Check if resolved tenant is in user's accessible tenants
		for _, accessibleTenantID := range accessibleTenants {
			if accessibleTenantID == resolvedTenantID {
				validated = true
				break
			}
		}

		if !validated {
			// Log security warning
			tr.logger.Warn().
				Str("user_id", userID.String()).
				Str("resolved_tenant_id", resolvedTenantID).
				Str("resolution_source", string(source)).
				Strs("accessible_tenants", accessibleTenants).
				Msg("Invalid tenant access attempt - resolved tenant not in user's accessible tenants")

			// Use fallback: user's first tenant
			if len(accessibleTenants) > 0 {
				resolvedTenantID = accessibleTenants[0]
				source = TenantSourceFallback
				validated = true
				fallbackUsed = true
			} else {
				return nil, ErrTenantAccessDenied
			}
		}
	} else {
		// No tenant resolved from any source - use user's first tenant
		if len(accessibleTenants) > 0 {
			resolvedTenantID = accessibleTenants[0]
			source = TenantSourceToken
			validated = true
			fallbackUsed = false
		} else {
			return nil, ErrNoAccessibleTenants
		}
	}

	return &TenantResolutionResult{
		TenantID:     resolvedTenantID,
		Source:       source,
		Validated:    validated,
		FallbackUsed: fallbackUsed,
	}, nil
}
