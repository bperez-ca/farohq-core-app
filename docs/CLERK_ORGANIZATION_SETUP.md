# Clerk Organization Setup Guide

## Overview

This guide explains how to set up Clerk Organizations to map to your application's tenant/agency IDs, so that when users are onboarded to an agency, the `org_id` is automatically included in their JWT tokens.

## How Clerk Organizations Work

Clerk uses a nested `o` (organization) claim in session tokens when a user is part of an active organization. The structure is:

```json
{
  "o": {
    "id": "org_xxx",      // Organization ID
    "slg": "agency-slug", // Organization slug
    "rol": "admin",       // User's role (without "org:" prefix)
    "per": [...],         // Permissions
    "fpm": {...}          // Feature-permission map
  }
}
```

## Setting Up Organizations

### Option 1: Manual Setup (For Testing)

1. **Create Organization in Clerk Dashboard**:
   - Go to your Clerk Dashboard → Organizations
   - Create a new organization
   - Use your tenant/agency ID as the organization ID (or create a mapping)
   - Set the organization slug to match your tenant slug

2. **Add Users to Organization**:
   - When a user accepts an invite in your app, manually add them to the corresponding Clerk organization
   - Assign the appropriate role (owner, admin, staff, viewer)

### Option 2: Automated Setup (Recommended)

Use Clerk's Backend API to automatically create organizations and manage memberships:

#### When Creating a Tenant

When a tenant (agency) is created, create a corresponding Clerk organization:

```go
// In create_tenant.go use case, after tenant is created:
// 1. Create Clerk organization with tenant ID
// 2. Store the mapping (tenant_id -> clerk_org_id) if needed
```

**Example using Clerk Backend API**:
```bash
curl -X POST https://api.clerk.com/v1/organizations \
  -H "Authorization: Bearer YOUR_CLERK_SECRET_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Agency Name",
    "slug": "agency-slug",
    "max_allowed_memberships": 100
  }'
```

#### When User Accepts Invite

When a user accepts an invite (`accept_invite.go`), add them to the Clerk organization:

```go
// After creating TenantMember, add user to Clerk organization:
// 1. Get the Clerk organization ID for the tenant
// 2. Add the user to the organization via Clerk API
// 3. Assign the role from the invite
```

**Example using Clerk Backend API**:
```bash
curl -X POST https://api.clerk.com/v1/organizations/{org_id}/memberships \
  -H "Authorization: Bearer YOUR_CLERK_SECRET_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_xxx",
    "role": "admin"
  }'
```

## Mapping Tenant ID to Clerk Organization ID

You have two options:

### Option A: Use Tenant ID as Clerk Organization ID

- When creating a Clerk organization, use your tenant UUID as the organization ID
- This requires using Clerk's Backend API with a custom ID
- **Pros**: Direct mapping, no lookup needed
- **Cons**: Requires API access, more complex setup

### Option B: Store Mapping in Database

- Create a `clerk_organizations` table to map `tenant_id` → `clerk_org_id`
- When creating a tenant, create a Clerk organization and store the mapping
- **Pros**: Flexible, can use Clerk Dashboard
- **Cons**: Requires database lookup

**Example migration**:
```sql
CREATE TABLE clerk_organizations (
  tenant_id UUID PRIMARY KEY REFERENCES tenants(id),
  clerk_org_id TEXT NOT NULL UNIQUE,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

## Implementation Steps

### 1. Add Clerk Backend API Client

Create a service to interact with Clerk's Backend API:

```go
// internal/platform/clerk/client.go
type Client struct {
  apiKey string
  httpClient *http.Client
}

func (c *Client) CreateOrganization(ctx context.Context, tenantID uuid.UUID, name, slug string) (string, error) {
  // Create organization via Clerk API
  // Return clerk_org_id
}

func (c *Client) AddUserToOrganization(ctx context.Context, orgID, userID, role string) error {
  // Add user to organization via Clerk API
}
```

### 2. Update Create Tenant Use Case

Modify `create_tenant.go` to create a Clerk organization:

```go
// After tenant is created:
clerkOrgID, err := clerkClient.CreateOrganization(ctx, tenant.ID(), tenant.Name(), tenant.Slug())
if err != nil {
  // Log error but don't fail tenant creation
  logger.Error().Err(err).Msg("Failed to create Clerk organization")
}
// Store mapping if using Option B
```

### 3. Update Accept Invite Use Case

Modify `accept_invite.go` to add user to Clerk organization:

```go
// After TenantMember is created:
// 1. Get Clerk organization ID for the tenant
clerkOrgID := getClerkOrgIDForTenant(ctx, invite.TenantID())

// 2. Get Clerk user ID (from token or user lookup)
clerkUserID := getClerkUserID(ctx, req.UserID)

// 3. Add user to organization
err = clerkClient.AddUserToOrganization(ctx, clerkOrgID, clerkUserID, invite.Role().String())
if err != nil {
  // Log error but don't fail invite acceptance
  logger.Error().Err(err).Msg("Failed to add user to Clerk organization")
}
```

### 4. Update Auth Middleware

The auth middleware has been updated to extract organization claims from both:
- Clerk's standard `o` claim (nested structure)
- Flat `org_id`, `org_slug`, `org_role` claims (for backward compatibility)

## Testing

1. **Create a tenant** and verify Clerk organization is created
2. **Accept an invite** and verify user is added to organization
3. **Check JWT token** - should contain `o` claim with organization info
4. **Verify backend** - `org_id` should be available in request context

## Troubleshooting

### `org_id` is null in tokens

- User may not be part of any organization
- Check if user was added to Clerk organization
- Verify organization was created for the tenant

### Organization claims not matching tenant

- Verify the mapping between tenant ID and Clerk organization ID
- Check if the correct organization is being used
- Ensure user is added to the correct organization

### Role mismatch

- Clerk roles should match your application roles (owner, admin, staff, viewer)
- Verify role is set correctly when adding user to organization

## Environment Variables

Add to your `.env`:

```bash
CLERK_SECRET_KEY=sk_test_xxx  # For Backend API calls
CLERK_JWKS_URL=https://your-instance.clerk.accounts.dev/.well-known/jwks.json
```

## Next Steps

1. Implement Clerk Backend API client
2. Update tenant creation to create organizations
3. Update invite acceptance to add users to organizations
4. Test the flow end-to-end
5. Monitor logs to ensure organizations are created and users are added correctly
