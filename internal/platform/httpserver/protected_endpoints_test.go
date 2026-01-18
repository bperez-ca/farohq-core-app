package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProtectedEndpoints_AuthRequired tests that protected endpoints require authentication
func TestProtectedEndpoints_AuthRequired(t *testing.T) {
	keyPair, err := GenerateTestKeyPair()
	require.NoError(t, err)

	auth, jwksServer, err := CreateTestRequireAuth(keyPair)
	require.NoError(t, err)
	defer jwksServer.Close()

	// Create a mock protected endpoint handler
	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"success"}`))
	})

	// Wrap with auth middleware
	handler := auth.RequireAuth(protectedHandler)

	// Test without authentication
	req := MakeUnauthenticatedRequest("GET", "/api/v1/protected")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Authorization header required")
}

// TestProtectedEndpoints_ValidToken tests that protected endpoints work with valid tokens
func TestProtectedEndpoints_ValidToken(t *testing.T) {
	keyPair, err := GenerateTestKeyPair()
	require.NoError(t, err)

	auth, jwksServer, err := CreateTestRequireAuth(keyPair)
	require.NoError(t, err)
	defer jwksServer.Close()

	claims := map[string]interface{}{
		"sub":   "test-user",
		"email": "test@example.com",
	}
	token, err := CreateMockJWT(keyPair, claims)
	require.NoError(t, err)

	// Create a mock protected endpoint handler
	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := ctx.Value("user_id")
		if userID == nil {
			http.Error(w, "user_id not in context", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"user_id":"` + userID.(string) + `"}`))
	})

	// Wrap with auth middleware
	handler := auth.RequireAuth(protectedHandler)

	// Test with valid token from Authorization header
	req := MakeAuthenticatedRequest("GET", "/api/v1/protected", token, TokenSourceAuthorization)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "test-user")
}

// TestProtectedEndpoints_TenantEndpoints tests tenant-scoped endpoints
func TestProtectedEndpoints_TenantEndpoints(t *testing.T) {
	keyPair, err := GenerateTestKeyPair()
	require.NoError(t, err)

	auth, jwksServer, err := CreateTestRequireAuth(keyPair)
	require.NoError(t, err)
	defer jwksServer.Close()

	claims := map[string]interface{}{
		"sub":    "tenant-user",
		"email":  "tenant@example.com",
		"org_id": "tenant-123",
	}
	token, err := CreateMockJWT(keyPair, claims)
	require.NoError(t, err)

	// Mock tenant endpoint handler
	tenantHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := ctx.Value("user_id")
		orgID := ctx.Value("org_id")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"user_id":"` + userID.(string) + `","org_id":"` + orgID.(string) + `"}`))
	})

	handler := auth.RequireAuth(tenantHandler)

	// Test GET /api/v1/tenants/{id}
	req := MakeAuthenticatedRequest("GET", "/api/v1/tenants/tenant-123", token, TokenSourceAuthorization)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "tenant-user")
	assert.Contains(t, rr.Body.String(), "tenant-123")
}

// TestProtectedEndpoints_BrandEndpoints tests brand endpoints
func TestProtectedEndpoints_BrandEndpoints(t *testing.T) {
	keyPair, err := GenerateTestKeyPair()
	require.NoError(t, err)

	auth, jwksServer, err := CreateTestRequireAuth(keyPair)
	require.NoError(t, err)
	defer jwksServer.Close()

	claims := map[string]interface{}{
		"sub":   "brand-user",
		"email": "brand@example.com",
	}
	token, err := CreateMockJWT(keyPair, claims)
	require.NoError(t, err)

	// Mock brand endpoint handler
	brandHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := ctx.Value("user_id")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"user_id":"` + userID.(string) + `","brands":[]}`))
	})

	handler := auth.RequireAuth(brandHandler)

	// Test GET /api/v1/brands
	req := MakeAuthenticatedRequest("GET", "/api/v1/brands", token, TokenSourceAuthorization)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "brand-user")
}

// TestProtectedEndpoints_FileEndpoints tests file endpoints
func TestProtectedEndpoints_FileEndpoints(t *testing.T) {
	keyPair, err := GenerateTestKeyPair()
	require.NoError(t, err)

	auth, jwksServer, err := CreateTestRequireAuth(keyPair)
	require.NoError(t, err)
	defer jwksServer.Close()

	claims := map[string]interface{}{
		"sub":   "file-user",
		"email": "file@example.com",
	}
	token, err := CreateMockJWT(keyPair, claims)
	require.NoError(t, err)

	// Mock file endpoint handler
	fileHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := ctx.Value("user_id")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"user_id":"` + userID.(string) + `","files":[]}`))
	})

	handler := auth.RequireAuth(fileHandler)

	// Test GET /api/v1/files
	req := MakeAuthenticatedRequest("GET", "/api/v1/files", token, TokenSourceAuthorization)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "file-user")
}

// TestProtectedEndpoints_AuthMeEndpoint tests /api/v1/auth/me endpoint
func TestProtectedEndpoints_AuthMeEndpoint(t *testing.T) {
	keyPair, err := GenerateTestKeyPair()
	require.NoError(t, err)

	auth, jwksServer, err := CreateTestRequireAuth(keyPair)
	require.NoError(t, err)
	defer jwksServer.Close()

	claims := map[string]interface{}{
		"sub":   "me-user",
		"email": "me@example.com",
	}
	token, err := CreateMockJWT(keyPair, claims)
	require.NoError(t, err)

	// Mock /auth/me handler
	authMeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := ctx.Value("user_id")
		email := ctx.Value("email")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"user_id":"` + userID.(string) + `","email":"` + email.(string) + `"}`))
	})

	handler := auth.RequireAuth(authMeHandler)

	// Test GET /api/v1/auth/me
	req := MakeAuthenticatedRequest("GET", "/api/v1/auth/me", token, TokenSourceAuthorization)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "me-user")
	assert.Contains(t, rr.Body.String(), "me@example.com")
}

