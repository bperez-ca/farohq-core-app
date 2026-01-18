package tenant

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a test database connection
// This should be used for integration tests
func setupTestDB(t *testing.T) *pgxpool.Pool {
	dbURL := "postgres://postgres:password@localhost:5432/localvisibilityos?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Skipf("Database not available, skipping integration test: %v", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		t.Skipf("Database not available, skipping integration test: %v", err)
	}

	return pool
}

func TestExtractTenantIDFromURL(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "valid tenant ID in URL",
			path:     "/api/v1/tenants/123e4567-e89b-12d3-a456-426614174000/invites",
			expected: "123e4567-e89b-12d3-a456-426614174000",
		},
		{
			name:     "tenant ID at end of path",
			path:     "/api/v1/tenants/123e4567-e89b-12d3-a456-426614174000",
			expected: "123e4567-e89b-12d3-a456-426614174000",
		},
		{
			name:     "invalid UUID format",
			path:     "/api/v1/tenants/not-a-uuid/invites",
			expected: "",
		},
		{
			name:     "no tenants in path",
			path:     "/api/v1/brands/123e4567-e89b-12d3-a456-426614174000",
			expected: "",
		},
		{
			name:     "empty path",
			path:     "",
			expected: "",
		},
		{
			name:     "path without tenant ID",
			path:     "/api/v1/tenants",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractTenantIDFromURL(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResolver_ValidateUserAccess(t *testing.T) {
	pool := setupTestDB(t)
	logger := zerolog.Nop()
	resolver := NewResolver(pool, logger)

	ctx := context.Background()

	// Create test user and tenant
	userID := uuid.New()
	tenantID := uuid.New()
	otherTenantID := uuid.New()

	// Setup: Create tenant_members entry
	_, err := pool.Exec(ctx, `
		INSERT INTO tenant_members (id, tenant_id, user_id, role, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, 'owner', NOW(), NOW())
	`, tenantID, userID)
	require.NoError(t, err)

	// Cleanup
	defer func() {
		pool.Exec(ctx, "DELETE FROM tenant_members WHERE user_id = $1", userID)
	}()

	tests := []struct {
		name          string
		userID        uuid.UUID
		tenantID      string
		expectedValid bool
		expectedError bool
	}{
		{
			name:          "user has access to tenant",
			userID:        userID,
			tenantID:      tenantID.String(),
			expectedValid: true,
			expectedError: false,
		},
		{
			name:          "user does not have access to tenant",
			userID:        userID,
			tenantID:      otherTenantID.String(),
			expectedValid: false,
			expectedError: false,
		},
		{
			name:          "invalid tenant ID format",
			userID:        userID,
			tenantID:      "not-a-uuid",
			expectedValid: false,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := resolver.ValidateUserAccess(ctx, tt.userID, tt.tenantID)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedValid, valid)
		})
	}
}

func TestResolver_GetUserTenantIDs(t *testing.T) {
	pool := setupTestDB(t)
	logger := zerolog.Nop()
	resolver := NewResolver(pool, logger)

	ctx := context.Background()

	// Create test user and tenants
	userID := uuid.New()
	tenant1ID := uuid.New()
	tenant2ID := uuid.New()

	// Setup: Create tenant_members entries
	_, err := pool.Exec(ctx, `
		INSERT INTO tenant_members (id, tenant_id, user_id, role, created_at, updated_at)
		VALUES 
			(gen_random_uuid(), $1, $3, 'owner', NOW(), NOW()),
			(gen_random_uuid(), $2, $3, 'admin', NOW(), NOW())
	`, tenant1ID, tenant2ID, userID)
	require.NoError(t, err)

	// Cleanup
	defer func() {
		pool.Exec(ctx, "DELETE FROM tenant_members WHERE user_id = $1", userID)
	}()

	tenantIDs, err := resolver.GetUserTenantIDs(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, tenantIDs, 2)
	assert.Contains(t, tenantIDs, tenant1ID.String())
	assert.Contains(t, tenantIDs, tenant2ID.String())
}

