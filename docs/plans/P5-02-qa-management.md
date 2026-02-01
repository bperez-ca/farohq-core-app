# P5-02: Q&A management

**Task ID:** P5-02  
**Owner:** Full-stack  
**Phase:** 5 – Posts, listings, reports & integrations

## Objective

Enable agencies to monitor and respond to Google Business Profile Q&A (Questions and Answers). Questions appear in the FARO dashboard; agents can post owner answers directly.

## Scope (agent-runnable)

- **Q&A sync:** Fetch Q&A from GBP API for each location. Store in `gbp_questions` table (id, location_id, agency_id, gbp_question_id, question_text, author_name, created_at, answers JSONB, owner_answer, owner_answer_at).
- **Sync job:** Periodic job (cron) or on-demand sync to fetch new questions and update existing.
- **Answer API:** `POST /api/v1/questions/:id/answer` with `{ answer_text }`. Calls GBP API to post owner answer; stores locally.
- **List API:** `GET /api/v1/questions?location_id=...` – List questions with answers.
- **UI:**
  - Q&A page: List questions for location; filter by unanswered.
  - Question detail: Show question, existing answers (community + owner); form to post owner answer.
- **Tier gating:** Q&A management at Plus tier and above.

## References

- [Agency-Pricing-Tiers.md](../../Agency-Pricing-Tiers.md) – Q&A management
- [SMB-Pricing-Tiers.md](../../SMB-Pricing-Tiers.md) – Q&A management in Plus+
- [P2-02-gbp-oauth-sync.md](P2-02-gbp-oauth-sync.md) – GBP OAuth
- Google My Business API – Q&A

## Dependencies

- P2-02 (GBP OAuth)
- P1-03 (locations)
- P3-08 (tier enforcement)

## Files to create or modify

| File | Action |
|------|--------|
| `migrations/` | Create – `gbp_questions` table |
| `internal/domains/qa/` | Create – Sync, list, answer use cases and handlers |
| `internal/jobs/` | Add – Q&A sync job |
| `api/openapi.yaml` | Add – Questions list and answer endpoints |
| Portal: Q&A page | Create – List questions, answer form |

## Acceptance criteria

- [ ] Questions are synced from GBP and stored locally.
- [ ] Agency can view questions for a location; filter by unanswered.
- [ ] Agent can post an owner answer; answer is sent to GBP and stored locally.
- [ ] Existing community answers are displayed (read-only).
- [ ] Q&A is scoped to agency (RLS).
- [ ] Q&A management is gated to Plus tier and above.
