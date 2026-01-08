package model

import (
	"time"

	"github.com/google/uuid"
)

// ClientMember represents a member of a client
type ClientMember struct {
	id         uuid.UUID
	clientID   uuid.UUID
	userID     uuid.UUID
	role       Role
	locationID *uuid.UUID
	createdAt  time.Time
	updatedAt  time.Time
	deletedAt  *time.Time
}

// NewClientMember creates a new client member entity
func NewClientMember(clientID, userID uuid.UUID, role Role, locationID *uuid.UUID) *ClientMember {
	now := time.Now()
	return &ClientMember{
		id:         uuid.New(),
		clientID:   clientID,
		userID:     userID,
		role:       role,
		locationID: locationID,
		createdAt:  now,
		updatedAt:  now,
		deletedAt:  nil,
	}
}

// NewClientMemberWithID creates a client member entity with a specific ID
func NewClientMemberWithID(id, clientID, userID uuid.UUID, role Role, locationID *uuid.UUID, createdAt, updatedAt time.Time, deletedAt *time.Time) *ClientMember {
	return &ClientMember{
		id:         id,
		clientID:   clientID,
		userID:     userID,
		role:       role,
		locationID: locationID,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
		deletedAt:  deletedAt,
	}
}

// ID returns the member ID
func (m *ClientMember) ID() uuid.UUID {
	return m.id
}

// ClientID returns the client ID
func (m *ClientMember) ClientID() uuid.UUID {
	return m.clientID
}

// UserID returns the user ID
func (m *ClientMember) UserID() uuid.UUID {
	return m.userID
}

// Role returns the member role
func (m *ClientMember) Role() Role {
	return m.role
}

// LocationID returns the location ID (if assigned to a specific location)
func (m *ClientMember) LocationID() *uuid.UUID {
	return m.locationID
}

// CreatedAt returns the creation timestamp
func (m *ClientMember) CreatedAt() time.Time {
	return m.createdAt
}

// UpdatedAt returns the update timestamp
func (m *ClientMember) UpdatedAt() time.Time {
	return m.updatedAt
}

// DeletedAt returns the deletion timestamp
func (m *ClientMember) DeletedAt() *time.Time {
	return m.deletedAt
}

// SetRole sets the member role
func (m *ClientMember) SetRole(role Role) {
	m.role = role
	m.updatedAt = time.Now()
}

// Delete marks the member as deleted (soft delete)
func (m *ClientMember) Delete() {
	now := time.Now()
	m.deletedAt = &now
	m.updatedAt = now
}

// IsDeleted checks if the member is deleted
func (m *ClientMember) IsDeleted() bool {
	return m.deletedAt != nil
}

