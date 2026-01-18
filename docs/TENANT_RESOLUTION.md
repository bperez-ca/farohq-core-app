# Tenant Resolution Documentation

## Overview

The tenant resolution system provides secure, multi-source tenant identification with access validation. It ensures that users can only access tenants they have permission to access, regardless of how the tenant is specified (domain, header, URL parameter, etc.).

## Architecture

### Middleware Order

The tenant resolution middleware runs **after** authentication to ensure user context is available:

```
Request → RequireAuth → TenantResolutionWithAuth → RequireTenantContext → Handler
         ✅ Extract user_id  ✅ Validate access      ✅ Ensure exists
```

### Resolution Priority

Tenant resolution follows this priority order:

1. **Domain-based**: Query `branding` table by host/domain
2. **X-Tenant-ID header**: Use header value if present
3. **URL parameter**: Extract from `/api/v1/tenants/{id}/...` paths
4. **User's first tenant**: Fallback to user's first accessible tenant
5. **Error**: If no accessible tenants, return error

### Access Validation

All resolved tenants are validated against the user's accessible tenants list:

- If resolved tenant is in user's accessible tenants → ✅ Allow
- If resolved tenant is NOT in user's accessible tenants → ⚠️ Log security warning + fallback to user's first tenant
- If user has no accessible tenants → ❌ Return 403 Forbidden

## Usage

### Basic Setup

```go
// In main.go
tenantResolver := tenant.NewResolver(pool, logger)
userRepo := appComposition.UserRepo

// Create tenant cache (optional)
tenantCache := tenant.NewTenantCache(5*time.Minute, logger)

// Apply middleware AFTER authentication
r.Use(authMiddleware.RequireAuth)
r.Use(httpserver.TenantResolutionWithAuth(
    tenantResolver,
    tenantCache,  // Can be nil to disable caching
    userRepo,
    pool,
    logger,
))
```

### Public Routes

These routes skip tenant resolution entirely:

- `/healthz`, `/readyz`, `/`
- `/api/v1/brand/by-domain` (public brand lookup)
- `/api/v1/brand/by-host` (public brand lookup)

### Routes Without Tenant Context

These routes require auth but not tenant:

- `/api/v1/tenants/my-orgs`
- `/api/v1/auth/me`
- `/api/v1/users/sync`
- `POST /api/v1/tenants` (creating new tenant)
- `POST /api/v1/tenants/onboard`

## API

### Resolver Methods

#### `GetUserTenantIDs(ctx, userID) ([]string, error)`

Returns all tenant IDs the user has access to.

```go
tenantIDs, err := resolver.GetUserTenantIDs(ctx, userID)
```

#### `ValidateUserAccess(ctx, userID, tenantID) (bool, error)`

Validates if user has access to a specific tenant.

```go
hasAccess, err := resolver.ValidateUserAccess(ctx, userID, tenantID)
```

#### `ResolveTenantWithValidation(ctx, userID, host, tenantIDHeader, urlPath) (*TenantResolutionResult, error)`

Resolves tenant from multiple sources with access validation.

```go
result, err := resolver.ResolveTenantWithValidation(
    ctx,
    userID,
    "example.com",
    "tenant-uuid-from-header",
    "/api/v1/tenants/tenant-uuid/invites",
)

// result.TenantID - resolved tenant ID
// result.Source - where tenant was resolved from (domain/header/url/token/fallback)
// result.Validated - whether access was validated
// result.FallbackUsed - whether fallback strategy was used
```

### Error Types

```go
var (
    ErrTenantNotFound      = errors.New("tenant not found")
    ErrTenantAccessDenied  = errors.New("user does not have access to tenant")
    ErrNoAccessibleTenants = errors.New("user has no accessible tenants")
    ErrInvalidTenantID     = errors.New("invalid tenant ID format")
)
```

## Security Features

### Access Validation

- **Always validates** user access to resolved tenant
- **Never trusts** X-Tenant-ID header without validation
- **Logs all** invalid access attempts
- **Uses fallback** only to user's accessible tenants

### Security Logging

All tenant resolution events are logged with:

- `user_id` - Database UUID of the user
- `clerk_user_id` - Clerk user ID
- `resolved_tenant_id` - Tenant ID that was resolved
- `resolution_source` - Where tenant was resolved from (domain/header/url/token/fallback)
- `validated` - Whether access was validated
- `fallback_used` - Whether fallback strategy was used
- `accessible_tenants` - List of user's accessible tenants (for security warnings)

