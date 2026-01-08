package model

import (
	"time"

	"github.com/google/uuid"
)

// TenantMember represents a member of a tenant (agency)
type TenantMember struct {
	id        uuid.UUID
	tenantID  uuid.UUID
	userID    uuid.UUID
	role      Role
	clientID  *uuid.UUID
	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time
}

// Role represents a user role
type Role string

const (
	RoleOwner       Role = "owner"
	RoleAdmin       Role = "admin"
	RoleStaff       Role = "staff"
	RoleViewer      Role = "viewer"
	RoleClientViewer Role = "client_viewer"
)

// NewTenantMember creates a new tenant member entity
func NewTenantMember(tenantID, userID uuid.UUID, role Role) *TenantMember {
	now := time.Now()
	return &TenantMember{
		id:        uuid.New(),
		tenantID:  tenantID,
		userID:    userID,
		role:      role,
		clientID:  nil,
		createdAt: now,
		updatedAt: now,
		deletedAt: nil,
	}
}

// NewTenantMemberWithID creates a tenant member entity with a specific ID
func NewTenantMemberWithID(id, tenantID, userID uuid.UUID, role Role, clientID *uuid.UUID, createdAt, updatedAt time.Time, deletedAt *time.Time) *TenantMember {
	return &TenantMember{
		id:        id,
		tenantID:  tenantID,
		userID:    userID,
		role:      role,
		clientID:  clientID,
		createdAt: createdAt,
		updatedAt: updatedAt,
		deletedAt: deletedAt,
	}
}

// ID returns the member ID
func (m *TenantMember) ID() uuid.UUID {
	return m.id
}

// TenantID returns the tenant ID
func (m *TenantMember) TenantID() uuid.UUID {
	return m.tenantID
}

// UserID returns the user ID
func (m *TenantMember) UserID() uuid.UUID {
	return m.userID
}

// Role returns the member role
func (m *TenantMember) Role() Role {
	return m.role
}

// ClientID returns the client ID (if assigned to a specific client)
func (m *TenantMember) ClientID() *uuid.UUID {
	return m.clientID
}

// CreatedAt returns the creation timestamp
func (m *TenantMember) CreatedAt() time.Time {
	return m.createdAt
}

// UpdatedAt returns the update timestamp
func (m *TenantMember) UpdatedAt() time.Time {
	return m.updatedAt
}

// DeletedAt returns the deletion timestamp
func (m *TenantMember) DeletedAt() *time.Time {
	return m.deletedAt
}

// SetRole sets the member role
func (m *TenantMember) SetRole(role Role) {
	m.role = role
	m.updatedAt = time.Now()
}

// Delete marks the member as deleted (soft delete)
func (m *TenantMember) Delete() {
	now := time.Now()
	m.deletedAt = &now
	m.updatedAt = now
}

// IsDeleted checks if the member is deleted
func (m *TenantMember) IsDeleted() bool {
	return m.deletedAt != nil
}

