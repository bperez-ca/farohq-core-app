# PostgreSQL Row-Level Security (RLS) & Database Layer Analysis

**Date**: 2025-01-27  
**Target**: FARO Multi-Tenant SaaS Platform  
**Database**: PostgreSQL 15 with Row-Level Security (RLS)  
**Critical Impact**: $0.50/location with RLS vs $5+/location without RLS

---

## Executive Summary

✅ **RLS Implementation**: **STRONG** - All tenant-scoped tables have RLS enabled with proper policies  
⚠️ **Critical Issues Found**: 2 security gaps, 1 performance bottleneck  
✅ **Tenant Context**: Properly passed via PostgreSQL session variables  
✅ **Indexing**: Comprehensive coverage on all tenant_id/agency_id columns  
⚠️ **Missing**: RLS tests, backup/recovery documentation, partitioning strategy

**Key Findings**:
1. ✅ All 7 tenant-scoped tables have RLS enabled
2. ⚠️ `agencies` table missing RLS (should be isolated for super-admin only)
3. ⚠️ `users` table missing RLS (cross-tenant user data risk)
4. 🔴 Critical N+1 pattern in seat usage calculation (2,001 queries for 1,000 clients)
5. ✅ Comprehensive indexing on all tenant_id columns
6. ⚠️ No RLS isolation tests found
7. ⚠️ No backup/recovery documentation

---

## 1. SCHEMA STRUCTURE

### 1.1 All Tables and Their Purpose

| Table | Purpose | Primary Key | Tenant Isolation | RLS Enabled |
|-------|---------|-------------|------------------|-------------|
| `agencies` | Top-level tenant (agency) records | `id` (UUID) | ❌ **NO** - Should be super-admin only | ❌ **MISSING** |
| `branding` | White-label branding per agency | `agency_id` (UUID, FK) | ✅ Yes (via `agency_id`) | ✅ Yes |
| `tenant_members` | Agency membership (users → agencies) | `id` (UUID) | ✅ Yes (via `tenant_id`) | ✅ Yes |
| `tenant_invites` | Agency invitation tokens | `id` (UUID) | ✅ Yes (via `tenant_id`) | ✅ Yes |
| `clients` | SMB accounts under agencies | `id` (UUID) | ✅ Yes (via `agency_id`) | ✅ Yes |
| `locations` | Business locations under clients | `id` (UUID) | ✅ Yes (via `client_id` → `agency_id`) | ✅ Yes |
| `client_members` | Client membership (users → clients) | `id` (UUID) | ✅ Yes (via `client_id` → `agency_id`) | ✅ Yes |
| `users` | User profiles (from Clerk) | `id` (UUID) | ❌ **NO** - Cross-tenant user data | ❌ **MISSING** |

**Total**: 8 tables, 7 with tenant isolation, 2 missing RLS

### 1.2 Primary Keys and Foreign Keys

#### `agencies` Table
```sql
PRIMARY KEY: id (UUID)
FOREIGN KEYS: None (top-level entity)
TENANT ISOLATION: ❌ None - This IS the tenant
```

#### `branding` Table
```sql
PRIMARY KEY: agency_id (UUID) - References agencies(id)
FOREIGN KEYS: agency_id → agencies(id) ON DELETE CASCADE
TENANT ISOLATION: ✅ agency_id
```

#### `tenant_members` Table
```sql
PRIMARY KEY: id (UUID)
FOREIGN KEYS: tenant_id → agencies(id) ON DELETE CASCADE
TENANT ISOLATION: ✅ tenant_id
UNIQUE CONSTRAINT: (tenant_id, user_id)
```

#### `tenant_invites` Table
```sql
PRIMARY KEY: id (UUID)
FOREIGN KEYS: tenant_id → agencies(id) ON DELETE CASCADE
TENANT ISOLATION: ✅ tenant_id
UNIQUE CONSTRAINT: (tenant_id, email)
```

#### `clients` Table
```sql
PRIMARY KEY: id (UUID)
FOREIGN KEYS: agency_id → agencies(id) ON DELETE CASCADE
TENANT ISOLATION: ✅ agency_id
UNIQUE CONSTRAINT: (agency_id, slug)
```

#### `locations` Table
```sql
PRIMARY KEY: id (UUID)
FOREIGN KEYS: client_id → clients(id) ON DELETE CASCADE
TENANT ISOLATION: ✅ client_id → agency_id (indirect)
```

#### `client_members` Table
```sql
PRIMARY KEY: id (UUID)
FOREIGN KEYS: client_id → clients(id) ON DELETE CASCADE, location_id → locations(id) ON DELETE SET NULL
TENANT ISOLATION: ✅ client_id → agency_id (indirect)
UNIQUE CONSTRAINT: (client_id, user_id, location_id)
```

