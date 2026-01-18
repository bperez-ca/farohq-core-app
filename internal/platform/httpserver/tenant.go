package httpserver

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"farohq-core-app/internal/platform/tenant"
	users_outbound "farohq-core-app/internal/domains/users/domain/ports/outbound"
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


// TenantResolutionWithAuth middleware that resolves tenant AFTER authentication
// This ensures user context is available for access validation
// tenantCache is optional - if nil, cache is not used
func TenantResolutionWithAuth(
	tenantResolver *tenant.Resolver,
	tenantCache *tenant.TenantCache,
	userRepo users_outbound.UserRepository,
	db *pgxpool.Pool,
	logger zerolog.Logger,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip tenant resolution for public routes
			if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" || r.URL.Path == "/" {
				next.ServeHTTP(w, r)
				return
			}

			// Skip tenant resolution for routes that don't need tenant context
			// These routes need auth but not tenant: /api/v1/tenants/my-orgs, /api/v1/auth/me, /api/v1/users/sync, POST /api/v1/tenants, POST /api/v1/invites/accept
			if r.URL.Path == "/api/v1/tenants/my-orgs" ||
				r.URL.Path == "/api/v1/auth/me" ||
				r.URL.Path == "/api/v1/users/sync" ||
				(r.Method == "POST" && r.URL.Path == "/api/v1/tenants") ||
				(r.Method == "POST" && r.URL.Path == "/api/v1/tenants/onboard") ||
				(r.Method == "POST" && r.URL.Path == "/api/v1/invites/accept") {
				next.ServeHTTP(w, r)
				return
			}

			// Extract user_id from context (set by RequireAuth middleware)
			clerkUserID, ok := r.Context().Value("user_id").(string)
			if !ok {
				logger.Warn().
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Msg("TenantResolutionWithAuth: user_id not found in context (auth middleware may not have run)")
				// If auth middleware hasn't run, we can't resolve tenant
				// Let the handler deal with it (it will fail auth check)
				next.ServeHTTP(w, r)
				return
			}

			// Look up user by Clerk user ID to get database UUID
			user, err := userRepo.FindByClerkUserID(r.Context(), clerkUserID)
			if err != nil {
				logger.Error().
					Str("clerk_user_id", clerkUserID).
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Err(err).
					Msg("Failed to find user by Clerk user ID for tenant resolution")
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}

			// Extract resolution sources
			host := r.Host
			if host == "" {
				host = r.Header.Get("Host")
			}
			tenantIDHeader := r.Header.Get("X-Tenant-ID")
			urlPath := r.URL.Path

			// Resolve tenant with validation
			// Cache is automatically used by resolver if set
			result, err := tenantResolver.ResolveTenantWithValidation(
				r.Context(),
				user.ID(),
				host,
				tenantIDHeader,
				urlPath,
			)

			if err != nil {
				// Handle specific errors
				switch err {
				case tenant.ErrNoAccessibleTenants:
					logger.Warn().
						Str("user_id", user.ID().String()).
						Str("clerk_user_id", clerkUserID).
						Str("method", r.Method).
						Str("path", r.URL.Path).
						Msg("User has no accessible tenants")
					http.Error(w, "You don't have access to any organizations. Please contact support.", http.StatusForbidden)
					return
				case tenant.ErrTenantAccessDenied:
					logger.Warn().
						Str("user_id", user.ID().String()).
						Str("clerk_user_id", clerkUserID).
						Str("method", r.Method).
						Str("path", r.URL.Path).
						Msg("User denied access to requested tenant")
					http.Error(w, "You don't have access to this organization.", http.StatusForbidden)
					return
				default:
					logger.Error().
						Str("user_id", user.ID().String()).
						Str("clerk_user_id", clerkUserID).
						Str("method", r.Method).
						Str("path", r.URL.Path).
						Err(err).
						Msg("Failed to resolve tenant with validation")
					http.Error(w, "Failed to resolve tenant", http.StatusInternalServerError)
					return
				}
			}

			// Log tenant resolution result
			logEvent := logger.Info().
				Str("user_id", user.ID().String()).
				Str("clerk_user_id", clerkUserID).
				Str("resolved_tenant_id", result.TenantID).
				Str("resolution_source", string(result.Source)).
				Bool("validated", result.Validated).
				Bool("fallback_used", result.FallbackUsed).
				Str("method", r.Method).
				Str("path", r.URL.Path)

			if result.FallbackUsed {
				logEvent.Msg("Tenant resolved with fallback (invalid access attempt detected)")
			} else {
				logEvent.Msg("Tenant resolved successfully")
			}

			// Set tenant context in Go context
			ctx := tenantResolver.SetTenantContext(r.Context(), result.TenantID)
			r = r.WithContext(ctx)

			// Resolve client (optional - from query param or header)
			clientID := r.URL.Query().Get("client_id")
			if clientID == "" {
				clientID = r.Header.Get("X-Client-ID")
			}

			if clientID != "" {
				ctx = tenantResolver.SetClientContext(ctx, clientID)
				r = r.WithContext(ctx)
			}

			// Set RLS context in database session
			conn, err := db.Acquire(r.Context())
			if err != nil {
				logger.Error().Err(err).Msg("Failed to acquire database connection for RLS context")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			defer conn.Release()

			// Set tenant_id for RLS
			_, err = conn.Exec(r.Context(), "SELECT set_config('lv.tenant_id', $1, true)", result.TenantID)
			if err != nil {
				logger.Error().
					Str("tenant_id", result.TenantID).
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
				`, clientID, result.TenantID).Scan(&validClientID)
				if err != nil {
					logger.Warn().
						Str("client_id", clientID).
						Str("tenant_id", result.TenantID).
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
