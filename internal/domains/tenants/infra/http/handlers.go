package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/model"
	"farohq-core-app/internal/domains/tenants/app/usecases"
)

// Handlers provides HTTP handlers for the tenants domain
type Handlers struct {
	logger            zerolog.Logger
	createTenant      *usecases.CreateTenant
	getTenant         *usecases.GetTenant
	updateTenant      *usecases.UpdateTenant
	inviteMember      *usecases.InviteMember
	acceptInvite      *usecases.AcceptInvite
	listMembers       *usecases.ListMembers
	removeMember      *usecases.RemoveMember
	listRoles         *usecases.ListRoles
	createClient      *usecases.CreateClient
	listClients       *usecases.ListClients
	getClient         *usecases.GetClient
	updateClient      *usecases.UpdateClient
	addClientMember   *usecases.AddClientMember
	listClientMembers *usecases.ListClientMembers
	removeClientMember *usecases.RemoveClientMember
	createLocation    *usecases.CreateLocation
	listLocations     *usecases.ListLocations
	updateLocation    *usecases.UpdateLocation
	getSeatUsage      *usecases.GetSeatUsage
}

// NewHandlers creates new tenants HTTP handlers
func NewHandlers(
	logger zerolog.Logger,
	createTenant *usecases.CreateTenant,
	getTenant *usecases.GetTenant,
	updateTenant *usecases.UpdateTenant,
	inviteMember *usecases.InviteMember,
	acceptInvite *usecases.AcceptInvite,
	listMembers *usecases.ListMembers,
	removeMember *usecases.RemoveMember,
	listRoles *usecases.ListRoles,
	createClient *usecases.CreateClient,
	listClients *usecases.ListClients,
	getClient *usecases.GetClient,
	updateClient *usecases.UpdateClient,
	addClientMember *usecases.AddClientMember,
	listClientMembers *usecases.ListClientMembers,
	removeClientMember *usecases.RemoveClientMember,
	createLocation *usecases.CreateLocation,
	listLocations *usecases.ListLocations,
	updateLocation *usecases.UpdateLocation,
	getSeatUsage *usecases.GetSeatUsage,
) *Handlers {
	return &Handlers{
		logger:            logger,
		createTenant:      createTenant,
		getTenant:         getTenant,
		updateTenant:      updateTenant,
		inviteMember:      inviteMember,
		acceptInvite:      acceptInvite,
		listMembers:       listMembers,
		removeMember:      removeMember,
		listRoles:         listRoles,
		createClient:      createClient,
		listClients:       listClients,
		getClient:         getClient,
		updateClient:      updateClient,
		addClientMember:   addClientMember,
		listClientMembers: listClientMembers,
		removeClientMember: removeClientMember,
		createLocation:    createLocation,
		listLocations:     listLocations,
		updateLocation:    updateLocation,
		getSeatUsage:      getSeatUsage,
	}
}

// CreateTenantHandler handles POST /api/v1/tenants
func (h *Handlers) CreateTenantHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	createReq := &usecases.CreateTenantRequest{
		Name: req.Name,
		Slug: req.Slug,
	}

	resp, err := h.createTenant.Execute(r.Context(), createReq)
	if err != nil {
		if err == domain.ErrTenantAlreadyExists {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		if err == domain.ErrInvalidTenantName || err == domain.ErrInvalidTenantSlug {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to create tenant")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         resp.Tenant.ID().String(),
		"name":       resp.Tenant.Name(),
		"slug":       resp.Tenant.Slug(),
		"status":     string(resp.Tenant.Status()),
		"created_at": resp.Tenant.CreatedAt().Format(time.RFC3339),
	})
}

