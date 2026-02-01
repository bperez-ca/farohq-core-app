# P4-03: Email inbox integration

**Task ID:** P4-03  
**Owner:** Backend + Frontend  
**Phase:** 4 – Enhanced inbox & AI

## Objective

Integrate email into the unified inbox. Allow SMBs to connect their email (via IMAP/SMTP or email forwarding) so customer emails appear in the inbox and agents can respond directly.

## Scope (agent-runnable)

- **Email ingestion options:**
  - **Option A (Forwarding):** SMB forwards emails to a unique FARO address (e.g., `inbox-{location_id}@mail.farohq.com`). Use inbound email service (SendGrid Inbound Parse, Mailgun Routes, or Postmark) to receive and parse.
  - **Option B (IMAP/OAuth):** Connect email via OAuth (Gmail, Outlook) or IMAP credentials. Poll or use push notifications (Gmail Push, MS Graph).
- **Email parsing:** Extract sender, subject, body (plain text + HTML), attachments (store in S3, link in message).
- **Conversation threading:** Group emails by thread (In-Reply-To, References headers, or subject + sender).
- **Send:** Reply via SMTP or email API. Include proper threading headers.
- **Channel type:** Add `email` as a channel in `conversations` table.
- **Tier gating:** Email available at Plus tier and above.

## References

- [Agency-Pricing-Tiers.md](../../Agency-Pricing-Tiers.md) – Plus tier includes Email
- [SMB-Pricing-Tiers.md](../../SMB-Pricing-Tiers.md) – Plus tier unified inbox channels
- [P2-01-conversations-messages-schema.md](P2-01-conversations-messages-schema.md) – Conversations schema
- SendGrid Inbound Parse / Mailgun / Postmark docs
- Gmail API / Microsoft Graph API docs

## Dependencies

- P2-01 (conversations and messages tables)
- P2-04 (conversations API)
- P3-08 (tier enforcement)

## Files to create or modify

| File | Action |
|------|--------|
| `internal/domains/email/` | Create – Inbound webhook/parser, SMTP send, email API client |
| `migrations/` | Modify – Add `email` to channel enum; add email config table (forwarding address or OAuth tokens) |
| `api/openapi.yaml` | Add – Email configuration and inbound webhook endpoints |
| Webhook route | Register `POST /api/webhooks/email` for inbound parse |
| `internal/domains/conversations/` | Extend – Support `email` channel, threading logic |
| `messages` table | Extend – Add `subject`, `thread_id` or use existing fields |
| Portal: settings/integrations | Add – Configure email (forwarding address or connect Gmail/Outlook) |
| Portal: inbox | Extend – Display email conversations with subject line, email icon |

## Acceptance criteria

- [ ] SMB can configure email forwarding or connect email account.
- [ ] Incoming emails create conversations with `channel = 'email'` and persist messages with subject and body.
- [ ] Email threads are grouped correctly (replies to same conversation).
- [ ] Agents can reply to emails from the inbox; replies are sent with proper threading headers.
- [ ] Attachments are stored and linked in messages.
- [ ] Email channel is gated to Plus tier and above.
- [ ] Email credentials/tokens stored securely.
