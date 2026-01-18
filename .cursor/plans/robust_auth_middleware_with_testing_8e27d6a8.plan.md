---
name: Robust Auth Middleware with Testing
overview: Implement a centralized, bulletproof authentication middleware that supports multiple token header sources, enhanced logging, and comprehensive test coverage for all authenticated endpoints.
todos:
  - id: enhance-auth-middleware
    content: Enhance auth middleware with multi-source token extraction (Authorization, x-clerk-auth-token, X-Auth-Token) and improved error handling
    status: completed
  - id: add-structured-logging
    content: Add structured logging throughout auth flow with consistent fields (method, path, token_source, auth_result, etc.)
    status: completed
    dependencies:
      - enhance-auth-middleware
  - id: create-test-helpers
    content: Create test helpers for mock JWT generation, JWKS server, and test request creation
    status: completed
  - id: write-unit-tests
    content: Write comprehensive unit tests for token extraction, validation, and error handling
    status: completed
    dependencies:
      - create-test-helpers
      - enhance-auth-middleware
  - id: write-integration-tests
    content: Write integration tests with real JWKS endpoint and full middleware chain
    status: completed
    dependencies:
      - write-unit-tests
  - id: test-protected-endpoints
    content: Create test suite for all protected endpoints (tenants, brands, files, auth, users)
    status: completed
    dependencies:
      - write-integration-tests
  - id: update-documentation
    content: Update API_AUTHENTICATION.md and create TESTING_AUTH.md with new token sources and testing guide
    status: completed
    dependencies:
      - test-protected-endpoints
---

# Robust Authentication Middleware Implementation Plan

## Overview

This plan implements Option 3: a centralized authentication middleware that supports multiple token sources, provides detailed logging, and includes comprehensive test coverage for all authenticated endpoints.

## Architecture

The authentication flow will support multiple token sources with priority order:

```
1. Authorization: Bearer <token> (standard)
2. x-clerk-auth-token (Clerk's automatic header)
3. X-Auth-Token (custom fallback)
```

## Implementation Steps

### 1. Enhance Auth Middleware (`internal/platform/httpserver/auth.go`)

**1.1 Add Token Extraction Helper**

- Create `extractTokenFromRequest()` method that checks headers in priority order
- Return both token string and source header name for logging
- Log which header was used for debugging

**1.2 Improve Error Handling**

- Distinguish between: missing token, invalid format, expired token, JWKS failure
- Return specific error messages for each case
- Add structured logging with request context (method, path, remote_addr, checked headers)

**1.3 Enhanced Logging**

- Log token extraction source (which header was used)
- Log all checked headers when token is missing
- Add debug logs for JWKS cache hits/misses
- Log token validation steps (extraction → JWKS lookup → verification → claims extraction)
- Include timing information for performance monitoring

**1.4 Token Validation Improvements**

- Add validation result logging (success/failure with reason)
- Log extracted claims structure for debugging
- Add metrics-ready log events (auth_success, auth_failure with reason)

### 2. Create Test Infrastructure

**2.1 Test Helpers (`internal/platform/httpserver/auth_test_helpers.go`)**

- `createMockJWT()` - Generate test JWT tokens with custom claims
- `createMockJWKS()` - Create mock JWKS server for testing
- `createTestRequest()` - Helper to create HTTP requests with various header configurations
- `createTestRequireAuth()` - Helper to create RequireAuth instance with mock JWKS

**2.2 Unit Tests (`internal/platform/httpserver/auth_test.go`)**

- Test token extraction from all three header sources
- Test priority order (Authorization > x-clerk-auth-token > X-Auth-Token)
- Test missing token scenarios
- Test invalid token format scenarios
- Test JWKS cache behavior (hit, miss, refresh)
- Test token verification (valid, expired, malformed, wrong signature)
- Test claims extraction (user_id, email, org_id, etc.)
- Test context value setting

**2.3 Integration Tests (`internal/platform/httpserver/auth_integration_test.go`)**

