# P3-06: Dashboard KPIs and "Needs attention"

**Task ID:** P3-06  
**Owner:** Frontend  
**Phase:** 3 – Value demonstration

## Objective

Agency dashboard: real KPIs (locations, clients, unread/recent conversations, review response rate, etc.); “Needs attention” list (e.g. unanswered reviews, old conversations); optional churn watch. Data from existing + new APIs.

## Scope (agent-runnable)

- Agency dashboard (extend P1-04): replace or augment placeholders with real data.
- KPIs: tenant summary, client count, location count; unread or recent conversation count; review count and response rate (e.g. % replied); optional: leads or revenue metrics if available.
- “Needs attention”: list of items requiring action – e.g. reviews without reply, conversations with no reply in 24h; link to inbox or reviews.
- Optional: “Churn watch” – e.g. clients with no activity or at-risk criteria (can be minimal or placeholder).
- Data from existing APIs (tenants, clients, locations, seat usage) and new (conversations, reviews) as implemented in Phase 2–3.

## References

- [FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md](../FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md) – Week 11 Dashboard & KPIs
- P1-04 (basic agency dashboard); P2-04 (conversations); P3-02 (reviews)

## Dependencies

- P1-04, P2-04, P3-02 (APIs for conversations and reviews).

## Files to create or modify

| File | Action |
|------|--------|
| Portal: agency dashboard page | Extend – KPI cards with real API data |
| Portal: “Needs attention” component | Add – fetch unanswered reviews / stale conversations; links to inbox/reviews |
| Backend (optional) | Add – GET /api/v1/dashboard/summary or similar aggregating counts and attention items |
| Portal: “Churn watch” (optional) | Add – placeholder or minimal at-risk list |

## Acceptance criteria

- [ ] Dashboard shows real client count, location count (from existing APIs).
- [ ] Dashboard shows conversation and/or review metrics (e.g. unread count, review response rate) when APIs exist.
- [ ] “Needs attention” lists at least one type of item (e.g. unanswered reviews) with link to fix.
- [ ] “Churn watch” section exists (data or placeholder).
