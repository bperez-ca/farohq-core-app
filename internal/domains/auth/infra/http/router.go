package http

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all auth domain routes
func (h *Handlers) RegisterRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Get("/me", h.MeHandler)
	})
}

