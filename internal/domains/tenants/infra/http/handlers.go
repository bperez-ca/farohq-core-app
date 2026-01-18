package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"farohq-core-app/internal/domains/tenants/app/usecases"
	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/model"
	tenants_outbound "farohq-core-app/internal/domains/tenants/domain/ports/outbound"
	users_domain "farohq-core-app/internal/domains/users/domain"
	users_outbound "farohq-core-app/internal/domains/users/domain/ports/outbound"

	// Brand domain for fetching branding info
	brand_outbound "farohq-core-app/internal/domains/brand/domain/ports/outbound"
)

// Handlers provides HTTP handlers for the tenants domain
type Handlers struct {
	logger             zerolog.Logger
	createTenant       *usecases.CreateTenant
	getTenant          *usecases.GetTenant
	updateTenant       *usecases.UpdateTenant
	inviteMember       *usecases.InviteMember
	acceptInvite       *usecases.AcceptInvite
	listInvites        *usecases.ListInvites
	findInvitesByEmail *usecases.FindInvitesByEmail
	revokeInvite       *usecases.RevokeInvite
	deleteInvite       *usecases.DeleteInvite
	listMembers        *usecases.ListMembers
	removeMember       *usecases.RemoveMember
	listRoles          *usecases.ListRoles
	createClient       *usecases.CreateClient
	listClients        *usecases.ListClients
	getClient          *usecases.GetClient
	updateClient       *usecases.UpdateClient
	addClientMember    *usecases.AddClientMember
	listClientMembers  *usecases.ListClientMembers
	removeClientMember *usecases.RemoveClientMember
	createLocation     *usecases.CreateLocation
	listLocations      *usecases.ListLocations
	updateLocation     *usecases.UpdateLocation
	getSeatUsage       *usecases.GetSeatUsage
	listTenantsByUser  *usecases.ListTenantsByUser
	validateSlug       *usecases.ValidateSlug
	onboardTenant      *usecases.OnboardTenant
	userRepo           users_outbound.UserRepository
	inviteRepo         tenants_outbound.InviteRepository
	tenantRepo         tenants_outbound.TenantRepository
	brandRepo          brand_outbound.BrandRepository // For fetching branding info in invite details
}

