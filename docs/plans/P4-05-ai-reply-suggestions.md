# P4-05: AI reply suggestions

**Task ID:** P4-05  
**Owner:** Backend + Frontend  
**Phase:** 4 – Enhanced inbox & AI

## Objective

Provide AI-generated reply suggestions for inbox messages and reviews. When an agent views a conversation or review, they can request 3 AI-suggested replies (formal, friendly, short) to speed up response time.

## Scope (agent-runnable)

- **API endpoint:** `POST /api/v1/ai/suggest-reply` with body `{ conversation_id | review_id, context }`. Returns 3 suggested replies with different tones (formal, friendly, concise).
- **LLM integration:** Use OpenAI GPT-3.5/4 (or configurable). Pass conversation history or review content as context. System prompt defines tone variations.
- **Context building:** For conversations, include last N messages. For reviews, include rating, content, business name/type.
- **Localization:** Support ES/PT/EN based on conversation language or location settings.
- **AI credits:** Each suggestion request consumes AI credits. Log to `usage_events` (P3-04/P3-07).
- **Tier gating:** AI suggestions available at Plus tier and above (per [Agency-Pricing-Tiers.md](../../Agency-Pricing-Tiers.md)).

## References

- [Agency-Pricing-Tiers.md](../../Agency-Pricing-Tiers.md) – AI credits by tier (300/1000/5000/15000)
- [SMB-Pricing-Tiers.md](../../SMB-Pricing-Tiers.md) – AI reply suggestions feature
- [P2-05-voice-transcription-whisper.md](P2-05-voice-transcription-whisper.md) – Similar OpenAI integration pattern
- [P4-06-ai-credits-tracking.md](P4-06-ai-credits-tracking.md) – AI credits tracking
- OpenAI API docs

## Dependencies

- P2-04 (conversations API for context)
- P3-02 (reviews API for context)
- P3-04/P3-07 (usage_events for credits)
- P3-08 (tier enforcement)

## Files to create or modify

| File | Action |
|------|--------|
| `internal/domains/ai/` | Create – AI service, suggest-reply use case, OpenAI client |
| `api/openapi.yaml` | Add – `POST /api/v1/ai/suggest-reply` endpoint and schema |
| Config / env | Document – `OPENAI_API_KEY`, model selection |
| `usage_events` | Log – `ai_suggestion` events with token count |
| Portal: inbox thread | Add – "Suggest reply" button; show 3 suggestions; click to populate input |
| Portal: reviews detail | Add – "Suggest reply" button for review responses |

## Acceptance criteria

- [ ] Agent can request AI suggestions for a conversation; 3 replies returned (formal, friendly, concise).
- [ ] Agent can request AI suggestions for a review; 3 replies returned.
- [ ] Suggestions respect conversation language (ES/PT/EN).
- [ ] Each request logs AI usage to `usage_events` with token count.
- [ ] Agent can click a suggestion to populate the reply input.
- [ ] AI suggestions are gated to Plus tier and above; lower tiers see "Upgrade" prompt.
- [ ] Error handling for OpenAI API failures (rate limit, timeout).