// GetTenantHandler handles GET /api/v1/tenants/{id}
func (h *Handlers) GetTenantHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "id")
	if tenantID == "" {
		http.Error(w, "tenant ID is required", http.StatusBadRequest)
		return
	}

	id, err := parseUUID(tenantID)
	if err != nil {
		http.Error(w, "invalid tenant ID", http.StatusBadRequest)
		return
	}

	getReq := &usecases.GetTenantRequest{
		TenantID: id,
	}

	resp, err := h.getTenant.Execute(r.Context(), getReq)
	if err != nil {
		if err == domain.ErrTenantNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to get tenant")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         resp.Tenant.ID().String(),
		"name":       resp.Tenant.Name(),
		"slug":       resp.Tenant.Slug(),
		"status":     string(resp.Tenant.Status()),
		"created_at": resp.Tenant.CreatedAt().Format(time.RFC3339),
	})
}

// UpdateTenantHandler handles PUT /api/v1/tenants/{id}
func (h *Handlers) UpdateTenantHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "id")
	if tenantID == "" {
		http.Error(w, "tenant ID is required", http.StatusBadRequest)
		return
	}

	id, err := parseUUID(tenantID)
	if err != nil {
		http.Error(w, "invalid tenant ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name   *string `json:"name"`
		Slug   *string `json:"slug"`
		Status *string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	updateReq := &usecases.UpdateTenantRequest{
		TenantID: id,
		Name:     req.Name,
		Slug:     req.Slug,
	}

	if req.Status != nil {
		status := model.TenantStatus(*req.Status)
		updateReq.Status = &status
	}

	resp, err := h.updateTenant.Execute(r.Context(), updateReq)
	if err != nil {
		if err == domain.ErrTenantNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err == domain.ErrTenantAlreadyExists {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		if err == domain.ErrInvalidTenantName || err == domain.ErrInvalidTenantSlug {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to update tenant")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         resp.Tenant.ID().String(),
		"name":       resp.Tenant.Name(),
		"slug":       resp.Tenant.Slug(),
		"status":     string(resp.Tenant.Status()),
		"created_at": resp.Tenant.CreatedAt().Format(time.RFC3339),
	})
}

// InviteMemberHandler handles POST /api/v1/tenants/{id}/invites
func (h *Handlers) InviteMemberHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "id")
	if tenantID == "" {
		http.Error(w, "tenant ID is required", http.StatusBadRequest)
		return
	}

	id, err := parseUUID(tenantID)
	if err != nil {
		http.Error(w, "invalid tenant ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context (set by auth middleware)
	userIDStr, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "user ID required", http.StatusUnauthorized)
		return
	}
	userID, err := parseUUID(userIDStr)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	role := model.Role(req.Role)
	inviteReq := &usecases.InviteMemberRequest{
		TenantID:  id,
		Email:     req.Email,
		Role:      role,
		CreatedBy: userID,
	}

	resp, err := h.inviteMember.Execute(r.Context(), inviteReq)
	if err != nil {
		if err == domain.ErrTenantNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err == domain.ErrInvalidEmail || err == domain.ErrInvalidRole {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to invite member")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         resp.Invite.ID().String(),
		"email":      resp.Invite.Email(),
		"role":       string(resp.Invite.Role()),
		"token":      resp.Invite.Token(),
		"expires_at": resp.Invite.ExpiresAt().Format(time.RFC3339),
		"created_at": resp.Invite.CreatedAt().Format(time.RFC3339),
	})
}

// ListMembersHandler handles GET /api/v1/tenants/{id}/members
func (h *Handlers) ListMembersHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "id")
	if tenantID == "" {
		http.Error(w, "tenant ID is required", http.StatusBadRequest)
		return
	}

	id, err := parseUUID(tenantID)
	if err != nil {
		http.Error(w, "invalid tenant ID", http.StatusBadRequest)
		return
	}

	listReq := &usecases.ListMembersRequest{
		TenantID: id,
	}

	resp, err := h.listMembers.Execute(r.Context(), listReq)
	if err != nil {
		if err == domain.ErrTenantNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to list members")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	members := make([]map[string]interface{}, len(resp.Members))
	for i, member := range resp.Members {
		members[i] = map[string]interface{}{
			"id":         member.ID().String(),
			"tenant_id":  member.TenantID().String(),
			"user_id":    member.UserID().String(),
			"role":       string(member.Role()),
			"created_at": member.CreatedAt().Format(time.RFC3339),
			"updated_at": member.UpdatedAt().Format(time.RFC3339),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"members": members,
	})
}

