package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractTokenFromRequest(t *testing.T) {
	keyPair, err := GenerateTestKeyPair()
	require.NoError(t, err)

	auth, jwksServer, err := CreateTestRequireAuth(keyPair)
	require.NoError(t, err)
	defer jwksServer.Close()

	tests := []struct {
		name           string
		headers        map[string]string
		expectedToken  string
		expectedSource TokenSource
		expectedFound  bool
	}{
		{
			name: "token from Authorization header",
			headers: map[string]string{
				"Authorization": "Bearer test-token-123",
			},
			expectedToken:  "test-token-123",
			expectedSource: TokenSourceAuthorization,
			expectedFound: true,
		},
		{
			name: "token from x-clerk-auth-token header",
			headers: map[string]string{
				"x-clerk-auth-token": "clerk-token-456",
			},
			expectedToken:  "clerk-token-456",
			expectedSource: TokenSourceClerkAuthToken,
			expectedFound:  true,
		},
		{
			name: "token from X-Auth-Token header",
			headers: map[string]string{
				"X-Auth-Token": "custom-token-789",
			},
			expectedToken:  "custom-token-789",
			expectedSource: TokenSourceXAuthToken,
			expectedFound:  true,
		},
		{
			name: "Authorization header takes priority over x-clerk-auth-token",
			headers: map[string]string{
				"Authorization":     "Bearer priority-token",
				"x-clerk-auth-token": "clerk-token",
			},
			expectedToken:  "priority-token",
			expectedSource: TokenSourceAuthorization,
			expectedFound:  true,
		},
		{
			name: "Authorization header takes priority over X-Auth-Token",
			headers: map[string]string{
				"Authorization": "Bearer priority-token",
				"X-Auth-Token":  "custom-token",
			},
			expectedToken:  "priority-token",
			expectedSource: TokenSourceAuthorization,
			expectedFound:  true,
		},
		{
			name: "x-clerk-auth-token takes priority over X-Auth-Token",
			headers: map[string]string{
				"x-clerk-auth-token": "clerk-token",
				"X-Auth-Token":       "custom-token",
			},
			expectedToken:  "clerk-token",
			expectedSource: TokenSourceClerkAuthToken,
			expectedFound:  true,
		},
		{
			name: "no token in any header",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			expectedFound: false,
		},
		{
			name: "Authorization header without Bearer prefix",
			headers: map[string]string{
				"Authorization": "InvalidFormat token",
			},
			expectedFound: false,
		},
		{
			name: "Authorization header with empty Bearer token",
			headers: map[string]string{
				"Authorization": "Bearer ",
			},
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateTestRequest("GET", "/test", tt.headers)
			token, source, found := auth.extractTokenFromRequest(req)

			assert.Equal(t, tt.expectedFound, found)
			if tt.expectedFound {
				assert.Equal(t, tt.expectedToken, token)
				assert.Equal(t, tt.expectedSource, source)
			}
		})
	}
}

func TestRequireAuth_MissingToken(t *testing.T) {
	keyPair, err := GenerateTestKeyPair()
	require.NoError(t, err)

	auth, jwksServer, err := CreateTestRequireAuth(keyPair)
	require.NoError(t, err)
	defer jwksServer.Close()

	handler := auth.RequireAuth(CreateTestHandler())
	req := MakeUnauthenticatedRequest("GET", "/test")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Authorization header required")
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	keyPair, err := GenerateTestKeyPair()
	require.NoError(t, err)

	auth, jwksServer, err := CreateTestRequireAuth(keyPair)
	require.NoError(t, err)
	defer jwksServer.Close()

	tests := []struct {
		name        string
		token       string
		tokenSource TokenSource
	}{
		{
			name:        "malformed token",
			token:       "not-a-valid-jwt",
			tokenSource: TokenSourceAuthorization,
		},
		{
			name:        "token signed with different key",
			token:       "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0LXVzZXIiLCJpYXQiOjE2MDAwMDAwMDAsImV4cCI6OTk5OTk5OTk5OX0.invalid-signature",
			tokenSource: TokenSourceAuthorization,
		},
		{
			name:        "token from x-clerk-auth-token with invalid format",
			token:       "invalid-token",
			tokenSource: TokenSourceClerkAuthToken,
		},
		{
			name:        "token from X-Auth-Token with invalid format",
			token:       "invalid-token",
			tokenSource: TokenSourceXAuthToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := auth.RequireAuth(CreateTestHandler())
			req := MakeAuthenticatedRequest("GET", "/test", tt.token, tt.tokenSource)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusUnauthorized, rr.Code)
			assert.Contains(t, rr.Body.String(), "Invalid token")
		})
	}
}

