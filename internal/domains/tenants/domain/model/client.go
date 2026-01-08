package model

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Client represents a client (SMB) account under an agency
type Client struct {
	id        uuid.UUID
	agencyID  uuid.UUID
	name      string
	slug      string
	tier      Tier
	status    ClientStatus
	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time
}

// ClientStatus represents the status of a client
type ClientStatus string

const (
	ClientStatusActive    ClientStatus = "active"
	ClientStatusInactive  ClientStatus = "inactive"
	ClientStatusSuspended ClientStatus = "suspended"
)

// NewClient creates a new client entity
func NewClient(agencyID uuid.UUID, name, slug string, tier Tier) *Client {
	now := time.Now()
	return &Client{
		id:        uuid.New(),
		agencyID:  agencyID,
		name:      strings.TrimSpace(name),
		slug:      normalizeSlug(slug),
		tier:      tier,
		status:    ClientStatusActive,
		createdAt: now,
		updatedAt: now,
		deletedAt: nil,
	}
}

// NewClientWithID creates a client entity with a specific ID (used for reconstruction from database)
func NewClientWithID(id, agencyID uuid.UUID, name, slug string, tier Tier, status ClientStatus, createdAt, updatedAt time.Time, deletedAt *time.Time) *Client {
	return &Client{
		id:        id,
		agencyID:  agencyID,
		name:      name,
		slug:      slug,
		tier:      tier,
		status:    status,
		createdAt: createdAt,
		updatedAt: updatedAt,
		deletedAt: deletedAt,
	}
}

// ID returns the client ID
func (c *Client) ID() uuid.UUID {
	return c.id
}

// AgencyID returns the agency ID
func (c *Client) AgencyID() uuid.UUID {
	return c.agencyID
}

// Name returns the client name
func (c *Client) Name() string {
	return c.name
}

// Slug returns the client slug
func (c *Client) Slug() string {
	return c.slug
}

// Tier returns the client tier
func (c *Client) Tier() Tier {
	return c.tier
}

// Status returns the client status
func (c *Client) Status() ClientStatus {
	return c.status
}

// CreatedAt returns the creation timestamp
func (c *Client) CreatedAt() time.Time {
	return c.createdAt
}

// UpdatedAt returns the update timestamp
func (c *Client) UpdatedAt() time.Time {
	return c.updatedAt
}

// DeletedAt returns the deletion timestamp
func (c *Client) DeletedAt() *time.Time {
	return c.deletedAt
}

// SetName sets the client name
func (c *Client) SetName(name string) {
	c.name = strings.TrimSpace(name)
	c.updatedAt = time.Now()
}

// SetSlug sets the client slug
func (c *Client) SetSlug(slug string) {
	c.slug = normalizeSlug(slug)
	c.updatedAt = time.Now()
}

// SetTier sets the client tier
func (c *Client) SetTier(tier Tier) {
	c.tier = tier
	c.updatedAt = time.Now()
}

// SetStatus sets the client status
func (c *Client) SetStatus(status ClientStatus) {
	c.status = status
	c.updatedAt = time.Now()
}

// IsActive checks if the client is active
func (c *Client) IsActive() bool {
	return c.status == ClientStatusActive && !c.IsDeleted()
}

// Delete marks the client as deleted (soft delete)
func (c *Client) Delete() {
	now := time.Now()
	c.deletedAt = &now
	c.updatedAt = now
}

// IsDeleted checks if the client is deleted
func (c *Client) IsDeleted() bool {
	return c.deletedAt != nil
}

// Restore restores a deleted client
func (c *Client) Restore() {
	c.deletedAt = nil
	c.updatedAt = time.Now()
}

// normalizeSlug normalizes a slug string
func normalizeSlug(slug string) string {
	slug = strings.ToLower(strings.TrimSpace(slug))
	slug = strings.ReplaceAll(slug, " ", "-")
	return slug
}