// RemoveMemberHandler handles DELETE /api/v1/tenants/{id}/members/{user_id}
func (h *Handlers) RemoveMemberHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "id")
	if tenantID == "" {
		http.Error(w, "tenant ID is required", http.StatusBadRequest)
		return
	}

	userID := chi.URLParam(r, "user_id")
	if userID == "" {
		http.Error(w, "user ID is required", http.StatusBadRequest)
		return
	}

	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		http.Error(w, "invalid tenant ID", http.StatusBadRequest)
		return
	}

	userUUID, err := parseUUID(userID)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	removeReq := &usecases.RemoveMemberRequest{
		TenantID: tenantUUID,
		UserID:   userUUID,
	}

	resp, err := h.removeMember.Execute(r.Context(), removeReq)
	if err != nil {
		if err == domain.ErrTenantNotFound || err == domain.ErrMemberNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to remove member")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": resp.Success,
	})
}

// ListRolesHandler handles GET /api/v1/tenants/{id}/roles
func (h *Handlers) ListRolesHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "id")
	if tenantID == "" {
		http.Error(w, "tenant ID is required", http.StatusBadRequest)
		return
	}

	id, err := parseUUID(tenantID)
	if err != nil {
		http.Error(w, "invalid tenant ID", http.StatusBadRequest)
		return
	}

	listReq := &usecases.ListRolesRequest{
		TenantID: id,
	}

	resp, err := h.listRoles.Execute(r.Context(), listReq)
	if err != nil {
		if err == domain.ErrTenantNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to list roles")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"roles": resp.Roles,
	})
}

