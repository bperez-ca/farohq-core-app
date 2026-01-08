# Migration Map: LVOS Core App → FaroHQ Core App

This document maps the original `lvos/services/core-app` structure to the new `farohq-core-app` hexagonal architecture.

## Overview

The refactoring transforms a monolithic module-based structure into a domain-driven hexagonal architecture with vertical slicing.

## Current Structure (LVOS Core App)

```
services/core-app/
├── cmd/core-app/main.go
├── internal/
│   ├── auth/module.go
│   ├── brand/module.go
│   ├── files/module.go
│   ├── tenants/
│   │   ├── domain/          # Already hexagonal!
│   │   ├── app/             # Use cases
│   │   ├── infra/postgres/  # Repositories
│   │   └── ports/           # Interfaces
│   ├── middleware/
│   │   ├── auth.go          # Clerk JWT validation
│   │   └── tenant.go        # Tenant resolution + RLS
│   └── shared/
│       ├── config.go        # Configuration
│       └── tenant_resolver.go
├── migrations/
│   ├── 0001_init.sql
│   ├── 0002_seed_data.sql
│   ├── 0003_create_tenants.sql
│   └── 0004_agency_hierarchy.sql
└── api/openapi.yaml
```

## New Structure (FaroHQ Core App)

```
farohq-core-app/
├── cmd/server/main.go
├── internal/
│   ├── platform/            # Cross-cutting concerns
│   │   ├── config/
│   │   ├── logging/
│   │   ├── db/
│   │   ├── httpserver/
│   │   └── tenant/
│   ├── domains/              # Business domains (vertical slices)
│   │   ├── tenants/
│   │   │   ├── domain/
│   │   │   │   ├── model/
│   │   │   │   ├── services/
│   │   │   │   └── ports/
│   │   │   ├── app/usecases/
│   │   │   └── infra/
│   │   │       ├── db/
│   │   │       └── http/
│   │   ├── brand/
│   │   ├── files/
│   │   └── auth/
│   └── app/
│       ├── composition/      # Dependency wiring
│       └── health/
├── migrations/
│   ├── 000001_init.up.sql
│   ├── 000002_seed_data.up.sql
│   ├── 000003_create_tenants.up.sql
│   └── 000004_agency_hierarchy.up.sql
└── docs/
```

## Mapping: Current Code → New Domain Slices

### 1. Tenants Domain

**Current:** `internal/tenants/` (already hexagonal)

**New:** `internal/domains/tenants/`

| Current | New | Notes |
|---------|-----|-------|
| `domain/tenant.go` | `domain/model/tenant.go` | Renamed for clarity |
| `domain/client.go` | `domain/model/client.go` | Renamed for clarity |
| `domain/location.go` | `domain/model/location.go` | Renamed for clarity |
| `domain/tenant_member.go` | `domain/model/tenant_member.go` | New file (extracted) |
| `domain/client_member.go` | `domain/model/client_member.go` | New file (extracted) |
| `domain/invite.go` | `domain/model/invite.go` | Renamed for clarity |
| `domain/tier.go` | `domain/model/tier.go` | Renamed for clarity |
| `domain/errors.go` | `domain/errors.go` | Same |
| `domain/seat_validator.go` | `domain/services/seat_validator.go` | Moved to services |
| `app/*.go` | `app/usecases/*.go` | All use cases preserved |
| `infra/postgres/*.go` | `infra/db/*.go` | Repository implementations |
| `ports/*.go` | `domain/ports/outbound/*.go` | Outbound ports |
| N/A | `infra/http/*.go` | New HTTP handlers layer |

**Use Cases Ported:**
- `create_tenant.go`
- `get_tenant.go`
- `update_tenant.go`
- `invite_member.go`
- `accept_invite.go`
- `list_members.go`
- `remove_member.go`
- `list_roles.go`
- `create_client.go`
- `list_clients.go`
- `get_client.go`
- `update_client.go`
- `add_client_member.go`
- `list_client_members.go`
- `remove_client_member.go`
- `create_location.go`
- `list_locations.go`
- `update_location.go`
- `get_seat_usage.go`

**Repositories Ported:**
- `tenant_repository.go`
- `tenant_member_repository.go`
- `invite_repository.go`
- `client_repository.go`
- `location_repository.go`
- `client_member_repository.go`

### 2. Brand Domain

**Current:** `internal/brand/module.go` (monolithic)

**New:** `internal/domains/brand/` (hexagonal)

