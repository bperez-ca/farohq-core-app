# P2-02: GBP OAuth and basic sync

**Task ID:** P2-02  
**Owner:** Backend  
**Phase:** 2 – Core integrations

## Objective

Implement Google Business Profile OAuth flow and basic NAP (name, address, phone) sync. Store tokens per location/tenant; expose endpoints for OAuth URL, callback, and sync.

## Scope (agent-runnable)

- New domain or module (e.g. `internal/domains/gbp/`): OAuth flow (consent screen, callback), store tokens per location/tenant.
- Sync NAP from GBP API to location (or return for UI to display).
- Endpoints (example paths):
  - `GET /api/v1/gbp/oauth/url` – return authorization URL (state can include tenant_id/location_id).
  - `GET /api/v1/gbp/oauth/callback` – handle callback, exchange code, store tokens, redirect.
  - `POST /api/v1/gbp/sync/:locationId` – trigger sync for location (read from GBP, update location or return data).
- Use tenant/location (not workspace); follow [FARO-CURSOR-TASKS.md](../../FARO-CURSOR-TASKS.md) pattern adapted to agency_id/location_id.

## References

- [FARO-CURSOR-TASKS.md](../../FARO-CURSOR-TASKS.md) – similar OAuth/sync pattern
- [FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md](../FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md) – Week 5 GBP OAuth + Sync
- Google Business Profile API docs (OAuth 2.0, My Business API)

## Dependencies

- P2-01 not required for OAuth/sync logic but locations must exist (already in Phase 1).

## Files to create or modify

| File | Action |
|------|--------|
| `internal/domains/gbp/` (or equivalent) | Create – domain, app, infra (OAuth client, token storage, sync) |
| `migrations/` | Add – table for GBP tokens per location if not in P2-01 |
| `api/openapi.yaml` | Add – GBP OAuth and sync endpoints |
| Router / composition | Register GBP routes with auth and tenant context |

## Acceptance criteria

- [ ] GET /api/v1/gbp/oauth/url returns a valid Google OAuth URL with state.
- [ ] Callback exchanges code, stores tokens for the location/tenant, redirects to success URL.
- [ ] POST /api/v1/gbp/sync/:locationId runs sync (NAP) for that location; tenant can only sync own locations.
- [ ] Tokens stored securely (e.g. encrypted at rest); RLS restricts access by agency.
