# P5-01: GBP Posts and scheduling

**Task ID:** P5-01  
**Owner:** Full-stack  
**Phase:** 5 – Posts, listings, reports & integrations

## Objective

Enable agencies to create, schedule, and publish posts to Google Business Profile. Posts can include text, images, CTAs, and events/offers. Scheduled posts are published automatically at the specified time.

## Scope (agent-runnable)

- **Posts table:** `gbp_posts` (id, location_id, agency_id, type `update | event | offer`, title, content, image_url, cta_type, cta_url, event_start, event_end, offer_terms, status `draft | scheduled | published | failed`, scheduled_at, published_at, gbp_post_id, created_at, updated_at).
- **CRUD API:**
  - `GET /api/v1/posts?location_id=...` – List posts for location.
  - `POST /api/v1/posts` – Create post (draft or scheduled).
  - `PUT /api/v1/posts/:id` – Update draft/scheduled post.
  - `DELETE /api/v1/posts/:id` – Delete post.
  - `POST /api/v1/posts/:id/publish` – Publish immediately.
- **GBP integration:** Use Google My Business API (LocalPosts) to create/update posts. Requires GBP OAuth tokens (P2-02).
- **Scheduler:** Background job (cron or queue) checks for posts where `status = 'scheduled'` and `scheduled_at <= now()`; publishes to GBP; updates status.
- **Image upload:** Upload images to S3; pass URL to GBP API (or use media upload API).
- **Tier gating:** GBP posts available at Plus tier and above; bulk scheduling at Pro+.

## References

- [Agency-Pricing-Tiers.md](../../Agency-Pricing-Tiers.md) – GBP post scheduling
- [SMB-Pricing-Tiers.md](../../SMB-Pricing-Tiers.md) – Posts & Offers features
- [P2-02-gbp-oauth-sync.md](P2-02-gbp-oauth-sync.md) – GBP OAuth tokens
- Google My Business API – LocalPosts

## Dependencies

- P2-02 (GBP OAuth and tokens)
- P1-03 (locations)
- P3-08 (tier enforcement)

## Files to create or modify

| File | Action |
|------|--------|
| `migrations/` | Create – `gbp_posts` table |
| `internal/domains/posts/` | Create – CRUD, publish use case, GBP LocalPosts client |
| `internal/jobs/` or `cmd/worker/` | Add – Scheduled post publisher job |
| `api/openapi.yaml` | Add – Posts CRUD and publish endpoints |
| Portal: posts page | Create – List posts, create/edit post form (text, image, CTA, schedule) |
| Portal: post detail | Create – View post status, preview |

## Acceptance criteria

- [ ] Agency can create a post for a location with content, optional image, and CTA.
- [ ] Post can be saved as draft or scheduled for a future time.
- [ ] Scheduled posts are automatically published at the scheduled time.
- [ ] Published posts appear on the location's GBP.
- [ ] Post status (draft/scheduled/published/failed) is tracked and displayed.
- [ ] Failed posts show error message; can be retried.
- [ ] Posts are scoped to agency (RLS).
