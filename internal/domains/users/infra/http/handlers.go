package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"farohq-core-app/internal/domains/users/domain/ports/inbound"
)

// Handlers provides HTTP handlers for the users domain
type Handlers struct {
	logger    zerolog.Logger
	syncUser  inbound.SyncUser
}

// NewHandlers creates new user HTTP handlers
func NewHandlers(
	logger zerolog.Logger,
	syncUser inbound.SyncUser,
) *Handlers {
	return &Handlers{
		logger:   logger,
		syncUser: syncUser,
	}
}

// SyncUserHandler handles POST /api/v1/users/sync
func (h *Handlers) SyncUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req inbound.SyncUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to decode sync user request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.ClerkUserID == "" {
		http.Error(w, "clerk_user_id is required", http.StatusBadRequest)
		return
	}

	resp, err := h.syncUser.Execute(r.Context(), &req)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to sync user")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":             resp.User.ID().String(),
		"clerk_user_id":  resp.User.ClerkUserID(),
		"email":          resp.User.Email(),
		"first_name":     resp.User.FirstName(),
		"last_name":      resp.User.LastName(),
		"full_name":      resp.User.FullName(),
		"image_url":      resp.User.ImageURL(),
		"phone_numbers":  resp.User.PhoneNumbers(),
		"created_at":     resp.User.CreatedAt().Format(time.RFC3339),
		"updated_at":     resp.User.UpdatedAt().Format(time.RFC3339),
		"last_sign_in_at": func() *string {
			if resp.User.LastSignInAt() != nil {
				s := resp.User.LastSignInAt().Format(time.RFC3339)
				return &s
			}
			return nil
		}(),
	})
}
