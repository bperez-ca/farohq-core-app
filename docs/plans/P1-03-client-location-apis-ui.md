# P1-03: Client and location APIs and UI

**Task ID:** P1-03  
**Owner:** Full-stack  
**Phase:** 1 – MVP foundation

## Objective

Ensure CRUD for clients and locations is complete and used by the portal; client invite flow (email link → accept) works; list clients/locations is filtered by tenant.

## Scope (agent-runnable)

- Ensure CRUD for clients and locations is complete and used by portal.
- Client invite flow: email link → accept (invite accept flow works for tenant invites and optionally client invites if applicable).
- List clients/locations filtered by tenant (RLS and API both enforce tenant).

Align with [FARO-CURSOR-TASKS.md](../../FARO-CURSOR-TASKS.md) MVP-009 style but for agency/client/location (not workspace/contact).

## References

- [FARO-CURSOR-TASKS.md](../../FARO-CURSOR-TASKS.md) – MVP-009 (Contacts) pattern, adapt for clients/locations
- [internal/domains/tenants/](../../internal/domains/tenants/) – tenant, client, location use cases and HTTP
- Portal: agency settings, clients list, locations, invite accept pages

## Dependencies

- P1-01 (tenant context and RLS).

## Files to create or modify

| File | Action |
|------|--------|
| `internal/domains/tenants/infra/http/handlers.go` | Verify/complete – client and location CRUD handlers |
| `api/openapi.yaml` | Verify – clients and locations endpoints documented |
| Portal: clients list, client detail, locations list | Verify/wire – use core-app API |
| Portal: invite accept flow | Verify – email link → accept → redirect |

## Acceptance criteria

- [ ] GET/POST /api/v1/tenants/:id/clients and GET/PUT /api/v1/clients/:id work and return only tenant’s clients.
- [ ] GET/POST /api/v1/clients/:id/locations and GET/PUT /api/v1/locations/:id work and return only tenant’s locations (via client).
- [ ] Client invite flow: user receives email with link; clicking link and accepting adds them to tenant (or client); redirect works.
- [ ] Portal pages for clients and locations load data from core-app API and respect tenant (X-Tenant-ID or domain).
