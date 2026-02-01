# P2-05: Voice transcription (Whisper)

**Task ID:** P2-05  
**Owner:** Backend  
**Phase:** 2 – Core integrations

## Objective

On inbound WhatsApp voice message: download media, call Whisper API, store transcript on message. Support language hint (ES/PT/EN). Log usage for billing.

## Scope (agent-runnable)

- When processing an inbound WhatsApp message with voice/media: if media type is audio, download file, call OpenAI Whisper API, store result in `messages.transcript`.
- Language hint: pass ES/PT/EN (or auto) per [Strategic Roadmap Part 4.2] (LATAM).
- Log usage for billing (e.g. `usage_events` table or equivalent: event_type `voice_transcription`, quantity, unit_cost).
- Per [Strategic Roadmap Part 4.2]: cost $0.006/minute; cap or limit audio length (e.g. max 5 min) if specified.

## References

- [FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md](../FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md) – Part 4.2 AI cost management
- OpenAI Whisper API docs
- P2-03 webhook flow – extend to detect voice and run transcription after persisting message

## Dependencies

- P2-01 (messages table with transcript column), P2-03 (inbound webhook).

## Files to create or modify

| File | Action |
|------|--------|
| `internal/domains/conversations/` or shared service | Add – Whisper client, download media, transcribe, update message.transcript |
| Webhook handler (P2-03) | Extend – after creating message, if voice, enqueue or run transcription |
| `usage_events` table or equivalent | Create or use – log voice_transcription events |
| Config / env | Document – OPENAI_API_KEY (or similar) for Whisper |

## Acceptance criteria

- [ ] Inbound WhatsApp voice message results in message row with `transcript` populated (async or sync).
- [ ] Language hint can be passed (e.g. ES/PT/EN); optional for MVP.
- [ ] Usage event is recorded for billing (voice_transcription, quantity, unit_cost).
- [ ] Audio length is capped (e.g. 5 min) to control cost; longer audio truncated or rejected.
