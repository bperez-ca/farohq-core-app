# P3-05: Shared diagnostic page (public)

**Task ID:** P3-05  
**Owner:** Full-stack  
**Phase:** 3 – Value demonstration

## Objective

Public page (e.g. `/share/diagnostic/:token`) that loads diagnostic by token and displays scores and loss; view tracking. Wire to real API; enforce no PII in URL.

## Scope (agent-runnable)

- Public route: `/share/diagnostic/[token]` (portal). No auth required.
- Fetch diagnostic by token: GET /api/v1/diagnostics/share/:token (or similar public endpoint) returning scores and estimated loss; no PII in response/URL.
- Display: presence score, reviews score, speed score, estimated monthly loss; optional CTA.
- View tracking: call backend to increment view_count when page is viewed (optional idempotency by session or IP to avoid abuse).
- Portal already has [share/diagnostic/[token]](farohq-portal/src/app/share/diagnostic); wire to real API and enforce no PII in URL.

## References

- Portal: [app/share/diagnostic/[token]/](../../../farohq-portal/src/app/share/diagnostic/)
- P3-04 (diagnostics API, get-by-token)

## Dependencies

- P3-04 (diagnostics domain and get-by-token endpoint).

## Files to create or modify

| File | Action |
|------|--------|
| Portal: share/diagnostic/[token] page | Wire – fetch from API by token, display scores and loss |
| Backend: view_count increment | Optional – endpoint or same get-by-token with side-effect |
| Ensure token in URL is opaque (no PII) | Verify – token is share_token from diagnostics table |

## Acceptance criteria

- [ ] Visiting /share/diagnostic/:token loads diagnostic from API and shows scores and loss.
- [ ] URL contains only token (no client name, email, or other PII).
- [ ] View count is incremented when page is viewed (optional; document if skipped).
- [ ] Unused or invalid token returns 404 or friendly “not found” message.