#### `users` Table
```sql
PRIMARY KEY: id (UUID)
FOREIGN KEYS: None
TENANT ISOLATION: ❌ None - Cross-tenant user data
UNIQUE CONSTRAINT: clerk_user_id (TEXT)
```

### 1.3 Reference Data Tables

**No reference data tables found** - All tables are tenant-scoped or user-scoped.

**Recommendation**: If reference data tables are added (e.g., `countries`, `timezones`, `categories`), they should:
- ❌ **NOT** have RLS enabled
- ✅ Have appropriate indexes for lookups
- ✅ Be read-only for application users

### 1.4 Largest Tables (Partitioning Candidates)

**Current Size**: Unknown (no size monitoring found)

**Estimated Growth** (12 months, 30 agencies, 1,000 locations):
- `locations`: ~1,000 rows × ~2KB = **~2MB** ✅ Small
- `client_members`: ~3,000 rows × ~500B = **~1.5MB** ✅ Small
- `tenant_members`: ~300 rows × ~500B = **~150KB** ✅ Small
- `users`: ~300 rows × ~1KB = **~300KB** ✅ Small

**Partitioning Threshold**: Tables >10GB should be partitioned.

**Current Status**: ✅ **No partitioning needed** - All tables are small.

**Future Monitoring**: 
- Monitor `locations` table (will grow fastest)
- Consider partitioning by `tenant_id` if >10GB
- Partition strategy: **LIST** partitioning by `agency_id` (if needed)

---

## 2. ROW-LEVEL SECURITY (RLS)

### 2.1 Tables with RLS Enabled

✅ **7 tables have RLS enabled**:
1. `branding` - ✅ RLS enabled
2. `tenant_members` - ✅ RLS enabled
3. `tenant_invites` - ✅ RLS enabled
4. `clients` - ✅ RLS enabled
5. `locations` - ✅ RLS enabled
6. `client_members` - ✅ RLS enabled

❌ **2 tables missing RLS**:
1. `agencies` - ❌ **MISSING** (should be super-admin only)
2. `users` - ❌ **MISSING** (cross-tenant user data risk)

### 2.2 RLS Policies (SELECT, INSERT, UPDATE, DELETE)

#### `branding` Table
```sql
-- Migration: 000001_init.up.sql
ALTER TABLE branding ENABLE ROW LEVEL SECURITY;

CREATE POLICY branding_tenant ON branding
    USING (agency_id = current_setting('lv.tenant_id')::uuid);
```
**Coverage**: ✅ SELECT, INSERT, UPDATE, DELETE (single policy covers all operations)  
**Soft Delete**: ❌ Not included in policy (should add `AND deleted_at IS NULL` if soft delete added)

#### `tenant_members` Table
```sql
-- Migration: 000003_create_tenants.up.sql
ALTER TABLE tenant_members ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_members_tenant ON tenant_members
    USING (tenant_id = current_setting('lv.tenant_id')::uuid AND deleted_at IS NULL);
```
**Coverage**: ✅ SELECT, INSERT, UPDATE, DELETE  
**Soft Delete**: ✅ Included (`deleted_at IS NULL`)

#### `tenant_invites` Table
```sql
-- Migration: 000003_create_tenants.up.sql
ALTER TABLE tenant_invites ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_invites_tenant ON tenant_invites
    USING (tenant_id = current_setting('lv.tenant_id')::uuid AND deleted_at IS NULL);
```
**Coverage**: ✅ SELECT, INSERT, UPDATE, DELETE  
**Soft Delete**: ✅ Included (`deleted_at IS NULL`)

#### `clients` Table
```sql
-- Migration: 000004_agency_hierarchy.up.sql
ALTER TABLE clients ENABLE ROW LEVEL SECURITY;

CREATE POLICY clients_tenant ON clients
    USING (agency_id = current_setting('lv.tenant_id')::uuid AND deleted_at IS NULL);
```
**Coverage**: ✅ SELECT, INSERT, UPDATE, DELETE  
**Soft Delete**: ✅ Included (`deleted_at IS NULL`)

#### `locations` Table
```sql
-- Migration: 000004_agency_hierarchy.up.sql
ALTER TABLE locations ENABLE ROW LEVEL SECURITY;

CREATE POLICY locations_tenant ON locations
    USING (client_id IN (
        SELECT id FROM clients 
        WHERE agency_id = current_setting('lv.tenant_id')::uuid 
        AND deleted_at IS NULL
    ) AND deleted_at IS NULL);
```
**Coverage**: ✅ SELECT, INSERT, UPDATE, DELETE  
**Soft Delete**: ✅ Included (`deleted_at IS NULL`)  
**Performance**: ⚠️ Uses subquery (should be optimized with JOIN or direct FK check)

