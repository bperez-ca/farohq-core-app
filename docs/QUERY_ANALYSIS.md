# Database Query Pattern Analysis & Optimization Report

**Date**: 2025-01-27  
**Target**: 1,000 locations, 30 agencies at Month 12  
**Budget**: <$50/month (RDS db.t3.small)  
**Goal**: 80%+ cache hit rate for frequently accessed data

---

## Executive Summary

This analysis identifies **critical N+1 patterns**, **missing cache opportunities**, and **tenant isolation risks** that could significantly impact database costs and performance at scale. Key findings:

- **1 critical N+1 pattern** in seat usage calculation (2N+1 queries for N clients)
- **3 queries executed on every API request** (should be cached)
- **No slow query logging** currently implemented
- **RLS + app-layer filtering** provides defense-in-depth but adds overhead
- **No external API integrations** found (Data4SEO, Meta APIs not yet implemented)

---

## 1. QUERY VOLUME & FREQUENCY

### Top 10 Most Executed Queries (by Call Frequency)

Based on codebase analysis, these queries are executed most frequently:

#### 1. **Tenant Resolution by Domain** (Every Request)
```sql
SELECT agency_id::text FROM branding WHERE domain = $1 AND deleted_at IS NULL
```
- **Location**: `internal/platform/tenant/resolver.go:70,265`
- **Frequency**: **Every authenticated request** (via middleware)
- **Estimated calls/day**: ~10,000-50,000 (assuming 100-500 req/min)
- **Cache opportunity**: ✅ **HIGH** - Domain→tenant mapping rarely changes
- **Current caching**: ❌ None (queries DB on every request)

#### 2. **Get User Tenant IDs** (Every Authenticated Request)
```sql
SELECT DISTINCT ON (tenant_id) tenant_id::text, created_at
FROM tenant_members
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY tenant_id, created_at ASC
```
- **Location**: `internal/platform/tenant/resolver.go:182-187`
- **Frequency**: **Every authenticated request** (tenant validation)
- **Estimated calls/day**: ~10,000-50,000
- **Cache opportunity**: ✅ **HIGH** - User tenant membership changes infrequently
- **Current caching**: ✅ **EXISTS** - Redis cache with 5min TTL (`internal/platform/tenant/cache.go`)

#### 3. **Client Validation Query** (Every Request with client_id)
```sql
SELECT id::text FROM clients WHERE id = $1 AND agency_id = $2 AND deleted_at IS NULL
```
- **Location**: `internal/platform/httpserver/tenant.go:90-94,287-291`
- **Frequency**: **Every request with X-Client-ID header or query param**
- **Estimated calls/day**: ~5,000-25,000 (assuming 50% of requests include client_id)
- **Cache opportunity**: ✅ **MEDIUM** - Client existence rarely changes
- **Current caching**: ❌ None

#### 4. **User Lookup by Clerk ID** (Every Authenticated Request)
```sql
SELECT id, clerk_user_id, email, first_name, last_name, full_name, image_url, phone_numbers, created_at, updated_at, last_sign_in_at
FROM users
WHERE clerk_user_id = $1
```
- **Location**: `internal/domains/users/infra/db/user_repository.go:31-35`
- **Frequency**: **Every authenticated request** (auth middleware)
- **Estimated calls/day**: ~10,000-50,000
- **Cache opportunity**: ✅ **HIGH** - User data changes infrequently
- **Current caching**: ❌ None

#### 5. **Brand Theme by Domain** (Every Page Load)
```sql
SELECT agency_id, domain, subdomain, domain_type, website, verified_at, logo_url, favicon_url, 
       primary_color, secondary_color, theme_json, hide_powered_by, email_domain, 
       cloudflare_zone_id, domain_verification_token, ssl_status, updated_at
FROM branding
WHERE domain = $1 AND domain_type = 'custom'
```
- **Location**: `internal/domains/brand/infra/db/brand_repository.go:125-132`
- **Frequency**: **Every page load** (frontend fetches theme)
- **Estimated calls/day**: ~5,000-25,000
- **Cache opportunity**: ✅ **HIGH** - Theme changes rarely
- **Current caching**: ✅ **EXISTS** - Frontend localStorage (15min TTL), but backend not cached

