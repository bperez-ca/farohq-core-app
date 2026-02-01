# P1-04: Basic agency dashboard

**Task ID:** P1-04  
**Owner:** Frontend  
**Phase:** 1 – MVP foundation

## Objective

Implement or finalize the agency dashboard with at least: tenant summary, client count, location count, and placeholders for “Needs attention” and “Churn watch” (data can be mocked or minimal).

## Scope (agent-runnable)

- Agency dashboard shows: tenant summary, client count, location count.
- Placeholders for “Needs attention” and “Churn watch” (empty or mocked list is acceptable for Phase 1).
- Use [Strategic Roadmap Week 4] checklist for “Basic agency dashboard”.

## References

- [FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md](../FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md) – Week 4 checklist
- Portal: [app/agency/dashboard/](../../../farohq-portal/src/app/agency/dashboard/) (or equivalent)

## Dependencies

- P1-01, P1-03 (tenant context and client/location APIs for real counts).

## Files to create or modify

| File | Action |
|------|--------|
| Portal: agency dashboard page | Implement or finalize – summary, client count, location count |
| Portal: “Needs attention” / “Churn watch” sections | Add – placeholders (e.g. “Coming soon” or empty list) |

## Acceptance criteria

- [ ] Agency dashboard page shows tenant name/summary.
- [ ] Client count and location count are displayed (from API or mocked).
- [ ] “Needs attention” section exists (placeholder or empty).
- [ ] “Churn watch” section exists (placeholder or empty).
- [ ] Page is reachable from agency navigation and uses tenant context.
