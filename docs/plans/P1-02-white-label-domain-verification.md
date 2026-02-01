# P1-02: White-label and domain verification

**Task ID:** P1-02  
**Owner:** Full-stack  
**Phase:** 1 – MVP foundation

## Objective

Confirm the theme-by-domain flow works end-to-end and that custom domain CNAME instructions and verification (e.g. Vercel) work. Document any env vars needed for agents.

## Scope (agent-runnable)

- Confirm theme-by-domain flow end-to-end: portal uses brand by host (e.g. `GET /v1/brand/by-host` or by-domain).
- Ensure custom domain CNAME instructions and verification (Vercel) work per [BRANDING_IMPLEMENTATION_SUMMARY.md](../BRANDING_IMPLEMENTATION_SUMMARY.md).
- Document env vars for agents (e.g. Vercel API token, project ID, web URL).

## References

- [docs/BRANDING_IMPLEMENTATION_SUMMARY.md](../BRANDING_IMPLEMENTATION_SUMMARY.md)
- [internal/domains/brand/](../../internal/domains/brand/) – brand use cases and HTTP
- Portal: brand theme provider, domain verification UI

## Dependencies

- P1-01 (auth/tenant context available for brand endpoints).

## Files to create or modify

| File | Action |
|------|--------|
| `docs/BRANDING_IMPLEMENTATION_SUMMARY.md` | Update if gaps found |
| `docs/plans/ENV_VARS_AGENTS.md` or section in README | Create/update – list env vars for branding (Vercel, Web URL, etc.) |
| Portal brand theme / by-host usage | Verify – no code change if already correct |

## Acceptance criteria

- [ ] Portal loads theme when accessed by custom domain (brand resolved by host).
- [ ] Custom domain CNAME instructions are shown in UI and match Vercel expectations.
- [ ] Domain verification (Vercel) succeeds when CNAME is correctly set.
- [ ] A short doc lists env vars required for branding/domain verification (for agents and local dev).
