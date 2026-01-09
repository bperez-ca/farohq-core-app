package http

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all user domain routes
func (h *Handlers) RegisterRoutes(r chi.Router) {
	r.Route("/users", func(r chi.Router) {
		r.Post("/sync", h.SyncUserHandler)
	})
}