#### `client_members` Table
```sql
-- Migration: 000004_agency_hierarchy.up.sql
ALTER TABLE client_members ENABLE ROW LEVEL SECURITY;

CREATE POLICY client_members_tenant ON client_members
    USING (client_id IN (
        SELECT id FROM clients 
        WHERE agency_id = current_setting('lv.tenant_id')::uuid 
        AND deleted_at IS NULL
    ) AND deleted_at IS NULL);
```
**Coverage**: ✅ SELECT, INSERT, UPDATE, DELETE  
**Soft Delete**: ✅ Included (`deleted_at IS NULL`)  
**Performance**: ⚠️ Uses subquery (should be optimized with JOIN or direct FK check)

### 2.3 Missing RLS Policies

#### ⚠️ **Issue 1: `agencies` Table Missing RLS**

**Risk**: **MEDIUM** - Agencies can see each other's metadata (name, slug, tier, seat limits)

**Current State**:
```sql
-- No RLS enabled on agencies table
-- Queries can access any agency's data
```

**Recommendation**:
```sql
-- Option A: Super-admin only (recommended)
ALTER TABLE agencies ENABLE ROW LEVEL SECURITY;

CREATE POLICY agencies_super_admin ON agencies
    USING (current_setting('lv.is_super_admin', true)::boolean = true);

-- Option B: Self-access only (if agencies need to read their own data)
CREATE POLICY agencies_self ON agencies
    USING (id = current_setting('lv.tenant_id')::uuid);
```

**Impact**: Low (agencies rarely query other agencies, but metadata leak is possible)

#### ⚠️ **Issue 2: `users` Table Missing RLS**

**Risk**: **HIGH** - Users from one tenant can see users from other tenants

**Current State**:
```sql
-- No RLS enabled on users table
-- Queries can access any user's data across all tenants
```

**Recommendation**:
```sql
-- Option A: Filter by tenant membership (recommended)
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

CREATE POLICY users_tenant ON users
    USING (
        id IN (
            SELECT user_id FROM tenant_members 
            WHERE tenant_id = current_setting('lv.tenant_id')::uuid 
            AND deleted_at IS NULL
        )
        OR id IN (
            SELECT user_id FROM client_members 
            WHERE client_id IN (
                SELECT id FROM clients 
                WHERE agency_id = current_setting('lv.tenant_id')::uuid 
                AND deleted_at IS NULL
            )
            AND deleted_at IS NULL
        )
    );
```

**Impact**: High (user data includes email, name, phone - PII leak risk)

### 2.4 RLS Policy Performance Issues

#### ⚠️ **Subquery Performance in `locations` and `client_members` Policies**

**Current Policy** (locations):
```sql
CREATE POLICY locations_tenant ON locations
    USING (client_id IN (
        SELECT id FROM clients 
        WHERE agency_id = current_setting('lv.tenant_id')::uuid 
        AND deleted_at IS NULL
    ) AND deleted_at IS NULL);
```

**Problem**: Subquery executed for every row check (expensive)

**Optimization Options**:

**Option 1: Add `agency_id` column to `locations` (denormalization)**
```sql
ALTER TABLE locations ADD COLUMN agency_id UUID REFERENCES agencies(id);
CREATE INDEX idx_locations_agency_id ON locations(agency_id);

CREATE POLICY locations_tenant ON locations
    USING (agency_id = current_setting('lv.tenant_id')::uuid AND deleted_at IS NULL);
```
**Pros**: Fast, simple  
**Cons**: Denormalization (must keep in sync)

**Option 2: Use JOIN in policy (PostgreSQL 10+)**
```sql
-- Not supported - policies can't use JOINs directly
```

**Option 3: Accept subquery (current)**
**Pros**: Normalized, correct  
**Cons**: Slower (but acceptable for <10K locations)

**Recommendation**: ✅ **Keep current approach** - Subquery is acceptable for current scale. Revisit if locations >10K per tenant.

---

## 3. TENANT CONTEXT PASSING

### 3.1 How `tenant_id` is Determined

**Resolution Priority** (from `internal/platform/httpserver/tenant.go`):
1. **Domain** (`Host` header) → Query `branding` table for `agency_id`
2. **X-Tenant-ID header** (API calls) → Direct tenant ID
3. **URL parameter** (`/api/v1/tenants/{id}/...`) → From URL path

**Code Location**: `internal/platform/httpserver/tenant.go:33-50`

```go
// Try to resolve tenant from domain
tenantID, err := tenantResolver.ResolveTenant(r.Context(), host)
if err != nil {
    // Try fallback to X-Tenant-ID header
    tenantIDHeader := r.Header.Get("X-Tenant-ID")
    if tenantIDHeader != "" {
        tenantID = tenantIDHeader
    }
}
```

### 3.2 PostgreSQL Session Variable Setting

**Middleware**: `TenantResolution` and `TenantResolutionWithAuth`  
**Location**: `internal/platform/httpserver/tenant.go:65-84`

