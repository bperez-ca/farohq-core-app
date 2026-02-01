# P5-10: Approval workflows (basic to advanced)

**Task ID:** P5-10  
**Owner:** Full-stack  
**Phase:** 5 – Posts, listings, reports & integrations

## Objective

Implement approval workflows for review replies, inbox messages, and posts. Staff members submit content for approval; managers/admins approve or reject before publishing.

## Scope (agent-runnable)

- **Approval types:**
  - Review replies: Before sending reply to GBP, require approval.
  - Inbox messages: Before sending to customer, require approval (optional per agency).
  - Posts: Before publishing to GBP, require approval.
- **Approval model:**
  - `approvals` table (id, agency_id, type `review_reply | message | post`, resource_id, content, status `pending | approved | rejected`, submitted_by, reviewed_by, reviewed_at, comments, created_at).
  - Or add `approval_status` column to existing tables (reviews, messages, posts).
- **Workflow:**
  1. Staff submits content → status = `pending_approval`.
  2. Manager/Admin views pending approvals → approves or rejects with comment.
  3. On approve → original action executes (send reply, send message, publish post).
  4. On reject → submitter notified with comment; can revise and resubmit.
- **Settings:**
  - Agency can enable/disable approvals per type.
  - RBAC: Only Manager/Admin can approve; Staff require approval.
- **API:**
  - `GET /api/v1/approvals?status=pending` – List pending approvals.
  - `POST /api/v1/approvals/:id/approve` – Approve and execute.
  - `POST /api/v1/approvals/:id/reject` – Reject with comment.
- **UI:**
  - Approvals queue: List pending approvals with preview.
  - Approve/reject actions with comment field.
  - Notification: Alert managers of pending approvals; alert staff of approval/rejection.
- **Tier gating:**
  - Growth: Basic approvals (review replies only).
  - Scale: Full approvals (reviews, posts, messages).
  - Enterprise: Multi-level approvals, approval policies.

## References

- [Agency-Pricing-Tiers.md](../../Agency-Pricing-Tiers.md) – Approvals by tier
- [SMB-Pricing-Tiers.md](../../SMB-Pricing-Tiers.md) – Approval flows
- [P1-01-auth-multi-tenancy-hardening.md](P1-01-auth-multi-tenancy-hardening.md) – RBAC roles

## Dependencies

- P1-01 (RBAC for role-based approval)
- P3-02 (reviews reply)
- P2-04 (messages)
- P5-01 (posts)
- P3-08 (tier enforcement)

## Files to create or modify

| File | Action |
|------|--------|
| `migrations/` | Create – `approvals` table or add `approval_status` to existing |
| `internal/domains/approvals/` | Create – Submit, approve, reject use cases |
| `api/openapi.yaml` | Add – Approvals endpoints |
| Review reply, message send, post publish | Extend – Check if approval required; create approval instead of executing |
| Portal: approvals page | Create – Pending approvals queue |
| Portal: approve/reject UI | Add – Approval actions with comment |
| Portal: notifications | Add – Notify on pending/approved/rejected |

## Acceptance criteria

- [ ] When approval is enabled and staff submits content, it goes to pending queue.
- [ ] Manager/Admin can view pending approvals.
- [ ] Approving executes the original action (send/publish).
- [ ] Rejecting notifies submitter with comments.
- [ ] Agency can enable/disable approvals per type.
- [ ] Only roles with approval permission can approve.
- [ ] Approval feature is gated by tier (Growth: basic, Scale: full).
