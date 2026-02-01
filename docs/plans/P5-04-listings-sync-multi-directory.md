# P5-04: Listings sync (multi-directory)

**Task ID:** P5-04  
**Owner:** Backend + Frontend  
**Phase:** 5 – Posts, listings, reports & integrations

## Objective

Sync business NAP (Name, Address, Phone) and other listing data to multiple directories beyond Google (Apple Maps, Yelp, Bing Places, Facebook, industry-specific sites). Ensure consistent information across the web.

## Scope (agent-runnable)

- **Directories by tier:**
  - Basic: 5 directories (Google, Facebook, Apple Maps, core platforms)
  - Plus: 10 directories (+ Yelp, Bing, key verticals)
  - Pro: 20 directories (extended coverage)
  - Elite: Custom/extended
- **Data model:**
  - `listing_directories` table (id, name, slug, api_type, config).
  - `location_listings` table (id, location_id, directory_id, status `synced | pending | error`, external_id, last_synced_at, error_message).
- **Sync logic:**
  - Option A (API): Direct API integration for major platforms (Yelp, Facebook, Bing).
  - Option B (Aggregator): Use third-party aggregator (Yext, Synup, Uberall) via API.
  - Option C (Manual): Generate submission-ready data for directories without API.
- **Sync job:** Periodic job pushes NAP updates to directories; tracks status per directory.
- **UI:**
  - Listings page: Show all directories with sync status per location.
  - Sync button: Trigger immediate sync.
  - Edit: Update location NAP and propagate to all directories.
- **Tier gating:** Number of directories gated by tier.

## References

- [Agency-Pricing-Tiers.md](../../Agency-Pricing-Tiers.md) – Listings sync (5/10/20 directories)
- [SMB-Pricing-Tiers.md](../../SMB-Pricing-Tiers.md) – Listings sync by tier
- Yext / Synup / Uberall API docs (if using aggregator)
- Yelp, Bing, Facebook, Apple business APIs

## Dependencies

- P1-03 (locations with NAP data)
- P3-08 (tier enforcement for directory count)

## Files to create or modify

| File | Action |
|------|--------|
| `migrations/` | Create – `listing_directories`, `location_listings` tables |
| `internal/domains/listings/` | Create – Directory adapters, sync use case, status tracking |
| `internal/jobs/` | Add – Listings sync job |
| `api/openapi.yaml` | Add – Listings status and sync endpoints |
| Config | Add – API keys/config for each directory or aggregator |
| Portal: listings page | Create – Directory list with status, sync button |

## Acceptance criteria

- [ ] Location NAP data can be synced to multiple directories.
- [ ] Sync status is tracked per directory (synced/pending/error).
- [ ] Number of directories is gated by tier.
- [ ] Agent can view sync status for all directories.
- [ ] Agent can trigger manual sync.
- [ ] Errors are logged with messages; retry available.
- [ ] Adding/updating NAP triggers sync to all enabled directories.