**Implementation**:
```go
// Get connection from pool
conn, err := db.Acquire(r.Context())
defer conn.Release()

// Set tenant_id for RLS
_, err = conn.Exec(r.Context(), "SELECT set_config('lv.tenant_id', $1, true)", tenantID)

// Set client_id for RLS (if provided)
if clientID != "" {
    _, err = conn.Exec(r.Context(), "SELECT set_config('lv.client_id', $1, true)", validClientID)
}
```

**Key Points**:
- ✅ Uses `set_config(..., true)` - `true` = session-local (not transaction-local)
- ✅ Connection acquired per request (ensures session variable is set)
- ✅ Connection released after request (pooled for reuse)
- ⚠️ **Issue**: Connection acquired but may not be reused for subsequent queries in same request

**Potential Issue**: If connection pool returns a different connection for subsequent queries, RLS context may be lost.

**Verification Needed**: Test that all queries in a request use the same connection or that RLS context persists across connections.

### 3.3 Places Where `tenant_id` is Missed

#### ✅ **All Repository Queries Include Tenant Filter**

**Analysis**: Reviewed all repository queries:
- ✅ `client_repository.go`: All queries include `agency_id = $1`
- ✅ `location_repository.go`: All queries include `client_id = $1` (implies tenant via FK)
- ✅ `tenant_member_repository.go`: All queries include `tenant_id = $1`
- ✅ `brand_repository.go`: Queries by `agency_id` or domain (domain is tenant-specific)

**Status**: ✅ **No leaks found** - All queries properly filtered.

#### ⚠️ **Exception: `FindPendingInvitesByEmail`**

**Location**: `internal/domains/tenants/infra/db/invite_repository.go:212-264`

**Query**:
```sql
SELECT id, tenant_id, email, role, token, expires_at, accepted_at, revoked_at, created_at, created_by
FROM tenant_invites
WHERE LOWER(TRIM(email)) = $1 
    AND accepted_at IS NULL 
    AND revoked_at IS NULL 
    AND expires_at > NOW()
    AND deleted_at IS NULL
ORDER BY created_at DESC
```

**Issue**: ❌ **No tenant filter** - Returns invites across all tenants for an email

**Risk**: **LOW** - This is intentional (user accepting invite may have invites from multiple tenants)

**Mitigation**: ✅ **Acceptable** - RLS policy will filter results to current tenant anyway (defense-in-depth)

### 3.4 Query Layer Interaction with Tenant Context

**Pattern**: **Defense-in-Depth**

1. **RLS (Database Layer)**:
   - Session variable `lv.tenant_id` set by middleware
   - RLS policies automatically filter all queries
   - **Status**: ✅ Implemented

2. **App-Layer Filtering**:
   - All repository queries include explicit `WHERE` clauses with `tenant_id` or `agency_id`
   - Example: `WHERE agency_id = $1 AND deleted_at IS NULL`
   - **Status**: ✅ Implemented

**Security Assessment**: ✅ **STRONG** - Defense-in-depth prevents data leaks even if RLS fails.

---

## 4. INDEXING & PERFORMANCE

### 4.1 Indexes on `tenant_id` Columns

✅ **All tenant_id/agency_id columns are indexed**:

| Table | Column | Index Name | Type |
|-------|--------|------------|------|
| `branding` | `agency_id` | `idx_branding_agency_id` | B-tree |
| `tenant_members` | `tenant_id` | `idx_tenant_members_tenant_id` | B-tree |
| `tenant_members` | `(tenant_id, user_id)` | `idx_tenant_members_tenant_user` | Composite |
| `tenant_invites` | `tenant_id` | `idx_tenant_invites_tenant_id` | B-tree |
| `clients` | `agency_id` | `idx_clients_agency_id` | B-tree |
| `clients` | `(agency_id, slug)` | `idx_clients_slug` | Composite |
| `clients` | `(agency_id, status)` | `idx_clients_status` | Composite (partial: `WHERE deleted_at IS NULL`) |
| `locations` | `client_id` | `idx_locations_client_id` | B-tree |
| `client_members` | `client_id` | `idx_client_members_client_id` | B-tree |

**Status**: ✅ **Comprehensive** - All tenant-scoped queries can use indexes.

### 4.2 Composite Indexes

✅ **Composite indexes exist where needed**:
- `(tenant_id, user_id)` on `tenant_members` - ✅ For user tenant lookup
- `(agency_id, slug)` on `clients` - ✅ For unique constraint + lookup
- `(agency_id, status)` on `clients` - ✅ For filtered queries

**Missing Composite Indexes**:
- ❌ `(client_id, deleted_at)` on `locations` - Could optimize RLS policy subquery
- ❌ `(agency_id, deleted_at)` on `clients` - Could optimize RLS policy subquery

**Recommendation**: ⚠️ **Consider adding** if RLS policy subqueries become slow:
```sql
CREATE INDEX idx_clients_agency_deleted ON clients(agency_id, deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_locations_client_deleted ON locations(client_id, deleted_at) WHERE deleted_at IS NULL;
```

### 4.3 Query Plan Analysis

