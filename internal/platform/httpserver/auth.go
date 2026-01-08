package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/rs/zerolog"
)

// RequireAuth middleware that validates Clerk JWT tokens via JWKS
type RequireAuth struct {
	jwksURL string
	cache   *jwk.Cache
	logger  zerolog.Logger
}

// NewRequireAuth creates a new Clerk authentication middleware with JWKS verification
func NewRequireAuth(jwksURL string, logger zerolog.Logger) (*RequireAuth, error) {
	if jwksURL == "" {
		return nil, fmt.Errorf("CLERK_JWKS_URL is required")
	}

	cache := jwk.NewCache(context.Background())

	// Register the JWKS URL with the cache
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Register the URL first
	if err := cache.Register(jwksURL); err != nil {
		return nil, fmt.Errorf("failed to register JWKS URL: %w", err)
	}

	// Then fetch initial JWKS
	_, err := cache.Refresh(ctx, jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch initial JWKS: %w", err)
	}

	return &RequireAuth{
		jwksURL: jwksURL,
		cache:   cache,
		logger:  logger,
	}, nil
}

// RequireAuth middleware that validates Clerk JWT tokens
func (ra *RequireAuth) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			ra.logger.Warn().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Msg("401 Unauthorized: Missing Authorization header")
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Check if it's a Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			authPrefix := authHeader
			if len(authPrefix) > 20 {
				authPrefix = authPrefix[:20] + "..."
			}
			ra.logger.Warn().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Str("auth_header_prefix", authPrefix).
				Msg("401 Unauthorized: Invalid authorization header format (not Bearer token)")
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Get the JWKS set from cache
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		keySet, err := ra.cache.Get(ctx, ra.jwksURL)
		if err != nil {
			ra.logger.Debug().Err(err).Msg("Failed to get JWKS from cache")
			// Try to refresh the cache
			_, refreshErr := ra.cache.Refresh(ctx, ra.jwksURL)
			if refreshErr != nil {
				ra.logger.Error().
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Str("remote_addr", r.RemoteAddr).
					Err(refreshErr).
					Msg("Failed to refresh JWKS cache")
			} else {
				// Retry getting the key set
				keySet, err = ra.cache.Get(ctx, ra.jwksURL)
			}

			if err != nil {
				ra.logger.Error().
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Str("remote_addr", r.RemoteAddr).
					Err(err).
					Msg("401 Unauthorized: Failed to get JWKS after refresh")
				http.Error(w, "Failed to verify token", http.StatusUnauthorized)
				return
			}
		}

		// Verify token with the key set
		verifiedToken, err := jwt.Parse(
			[]byte(tokenString),
			jwt.WithKeySet(keySet),
			jwt.WithValidate(true),
		)
		if err != nil {
			ra.logger.Warn().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Err(err).
				Msg("401 Unauthorized: Token verification failed (invalid, expired, or malformed token)")
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Extract claims
		userID, _ := verifiedToken.Get("sub")
		email, _ := verifiedToken.Get("email")
		
		// Extract user profile information (Clerk may include these in JWT)
		firstName, _ := verifiedToken.Get("firstName")
		lastName, _ := verifiedToken.Get("lastName")
		// Fallback to snake_case if camelCase not found
		if firstName == nil {
			firstName, _ = verifiedToken.Get("first_name")
		}
		if lastName == nil {
			lastName, _ = verifiedToken.Get("last_name")
		}
		// Try full name claim
		fullName, _ := verifiedToken.Get("name")
		// Extract created_at or use iat (issued at) as fallback
		createdAt, _ := verifiedToken.Get("created_at")
		if createdAt == nil {
			createdAt, _ = verifiedToken.Get("createdAt")
		}
		iat, _ := verifiedToken.Get("iat") // Issued at timestamp

		allClaims, _ := verifiedToken.AsMap(ctx)

		// Clerk uses a nested "o" (organization) claim in session tokens
		// Structure: o.id, o.slg (slug), o.rol (role), o.per (permissions), o.fpm (feature-permission map)
		var orgID, orgSlug, orgRole interface{}

		// Try to get organization from nested "o" claim (Clerk's standard format)
		orgClaim, exists := verifiedToken.Get("o")
		logEvent := ra.logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Interface("user_id", userID).
			Interface("email", email).
			Interface("first_name", firstName).
			Interface("last_name", lastName).
			Interface("name", fullName).
			Interface("created_at", createdAt).
			Interface("iat", iat).
			Bool("org_claim_exists", exists).
			Interface("org_claim_value", orgClaim)

		if exists && orgClaim != nil {
			logEvent = logEvent.Str("org_claim_type", fmt.Sprintf("%T", orgClaim))
			if orgMap, ok := orgClaim.(map[string]interface{}); ok {
				orgID = orgMap["id"]
				orgSlug = orgMap["slg"] // Clerk uses "slg" for slug
				orgRole = orgMap["rol"] // Clerk uses "rol" for role (without "org:" prefix)
				logEvent = logEvent.
					Strs("org_map_keys", getMapKeys(orgMap)).
					Interface("org_id_from_map", orgID).
					Interface("org_slug_from_map", orgSlug).
					Interface("org_role_from_map", orgRole)
			} else {
				logEvent = logEvent.Str("org_claim_not_map", "org claim exists but is not a map")
			}
		} else {
			logEvent = logEvent.Str("org_claim_missing", "organization claim 'o' not found in token")
		}

		// Fallback: Try flat claims (for custom tokens or backward compatibility)
		if orgID == nil {
			orgID, _ = verifiedToken.Get("org_id")
			if orgID != nil {
				logEvent = logEvent.Str("org_id_source", "flat_claim")
			}
		}
		if orgSlug == nil {
			orgSlug, _ = verifiedToken.Get("org_slug")
			if orgSlug != nil {
				logEvent = logEvent.Str("org_slug_source", "flat_claim")
			}
		}
		if orgRole == nil {
			orgRole, _ = verifiedToken.Get("org_role")
			if orgRole != nil {
				logEvent = logEvent.Str("org_role_source", "flat_claim")
			}
		}

		// Log all available claim keys for debugging
		claimKeys := make([]string, 0, len(allClaims))
		for k := range allClaims {
			claimKeys = append(claimKeys, k)
		}
		logEvent = logEvent.Strs("available_claim_keys", claimKeys)

		// Final extracted values
		logEvent = logEvent.
			Interface("org_id", orgID).
			Interface("org_slug", orgSlug).
			Interface("org_role", orgRole)

		if orgID == nil {
			logEvent.Msg("⚠️  WARNING: org_id is null - User may not be part of any Clerk organization")
		} else {
			logEvent.Msg("Authentication successful: Token verified and claims extracted")
		}

		// Add user info to context
		ctx = r.Context()
		if userID != nil {
			ctx = context.WithValue(ctx, "user_id", userID)
		}
		if email != nil {
			ctx = context.WithValue(ctx, "email", email)
		}
		if firstName != nil {
			ctx = context.WithValue(ctx, "first_name", firstName)
		}
		if lastName != nil {
			ctx = context.WithValue(ctx, "last_name", lastName)
		}
		if fullName != nil {
			ctx = context.WithValue(ctx, "name", fullName)
		}
		if createdAt != nil {
			ctx = context.WithValue(ctx, "created_at", createdAt)
		} else if iat != nil {
			// Use issued at as fallback for created_at if available
			ctx = context.WithValue(ctx, "created_at", iat)
		}
		if orgID != nil {
			ctx = context.WithValue(ctx, "org_id", orgID)
		}
		if orgSlug != nil {
			ctx = context.WithValue(ctx, "org_slug", orgSlug)
		}
		if orgRole != nil {
			ctx = context.WithValue(ctx, "org_role", orgRole)
		}
		ctx = context.WithValue(ctx, "clerk_token", verifiedToken)

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// getMapKeys returns all keys from a map for logging purposes
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
