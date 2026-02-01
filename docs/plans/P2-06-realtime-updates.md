# P2-06: Real-time updates (optional for MVP)

**Task ID:** P2-06  
**Owner:** Backend  
**Phase:** 2 – Core integrations

## Objective

Either Supabase Realtime (if moving to Supabase) or WebSocket: broadcast new message and status updates to tenant. If skipped for MVP, portal can poll; document the decision.

## Scope (agent-runnable)

- Option A: Implement WebSocket (or SSE) server: authenticate by JWT, subscribe by tenant_id; on new message or status update, broadcast to that tenant’s connections.
- Option B: Use Supabase Realtime: subscribe to `messages` (and optionally `conversations`) changes filtered by agency_id; document how portal connects.
- Option C: Skip for MVP – document that portal should poll (e.g. GET conversations/messages every N seconds or on focus); add a short “Real-time” section in docs with the decision and future plan.

## References

- [FARO-CURSOR-IMPLEMENTATION-GUIDE.md](../../FARO-CURSOR-IMPLEMENTATION-GUIDE.md) – real-time architecture
- [FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md](../FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md) – Part 4.3 Real-time

## Dependencies

- P2-03, P2-04 (messages and conversations exist).

## Files to create or modify

| File | Action |
|------|--------|
| Either: WebSocket hub + handler, or Supabase Realtime doc | Implement or document |
| `docs/plans/REALTIME_DECISION.md` or section in E2E_README | Create – document MVP choice (poll vs WS vs Supabase) and next steps |

## Acceptance criteria

- [ ] If implemented: new message or status update is pushed to connected clients for that tenant.
- [ ] If skipped: decision and polling recommendation (or “no real-time”) are documented; portal can still work via polling.
