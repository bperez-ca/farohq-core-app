# P4-07: Conversation assignment and team collaboration

**Task ID:** P4-07  
**Owner:** Full-stack  
**Phase:** 4 – Enhanced inbox & AI

## Objective

Enable team collaboration in the inbox. Agents can assign conversations to team members, add internal notes (not visible to customers), and @mention colleagues. Provides visibility into who is handling each conversation.

## Scope (agent-runnable)

- **Assignment:**
  - Add `assigned_to` (user_id, nullable) to `conversations` table.
  - API: `PATCH /api/v1/conversations/:id/assign` with `{ user_id }` to assign/unassign.
  - Filter: `GET /api/v1/conversations?assigned_to=me` or `?assigned_to={user_id}`.
  - UI: Dropdown to assign conversation to team member; "Assigned to me" filter.
- **Internal notes:**
  - Add `internal_notes` table (id, conversation_id, author_id, content, mentions, created_at) or add `is_internal` flag to `messages`.
  - API: `POST /api/v1/conversations/:id/notes` to add internal note.
  - UI: Internal notes section in conversation thread (visually distinct, labeled "Internal").
- **@Mentions:**
  - Parse `@username` in notes; store mentioned user IDs.
  - Optional: Send notification to mentioned users.
- **Tier gating:** Basic assignment available at all tiers; internal notes and mentions at Growth+ (or all tiers).

## References

- [Agency-Pricing-Tiers.md](../../Agency-Pricing-Tiers.md) – Shared inbox collaboration features
- [SMB-Pricing-Tiers.md](../../SMB-Pricing-Tiers.md) – Team collaboration
- [P2-01-conversations-messages-schema.md](P2-01-conversations-messages-schema.md) – Conversations schema

## Dependencies

- P2-01 (conversations table)
- P2-04 (conversations API)
- P1-01 (RBAC for user list)

## Files to create or modify

| File | Action |
|------|--------|
| `migrations/` | Add – `assigned_to` column on conversations; `internal_notes` table |
| `internal/domains/conversations/` | Extend – Assign use case, notes use case |
| `api/openapi.yaml` | Add – Assign and notes endpoints |
| Portal: inbox | Add – Assignment dropdown, "Assigned to me" filter |
| Portal: conversation thread | Add – Internal notes section, note input with @mention autocomplete |

## Acceptance criteria

- [ ] Agent can assign a conversation to a team member; assignment shows in conversation list.
- [ ] "Assigned to me" filter returns only conversations assigned to the current user.
- [ ] Agent can add internal notes; notes are visible in thread but marked as internal.
- [ ] @mentions are parsed and stored; mentioned users highlighted.
- [ ] Assignment and notes respect tenant context (RLS).
- [ ] Optional: Notification sent to assigned user or mentioned users.
