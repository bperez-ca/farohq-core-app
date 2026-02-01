# Environment Variables for Agents and Local Development

This document lists environment variables required for branding, domain verification, and related features. Use this when configuring local development, CI, or automated agents.

## Core App (farohq-core-app)

### Vercel (Custom Domain Verification)

Required for domain verification and CNAME instructions (Scale tier only). Without these, domain-instructions and domain-status endpoints will return errors when calling the Vercel API.

| Variable | Description | Required |
|----------|-------------|----------|
| `VERCEL_API_TOKEN` | Vercel API token. Create at [Vercel Dashboard](https://vercel.com/account/tokens). | For domain verification |
| `VERCEL_PROJECT_ID` | Vercel project ID (portal deployment). Found in project settings. | For domain verification |
| `VERCEL_TEAM_ID` | Vercel team ID. Optional; required only when using a team. | Optional |

### Web and API

| Variable | Description | Default |
|----------|-------------|---------|
| `WEB_URL` | Base web URL for the portal (e.g. `http://localhost:3000`). | `http://localhost:3000` |
| `PORT` | HTTP server port. | `8080` |

### Other (reference)

- `NEXT_PUBLIC_API_URL` / `API_URL` – Used by the portal to reach the core API. Set in the portal's environment.

## Portal (farohq-portal)

| Variable | Description | Default |
|----------|-------------|---------|
| `NEXT_PUBLIC_API_URL` or `API_URL` | Core API base URL (e.g. `http://localhost:8080`). | `http://localhost:8080` |
| `NEXT_PUBLIC_APP_DOMAIN` | App domain for branding context. | `app.thefaro.co` |
| `NEXT_PUBLIC_PORTAL_WILDCARD` | Portal wildcard domain (e.g. `portal.thefaro.co`). | `portal.thefaro.co` |

## Brand Resolution Notes

- **Subdomain resolution** (`get_by_host`): Resolves brands by host. Subdomain resolution assumes the pattern `*.portal.farohq.com` (or equivalent). This is currently hardcoded in `internal/domains/brand/app/usecases/get_by_host.go`.
- **Custom domain resolution**: Uses the `domain` field from the brand. When a user accesses the portal via a custom domain (e.g. `portal.agency.com`), the brand is resolved by matching the host to the stored `domain` value.

## Related Documentation

- [BRANDING_IMPLEMENTATION_SUMMARY.md](../BRANDING_IMPLEMENTATION_SUMMARY.md) – Component styles and domain verification overview
- [P1-02-white-label-domain-verification.md](P1-02-white-label-domain-verification.md) – White-label and domain verification plan