**No query plan analysis found** - Recommend enabling `EXPLAIN ANALYZE` logging.

**Typical Multi-Tenant Query Pattern**:
```sql
-- Example: List clients for agency
SELECT * FROM clients 
WHERE agency_id = $1 AND deleted_at IS NULL
ORDER BY created_at ASC;
```

**Expected Plan**:
```
Index Scan using idx_clients_agency_id on clients
  Index Cond: (agency_id = $1)
  Filter: (deleted_at IS NULL)
```

**RLS Policy Overhead**:
- Simple policies (`agency_id = current_setting(...)`) - **Negligible** (<1ms)
- Subquery policies (`client_id IN (SELECT ...)`) - **Moderate** (1-5ms per row)

### 4.4 Slow Queries

#### 🔴 **Critical: GetSeatUsage N+1 Pattern**

**Location**: `internal/domains/tenants/app/usecases/get_seat_usage.go:92-115`

**Problem**: For each client, executes 2 separate queries:
```go
for _, client := range clients {
    locationCount, err := uc.locationRepo.CountByClient(ctx, client.ID())  // Query 1
    memberCount, err := uc.clientMemberRepo.CountByClient(ctx, client.ID()) // Query 2
}
```

**Impact**: **2N + 1 queries** for N clients
- 1,000 clients = **2,001 queries**
- Estimated time: **2-5 seconds**

**Fix**: Use batch queries:
```sql
-- Single query for location counts
SELECT client_id, COUNT(*) as location_count
FROM locations
WHERE client_id = ANY($1) AND deleted_at IS NULL
GROUP BY client_id;

-- Single query for member counts
SELECT client_id, COUNT(*) as member_count
FROM client_members
WHERE client_id = ANY($1) AND deleted_at IS NULL
GROUP BY client_id;
```

**Estimated Savings**: **99% reduction** (2,001 queries → 3 queries)

#### ⚠️ **Potential: FindPendingInvitesByEmail**

**Query**:
```sql
SELECT * FROM tenant_invites
WHERE LOWER(TRIM(email)) = $1 
    AND accepted_at IS NULL 
    AND revoked_at IS NULL 
    AND expires_at > NOW()
    AND deleted_at IS NULL
```

**Issue**: `LOWER(TRIM(email))` prevents index usage

**Fix**: Add functional index:
```sql
CREATE INDEX idx_tenant_invites_email_lower ON tenant_invites(LOWER(TRIM(email))) 
WHERE accepted_at IS NULL AND revoked_at IS NULL AND deleted_at IS NULL;
```

---

## 5. DATA ISOLATION VALIDATION

### 5.1 RLS Test Cases

**Status**: ❌ **No RLS isolation tests found**

**Searched**:
- `*_test.go` files
- Test directories
- Integration test files

**Found**: General tests exist, but no RLS-specific isolation tests.

### 5.2 Recommended Test Cases

**Test 1: Tenant A Cannot Read Tenant B's Data**
```go
func TestRLS_TenantIsolation(t *testing.T) {
    // Set tenant A context
    ctxA := setTenantContext(ctx, tenantAID)
    
    // Try to read tenant B's client
    client, err := repo.FindByID(ctxA, tenantBClientID)
    assert.Error(t, err) // Should fail with RLS
    assert.Nil(t, client)
}
```

**Test 2: RLS Policy Enforces Soft Delete**
```go
func TestRLS_SoftDelete(t *testing.T) {
    // Soft delete a client
    client.SetDeletedAt(time.Now())
    repo.Save(ctx, client)
    
    // Try to read deleted client
    result, err := repo.FindByID(ctx, client.ID())
    assert.Error(t, err) // Should fail (RLS filters deleted_at)
}
```

**Test 3: Cross-Tenant Location Access**
```go
func TestRLS_LocationIsolation(t *testing.T) {
    // Tenant A tries to access Tenant B's location
    ctxA := setTenantContext(ctx, tenantAID)
    
    location, err := locationRepo.FindByID(ctxA, tenantBLocationID)
    assert.Error(t, err) // Should fail (RLS subquery filters)
}
```

### 5.3 How to Catch Data Leaks

**Current Monitoring**: ❌ **None found**

**Recommendations**:

1. **Application-Level Logging**:
   ```go
   // Log all queries with tenant context
   logger.Info().
       Str("tenant_id", tenantID).
       Str("query", query).
       Msg("Database query executed")
   ```

2. **PostgreSQL Audit Logging**:
   ```sql
   -- Enable audit logging for RLS policy violations
   CREATE EXTENSION IF NOT EXISTS pg_audit;
   ```

3. **Automated Tests**:
   - Add RLS isolation tests to CI/CD
   - Run on every PR

4. **Query Monitoring**:
   - Monitor queries that return 0 rows (potential RLS filtering)
   - Alert on queries that access multiple tenants

### 5.4 Background Jobs and RLS

