package inbound

import (
	"context"

	"farohq-core-app/internal/domains/users/domain/model"
)

// SyncUserRequest represents the request to sync a user from Clerk
type SyncUserRequest struct {
	ClerkUserID  string   `json:"clerk_user_id"`
	Email        string   `json:"email"`
	FirstName    string   `json:"first_name"`
	LastName     string   `json:"last_name"`
	FullName     string   `json:"full_name"`
	ImageURL     string   `json:"image_url"`
	PhoneNumbers []string `json:"phone_numbers"`
	LastSignInAt *int64   `json:"last_sign_in_at"` // Unix timestamp in seconds
}

// SyncUserResponse represents the response from syncing a user
type SyncUserResponse struct {
	User *model.User
}

// SyncUser defines the interface for syncing a user from Clerk
type SyncUser interface {
	Execute(ctx context.Context, req *SyncUserRequest) (*SyncUserResponse, error)
}
