# P3-07: Usage and billing (Stripe)

**Task ID:** P3-07  
**Owner:** Backend  
**Phase:** 3 – Value demonstration

## Objective

Implement usage tracking and Stripe billing: subscription (e.g. $99/mo + $49/client), webhooks, tier enforcement. Optional: invoice generation. Per [Strategic Roadmap Week 12].

## Scope (agent-runnable)

- `usage_events` (or equivalent): record billing-relevant events (e.g. message count, voice transcription, diagnostic views); already partially specified in P2-05 and P3-04.
- Stripe integration: create customer per tenant (agency); subscription for base plan ($99/mo) and per-client add-on ($49/mo); handle webhooks (invoice.paid, customer.subscription.updated/deleted, etc.).
- Tier enforcement: read plan and limits from Stripe or local cache; used by P3-08 to enforce client/location/message limits.
- Optional: invoice generation (Stripe Invoicing or dashboard); document how agencies view invoices.

## References

- [FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md](../FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md) – Part 1.2 Pricing, Week 12 Billing
- Stripe: Subscriptions, Customer, Webhooks, optional Invoicing

## Dependencies

- P1-01 (tenant/agency); usage_events or equivalent (P2-05/P3-04).

## Files to create or modify

| File | Action |
|------|--------|
| `internal/domains/billing/` or equivalent | Create – Stripe client, create customer, create/update subscription, webhook handler |
| `migrations/` | Add – stripe_customer_id on agencies if not present; usage_events if not done |
| Webhook route | POST /api/webhooks/stripe – verify signature, handle events |
| Config / env | Document – STRIPE_SECRET_KEY, STRIPE_WEBHOOK_SECRET, price IDs |
| P3-08 | Consume – tier/limits from billing for enforcement |

## Acceptance criteria

- [ ] Agency can have a Stripe customer and subscription (base + per-client); webhook updates local state.
- [ ] usage_events (or equivalent) records events used for billing or limits.
- [ ] Tier/plan is available for enforcement (P3-08); e.g. agency.tier or subscription metadata.
- [ ] Webhook signature is verified; idempotency handled for duplicate events.
