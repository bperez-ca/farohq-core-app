package http

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"
)

// Handlers provides HTTP handlers for the auth domain
type Handlers struct {
	logger zerolog.Logger
}

// NewHandlers creates new auth HTTP handlers
func NewHandlers(logger zerolog.Logger) *Handlers {
	return &Handlers{
		logger: logger,
	}
}

// MeHandler returns current user info from Clerk token in context
func (h *Handlers) MeHandler(w http.ResponseWriter, r *http.Request) {
	// Get user info from context (set by RequireAuth middleware)
	userID := r.Context().Value("user_id")
	email := r.Context().Value("email")
	orgID := r.Context().Value("org_id")
	orgSlug := r.Context().Value("org_slug")
	orgRole := r.Context().Value("org_role")

	response := map[string]interface{}{
		"user_id":  userID,
		"email":    email,
		"org_id":   orgID,
		"org_slug": orgSlug,
		"org_role": orgRole,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

