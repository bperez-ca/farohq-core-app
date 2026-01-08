package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"farohq-core-app/internal/platform/db"
)

// Handlers provides health check endpoints
type Handlers struct {
	db *pgxpool.Pool
}

// NewHandlers creates new health check handlers
func NewHandlers(db *pgxpool.Pool) *Handlers {
	return &Handlers{
		db: db,
	}
}

// Healthz handles liveness probe (always returns 200)
func (h *Handlers) Healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Readyz handles readiness probe (checks database connectivity)
func (h *Handlers) Readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := db.HealthCheck(ctx, h.db); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}

