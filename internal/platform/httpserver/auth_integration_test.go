package httpserver

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRequireAuth_RealJWKS tests with a real JWKS endpoint (if CLERK_JWKS_URL is set)
// This test is skipped if CLERK_JWKS_URL is not set
func TestRequireAuth_RealJWKS(t *testing.T) {
	jwksURL := os.Getenv("CLERK_JWKS_URL")
	if jwksURL == "" {
		t.Skip("CLERK_JWKS_URL not set, skipping integration test")
	}

	logger := zerolog.Nop()
	auth, err := NewRequireAuth(jwksURL, logger)
	require.NoError(t, err)

	// This test requires a valid Clerk token
	// In a real scenario, you would get this from Clerk after authentication
	// For now, we'll just test that the middleware initializes correctly
	assert.NotNil(t, auth)
	assert.Equal(t, jwksURL, auth.jwksURL)
}

// TestRequireAuth_MiddlewareChain tests the full middleware chain
func TestRequireAuth_MiddlewareChain(t *testing.T) {
	keyPair, err := GenerateTestKeyPair()
	require.NoError(t, err)

	auth, jwksServer, err := CreateTestRequireAuth(keyPair)
	require.NoError(t, err)
	defer jwksServer.Close()

	// Create a chain: auth -> test handler
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := ctx.Value("user_id")
		if userID == nil {
			http.Error(w, "user_id not in context", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler := auth.RequireAuth(finalHandler)

	claims := map[string]interface{}{
		"sub": "test-user",
	}
	token, err := CreateMockJWT(keyPair, claims)
	require.NoError(t, err)

	req := MakeAuthenticatedRequest("GET", "/test", token, TokenSourceAuthorization)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "OK", rr.Body.String())
}

// TestRequireAuth_ErrorResponses tests various error response scenarios
func TestRequireAuth_ErrorResponses(t *testing.T) {
	keyPair, err := GenerateTestKeyPair()
	require.NoError(t, err)

	auth, jwksServer, err := CreateTestRequireAuth(keyPair)
	require.NoError(t, err)
	defer jwksServer.Close()

	handler := auth.RequireAuth(CreateTestHandler())

	tests := []struct {
		name           string
		request        *http.Request
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "missing token",
			request:        MakeUnauthenticatedRequest("GET", "/test"),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Authorization header required",
		},
		{
			name: "invalid token format",
			request: CreateTestRequest("GET", "/test", map[string]string{
				"Authorization": "InvalidFormat token",
			}),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Authorization header required",
		},
		{
			name: "malformed JWT",
			request: CreateTestRequest("GET", "/test", map[string]string{
				"Authorization": "Bearer not.a.valid.jwt",
			}),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid token",
		},
		{
			name: "expired token",
			request: func() *http.Request {
				claims := map[string]interface{}{
					"sub": "test-user",
				}
				expiredToken, _ := CreateExpiredJWT(keyPair, claims)
				return MakeAuthenticatedRequest("GET", "/test", expiredToken, TokenSourceAuthorization)
			}(),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, tt.request)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)
		})
	}
}

// TestRequireAuth_AllTokenSources tests that all token sources work correctly
func TestRequireAuth_AllTokenSources(t *testing.T) {
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

	sources := []TokenSource{
		TokenSourceAuthorization,
		TokenSourceClerkAuthToken,
		TokenSourceXAuthToken,
	}

	for _, source := range sources {
		t.Run(string(source), func(t *testing.T) {
			handler := auth.RequireAuth(CreateTestHandler())
			req := MakeAuthenticatedRequest("GET", "/test", token, source)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			assert.Contains(t, rr.Body.String(), "test-user")
		})
	}
}

// TestRequireAuth_ContextPropagation tests that context values are properly propagated
func TestRequireAuth_ContextPropagation(t *testing.T) {
	keyPair, err := GenerateTestKeyPair()
	require.NoError(t, err)

	auth, jwksServer, err := CreateTestRequireAuth(keyPair)
	require.NoError(t, err)
	defer jwksServer.Close()

	claims := map[string]interface{}{
		"sub":     "user-123",
		"email":   "user@example.com",
		"org_id":  "org-456",
		"org_slug": "test-org",
		"org_role": "admin",
	}
	token, err := CreateMockJWT(keyPair, claims)
	require.NoError(t, err)

	// Create handler that verifies all context values
	handler := auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID := ctx.Value("user_id")
		email := ctx.Value("email")
		orgID := ctx.Value("org_id")
		orgSlug := ctx.Value("org_slug")
		orgRole := ctx.Value("org_role")

		assert.Equal(t, "user-123", userID)
		assert.Equal(t, "user@example.com", email)
		assert.Equal(t, "org-456", orgID)
		assert.Equal(t, "test-org", orgSlug)
		assert.Equal(t, "admin", orgRole)

		w.WriteHeader(http.StatusOK)
	}))

	req := MakeAuthenticatedRequest("GET", "/test", token, TokenSourceAuthorization)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestRequireAuth_ConcurrentRequests tests concurrent request handling
func TestRequireAuth_ConcurrentRequests(t *testing.T) {
	keyPair, err := GenerateTestKeyPair()
	require.NoError(t, err)

	auth, jwksServer, err := CreateTestRequireAuth(keyPair)
	require.NoError(t, err)
	defer jwksServer.Close()

	claims := map[string]interface{}{
		"sub": "concurrent-user",
	}
	token, err := CreateMockJWT(keyPair, claims)
	require.NoError(t, err)

	handler := auth.RequireAuth(CreateTestHandler())

	// Make multiple concurrent requests
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			req := MakeAuthenticatedRequest("GET", "/test", token, TokenSourceAuthorization)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusOK, rr.Code)
			done <- true
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
