package usecases

import (
	"context"
	"time"

	"farohq-core-app/internal/domains/users/domain/model"
	"farohq-core-app/internal/domains/users/domain/ports/inbound"
	"farohq-core-app/internal/domains/users/domain/ports/outbound"
)

// SyncUserUseCase implements the SyncUser use case
type SyncUserUseCase struct {
	userRepo outbound.UserRepository
}

// NewSyncUser creates a new sync user use case
func NewSyncUser(userRepo outbound.UserRepository) inbound.SyncUser {
	return &SyncUserUseCase{
		userRepo: userRepo,
	}
}

// Execute syncs a user from Clerk data (creates if not exists, updates if exists)
func (uc *SyncUserUseCase) Execute(ctx context.Context, req *inbound.SyncUserRequest) (*inbound.SyncUserResponse, error) {
	// Try to find existing user by Clerk user ID
	existingUser, err := uc.userRepo.FindByClerkUserID(ctx, req.ClerkUserID)
	
	var lastSignInAt *time.Time
	if req.LastSignInAt != nil {
		t := time.Unix(*req.LastSignInAt, 0)
		lastSignInAt = &t
	}

	if err != nil {
		// User doesn't exist, create new one
		newUser := model.NewUser(
			req.ClerkUserID,
			req.Email,
			req.FirstName,
			req.LastName,
			req.FullName,
			req.ImageURL,
			req.PhoneNumbers,
		)
		
		// Set last sign in if provided
		if lastSignInAt != nil {
			newUser.UpdateFromClerk(
				req.Email,
				req.FirstName,
				req.LastName,
				req.FullName,
				req.ImageURL,
				req.PhoneNumbers,
				lastSignInAt,
			)
		}

		if err := uc.userRepo.Save(ctx, newUser); err != nil {
			return nil, err
		}

		return &inbound.SyncUserResponse{
			User: newUser,
		}, nil
	}

	// User exists, update it
	existingUser.UpdateFromClerk(
		req.Email,
		req.FirstName,
		req.LastName,
		req.FullName,
		req.ImageURL,
		req.PhoneNumbers,
		lastSignInAt,
	)

	if err := uc.userRepo.Update(ctx, existingUser); err != nil {
		return nil, err
	}

	return &inbound.SyncUserResponse{
		User: existingUser,
	}, nil
}
