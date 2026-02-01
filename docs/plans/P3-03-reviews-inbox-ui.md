# P3-03: Reviews Inbox UI

**Task ID:** P3-03  
**Owner:** Frontend  
**Phase:** 3 – Value demonstration

## Objective

Reviews list and detail in the portal; compose and send reply; show status. Reuse or extend existing business/reviews page.

## Scope (agent-runnable)

- Reviews list page: fetch GET /api/v1/reviews (optional location filter); show rating, author, content, reply status.
- Review detail or inline: click to expand or open detail; show full content and existing reply.
- Compose reply: text input + submit; call POST /api/v1/reviews/:id/reply; show success/error; refresh or update UI.
- Reuse or extend [farohq-portal](farohq-portal) business/reviews page.

## References

- Portal: [app/business/reviews/](../../../farohq-portal/src/app/business/reviews/)
- P3-02 (reviews API)

## Dependencies

- P3-02 (reviews API).

## Files to create or modify

| File | Action |
|------|--------|
| Portal: reviews page and components | Create or extend – list, detail, reply form |
| Portal: API client | Add – fetch reviews, submit reply (with tenant context) |

## Acceptance criteria

- [ ] Reviews page shows list of reviews for the tenant (optionally filtered by location).
- [ ] User can open a review and see full content and existing reply.
- [ ] User can submit a reply; success shows updated reply and status.
- [ ] All requests use tenant context.