### Security Warnings

Warnings are logged for:

- Invalid tenant access attempts
- Tenant resolution failures
- Fallback usage
- Missing tenant context

## Caching

### Tenant Cache

The tenant cache reduces database queries by caching user's accessible tenants:

```go
// Create cache with 5 minute TTL
cache := tenant.NewTenantCache(5*time.Minute, logger)

// Use in middleware
r.Use(httpserver.TenantResolutionWithAuth(
    tenantResolver,
    cache,  // Pass cache instance
    userRepo,
    pool,
    logger,
))
```

### Cache Invalidation

Cache should be invalidated when:

- User is added to tenant
- User is removed from tenant
- User's role changes
- Tenant is deleted

```go
// Invalidate cache for specific user
cache.Invalidate(userID)

// Invalidate all cache entries
cache.InvalidateAll()
```

### Cache Cleanup

Periodically clean up expired entries:

```go
// Run cleanup in background goroutine
go func() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    for range ticker.C {
        cache.Cleanup()
    }
}()
```

## Testing

### Unit Tests

Test individual resolver methods:

```bash
go test ./internal/platform/tenant -v
```

### Integration Tests

Test middleware chain with database:

```bash
# Ensure test database is running
make up

# Run integration tests
go test ./internal/platform/httpserver -v -run TestTenantResolutionWithAuth
```

### Test Database Setup

Tests use a test database connection. Ensure PostgreSQL is running:

```bash
# Start test database
make up

# Or use E2E database
make e2e-start
```

## Troubleshooting

### "User has no accessible tenants"

**Cause**: User is not a member of any tenant.

**Solution**: Add user to a tenant via `tenant_members` table or use invite flow.

### "Invalid tenant access attempt"

**Cause**: User tried to access a tenant they don't belong to.

**Solution**: This is expected behavior - the system will fallback to user's first accessible tenant. Check logs for security warnings.

### "Failed to resolve tenant"

**Cause**: No tenant could be resolved from any source and user has no accessible tenants.

**Solution**: Ensure user is a member of at least one tenant.

### Tenant resolution runs before authentication

**Cause**: Middleware order is incorrect.

**Solution**: Ensure `TenantResolutionWithAuth` runs AFTER `RequireAuth` middleware.

## Migration Guide

### From Old Tenant Resolution

The old `TenantResolution` middleware ran before authentication and didn't validate access. To migrate:

1. **Remove old middleware**:
   ```go
   // OLD - Remove this
   r.Use(httpserver.TenantResolution(tenantResolver, pool, logger))
   ```

2. **Add new middleware after auth**:
   ```go
   // NEW - Add this after RequireAuth
   r.Use(authMiddleware.RequireAuth)
   r.Use(httpserver.TenantResolutionWithAuth(
       tenantResolver,
       nil,  // or tenantCache
       appComposition.UserRepo,
       pool,
       logger,
   ))
   ```

3. **Update frontend**: Remove manual `X-Tenant-ID` header setting - backend will resolve automatically.

### Breaking Changes

- Tenant resolution now requires authentication
- Invalid tenant access attempts are logged and fallback is used
- Routes that don't need tenant context must be explicitly listed in middleware

## Performance Considerations

### Database Queries

- Use prepared statements for tenant queries
- Cache user's accessible tenants (5 min TTL recommended)
- Limit tenant list size (users shouldn't have 100s of tenants)

### Cache Performance

- Cache hit rate should be > 80% for optimal performance
- Monitor cache invalidation frequency
- Adjust TTL based on tenant membership change frequency

## Best Practices

1. **Always validate access** - Never trust tenant IDs from headers/URLs
2. **Log security events** - Monitor invalid access attempts
3. **Use caching** - Reduce database queries for user tenant lists
4. **Invalidate cache** - Ensure cache is invalidated on membership changes
5. **Handle errors gracefully** - Provide clear error messages to users
6. **Test thoroughly** - Test all resolution sources and edge cases

## Related Documentation

- [API Authentication](./API_AUTHENTICATION.md) - Authentication middleware
- [Testing Auth](./TESTING_AUTH.md) - Testing authentication flow