#### 6. **List Clients by Agency** (Dashboard/Agency Pages)
```sql
SELECT id, agency_id, name, slug, tier, status, created_at, updated_at, deleted_at
FROM clients
WHERE agency_id = $1 AND deleted_at IS NULL
ORDER BY created_at ASC
```
- **Location**: `internal/domains/tenants/infra/db/client_repository.go:148-154`
- **Frequency**: **Every dashboard load**
- **Estimated calls/day**: ~1,000-5,000
- **Cache opportunity**: ✅ **MEDIUM** - Client list changes infrequently
- **Current caching**: ❌ None

#### 7. **List Locations by Client** (Location Pages)
```sql
SELECT id, client_id, name, address, phone, business_hours, categories, is_active, created_at, updated_at, deleted_at
FROM locations
WHERE client_id = $1 AND deleted_at IS NULL
ORDER BY created_at ASC
```
- **Location**: `internal/domains/tenants/infra/db/location_repository.go:122-129`
- **Frequency**: **Every location list page load**
- **Estimated calls/day**: ~2,000-10,000
- **Cache opportunity**: ⚠️ **LOW** - Locations change frequently (active business data)
- **Current caching**: ❌ None

#### 8. **RLS Context Setting** (Every Request)
```sql
SELECT set_config('lv.tenant_id', $1, true)
SELECT set_config('lv.client_id', $1, true)
```
- **Location**: `internal/platform/httpserver/tenant.go:76,103`
- **Frequency**: **Every request** (2 queries per request)
- **Estimated calls/day**: ~20,000-100,000 (2 queries × requests)
- **Cache opportunity**: ❌ **NONE** - Required for RLS security
- **Note**: These are `SELECT` statements (not data queries), but still add overhead

#### 9. **Find Pending Invites by Email** (Invite Acceptance Flow)
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
- **Location**: `internal/domains/tenants/infra/db/invite_repository.go:255-264`
- **Frequency**: **Every invite acceptance**
- **Estimated calls/day**: ~100-500
- **Cache opportunity**: ⚠️ **LOW** - Invites are time-sensitive
- **Current caching**: ❌ None

#### 10. **Tenant Member Lookup** (Access Validation)
```sql
SELECT COUNT(*) FROM tenant_members
WHERE user_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
```
- **Location**: `internal/platform/tenant/resolver.go:234-238`
- **Frequency**: **Every tenant-scoped request** (access validation)
- **Estimated calls/day**: ~5,000-25,000
- **Cache opportunity**: ✅ **MEDIUM** - Can be inferred from cached user tenant list
- **Current caching**: ❌ None (but could use cached user tenant IDs)

### Queries Triggered on Every API Request

These queries run on **every single API request** and should be cached:

1. ✅ **Tenant Resolution by Domain** - Currently **NOT cached** (critical)
2. ✅ **Get User Tenant IDs** - Currently **cached** (5min TTL via Redis)
3. ✅ **User Lookup by Clerk ID** - Currently **NOT cached** (critical)
4. ✅ **Client Validation** (if client_id provided) - Currently **NOT cached**
5. ✅ **RLS Context Setting** - Cannot be cached (security requirement)

**Impact**: Without caching, **3-4 database queries per request** = 30,000-200,000 queries/day at 100-500 req/min.

### Background Jobs

**No background jobs found** in the current codebase. Based on roadmap (`lvos/docs/FARO_IMPLEMENTATION_PLAN.md`), planned jobs include:
- GBP sync jobs (not yet implemented)
- Review ingestion (not yet implemented)
- Post scheduling (not yet implemented)

### Peak Query Throughput

**Estimated peak throughput** (business hours, 9am-5pm):
- **Requests/min**: 100-500 (conservative estimate for 30 agencies)
- **Queries/request**: 3-5 (without caching), 1-2 (with caching)
- **Peak queries/min**: **300-2,500** (without caching), **100-1,000** (with caching)
- **Peak queries/hour**: **18,000-150,000** (without caching), **6,000-60,000** (with caching)

