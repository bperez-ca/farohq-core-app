# P5-03: Content calendar UI

**Task ID:** P5-03  
**Owner:** Frontend  
**Phase:** 5 – Posts, listings, reports & integrations

## Objective

Provide a visual content calendar for managing scheduled posts across locations. Agencies can view, create, edit, and reschedule posts from a calendar interface.

## Scope (agent-runnable)

- **Calendar view:**
  - Month/week view showing scheduled and published posts.
  - Color-coded by status (draft, scheduled, published, failed).
  - Filter by location (single or all).
- **Interactions:**
  - Click on a date to create a new post for that date.
  - Click on a post to view/edit details.
  - Drag-and-drop to reschedule (update `scheduled_at`).
- **Multi-location:**
  - Agency view: Show posts for all locations (with location label).
  - Location view: Show posts for selected location only.
- **Integration:** Uses P5-01 posts API; no new backend endpoints required (calendar is a UI layer).
- **Tier gating:** Content calendar at Plus tier and above; bulk scheduling at Pro+.

## References

- [Agency-Pricing-Tiers.md](../../Agency-Pricing-Tiers.md) – Content calendar
- [SMB-Pricing-Tiers.md](../../SMB-Pricing-Tiers.md) – Content calendar in Pro+
- [P5-01-gbp-posts-scheduling.md](P5-01-gbp-posts-scheduling.md) – Posts API

## Dependencies

- P5-01 (posts CRUD and scheduling API)

## Files to create or modify

| File | Action |
|------|--------|
| Portal: calendar page | Create – Content calendar component (month/week view) |
| Portal: calendar components | Create – CalendarGrid, CalendarDay, PostCard |
| Portal: post create/edit modal | Reuse/extend from P5-01 posts page |

## Acceptance criteria

- [ ] Calendar displays scheduled and published posts in month/week view.
- [ ] Posts are color-coded by status.
- [ ] Clicking a date opens create post modal for that date.
- [ ] Clicking a post opens edit modal.
- [ ] Drag-and-drop reschedules post (updates `scheduled_at` via API).
- [ ] Calendar can be filtered by location.
- [ ] Calendar is gated to Plus tier and above.
