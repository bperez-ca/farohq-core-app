package httpserver

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"farohq-core-app/internal/platform/tenant"
)

// TenantResolution middleware that resolves tenant from request and sets RLS context
func TenantResolution(tenantResolver *tenant.Resolver, db *pgxpool.Pool, logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip tenant resolution for health endpoints
			if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" || r.URL.Path == "/" {
				next.ServeHTTP(w, r)
				return
			}

			// Extract host from request
			host := r.Host
			if host == "" {
				host = r.Header.Get("Host")
			}

			// Try to resolve tenant from domain
			tenantID, err := tenantResolver.ResolveTenant(r.Context(), host)
			if err != nil {
				// Try fallback to X-Tenant-ID header
				tenantIDHeader := r.Header.Get("X-Tenant-ID")
				if tenantIDHeader != "" {
					tenantID = tenantIDHeader
					err = nil
				} else {
					logger.Debug().
						Str("host", host).
						Err(err).
						Msg("Failed to resolve tenant by domain")
					// For public routes, allow proceeding without tenant context
					// The route handler will handle authorization
					next.ServeHTTP(w, r)
					return
				}
			}

			// Resolve client (optional - from query param or header)
			clientID := r.URL.Query().Get("client_id")
			if clientID == "" {
				clientID = r.Header.Get("X-Client-ID")
			}

			// Set tenant context in Go context
			ctx := tenantResolver.SetTenantContext(r.Context(), tenantID)
			if clientID != "" {
				ctx = tenantResolver.SetClientContext(ctx, clientID)
			}
			r = r.WithContext(ctx)

			// Set RLS context in database session
			// Get a connection from the pool to set session variables
			conn, err := db.Acquire(r.Context())
			if err != nil {
				logger.Error().Err(err).Msg("Failed to acquire database connection for RLS context")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			defer conn.Release()

			// Set tenant_id for RLS
			_, err = conn.Exec(r.Context(), "SELECT set_config('lv.tenant_id', $1, true)", tenantID)
			if err != nil {
				logger.Error().
					Str("tenant_id", tenantID).
					Err(err).
					Msg("Failed to set tenant context in database")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Set client_id for RLS (if provided)
			if clientID != "" {
				// Validate client belongs to tenant
				var validClientID string
				err := conn.QueryRow(r.Context(), `
					SELECT id::text 
					FROM clients 
					WHERE id = $1 AND agency_id = $2 AND deleted_at IS NULL
				`, clientID, tenantID).Scan(&validClientID)
				if err != nil {
					logger.Warn().
						Str("client_id", clientID).
						Str("tenant_id", tenantID).
						Err(err).
						Msg("Client not found or doesn't belong to tenant, skipping client context")
					// Don't fail the request, just skip client context
				} else {
					_, err = conn.Exec(r.Context(), "SELECT set_config('lv.client_id', $1, true)", validClientID)
					if err != nil {
						logger.Error().
							Str("client_id", validClientID).
							Err(err).
							Msg("Failed to set client context in database")
						// Don't fail the request, RLS will work with just tenant_id
					}
				}
			} else {
				// Clear client_id if not provided
				conn.Exec(r.Context(), "SELECT set_config('lv.client_id', '', true)")
			}

			next.ServeHTTP(w, r)
		})
	}
}


// RequireTenantContext middleware ensures tenant context exists
// This should be applied to protected routes that require tenant context
func RequireTenantContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// #region agent log
		logFile, _ := os.OpenFile("/Users/bperez/Projects/farohq-core-app/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		json.NewEncoder(logFile).Encode(map[string]interface{}{"timestamp": time.Now().UnixMilli(), "location": "tenant.go:122", "message": "RequireTenantContext middleware executed", "hypothesisId": "H2", "sessionId": "debug-session", "runId": "run1", "data": map[string]interface{}{"path": r.URL.Path, "method": r.Method}})
		logFile.Close()
		// #endregion
		tenantID, ok := tenant.GetTenantFromContext(r.Context())
		if !ok {
			// #region agent log
			logFile2, _ := os.OpenFile("/Users/bperez/Projects/farohq-core-app/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			json.NewEncoder(logFile2).Encode(map[string]interface{}{"timestamp": time.Now().UnixMilli(), "location": "tenant.go:125", "message": "RequireTenantContext: tenant context missing, rejecting", "hypothesisId": "H2", "sessionId": "debug-session", "runId": "run1", "data": map[string]interface{}{"path": r.URL.Path, "method": r.Method}})
			logFile2.Close()
			// #endregion
			http.Error(w, "Failed to resolve tenant. Provide X-Tenant-ID header or use a tenant domain.", http.StatusBadRequest)
			return
		}
		// Ensure tenantID is not empty
		if tenantID == "" {
			http.Error(w, "Failed to resolve tenant. Provide X-Tenant-ID header or use a tenant domain.", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}