// CreateClientHandler handles POST /api/v1/tenants/{id}/clients
func (h *Handlers) CreateClientHandler(w http.ResponseWriter, r *http.Request) {
	agencyID := chi.URLParam(r, "id")
	if agencyID == "" {
		http.Error(w, "agency ID is required", http.StatusBadRequest)
		return
	}

	id, err := parseUUID(agencyID)
	if err != nil {
		http.Error(w, "invalid agency ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
		Tier string `json:"tier"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	createReq := &usecases.CreateClientRequest{
		AgencyID: id,
		Name:     req.Name,
		Slug:     req.Slug,
		Tier:     model.Tier(req.Tier),
	}

	resp, err := h.createClient.Execute(r.Context(), createReq)
	if err != nil {
		if err == domain.ErrClientAlreadyExists {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		if err == domain.ErrInvalidTenantName || err == domain.ErrInvalidTenantSlug {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to create client")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         resp.Client.ID().String(),
		"agency_id":  resp.Client.AgencyID().String(),
		"name":       resp.Client.Name(),
		"slug":       resp.Client.Slug(),
		"tier":       resp.Client.Tier().String(),
		"status":     string(resp.Client.Status()),
		"created_at": resp.Client.CreatedAt().Format(time.RFC3339),
		"updated_at": resp.Client.UpdatedAt().Format(time.RFC3339),
	})
}

// ListClientsHandler handles GET /api/v1/tenants/{id}/clients
func (h *Handlers) ListClientsHandler(w http.ResponseWriter, r *http.Request) {
	agencyID := chi.URLParam(r, "id")
	if agencyID == "" {
		http.Error(w, "agency ID is required", http.StatusBadRequest)
		return
	}

	id, err := parseUUID(agencyID)
	if err != nil {
		http.Error(w, "invalid agency ID", http.StatusBadRequest)
		return
	}

	listReq := &usecases.ListClientsRequest{
		AgencyID: id,
	}

	resp, err := h.listClients.Execute(r.Context(), listReq)
	if err != nil {
		if err == domain.ErrTenantNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to list clients")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	clients := make([]map[string]interface{}, len(resp.Clients))
	for i, client := range resp.Clients {
		clients[i] = map[string]interface{}{
			"id":         client.ID().String(),
			"agency_id":  client.AgencyID().String(),
			"name":       client.Name(),
			"slug":       client.Slug(),
			"tier":       client.Tier().String(),
			"status":     string(client.Status()),
			"created_at": client.CreatedAt().Format(time.RFC3339),
			"updated_at": client.UpdatedAt().Format(time.RFC3339),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"clients": clients,
	})
}

// GetClientHandler handles GET /api/v1/clients/{id}
func (h *Handlers) GetClientHandler(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "id")
	if clientID == "" {
		http.Error(w, "client ID is required", http.StatusBadRequest)
		return
	}

	id, err := parseUUID(clientID)
	if err != nil {
		http.Error(w, "invalid client ID", http.StatusBadRequest)
		return
	}

	getReq := &usecases.GetClientRequest{
		ClientID: id,
	}

	resp, err := h.getClient.Execute(r.Context(), getReq)
	if err != nil {
		if err == domain.ErrClientNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to get client")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         resp.Client.ID().String(),
		"agency_id":  resp.Client.AgencyID().String(),
		"name":       resp.Client.Name(),
		"slug":       resp.Client.Slug(),
		"tier":       resp.Client.Tier().String(),
		"status":     string(resp.Client.Status()),
		"created_at": resp.Client.CreatedAt().Format(time.RFC3339),
		"updated_at": resp.Client.UpdatedAt().Format(time.RFC3339),
	})
}

// UpdateClientHandler handles PUT /api/v1/clients/{id}
func (h *Handlers) UpdateClientHandler(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "id")
	if clientID == "" {
		http.Error(w, "client ID is required", http.StatusBadRequest)
		return
	}

	id, err := parseUUID(clientID)
	if err != nil {
		http.Error(w, "invalid client ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name   *string `json:"name"`
		Slug   *string `json:"slug"`
		Status *string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	updateReq := &usecases.UpdateClientRequest{
		ClientID: id,
		Name:     req.Name,
		Slug:     req.Slug,
	}

	if req.Status != nil {
		status := model.ClientStatus(*req.Status)
		updateReq.Status = &status
	}

	resp, err := h.updateClient.Execute(r.Context(), updateReq)
	if err != nil {
		if err == domain.ErrClientNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to update client")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         resp.Client.ID().String(),
		"agency_id":  resp.Client.AgencyID().String(),
		"name":       resp.Client.Name(),
		"slug":       resp.Client.Slug(),
		"tier":       resp.Client.Tier().String(),
		"status":     string(resp.Client.Status()),
		"created_at": resp.Client.CreatedAt().Format(time.RFC3339),
		"updated_at": resp.Client.UpdatedAt().Format(time.RFC3339),
	})
}

// AddClientMemberHandler handles POST /api/v1/clients/{id}/members
func (h *Handlers) AddClientMemberHandler(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "id")
	if clientID == "" {
		http.Error(w, "client ID is required", http.StatusBadRequest)
		return
	}

	id, err := parseUUID(clientID)
	if err != nil {
		http.Error(w, "invalid client ID", http.StatusBadRequest)
		return
	}

	var req struct {
		UserID     string  `json:"user_id"`
		Role       string  `json:"role"`
		LocationID *string `json:"location_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	userUUID, err := parseUUID(req.UserID)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	var locationUUID *uuid.UUID
	if req.LocationID != nil {
		locUUID, err := parseUUID(*req.LocationID)
		if err != nil {
			http.Error(w, "invalid location ID", http.StatusBadRequest)
			return
		}
		locationUUID = &locUUID
	}

	addReq := &usecases.AddClientMemberRequest{
		ClientID:   id,
		UserID:     userUUID,
		Role:       model.Role(req.Role),
		LocationID: locationUUID,
	}

	resp, err := h.addClientMember.Execute(r.Context(), addReq)
	if err != nil {
		if err == domain.ErrMemberAlreadyExists {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		if err == domain.ErrInvalidRole || err == domain.ErrClientSeatLimitExceeded {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to add client member")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var locationIDStr *string
	if resp.Member.LocationID() != nil {
		s := resp.Member.LocationID().String()
		locationIDStr = &s
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          resp.Member.ID().String(),
		"client_id":   resp.Member.ClientID().String(),
		"user_id":     resp.Member.UserID().String(),
		"role":        string(resp.Member.Role()),
		"location_id": locationIDStr,
		"created_at":  resp.Member.CreatedAt().Format(time.RFC3339),
		"updated_at":  resp.Member.UpdatedAt().Format(time.RFC3339),
	})
}

// ListClientMembersHandler handles GET /api/v1/clients/{id}/members
func (h *Handlers) ListClientMembersHandler(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "id")
	if clientID == "" {
		http.Error(w, "client ID is required", http.StatusBadRequest)
		return
	}

	id, err := parseUUID(clientID)
	if err != nil {
		http.Error(w, "invalid client ID", http.StatusBadRequest)
		return
	}

	locationIDStr := r.URL.Query().Get("location_id")
	var locationID *uuid.UUID
	if locationIDStr != "" {
		locID, err := parseUUID(locationIDStr)
		if err == nil {
			locationID = &locID
		}
	}

	listReq := &usecases.ListClientMembersRequest{
		ClientID:   id,
		LocationID: locationID,
	}

	resp, err := h.listClientMembers.Execute(r.Context(), listReq)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list client members")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	members := make([]map[string]interface{}, len(resp.Members))
	for i, member := range resp.Members {
		var locationIDStr *string
		if member.LocationID() != nil {
			s := member.LocationID().String()
			locationIDStr = &s
		}
		members[i] = map[string]interface{}{
			"id":          member.ID().String(),
			"client_id":   member.ClientID().String(),
			"user_id":     member.UserID().String(),
			"role":        string(member.Role()),
			"location_id": locationIDStr,
			"created_at":  member.CreatedAt().Format(time.RFC3339),
			"updated_at":  member.UpdatedAt().Format(time.RFC3339),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"members": members,
	})
}

// RemoveClientMemberHandler handles DELETE /api/v1/clients/{id}/members/{memberId}
func (h *Handlers) RemoveClientMemberHandler(w http.ResponseWriter, r *http.Request) {
	memberID := chi.URLParam(r, "memberId")
	if memberID == "" {
		http.Error(w, "member ID is required", http.StatusBadRequest)
		return
	}

	id, err := parseUUID(memberID)
	if err != nil {
		http.Error(w, "invalid member ID", http.StatusBadRequest)
		return
	}

	removeReq := &usecases.RemoveClientMemberRequest{
		MemberID: id,
	}

	resp, err := h.removeClientMember.Execute(r.Context(), removeReq)
	if err != nil {
		if err == domain.ErrClientMemberNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to remove client member")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": resp.Success,
	})
}

// CreateLocationHandler handles POST /api/v1/clients/{id}/locations
func (h *Handlers) CreateLocationHandler(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "id")
	if clientID == "" {
		http.Error(w, "client ID is required", http.StatusBadRequest)
		return
	}

	id, err := parseUUID(clientID)
	if err != nil {
		http.Error(w, "invalid client ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	createReq := &usecases.CreateLocationRequest{
		ClientID: id,
		Name:     req.Name,
	}

	resp, err := h.createLocation.Execute(r.Context(), createReq)
	if err != nil {
		if err == domain.ErrClientNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to create location")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":             resp.Location.ID().String(),
		"client_id":      resp.Location.ClientID().String(),
		"name":           resp.Location.Name(),
		"address":        resp.Location.Address(),
		"phone":          resp.Location.Phone(),
		"business_hours": resp.Location.BusinessHours(),
		"categories":     resp.Location.Categories(),
		"is_active":      resp.Location.IsActive(),
		"created_at":     resp.Location.CreatedAt().Format(time.RFC3339),
		"updated_at":     resp.Location.UpdatedAt().Format(time.RFC3339),
	})
}

// ListLocationsHandler handles GET /api/v1/clients/{id}/locations
func (h *Handlers) ListLocationsHandler(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "id")
	if clientID == "" {
		http.Error(w, "client ID is required", http.StatusBadRequest)
		return
	}

	id, err := parseUUID(clientID)
	if err != nil {
		http.Error(w, "invalid client ID", http.StatusBadRequest)
		return
	}

	listReq := &usecases.ListLocationsRequest{
		ClientID: id,
	}

	resp, err := h.listLocations.Execute(r.Context(), listReq)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list locations")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	locations := make([]map[string]interface{}, len(resp.Locations))
	for i, location := range resp.Locations {
		locations[i] = map[string]interface{}{
			"id":             location.ID().String(),
			"client_id":      location.ClientID().String(),
			"name":           location.Name(),
			"address":        location.Address(),
			"phone":          location.Phone(),
			"business_hours": location.BusinessHours(),
			"categories":     location.Categories(),
			"is_active":      location.IsActive(),
			"created_at":     location.CreatedAt().Format(time.RFC3339),
			"updated_at":     location.UpdatedAt().Format(time.RFC3339),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"locations": locations,
	})
}