| Current | New | Notes |
|---------|-----|-------|
| `Branding` struct | `domain/model/branding.go` | Extracted domain entity |
| `Agency` struct | Uses `tenants` domain | Agencies are tenants |
| `GetByDomainHandler` | `app/usecases/get_by_domain.go` + `infra/http/handlers.go` | Separated concerns |
| `GetByHostHandler` | `app/usecases/get_by_host.go` + `infra/http/handlers.go` | Separated concerns |
| `ListBrandsHandler` | `app/usecases/list_brands.go` + `infra/http/handlers.go` | Separated concerns |
| `CreateBrandHandler` | `app/usecases/create_brand.go` + `infra/http/handlers.go` | Separated concerns |
| `GetBrandHandler` | `app/usecases/get_brand.go` + `infra/http/handlers.go` | Separated concerns |
| `UpdateBrandHandler` | `app/usecases/update_brand.go` + `infra/http/handlers.go` | Separated concerns |
| `DeleteBrandHandler` | `app/usecases/delete_brand.go` + `infra/http/handlers.go` | Separated concerns |
| Direct DB access | `infra/db/brand_repository.go` | Repository pattern |
| N/A | `domain/ports/outbound/brand_repository.go` | Outbound port |
| N/A | `domain/ports/inbound/*.go` | Inbound ports (use case interfaces) |

### 3. Files Domain

**Current:** `internal/files/module.go` (monolithic)

**New:** `internal/domains/files/` (hexagonal)

| Current | New | Notes |
|---------|-----|-------|
| `SignHandler` | `app/usecases/sign_upload.go` + `infra/http/handlers.go` | Separated concerns |
| `DeleteFileHandler` | `app/usecases/delete_file.go` + `infra/http/handlers.go` | Separated concerns |
| Direct S3 client | `infra/s3/storage.go` | Storage adapter |
| `AllowedAssets` | `domain/services/asset_validator.go` | Domain service |
| `generateObjectKey` | `domain/services/key_generator.go` | Domain service |
| N/A | `domain/ports/outbound/storage.go` | Outbound port |
| N/A | `domain/ports/inbound/*.go` | Inbound ports |

### 4. Auth Domain

**Current:** `internal/auth/module.go` (minimal)

**New:** `internal/domains/auth/infra/http/` (minimal, mostly infra)

| Current | New | Notes |
|---------|-----|-------|
| `MeHandler` | `infra/http/handlers.go` | Preserved as-is |
| N/A | `infra/http/router.go` | Route registration |

**Note:** Auth is mostly infrastructure (middleware in `platform/httpserver/auth.go`). The domain is minimal.

### 5. Platform Layer

**Current:** `internal/shared/`, `internal/middleware/`

**New:** `internal/platform/`

| Current | New | Notes |
|---------|-----|-------|
| `shared/config.go` | `platform/config/config.go` | Preserved |
| `shared/tenant_resolver.go` | `platform/tenant/resolver.go` | Preserved |
| `middleware/auth.go` | `platform/httpserver/auth.go` | Preserved |
| `middleware/tenant.go` | `platform/httpserver/tenant.go` | Preserved |
| N/A | `platform/logging/logger.go` | New structured logger |
| N/A | `platform/db/pool.go` | Database pool management |
| N/A | `platform/db/health.go` | Health check utilities |
| N/A | `platform/httpserver/middleware.go` | Common middleware |
| N/A | `platform/httpserver/server.go` | HTTP server wrapper |

### 6. Main Entry Point

**Current:** `cmd/core-app/main.go`

**New:** `cmd/server/main.go`

| Current | New | Notes |
|---------|-----|-------|
| Module initialization | `app/composition/composition.go` | Dependency wiring |
| Route registration | `app/composition/router.go` | Unified routing |
| Health endpoints | `app/health/handlers.go` | Extracted |

### 7. Migrations

**Current:** `migrations/0001_init.sql`, etc.

**New:** `migrations/000001_init.up.sql`, etc.

| Current | New | Notes |
|---------|-----|-------|
| `0001_init.sql` | `000001_init.up.sql` | Renamed for golang-migrate |
| `0002_seed_data.sql` | `000002_seed_data.up.sql` | Renamed for golang-migrate |
| `0003_create_tenants.sql` | `000003_create_tenants.up.sql` | Renamed for golang-migrate |
| `0004_agency_hierarchy.sql` | `000004_agency_hierarchy.up.sql` | Renamed for golang-migrate |

**Migration Tool:** Changed from custom scripts to `golang-migrate/migrate`

## API Endpoints (Backward Compatible)

All API endpoints remain backward compatible:

