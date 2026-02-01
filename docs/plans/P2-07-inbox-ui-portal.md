# P2-07: Inbox UI (portal)

**Task ID:** P2-07  
**Owner:** Frontend  
**Phase:** 2 – Core integrations

## Objective

Build inbox UI in the portal: conversation list, message thread, send reply. Optional: contact panel, assignment. Use new conversations/messages API.

## Scope (agent-runnable)

- Conversation list: fetch GET /api/v1/conversations; show contact/last message/preview; click to select.
- Message thread: for selected conversation, GET /api/v1/conversations/:id/messages; display messages; send via POST /api/v1/conversations/:id/messages.
- Layout: list (e.g. left) + thread (right); optional contact panel (right) with contact details.
- Optional: assignment dropdown; filter “Assigned to me”.
- Follow [FARO-CURSOR-IMPLEMENTATION-GUIDE.md](../../FARO-CURSOR-IMPLEMENTATION-GUIDE.md) Section 9 inbox components and [FARO-CURSOR-TASKS.md](../../FARO-CURSOR-TASKS.md) MVP-007.

## References

- [FARO-CURSOR-IMPLEMENTATION-GUIDE.md](../../FARO-CURSOR-IMPLEMENTATION-GUIDE.md) – Section 9 inbox components
- [FARO-CURSOR-TASKS.md](../../FARO-CURSOR-TASKS.md) – MVP-007 Inbox UI
- Portal: [app/business/inbox/](../../../farohq-portal/src/app/business/inbox/) (or equivalent)

## Dependencies

- P2-04 (conversations and messages API); P2-06 optional (polling is acceptable).

## Files to create or modify

| File | Action |
|------|--------|
| Portal: inbox page and components | Create or extend – ConversationList, ConversationThread, MessageBubble, MessageInput |
| Portal: API client | Add – fetch conversations, messages, send message (with tenant context) |
| Navigation | Ensure inbox is reachable from agency/business nav |

## Acceptance criteria

- [ ] Inbox page shows list of conversations for the current tenant.
- [ ] Selecting a conversation loads and displays messages; user can send a reply (POST message).
- [ ] Sent message appears in thread (after refresh or real-time if P2-06 done).
- [ ] UI uses tenant context (X-Tenant-ID or domain) for all API calls.
