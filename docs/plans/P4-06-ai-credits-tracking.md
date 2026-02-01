# P4-06: AI credits tracking and tier limits

**Task ID:** P4-06  
**Owner:** Backend + Frontend  
**Phase:** 4 – Enhanced inbox & AI

## Objective

Track AI credit usage per client/agency and enforce tier-based limits. Each tier includes a monthly AI credit quota; usage beyond quota is blocked or charged as overage via the usage wallet.

## Scope (agent-runnable)

- **AI credit allocation by tier:**
  - Basic: 300 credits/month
  - Plus: 1,000 credits/month
  - Pro: 5,000 credits/month
  - Elite: 15,000 credits/month
- **Credit consumption:** Each AI operation costs credits (e.g., reply suggestion ~10-50 credits based on tokens, transcription based on audio length).
- **Usage tracking:** Log each AI operation to `usage_events` with `event_type = 'ai_credits'`, quantity (credits used), metadata (operation type, token count).
- **Limit enforcement:** Before AI operation, check current period usage vs tier limit. If at/over limit, return 402/403 with "AI credits exhausted" message and upgrade CTA.
- **Overage (optional):** If usage wallet has balance, allow overage and deduct from wallet. Otherwise block.
- **Reset:** Credits reset monthly (per billing cycle or calendar month).
- **Dashboard:** Show current AI credit usage on agency/client dashboard ("X of Y credits used this month").

## References

- [Agency-Pricing-Tiers.md](../../Agency-Pricing-Tiers.md) – AI credits by tier
- [SMB-Pricing-Tiers.md](../../SMB-Pricing-Tiers.md) – AI credits by tier
- [P3-04-diagnostics-domain-scores.md](P3-04-diagnostics-domain-scores.md) – usage_events table
- [P3-07-usage-billing-stripe.md](P3-07-usage-billing-stripe.md) – Billing integration
- [P4-05-ai-reply-suggestions.md](P4-05-ai-reply-suggestions.md) – AI suggestion credits

## Dependencies

- P3-04 (usage_events table)
- P3-07 (billing and tier info)
- P3-08 (tier enforcement pattern)

## Files to create or modify

| File | Action |
|------|--------|
| `internal/domains/ai/` | Add – Credit check before operations, credit deduction after |
| `internal/domains/billing/` or `usage/` | Add – Get AI credits used this period, get limit by tier |
| `usage_events` | Use – Log AI credit usage events |
| `api/openapi.yaml` | Add – `GET /api/v1/usage/ai-credits` endpoint (current usage and limit) |
| Portal: dashboard | Add – AI credits usage card ("X of Y credits used") |
| Portal: AI operations | Add – Show "credits exhausted" message when at limit |

## Acceptance criteria

- [ ] AI operations log credit usage to `usage_events`.
- [ ] API enforces tier limits; returns 402/403 when credits exhausted with clear message.
- [ ] `GET /api/v1/usage/ai-credits` returns current usage and limit for the period.
- [ ] Dashboard shows AI credit usage ("X of Y credits used this month").
- [ ] Credits reset monthly (document reset logic: billing cycle or calendar month).
- [ ] Optional: Overage allowed if usage wallet has balance; deducted from wallet.
