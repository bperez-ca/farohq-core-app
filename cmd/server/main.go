package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	app_composition "farohq-core-app/internal/app/composition"
	"farohq-core-app/internal/app/health"
	"farohq-core-app/internal/platform/cache"
	"farohq-core-app/internal/platform/config"
	"farohq-core-app/internal/platform/db"
	"farohq-core-app/internal/platform/httpserver"
	"farohq-core-app/internal/platform/logging"
	"farohq-core-app/internal/platform/tenant"
)

func main() {
	// Load .env file if it exists
	godotenv.Load()

	// Setup structured logger
	logger := logging.NewLogger()

	// Initialize configuration
	cfg := config.NewConfig()

	// Initialize database connection
	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL(), logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer pool.Close()

	// Initialize Redis/Dragonfly client for caching
	var redisClient *redis.Client
	var tenantCache *tenant.TenantCache
	if cfg.RedisURL != "" {
		client, err := cache.NewRedisClient(cfg.RedisURL, logger)
		if err != nil {
			logger.Warn().
				Err(err).
				Str("redis_url", cfg.RedisURL).
				Msg("Failed to connect to Redis/Dragonfly, continuing without cache")
		} else {
			redisClient = client
			// Create tenant cache with 5 minute TTL
			tenantCache = tenant.NewTenantCache(client, 5*time.Minute, logger)
			logger.Info().Msg("Tenant cache enabled with Dragonfly/Redis")
		}
	} else {
		logger.Info().Msg("Redis URL not configured, tenant cache disabled")
	}

	// Initialize tenant resolver
	tenantResolver := tenant.NewResolver(pool, logger)

	// Set cache on resolver if available
	if tenantCache != nil {
		tenantResolver.SetCache(tenantCache)
	}

	// Initialize authentication middleware
	authMiddleware, err := httpserver.NewRequireAuth(cfg.ClerkJWKSURL, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize authentication middleware")
	}

	// Setup router
	r := chi.NewRouter()

	// Apply common middleware
	for _, mw := range httpserver.CommonMiddleware(logger) {
		r.Use(mw)
	}

	// Initialize composition (wires all domains together) - needed for user repo
	appComposition := app_composition.NewComposition(pool, cfg, logger)

	// Initialize health handlers
	healthHandlers := health.NewHandlers(pool)

	// Health check endpoints
	r.Get("/healthz", healthHandlers.Healthz)
	r.Get("/readyz", healthHandlers.Readyz)

	// API info endpoint
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"name": "FaroHQ Core App",
			"version": "1.0.0",
			"status": "running",
		"timestamp": "%s"
	}`, time.Now().Format(time.RFC3339))
	})

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// CRITICAL: Register /invites/by-email FIRST (before public routes) to avoid route conflict with /invites/{token}
		// This specific route must be registered before the parameterized public route
		// NOTE: This route does NOT require tenant resolution because it's used by users who don't have a tenant yet
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequireAuth)
			// #region agent log
			r.Get("/invites/by-email", func(w http.ResponseWriter, r *http.Request) {
				logFile, _ := os.OpenFile("/Users/bperez/Projects/farohq-core-app/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				json.NewEncoder(logFile).Encode(map[string]interface{}{"timestamp": time.Now().UnixMilli(), "location": "main.go:150", "message": "invites/by-email route matched - BEFORE handler", "hypothesisId": "ROUTE", "sessionId": "debug-session", "runId": "run1", "data": map[string]interface{}{"path": r.URL.Path, "method": r.Method, "query": r.URL.RawQuery}})
				logFile.Close()
				// #endregion
				appComposition.TenantHandlers.FindInvitesByEmailHandler(w, r)
			})
			// #endregion
		})

		// Public routes (no auth required) - MUST be registered AFTER specific routes to avoid conflicts
		appComposition.RegisterPublicRoutes(r)

		// Protected routes (auth required) - Register specific routes FIRST to avoid conflicts with parameterized public routes
		r.Group(func(r chi.Router) {
			// 1. Authenticate first
			r.Use(authMiddleware.RequireAuth)

			// 2. Resolve tenant with validation (runs AFTER authentication)
			r.Use(httpserver.TenantResolutionWithAuth(
				tenantResolver,
				tenantCache,
				appComposition.UserRepo,
				pool,
				logger,
			))

			// Routes that don't require tenant context
			r.Route("/tenants", func(r chi.Router) {
				r.Post("/", appComposition.TenantHandlers.CreateTenantHandler)
				r.Post("/onboard", appComposition.TenantHandlers.OnboardTenantHandler)
			})
			// Register /tenants/my-orgs at top level (outside Route) to ensure it matches before Group middleware
			// #region agent log
			r.Get("/tenants/my-orgs", func(w http.ResponseWriter, r *http.Request) {
				logFile, _ := os.OpenFile("/Users/bperez/Projects/farohq-core-app/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				json.NewEncoder(logFile).Encode(map[string]interface{}{"timestamp": time.Now().UnixMilli(), "location": "main.go:100", "message": "my-orgs route matched - BEFORE RequireTenantContext", "hypothesisId": "H1,H2,H5", "sessionId": "debug-session", "runId": "run3", "data": map[string]interface{}{"path": r.URL.Path, "method": r.Method}})
				logFile.Close()
				appComposition.TenantHandlers.ListTenantsByUserHandler(w, r)
			})
			// #endregion
			r.Route("/auth", func(r chi.Router) {
				r.Get("/me", appComposition.AuthHandlers.MeHandler)
			})
			r.Route("/users", func(r chi.Router) {
				r.Post("/sync", appComposition.UserHandlers.SyncUserHandler)
			})
			r.Route("/invites", func(r chi.Router) {
				r.Post("/accept", appComposition.TenantHandlers.AcceptInviteHandler)
			})
		})

		// All other protected routes require tenant context
		// #region agent log
		r.Group(func(r chi.Router) {
			// 1. Authenticate first
			r.Use(authMiddleware.RequireAuth)

			// 2. Resolve tenant with validation (runs AFTER authentication)
			r.Use(httpserver.TenantResolutionWithAuth(
				tenantResolver,
				tenantCache,
				appComposition.UserRepo,
				pool,
				logger,
			))

			// #region agent log
			mw := httpserver.RequireTenantContext
			wrappedMw := func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					logFile, _ := os.OpenFile("/Users/bperez/Projects/farohq-core-app/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					json.NewEncoder(logFile).Encode(map[string]interface{}{"timestamp": time.Now().UnixMilli(), "location": "main.go:116", "message": "Group middleware wrapper: about to apply RequireTenantContext", "hypothesisId": "H4", "sessionId": "debug-session", "runId": "run2", "data": map[string]interface{}{"path": r.URL.Path, "method": r.Method}})
					logFile.Close()
					// #endregion
					mw(next).ServeHTTP(w, r)
				})
			}
			r.Use(wrappedMw)
			// #endregion
			appComposition.RegisterProtectedRoutesWithTenant(r)
		})
		// #endregion
	})

	// Start server
	port := cfg.Port
	server := httpserver.NewServer(":"+port, r, logger)

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info().Str("port", port).Msg("Starting FaroHQ Core App")
		if err := server.Start(); err != nil {
			logger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	// Close Redis connection if it was opened
	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			logger.Warn().Err(err).Msg("Error closing Redis connection")
		} else {
			logger.Info().Msg("Redis connection closed")
		}
	}

	logger.Info().Msg("Core App exited gracefully")
}
