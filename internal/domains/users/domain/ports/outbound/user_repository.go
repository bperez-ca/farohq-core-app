package outbound

import (
	"context"

	"farohq-core-app/internal/domains/users/domain/model"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	FindByClerkUserID(ctx context.Context, clerkUserID string) (*model.User, error)
	Save(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
}
