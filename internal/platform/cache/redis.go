package cache

import (
	"context"
	"fmt"
	"net/url"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// NewRedisClient creates a new Redis/Dragonfly client from a URL
// Supports both redis:// and rediss:// (TLS) URLs
func NewRedisClient(redisURL string, logger zerolog.Logger) (*redis.Client, error) {
	if redisURL == "" {
		return nil, fmt.Errorf("redis URL is required")
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		// If ParseURL fails, try manual parsing for custom formats
		parsedURL, parseErr := url.Parse(redisURL)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
		}

		opt = &redis.Options{
			Addr:     parsedURL.Host,
			Password: "",
			DB:       0,
		}

		if parsedURL.User != nil {
			opt.Password, _ = parsedURL.User.Password()
		}

		if parsedURL.Path != "" {
			// Extract DB number from path if present (e.g., /0)
			var dbNum int
			if _, scanErr := fmt.Sscanf(parsedURL.Path, "/%d", &dbNum); scanErr == nil {
				opt.DB = dbNum
			}
		}
	}

	client := redis.NewClient(opt)

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info().
		Str("addr", opt.Addr).
		Int("db", opt.DB).
		Msg("Connected to Redis/Dragonfly")

	return client, nil
}
