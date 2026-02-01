# P3-01: Reviews schema and GBP reviews sync

**Task ID:** P3-01  
**Owner:** Backend  
**Phase:** 3 – Value demonstration

## Objective

Add `reviews` table and a job or endpoint to fetch GBP reviews per location and upsert. Store platform, author, rating, content, reply, replied_at.

## Scope (agent-runnable)

- Migration for `reviews` table per [Strategic Roadmap Part 2.3]: id, location_id, agency_id, platform, platform_review_id, author_name, rating, content, reply, replied_at, sentiment, created_at, fetched_at; unique (platform, platform_review_id).
- RLS: reviews scoped by agency_id (or via location → client → agency).
- Job or endpoint (e.g. POST /api/v1/gbp/sync/:locationId/reviews or cron): fetch reviews from GBP API for location, upsert into `reviews`. Requires GBP tokens (P2-02).

## References

- [FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md](../FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md) – Part 2.3 reviews table
- P2-02 (GBP OAuth and tokens per location)

## Dependencies

- P2-02 (GBP tokens and location linkage).

## Files to create or modify

| File | Action |
|------|--------|
| `migrations/000013_reviews.up.sql` | Create – reviews table, RLS, indexes |
| `migrations/000013_reviews.down.sql` | Create – rollback |
| `internal/domains/gbp/` or `internal/domains/reviews/` | Add – fetch GBP reviews, upsert logic |
| Endpoint or job | Add – trigger sync for a location (or all locations for tenant) |

## Acceptance criteria

- [ ] Migration applies; `reviews` table exists with required columns and unique (platform, platform_review_id).
- [ ] RLS restricts access by agency.
- [ ] Sync job/endpoint fetches GBP reviews for a location and upserts; no duplicate rows for same platform_review_id.
- [ ] Down migration rolls back cleanly.