| Endpoint | Method | Domain | Handler |
|----------|--------|--------|---------|
| `/api/v1/auth/me` | GET | auth | `AuthHandlers.MeHandler` |
| `/api/v1/tenants` | POST | tenants | `TenantHandlers.CreateTenantHandler` |
| `/api/v1/tenants/{id}` | GET | tenants | `TenantHandlers.GetTenantHandler` |
| `/api/v1/tenants/{id}` | PUT | tenants | `TenantHandlers.UpdateTenantHandler` |
| `/api/v1/tenants/{id}/invites` | POST | tenants | `TenantHandlers.InviteMemberHandler` |
| `/api/v1/tenants/{id}/members` | GET | tenants | `TenantHandlers.ListMembersHandler` |
| `/api/v1/tenants/{id}/members/{user_id}` | DELETE | tenants | `TenantHandlers.RemoveMemberHandler` |
| `/api/v1/tenants/{id}/roles` | GET | tenants | `TenantHandlers.ListRolesHandler` |
| `/api/v1/tenants/{id}/clients` | POST | tenants | `TenantHandlers.CreateClientHandler` |
| `/api/v1/tenants/{id}/clients` | GET | tenants | `TenantHandlers.ListClientsHandler` |
| `/api/v1/clients/{id}` | GET | tenants | `TenantHandlers.GetClientHandler` |
| `/api/v1/clients/{id}` | PUT | tenants | `TenantHandlers.UpdateClientHandler` |
| `/api/v1/clients/{id}/members` | POST | tenants | `TenantHandlers.AddClientMemberHandler` |
| `/api/v1/clients/{id}/members` | GET | tenants | `TenantHandlers.ListClientMembersHandler` |
| `/api/v1/clients/{id}/locations` | POST | tenants | `TenantHandlers.CreateLocationHandler` |
| `/api/v1/clients/{id}/locations` | GET | tenants | `TenantHandlers.ListLocationsHandler` |
| `/api/v1/locations/{id}` | PUT | tenants | `TenantHandlers.UpdateLocationHandler` |
| `/api/v1/brand/by-domain` | GET | brand | `BrandHandlers.GetByDomainHandler` |
| `/api/v1/brand/by-host` | GET | brand | `BrandHandlers.GetByHostHandler` |
| `/api/v1/brands` | GET | brand | `BrandHandlers.ListBrandsHandler` |
| `/api/v1/brands` | POST | brand | `BrandHandlers.CreateBrandHandler` |
| `/api/v1/brands/{brandId}` | GET | brand | `BrandHandlers.GetBrandHandler` |
| `/api/v1/brands/{brandId}` | PUT | brand | `BrandHandlers.UpdateBrandHandler` |
| `/api/v1/brands/{brandId}` | DELETE | brand | `BrandHandlers.DeleteBrandHandler` |
| `/api/v1/files/sign` | POST | files | `FilesHandlers.SignHandler` |
| `/api/v1/files/{key}` | DELETE | files | `FilesHandlers.DeleteFileHandler` |

## Database Schema

**No changes** - all migrations preserved exactly as-is. The database schema remains identical.

## External Integrations

| Integration | Current | New | Notes |
|-------------|---------|-----|-------|
| **Clerk** | `middleware/auth.go` | `platform/httpserver/auth.go` | JWT validation via JWKS |
| **PostgreSQL** | Direct `pgxpool.Pool` | `platform/db/pool.go` | Connection management |
| **AWS S3** | Direct in `files/module.go` | `domains/files/infra/s3/storage.go` | Storage adapter |
| **Row Level Security** | `middleware/tenant.go` | `platform/httpserver/tenant.go` | RLS context setting |

## Environment Variables

**No changes** - all environment variables preserved:
- `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`, `DB_SSLMODE`
- `DATABASE_URL` (alternative to individual DB vars)
- `CLERK_JWKS_URL`
- `AWS_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `S3_BUCKET_NAME`
- `AWS_ENDPOINT_URL` (for LocalStack)
- `PORT`, `WEB_URL`

## Key Architectural Changes

1. **Separation of Concerns**: Business logic separated from infrastructure
2. **Dependency Inversion**: Domain depends on ports (interfaces), not implementations
3. **Testability**: Use cases can be tested without database/HTTP
4. **Composition Root**: All dependencies wired in `app/composition/`
5. **Vertical Slicing**: Each domain owns its complete slice
6. **Screaming Architecture**: Directory structure reveals business domains

## Non-Negotiables Preserved

✅ **No business behavior changes** - All use cases preserve exact logic  
✅ **Backward-compatible API** - All routes and JSON contracts unchanged  
✅ **Runnable migrations** - All migrations preserved and runnable  
✅ **Passing tests** - All existing tests should pass (when ported)

## Next Steps

1. ✅ Scaffold repository structure
2. ✅ Port tenants domain
3. ✅ Port brand domain
4. ✅ Port files domain
5. ✅ Port auth domain
6. ✅ Wire composition layer
7. ⏳ Add tests
8. ⏳ Add CI/CD
9. ⏳ Deploy to Cloud Run

