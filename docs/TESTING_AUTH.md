# Authentication Testing Guide

This guide explains how to test authentication in the FaroHQ Core App.

## Overview

The authentication system uses Clerk JWT tokens validated via JWKS (JSON Web Key Set). All protected endpoints require a valid token in one of the supported headers.

## Running Tests

### Unit Tests

Run all authentication unit tests:

```bash
go test ./internal/platform/httpserver -v -run TestExtractTokenFromRequest
go test ./internal/platform/httpserver -v -run TestRequireAuth
```

### Integration Tests

Run integration tests (requires `CLERK_JWKS_URL` environment variable):

```bash
CLERK_JWKS_URL=https://your-clerk-instance.clerk.accounts.dev/.well-known/jwks.json \
go test ./internal/platform/httpserver -v -run TestRequireAuth_RealJWKS
```

### Protected Endpoint Tests

Test all protected endpoints:

```bash
go test ./internal/platform/httpserver -v -run TestProtectedEndpoints
```

### All Authentication Tests

Run all authentication-related tests:

```bash
go test ./internal/platform/httpserver -v -run "Test.*Auth|TestProtected"
```

## Test Helpers

The test suite includes helpers in `auth_test_helpers.go`:

### Creating Test Tokens

```go
// Generate a test key pair
keyPair, err := GenerateTestKeyPair()
require.NoError(t, err)

// Create a JWT with custom claims
claims := map[string]interface{}{
    "sub":   "test-user",
    "email": "test@example.com",
    "org_id": "org-123",
}
token, err := CreateMockJWT(keyPair, claims)
require.NoError(t, err)

// Create an expired token
expiredToken, err := CreateExpiredJWT(keyPair, claims)
require.NoError(t, err)
```

### Creating Mock JWKS Server

```go
// Create a mock JWKS server
keyPair, err := GenerateTestKeyPair()
require.NoError(t, err)

jwksServer, err := CreateMockJWKSServer(keyPair)
require.NoError(t, err)
defer jwksServer.Close()

// Create RequireAuth with mock JWKS
auth, err := CreateTestRequireAuth(keyPair)
require.NoError(t, err)
defer jwksServer.Close()
```

### Creating Test Requests

```go
// Request with Authorization header
req := MakeAuthenticatedRequest("GET", "/api/v1/test", token, TokenSourceAuthorization)

// Request with x-clerk-auth-token header
req := MakeAuthenticatedRequest("GET", "/api/v1/test", token, TokenSourceClerkAuthToken)

// Request with X-Auth-Token header
req := MakeAuthenticatedRequest("GET", "/api/v1/test", token, TokenSourceXAuthToken)

// Request without authentication
req := MakeUnauthenticatedRequest("GET", "/api/v1/test")
```

### Asserting Auth Errors

```go
handler := auth.RequireAuth(testHandler)
req := MakeUnauthenticatedRequest("GET", "/test")
rr := httptest.NewRecorder()
handler.ServeHTTP(rr, req)

AssertAuthError(t, rr.Result(), "Authorization header required")
```

## Test Scenarios

### Token Extraction Tests

Test that tokens are extracted from all supported headers:

```go
func TestTokenExtraction(t *testing.T) {
    // Test Authorization header
    // Test x-clerk-auth-token header
    // Test X-Auth-Token header
    // Test priority order
    // Test missing token
}
```

### Token Validation Tests

Test token validation scenarios:

```go
func TestTokenValidation(t *testing.T) {
    // Test valid token
    // Test expired token
    // Test malformed token
    // Test token with wrong signature
    // Test empty token
}
```

### Context Propagation Tests

Test that user context is properly propagated:

```go
func TestContextPropagation(t *testing.T) {
    // Test user_id in context
    // Test email in context
    // Test org_id in context
    // Test org_slug in context
    // Test org_role in context
}
```

### Protected Endpoint Tests

Test that protected endpoints require authentication:

```go
func TestProtectedEndpoints(t *testing.T) {
    // Test /api/v1/tenants/*
    // Test /api/v1/brands/*
    // Test /api/v1/files/*
    // Test /api/v1/auth/me
    // Test /api/v1/users/sync
}
```

## Common Test Patterns

### Testing Success Case

