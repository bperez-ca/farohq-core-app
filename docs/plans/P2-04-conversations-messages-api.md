# P2-04: Conversations and messages API

**Task ID:** P2-04  
**Owner:** Backend  
**Phase:** 2 – Core integrations

## Objective

Expose REST API for conversations and messages: list conversations (by tenant/location, filters), get conversation, list messages (paginated), send message (delegate to Twilio). All with tenant context and RLS.

## Scope (agent-runnable)

- REST endpoints:
  - `GET /api/v1/conversations` – query params: location_id, status, limit, offset; response: list of conversations (with last message preview, contact) and total; RLS by agency_id.
  - `GET /api/v1/conversations/:id` – single conversation with contact/channel info; tenant must own it.
  - `GET /api/v1/conversations/:id/messages` – paginated messages (e.g. before, limit); newest first.
  - `POST /api/v1/conversations/:id/messages` – body: content, optional media; create message and send via Twilio (see P2-03).
- All endpoints require auth and tenant context; OpenAPI spec in [api/openapi.yaml](../../api/openapi.yaml).

## References

- [FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md](../FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md) – Part 2.4 API design
- [api/openapi.yaml](../../api/openapi.yaml) – existing patterns
- P2-01 (schema), P2-03 (send implementation)

## Dependencies

- P2-01 (schema), P2-03 (send path).

## Files to create or modify

| File | Action |
|------|--------|
| `internal/domains/conversations/` (or same as P2-03) | Add – list conversation, get conversation, list messages handlers |
| `api/openapi.yaml` | Add – conversations and messages endpoints and schemas |
| Router / composition | Register routes with RequireAuth and RequireTenantContext |

## Acceptance criteria

- [ ] GET /api/v1/conversations returns only tenant’s conversations; supports location_id and pagination.
- [ ] GET /api/v1/conversations/:id returns 404 if conversation not in tenant.
- [ ] GET /api/v1/conversations/:id/messages returns paginated messages for that conversation.
- [ ] POST /api/v1/conversations/:id/messages creates and sends message; 403 if conversation not in tenant.
- [ ] OpenAPI spec includes request/response schemas for these endpoints.
