# Redis Caching Strategy Analysis & Optimization Plan

**Date**: 2025-01-27  
**Target**: 80%+ cache hit rate to reduce infrastructure costs from $2,000+/mo to $1,200/mo  
**Current Status**: ⚠️ **Minimal caching implemented** - Only tenant membership cached

---

## Executive Summary

**Current State**:
- ✅ **1 cache implementation**: User tenant IDs (5min TTL)
- ❌ **5 critical cache opportunities**: Not implemented
- ❌ **No cache monitoring**: No hit rate metrics
- ❌ **No cache invalidation**: Manual invalidation not called on data changes
- ❌ **No external API caching**: Data4SEO/Meta APIs not yet implemented

**Estimated Current Cache Hit Rate**: **~15-20%** (only tenant membership cached, but 4 other high-frequency queries uncached)

**Target**: **80%+ cache hit rate** after implementing all critical caches

**Cost Impact**: 
- **Current**: ~200,000 queries/day → **$2,000+/mo** (db.t3.medium required)
- **After optimization**: ~30,000 queries/day → **$1,200/mo** (db.t3.small sufficient)
- **Savings**: **$800+/month** (40% reduction)

---

## 1. CURRENT CACHE USAGE

### 1.1 What Data is Currently Cached?

#### ✅ **User Tenant IDs** (IMPLEMENTED)
- **Location**: `internal/platform/tenant/cache.go`
- **Cache Key Pattern**: `tenant_cache:{user_id}`
- **TTL**: 5 minutes (fixed, configurable via constructor)
- **Data Type**: JSON array of tenant ID strings
- **Frequency**: Every authenticated request
- **Estimated Hit Rate**: 85-90% (assuming 5min TTL with typical session duration)

**Code Reference**:
```34:36:internal/platform/tenant/cache.go
// key generates a cache key for a user ID
func (tc *TenantCache) key(userID uuid.UUID) string {
	return tc.prefix + userID.String()
}
```

**Cache Pattern**: Cache-aside (check cache, if miss fetch DB, populate cache)

#### ❌ **Tenant Resolution by Domain** (NOT CACHED)
- **Query**: `SELECT agency_id FROM branding WHERE domain = $1`
- **Frequency**: Every request
- **Estimated calls/day**: ~50,000
- **Cache opportunity**: 95%+ hit rate (domains rarely change)

#### ❌ **User Lookup by Clerk ID** (NOT CACHED)
- **Query**: `SELECT * FROM users WHERE clerk_user_id = $1`
- **Frequency**: Every authenticated request
- **Estimated calls/day**: ~50,000
- **Cache opportunity**: 90%+ hit rate (user data changes infrequently)

#### ❌ **Brand Theme by Domain** (NOT CACHED - Backend)
- **Query**: `SELECT * FROM branding WHERE domain = $1`
- **Frequency**: Every page load
- **Estimated calls/day**: ~25,000
- **Frontend caching**: ✅ localStorage (15min TTL) - `farohq-portal/src/hooks/useBrandTheme.ts`
- **Backend caching**: ❌ None
- **Cache opportunity**: 95%+ hit rate (theme changes rarely)

#### ❌ **Client Validation** (NOT CACHED)
- **Query**: `SELECT id FROM clients WHERE id = $1 AND agency_id = $2`
- **Frequency**: Every request with client_id
- **Estimated calls/day**: ~25,000
- **Cache opportunity**: 85%+ hit rate

#### ❌ **List Clients by Agency** (NOT CACHED)
- **Query**: `SELECT * FROM clients WHERE agency_id = $1`
- **Frequency**: Dashboard loads
- **Estimated calls/day**: ~5,000
- **Cache opportunity**: 80%+ hit rate

### 1.2 Cache Key Structure

**Current Pattern**: `{prefix}:{identifier}`

**Examples**:
- `tenant_cache:{user_id}` - User's accessible tenant IDs

**Issues**:
1. ❌ **No tenant isolation in key structure** - Keys don't include tenant_id, but current data is user-scoped (safe)
2. ❌ **No versioning** - Cache keys don't include version numbers for invalidation
3. ⚠️ **Simple prefix** - Only `tenant_cache:` prefix, no namespace hierarchy

**Recommended Pattern**: `{namespace}:{resource}:{tenant_id}:{identifier}`

