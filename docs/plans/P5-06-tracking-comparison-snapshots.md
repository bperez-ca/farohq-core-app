# P5-06: Tracking and comparison (snapshots)

**Task ID:** P5-06  
**Owner:** Full-stack  
**Phase:** 5 – Posts, listings, reports & integrations

## Objective

Track metrics over time (review count, rating, presence score, Maps visibility) and provide before/after comparison views. Enables ROI proof by showing improvement since agency started managing the location.

## Scope (agent-runnable)

- **Metrics to track:**
  - Review count and average rating
  - Presence score (from diagnostics)
  - Reviews score, speed score
  - GBP profile completeness
  - Optional: Maps grid rank (Local Falcon/Dominator style – may require third-party API)
- **Snapshots table:** `metric_snapshots` (id, location_id, agency_id, snapshot_date, metrics JSONB, created_at).
- **Snapshot job:** Daily or weekly job captures current metrics for each location.
- **Comparison views:**
  - Before/After: Compare first snapshot to latest.
  - Period comparison: This month vs. last month; this quarter vs. last.
  - Trend charts: Line graphs showing metric changes over time.
- **API:**
  - `GET /api/v1/locations/:id/snapshots` – List snapshots for location.
  - `GET /api/v1/locations/:id/comparison?from=...&to=...` – Compare two periods.
- **UI:**
  - Tracking page: Trend charts for key metrics.
  - Comparison cards: "Reviews increased from X to Y (+Z%)" style.
- **Tier gating:** Tracking at Pro tier and above; comparison reports at Elite.

## References

- [Agency-Pricing-Tiers.md](../../Agency-Pricing-Tiers.md) – Tracking and comparison
- [SMB-Pricing-Tiers.md](../../SMB-Pricing-Tiers.md) – Tracking reports in Premium layer
- [SMB-client-creation-and-layers.md](SMB-client-creation-and-layers.md) – Tracking/comparison for ROI proof
- Local Falcon / Local Dominator style tracking

## Dependencies

- P3-04 (diagnostic scores)
- P3-01/P3-02 (reviews data)
- P1-03 (locations)
- P3-08 (tier enforcement)

## Files to create or modify

| File | Action |
|------|--------|
| `migrations/` | Create – `metric_snapshots` table |
| `internal/domains/tracking/` | Create – Snapshot capture, comparison logic |
| `internal/jobs/` | Add – Daily/weekly snapshot job |
| `api/openapi.yaml` | Add – Snapshots and comparison endpoints |
| Portal: tracking page | Create – Trend charts, comparison cards |

## Acceptance criteria

- [ ] Metrics are captured as snapshots on a regular schedule (daily/weekly).
- [ ] Agency can view trend charts for key metrics over time.
- [ ] Comparison view shows before/after improvement.
- [ ] Period comparison (month over month) is available.
- [ ] Tracking is gated to Pro tier and above.
- [ ] Snapshots are scoped to agency (RLS).