**db.t3.small capacity**: ~1,000-2,000 queries/sec = **60,000-120,000 queries/hour**
- **Without caching**: ⚠️ **At risk** during peak hours
- **With caching**: ✅ **Within capacity**

---

## 2. N+1 & INEFFICIENCY DETECTION

### Critical N+1 Pattern Found

#### **N+1 in GetSeatUsage Use Case** 🔴 **CRITICAL**

**Location**: `internal/domains/tenants/app/usecases/get_seat_usage.go:92-115`

**Problem**: For each client in an agency, the code executes 2 separate queries:

```go
for _, client := range clients {
    locationCount, err := uc.locationRepo.CountByClient(ctx, client.ID())  // Query 1
    memberCount, err := uc.clientMemberRepo.CountByClient(ctx, client.ID()) // Query 2
    // ...
}
```

**Query Pattern**:
- 1 query to list clients: `SELECT * FROM clients WHERE agency_id = $1`
- N queries for location counts: `SELECT COUNT(*) FROM locations WHERE client_id = $1` (N times)
- N queries for member counts: `SELECT COUNT(*) FROM client_members WHERE client_id = $1` (N times)

**Total**: **2N + 1 queries** for N clients

**Impact at Scale**:
- 30 agencies × ~33 clients/agency = 1,000 clients
- **2,001 queries** to calculate seat usage for one agency
- If called every dashboard load: **2,000+ queries/day** just for seat usage

**Fix**: Use batch queries with `GROUP BY`:

```sql
-- Single query for location counts
SELECT client_id, COUNT(*) as location_count
FROM locations
WHERE client_id = ANY($1) AND deleted_at IS NULL
GROUP BY client_id

-- Single query for member counts  
SELECT client_id, COUNT(*) as member_count
FROM client_members
WHERE client_id = ANY($1) AND deleted_at IS NULL
GROUP BY client_id
```

**Estimated Cost Savings**: **99% reduction** (2,001 queries → 3 queries)

### Other Potential N+1 Patterns

#### **List Locations with Client Details** (Potential)
**Location**: `internal/domains/tenants/infra/db/location_repository.go:122-175`

**Current**: Locations are listed without client details. If frontend needs client info, it would trigger N queries.

**Status**: ⚠️ **Not currently a problem**, but could become one if frontend fetches client details per location.

**Prevention**: Use JOIN or batch fetch if client details are needed.

#### **List Client Members with User Details** (Potential)
**Location**: `internal/domains/tenants/infra/db/client_member_repository.go:147-201`

**Current**: Members are listed without user details. If frontend needs user info, it would trigger N queries.

**Status**: ⚠️ **Not currently a problem**, but could become one.

**Prevention**: Use JOIN or batch fetch if user details are needed.

### Batch Query Opportunities

**Current State**: Most repositories use single-entity queries. No batch query methods found.

**Opportunities**:
1. ✅ **Batch count queries** (seat usage fix above)
2. ✅ **Batch user lookups** (if multiple users needed)
3. ✅ **Batch client validations** (if multiple clients validated)

---

## 3. TENANT FILTERING

### Tenant Isolation Enforcement

#### **Defense-in-Depth Approach** ✅

The codebase uses **both RLS (Row Level Security) and app-layer filtering**:

1. **RLS (Database Layer)**:
   - Location: `internal/platform/httpserver/tenant.go:76,103`
   - Method: `SELECT set_config('lv.tenant_id', $1, true)` on every request
   - **Status**: ✅ Implemented

2. **App-Layer Filtering**:
   - All repository queries include explicit `WHERE` clauses with `tenant_id` or `agency_id`
   - Example: `WHERE agency_id = $1 AND deleted_at IS NULL`

**Security Assessment**: ✅ **Strong** - Defense-in-depth prevents data leaks even if RLS fails.

### Potential Data Leak Risks

#### **Risk 1: Missing Tenant Filter in Query** ⚠️ **LOW RISK**