**Examples**:
- `cache:tenant_members:{user_id}` - User's tenant memberships (user-scoped, no tenant needed)
- `cache:domain_tenant:{domain}` - Domain → tenant mapping (domain-scoped)
- `cache:user:{clerk_user_id}` - User data (user-scoped)
- `cache:brand_theme:{domain}` - Brand theme (domain-scoped)
- `cache:client:{tenant_id}:{client_id}` - Client validation (tenant-scoped)

### 1.3 Cache TTLs

**Current TTLs**:
| Data Type | TTL | Configurable? | Rationale |
|-----------|-----|---------------|-----------|
| User Tenant IDs | 5 minutes | ✅ Yes (via constructor) | User membership changes infrequently |

**Recommended TTLs** (from QUERY_ANALYSIS.md):
| Data Type | Recommended TTL | Rationale |
|-----------|----------------|-----------|
| User Tenant IDs | 5 minutes | ✅ Current (good) |
| Tenant Resolution by Domain | 15 minutes | Domains rarely change |
| User Lookup by Clerk ID | 5 minutes | User data changes infrequently |
| Brand Theme by Domain | 15 minutes | Theme changes rarely |
| Client Validation | 10 minutes | Client existence rarely changes |
| List Clients by Agency | 5 minutes | Client list changes infrequently |

**TTL Configuration**:
- ❌ **Not configurable per tenant** - All tenants use same TTL
- ❌ **No TTL override mechanism** - Cannot adjust TTL for specific use cases
- ⚠️ **Fixed TTL** - No adaptive TTL based on data change frequency

**Recommendation**: Make TTLs configurable via environment variables or config file.

---

## 2. CACHE PATTERNS

### 2.1 Cache-Aside Pattern (Current)

**Status**: ✅ **Used** for tenant membership cache

**Implementation**:
```145:181:internal/platform/tenant/cache.go
// GetWithResolver gets tenant IDs from cache or resolves from database
// This method is used by the cache itself to avoid circular dependency
func (tc *TenantCache) GetWithResolver(
	ctx context.Context,
	userID uuid.UUID,
	resolver *Resolver,
) ([]string, error) {
	// Try cache first
	if tenantIDs, found := tc.Get(ctx, userID); found {
		tc.logger.Debug().
			Str("user_id", userID.String()).
			Int("tenant_count", len(tenantIDs)).
			Msg("Tenant cache hit")
		return tenantIDs, nil
	}

	// Cache miss - resolve from database (use direct DB query to avoid recursion)
	tc.logger.Debug().
		Str("user_id", userID.String()).
		Msg("Tenant cache miss - resolving from database")

	tenantIDs, err := resolver.getUserTenantIDsFromDB(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := tc.Set(ctx, userID, tenantIDs); err != nil {
		// Log error but don't fail the request
		tc.logger.Warn().
			Str("user_id", userID.String()).
			Err(err).
			Msg("Failed to cache tenant IDs, continuing without cache")
	}

	return tenantIDs, nil
}
```

