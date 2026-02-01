# P3-02: Reviews API and reply

**Task ID:** P3-02  
**Owner:** Backend  
**Phase:** 3 – Value demonstration

## Objective

Expose GET /api/v1/reviews (filter by location/tenant) and POST /api/v1/reviews/:id/reply (submit reply to GBP). OpenAPI + RLS.

## Scope (agent-runnable)

- `GET /api/v1/reviews` – query params: location_id, agency (implicit from tenant), status, limit, offset; response: list of reviews and total; RLS by agency.
- `POST /api/v1/reviews/:id/reply` – body: reply text; call GBP API to post reply; update `reviews.reply` and `reviews.replied_at`; tenant must own the review (via location/agency).

## References

- [FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md](../FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md) – Part 2.3 reviews
- Google Business Profile API – reply to review

## Dependencies

- P3-01 (reviews table and GBP sync); P2-02 (GBP tokens).

## Files to create or modify

| File | Action |
|------|--------|
| `internal/domains/reviews/` (or gbp) | Add – list reviews, get review, reply use case and HTTP handlers |
| `api/openapi.yaml` | Add – reviews list and reply endpoints |
| Router | Register routes with auth and tenant context |

## Acceptance criteria

- [ ] GET /api/v1/reviews returns only tenant’s reviews; supports location_id and pagination.
- [ ] POST /api/v1/reviews/:id/reply returns 404 if review not in tenant; on success, reply is sent to GBP and stored in DB.
- [ ] OpenAPI spec includes reviews endpoints and schemas.
