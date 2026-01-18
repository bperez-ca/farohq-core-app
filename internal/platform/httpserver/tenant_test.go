package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"farohq-core-app/internal/domains/users/infra/db"
	"farohq-core-app/internal/platform/tenant"
)

// setupTestDBForTenant creates a test database connection for tenant tests
func setupTestDBForTenant(t *testing.T) *pgxpool.Pool {
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

func TestTenantResolutionWithAuth_Success(t *testing.T) {
	pool := setupTestDBForTenant(t)
	logger := zerolog.Nop()
	tenantResolver := tenant.NewResolver(pool, logger)
	userRepo := db.NewUserRepository(pool)

	ctx := context.Background()

	// Create test user
	userID := uuid.New()
	clerkUserID := "clerk_test_user_123"
	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, clerk_user_id, email, created_at, updated_at)
		VALUES ($1, $2, 'test@example.com', NOW(), NOW())
	`, userID, clerkUserID)
	require.NoError(t, err)

	// Create test tenant
	tenantID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO agencies (id, name, slug, status, tier, agency_seat_limit, created_at, updated_at)
		VALUES ($1, 'Test Agency', 'test-agency', 'active', 'growth', 10, NOW(), NOW())
	`, tenantID)
	require.NoError(t, err)

	// Create tenant_members entry
	_, err = pool.Exec(ctx, `
		INSERT INTO tenant_members (id, tenant_id, user_id, role, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, 'owner', NOW(), NOW())
	`, tenantID, userID)
	require.NoError(t, err)

	// Cleanup
	defer func() {
		pool.Exec(ctx, "DELETE FROM tenant_members WHERE user_id = $1", userID)
		pool.Exec(ctx, "DELETE FROM agencies WHERE id = $1", tenantID)
		pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	}()

	// Create middleware
	middleware := TenantResolutionWithAuth(tenantResolver, nil, userRepo, pool, logger)

	// Create test handler
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify tenant context is set
		tenantIDFromCtx, ok := tenant.GetTenantFromContext(r.Context())
		assert.True(t, ok, "Tenant ID should be in context")
		assert.Equal(t, tenantID.String(), tenantIDFromCtx)
		w.WriteHeader(http.StatusOK)
	}))

	// Create request with user_id in context (set by auth middleware)
	req := httptest.NewRequest("GET", "/api/v1/tenants/"+tenantID.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", clerkUserID))

	// Add X-Tenant-ID header
	req.Header.Set("X-Tenant-ID", tenantID.String())

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTenantResolutionWithAuth_InvalidAccess(t *testing.T) {
	pool := setupTestDBForTenant(t)
	logger := zerolog.Nop()
	tenantResolver := tenant.NewResolver(pool, logger)
	userRepo := db.NewUserRepository(pool)

	ctx := context.Background()

	// Create test user
	userID := uuid.New()
	clerkUserID := "clerk_test_user_456"
	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, clerk_user_id, email, created_at, updated_at)
		VALUES ($1, $2, 'test@example.com', NOW(), NOW())
	`, userID, clerkUserID)
	require.NoError(t, err)

	// Create test tenant that user has access to
	accessibleTenantID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO agencies (id, name, slug, status, tier, agency_seat_limit, created_at, updated_at)
		VALUES ($1, 'Accessible Agency', 'accessible-agency', 'active', 'growth', 10, NOW(), NOW())
	`, accessibleTenantID)
	require.NoError(t, err)

	// Create tenant_members entry for accessible tenant
	_, err = pool.Exec(ctx, `
		INSERT INTO tenant_members (id, tenant_id, user_id, role, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, 'owner', NOW(), NOW())
	`, accessibleTenantID, userID)
	require.NoError(t, err)

	// Create tenant that user does NOT have access to
	inaccessibleTenantID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO agencies (id, name, slug, status, tier, agency_seat_limit, created_at, updated_at)
		VALUES ($1, 'Inaccessible Agency', 'inaccessible-agency', 'active', 'growth', 10, NOW(), NOW())
	`, inaccessibleTenantID)
	require.NoError(t, err)

	// Cleanup
	defer func() {
		pool.Exec(ctx, "DELETE FROM tenant_members WHERE user_id = $1", userID)
		pool.Exec(ctx, "DELETE FROM agencies WHERE id IN ($1, $2)", accessibleTenantID, inaccessibleTenantID)
		pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	}()

	// Create middleware
	middleware := TenantResolutionWithAuth(tenantResolver, nil, userRepo, pool, logger)

	// Create test handler
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should not reach here, but if it does, verify fallback tenant
		tenantIDFromCtx, ok := tenant.GetTenantFromContext(r.Context())
		assert.True(t, ok, "Tenant ID should be in context (fallback)")
		// Should fallback to accessible tenant
		assert.Equal(t, accessibleTenantID.String(), tenantIDFromCtx)
		w.WriteHeader(http.StatusOK)
	}))

	// Create request with user_id in context
	req := httptest.NewRequest("GET", "/api/v1/tenants/"+inaccessibleTenantID.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", clerkUserID))

	// Add X-Tenant-ID header with inaccessible tenant
	req.Header.Set("X-Tenant-ID", inaccessibleTenantID.String())

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Should succeed but use fallback tenant
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTenantResolutionWithAuth_NoAccessibleTenants(t *testing.T) {
	pool := setupTestDBForTenant(t)
	logger := zerolog.Nop()
	tenantResolver := tenant.NewResolver(pool, logger)
	userRepo := db.NewUserRepository(pool)

	ctx := context.Background()

	// Create test user with NO tenant access
	userID := uuid.New()
	clerkUserID := "clerk_test_user_789"
	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, clerk_user_id, email, created_at, updated_at)
		VALUES ($1, $2, 'test@example.com', NOW(), NOW())
	`, userID, clerkUserID)
	require.NoError(t, err)

	// Cleanup
	defer func() {
		pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	}()

	// Create middleware
	middleware := TenantResolutionWithAuth(tenantResolver, nil, userRepo, pool, logger)

	// Create test handler
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Handler should not be called when user has no accessible tenants")
	}))

	// Create request with user_id in context
	req := httptest.NewRequest("GET", "/api/v1/tenants/test", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", clerkUserID))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Should return 403 Forbidden
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestTenantResolutionWithAuth_PublicRoutes(t *testing.T) {
	pool := setupTestDBForTenant(t)
	logger := zerolog.Nop()
	tenantResolver := tenant.NewResolver(pool, logger)
	userRepo := db.NewUserRepository(pool)

	// Create middleware
	middleware := TenantResolutionWithAuth(tenantResolver, nil, userRepo, pool, logger)

	// Create test handler
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	publicRoutes := []string{
		"/healthz",
		"/readyz",
		"/",
		"/api/v1/tenants/my-orgs",
		"/api/v1/auth/me",
		"/api/v1/users/sync",
	}

	for _, route := range publicRoutes {
		t.Run(route, func(t *testing.T) {
			req := httptest.NewRequest("GET", route, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			// Public routes should pass through without tenant resolution
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestTenantResolutionWithAuth_UserNotFound(t *testing.T) {
	pool := setupTestDBForTenant(t)
	logger := zerolog.Nop()
	tenantResolver := tenant.NewResolver(pool, logger)
	userRepo := db.NewUserRepository(pool)

	// Create middleware
	middleware := TenantResolutionWithAuth(tenantResolver, nil, userRepo, pool, logger)

	// Create test handler
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Handler should not be called when user is not found")
	}))

	// Create request with non-existent user_id in context
	req := httptest.NewRequest("GET", "/api/v1/tenants/test", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", "clerk_nonexistent_user"))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Should return 404 Not Found
	assert.Equal(t, http.StatusNotFound, w.Code)
}
