# P4-10: Status tracking and collision detection

**Task ID:** P4-10  
**Owner:** Full-stack  
**Phase:** 4 – Enhanced inbox & AI

## Objective

Implement conversation status tracking (Open/In Progress/Resolved) and collision detection to prevent multiple agents from replying to the same conversation simultaneously.

## Scope (agent-runnable)

- **Status tracking:**
  - Add `status` to `conversations` table: `open | in_progress | resolved` (or similar).
  - API: `PATCH /api/v1/conversations/:id/status` with `{ status }`.
  - Auto-update: New inbound message sets status to `open`; agent reply can set to `in_progress`; explicit resolve action sets to `resolved`.
  - UI: Status badge in conversation list; status dropdown in conversation detail.
  - Filters: Filter conversation list by status.
- **Collision detection:**
  - Track which agent is currently viewing/typing in a conversation.
  - Option A (real-time): WebSocket/SSE presence – broadcast "User X is viewing/typing" to other agents viewing same conversation.
  - Option B (simple): Lock on reply – when agent starts typing, set `locked_by` and `locked_at`; if another agent tries to reply, show warning "User X is currently replying".
  - Lock expires after inactivity (e.g., 2 minutes) or on reply sent.
- **UI:**
  - Show indicator when another agent is viewing/typing.
  - Warning modal if agent tries to send while locked by another.
- **Tier gating:** Status tracking at all tiers; collision detection at Growth+ (or all tiers).

## References

- [Agency-Pricing-Tiers.md](../../Agency-Pricing-Tiers.md) – Status tracking, collision detection in shared inbox
- [SMB-Pricing-Tiers.md](../../SMB-Pricing-Tiers.md) – Team collaboration features
- [P2-06-realtime-updates.md](P2-06-realtime-updates.md) – Real-time infrastructure for presence

## Dependencies

- P2-01 (conversations table)
- P2-04 (conversations API)
- P2-06 (real-time for presence, optional)

## Files to create or modify

| File | Action |
|------|--------|
| `migrations/` | Modify – Add `status` to conversations; optionally `locked_by`, `locked_at` |
| `internal/domains/conversations/` | Extend – Status update, lock/unlock logic |
| `api/openapi.yaml` | Add – Status update endpoint, presence/lock endpoints |
| Real-time hub (if P2-06 implemented) | Extend – Presence channel for conversation viewing |
| Portal: conversation list | Add – Status badge, status filter |
| Portal: conversation detail | Add – Status dropdown, presence indicator ("X is typing"), lock warning |

## Acceptance criteria

- [ ] Conversations have a status (open/in_progress/resolved); status can be updated via API.
- [ ] New inbound message sets conversation to `open`.
- [ ] Conversation list can be filtered by status.
- [ ] When agent is typing, other agents see indicator (presence or lock).
- [ ] If agent tries to send while another agent has lock, warning is shown.
- [ ] Lock expires after inactivity or on message sent.
- [ ] Status and presence respect tenant context.
