# FARO MVP Implementation Plans

Each **todo/task** from the FARO MVP Implementation Plan has a dedicated **plan MD file**. For every todo check:

1. **Plan first:** Read the corresponding plan MD file in this folder before implementing.
2. **Then implement:** Follow the scope, dependencies, and files listed in that plan.
3. **Verify:** Use the acceptance criteria in the plan to confirm the task is done.

Plan files define: objective, scope (agent-runnable), references, dependencies, files to create/modify, and acceptance criteria.

## Phase 1: MVP foundation (weeks 1–4)

| Done | Task   | Plan file | Description |
|------|--------|-----------|-------------|
| [x]  | P1-01  | [P1-01-auth-multi-tenancy-hardening.md](P1-01-auth-multi-tenancy-hardening.md) | Auth and multi-tenancy hardening |
| [ ]  | P1-02  | [P1-02-white-label-domain-verification.md](P1-02-white-label-domain-verification.md) | White-label and domain verification |
| [ ]  | P1-03  | [P1-03-client-location-apis-ui.md](P1-03-client-location-apis-ui.md) | Client and location APIs and UI |
| [ ]  | P1-04  | [P1-04-basic-agency-dashboard.md](P1-04-basic-agency-dashboard.md) | Basic agency dashboard |

## Phase 2: Core integrations (weeks 5–8)

| Done | Task   | Plan file | Description |
|------|--------|-----------|-------------|
| [ ]  | P2-00  | [SMB-client-creation-and-layers.md](SMB-client-creation-and-layers.md) | SMB creation via Google business search (Places API) + manual fallback; lead vs active lifecycle; per-client layers (Basic/Growth/Premium) |
| [ ]  | P2-01  | [P2-01-conversations-messages-schema.md](P2-01-conversations-messages-schema.md) | Database schema for conversations and messages |
| [ ]  | P2-02  | [P2-02-gbp-oauth-sync.md](P2-02-gbp-oauth-sync.md) | GBP OAuth and basic sync |
| [ ]  | P2-03  | [P2-03-twilio-whatsapp-webhook.md](P2-03-twilio-whatsapp-webhook.md) | Twilio WhatsApp webhook and send |
| [ ]  | P2-04  | [P2-04-conversations-messages-api.md](P2-04-conversations-messages-api.md) | Conversations and messages API |
| [ ]  | P2-05  | [P2-05-voice-transcription-whisper.md](P2-05-voice-transcription-whisper.md) | Voice transcription (Whisper) |
| [ ]  | P2-06  | [P2-06-realtime-updates.md](P2-06-realtime-updates.md) | Real-time updates (optional) |
| [ ]  | P2-07  | [P2-07-inbox-ui-portal.md](P2-07-inbox-ui-portal.md) | Inbox UI (portal) |

## Phase 3: Value demonstration (weeks 9–12)

| Done | Task   | Plan file | Description |
|------|--------|-----------|-------------|
| [ ]  | P3-01  | [P3-01-reviews-schema-gbp-sync.md](P3-01-reviews-schema-gbp-sync.md) | Reviews schema and GBP reviews sync |
| [ ]  | P3-02  | [P3-02-reviews-api-reply.md](P3-02-reviews-api-reply.md) | Reviews API and reply |
| [ ]  | P3-03  | [P3-03-reviews-inbox-ui.md](P3-03-reviews-inbox-ui.md) | Reviews Inbox UI |
| [ ]  | P3-04  | [P3-04-diagnostics-domain-scores.md](P3-04-diagnostics-domain-scores.md) | Diagnostics domain and scores |
| [ ]  | P3-05  | [P3-05-shared-diagnostic-page.md](P3-05-shared-diagnostic-page.md) | Shared diagnostic page (public) |
| [ ]  | P3-06  | [P3-06-dashboard-kpis-needs-attention.md](P3-06-dashboard-kpis-needs-attention.md) | Dashboard KPIs and "Needs attention" |
| [ ]  | P3-07  | [P3-07-usage-billing-stripe.md](P3-07-usage-billing-stripe.md) | Usage and billing (Stripe) |
| [ ]  | P3-08  | [P3-08-tier-enforcement-limits.md](P3-08-tier-enforcement-limits.md) | Tier enforcement and limits |

