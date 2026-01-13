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

	app_composition "farohq-core-app/internal/app/composition"
	"farohq-core-app/internal/app/health"
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

	// Initialize tenant resolver
	tenantResolver := tenant.NewResolver(pool, logger)

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

	// Add tenant resolution middleware (with database connection for RLS)
	r.Use(httpserver.TenantResolution(tenantResolver, pool, logger))

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

	// Initialize composition (wires all domains together)
	appComposition := app_composition.NewComposition(pool, cfg, logger)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes (no auth required)
		appComposition.RegisterPublicRoutes(r)

		// Protected routes (auth required)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequireAuth)

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

			// All other protected routes require tenant context
			// #region agent log
			r.Group(func(r chi.Router) {
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

	logger.Info().Msg("Core App exited gracefully")
}