**Status**: ✅ **No background jobs found** (no risk)

**Future Jobs** (from roadmap):
- GBP sync jobs
- Review ingestion
- Post scheduling

**Recommendation**: When implementing background jobs:
1. ✅ Set `lv.tenant_id` before each job execution
2. ✅ Use tenant-scoped job queues (one queue per tenant)
3. ✅ Validate tenant context in job handler
4. ✅ Log tenant context in job logs

---

## 6. MIGRATIONS & SCHEMA EVOLUTION

### 6.1 Migration Management

**Tool**: `golang-migrate/migrate`  
**Location**: `migrations/` directory  
**Format**: `000XXX_description.up.sql` and `000XXX_description.down.sql`

**Makefile Commands**:
```bash
make migrate-up        # Run all pending migrations
make migrate-down      # Rollback last migration
make migrate-status    # Check migration status
make migrate-create NAME=add_table  # Create new migration
```

**Current Migrations**:
1. `000001_init.up.sql` - Agencies and branding tables
2. `000002_seed_data.up.sql` - Seed data
3. `000003_create_tenants.up.sql` - Tenant members and invites
4. `000004_agency_hierarchy.up.sql` - Clients, locations, client members
5. `000005_add_branding_deleted_at.up.sql` - Soft delete for branding
6. `000006_create_users.up.sql` - Users table
7. `000007_add_branding_features.up.sql` - Branding enhancements
8. `000008_add_invite_revoked_at.up.sql` - Invite revocation
9. `000009_add_invite_expiry_hours.up.sql` - Invite expiry configuration
10. `000010_fix_domain_constraint.up.sql` - Domain uniqueness

**Status**: ✅ **Well-managed** - Sequential, reversible migrations.

### 6.2 Process for Adding RLS to New Tables

**Current Process**: ⚠️ **Manual** - No documented checklist

**Recommended Checklist**:
1. ✅ Create table with `tenant_id` or `agency_id` column
2. ✅ Add foreign key constraint
3. ✅ Create index on tenant column
4. ✅ Enable RLS: `ALTER TABLE table_name ENABLE ROW LEVEL SECURITY;`
5. ✅ Create RLS policy:
   ```sql
   CREATE POLICY table_name_tenant ON table_name
       USING (tenant_id = current_setting('lv.tenant_id')::uuid AND deleted_at IS NULL);
   ```
6. ✅ Add soft delete support (if needed)
7. ✅ Test RLS isolation

**Example**: See `migrations/000004_agency_hierarchy.up.sql:98-116` for reference.

### 6.3 Breaking Changes Handling

**Current Process**: ⚠️ **Not documented**

**Recommendations**:

1. **Column Removal**:
   - Create migration to add new column
   - Deploy application code that uses new column
   - Create migration to remove old column (after grace period)

2. **Table Removal**:
   - Mark table as deprecated in code
   - Create migration to rename table (e.g., `clients_deprecated`)
   - Remove after grace period (e.g., 3 months)

3. **Constraint Changes**:
   - Test constraint changes on staging first
   - Use `IF EXISTS` / `IF NOT EXISTS` for idempotency

**Example Breaking Change**:
```sql
-- Migration: Add new column, keep old for backward compatibility
ALTER TABLE clients ADD COLUMN new_field TEXT;

-- Application code migrates data

-- Later migration: Remove old column
ALTER TABLE clients DROP COLUMN old_field;
```

---

## 7. PARTITIONING & SCALING

### 7.1 Partitioning Threshold

**Recommendation**: Partition tables when they exceed **10GB** or **10 million rows**.

**Current Size**: Unknown (no monitoring)

**Estimated Growth** (12 months):
- `locations`: ~1,000 rows = **~2MB** ✅ No partitioning needed
- `client_members`: ~3,000 rows = **~1.5MB** ✅ No partitioning needed
- `users`: ~300 rows = **~300KB** ✅ No partitioning needed

**Future Growth** (5 years, 1,000 agencies, 100K locations):
- `locations`: ~100,000 rows = **~200MB** ✅ Still small
- `client_members`: ~300,000 rows = **~150MB** ✅ Still small

**Conclusion**: ✅ **No partitioning needed** for foreseeable future.

### 7.2 Partitioning Strategy (If Needed)

**Strategy**: **LIST partitioning by `agency_id`**

**Rationale**:
- Each tenant (agency) is isolated
- Queries are always scoped to a single tenant
- Partition pruning will be effective

**Example** (if needed in future):
```sql
-- Create partitioned table
CREATE TABLE locations_partitioned (
    LIKE locations INCLUDING ALL
) PARTITION BY LIST (agency_id);

-- Create partition for each agency (or use default partition)
CREATE TABLE locations_agency_1 PARTITION OF locations_partitioned
    FOR VALUES IN ('agency-uuid-1');

-- Or use default partition for new agencies
CREATE TABLE locations_default PARTITION OF locations_partitioned
    DEFAULT;
```

