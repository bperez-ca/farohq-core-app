# P4-04: Web chat widget

**Task ID:** P4-04  
**Owner:** Full-stack  
**Phase:** 4 – Enhanced inbox & AI

## Objective

Provide an embeddable web chat widget that SMBs can add to their websites. Visitor messages appear in the unified inbox; agents can respond in real-time.

## Scope (agent-runnable)

- **Widget:** Lightweight JavaScript widget (iframe or inline) that SMBs embed via script tag. Configurable colors (match SMB branding), position, welcome message.
- **Widget ID:** Each location gets a unique widget ID. Widget loads config (branding, welcome message) from API.
- **WebSocket/SSE:** Real-time communication between widget and backend. Messages sent from widget appear in inbox immediately; agent replies appear in widget.
- **Visitor identification:** Anonymous by default (session ID). Optionally collect name/email before chat starts (pre-chat form).
- **Conversation creation:** First message from visitor creates conversation with `channel = 'webchat'`.
- **Persistence:** Messages stored in `messages` table; conversation persists even if visitor closes browser (resume via cookie/localStorage).
- **Tier gating:** Web chat available at Plus tier and above.

## References

- [Agency-Pricing-Tiers.md](../../Agency-Pricing-Tiers.md) – Plus tier includes Web chat
- [SMB-Pricing-Tiers.md](../../SMB-Pricing-Tiers.md) – Plus tier unified inbox channels
- [P2-01-conversations-messages-schema.md](P2-01-conversations-messages-schema.md) – Conversations schema
- [P2-06-realtime-updates.md](P2-06-realtime-updates.md) – Real-time infrastructure

## Dependencies

- P2-01 (conversations and messages tables)
- P2-04 (conversations API)
- P2-06 (real-time updates preferred, or polling fallback)
- P3-08 (tier enforcement)

## Files to create or modify

| File | Action |
|------|--------|
| `internal/domains/webchat/` | Create – Widget config API, message receive/send, WebSocket/SSE handler |
| `migrations/` | Modify – Add `webchat` to channel enum; add `webchat_configs` table (widget_id, location_id, settings) |
| `api/openapi.yaml` | Add – Widget config, chat endpoints, WebSocket/SSE spec |
| `packages/webchat-widget/` or `public/widget.js` | Create – Embeddable widget (JS + CSS) |
| Portal: settings/integrations | Add – Web chat setup (embed code, customize widget) |
| Portal: inbox | Extend – Display webchat conversations with chat icon |

## Acceptance criteria

- [ ] SMB can generate embed code for their website from portal settings.
- [ ] Widget loads on SMB website with correct branding (colors from location/agency theme).
- [ ] Visitor messages create conversations with `channel = 'webchat'` and persist messages.
- [ ] Agent replies in inbox appear in widget in real-time (or near-real-time).
- [ ] Visitor can resume chat after closing/reopening browser (within session window).
- [ ] Pre-chat form (optional) collects visitor name/email before chat starts.
- [ ] Web chat is gated to Plus tier and above.
