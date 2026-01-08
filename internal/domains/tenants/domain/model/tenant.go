package model

import (
	"time"

	"github.com/google/uuid"
)

// Tenant represents a tenant (agency) in the system
// Note: This references the existing agencies table
type Tenant struct {
	id              uuid.UUID
	name            string
	slug            string
	status          TenantStatus
	tier            *Tier
	agencySeatLimit int
	createdAt       time.Time
	updatedAt       time.Time
	deletedAt       *time.Time
}

// TenantStatus represents the status of a tenant
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusInactive  TenantStatus = "inactive"
	TenantStatusSuspended TenantStatus = "suspended"
)

// NewTenant creates a new tenant entity
func NewTenant(name, slug string, tier *Tier, agencySeatLimit int) *Tenant {
	now := time.Now()
	return &Tenant{
		id:              uuid.New(),
		name:            name,
		slug:            slug,
		status:          TenantStatusActive,
		tier:            tier,
		agencySeatLimit: agencySeatLimit,
		createdAt:       now,
		updatedAt:       now,
		deletedAt:       nil,
	}
}

// NewTenantWithID creates a tenant entity with a specific ID (used for reconstruction from database)
func NewTenantWithID(id uuid.UUID, name, slug string, status TenantStatus, tier *Tier, agencySeatLimit int, createdAt, updatedAt time.Time, deletedAt *time.Time) *Tenant {
	return &Tenant{
		id:              id,
		name:            name,
		slug:            slug,
		status:          status,
		tier:            tier,
		agencySeatLimit: agencySeatLimit,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
		deletedAt:       deletedAt,
	}
}

// ID returns the tenant ID
func (t *Tenant) ID() uuid.UUID {
	return t.id
}

// Name returns the tenant name
func (t *Tenant) Name() string {
	return t.name
}

// Slug returns the tenant slug
func (t *Tenant) Slug() string {
	return t.slug
}

// Status returns the tenant status
func (t *Tenant) Status() TenantStatus {
	return t.status
}

// CreatedAt returns the creation timestamp
func (t *Tenant) CreatedAt() time.Time {
	return t.createdAt
}

// UpdatedAt returns the update timestamp
func (t *Tenant) UpdatedAt() time.Time {
	return t.updatedAt
}

// DeletedAt returns the deletion timestamp
func (t *Tenant) DeletedAt() *time.Time {
	return t.deletedAt
}

// Tier returns the tenant tier
func (t *Tenant) Tier() *Tier {
	return t.tier
}

// AgencySeatLimit returns the agency seat limit
func (t *Tenant) AgencySeatLimit() int {
	return t.agencySeatLimit
}

// SetName sets the tenant name
func (t *Tenant) SetName(name string) {
	t.name = name
	t.updatedAt = time.Now()
}

// SetSlug sets the tenant slug
func (t *Tenant) SetSlug(slug string) {
	t.slug = slug
	t.updatedAt = time.Now()
}

// SetStatus sets the tenant status
func (t *Tenant) SetStatus(status TenantStatus) {
	t.status = status
	t.updatedAt = time.Now()
}

// SetTier sets the tenant tier
func (t *Tenant) SetTier(tier *Tier) {
	t.tier = tier
	t.updatedAt = time.Now()
}

// SetAgencySeatLimit sets the agency seat limit
func (t *Tenant) SetAgencySeatLimit(limit int) {
	t.agencySeatLimit = limit
	t.updatedAt = time.Now()
}

// IsActive checks if the tenant is active
func (t *Tenant) IsActive() bool {
	return t.status == TenantStatusActive && !t.IsDeleted()
}

// Delete marks the tenant as deleted (soft delete)
func (t *Tenant) Delete() {
	now := time.Now()
	t.deletedAt = &now
	t.updatedAt = now
}

// IsDeleted checks if the tenant is deleted
func (t *Tenant) IsDeleted() bool {
	return t.deletedAt != nil
}

// Restore restores a deleted tenant
func (t *Tenant) Restore() {
	t.deletedAt = nil
	t.updatedAt = time.Now()
}

