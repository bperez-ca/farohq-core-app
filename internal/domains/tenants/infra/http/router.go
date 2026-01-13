package http

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all tenant domain routes
func (h *Handlers) RegisterRoutes(r chi.Router) {
	// Tenants routes
	r.Route("/tenants", func(r chi.Router) {
		r.Post("/", h.CreateTenantHandler)
		r.Post("/onboard", h.OnboardTenantHandler)
		r.Get("/my-orgs", h.ListTenantsByUserHandler)
		r.Get("/validate-slug", h.ValidateSlugHandler)
		r.Get("/{id}", h.GetTenantHandler)
		r.Put("/{id}", h.UpdateTenantHandler)
		r.Post("/{id}/invites", h.InviteMemberHandler)
		r.Get("/{id}/members", h.ListMembersHandler)
		r.Delete("/{id}/members/{user_id}", h.RemoveMemberHandler)
		r.Get("/{id}/roles", h.ListRolesHandler)
		r.Get("/{id}/seat-usage", h.GetSeatUsageHandler)
		// Client routes
		r.Post("/{id}/clients", h.CreateClientHandler)
		r.Get("/{id}/clients", h.ListClientsHandler)
	})

	// Clients routes
	r.Route("/clients", func(r chi.Router) {
		r.Get("/{id}", h.GetClientHandler)
		r.Put("/{id}", h.UpdateClientHandler)
		r.Post("/{id}/members", h.AddClientMemberHandler)
		r.Get("/{id}/members", h.ListClientMembersHandler)
		r.Delete("/{id}/members/{memberId}", h.RemoveClientMemberHandler)
		r.Post("/{id}/locations", h.CreateLocationHandler)
		r.Get("/{id}/locations", h.ListLocationsHandler)
	})

	// Locations routes
	r.Route("/locations", func(r chi.Router) {
		r.Put("/{id}", h.UpdateLocationHandler)
	})
}

