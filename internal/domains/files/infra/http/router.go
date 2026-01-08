package http

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all files domain routes
func (h *Handlers) RegisterRoutes(r chi.Router) {
	r.Route("/files", func(r chi.Router) {
		r.Get("/", h.ListFilesHandler) // List files (placeholder - returns empty list)
		r.Post("/sign", h.SignHandler)
		r.Delete("/{key}", h.DeleteFileHandler)
	})
}