func TestRequireAuth_ExpiredToken(t *testing.T) {
	keyPair, err := GenerateTestKeyPair()
	require.NoError(t, err)

	auth, jwksServer, err := CreateTestRequireAuth(keyPair)
	require.NoError(t, err)
	defer jwksServer.Close()

	// Create expired token
	claims := map[string]interface{}{
		"sub": "test-user",
		"email": "test@example.com",
	}
	expiredToken, err := CreateExpiredJWT(keyPair, claims)
	require.NoError(t, err)

	handler := auth.RequireAuth(CreateTestHandler())
	req := MakeAuthenticatedRequest("GET", "/test", expiredToken, TokenSourceAuthorization)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid token")
}

func TestRequireAuth_ValidToken(t *testing.T) {
	keyPair, err := GenerateTestKeyPair()
	require.NoError(t, err)

	auth, jwksServer, err := CreateTestRequireAuth(keyPair)
	require.NoError(t, err)
	defer jwksServer.Close()

	tests := []struct {
		name        string
		tokenSource TokenSource
		claims      map[string]interface{}
	}{
		{
			name:        "valid token from Authorization header",
			tokenSource: TokenSourceAuthorization,
			claims: map[string]interface{}{
				"sub":   "user-123",
				"email": "user@example.com",
			},
		},
		{
			name:        "valid token from x-clerk-auth-token header",
			tokenSource: TokenSourceClerkAuthToken,
			claims: map[string]interface{}{
				"sub":   "user-456",
				"email": "user2@example.com",
			},
		},
		{
			name:        "valid token from X-Auth-Token header",
			tokenSource: TokenSourceXAuthToken,
			claims: map[string]interface{}{
				"sub":   "user-789",
				"email": "user3@example.com",
			},
		},
		{
			name:        "valid token with org claims",
			tokenSource: TokenSourceAuthorization,
			claims: map[string]interface{}{
				"sub":    "user-org",
				"email":  "org@example.com",
				"org_id": "org-123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create valid token
			token, err := CreateMockJWT(keyPair, tt.claims)
			require.NoError(t, err)

			handler := auth.RequireAuth(CreateTestHandler())
			req := MakeAuthenticatedRequest("GET", "/test", token, tt.tokenSource)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.claims["sub"].(string))
		})
	}
}

func TestRequireAuth_ContextValues(t *testing.T) {
	keyPair, err := GenerateTestKeyPair()
	require.NoError(t, err)

	auth, jwksServer, err := CreateTestRequireAuth(keyPair)
	require.NoError(t, err)
	defer jwksServer.Close()

	claims := map[string]interface{}{
		"sub":   "test-user-id",
		"email": "test@example.com",
		"org_id": "org-123",
	}
	token, err := CreateMockJWT(keyPair, claims)
	require.NoError(t, err)

	// Create handler that checks context values
	handler := auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := ctx.Value("user_id")
		email := ctx.Value("email")
		orgID := ctx.Value("org_id")

		assert.Equal(t, "test-user-id", userID)
		assert.Equal(t, "test@example.com", email)
		assert.Equal(t, "org-123", orgID)
	}))

	req := MakeAuthenticatedRequest("GET", "/test", token, TokenSourceAuthorization)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestNewRequireAuth_InvalidJWKSURL(t *testing.T) {
	logger := zerolog.Nop()
	auth, err := NewRequireAuth("", logger)
	assert.Error(t, err)
	assert.Nil(t, auth)
	assert.Contains(t, err.Error(), "CLERK_JWKS_URL is required")
}

func TestNewRequireAuth_InvalidJWKSEndpoint(t *testing.T) {
	logger := zerolog.Nop()
	// Use a non-existent URL
	auth, err := NewRequireAuth("http://localhost:99999/.well-known/jwks.json", logger)
	assert.Error(t, err)
	assert.Nil(t, auth)
}

func TestRequireAuth_EmptyTokenString(t *testing.T) {
	keyPair, err := GenerateTestKeyPair()
	require.NoError(t, err)

	auth, jwksServer, err := CreateTestRequireAuth(keyPair)
	require.NoError(t, err)
	defer jwksServer.Close()

	// Test with empty token in Authorization header
	handler := auth.RequireAuth(CreateTestHandler())
	req := CreateTestRequest("GET", "/test", map[string]string{
		"Authorization": "Bearer ",
	})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid token")
}

func TestRequireAuth_JWKSCacheRefresh(t *testing.T) {
	keyPair, err := GenerateTestKeyPair()
	require.NoError(t, err)

	auth, jwksServer, err := CreateTestRequireAuth(keyPair)
	require.NoError(t, err)
	defer jwksServer.Close()

	claims := map[string]interface{}{
		"sub": "test-user",
	}
	token, err := CreateMockJWT(keyPair, claims)
	require.NoError(t, err)

	handler := auth.RequireAuth(CreateTestHandler())
	req := MakeAuthenticatedRequest("GET", "/test", token, TokenSourceAuthorization)
	rr := httptest.NewRecorder()

	// First request should work
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Second request should use cache
	rr2 := httptest.NewRecorder()
	req2 := MakeAuthenticatedRequest("GET", "/test", token, TokenSourceAuthorization)
	handler.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusOK, rr2.Code)
}