**Analysis**: Reviewed all repository queries. All queries that should filter by tenant include explicit filters:
- ✅ `client_repository.go`: All queries include `agency_id = $1`
- ✅ `location_repository.go`: All queries include `client_id = $1` (which implies tenant via FK)
- ✅ `tenant_member_repository.go`: All queries include `tenant_id = $1`
- ✅ `brand_repository.go`: Queries by `agency_id` or domain (domain is tenant-specific)

**Status**: ✅ **No leaks found** - All queries properly filtered.

#### **Risk 2: RLS Not Set on Connection** ⚠️ **MEDIUM RISK**

**Location**: `internal/platform/httpserver/tenant.go:264-281`

**Issue**: If `conn.Acquire()` fails or `set_config()` fails, the request continues without RLS context.

**Current Behavior**: 
- If RLS setup fails, request returns 500 error (line 267-280)
- ✅ **Safe** - Request fails rather than proceeding without RLS

**Recommendation**: ✅ **Current implementation is safe**.

#### **Risk 3: Cross-Tenant Query via Domain Resolution** ⚠️ **LOW RISK**

**Location**: `internal/platform/tenant/resolver.go:257-280`

**Issue**: Domain→tenant resolution queries `branding` table without explicit tenant filter.

**Analysis**: 
- Query: `SELECT agency_id::text FROM branding WHERE domain = $1`
- **Risk**: If domain is not unique, could return wrong tenant
- **Mitigation**: Database should have `UNIQUE` constraint on `domain` (verify in migrations)

**Status**: ⚠️ **Verify uniqueness constraint exists**.

### Database-Level RLS

**Current Implementation**: RLS is set via session variables (`lv.tenant_id`, `lv.client_id`).

**Missing**: No actual PostgreSQL RLS policies found in migrations. RLS relies on:
1. Session variables set by application
2. Application queries manually filtering by tenant_id

**Recommendation**: 
- ✅ **Current approach works** but adds overhead (2 queries per request)
- ⚠️ **Consider**: Native PostgreSQL RLS policies for additional security layer

---

## 4. CACHE OPPORTUNITIES

### High-Impact Cache Opportunities (Ranked)

#### 1. **Tenant Resolution by Domain** 🔴 **CRITICAL**
- **Query**: `SELECT agency_id FROM branding WHERE domain = $1`
- **Frequency**: Every request
- **Cache TTL**: 15 minutes (domains rarely change)
- **Estimated Hit Rate**: 95%+
- **Cost Savings**: ~45,000 queries/day → ~2,250 queries/day (**95% reduction**)
- **Implementation**: Add Redis cache in `internal/platform/tenant/resolver.go:257-280`

#### 2. **User Lookup by Clerk ID** 🔴 **CRITICAL**
- **Query**: `SELECT * FROM users WHERE clerk_user_id = $1`
- **Frequency**: Every authenticated request
- **Cache TTL**: 5 minutes (user data changes infrequently)
- **Estimated Hit Rate**: 90%+
- **Cost Savings**: ~40,000 queries/day → ~4,000 queries/day (**90% reduction**)
- **Implementation**: Add Redis cache in `internal/domains/users/infra/db/user_repository.go:30-35`

#### 3. **Brand Theme by Domain** 🟡 **HIGH**
- **Query**: `SELECT * FROM branding WHERE domain = $1`
- **Frequency**: Every page load
- **Cache TTL**: 15 minutes (already cached on frontend, but backend not cached)
- **Estimated Hit Rate**: 95%+
- **Cost Savings**: ~20,000 queries/day → ~1,000 queries/day (**95% reduction**)
- **Implementation**: Add Redis cache in `internal/domains/brand/infra/db/brand_repository.go:125-132`

#### 4. **Client Validation** 🟡 **MEDIUM**
- **Query**: `SELECT id FROM clients WHERE id = $1 AND agency_id = $2`
- **Frequency**: Every request with client_id
- **Cache TTL**: 10 minutes
- **Estimated Hit Rate**: 85%+
- **Cost Savings**: ~20,000 queries/day → ~3,000 queries/day (**85% reduction**)
- **Implementation**: Add Redis cache in `internal/platform/httpserver/tenant.go:90-94`

