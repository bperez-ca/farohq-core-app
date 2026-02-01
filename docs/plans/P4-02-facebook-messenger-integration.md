# P4-02: Facebook Messenger integration

**Task ID:** P4-02  
**Owner:** Backend + Frontend  
**Phase:** 4 – Enhanced inbox & AI

## Objective

Integrate Facebook Messenger into the unified inbox. Connect Facebook Pages via Facebook Graph API, receive incoming messages via webhooks, and enable agents to respond from the FARO inbox.

## Scope (agent-runnable)

- **OAuth flow:** Connect Facebook Page. Use Facebook Login with `pages_messaging` and `pages_manage_metadata` permissions. Store page access token per location/agency.
- **Webhook:** Receive incoming Messenger messages via Facebook webhook (`messages` event on Page). Validate signature, parse message, create/find contact and conversation, persist message.
- **Send:** Send replies via Messenger Send API (`POST /{page-id}/messages`). Create message record, call API, update status.
- **Channel type:** Add `facebook` as a channel in `conversations` table.
- **Tier gating:** Facebook Messenger available at Plus tier and above.

## References

- [Agency-Pricing-Tiers.md](../../Agency-Pricing-Tiers.md) – Plus tier includes Facebook Messenger
- [SMB-Pricing-Tiers.md](../../SMB-Pricing-Tiers.md) – Plus tier unified inbox channels
- [P2-01-conversations-messages-schema.md](P2-01-conversations-messages-schema.md) – Conversations schema
- [P4-01-instagram-dms-integration.md](P4-01-instagram-dms-integration.md) – Similar Facebook Graph API pattern
- Facebook Messenger Platform docs

## Dependencies

- P2-01 (conversations and messages tables)
- P2-04 (conversations API)
- P3-08 (tier enforcement)
- P4-01 (shares Facebook OAuth infrastructure)

## Files to create or modify

| File | Action |
|------|--------|
| `internal/domains/facebook/` or extend `internal/domains/instagram/` | Create/extend – OAuth flow, webhook handler, send message, FB API client |
| `migrations/` | Modify – Add `facebook` to channel enum; add `facebook_page_tokens` table or extend shared token table |
| `api/openapi.yaml` | Add – Facebook OAuth and webhook endpoints |
| Webhook route | Register `POST /api/webhooks/facebook` (or combined with Instagram if using same endpoint) |
| `internal/domains/conversations/` | Extend – Support `facebook` channel |
| Portal: settings/integrations | Add – Connect Facebook Page UI |
| Portal: inbox | Extend – Display Facebook conversations with FB icon/badge |

## Acceptance criteria

- [ ] Agency can connect a Facebook Page via OAuth.
- [ ] Incoming Messenger messages create conversations with `channel = 'facebook'` and persist messages.
- [ ] Agents can reply to Messenger from the inbox; replies are sent via FB Send API.
- [ ] Webhook validates Facebook signature; invalid signature returns 401.
- [ ] Facebook channel is gated to Plus tier and above.
- [ ] Page tokens stored securely; RLS restricts access by agency.
