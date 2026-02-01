# P3-08: Tier enforcement and limits

**Task ID:** P3-08  
**Owner:** Full-stack  
**Phase:** 3 – Value demonstration

## Objective

Enforce client/location/message limits by tier (e.g. from [Strategic Roadmap Part 1.2]); block or warn in API and portal when over limit.

## Scope (agent-runnable)

- Limits per tier (example from roadmap): Agency Plan 10 clients; Growth 50; Scale 200. Per-client add-on. Message or diagnostic limits if defined.
- Backend: before creating client/location (or sending message, creating diagnostic), check current usage vs tier limits; return 403 or 402 with clear error if over limit.
- Portal: before inviting or adding client, check limit; show upgrade CTA when at or over limit. Optional: display “X of Y clients used” on dashboard.
- Tier source: from Stripe subscription (P3-07) or from agencies.tier column; keep in sync with billing.

## References

- [FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md](../FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md) – Part 1.2 Pricing (tiers and limits)
- P3-07 (billing and tier/plan state)

## Dependencies

- P3-07 (tier/plan available from Stripe or DB).

## Files to create or modify

| File | Action |
|------|--------|
| Backend: tenant or billing service | Add – get tier and limits for agency; check client/location count vs limit |
| Backend: create client, create location (and optionally message/diagnostic) | Add – enforce limits; return 403/402 with message |
| Portal: create client/location flows | Add – check limit before submit; show upgrade CTA when at limit |
| Portal: dashboard or settings | Optional – show “X of Y clients” and link to upgrade |

## Acceptance criteria

- [ ] Creating a client when at tier limit returns 403 (or 402) with clear message; creation is blocked.
- [ ] Creating a location (if limited by tier) is blocked when over limit.
- [ ] Portal shows upgrade or limit message when user is at limit (e.g. on “Add client” or dashboard).
- [ ] Tier and limits are read from billing/Stripe or agencies table consistently.
