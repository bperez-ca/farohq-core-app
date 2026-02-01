# P3-04: Diagnostics domain and scores

**Task ID:** P3-04  
**Owner:** Backend  
**Phase:** 3 – Value demonstration

## Objective

Add diagnostics domain: migrations for `diagnostics` (and `usage_events` if not done); compute presence/reviews/speed scores and estimated monthly loss per [Strategic Roadmap Part 4.4]. Endpoints: create/get diagnostic, get shareable by token.

## Scope (agent-runnable)

- Migration: `diagnostics` table (id, agency_id, client_id, share_token, presence_score, reviews_score, speed_score, estimated_monthly_loss, view_count, created_at); `usage_events` if not yet present (agency_id, event_type, quantity, unit_cost, metadata, created_at).
- Score formulas per [Strategic Roadmap Part 4.4]:
  - Presence: weighted sum of platform status (e.g. GBP, Meta, etc.).
  - Reviews: rating component + velocity + recency.
  - Speed: based on avg reply time buckets.
  - Estimated monthly loss: presence_loss + reviews_loss + speed_loss (formula in roadmap).
- Endpoints: create diagnostic (compute scores), get diagnostic by id (tenant-scoped), get by share_token (public, no PII in URL).

## References

- [FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md](../FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md) – Part 4.4 Diagnostic score calculations
- Part 2.3 diagnostics and usage_events schema

## Dependencies

- P2-01/P3-01 (reviews and conversations data for score inputs); locations and clients exist.

## Files to create or modify

| File | Action |
|------|--------|
| `migrations/000014_diagnostics_usage_events.up.sql` | Create – diagnostics, usage_events if needed |
| `internal/domains/diagnostics/` | Create – domain model, score calculation, create/get/get-by-token |
| `api/openapi.yaml` | Add – diagnostic endpoints (create, get, get-by-token) |
| Router | Register – protected create/get; public get-by-token (no auth) |

## Acceptance criteria

- [ ] Diagnostic can be created for a client/location; scores and loss are computed from current data.
- [ ] GET by id returns diagnostic only if tenant owns it; GET by share_token returns diagnostic without auth (for public share page).
- [ ] Score formulas match Strategic Roadmap Part 4.4 (or document deviations).
- [ ] usage_events table exists and can record billing-relevant events.