func TestResolver_ResolveTenantByDomain(t *testing.T) {
	pool := setupTestDB(t)
	logger := zerolog.Nop()
	resolver := NewResolver(pool, logger)

	ctx := context.Background()

	// Create test tenant and branding
	tenantID := uuid.New()
	domain := "test.example.com"

	// Setup: Create branding entry
	_, err := pool.Exec(ctx, `
		INSERT INTO branding (id, agency_id, domain, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, NOW(), NOW())
	`, tenantID, domain)
	require.NoError(t, err)

	// Cleanup
	defer func() {
		pool.Exec(ctx, "DELETE FROM branding WHERE domain = $1", domain)
	}()

	tests := []struct {
		name          string
		host          string
		expectedID    string
		expectedError bool
	}{
		{
			name:          "valid domain",
			host:          domain,
			expectedID:    tenantID.String(),
			expectedError: false,
		},
		{
			name:          "domain with port",
			host:          domain + ":8080",
			expectedID:    tenantID.String(),
			expectedError: false,
		},
		{
			name:          "case insensitive domain",
			host:          "TEST.EXAMPLE.COM",
			expectedID:    tenantID.String(),
			expectedError: false,
		},
		{
			name:          "non-existent domain",
			host:          "nonexistent.example.com",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.ResolveTenantByDomain(ctx, tt.host)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, ErrTenantNotFound, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, result)
			}
		})
	}
}

// Note: TestResolver_ResolveTenantWithValidation requires a full database setup
// with users, tenants, tenant_members, and branding tables.
// This is an integration test that should be run with a test database.
func TestResolver_ResolveTenantWithValidation_Integration(t *testing.T) {
	pool := setupTestDB(t)
	logger := zerolog.Nop()
	resolver := NewResolver(pool, logger)

	ctx := context.Background()

	// Create test user and tenants
	userID := uuid.New()
	tenant1ID := uuid.New()
	tenant2ID := uuid.New()
	domain := "test-tenant.example.com"

	// Setup: Create user
	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, clerk_user_id, email, created_at, updated_at)
		VALUES ($1, 'clerk_' || $1::text, 'test@example.com', NOW(), NOW())
	`, userID)
	require.NoError(t, err)

	// Setup: Create tenant_members entries
	_, err = pool.Exec(ctx, `
		INSERT INTO tenant_members (id, tenant_id, user_id, role, created_at, updated_at)
		VALUES 
			(gen_random_uuid(), $1, $3, 'owner', NOW(), NOW()),
			(gen_random_uuid(), $2, $3, 'admin', NOW(), NOW())
	`, tenant1ID, tenant2ID, userID)
	require.NoError(t, err)

	// Setup: Create branding entry for tenant1
	_, err = pool.Exec(ctx, `
		INSERT INTO branding (id, agency_id, domain, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, NOW(), NOW())
	`, tenant1ID, domain)
	require.NoError(t, err)

	// Cleanup
	defer func() {
		pool.Exec(ctx, "DELETE FROM branding WHERE domain = $1", domain)
		pool.Exec(ctx, "DELETE FROM tenant_members WHERE user_id = $1", userID)
		pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	}()

	tests := []struct {
		name            string
		host            string
		tenantIDHeader  string
		urlPath         string
		expectedTenant  string
		expectedSource  TenantSource
		expectedValid   bool
		expectedError   bool
	}{
		{
			name:           "resolve by domain",
			host:            domain,
			expectedTenant:  tenant1ID.String(),
			expectedSource:  TenantSourceDomain,
			expectedValid:   true,
			expectedError:   false,
		},
		{
			name:           "resolve by header",
			tenantIDHeader: tenant1ID.String(),
			expectedTenant: tenant1ID.String(),
			expectedSource: TenantSourceHeader,
			expectedValid:  true,
			expectedError:  false,
		},
		{
			name:           "resolve by URL",
			urlPath:        "/api/v1/tenants/" + tenant1ID.String() + "/invites",
			expectedTenant: tenant1ID.String(),
			expectedSource: TenantSourceURL,
			expectedValid:  true,
			expectedError:  false,
		},
		{
			name:           "fallback to first tenant when no source",
			expectedTenant: tenant1ID.String(), // First tenant by created_at
			expectedSource: TenantSourceToken,
			expectedValid:     true,
			expectedError:    false,
		},
		{
			name:           "invalid tenant in header falls back to first tenant",
			tenantIDHeader: uuid.New().String(), // Tenant user doesn't have access to
			expectedTenant: tenant1ID.String(),   // Falls back to first accessible tenant
			expectedSource: TenantSourceFallback,
			expectedValid:  true,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.ResolveTenantWithValidation(
				ctx,
				userID,
				tt.host,
				tt.tenantIDHeader,
				tt.urlPath,
			)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedTenant, result.TenantID)
				assert.Equal(t, tt.expectedSource, result.Source)
				assert.Equal(t, tt.expectedValid, result.Validated)
			}
		})
	}
}
