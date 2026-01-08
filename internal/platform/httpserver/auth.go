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
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Check if it's a Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
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
				ra.logger.Error().Err(refreshErr).Msg("Failed to refresh JWKS cache")
			} else {
				// Retry getting the key set
				keySet, err = ra.cache.Get(ctx, ra.jwksURL)
			}

			if err != nil {
				ra.logger.Debug().Err(err).Msg("Failed to get JWKS after refresh")
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
			ra.logger.Debug().Err(err).Msg("Token verification failed")
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Extract claims
		userID, _ := verifiedToken.Get("sub")
		email, _ := verifiedToken.Get("email")
		orgID, _ := verifiedToken.Get("org_id")
		orgSlug, _ := verifiedToken.Get("org_slug")
		orgRole, _ := verifiedToken.Get("org_role")

		// Add user info to context
		ctx = r.Context()
		if userID != nil {
			ctx = context.WithValue(ctx, "user_id", userID)
		}
		if email != nil {
			ctx = context.WithValue(ctx, "email", email)
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