## Phase 4: Enhanced inbox & AI (post-MVP)

| Done | Task   | Plan file | Description |
|------|--------|-----------|-------------|
| [ ]  | P4-01  | [P4-01-instagram-dms-integration.md](P4-01-instagram-dms-integration.md) | Instagram DMs integration |
| [ ]  | P4-02  | [P4-02-facebook-messenger-integration.md](P4-02-facebook-messenger-integration.md) | Facebook Messenger integration |
| [ ]  | P4-03  | [P4-03-email-inbox-integration.md](P4-03-email-inbox-integration.md) | Email inbox integration |
| [ ]  | P4-04  | [P4-04-web-chat-widget.md](P4-04-web-chat-widget.md) | Web chat widget |
| [ ]  | P4-05  | [P4-05-ai-reply-suggestions.md](P4-05-ai-reply-suggestions.md) | AI reply suggestions (reviews + inbox) |
| [ ]  | P4-06  | [P4-06-ai-credits-tracking.md](P4-06-ai-credits-tracking.md) | AI credits tracking and tier limits |
| [ ]  | P4-07  | [P4-07-conversation-assignment.md](P4-07-conversation-assignment.md) | Conversation assignment and team collaboration |
| [ ]  | P4-08  | [P4-08-response-templates.md](P4-08-response-templates.md) | Templates for reviews and inbox |
| [ ]  | P4-09  | [P4-09-tags-routing-automations.md](P4-09-tags-routing-automations.md) | Tags and routing automations |
| [ ]  | P4-10  | [P4-10-status-collision-detection.md](P4-10-status-collision-detection.md) | Status tracking and collision detection |

## Phase 5: Posts, listings, reports & integrations

| Done | Task   | Plan file | Description |
|------|--------|-----------|-------------|
| [ ]  | P5-01  | [P5-01-gbp-posts-scheduling.md](P5-01-gbp-posts-scheduling.md) | GBP Posts and scheduling |
| [ ]  | P5-02  | [P5-02-qa-management.md](P5-02-qa-management.md) | Q&A management |
| [ ]  | P5-03  | [P5-03-content-calendar.md](P5-03-content-calendar.md) | Content calendar UI |
| [ ]  | P5-04  | [P5-04-listings-sync-multi-directory.md](P5-04-listings-sync-multi-directory.md) | Listings sync (multi-directory) |
| [ ]  | P5-05  | [P5-05-revenue-impact-reports.md](P5-05-revenue-impact-reports.md) | Monthly Revenue Impact Reports |
| [ ]  | P5-06  | [P5-06-tracking-comparison-snapshots.md](P5-06-tracking-comparison-snapshots.md) | Tracking and comparison (snapshots) |
| [ ]  | P5-07  | [P5-07-activity-logs-retention.md](P5-07-activity-logs-retention.md) | Activity logs and retention |
| [ ]  | P5-08  | [P5-08-rest-api-webhooks.md](P5-08-rest-api-webhooks.md) | REST API (public) and webhooks |
| [ ]  | P5-09  | [P5-09-usage-wallet-overage.md](P5-09-usage-wallet-overage.md) | Usage wallet and overage billing |
| [ ]  | P5-10  | [P5-10-approval-workflows.md](P5-10-approval-workflows.md) | Approval workflows (basic to advanced) |

## How to use

1. **Before implementing a task:** Read the corresponding plan MD file.
2. **Check dependencies:** Ensure any listed dependencies are done.
3. **Implement:** Follow the scope and touch the listed files.
4. **Verify:** Run through the acceptance criteria and mark the todo complete.
5. **Mark done:** Update the checkbox in this README when the task is complete.
