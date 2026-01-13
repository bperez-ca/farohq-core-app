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
		// Domain verification routes (Scale tier only)
		r.Post("/{brandId}/verify-domain", h.VerifyDomainHandler)
		r.Get("/{brandId}/domain-status", h.GetDomainStatusHandler)
		r.Get("/{brandId}/domain-instructions", h.GetDomainInstructionsHandler)
		r.Get("/{brandId}/ssl-status", h.GetSSLStatusHandler)
	})
}

// RegisterPublicRoutes registers public brand routes (no auth)
func (h *Handlers) RegisterPublicRoutes(r chi.Router) {
	r.Route("/brand", func(r chi.Router) {
		r.Get("/by-domain", h.GetByDomainHandler)
		r.Get("/by-host", h.GetByHostHandler)
		r.Get("/by-subdomain", h.GetBySubdomainHandler)
	})
}

