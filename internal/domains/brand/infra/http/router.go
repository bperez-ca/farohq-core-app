package http

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all brand domain routes (protected)
func (h *Handlers) RegisterRoutes(r chi.Router) {
	// Protected brand routes (auth required)
	r.Route("/brands", func(r chi.Router) {
		r.Get("/", h.ListBrandsHandler)
		r.Post("/", h.CreateBrandHandler)
		r.Get("/{brandId}", h.GetBrandHandler)
		r.Put("/{brandId}", h.UpdateBrandHandler)
		r.Delete("/{brandId}", h.DeleteBrandHandler)
	})
}

// RegisterPublicRoutes registers public brand routes (no auth)
func (h *Handlers) RegisterPublicRoutes(r chi.Router) {
	r.Route("/brand", func(r chi.Router) {
		r.Get("/by-domain", h.GetByDomainHandler)
		r.Get("/by-host", h.GetByHostHandler)
	})
}

