# P2-01: Database schema for conversations and messages

**Task ID:** P2-01  
**Owner:** Backend  
**Phase:** 2 – Core integrations

## Objective

Add migrations for `conversations` and `messages` (and optionally `contacts`) per Strategic Roadmap Part 2.3. Include `agency_id`/`location_id`, RLS policies, indexes. No business logic yet.

## Scope (agent-runnable)

- Add migrations for:
  - `conversations`: id, location_id, agency_id (denormalized for RLS), channel, contact_phone, contact_name, status, lead_status, lead_value, last_message_at, created_at.
  - `messages`: id, conversation_id, agency_id, direction, content, media_url, media_type, transcript, sent_at.
  - Optionally `contacts` if you want a separate contact entity (e.g. contact_phone, contact_name, agency_id).
- RLS: all tables scoped by `agency_id` (or conversation/location belonging to agency).
- Indexes: conversation list by agency_id and last_message_at; messages by conversation_id and sent_at.
- No application code yet (handlers/use cases in P2-03/P2-04).

## References

- [FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md](../FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md) – Part 2.3 schema (conversations, messages)
- [migrations/](../../migrations/) – existing patterns (agency_id, lv.tenant_id)

## Dependencies

- None (first Phase 2 task).

## Files to create or modify

| File | Action |
|------|--------|
| `migrations/000012_conversations_messages.up.sql` | Create – tables, RLS, indexes |
| `migrations/000012_conversations_messages.down.sql` | Create – drop tables |

## Acceptance criteria

- [ ] Migration applies cleanly; `conversations` and `messages` exist with required columns.
- [ ] RLS is enabled; policies use `agency_id` or subquery via location/client to agency.
- [ ] Indexes exist for listing conversations by agency and last_message_at, and messages by conversation_id.
- [ ] Down migration rolls back without errors.
