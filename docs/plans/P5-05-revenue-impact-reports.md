# P5-05: Monthly Revenue Impact Reports

**Task ID:** P5-05  
**Owner:** Full-stack  
**Phase:** 5 – Posts, listings, reports & integrations

## Objective

Generate comprehensive monthly revenue impact reports that show ROI: visibility score, lead breakdown (hot/warm/cold), estimated revenue by channel, and actionable priorities. This is FARO's signature differentiator vs. competitors.

## Scope (agent-runnable)

- **Report types by tier:**
  - Basic: Visibility Check (1/month) – visibility score, basic stats.
  - Plus: Monthly Revenue Impact (4/month) – full report.
  - Pro: Monthly Revenue Impact (12/month).
  - Elite: Competitive Intelligence (30/month) – includes competitor analysis.
- **Report data model:** `impact_reports` table (id, client_id, location_id, agency_id, type, period_start, period_end, visibility_score, presence_score, reviews_score, speed_score, leads JSONB, revenue_estimate, channel_attribution JSONB, priorities JSONB, created_at).
- **Revenue calculation:**
  - Track leads from conversations (contact created, source channel).
  - Use average booking value (configured per client/location).
  - Estimate conversion rate (configurable or learned).
  - Revenue = leads × conversion_rate × avg_booking_value.
- **Channel attribution:** Attribute revenue to channels (WhatsApp, GBP, Reviews, etc.) based on lead source.
- **Lead scoring:** Categorize leads as hot/warm/cold based on engagement, recency, intent signals.
- **Report generation:**
  - Scheduled job at month-end (or on-demand).
  - Compute metrics for period; store in `impact_reports`.
- **API:**
  - `POST /api/v1/reports/generate` – Generate report for client/location.
  - `GET /api/v1/reports?client_id=...` – List reports.
  - `GET /api/v1/reports/:id` – Get report details.
- **UI:**
  - Reports page: List generated reports.
  - Report detail: Visualize scores, leads breakdown, revenue by channel, priorities.
  - Export: PDF export for client presentation.
- **Tier gating:** Report frequency and type gated by tier.

## References

- [Agency-Pricing-Tiers.md](../../Agency-Pricing-Tiers.md) – Impact Reports section
- [SMB-Pricing-Tiers.md](../../SMB-Pricing-Tiers.md) – Impact Reports by tier
- [P3-04-diagnostics-domain-scores.md](P3-04-diagnostics-domain-scores.md) – Score calculations

## Dependencies

- P2-04 (conversations for lead data)
- P3-01/P3-02 (reviews data)
- P3-04 (diagnostic scores)
- P3-08 (tier enforcement)

## Files to create or modify

| File | Action |
|------|--------|
| `migrations/` | Create – `impact_reports` table |
| `internal/domains/reports/` | Create – Report generation, revenue calculation, lead scoring |
| `internal/jobs/` | Add – Monthly report generation job |
| `api/openapi.yaml` | Add – Reports endpoints |
| Portal: reports page | Create – List reports, view report detail |
| Portal: report visualization | Create – Score gauges, lead breakdown chart, revenue by channel |
| Portal: PDF export | Add – Generate PDF from report data |

## Acceptance criteria

- [ ] Monthly report can be generated for a client/location.
- [ ] Report includes visibility score, lead breakdown (hot/warm/cold count), revenue estimate.
- [ ] Revenue is attributed to channels (e.g., WhatsApp: $X, GBP: $Y).
- [ ] Report lists top 3 priorities/recommendations.
- [ ] Reports are stored and can be viewed later.
- [ ] Report frequency is gated by tier (1/4/12/30 per month).
- [ ] PDF export is available.
