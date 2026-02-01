# P2-03: Twilio WhatsApp webhook and send

**Task ID:** P2-03  
**Owner:** Backend  
**Phase:** 2 – Core integrations

## Objective

Implement WhatsApp receive (webhook) and send via Twilio. Validate Twilio signature, create/find contact and conversation, persist message; implement send path that calls Twilio.

## Scope (agent-runnable)

- `POST /api/webhooks/whatsapp` (or `/api/v1/webhooks/whatsapp`): validate Twilio signature; parse incoming message; create or find contact by phone; create or find conversation; persist message; optional: enqueue outbound send or respond with 200 immediately.
- Send path: e.g. `POST /api/v1/conversations/:id/messages` – create message record, call Twilio API to send, update status on delivery callback if needed.
- Use agency_id/location_id (conversations/messages schema from P2-01); map Twilio phone to a location or agency config if needed.
- Use [FARO-CURSOR-TASKS.md](../../FARO-CURSOR-TASKS.md) MVP-005 prompts; adapt to agency_id/location_id.

## References

- [FARO-CURSOR-TASKS.md](../../FARO-CURSOR-TASKS.md) – MVP-005 WhatsApp Integration
- [FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md](../FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md) – Part 2.3 messages schema
- Twilio WhatsApp webhook docs (signature validation, payload format)

## Dependencies

- P2-01 (conversations and messages tables).

## Files to create or modify

| File | Action |
|------|--------|
| `internal/domains/conversations/` or `internal/domains/whatsapp/` | Create – webhook handler, Twilio client, message create/send |
| Webhook route (no auth or auth by signature) | Register POST /api/webhooks/whatsapp |
| Send message use case / handler | Create – POST /api/v1/conversations/:id/messages calling Twilio |
| `api/openapi.yaml` | Add – webhook and send message contract if needed |

## Acceptance criteria

- [ ] POST /api/webhooks/whatsapp with valid Twilio payload and signature returns 200 and creates a message (and conversation/contact if new).
- [ ] Invalid signature returns 401.
- [ ] POST /api/v1/conversations/:id/messages with auth and tenant context sends via Twilio and persists message; response returns message record.
- [ ] Only conversations belonging to the tenant can be sent to.
