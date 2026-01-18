# API Authentication Guide

## Overview

Most API endpoints require authentication via Clerk JWT tokens. The authentication flow works as follows:

1. **Frontend** authenticates with Clerk and receives a JWT token
2. **Frontend** sends requests with `Authorization: Bearer <token>` header
3. **Backend** validates the token using Clerk's JWKS endpoint
4. **Backend** extracts user/org information from the token

## Required Headers

### Authentication

The backend supports multiple token header formats (checked in priority order):

1. **Authorization header (standard, highest priority)**
   ```
   Authorization: Bearer <clerk_jwt_token>
   ```

2. **x-clerk-auth-token header (Clerk's automatic header)**
   ```
   x-clerk-auth-token: <clerk_jwt_token>
   ```

3. **X-Auth-Token header (custom fallback)**
   ```
   X-Auth-Token: <clerk_jwt_token>
   ```

**Note**: The `Authorization: Bearer` format is the recommended standard. Other formats are provided for compatibility with different client configurations.

### Tenant Resolution
The backend resolves tenants in this order:
1. **Domain-based**: From the `Host` header (e.g., `agency1.example.com`)
2. **Header fallback**: `X-Tenant-ID` header (for API calls)
3. **Public routes**: Some routes don't require tenant context

## Endpoints

### Public Endpoints (No Auth Required)
- `GET /healthz` - Health check
- `GET /readyz` - Readiness check
- `GET /` - API info
- `GET /api/v1/brand/by-domain` - Get brand by domain
- `GET /api/v1/brand/by-host` - Get brand by host

### Protected Endpoints (Auth Required)
All endpoints under `/api/v1` except the public brand routes require authentication:

- `GET /api/v1/brands` - List brands (requires auth + tenant)
- `POST /api/v1/brands` - Create brand (requires auth + tenant)
- `GET /api/v1/files` - List files (requires auth + tenant)
- `POST /api/v1/files/sign` - Sign upload URL (requires auth + tenant)
- All tenant endpoints (requires auth + tenant)

## Error Responses

### 401 Unauthorized
- **Missing token**: `Authorization header required` - No authentication token found in any supported header
- **Invalid token**: `Invalid token` - Token is malformed, expired, or has invalid signature
- **Token verification failed**: `Failed to verify token` - JWKS unavailable or token verification error

### 500 Internal Server Error
- **JWKS refresh failed**: `Failed to verify token` - Unable to refresh JWKS cache

### 400 Bad Request
- **Missing tenant**: `Failed to resolve tenant` (when tenant resolution fails)
- **Invalid request**: Various validation errors

### 404 Not Found
- **Route not found**: `404 page not found`
- **Resource not found**: Resource-specific errors

## Testing with curl

### With Authentication Token (Authorization header)
```bash
curl -H "Authorization: Bearer YOUR_CLERK_TOKEN" \
     http://localhost:8080/api/v1/auth/me
```

### With Authentication Token (x-clerk-auth-token header)
```bash
curl -H "x-clerk-auth-token: YOUR_CLERK_TOKEN" \
     http://localhost:8080/api/v1/auth/me
```

### With Authentication Token (X-Auth-Token header)
```bash
curl -H "X-Auth-Token: YOUR_CLERK_TOKEN" \
     http://localhost:8080/api/v1/auth/me
```

### Testing Error Responses
```bash
# Missing token
curl http://localhost:8080/api/v1/auth/me
# Expected: 401 Unauthorized - "Authorization header required"

# Invalid token
curl -H "Authorization: Bearer invalid-token" \
     http://localhost:8080/api/v1/auth/me
# Expected: 401 Unauthorized - "Invalid token"
```

## Troubleshooting

### Token Not Accepted

If your token is being rejected:

1. **Check token format**: Ensure the token is a valid JWT
2. **Check token expiration**: Verify the token hasn't expired
3. **Check header format**: For Authorization header, ensure it's `Bearer <token>` (with space)
4. **Check JWKS URL**: Verify `CLERK_JWKS_URL` environment variable is set correctly
5. **Check logs**: Review backend logs for specific error messages:
   - `missing_token`: No token found in any header
   - `empty_token`: Token string is empty
   - `token_expired`: Token has expired
   - `token_signature_invalid`: Token signature doesn't match JWKS
   - `token_malformed`: Token is not a valid JWT
   - `jwks_unavailable`: Cannot fetch JWKS from Clerk

### Multiple Token Sources

The backend checks headers in priority order. If you're using multiple headers:
- `Authorization` header will always be used if present (even if other headers also have tokens)
- `x-clerk-auth-token` will be used if `Authorization` is missing
- `X-Auth-Token` will be used if both above are missing

### Logging

The backend logs all authentication attempts with structured fields:
- `token_source`: Which header provided the token
- `auth_result`: `success` or `failure`
- `auth_failure_reason`: Specific reason if authentication failed
- `user_id`: Extracted user ID from token
- `verify_duration_ms`: Time taken to verify token
- `total_duration_ms`: Total authentication timehttp://localhost:8080/api/v1/brands
```

### With Tenant ID Header
```bash
curl -H "Authorization: Bearer YOUR_CLERK_TOKEN" \
     -H "X-Tenant-ID: YOUR_TENANT_UUID" \
     http://localhost:8080/api/v1/brands
```

## Troubleshooting

### "Authorization header required"
- Ensure you're sending the `Authorization: Bearer <token>` header
- Check that the token is not expired

### "Invalid token"
- Token may be from a different Clerk instance
- Token may be malformed
- JWKS keys may have rotated (backend should auto-refresh)

### "Failed to resolve tenant"
- No tenant found for the domain in the `Host` header
- `X-Tenant-ID` header not provided or invalid
- For localhost testing, use `X-Tenant-ID` header

### "400 Bad Request" on brands/files endpoints
- Usually means tenant resolution failed
- Check backend logs for specific error
- Try providing `X-Tenant-ID` header

## Local Development

For local development without a real tenant:

1. **Create a test tenant** in the database
2. **Use X-Tenant-ID header** in requests:
   ```bash
   curl -H "Authorization: Bearer TOKEN" \
        -H "X-Tenant-ID: <tenant-uuid>" \
        http://localhost:8080/api/v1/brands
   ```

3. **Or set up domain mapping**:
   - Add a branding record with `domain = 'localhost'`
   - Access via `http://localhost:8080` (not `localhost:3001`)
