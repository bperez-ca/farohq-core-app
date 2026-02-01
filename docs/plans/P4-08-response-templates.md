# P4-08: Templates for reviews and inbox

**Task ID:** P4-08  
**Owner:** Full-stack  
**Phase:** 4 – Enhanced inbox & AI

## Objective

Provide pre-built and custom response templates for reviews and inbox messages. Agents can quickly insert common replies, improving response time and consistency.

## Scope (agent-runnable)

- **Templates table:** `response_templates` (id, agency_id, type `review | inbox`, name, content, category, shortcut, created_at, updated_at).
- **CRUD API:**
  - `GET /api/v1/templates?type=review` – List templates.
  - `POST /api/v1/templates` – Create template.
  - `PUT /api/v1/templates/:id` – Update template.
  - `DELETE /api/v1/templates/:id` – Delete template.
- **Pre-built templates:** Seed default templates (e.g., "Thank you for 5-star review", "Sorry for poor experience", "How can I help?").
- **Template insertion:** In inbox reply input and review reply input, add "Templates" button; show list; click to insert into input.
- **Variables (optional):** Support placeholders like `{{customer_name}}`, `{{business_name}}` that are replaced on insert.
- **Categories:** Group templates (e.g., "5-star reviews", "Complaints", "FAQs").
- **Tier gating:** Templates available at all tiers (basic functionality); advanced features (variables, unlimited templates) at Plus+.

## References

- [Agency-Pricing-Tiers.md](../../Agency-Pricing-Tiers.md) – Templates mentioned in reviews management
- [SMB-Pricing-Tiers.md](../../SMB-Pricing-Tiers.md) – Review response templates, preset templates

## Dependencies

- P2-04 (inbox API for context)
- P3-02 (reviews API for context)
- P1-01 (tenant context)

## Files to create or modify

| File | Action |
|------|--------|
| `migrations/` | Create – `response_templates` table |
| `internal/domains/templates/` | Create – CRUD use cases and handlers |
| `api/openapi.yaml` | Add – Templates CRUD endpoints |
| `scripts/` or `seeds/` | Add – Default template seeds |
| Portal: settings/templates | Create – Templates management page (list, create, edit, delete) |
| Portal: inbox reply input | Add – Templates button and picker |
| Portal: reviews reply input | Add – Templates button and picker |

## Acceptance criteria

- [ ] Agency can create, list, update, delete response templates.
- [ ] Templates are scoped to agency (RLS).
- [ ] Default templates are seeded on agency creation (or first access).
- [ ] Agent can insert a template into inbox reply input; content appears in input field.
- [ ] Agent can insert a template into review reply input.
- [ ] Optional: Variables like `{{customer_name}}` are replaced on insert.
- [ ] Templates can be categorized (e.g., by star rating or topic).