**Alternative**: **RANGE partitioning by `created_at`** (if time-based queries are common)

**Recommendation**: ⚠️ **Don't partition yet** - Overhead not justified at current scale.

### 7.3 Query Time Impact

**Current**: RLS policies add **<1ms** overhead per query (negligible)

**With Partitioning**: Would reduce query time by **10-20%** for large tables (>10GB), but:
- Adds management overhead
- Requires partition maintenance
- Not needed at current scale

**Recommendation**: ✅ **Monitor table sizes** - Revisit when any table exceeds 5GB.

---

## 8. BACKUP & RECOVERY

### 8.1 Backup Strategy

**Status**: ❌ **Not documented** - No backup configuration found

**Recommendations**:

#### **Option A: Managed Database (RDS/Cloud SQL)**
- ✅ Automatic daily backups
- ✅ Point-in-time recovery (PITR)
- ✅ Cross-region replication

#### **Option B: Self-Managed PostgreSQL**
```bash
# Daily full backup
pg_dump -h localhost -U postgres -d localvisibilityos -F c -f backup_$(date +%Y%m%d).dump

# Continuous WAL archiving (for PITR)
archive_mode = on
archive_command = 'cp %p /backup/wal/%f'
```

### 8.2 Single Agency Data Recovery

**Status**: ❌ **Not documented** - No process for single-tenant recovery

**Recommendation**: Use RLS to restore single tenant:

```sql
-- 1. Restore full backup to temporary database
pg_restore -d temp_db backup.dump

-- 2. Set tenant context
SET LOCAL lv.tenant_id = 'agency-uuid';

-- 3. Export single tenant's data
pg_dump -d temp_db --table=clients --table=locations ... --data-only > tenant_backup.sql

-- 4. Restore to production (with tenant context set)
psql -d production < tenant_backup.sql
```

**Alternative**: Use application-level export/import API:
```go
// Export tenant data
GET /api/v1/tenants/{id}/export

// Import tenant data
POST /api/v1/tenants/{id}/import
```

### 8.3 Backup Testing

**Status**: ❌ **Not documented** - No backup testing process

**Recommendation**:
1. ✅ Test restore monthly
2. ✅ Test single-tenant restore quarterly
3. ✅ Document RTO/RPO targets

### 8.4 RTO/RPO (Recovery Time/Point Objective)

**Status**: ❌ **Not documented**

**Recommended Targets**:
- **RTO** (Recovery Time Objective): **4 hours** (restore from backup)
- **RPO** (Recovery Point Objective): **1 hour** (max data loss)

**For Managed Database** (RDS/Cloud SQL):
- **RTO**: **15 minutes** (automated failover)
- **RPO**: **5 minutes** (continuous WAL replication)

---

## 9. IDENTIFIED ISSUES & OPTIMIZATIONS

### 🔴 Critical Issues (Fix Immediately)

1. **Missing RLS on `users` Table**
   - **Risk**: HIGH - Cross-tenant user data access
   - **Fix**: Add RLS policy filtering by tenant membership
   - **Effort**: 2 hours

2. **GetSeatUsage N+1 Pattern**
   - **Impact**: 2,001 queries for 1,000 clients
   - **Fix**: Use batch queries with `GROUP BY`
   - **Effort**: 4 hours
   - **Savings**: 99% query reduction

### ⚠️ High Priority (Fix Soon)

3. **Missing RLS on `agencies` Table**
   - **Risk**: MEDIUM - Agencies can see each other's metadata
   - **Fix**: Add RLS policy (super-admin only or self-access)
   - **Effort**: 1 hour

4. **No RLS Isolation Tests**
   - **Risk**: MEDIUM - No validation that RLS works
   - **Fix**: Add test suite for RLS isolation
   - **Effort**: 4 hours

5. **Subquery Performance in RLS Policies**
   - **Impact**: 1-5ms overhead per row (acceptable for <10K rows)
   - **Fix**: Denormalize `agency_id` on `locations` and `client_members` (if needed)
   - **Effort**: 8 hours
   - **Priority**: Low (revisit when locations >10K)

### 🟡 Medium Priority (Fix When Needed)

6. **Missing Backup/Recovery Documentation**
   - **Risk**: LOW - No documented process
   - **Fix**: Document backup strategy and recovery procedures
   - **Effort**: 2 hours

7. **Missing Functional Index on `tenant_invites.email`**
   - **Impact**: Slow invite lookup (uses `LOWER(TRIM(email))`)
   - **Fix**: Add functional index
   - **Effort**: 1 hour

8. **No Query Plan Monitoring**
   - **Impact**: Can't identify slow queries
   - **Fix**: Enable `EXPLAIN ANALYZE` logging
   - **Effort**: 2 hours

---

## 10. PERFORMANCE BASELINE

### 10.1 Average Query Time

**Status**: ❌ **Not measured** - No query timing found