// UpdateLocationHandler handles PUT /api/v1/locations/{id}
func (h *Handlers) UpdateLocationHandler(w http.ResponseWriter, r *http.Request) {
	locationID := chi.URLParam(r, "id")
	if locationID == "" {
		http.Error(w, "location ID is required", http.StatusBadRequest)
		return
	}

	id, err := parseUUID(locationID)
	if err != nil {
		http.Error(w, "invalid location ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name          *string                `json:"name"`
		Address       *map[string]interface{} `json:"address"`
		Phone         *string                 `json:"phone"`
		BusinessHours *map[string]interface{} `json:"business_hours"`
		Categories    *[]string               `json:"categories"`
		IsActive      *bool                   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	updateReq := &usecases.UpdateLocationRequest{
		LocationID:    id,
		Name:          req.Name,
		Address:       req.Address,
		Phone:         req.Phone,
		BusinessHours: req.BusinessHours,
		Categories:    req.Categories,
		IsActive:      req.IsActive,
	}

	resp, err := h.updateLocation.Execute(r.Context(), updateReq)
	if err != nil {
		if err == domain.ErrLocationNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to update location")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":             resp.Location.ID().String(),
		"client_id":      resp.Location.ClientID().String(),
		"name":           resp.Location.Name(),
		"address":        resp.Location.Address(),
		"phone":          resp.Location.Phone(),
		"business_hours": resp.Location.BusinessHours(),
		"categories":     resp.Location.Categories(),
		"is_active":      resp.Location.IsActive(),
		"created_at":     resp.Location.CreatedAt().Format(time.RFC3339),
		"updated_at":     resp.Location.UpdatedAt().Format(time.RFC3339),
	})
}

// GetSeatUsageHandler handles GET /api/v1/tenants/{id}/seat-usage
func (h *Handlers) GetSeatUsageHandler(w http.ResponseWriter, r *http.Request) {
	agencyID := chi.URLParam(r, "id")
	if agencyID == "" {
		http.Error(w, "agency ID is required", http.StatusBadRequest)
		return
	}

	id, err := parseUUID(agencyID)
	if err != nil {
		http.Error(w, "invalid agency ID", http.StatusBadRequest)
		return
	}

	getReq := &usecases.GetSeatUsageRequest{
		AgencyID: &id,
	}

	resp, err := h.getSeatUsage.Execute(r.Context(), getReq)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get seat usage")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	clientBreakdown := make([]map[string]interface{}, len(resp.Usage.ClientBreakdown))
	for i, info := range resp.Usage.ClientBreakdown {
		clientBreakdown[i] = map[string]interface{}{
			"client_id":   info.ClientID.String(),
			"client_name": info.ClientName,
			"locations":   info.Locations,
			"members":     info.Members,
			"seat_limit":  info.SeatLimit,
			"seats_used":  info.SeatsUsed,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agency_seats_used":  resp.Usage.AgencySeatsUsed,
		"agency_seats_limit": resp.Usage.AgencySeatsLimit,
		"client_seats_used":  resp.Usage.ClientSeatsUsed,
		"client_seats_limit": resp.Usage.ClientSeatsLimit,
		"total_clients":      resp.Usage.TotalClients,
		"total_locations":    resp.Usage.TotalLocations,
		"client_breakdown":   clientBreakdown,
	})
}

// parseUUID parses a UUID string
func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