#### 5. **List Clients by Agency** 🟢 **MEDIUM**
- **Query**: `SELECT * FROM clients WHERE agency_id = $1`
- **Frequency**: Dashboard loads
- **Cache TTL**: 5 minutes
- **Estimated Hit Rate**: 80%+
- **Cost Savings**: ~4,000 queries/day → ~800 queries/day (**80% reduction**)
- **Implementation**: Add Redis cache in `internal/domains/tenants/infra/db/client_repository.go:148-154`

#### 6. **User Tenant IDs** ✅ **ALREADY CACHED**
- **Query**: `SELECT tenant_id FROM tenant_members WHERE user_id = $1`
- **Frequency**: Every authenticated request
- **Cache TTL**: 5 minutes (current)
- **Status**: ✅ **Already implemented** (`internal/platform/tenant/cache.go`)

### Cache-Incompatible Data

**Write-Heavy Data** (should NOT be cached):
- ❌ **Locations** - Active business data, changes frequently
- ❌ **Client Members** - Membership changes frequently
- ❌ **Invites** - Time-sensitive, expires frequently
- ❌ **Reviews** (when implemented) - Real-time data
- ❌ **Posts** (when implemented) - Real-time data

### Hot Paths for Caching

**Identified Hot Paths**:
1. ✅ **Tenant resolution** - Every request (cache domain→tenant mapping)
2. ✅ **User lookup** - Every authenticated request (cache user data)
3. ✅ **Brand theme** - Every page load (cache theme JSON)
4. ✅ **Client validation** - Every request with client_id (cache client existence)

**Multi-Tenant Aggregations** (Precomputation Opportunities):
- ⚠️ **Seat usage** - Currently calculated on-demand (N+1 pattern). Could be precomputed and cached.
- ⚠️ **Client counts** - Could be cached per agency
- ⚠️ **Location counts** - Could be cached per client

---

## 5. EXTERNAL API CALLS

### Data4SEO API Calls

**Status**: ❌ **Not found in codebase**

**Expected Locations** (based on roadmap):
- GBP profile sync
- Rankings data
- Snapshots

**Recommendation**: When implemented, cache API responses:
- **GBP profiles**: Cache 15 minutes (profiles change infrequently)
- **Rankings**: Cache 1 hour (rankings update daily)
- **Snapshots**: Cache 24 hours (historical data)

### Meta Cloud API Calls

**Status**: ❌ **Not found in codebase**

**Expected Locations** (based on roadmap):
- WhatsApp Business API
- Instagram Messaging API
- Facebook Messenger API

**Recommendation**: When implemented:
- **Message threads**: Cache 5 minutes (active conversations)
- **User profiles**: Cache 15 minutes
- **Webhook validation**: No caching (security-critical)

### Current API Response Caching

**Status**: ❌ **No external API integrations found**

**When Implemented**: Use Redis for API response caching with appropriate TTLs.

---

## 6. SLOW QUERY LOGGING

### Current Implementation

**Status**: ❌ **No slow query logging found**

**Location**: No query timing or logging found in:
- Repository methods
- Database pool configuration
- Middleware

### Recommendations

#### **1. Enable PostgreSQL Slow Query Logging**

Add to `postgresql.conf` or connection string:
```sql
log_min_duration_statement = 100  -- Log queries >100ms
log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '
```

#### **2. Application-Level Query Timing**

Add query timing middleware:

```go
// In internal/platform/db/pool.go or middleware
func LogSlowQueries(ctx context.Context, query string, duration time.Duration) {
    if duration > 100*time.Millisecond {
        logger.Warn().
            Str("query", query).
            Dur("duration", duration).
            Msg("Slow query detected")
    }
}
```

#### **3. Expected Slow Queries**

Based on query patterns, these queries may be slow at scale:
- ⚠️ **GetSeatUsage** - N+1 pattern (2,001 queries for 1,000 clients)
- ⚠️ **List Locations** - Could be slow with 1,000+ locations
- ⚠️ **Find Pending Invites** - Uses `LOWER(TRIM(email))` (no index on computed value)

### Current Average Query Latency

**No metrics found** - Recommend implementing query timing.