- Test with real JWKS endpoint (using Clerk test keys)
- Test middleware chain (auth → tenant → handler)
- Test protected endpoint access
- Test error responses (401, 500)

### 3. Endpoint Testing

**3.1 Test Suite for Protected Endpoints**

Create test file: `internal/platform/httpserver/protected_endpoints_test.go`

- Test each protected endpoint category:
  - `/api/v1/tenants/*` endpoints
  - `/api/v1/brands/*` endpoints  
  - `/api/v1/files/*` endpoints
  - `/api/v1/auth/me`
  - `/api/v1/users/sync`
- For each endpoint, test:
  - Success with valid token
  - Failure with missing token
  - Failure with invalid token
  - Failure with expired token
  - Proper context propagation (user_id, tenant_id)

**3.2 Test Helpers for Endpoint Tests**

- `makeAuthenticatedRequest()` - Helper to make requests with valid tokens
- `makeUnauthenticatedRequest()` - Helper for testing auth failures
- `assertAuthError()` - Helper to verify proper 401 responses

### 4. Logging Enhancements

**4.1 Structured Logging**

- Use consistent log levels:
  - `Debug`: Token extraction details, JWKS cache operations
  - `Info`: Successful authentication with user context
  - `Warn`: Authentication failures (missing/invalid token)
  - `Error`: JWKS failures, token verification errors

**4.2 Log Fields**

Standard fields for all auth logs:

- `method` - HTTP method
- `path` - Request path
- `remote_addr` - Client IP
- `token_source` - Which header provided the token
- `user_id` - Extracted user ID (if available)
- `auth_result` - success/failure
- `auth_failure_reason` - Specific reason if failed

### 5. Documentation Updates

**5.1 Update API Authentication Docs (`docs/API_AUTHENTICATION.md`)**

- Document all supported token header formats
- Add examples for each header type
- Document error response codes and messages
- Add troubleshooting section

**5.2 Add Testing Documentation**

- Create `docs/TESTING_AUTH.md` with:
  - How to run auth tests
  - How to create test tokens
  - How to test protected endpoints
  - Common test scenarios

## File Changes

### Modified Files

- `internal/platform/httpserver/auth.go` - Enhanced middleware with multi-source token extraction
- `docs/API_AUTHENTICATION.md` - Updated documentation

### New Files

- `internal/platform/httpserver/auth_test.go` - Unit tests
- `internal/platform/httpserver/auth_integration_test.go` - Integration tests  
- `internal/platform/httpserver/auth_test_helpers.go` - Test helpers
- `internal/platform/httpserver/protected_endpoints_test.go` - Endpoint tests
- `docs/TESTING_AUTH.md` - Testing documentation

## Testing Strategy

### Unit Tests (Fast, Isolated)

- Mock JWKS server
- Test token extraction logic
- Test error handling
- Test claims extraction

### Integration Tests (Slower, Real Dependencies)

- Real JWKS endpoint (Clerk test keys)
- Full middleware chain
- Database integration for tenant context

### Test Coverage Goals

- 90%+ coverage for auth middleware
- All error paths tested
- All token sources tested
- All protected endpoints have at least one test

## Success Criteria

1. ✅ Middleware accepts tokens from all three header sources
2. ✅ Comprehensive logging for debugging auth issues
3. ✅ All protected endpoints have test coverage
4. ✅ Clear error messages for different failure scenarios
5. ✅ Documentation updated with new token sources
6. ✅ Tests pass with 90%+ coverage

## Implementation Order

1. **Phase 1**: Enhance auth middleware (token extraction + logging)
2. **Phase 2**: Create test infrastructure (helpers + unit tests)
3. **Phase 3**: Add integration tests
4. **Phase 4**: Test all protected endpoints
5. **Phase 5**: Update documentation

## Notes

- Maintain backward compatibility with existing `Authorization` header
- All changes should be non-breaking
- Logging should be production-ready (structured, not verbose)
- Tests should be fast and reliable
- Consider adding metrics/monitoring hooks for production observability