```go
func TestAuthSuccess(t *testing.T) {
    keyPair, _ := GenerateTestKeyPair()
    auth, jwksServer, _ := CreateTestRequireAuth(keyPair)
    defer jwksServer.Close()

    claims := map[string]interface{}{"sub": "user-123"}
    token, _ := CreateMockJWT(keyPair, claims)

    handler := auth.RequireAuth(CreateTestHandler())
    req := MakeAuthenticatedRequest("GET", "/test", token, TokenSourceAuthorization)
    rr := httptest.NewRecorder()

    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)
}
```

### Testing Failure Case

```go
func TestAuthFailure(t *testing.T) {
    keyPair, _ := GenerateTestKeyPair()
    auth, jwksServer, _ := CreateTestRequireAuth(keyPair)
    defer jwksServer.Close()

    handler := auth.RequireAuth(CreateTestHandler())
    req := MakeUnauthenticatedRequest("GET", "/test")
    rr := httptest.NewRecorder()

    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusUnauthorized, rr.Code)
    assert.Contains(t, rr.Body.String(), "Authorization header required")
}
```

### Testing All Token Sources

```go
func TestAllTokenSources(t *testing.T) {
    keyPair, _ := GenerateTestKeyPair()
    auth, jwksServer, _ := CreateTestRequireAuth(keyPair)
    defer jwksServer.Close()

    token, _ := CreateMockJWT(keyPair, map[string]interface{}{"sub": "user"})

    sources := []TokenSource{
        TokenSourceAuthorization,
        TokenSourceClerkAuthToken,
        TokenSourceXAuthToken,
    }

    for _, source := range sources {
        t.Run(string(source), func(t *testing.T) {
            req := MakeAuthenticatedRequest("GET", "/test", token, source)
            rr := httptest.NewRecorder()
            handler.ServeHTTP(rr, req)
            assert.Equal(t, http.StatusOK, rr.Code)
        })
    }
}
```

## Test Coverage Goals

- **90%+ coverage** for auth middleware
- **All error paths** tested
- **All token sources** tested
- **All protected endpoints** have at least one test

## Running Coverage

Generate coverage report:

```bash
go test ./internal/platform/httpserver -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Debugging Tests

### Enable Debug Logging

Set log level to debug in tests:

```go
logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)
auth, err := NewRequireAuth(jwksURL, logger)
```

### Inspect Token Claims

```go
// Decode token to inspect claims
parts := strings.Split(token, ".")
payload, _ := base64.RawURLEncoding.DecodeString(parts[1])
var claims map[string]interface{}
json.Unmarshal(payload, &claims)
fmt.Printf("Claims: %+v\n", claims)
```

### Check JWKS Response

```go
// Test JWKS endpoint directly
resp, err := http.Get(jwksServer.URL)
body, _ := io.ReadAll(resp.Body)
fmt.Printf("JWKS: %s\n", string(body))
```

## Integration with Real Clerk

For integration tests with real Clerk:

1. Set `CLERK_JWKS_URL` environment variable
2. Get a real token from Clerk (via sign-in flow)
3. Use the real token in tests

```bash
export CLERK_JWKS_URL=https://your-instance.clerk.accounts.dev/.well-known/jwks.json
go test ./internal/platform/httpserver -v -run TestRequireAuth_RealJWKS
```

## Best Practices

1. **Use test helpers**: Always use the provided test helpers instead of creating tokens manually
2. **Clean up resources**: Always close JWKS servers and clean up test resources
3. **Test all token sources**: Ensure tests cover all three token header formats
4. **Test error cases**: Don't just test success cases - test all failure scenarios
5. **Isolate tests**: Each test should be independent and not rely on other tests
6. **Use table-driven tests**: For multiple similar test cases, use table-driven tests
7. **Check context values**: Verify that context values are properly set after authentication

## Troubleshooting

### Tests Fail with "JWKS unavailable"

- Check that the mock JWKS server is running
- Verify the JWKS URL is correct
- Check network connectivity (for real JWKS tests)

### Tests Fail with "Invalid token"

- Verify the token is signed with the correct key
- Check token expiration (use `CreateExpiredJWT` for expired token tests)
- Ensure token format is correct (valid JWT)

### Tests Fail with "Token not found"

- Check that the request includes the correct header
- Verify the header name matches the token source
- Ensure the token string is not empty
