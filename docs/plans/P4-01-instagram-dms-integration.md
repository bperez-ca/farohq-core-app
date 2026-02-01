# P4-01: Instagram DMs integration

**Task ID:** P4-01  
**Owner:** Backend + Frontend  
**Phase:** 4 – Enhanced inbox & AI

## Objective

Integrate Instagram Direct Messages into the unified inbox. Connect Instagram Business accounts via Facebook Graph API, receive incoming DMs via webhooks, and enable agents to respond from the FARO inbox.

## Scope (agent-runnable)

- **OAuth flow:** Connect Instagram Business account linked to a Facebook Page. Use Facebook Graph API (Instagram Messaging API requires Facebook Login with `instagram_manage_messages` and `pages_messaging` permissions).
- **Webhook:** Receive incoming Instagram DMs via Facebook webhook (`instagram_messaging` subscription). Validate signature, parse message, create/find contact and conversation, persist message.
- **Send:** Send replies via Instagram Messaging API (`POST /{ig-user-id}/messages`). Create message record, call API, update status.
- **Channel type:** Add `instagram` as a channel in `conversations` table (alongside `whatsapp`, `gbp`).
- **Tier gating:** Instagram DMs available at Plus tier and above (per [Agency-Pricing-Tiers.md](../../Agency-Pricing-Tiers.md)).

## References

- [Agency-Pricing-Tiers.md](../../Agency-Pricing-Tiers.md) – Plus tier includes Instagram DMs
- [SMB-Pricing-Tiers.md](../../SMB-Pricing-Tiers.md) – Plus tier unified inbox channels
- [P2-01-conversations-messages-schema.md](P2-01-conversations-messages-schema.md) – Conversations schema (extend channel enum)
- [P2-03-twilio-whatsapp-webhook.md](P2-03-twilio-whatsapp-webhook.md) – Similar webhook pattern
- Facebook Graph API / Instagram Messaging API docs

## Dependencies

- P2-01 (conversations and messages tables)
- P2-04 (conversations API)
- P3-08 (tier enforcement for Plus+ gating)

## Files to create or modify

| File | Action |
|------|--------|
| `internal/domains/instagram/` | Create – OAuth flow, webhook handler, send message, IG API client |
| `migrations/` | Modify – Add `instagram` to channel enum if needed; add `instagram_tokens` table or extend `channel_tokens` |
| `api/openapi.yaml` | Add – Instagram OAuth and webhook endpoints |
| Webhook route | Register `POST /api/webhooks/instagram` (signature validation) |
| `internal/domains/conversations/` | Extend – Support `instagram` channel in conversation creation and listing |
| Portal: settings/integrations | Add – Connect Instagram account UI |
| Portal: inbox | Extend – Display Instagram conversations with IG icon/badge |

## Acceptance criteria

- [ ] Agency can connect an Instagram Business account via OAuth (Facebook Login flow).
- [ ] Incoming Instagram DMs create conversations with `channel = 'instagram'` and persist messages.
- [ ] Agents can reply to Instagram DMs from the inbox; replies are sent via IG Messaging API.
- [ ] Webhook validates Facebook signature; invalid signature returns 401.
- [ ] Instagram channel is gated to Plus tier and above; lower tiers see "Upgrade" prompt.
- [ ] Tokens stored securely (encrypted at rest); RLS restricts access by agency.