**Pattern Flow**:
1. ✅ Check cache first
2. ✅ If miss, fetch from database
3. ✅ Populate cache (non-blocking - errors logged but don't fail request)
4. ✅ Return data

**Strengths**:
- ✅ Graceful degradation (continues without cache if Redis fails)
- ✅ Non-blocking cache writes (errors don't fail requests)

**Weaknesses**:
- ❌ No write-through pattern for updates
- ❌ Cache invalidation not called on data changes (see §5)

### 2.2 Write-Through Pattern

**Status**: ❌ **NOT USED**

**Issue**: When tenant membership changes (user added/removed from tenant), cache is not invalidated.

**Current Behavior**:
- User added to tenant → Cache still shows old tenant list until TTL expires (up to 5 minutes stale)
- User removed from tenant → Cache still shows old tenant list (security risk!)

**Recommendation**: Implement write-through or event-based invalidation (see §5).

### 2.3 Cache Invalidation Issues

**Status**: ⚠️ **CRITICAL** - Cache invalidation methods exist but are **never called**

**Available Methods**:
```92:143:internal/platform/tenant/cache.go
// Invalidate removes cached tenant IDs for a user
func (tc *TenantCache) Invalidate(ctx context.Context, userID uuid.UUID) error {
	key := tc.key(userID)

	err := tc.client.Del(ctx, key).Err()
	if err != nil {
		tc.logger.Error().
			Str("user_id", userID.String()).
			Err(err).
			Msg("Failed to invalidate tenant cache in Redis")
		return err
	}

	tc.logger.Debug().
		Str("user_id", userID.String()).
		Msg("Invalidated tenant cache for user")

	return nil
}

// InvalidateAll clears all cached entries with the tenant cache prefix
func (tc *TenantCache) InvalidateAll(ctx context.Context) error {
	pattern := tc.prefix + "*"

	iter := tc.client.Scan(ctx, 0, pattern, 0).Iterator()
	keys := []string{}
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		tc.logger.Error().
			Err(err).
			Msg("Failed to scan tenant cache keys in Redis")
		return err
	}

	if len(keys) > 0 {
		err := tc.client.Del(ctx, keys...).Err()
		if err != nil {
			tc.logger.Error().
				Err(err).
				Msg("Failed to delete tenant cache keys in Redis")
			return err
		}
	}

	tc.logger.Debug().
		Int("deleted_keys", len(keys)).
		Msg("Invalidated all tenant cache entries")

	return nil
}
```

**Problem**: These methods are **never called** when:
- User is added to tenant (`tenant_members` INSERT)
- User is removed from tenant (`tenant_members` DELETE/UPDATE)
- User's role changes (`tenant_members` UPDATE)

**Impact**: 
- **Stale data**: Users may see outdated tenant lists for up to 5 minutes
- **Security risk**: Removed users may still access tenants until cache expires
- **Poor UX**: New tenant memberships not visible immediately

**Recommendation**: Call `Invalidate(userID)` in:
- `InviteMember` use case (after invite accepted)
- `RemoveMember` use case (if exists)
- `UpdateMemberRole` use case (if exists)

### 2.4 Cache Miss Handling

**Status**: ✅ **Good** - Graceful degradation implemented

**Behavior**:
- Cache miss → Fetch from database → Populate cache (non-blocking)
- Cache error → Log warning, continue without cache
- Redis unavailable → Continue without cache (no request failures)

**Code**:
```172:178:internal/platform/tenant/cache.go
	// Store in cache
	if err := tc.Set(ctx, userID, tenantIDs); err != nil {
		// Log error but don't fail the request
		tc.logger.Warn().
			Str("user_id", userID.String()).
			Err(err).
			Msg("Failed to cache tenant IDs, continuing without cache")
	}
```

**Recommendation**: ✅ Keep this pattern for all new cache implementations.

---

## 3. CACHE HIT RATE

### 3.1 Current Cache Hit Rate

**Estimated Current Hit Rate**: **~15-20%**

**Breakdown**:
- User Tenant IDs: **85-90%** hit rate (cached, 5min TTL)
  - Estimated: 50,000 queries/day → 42,500 hits, 7,500 misses
- Tenant Resolution by Domain: **0%** (not cached)
  - Estimated: 50,000 queries/day → 0 hits, 50,000 misses
- User Lookup by Clerk ID: **0%** (not cached)
  - Estimated: 50,000 queries/day → 0 hits, 50,000 misses
- Brand Theme: **0%** backend (frontend cached)
  - Estimated: 25,000 queries/day → 0 hits, 25,000 misses
- Client Validation: **0%** (not cached)
  - Estimated: 25,000 queries/day → 0 hits, 25,000 misses

**Total**:
- **Cached queries**: 50,000/day (user tenant IDs)
- **Uncached queries**: 200,000/day (all others)
- **Total queries**: 250,000/day
- **Cache hits**: ~42,500/day (17%)
- **Cache misses**: ~207,500/day (83%)

**Target**: 80%+ hit rate = **200,000+ hits/day**

### 3.2 Operations with Low Hit Rates

**Critical (Should be 90%+)**:
1. ❌ **Tenant Resolution by Domain** - 0% (should be 95%+)
2. ❌ **User Lookup by Clerk ID** - 0% (should be 90%+)
3. ❌ **Brand Theme by Domain** - 0% backend (should be 95%+)

**Medium Priority (Should be 80%+)**:
4. ❌ **Client Validation** - 0% (should be 85%+)
5. ❌ **List Clients by Agency** - 0% (should be 80%+)

### 3.3 Monitoring/Observability

**Status**: ❌ **NO MONITORING IMPLEMENTED**

**Missing Metrics**:
- ❌ Cache hit rate per key pattern
- ❌ Cache miss rate per key pattern
- ❌ Cache latency (p50, p95, p99)
- ❌ Redis memory usage
- ❌ Redis eviction rate
- ❌ Slow cache operations (>10ms)

**Current Logging**:
- ✅ Debug logs for cache hits/misses (tenant cache only)
- ❌ No structured metrics for dashboards
- ❌ No alerts for cache performance issues

**Recommendation**: Implement cache metrics using:
- Redis `INFO stats` command for hit/miss rates
- Application-level metrics (Prometheus/CloudWatch)
- Structured logging with cache operation timing

### 3.4 Hot Spots (Frequently Accessed, Not Cached)

**Identified Hot Spots**:
1. 🔴 **Tenant Resolution by Domain** - Every request, not cached
2. 🔴 **User Lookup by Clerk ID** - Every authenticated request, not cached
3. 🟡 **Brand Theme by Domain** - Every page load, backend not cached
4. 🟡 **Client Validation** - Every request with client_id, not cached

**Impact**: These 4 operations account for **~150,000 queries/day** that could be cached.

---

## 4. TENANT ISOLATION IN CACHE

### 4.1 Cache Key Namespacing

**Current Pattern**: `tenant_cache:{user_id}`

**Analysis**:
- ✅ **Safe for user-scoped data** - User IDs are unique, no tenant leakage risk
- ❌ **No tenant isolation for tenant-scoped data** - Future caches need tenant_id in key

**Example Risk** (if caching client data):
```go
// ❌ BAD - No tenant isolation
cacheKey := fmt.Sprintf("client:%s", clientID)

// ✅ GOOD - Tenant isolation
cacheKey := fmt.Sprintf("client:%s:%s", tenantID, clientID)
```

**Recommendation**: 
- Use tenant_id in cache keys for tenant-scoped data
- Validate tenant_id in cache keys to prevent injection

### 4.2 Tenant Data Leakage Risk

**Current Risk**: ⚠️ **LOW** - Only user-scoped data cached (tenant membership)

**Potential Risks** (if caching tenant-scoped data without tenant_id):
1. ❌ **Cache key collision** - Same client_id in different tenants could collide
2. ❌ **Cross-tenant access** - Tenant A could see Tenant B's cached data
3. ❌ **Cache injection** - Malicious tenant_id in cache key could access other tenants' data

**Mitigation**:
- Always include tenant_id in cache keys for tenant-scoped data
- Validate tenant_id format (UUID) before using in cache key
- Use Redis key patterns with tenant_id: `cache:{resource}:{tenant_id}:{id}`

### 4.3 Cache Key Validation

**Status**: ⚠️ **PARTIAL** - UUID validation exists but not enforced

**Current**: Cache keys use UUID strings directly (from `userID.String()`)

**Risk**: If user_id is not validated before cache key generation, could allow:
- SQL injection (if user_id comes from untrusted source)
- Cache key injection (if user_id contains special characters)

**Recommendation**: 
- Validate UUID format before cache key generation
- Sanitize cache keys (no special characters)
- Use structured key builders (not string concatenation)

### 4.4 Per-Tenant Cache Size Limits

**Status**: ❌ **NOT IMPLEMENTED**

**Risk**: "Noisy neighbor" problem - One tenant could fill Redis with cached data, evicting other tenants' cache.

**Mitigation Needed**:
- Per-tenant cache size limits (e.g., 10MB per tenant)
- Cache eviction policy: LRU with tenant-aware eviction
- Monitor cache size per tenant (Redis key patterns)

**Recommendation**: 
- Use Redis key patterns to track per-tenant cache size: `cache:{resource}:{tenant_id}:*`
- Set maxmemory-policy: `allkeys-lru` (evict least recently used)
- Monitor eviction rate (if high, increase Redis memory or reduce TTLs)

---

## 5. CACHE INVALIDATION STRATEGY

### 5.1 Current Invalidation Methods

**Available Methods**:
1. ✅ `Invalidate(ctx, userID)` - Invalidate specific user's cache
2. ✅ `InvalidateAll(ctx)` - Invalidate all tenant cache entries

**Problem**: ⚠️ **These methods are never called** when data changes.

### 5.2 Invalidation on Data Changes

**Status**: ❌ **NOT IMPLEMENTED**

**Missing Invalidation Points**:

#### User Added to Tenant
**Location**: `internal/domains/tenants/app/usecases/invite_member.go` (after invite accepted)

**Current**: No cache invalidation
**Needed**: `tenantCache.Invalidate(ctx, userID)` after `tenant_members` INSERT

#### User Removed from Tenant
**Location**: `internal/domains/tenants/infra/db/tenant_member_repository.go:207-221` (Delete method)

**Current**: No cache invalidation
**Needed**: `tenantCache.Invalidate(ctx, userID)` after `tenant_members` DELETE/UPDATE

#### User Role Changed
**Location**: `internal/domains/tenants/infra/db/tenant_member_repository.go:177-205` (Update method)

**Current**: No cache invalidation
**Needed**: `tenantCache.Invalidate(ctx, userID)` after `tenant_members` UPDATE (if role changed)

#### Domain → Tenant Mapping Changed
**Location**: `internal/domains/brand/infra/db/brand_repository.go` (when domain updated)

**Current**: Not cached (no invalidation needed yet)
**Needed**: When implementing domain cache, invalidate on domain update

#### Brand Theme Changed
**Location**: `internal/domains/brand/infra/db/brand_repository.go` (when theme updated)

**Current**: Not cached (backend)
**Needed**: When implementing brand cache, invalidate on theme update

### 5.3 Event-Based Invalidation

**Status**: ❌ **NOT IMPLEMENTED**

**Current**: Manual invalidation (but not called)

**Recommendation**: Implement event-based invalidation using:
- **Database triggers** → Publish events → Invalidate cache
- **Application events** → Invalidate cache on domain events
- **Redis Pub/Sub** → Real-time invalidation across instances

**Example** (Event-Based):
```go
// When user added to tenant
func (uc *InviteMember) Execute(ctx context.Context, ...) error {
    // ... add user to tenant ...
    
    // Invalidate cache
    if uc.tenantCache != nil {
        uc.tenantCache.Invalidate(ctx, userID)
    }
    
    // Publish event (optional, for multi-instance invalidation)
    uc.eventBus.Publish(ctx, "tenant.member.added", userID)
    
    return nil
}
```

### 5.4 Redis Pub/Sub for Real-Time Invalidation

**Status**: ❌ **NOT IMPLEMENTED**

**Use Case**: Multi-instance deployments (multiple app servers sharing Redis)

**Problem**: If Instance A updates data, Instance B's cache is stale until TTL expires.

**Solution**: Use Redis Pub/Sub to broadcast invalidation events:
```go
// Instance A: Invalidate and publish
tenantCache.Invalidate(ctx, userID)
redisClient.Publish(ctx, "cache:invalidate", "tenant_cache:"+userID.String())

// Instance B: Subscribe and invalidate
pubsub := redisClient.Subscribe(ctx, "cache:invalidate")
for msg := range pubsub.Channel() {
    redisClient.Del(ctx, msg.Payload)
}
```

**Recommendation**: Implement when scaling to multiple instances.

---

## 6. EXTERNAL API RESPONSES

### 6.1 Data4SEO API Caching

**Status**: ❌ **NOT IMPLEMENTED** (API integration not yet implemented)

**Expected Endpoints** (from `docs/API_COST_ANALYSIS.md`):
- GBP profile snapshots: `$0.50-1.50 per snapshot`
- Local rankings data: `$0.50-1.50 per snapshot`
- Competitive intelligence: `$0.50-1.50 per snapshot`

**Recommended Cache TTLs**:
| Endpoint | Cache TTL | Rationale |
|----------|-----------|-----------|
| GBP Profiles | 15 minutes | Profiles change infrequently |
| Rankings | 1 hour | Rankings update daily |
| Snapshots | 24 hours | Historical data, expensive |

**Cache Key Pattern**: `data4seo:{endpoint}:{location_id}`

**Cost Impact**:
- Without caching: `10 locations × 5 loads/day × $1.00 = $50/day = $1,500/month`
- With caching: `10 locations × 4 syncs/day × $1.00 = $40/day = $1,200/month`
- **Savings**: **$300/month (20% reduction)**

### 6.2 Meta Cloud API Caching

**Status**: ❌ **NOT IMPLEMENTED** (API integration not yet implemented)

**Expected Endpoints**:
- WhatsApp Business API: `$0.0015-0.045 per conversation`
- Instagram Messaging API: `$0.0015-0.045 per conversation`
- Facebook Messenger API: `$0.0015-0.045 per conversation`

**Recommended Cache TTLs**:
| Endpoint | Cache TTL | Rationale |
|----------|-----------|-----------|
| Message threads | 5 minutes | Active conversations |
| User profiles | 15 minutes | Profiles change infrequently |
| Webhook validation | No caching | Security-critical |

**Cache Key Pattern**: `meta:{platform}:{resource}:{id}`

### 6.3 Current API Response Caching

**Status**: ❌ **No external API integrations found**

**When Implemented**: Use Redis for API response caching with appropriate TTLs (see `docs/API_COST_ANALYSIS.md` for detailed recommendations).

---

## 7. SESSION CACHING

### 7.1 Session Storage

**Status**: ✅ **Clerk-managed sessions** (not in Redis)

**Implementation**: Clerk handles session management:
- Sessions stored in Clerk's infrastructure (not our Redis)
- JWT tokens issued by Clerk
- Tokens validated via JWKS (cached in memory, not Redis)

**Code Reference**: `internal/platform/httpserver/auth.go` - Validates Clerk JWT tokens

**Session TTL**: Clerk default 7 days (configurable in Clerk Dashboard)

### 7.2 Session Invalidation

**Status**: ✅ **Clerk-managed** - Clerk invalidates sessions on logout/expiry

**Our Role**: None - Clerk handles all session lifecycle

### 7.3 Device Sessions

**Status**: ✅ **Clerk-managed** - Clerk tracks device sessions

**Our Role**: None - Clerk handles device tracking

### 7.4 JWKS Caching

**Status**: ⚠️ **In-memory cache** (not Redis)

**Location**: `lvos/services/core-app/internal/middleware/auth.go:68-85` (different codebase)

**Current**: JWKS fetched from Clerk and cached in memory

**Recommendation**: Consider caching JWKS in Redis (TTL: 1 hour) for multi-instance deployments.

---

## 8. MONITORING & ALERTS

### 8.1 Cache Performance Dashboard

**Status**: ❌ **NOT IMPLEMENTED**

**Missing Dashboards**:
- Cache hit rate per key pattern
- Cache miss rate per key pattern
- Cache latency (p50, p95, p99)
- Redis memory usage
- Redis eviction rate
- Cache operations per second

**Recommendation**: Implement using:
- **Redis `INFO stats`** for hit/miss rates
- **Prometheus/CloudWatch** for application metrics
- **Grafana** for visualization

### 8.2 Cache Memory Usage Alerts

**Status**: ❌ **NOT IMPLEMENTED**

**Risk**: Redis memory full → Eviction → Cache misses → Database load spike

**Recommendation**: 
- Alert when Redis memory usage > 80%
- Alert when eviction rate > 100 keys/second
- Monitor `used_memory` and `used_memory_peak` via Redis `INFO memory`

### 8.3 Cache Hit Rate Monitoring

**Status**: ❌ **NOT IMPLEMENTED**

**Current**: Only debug logs (not metrics)

**Recommendation**: 
- Track hit/miss counts per key pattern
- Calculate hit rate: `hits / (hits + misses)`
- Alert when hit rate < 70% (target: 80%+)

### 8.4 Slow Query Logging for Cache Misses

**Status**: ❌ **NOT IMPLEMENTED**

**Recommendation**: 
- Log cache misses that trigger database queries
- Track database query latency after cache miss
- Alert when cache miss → DB query > 100ms

---

## 9. REDIS RESILIENCE

### 9.1 Redis Replication/Backup

**Status**: ⚠️ **UNKNOWN** - Docker Compose uses single Dragonfly instance

**Current Setup** (`docker-compose.yml`):
```31:41:docker-compose.yml
  dragonfly:
    image: docker.dragonflydb.io/dragonflydb/dragonfly:latest
    container_name: farohq-core-app-dragonfly
    ports:
      - "6379:6379"
    command: ["--logtostderr"]
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5
```

**Issues**:
- ❌ Single instance (no replication)
- ❌ No backup configured
- ❌ No persistence configured (data lost on restart)

**Recommendation**:
- **Development**: Current setup OK (single instance)
- **Production**: Use managed Redis (AWS ElastiCache, GCP Memorystore) with:
  - Replication (read replicas)
  - Automated backups
  - Persistence (AOF or RDB)

### 9.2 Redis Failure Handling

**Status**: ✅ **Graceful degradation implemented**

**Behavior**: If Redis unavailable, application continues without cache:
```49:55:cmd/server/main.go
		if err != nil {
			logger.Warn().
				Err(err).
				Str("redis_url", cfg.RedisURL).
				Msg("Failed to connect to Redis/Dragonfly, continuing without cache")
		} else {
			redisClient = client
			// Create tenant cache with 5 minute TTL
			tenantCache = tenant.NewTenantCache(client, 5*time.Minute, logger)
```

**Cache Operations**: Non-blocking - errors logged but don't fail requests:
```172:178:internal/platform/tenant/cache.go
	// Store in cache
	if err := tc.Set(ctx, userID, tenantIDs); err != nil {
		// Log error but don't fail the request
		tc.logger.Warn().
			Str("user_id", userID.String()).
			Err(err).
			Msg("Failed to cache tenant IDs, continuing without cache")
	}
```

**Recommendation**: ✅ Keep this pattern - graceful degradation is critical.

### 9.3 Fallback to Database

**Status**: ✅ **Implemented** - Cache miss → Database query

**Behavior**: If cache miss or Redis unavailable, fetch from database.

**Recommendation**: ✅ Keep this pattern.

### 9.4 Recovery Time

**Status**: ⚠️ **UNKNOWN** - No monitoring of Redis recovery

**Recommendation**: 
- Monitor Redis connection health
- Track time to reconnect after failure
- Alert if Redis unavailable > 1 minute

---

## 10. OPTIMIZATION OPPORTUNITIES

### 10.1 Critical (Implement Immediately)

#### 1. Cache Tenant Resolution by Domain 🔴
- **Impact**: 95% reduction (50,000 → 2,500 queries/day)
- **Effort**: Low (2-4 hours)
- **Cost Savings**: ~$400/month (prevents DB upgrade)
- **Implementation**: Add Redis cache in `internal/platform/tenant/resolver.go:257-280`

#### 2. Cache User Lookup by Clerk ID 🔴
- **Impact**: 90% reduction (50,000 → 5,000 queries/day)
- **Effort**: Low (2-4 hours)
- **Cost Savings**: ~$400/month
- **Implementation**: Add Redis cache in `internal/domains/users/infra/db/user_repository.go:30-35`

#### 3. Implement Cache Invalidation on Data Changes 🔴
- **Impact**: Prevents stale data, security risks
- **Effort**: Medium (4-8 hours)
- **Cost Savings**: Prevents security incidents, improves UX
- **Implementation**: Call `Invalidate()` in:
  - `InviteMember` use case
  - `TenantMemberRepository.Delete()`
  - `TenantMemberRepository.Update()` (if role changed)

### 10.2 High Priority (Implement Soon)

#### 4. Cache Brand Theme by Domain 🟡
- **Impact**: 95% reduction (25,000 → 1,250 queries/day)
- **Effort**: Low (2-4 hours)
- **Cost Savings**: ~$200/month
- **Implementation**: Add Redis cache in `internal/domains/brand/infra/db/brand_repository.go:125-132`

#### 5. Cache Client Validation 🟡
- **Impact**: 85% reduction (25,000 → 3,750 queries/day)
- **Effort**: Low (2-4 hours)
- **Cost Savings**: ~$150/month
- **Implementation**: Add Redis cache in `internal/platform/httpserver/tenant.go:90-94`

#### 6. Implement Cache Monitoring & Alerts 🟡
- **Impact**: Visibility into cache performance, early detection of issues
- **Effort**: Medium (8-16 hours)
- **Cost Savings**: Prevents cost overruns, optimizes TTLs
- **Implementation**: 
  - Redis `INFO stats` metrics
  - Prometheus/CloudWatch integration
  - Grafana dashboards

### 10.3 Medium Priority (Implement When Needed)

#### 7. Cache List Clients by Agency 🟢
- **Impact**: 80% reduction (5,000 → 1,000 queries/day)
- **Effort**: Low (2-4 hours)
- **Cost Savings**: ~$50/month
- **Implementation**: Add Redis cache in `internal/domains/tenants/infra/db/client_repository.go:148-154`

#### 8. Implement Redis Pub/Sub for Multi-Instance Invalidation 🟢
- **Impact**: Real-time cache invalidation across instances
- **Effort**: Medium (4-8 hours)
- **Cost Savings**: Prevents stale data in multi-instance deployments
- **Implementation**: Redis Pub/Sub for cache invalidation events

#### 9. Per-Tenant Cache Size Limits 🟢
- **Impact**: Prevents "noisy neighbor" problem
- **Effort**: Medium (4-8 hours)
- **Cost Savings**: Prevents cache eviction, improves hit rates
- **Implementation**: Monitor cache size per tenant, set limits

#### 10. Configurable TTLs per Cache Type 🟢
- **Impact**: Optimize TTLs based on data change frequency
- **Effort**: Low (2-4 hours)
- **Cost Savings**: Improves hit rates, reduces stale data
- **Implementation**: Environment variables or config file for TTLs

### 10.4 Estimated Total Cost Impact

**Current State**:
- Queries/day: ~250,000
- Cache hit rate: ~17%
- Database required: db.t3.medium ($2,000+/mo)

**After Critical + High Priority Optimizations**:
- Queries/day: ~30,000 (88% reduction)
- Cache hit rate: ~80%+
- Database required: db.t3.small ($1,200/mo)

**Savings**: **$800+/month (40% reduction)**

---

## 11. RECOMMENDED REDIS ARCHITECTURE

### 11.1 Development (Current)

**Setup**: Single Dragonfly instance via Docker Compose

**Status**: ✅ **Adequate** for development

**Recommendation**: Keep current setup.

### 11.2 Production (Recommended)

**Option 1: Managed Redis (Recommended)**
- **AWS**: ElastiCache (Redis 7.x)
  - Instance: `cache.t3.medium` (2 vCPU, 3.09 GB RAM) - $50-70/month
  - Replication: 1 primary + 1 replica
  - Backup: Daily automated backups
  - High Availability: Multi-AZ deployment
- **GCP**: Memorystore (Redis 7.x)
  - Instance: `standard-1` (1 vCPU, 4 GB RAM) - $60-80/month
  - Replication: 1 primary + 1 replica
  - Backup: Daily automated backups
  - High Availability: Regional deployment

**Option 2: Self-Managed Redis Cluster**
- **Setup**: Redis Cluster (3+ nodes)
- **Cost**: Lower ($30-50/month for compute)
- **Effort**: High (maintenance, monitoring, backups)
- **Recommendation**: Only if cost is critical and team has Redis expertise

### 11.3 Redis Configuration Recommendations

**Memory**: 4-8 GB (sufficient for 80%+ hit rate with 30 agencies)

**Eviction Policy**: `allkeys-lru` (evict least recently used when memory full)

**Persistence**: 
- **AOF** (Append-Only File) - Recommended for durability
- **RDB** (snapshots) - Optional for backups

**Connection Pooling**: 
- Max connections: 100-200 (adjust based on app instances)
- Timeout: 5 seconds

**Monitoring**:
- Enable Redis `INFO` commands
- Monitor: memory usage, hit rate, eviction rate, latency

---

## 12. IMPLEMENTATION ROADMAP

### Phase 1: Critical Caches (Week 1)
1. ✅ Cache tenant resolution by domain (15min TTL)
2. ✅ Cache user lookup by Clerk ID (5min TTL)
3. ✅ Implement cache invalidation on data changes

**Expected Impact**: 85%+ cache hit rate, ~$800/month savings

### Phase 2: High Priority Caches (Week 2)
4. ✅ Cache brand theme by domain (15min TTL)
5. ✅ Cache client validation (10min TTL)
6. ✅ Implement cache monitoring & alerts

**Expected Impact**: 80%+ cache hit rate, improved observability

### Phase 3: Medium Priority (Week 3-4)
7. ✅ Cache list clients by agency (5min TTL)
8. ✅ Configurable TTLs per cache type
9. ✅ Per-tenant cache size limits

**Expected Impact**: Optimized cache performance, prevents noisy neighbor

### Phase 4: Advanced Features (Month 2)
10. ✅ Redis Pub/Sub for multi-instance invalidation
11. ✅ External API response caching (when APIs implemented)
12. ✅ Cache warming strategies

**Expected Impact**: Real-time invalidation, API cost savings

---

## 13. CONCLUSION

**Current State**: Minimal caching (only tenant membership), ~17% hit rate

**Target State**: Comprehensive caching, 80%+ hit rate, $800+/month savings

**Key Findings**:
1. ✅ **1 cache implemented** (tenant membership) - Good pattern, needs invalidation
2. ❌ **5 critical caches missing** - High impact, low effort
3. ❌ **No cache invalidation** - Security risk, stale data
4. ❌ **No monitoring** - No visibility into cache performance
5. ❌ **No external API caching** - Will be critical when APIs go live

**Next Steps**:
1. Implement Phase 1 (critical caches + invalidation) - **Week 1**
2. Implement Phase 2 (high priority + monitoring) - **Week 2**
3. Monitor cache hit rates and adjust TTLs - **Ongoing**
4. Implement Phase 3-4 as needed - **Month 2+**

**Expected Outcome**: 80%+ cache hit rate, $800+/month infrastructure savings, improved performance, better scalability.
