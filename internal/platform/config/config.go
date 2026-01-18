package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// Config holds application configuration
type Config struct {
	// Database configuration (separate vars for security)
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string

	// Clerk
	ClerkJWKSURL string

	// Web
	WebURL string

	// AWS S3
	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	S3BucketName       string

	// GCP Cloud Storage
	GCSBucketName string
	GCSProjectID  string

	// Vercel API (for custom domain management)
	VercelAPIToken  string
	VercelProjectID string
	VercelTeamID    string

	// DNS (optional, for UX feedback only)
	DNSLookupEnabled bool

	// Email Service Configuration
	PostmarkAPIToken  string
	PostmarkFromEmail string
	MailhogHost       string
	MailhogPort       string
	Environment       string

	// Server
	Port string

	// Redis/Dragonfly Cache
	RedisURL string
}

// NewConfig creates a new configuration from environment variables
func NewConfig() *Config {
	cfg := &Config{
		// Database - use separate env vars for better security
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "localvisibilityos"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		// Clerk
		ClerkJWKSURL: getEnv("CLERK_JWKS_URL", ""),

		// Web
		WebURL: getEnv("WEB_URL", "http://localhost:3000"),

		// AWS S3
		AWSRegion:          getEnv("AWS_REGION", "us-east-1"),
		AWSAccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
		S3BucketName:       getEnv("S3_BUCKET_NAME", "lvos-files"),

		// GCP Cloud Storage
		GCSBucketName: getEnv("GCS_BUCKET_NAME", "farohq-files"),
		GCSProjectID:  getEnv("GCS_PROJECT_ID", ""),

		// Vercel API
		VercelAPIToken:  getEnv("VERCEL_API_TOKEN", ""),
		VercelProjectID: getEnv("VERCEL_PROJECT_ID", ""),
		VercelTeamID:    getEnv("VERCEL_TEAM_ID", ""),

		// DNS (optional)
		DNSLookupEnabled: getEnv("DNS_LOOKUP_ENABLED", "false") == "true",

		// Email Service
		PostmarkAPIToken:  getEnv("POSTMARK_API_TOKEN", ""),
		PostmarkFromEmail: getEnv("POSTMARK_FROM_EMAIL", ""),
		MailhogHost:       getEnv("MAILHOG_HOST", "localhost"),
		MailhogPort:       getEnv("MAILHOG_PORT", "8025"),
		Environment:       getEnv("ENVIRONMENT", "development"),

		// Server
		Port: getEnv("PORT", "8080"),

		// Redis/Dragonfly Cache
		RedisURL: getEnv("REDIS_URL", "redis://localhost:6379"),
	}

	return cfg
}

// DatabaseDSN builds a PostgreSQL connection string (DSN) from config
// Supports both TCP connections (local/dev) and Unix socket connections (Cloud Run)
//
// TCP format: postgres://user:password@host:port/dbname?sslmode=mode
// Unix socket format: postgres://user:password@/dbname?host=/cloudsql/project:region:instance&sslmode=mode
//
// Cloud Run uses Unix sockets when DB_HOST starts with /cloudsql/
func (c *Config) DatabaseDSN() string {
	// URL encode password to handle special characters
	encodedPassword := url.QueryEscape(c.DBPassword)
	encodedUser := url.QueryEscape(c.DBUser)
	encodedDBName := url.QueryEscape(c.DBName)

	// Check if using Unix socket (Cloud Run with Cloud SQL)
	if strings.HasPrefix(c.DBHost, "/cloudsql/") {
		// Unix socket connection format for Cloud Run
		// Format: postgres://user:password@/dbname?host=/cloudsql/instance&sslmode=mode
		return fmt.Sprintf("postgres://%s:%s@/%s?host=%s&sslmode=%s",
			encodedUser,
			encodedPassword,
			encodedDBName,
			c.DBHost, // e.g., /cloudsql/project:region:instance
			c.DBSSLMode,
		)
	}

	// TCP connection format (local/dev via proxy)
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		encodedUser,
		encodedPassword,
		c.DBHost,
		c.DBPort,
		encodedDBName,
		c.DBSSLMode,
	)
}

// DatabaseURL returns DATABASE_URL if set, otherwise builds DSN from config
func (c *Config) DatabaseURL() string {
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		return dbURL
	}
	return c.DatabaseDSN()
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
