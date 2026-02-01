# P5-08: REST API (public) and webhooks

**Task ID:** P5-08  
**Owner:** Backend  
**Phase:** 5 – Posts, listings, reports & integrations

## Objective

Expose a public REST API for Scale+ tier agencies and implement outbound webhooks for key events. Enables agencies to integrate FARO data into their own systems.

## Scope (agent-runnable)

- **Public API:**
  - Subset of existing endpoints exposed with API key authentication.
  - Endpoints: Locations, Conversations, Messages, Reviews, Leads.
  - Rate limiting: 1,000 requests/hour (Scale), custom (Enterprise).
  - API key management: Generate, rotate, revoke keys per agency.
- **API key auth:**
  - `api_keys` table (id, agency_id, key_hash, name, scopes, rate_limit, last_used_at, created_at, revoked_at).
  - Middleware: Validate API key, extract agency context, enforce rate limit.
- **Webhooks:**
  - `webhook_subscriptions` table (id, agency_id, url, events JSONB, secret, enabled, created_at).
  - Events: `conversation.created`, `message.received`, `review.created`, `lead.converted`, `sync.failed`.
  - Delivery: On event, queue webhook payload; deliver with retry logic (exponential backoff).
  - Signature: Sign payload with HMAC (webhook secret); include in header for verification.
- **Webhook management API:**
  - `POST /api/v1/webhooks` – Create subscription.
  - `GET /api/v1/webhooks` – List subscriptions.
  - `DELETE /api/v1/webhooks/:id` – Delete subscription.
  - `POST /api/v1/webhooks/:id/test` – Send test event.
- **Documentation:** OpenAPI spec covers public API; webhook event schemas documented.
- **Tier gating:** API access at Scale tier and above; webhooks at Scale+.

## References

- [Agency-Pricing-Tiers.md](../../Agency-Pricing-Tiers.md) – REST API and webhooks at Scale+
- [SMB-Pricing-Tiers.md](../../SMB-Pricing-Tiers.md) – API access via agency
- Standard API key and webhook patterns

## Dependencies

- P1-01 (auth and tenant context)
- P2-04, P3-02 (existing APIs to expose)
- P3-08 (tier enforcement)

## Files to create or modify

| File | Action |
|------|--------|
| `migrations/` | Create – `api_keys`, `webhook_subscriptions`, `webhook_deliveries` tables |
| `internal/domains/api/` | Create – API key management, key auth middleware |
| `internal/domains/webhooks/` | Create – Subscription CRUD, event dispatcher, delivery with retry |
| `internal/jobs/` | Add – Webhook delivery worker |
| `api/openapi.yaml` | Add – API key and webhook management endpoints; mark public API endpoints |
| Portal: settings/api | Create – API key management (generate, view, revoke) |
| Portal: settings/webhooks | Create – Webhook subscription management |

## Acceptance criteria

- [ ] Agency can generate and revoke API keys.
- [ ] API key authentication works for designated public endpoints.
- [ ] Rate limiting is enforced per API key.
- [ ] Agency can create webhook subscriptions for events.
- [ ] Events trigger webhook delivery with signed payload.
- [ ] Failed deliveries are retried with backoff.
- [ ] Test webhook endpoint sends sample event.
- [ ] API and webhooks are gated to Scale tier and above.