// NewHandlers creates new tenants HTTP handlers
func NewHandlers(
	logger zerolog.Logger,
	createTenant *usecases.CreateTenant,
	onboardTenant *usecases.OnboardTenant,
	getTenant *usecases.GetTenant,
	updateTenant *usecases.UpdateTenant,
	inviteMember *usecases.InviteMember,
	acceptInvite *usecases.AcceptInvite,
	listInvites *usecases.ListInvites,
	findInvitesByEmail *usecases.FindInvitesByEmail,
	revokeInvite *usecases.RevokeInvite,
	deleteInvite *usecases.DeleteInvite,
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
	listTenantsByUser *usecases.ListTenantsByUser,
	validateSlug *usecases.ValidateSlug,
	userRepo users_outbound.UserRepository,
	inviteRepo tenants_outbound.InviteRepository,
	tenantRepo tenants_outbound.TenantRepository,
	brandRepo brand_outbound.BrandRepository,
) *Handlers {
	return &Handlers{
		logger:             logger,
		createTenant:       createTenant,
		onboardTenant:      onboardTenant,
		getTenant:          getTenant,
		updateTenant:       updateTenant,
		inviteMember:       inviteMember,
		acceptInvite:       acceptInvite,
		listInvites:        listInvites,
		findInvitesByEmail: findInvitesByEmail,
		revokeInvite:       revokeInvite,
		deleteInvite:       deleteInvite,
		listMembers:        listMembers,
		removeMember:       removeMember,
		listRoles:          listRoles,
		createClient:       createClient,
		listClients:        listClients,
		getClient:          getClient,
		updateClient:       updateClient,
		addClientMember:    addClientMember,
		listClientMembers:  listClientMembers,
		removeClientMember: removeClientMember,
		createLocation:     createLocation,
		listLocations:      listLocations,
		updateLocation:     updateLocation,
		getSeatUsage:       getSeatUsage,
		listTenantsByUser:  listTenantsByUser,
		validateSlug:       validateSlug,
		userRepo:           userRepo,
		inviteRepo:         inviteRepo,
		tenantRepo:         tenantRepo,
		brandRepo:          brandRepo,
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

	// Get Clerk user ID from context (set by auth middleware)
	clerkUserID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "user ID required", http.StatusUnauthorized)
		return
	}

	// Look up user by Clerk user ID to get database UUID
	user, err := h.userRepo.FindByClerkUserID(r.Context(), clerkUserID)
	if err != nil {
		h.logger.Error().Err(err).Str("clerk_user_id", clerkUserID).Msg("Failed to find user by Clerk user ID")
		http.Error(w, "User not found", http.StatusNotFound)
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
		CreatedBy: user.ID(),
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
		if err == domain.ErrPendingInviteExists || err == domain.ErrInviteAlreadyAccepted {
			http.Error(w, err.Error(), http.StatusConflict)
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

// AcceptInviteHandler handles POST /api/v1/invites/accept
func (h *Handlers) AcceptInviteHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token  string `json:"token"`
		UserID string `json:"user_id"` // Optional - will use authenticated user if not provided
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Token == "" {
		http.Error(w, "token is required", http.StatusBadRequest)
		return
	}

	// Get authenticated user from context (set by RequireAuth middleware)
	clerkUserID, ok := r.Context().Value("user_id").(string)
	if !ok || clerkUserID == "" {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	// Look up user by Clerk user ID to get database UUID
	user, err := h.userRepo.FindByClerkUserID(r.Context(), clerkUserID)
	if err != nil {
		h.logger.Error().
			Str("clerk_user_id", clerkUserID).
			Err(err).
			Msg("Failed to find user by Clerk user ID for invite acceptance")
		http.Error(w, "User not found. Please ensure your account is synced.", http.StatusNotFound)
		return
	}

	// Find invite by token to validate email match
	invite, err := h.inviteRepo.FindByToken(r.Context(), req.Token)
	if err != nil {
		if err == domain.ErrInviteNotFound {
			http.Error(w, "Invitation not found", http.StatusNotFound)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to find invite by token")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Validate that the authenticated user's email matches the invite email
	if user.Email() != invite.Email() {
		h.logger.Warn().
			Str("user_email", user.Email()).
			Str("invite_email", invite.Email()).
			Str("invite_token", req.Token).
			Msg("User email does not match invite email")
		http.Error(w, "This invitation was sent to a different email address. Please sign in with the email that received the invitation.", http.StatusForbidden)
		return
	}

	// Use the database UUID from the authenticated user
	userID := user.ID()

	acceptReq := &usecases.AcceptInviteRequest{
		Token:  req.Token,
		UserID: userID,
	}

	resp, err := h.acceptInvite.Execute(r.Context(), acceptReq)
	if err != nil {
		if err == domain.ErrInviteNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err == domain.ErrInviteExpired {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err == domain.ErrInviteRevoked {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err == domain.ErrInviteAlreadyAccepted {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to accept invite")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"member": map[string]interface{}{
			"id":         resp.Member.ID().String(),
			"tenant_id":  resp.Member.TenantID().String(),
			"user_id":    resp.Member.UserID().String(),
			"role":       string(resp.Member.Role()),
			"created_at": resp.Member.CreatedAt().Format(time.RFC3339),
			"updated_at": resp.Member.UpdatedAt().Format(time.RFC3339),
		},
		"message": "Invitation accepted successfully",
	})
}

// GetInviteByTokenHandler handles GET /api/v1/invites/{token} (public endpoint for branding)
func (h *Handlers) GetInviteByTokenHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		http.Error(w, "token is required", http.StatusBadRequest)
		return
	}

	// Find invite by token
	invite, err := h.inviteRepo.FindByToken(r.Context(), token)
	if err != nil {
		if err == domain.ErrInviteNotFound {
			http.Error(w, "Invitation not found", http.StatusNotFound)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to find invite by token")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check invite status
	status := "pending"
	if invite.IsAccepted() {
		status = "accepted"
	} else if invite.IsRevoked() {
		status = "revoked"
	} else if invite.IsExpired() {
		status = "expired"
	}

	// Get tenant information for branding
	tenant, err := h.tenantRepo.FindByID(r.Context(), invite.TenantID())
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("tenant_id", invite.TenantID().String()).
			Msg("Failed to find tenant for invite")
		// Still return invite info even if tenant lookup fails
	}

	response := map[string]interface{}{
		"id":         invite.ID().String(),
		"email":      invite.Email(),
		"role":       string(invite.Role()),
		"expires_at": invite.ExpiresAt().Format(time.RFC3339),
		"created_at": invite.CreatedAt().Format(time.RFC3339),
		"status":     status,
	}

	// Add tenant info if available
	if tenant != nil {
		response["tenant"] = map[string]interface{}{
			"id":   tenant.ID().String(),
			"name": tenant.Name(),
			"slug": tenant.Slug(),
		}

		// Fetch branding information for white-label support
		if h.brandRepo != nil {
			branding, err := h.brandRepo.FindByAgencyID(r.Context(), tenant.ID())
			if err == nil && branding != nil {
				// Get tier to apply tier-based rules
				tier := tenant.Tier()
				hidePoweredBy := branding.HidePoweredBy()

				// Apply tier-based rules: only Growth+ can hide powered by
				if tier != nil && !model.TierCanHidePoweredBy(tier) {
					hidePoweredBy = false
				}

				response["branding"] = map[string]interface{}{
					"logo_url":        branding.LogoURL(),
					"favicon_url":     branding.FaviconURL(),
					"primary_color":   branding.PrimaryColor(),
					"secondary_color": branding.SecondaryColor(),
					"hide_powered_by": hidePoweredBy,
					"theme_json":      branding.ThemeJSON(),
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// FindInvitesByEmailHandler handles GET /api/v1/invites/by-email?email={email}
func (h *Handlers) FindInvitesByEmailHandler(w http.ResponseWriter, r *http.Request) {
	// #region agent log
	rawURL := r.URL.String()
	logData := map[string]interface{}{
		"location":     "handlers.go:546",
		"message":      "FindInvitesByEmailHandler: raw URL",
		"raw_url":      rawURL,
		"timestamp":    time.Now().UnixMilli(),
		"sessionId":    "debug-session",
		"runId":        "run1",
		"hypothesisId": "C",
	}
	if logBytes, err := json.Marshal(logData); err == nil {
		if f, err := os.OpenFile("/Users/bperez/Projects/farohq-core-app/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			f.Write(append(logBytes, '\n'))
			f.Close()
		}
	}
	// #endregion

	// Get email from query parameter
	// r.URL.Query().Get() automatically URL-decodes %40 to @
	email := r.URL.Query().Get("email")

	// #region agent log
	logData = map[string]interface{}{
		"location":     "handlers.go:551",
		"message":      "FindInvitesByEmailHandler: extracted email from URL",
		"raw_email":    email,
		"email_length": len(email),
		"timestamp":    time.Now().UnixMilli(),
		"sessionId":    "debug-session",
		"runId":        "run1",
		"hypothesisId": "C",
	}
	if logBytes, err := json.Marshal(logData); err == nil {
		if f, err := os.OpenFile("/Users/bperez/Projects/farohq-core-app/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			f.Write(append(logBytes, '\n'))
			f.Close()
		}
	}
	// #endregion

	if email == "" {
		http.Error(w, "email query parameter is required", http.StatusBadRequest)
		return
	}

	// Execute use case
	findReq := &usecases.FindInvitesByEmailRequest{
		Email: email,
	}

	resp, err := h.findInvitesByEmail.Execute(r.Context(), findReq)

	// #region agent log
	invitesCount := 0
	if resp != nil {
		invitesCount = len(resp.Invites)
	}
	logData = map[string]interface{}{
		"location":      "handlers.go:567",
		"message":       "FindInvitesByEmailHandler: after Execute",
		"error":         fmt.Sprintf("%v", err),
		"invites_count": invitesCount,
		"timestamp":     time.Now().UnixMilli(),
		"sessionId":     "debug-session",
		"runId":         "run1",
		"hypothesisId":  "E",
	}
	if logBytes, err2 := json.Marshal(logData); err2 == nil {
		if f, err := os.OpenFile("/Users/bperez/Projects/farohq-core-app/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			f.Write(append(logBytes, '\n'))
			f.Close()
		}
	}
	// #endregion

	if err != nil {
		if err == domain.ErrInvalidEmail {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.logger.Error().Err(err).Str("email", email).Msg("Failed to find invites by email")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// #region agent log
	if len(resp.Invites) == 0 {
		logData = map[string]interface{}{
			"location":     "handlers.go:580",
			"message":      "FindInvitesByEmailHandler: no invites found, returning empty array",
			"timestamp":    time.Now().UnixMilli(),
			"sessionId":    "debug-session",
			"runId":        "run1",
			"hypothesisId": "E",
		}
		if logBytes, err := json.Marshal(logData); err == nil {
			os.WriteFile("/Users/bperez/Projects/farohq-core-app/.cursor/debug.log", append(logBytes, '\n'), 0644)
		}
	}
	// #endregion

	// Build response with branding info for each invite
	invites := make([]map[string]interface{}, len(resp.Invites))
	for i, invite := range resp.Invites {
		// Get tenant information for branding
		tenant, err := h.tenantRepo.FindByID(r.Context(), invite.TenantID())
		if err != nil {
			h.logger.Warn().
				Err(err).
				Str("tenant_id", invite.TenantID().String()).
				Msg("Failed to find tenant for invite in FindInvitesByEmailHandler")
		}

		inviteMap := map[string]interface{}{
			"id":         invite.ID().String(),
			"email":      invite.Email(),
			"role":       string(invite.Role()),
			"token":      invite.Token(),
			"expires_at": invite.ExpiresAt().Format(time.RFC3339),
			"created_at": invite.CreatedAt().Format(time.RFC3339),
			"status":     "pending",
		}

		// Add tenant info if available
		if tenant != nil {
			inviteMap["tenant"] = map[string]interface{}{
				"id":   tenant.ID().String(),
				"name": tenant.Name(),
				"slug": tenant.Slug(),
			}

			// Fetch branding information for white-label support
			if h.brandRepo != nil {
				branding, err := h.brandRepo.FindByAgencyID(r.Context(), tenant.ID())
				if err == nil && branding != nil {
					// Get tier to apply tier-based rules
					tier := tenant.Tier()
					hidePoweredBy := branding.HidePoweredBy()

					// Apply tier-based rules: only Growth+ can hide powered by
					if tier != nil && !model.TierCanHidePoweredBy(tier) {
						hidePoweredBy = false
					}

					inviteMap["branding"] = map[string]interface{}{
						"logo_url":        branding.LogoURL(),
						"favicon_url":     branding.FaviconURL(),
						"primary_color":   branding.PrimaryColor(),
						"secondary_color": branding.SecondaryColor(),
						"hide_powered_by": hidePoweredBy,
						"theme_json":      branding.ThemeJSON(),
					}
				}
			}
		}

		invites[i] = inviteMap
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"invites": invites,
	})
}

// ListInvitesHandler handles GET /api/v1/tenants/{id}/invites
func (h *Handlers) ListInvitesHandler(w http.ResponseWriter, r *http.Request) {
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

	listReq := &usecases.ListInvitesRequest{
		TenantID: id,
	}

	resp, err := h.listInvites.Execute(r.Context(), listReq)
	if err != nil {
		if err == domain.ErrTenantNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to list invites")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	invites := make([]map[string]interface{}, len(resp.Invites))
	for i, invite := range resp.Invites {
		inviteMap := map[string]interface{}{
			"id":         invite.ID().String(),
			"email":      invite.Email(),
			"role":       string(invite.Role()),
			"expires_at": invite.ExpiresAt().Format(time.RFC3339),
			"created_at": invite.CreatedAt().Format(time.RFC3339),
			"created_by": invite.CreatedBy().String(),
		}

		if invite.AcceptedAt() != nil {
			inviteMap["accepted_at"] = invite.AcceptedAt().Format(time.RFC3339)
			inviteMap["status"] = "accepted"
		} else if invite.RevokedAt() != nil {
			inviteMap["revoked_at"] = invite.RevokedAt().Format(time.RFC3339)
			inviteMap["status"] = "revoked"
		} else if invite.IsExpired() {
			inviteMap["status"] = "expired"
		} else {
			inviteMap["status"] = "pending"
		}

		invites[i] = inviteMap
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"invites": invites,
	})
}

// RevokeInviteHandler handles DELETE /api/v1/tenants/{id}/invites/{invite_id}
// If ?permanent=true query parameter is present, it deletes the invite permanently instead of revoking
func (h *Handlers) RevokeInviteHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "id")
	if tenantID == "" {
		http.Error(w, "tenant ID is required", http.StatusBadRequest)
		return
	}

	inviteIDStr := chi.URLParam(r, "invite_id")
	if inviteIDStr == "" {
		http.Error(w, "invite ID is required", http.StatusBadRequest)
		return
	}

	tenantUUID, err := parseUUID(tenantID)
	if err != nil {
		http.Error(w, "invalid tenant ID", http.StatusBadRequest)
		return
	}

	inviteUUID, err := parseUUID(inviteIDStr)
	if err != nil {
		http.Error(w, "invalid invite ID", http.StatusBadRequest)
		return
	}

	// Check if permanent delete is requested
	if r.URL.Query().Get("permanent") == "true" {
		// Permanently delete the invite
		deleteReq := &usecases.DeleteInviteRequest{
			InviteID: inviteUUID,
			TenantID: tenantUUID,
		}

		resp, err := h.deleteInvite.Execute(r.Context(), deleteReq)
		if err != nil {
			if err == domain.ErrTenantNotFound || err == domain.ErrInviteNotFound {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			h.logger.Error().Err(err).Msg("Failed to delete invite")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": resp.Success,
			"message": "Invitation deleted permanently",
		})
		return
	}

	// Otherwise, revoke the invite (soft delete)
	revokeReq := &usecases.RevokeInviteRequest{
		InviteID: inviteUUID,
		TenantID: tenantUUID,
	}

	resp, err := h.revokeInvite.Execute(r.Context(), revokeReq)
	if err != nil {
		if err == domain.ErrTenantNotFound || err == domain.ErrInviteNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err == domain.ErrInviteAlreadyAccepted {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to revoke invite")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	inviteMap := map[string]interface{}{
		"id":         resp.Invite.ID().String(),
		"email":      resp.Invite.Email(),
		"role":       string(resp.Invite.Role()),
		"expires_at": resp.Invite.ExpiresAt().Format(time.RFC3339),
		"created_at": resp.Invite.CreatedAt().Format(time.RFC3339),
		"revoked_at": resp.Invite.RevokedAt().Format(time.RFC3339),
		"status":     "revoked",
	}

	if resp.Invite.AcceptedAt() != nil {
		inviteMap["accepted_at"] = resp.Invite.AcceptedAt().Format(time.RFC3339)
	}

	json.NewEncoder(w).Encode(inviteMap)
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
		Name          *string                 `json:"name"`
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

// OnboardTenantHandler handles POST /api/v1/tenants/onboard
func (h *Handlers) OnboardTenantHandler(w http.ResponseWriter, r *http.Request) {
	// Get Clerk user ID from context (set by auth middleware)
	clerkUserID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "user ID required", http.StatusUnauthorized)
		return
	}

	// Look up user by Clerk user ID to get database UUID
	user, err := h.userRepo.FindByClerkUserID(r.Context(), clerkUserID)
	if err != nil {
		h.logger.Error().Err(err).Str("clerk_user_id", clerkUserID).Msg("Failed to find user by Clerk user ID")
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	userID := user.ID()

	var req struct {
		Name         string `json:"name"`
		Slug         string `json:"slug"`
		Website      string `json:"website"`
		PrimaryColor string `json:"primary_color"`
		LogoURL      string `json:"logo_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Default to Starter tier for new tenants
	tier := model.TierStarter
	onboardReq := &usecases.OnboardTenantRequest{
		Name:            req.Name,
		Slug:            req.Slug,
		Tier:            &tier,
		AgencySeatLimit: 0, // Default seat limit
		UserID:          userID,
	}

	resp, err := h.onboardTenant.Execute(r.Context(), onboardReq)
	if err != nil {
		if err == domain.ErrTenantAlreadyExists {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		if err == domain.ErrInvalidTenantName || err == domain.ErrInvalidTenantSlug {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to onboard tenant")
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

// ListTenantsByUserHandler handles GET /api/v1/tenants/my-orgs
func (h *Handlers) ListTenantsByUserHandler(w http.ResponseWriter, r *http.Request) {
	// #region agent log
	logFile, _ := os.OpenFile("/Users/bperez/Projects/farohq-core-app/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	json.NewEncoder(logFile).Encode(map[string]interface{}{"timestamp": time.Now().UnixMilli(), "location": "handlers.go:1112", "message": "ListTenantsByUserHandler called", "hypothesisId": "H3", "sessionId": "debug-session", "runId": "run1", "data": map[string]interface{}{"path": r.URL.Path, "method": r.Method}})
	logFile.Close()
	// #endregion
	// Get Clerk user ID from context (set by auth middleware)
	clerkUserID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "user ID required", http.StatusUnauthorized)
		return
	}

	// Look up user by Clerk user ID to get database UUID
	user, err := h.userRepo.FindByClerkUserID(r.Context(), clerkUserID)
	if err != nil {
		// If user doesn't exist (e.g., during invite acceptance flow), return empty array
		// This is expected behavior when a new user signs up via Clerk but hasn't been synced yet
		if err == users_domain.ErrUserNotFound {
			h.logger.Debug().Str("clerk_user_id", clerkUserID).Msg("User not found in database, returning empty orgs list")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"count": 0,
				"orgs":  []map[string]interface{}{},
			})
			return
		}
		h.logger.Error().Err(err).Str("clerk_user_id", clerkUserID).Msg("Failed to find user by Clerk user ID")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	listReq := &usecases.ListTenantsByUserRequest{
		UserID: user.ID(),
	}

	resp, err := h.listTenantsByUser.Execute(r.Context(), listReq)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list tenants by user")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tenants := make([]map[string]interface{}, len(resp.Tenants))
	for i, tenantWithRole := range resp.Tenants {
		tenants[i] = map[string]interface{}{
			"id":         tenantWithRole.Tenant.ID().String(),
			"name":       tenantWithRole.Tenant.Name(),
			"slug":       tenantWithRole.Tenant.Slug(),
			"role":       string(tenantWithRole.Role),
			"created_at": tenantWithRole.Tenant.CreatedAt().Format(time.RFC3339),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count": len(tenants),
		"orgs":  tenants,
	})
}

// ValidateSlugHandler handles GET /api/v1/tenants/validate-slug?slug=xxx
func (h *Handlers) ValidateSlugHandler(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	if slug == "" {
		http.Error(w, "slug parameter is required", http.StatusBadRequest)
		return
	}

	validateReq := &usecases.ValidateSlugRequest{
		Slug: slug,
	}

	resp, err := h.validateSlug.Execute(r.Context(), validateReq)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to validate slug")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"available": resp.Available,
		"slug":      resp.Slug,
	})
}

// parseUUID parses a UUID string
func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