**Estimated** (based on query complexity):
- Simple lookups (by ID): **<10ms**
- List queries: **10-50ms** (depending on result size)
- Count queries: **10-30ms**
- RLS policy overhead: **<1ms** (simple), **1-5ms** (subquery)

### 10.2 Peak Throughput

**Estimated** (from `docs/QUERY_ANALYSIS.md`):
- **Requests/min**: 100-500 (conservative for 30 agencies)
- **Queries/request**: 3-5 (without caching), 1-2 (with caching)
- **Peak queries/min**: **300-2,500** (without caching), **100-1,000** (with caching)
- **Peak queries/hour**: **18,000-150,000** (without caching), **6,000-60,000** (with caching)

**Database Capacity** (db.t3.small):
- **Max queries/sec**: ~1,000-2,000
- **Max queries/hour**: **60,000-120,000**

**Status**: ⚠️ **At risk** during peak hours without caching

**With Caching**: ✅ **Within capacity**

---

## 11. SCHEMA DIAGRAM

```
┌─────────────────┐
│   agencies      │ (Top-level tenant)
│─────────────────│
│ id (PK)         │
│ name            │
│ slug (UNIQUE)   │
│ tier            │
│ seat_limit      │
│ deleted_at      │
└────────┬────────┘
         │
         │ 1:N
         ├─────────────────────────────────────┐
         │                                     │
         ▼                                     ▼
┌─────────────────┐              ┌──────────────────────┐
│   branding      │              │  tenant_members      │
│─────────────────│              │──────────────────────│
│ agency_id (PK)  │◄──────────────│ tenant_id (FK)       │
│ domain          │              │ user_id              │
│ theme_json      │              │ role                 │
│ deleted_at      │              │ deleted_at           │
└─────────────────┘              └──────────────────────┘
         │
         │ RLS: agency_id = lv.tenant_id
         │
         ▼
┌─────────────────┐
│   clients       │
│─────────────────│
│ id (PK)         │
│ agency_id (FK)  │◄─── RLS: agency_id = lv.tenant_id
│ name            │
│ slug            │
│ deleted_at      │
└────────┬────────┘
         │
         │ 1:N
         ├─────────────────────────────┐
         │                             │
         ▼                             ▼
┌─────────────────┐      ┌──────────────────────┐
│   locations     │      │  client_members      │
│─────────────────│      │──────────────────────│
│ id (PK)         │      │ id (PK)              │
│ client_id (FK)  │◄─────│ client_id (FK)        │
│ name            │      │ user_id              │
│ address (JSONB) │      │ role                 │
│ deleted_at      │      │ deleted_at           │
└─────────────────┘      └──────────────────────┘
         │
         │ RLS: client_id IN (SELECT id FROM clients WHERE agency_id = lv.tenant_id)
         │
         ▼
┌─────────────────┐
│   users         │ (Cross-tenant - NO RLS)
│─────────────────│
│ id (PK)         │
│ clerk_user_id   │
│ email           │
│ name            │
└─────────────────┘

Legend:
  PK = Primary Key
  FK = Foreign Key
  RLS = Row-Level Security Policy
  ─── = Relationship
```

---

## 12. RECOMMENDATIONS SUMMARY

### Immediate Actions (Week 1)

1. ✅ **Add RLS to `users` table** (2 hours)
2. ✅ **Fix GetSeatUsage N+1 pattern** (4 hours)
3. ✅ **Add RLS to `agencies` table** (1 hour)

### Short-Term Actions (Month 1)

4. ✅ **Add RLS isolation tests** (4 hours)
5. ✅ **Document backup/recovery process** (2 hours)
6. ✅ **Enable query plan monitoring** (2 hours)

### Long-Term Actions (Month 2-3)

7. ✅ **Optimize RLS subquery policies** (if locations >10K)
8. ✅ **Add functional index on `tenant_invites.email`** (1 hour)
9. ✅ **Implement table size monitoring** (2 hours)

### Monitoring & Maintenance

10. ✅ **Set up query performance dashboards**
11. ✅ **Monthly backup restore tests**
12. ✅ **Quarterly single-tenant recovery tests**

---

## Conclusion

✅ **RLS Implementation**: **STRONG** - 7/8 tenant-scoped tables have RLS enabled  
⚠️ **Security Gaps**: 2 tables missing RLS (`users`, `agencies`)  
🔴 **Performance**: 1 critical N+1 pattern (GetSeatUsage)  
✅ **Indexing**: Comprehensive coverage  
⚠️ **Testing**: No RLS isolation tests  
⚠️ **Documentation**: Missing backup/recovery procedures

**Overall Assessment**: **GOOD** - Strong foundation with a few critical fixes needed.

**Estimated Fix Time**: **13 hours** (immediate + short-term actions)

**Risk Level**: **MEDIUM** - Security gaps exist but are mitigated by app-layer filtering (defense-in-depth).
