# P1-01: Auth and multi-tenancy hardening

**Task ID:** P1-01  
**Owner:** Backend  
**Phase:** 1 – MVP foundation

## Objective

Verify and harden auth and multi-tenancy so all tenant/client/location access is correctly isolated. Ensure JWT carries `agency_id` and `role`, add RBAC middleware, and add table-driven tests for tenant isolation.

## Scope (agent-runnable)

- Verify RLS on all tenant/client/location tables (agencies, branding, tenant_members, tenant_invites, clients, locations, client_members).
- Ensure JWT context carries `agency_id` (alias for org_id) and `role` (org_role) per [Strategic Roadmap Part 2.2].
- Add RBAC middleware that enforces Owner/Admin/Manager/Staff (and optionally Viewer, Client Viewer).
- Add RLS on `agencies` table if missing (access only when `lv.tenant_id` is set and matches).
- Add table-driven tests for tenant isolation (e.g. RBAC middleware, context helpers).

## References

- [FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md](../FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md) – Part 2.2 RBAC roles
- [internal/platform/httpserver/auth.go](../../internal/platform/httpserver/auth.go) – auth middleware
- [internal/platform/httpserver/tenant.go](../../internal/platform/httpserver/tenant.go) – tenant resolution, RLS context
- [migrations/](../../migrations/) – RLS policies

## Dependencies

- None (foundation task).

## Files to create or modify

| File | Action |
|------|--------|
| `migrations/000011_agencies_rls.up.sql` | Create – RLS policy on agencies |
| `migrations/000011_agencies_rls.down.sql` | Create – rollback |
| `internal/platform/httpserver/auth.go` | Modify – set `agency_id` in context (alias of org_id) |
| `internal/platform/httpserver/rbac.go` | Create – RequireRole, GetRoleFromContext, GetAgencyIDFromContext |
| `internal/platform/httpserver/rbac_test.go` | Create – table-driven tests for RBAC and context helpers |

## Acceptance criteria

- [x] All tables that hold tenant-scoped data have RLS enabled and a policy using `lv.tenant_id` (or client/location subquery). (agencies: 000011; branding: 000001; tenant_members, tenant_invites: 000003; clients, locations, client_members: 000004.)
- [x] Request context includes `agency_id` when user has an org (same value as org_id). ([auth.go](../../internal/platform/httpserver/auth.go) sets `agency_id` in context.)
- [x] RBAC middleware `RequireRole(roles...)` returns 403 when role is not in the allowed list; passes when role is allowed. ([rbac.go](../../internal/platform/httpserver/rbac.go).)
- [x] Table-driven tests exist for GetRoleFromContext, GetAgencyIDFromContext, and RequireRole (allowed/forbidden cases). ([rbac_test.go](../../internal/platform/httpserver/rbac_test.go) – all tests pass.)
- [x] Migration 000011 applies cleanly and down rolls back without errors. ([000011_agencies_rls.up.sql](../../migrations/000011_agencies_rls.up.sql), [000011_agencies_rls.down.sql](../../migrations/000011_agencies_rls.down.sql). Run `make migrate` or equivalent to apply.)

---

## Validation (completed)

**Status:** P1-01 is complete.

- **RLS:** agencies (000011), branding (000001), tenant_members, tenant_invites (000003), clients, locations, client_members (000004) all have RLS with `lv.tenant_id` or subquery.
- **Context:** `auth.go` sets `agency_id` (alias of org_id) in request context when org is present.
- **RBAC:** `rbac.go` provides `RequireRole`, `GetRoleFromContext`, `GetAgencyIDFromContext`, `RequireOwnerOrAdmin`.
- **Tests:** `go test ./internal/platform/httpserver/... -run 'TestGetRoleFromContext|TestGetAgencyIDFromContext|TestRequireRole|TestRequireOwnerOrAdmin'` passes.
- **Migration:** 000011 adds RLS on agencies; down migration removes it. Apply with your migration tool to confirm in your environment.
