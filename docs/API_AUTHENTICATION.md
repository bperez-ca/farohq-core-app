# API Authentication Guide

## Overview

Most API endpoints require authentication via Clerk JWT tokens. The authentication flow works as follows:

1. **Frontend** authenticates with Clerk and receives a JWT token
2. **Frontend** sends requests with `Authorization: Bearer <token>` header
3. **Backend** validates the token using Clerk's JWKS endpoint
4. **Backend** extracts user/org information from the token

## Required Headers

### Authentication
```
Authorization: Bearer <clerk_jwt_token>
```

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
- **Missing token**: `Authorization header required`
- **Invalid token**: `Invalid token`
- **Token verification failed**: `Failed to verify token`

### 400 Bad Request
- **Missing tenant**: `Failed to resolve tenant` (when tenant resolution fails)
- **Invalid request**: Various validation errors

### 404 Not Found
- **Route not found**: `404 page not found`
- **Resource not found**: Resource-specific errors

## Testing with curl

### With Authentication Token
```bash
curl -H "Authorization: Bearer YOUR_CLERK_TOKEN" \
     http://localhost:8080/api/v1/brands
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