**Estimated Latency** (based on query complexity):
- Simple lookups (by ID): <10ms
- List queries: 10-50ms (depending on result size)
- Count queries: 10-30ms
- **N+1 patterns**: 100-500ms+ (depending on N)

---

## 7. TOP 10 MOST EXPENSIVE QUERIES

### By Cost (Query Frequency × Complexity)

1. **Tenant Resolution by Domain** - 50,000 queries/day × 1ms = **50s/day**
2. **Get User Tenant IDs** - 50,000 queries/day × 2ms = **100s/day** (cached, but cache misses)
3. **User Lookup by Clerk ID** - 50,000 queries/day × 1ms = **50s/day**
4. **Client Validation** - 25,000 queries/day × 1ms = **25s/day**
5. **GetSeatUsage (N+1)** - 100 queries/day × 2,000ms = **200s/day** 🔴 **CRITICAL**
6. **List Clients by Agency** - 5,000 queries/day × 10ms = **50s/day**
7. **List Locations by Client** - 10,000 queries/day × 20ms = **200s/day**
8. **RLS Context Setting** - 100,000 queries/day × 0.5ms = **50s/day**
9. **Brand Theme Lookup** - 25,000 queries/day × 2ms = **50s/day**
10. **Tenant Member Validation** - 25,000 queries/day × 1ms = **25s/day**

**Total Estimated Database Time**: ~800s/day = **13.3 minutes/day**

**At db.t3.small capacity**: Well within limits, but **GetSeatUsage N+1** is a critical bottleneck.

---

## 8. SPECIFIC N+1 PATTERNS WITH CODE LOCATIONS

### Pattern 1: GetSeatUsage - Client Loop 🔴 **CRITICAL**

**File**: `internal/domains/tenants/app/usecases/get_seat_usage.go`

**Lines**: 92-115

**Code**:
```go
for _, client := range clients {
    locationCount, err := uc.locationRepo.CountByClient(ctx, client.ID())  // N queries
    memberCount, err := uc.clientMemberRepo.CountByClient(ctx, client.ID()) // N queries
}
```

**Impact**: 2N+1 queries for N clients

**Fix**: Implement batch count methods:
```go
// New method in LocationRepository
func (r *LocationRepository) CountByClients(ctx context.Context, clientIDs []uuid.UUID) (map[uuid.UUID]int, error)

// New method in ClientMemberRepository  
func (r *ClientMemberRepository) CountByClients(ctx context.Context, clientIDs []uuid.UUID) (map[uuid.UUID]int, error)
```

**Estimated Savings**: 99% reduction (2,001 queries → 3 queries)

---

## 9. CACHING OPPORTUNITIES RANKED BY IMPACT

### Tier 1: Critical (Implement Immediately)

1. **Tenant Resolution by Domain**
   - **Impact**: 95% reduction in queries
   - **Effort**: Low (add Redis cache)
   - **ROI**: Very High

2. **User Lookup by Clerk ID**
   - **Impact**: 90% reduction in queries
   - **Effort**: Low (add Redis cache)
   - **ROI**: Very High

3. **Fix GetSeatUsage N+1**
   - **Impact**: 99% reduction in queries
   - **Effort**: Medium (add batch methods)
   - **ROI**: Very High

### Tier 2: High (Implement Soon)

4. **Brand Theme by Domain**
   - **Impact**: 95% reduction in queries
   - **Effort**: Low (add Redis cache)
   - **ROI**: High

5. **Client Validation**
   - **Impact**: 85% reduction in queries
   - **Effort**: Low (add Redis cache)
   - **ROI**: High

### Tier 3: Medium (Implement When Needed)

6. **List Clients by Agency**
   - **Impact**: 80% reduction in queries
   - **Effort**: Low (add Redis cache)
   - **ROI**: Medium

7. **Precompute Seat Usage**
   - **Impact**: Eliminates N+1 entirely
   - **Effort**: Medium (add background job)
   - **ROI**: Medium

---

## 10. RECOMMENDATIONS FOR QUERY OPTIMIZATION

### Immediate Actions (Week 1)

1. ✅ **Fix GetSeatUsage N+1 Pattern**
   - Add batch count methods to repositories
   - Refactor use case to use batch queries
   - **Estimated Savings**: 99% reduction (2,001 → 3 queries)

