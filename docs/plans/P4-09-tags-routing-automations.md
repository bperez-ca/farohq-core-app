# P4-09: Tags and routing automations

**Task ID:** P4-09  
**Owner:** Full-stack  
**Phase:** 4 – Enhanced inbox & AI

## Objective

Implement tags for conversations and basic routing automations. Agents can manually tag conversations; automations can auto-tag and route based on rules (e.g., keywords, channel, sentiment).

## Scope (agent-runnable)

- **Tags:**
  - Add `tags` table (id, agency_id, name, color) and `conversation_tags` junction table.
  - API: CRUD for tags; add/remove tags from conversation.
  - UI: Tag picker in conversation detail; tag filter in conversation list.
- **Routing automations (rules engine):**
  - Add `automation_rules` table (id, agency_id, name, trigger, conditions JSONB, actions JSONB, enabled, priority).
  - Triggers: `conversation.created`, `message.received`, `review.created`.
  - Conditions: channel equals, keyword contains, sentiment, time of day.
  - Actions: add_tag, assign_to, set_status, send_auto_reply.
  - Evaluation: On trigger, evaluate rules in priority order; execute matching actions.
- **UI:**
  - Automations settings page: List rules, create/edit rule (trigger, conditions, actions).
  - Basic rule builder (dropdowns for condition type, value, action type).
- **Tier gating:** Manual tags at all tiers; automations at Plus tier and above.

## References

- [Agency-Pricing-Tiers.md](../../Agency-Pricing-Tiers.md) – Automations: tags & routing at Plus+
- [SMB-Pricing-Tiers.md](../../SMB-Pricing-Tiers.md) – Tags and routing, workflow automations

## Dependencies

- P2-01 (conversations table)
- P2-04 (conversations API)
- P4-07 (assignment for routing)
- P3-08 (tier enforcement)

## Files to create or modify

| File | Action |
|------|--------|
| `migrations/` | Create – `tags`, `conversation_tags`, `automation_rules` tables |
| `internal/domains/tags/` | Create – Tags CRUD |
| `internal/domains/automations/` | Create – Rules engine, rule evaluation, action execution |
| `api/openapi.yaml` | Add – Tags and automations endpoints |
| Event hooks | Extend – On conversation/message/review created, evaluate automation rules |
| Portal: settings/tags | Create – Tag management page |
| Portal: settings/automations | Create – Automations management (rule builder) |
| Portal: conversation detail | Add – Tag picker |
| Portal: conversation list | Add – Tag filter |

## Acceptance criteria

- [ ] Agency can create, list, update, delete tags.
- [ ] Agent can add/remove tags from a conversation; tags display in conversation list and detail.
- [ ] Conversation list can be filtered by tag.
- [ ] Agency can create automation rules with trigger, conditions, actions.
- [ ] On trigger event, matching rules execute actions (e.g., auto-tag, auto-assign).
- [ ] Automations are gated to Plus tier and above.
- [ ] Rules can be enabled/disabled; priority determines order of evaluation.
