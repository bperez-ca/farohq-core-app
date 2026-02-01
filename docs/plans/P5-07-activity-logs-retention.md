# P5-07: Activity logs and retention

**Task ID:** P5-07  
**Owner:** Backend + Frontend  
**Phase:** 5 – Posts, listings, reports & integrations

## Objective

Implement activity logging for audit purposes. Track user actions (login, message sent, review replied, settings changed, etc.) with retention periods based on tier.

## Scope (agent-runnable)

- **Activity log table:** `activity_logs` (id, agency_id, user_id, action, resource_type, resource_id, metadata JSONB, ip_address, user_agent, created_at).
- **Actions to log:**
  - Auth: login, logout, password change
  - Conversations: message sent, conversation assigned, status changed
  - Reviews: reply sent
  - Settings: branding updated, domain added, team member invited
  - Clients/Locations: created, updated, deleted
  - Posts: created, scheduled, published
- **Logging mechanism:** Middleware or event-based; log after successful actions.
- **Retention by tier:**
  - Basic: 7 days
  - Plus: 30 days
  - Pro: 180 days
  - Elite: 1 year (365 days)
- **Retention job:** Periodic job deletes logs older than tier's retention period.
- **API:**
  - `GET /api/v1/activity-logs?resource_type=...&user_id=...` – List logs with filters.
- **UI:**
  - Activity log page in settings: Filterable list of recent activity.
- **Tier gating:** All tiers have activity logs; retention period varies by tier.

## References

- [Agency-Pricing-Tiers.md](../../Agency-Pricing-Tiers.md) – Activity logs retention (7/30/180/365 days)
- [SMB-Pricing-Tiers.md](../../SMB-Pricing-Tiers.md) – Activity logs by tier

## Dependencies

- P1-01 (auth context for user info)
- P3-08 (tier for retention period)

## Files to create or modify

| File | Action |
|------|--------|
| `migrations/` | Create – `activity_logs` table |
| `internal/domains/audit/` or `internal/platform/audit/` | Create – Logger service, log write use case |
| `internal/jobs/` | Add – Log retention cleanup job |
| `api/openapi.yaml` | Add – Activity logs list endpoint |
| Middleware / handlers | Extend – Log actions after success |
| Portal: settings/activity-logs | Create – Activity log viewer |

## Acceptance criteria

- [ ] User actions are logged to `activity_logs` table.
- [ ] Logs include user, action, resource, timestamp, and metadata.
- [ ] `GET /api/v1/activity-logs` returns logs filtered by agency context.
- [ ] Logs older than tier's retention period are deleted by cleanup job.
- [ ] Activity log page shows recent activity with filters.
- [ ] Logs are scoped to agency (RLS).