2. ✅ **Cache Tenant Resolution by Domain**
   - Add Redis cache in `tenant/resolver.go`
   - TTL: 15 minutes
   - **Estimated Savings**: 95% reduction (50,000 → 2,500 queries/day)

3. ✅ **Cache User Lookup by Clerk ID**
   - Add Redis cache in `users/infra/db/user_repository.go`
   - TTL: 5 minutes
   - **Estimated Savings**: 90% reduction (50,000 → 5,000 queries/day)

### Short-Term Actions (Month 1)

4. ✅ **Cache Brand Theme Lookup**
   - Add Redis cache in `brand/infra/db/brand_repository.go`
   - TTL: 15 minutes
   - **Estimated Savings**: 95% reduction (25,000 → 1,250 queries/day)

5. ✅ **Cache Client Validation**
   - Add Redis cache in `httpserver/tenant.go`
   - TTL: 10 minutes
   - **Estimated Savings**: 85% reduction (25,000 → 3,750 queries/day)

6. ✅ **Implement Slow Query Logging**
   - Add query timing middleware
   - Enable PostgreSQL slow query log
   - **Benefit**: Identify bottlenecks early

### Long-Term Actions (Month 2-3)

7. ✅ **Add Database Indexes**
   - Index on `branding.domain` (if not exists)
   - Index on `users.clerk_user_id` (if not exists)
   - Index on `tenant_invites.email` (for case-insensitive lookup)

8. ✅ **Consider Read Replicas**
   - If query volume exceeds db.t3.small capacity
   - Route read queries to replica
   - **Benefit**: Scale reads independently

9. ✅ **Implement Query Result Caching**
   - Cache list queries (clients, locations) with short TTL
   - Invalidate on writes
   - **Benefit**: Reduce load during peak hours

### Estimated Cost Savings

**Current State** (without optimizations):
- Estimated queries/day: ~200,000
- Database load: Medium-High
- Risk: Exceeds db.t3.small capacity during peaks

**After Optimizations**:
- Estimated queries/day: ~30,000 (85% reduction)
- Database load: Low
- Risk: ✅ Well within db.t3.small capacity

**Monthly Cost Impact**:
- **Current**: db.t3.small = ~$15-20/month
- **After**: db.t3.small = ~$15-20/month (same, but with headroom)
- **Savings**: Prevents need to upgrade to db.t3.medium ($30-40/month)
- **Avoided Cost**: **$15-20/month** (by staying on small instance)

---

## 11. IMPLEMENTATION PRIORITY

### Priority 1: Critical (Do First)
1. Fix GetSeatUsage N+1 pattern
2. Cache tenant resolution by domain
3. Cache user lookup by Clerk ID

### Priority 2: High (Do Soon)
4. Cache brand theme lookup
5. Cache client validation
6. Implement slow query logging

### Priority 3: Medium (Do When Needed)
7. Cache list queries (clients, locations)
8. Add database indexes
9. Precompute aggregations

---

## 12. MONITORING & METRICS

### Recommended Metrics

1. **Query Count by Type** (per hour/day)
2. **Cache Hit Rate** (per cache key)
3. **Average Query Latency** (p50, p95, p99)
4. **Slow Query Count** (>100ms, >500ms)
5. **Database Connection Pool Usage**

### Tools

- **PostgreSQL**: `pg_stat_statements` extension
- **Redis**: `INFO stats` for cache metrics
- **Application**: Structured logging with query timing

---

## Conclusion

The codebase has **strong tenant isolation** and **good query patterns**, but **critical optimizations** are needed:

1. ✅ **Fix N+1 pattern** in GetSeatUsage (99% reduction)
2. ✅ **Cache high-frequency queries** (85% overall reduction)
3. ✅ **Implement slow query logging** (identify bottlenecks)

**Estimated Impact**: 
- **85% reduction** in database queries
- **$15-20/month savings** (avoid upgrade to larger instance)
- **Improved performance** (faster response times)
- **Better scalability** (headroom for growth)

**Next Steps**: Implement Priority 1 optimizations immediately.
