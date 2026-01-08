package outbound

import (
	"context"

	"farohq-core-app/internal/domains/tenants/domain/model"

	"github.com/google/uuid"
)

// ClientMemberRepository defines the interface for client member persistence
type ClientMemberRepository interface {
	// Save saves or updates a client member
	Save(ctx context.Context, member *model.ClientMember) error

	// FindByID finds a client member by ID
	FindByID(ctx context.Context, id uuid.UUID) (*model.ClientMember, error)

	// FindByClientAndUser finds a client member by client ID and user ID
	FindByClientAndUser(ctx context.Context, clientID, userID uuid.UUID, locationID *uuid.UUID) (*model.ClientMember, error)

	// ListByClient lists all members for a client
	ListByClient(ctx context.Context, clientID uuid.UUID, locationID *uuid.UUID) ([]*model.ClientMember, error)

	// CountByClient counts members for a client (excluding soft-deleted)
	CountByClient(ctx context.Context, clientID uuid.UUID) (int, error)

	// CountByClientAndLocation counts members for a client and location (excluding soft-deleted)
	CountByClientAndLocation(ctx context.Context, clientID uuid.UUID, locationID uuid.UUID) (int, error)
}