// TestProtectedEndpoints_UserSyncEndpoint tests /api/v1/users/sync endpoint
func TestProtectedEndpoints_UserSyncEndpoint(t *testing.T) {
	keyPair, err := GenerateTestKeyPair()
	require.NoError(t, err)

	auth, jwksServer, err := CreateTestRequireAuth(keyPair)
	require.NoError(t, err)
	defer jwksServer.Close()

	claims := map[string]interface{}{
		"sub":   "sync-user",
		"email": "sync@example.com",
	}
	token, err := CreateMockJWT(keyPair, claims)
	require.NoError(t, err)

	// Mock /users/sync handler
	userSyncHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := ctx.Value("user_id")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"user_id":"` + userID.(string) + `","synced":true}`))
	})

	handler := auth.RequireAuth(userSyncHandler)

	// Test POST /api/v1/users/sync
	req := MakeAuthenticatedRequest("POST", "/api/v1/users/sync", token, TokenSourceAuthorization)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "sync-user")
}

// TestProtectedEndpoints_AllTokenSources tests that all token sources work for protected endpoints
func TestProtectedEndpoints_AllTokenSources(t *testing.T) {
	keyPair, err := GenerateTestKeyPair()
	require.NoError(t, err)

	auth, jwksServer, err := CreateTestRequireAuth(keyPair)
	require.NoError(t, err)
	defer jwksServer.Close()

	claims := map[string]interface{}{
		"sub": "multi-source-user",
	}
	token, err := CreateMockJWT(keyPair, claims)
	require.NoError(t, err)

	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	handler := auth.RequireAuth(protectedHandler)

	sources := []TokenSource{
		TokenSourceAuthorization,
		TokenSourceClerkAuthToken,
		TokenSourceXAuthToken,
	}

	for _, source := range sources {
		t.Run(string(source), func(t *testing.T) {
			req := MakeAuthenticatedRequest("GET", "/api/v1/protected", token, source)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			assert.Contains(t, rr.Body.String(), "ok")
		})
	}
}

// TestProtectedEndpoints_InvalidToken tests that protected endpoints reject invalid tokens
func TestProtectedEndpoints_InvalidToken(t *testing.T) {
	keyPair, err := GenerateTestKeyPair()
	require.NoError(t, err)

	auth, jwksServer, err := CreateTestRequireAuth(keyPair)
	require.NoError(t, err)
	defer jwksServer.Close()

	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := auth.RequireAuth(protectedHandler)

	tests := []struct {
		name        string
		token       string
		tokenSource TokenSource
	}{
		{
			name:        "malformed token",
			token:       "invalid.jwt.token",
			tokenSource: TokenSourceAuthorization,
		},
		{
			name:        "expired token",
			token: func() string {
				claims := map[string]interface{}{"sub": "user"}
				token, _ := CreateExpiredJWT(keyPair, claims)
				return token
			}(),
			tokenSource: TokenSourceAuthorization,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := MakeAuthenticatedRequest("GET", "/api/v1/protected", tt.token, tt.tokenSource)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusUnauthorized, rr.Code)
			assert.Contains(t, rr.Body.String(), "Invalid token")
		})
	}
}

// TestProtectedEndpoints_ContextPropagation tests that context values are properly propagated to handlers
func TestProtectedEndpoints_ContextPropagation(t *testing.T) {
	keyPair, err := GenerateTestKeyPair()
	require.NoError(t, err)

	auth, jwksServer, err := CreateTestRequireAuth(keyPair)
	require.NoError(t, err)
	defer jwksServer.Close()

	claims := map[string]interface{}{
		"sub":     "context-user",
		"email":   "context@example.com",
		"org_id":  "context-org",
		"org_slug": "context-slug",
		"org_role": "admin",
	}
	token, err := CreateMockJWT(keyPair, claims)
	require.NoError(t, err)

	// Handler that checks all context values
	contextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID := ctx.Value("user_id")
		email := ctx.Value("email")
		orgID := ctx.Value("org_id")
		orgSlug := ctx.Value("org_slug")
		orgRole := ctx.Value("org_role")

		// Verify all values are present
		assert.NotNil(t, userID)
		assert.NotNil(t, email)
		assert.NotNil(t, orgID)
		assert.NotNil(t, orgSlug)
		assert.NotNil(t, orgRole)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"context":"ok"}`))
	})

	handler := auth.RequireAuth(contextHandler)

	req := MakeAuthenticatedRequest("GET", "/api/v1/test", token, TokenSourceAuthorization)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
