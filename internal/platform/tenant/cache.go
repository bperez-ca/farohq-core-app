package tenant

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// TenantCache caches user's accessible tenants using Dragonfly/Redis
type TenantCache struct {
	client *redis.Client
	ttl    time.Duration
	logger zerolog.Logger
	prefix string
}

// NewTenantCache creates a new tenant cache using Redis/Dragonfly
func NewTenantCache(client *redis.Client, ttl time.Duration, logger zerolog.Logger) *TenantCache {
	if ttl == 0 {
		ttl = 5 * time.Minute // Default TTL
	}
	return &TenantCache{
		client: client,
		ttl:    ttl,
		logger: logger,
		prefix: "tenant_cache:",
	}
}

// key generates a cache key for a user ID
func (tc *TenantCache) key(userID uuid.UUID) string {
	return tc.prefix + userID.String()
}

// Get retrieves cached tenant IDs for a user
func (tc *TenantCache) Get(ctx context.Context, userID uuid.UUID) ([]string, bool) {
	key := tc.key(userID)

	val, err := tc.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, false
		}
		tc.logger.Debug().
			Str("user_id", userID.String()).
			Err(err).
			Msg("Failed to get tenant cache from Redis")
		return nil, false
	}

	var tenantIDs []string
	if err := json.Unmarshal([]byte(val), &tenantIDs); err != nil {
		tc.logger.Warn().
			Str("user_id", userID.String()).
			Err(err).
			Msg("Failed to unmarshal tenant cache from Redis")
		return nil, false
	}

	return tenantIDs, true
}

// Set stores tenant IDs for a user with TTL
func (tc *TenantCache) Set(ctx context.Context, userID uuid.UUID, tenantIDs []string) error {
	key := tc.key(userID)

	data, err := json.Marshal(tenantIDs)
	if err != nil {
		tc.logger.Error().
			Str("user_id", userID.String()).
			Err(err).
			Msg("Failed to marshal tenant IDs for cache")
		return err
	}

	err = tc.client.Set(ctx, key, data, tc.ttl).Err()
	if err != nil {
		tc.logger.Error().
			Str("user_id", userID.String()).
			Err(err).
			Msg("Failed to set tenant cache in Redis")
		return err
	}

	return nil
}

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

// Ping checks if the Redis connection is alive
func (tc *TenantCache) Ping(ctx context.Context) error {
	return tc.client.Ping(ctx).Err()
}
