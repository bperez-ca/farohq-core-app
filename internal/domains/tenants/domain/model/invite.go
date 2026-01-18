package model

import (
	"time"

	"github.com/google/uuid"
)

// Invite represents an invitation to join a tenant
type Invite struct {
	id         uuid.UUID
	tenantID   uuid.UUID
	email      string
	role       Role
	token      string
	expiresAt  time.Time
	acceptedAt *time.Time
	revokedAt  *time.Time
	createdAt  time.Time
	createdBy  uuid.UUID
}

// NewInvite creates a new invite entity
func NewInvite(tenantID uuid.UUID, email string, role Role, token string, createdBy uuid.UUID, expiresIn time.Duration) *Invite {
	now := time.Now()
	return &Invite{
		id:        uuid.New(),
		tenantID:  tenantID,
		email:     email,
		role:      role,
		token:     token,
		expiresAt: now.Add(expiresIn),
		createdAt: now,
		createdBy: createdBy,
	}
}

// NewInviteWithID creates an invite entity with a specific ID (used for reconstruction from database)
func NewInviteWithID(id, tenantID uuid.UUID, email string, role Role, token string, expiresAt time.Time, acceptedAt *time.Time, revokedAt *time.Time, createdAt time.Time, createdBy uuid.UUID) *Invite {
	return &Invite{
		id:         id,
		tenantID:   tenantID,
		email:      email,
		role:       role,
		token:      token,
		expiresAt:  expiresAt,
		acceptedAt: acceptedAt,
		revokedAt:  revokedAt,
		createdAt:  createdAt,
		createdBy:  createdBy,
	}
}

// ID returns the invite ID
func (i *Invite) ID() uuid.UUID {
	return i.id
}

// TenantID returns the tenant ID
func (i *Invite) TenantID() uuid.UUID {
	return i.tenantID
}

// Email returns the invite email
func (i *Invite) Email() string {
	return i.email
}

// Role returns the invite role
func (i *Invite) Role() Role {
	return i.role
}

// Token returns the invite token
func (i *Invite) Token() string {
	return i.token
}

// ExpiresAt returns the expiration timestamp
func (i *Invite) ExpiresAt() time.Time {
	return i.expiresAt
}

// AcceptedAt returns the acceptance timestamp (nil if not accepted)
func (i *Invite) AcceptedAt() *time.Time {
	return i.acceptedAt
}

// CreatedAt returns the creation timestamp
func (i *Invite) CreatedAt() time.Time {
	return i.createdAt
}

// CreatedBy returns the creator user ID
func (i *Invite) CreatedBy() uuid.UUID {
	return i.createdBy
}

// IsExpired checks if the invite has expired
func (i *Invite) IsExpired() bool {
	return time.Now().After(i.expiresAt)
}

// IsAccepted checks if the invite has been accepted
func (i *Invite) IsAccepted() bool {
	return i.acceptedAt != nil
}

// Accept marks the invite as accepted
func (i *Invite) Accept() {
	now := time.Now()
	i.acceptedAt = &now
}

// RevokedAt returns the revocation timestamp (nil if not revoked)
func (i *Invite) RevokedAt() *time.Time {
	return i.revokedAt
}

// IsRevoked checks if the invite has been revoked
func (i *Invite) IsRevoked() bool {
	return i.revokedAt != nil
}

// Revoke marks the invite as revoked
func (i *Invite) Revoke() {
	now := time.Now()
	i.revokedAt = &now
}

