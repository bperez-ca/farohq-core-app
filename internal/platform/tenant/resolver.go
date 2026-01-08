package tenant

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

// Resolver handles tenant resolution logic
type Resolver struct {
	db     *pgxpool.Pool
	logger zerolog.Logger
}

// NewResolver creates a new tenant resolver
func NewResolver(db *pgxpool.Pool, logger zerolog.Logger) *Resolver {
	return &Resolver{
		db:     db,
		logger: logger,
	}
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

