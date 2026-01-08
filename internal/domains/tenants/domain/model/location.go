package model

import (
	"time"

	"github.com/google/uuid"
)

// Location represents a business location for a client
type Location struct {
	id            uuid.UUID
	clientID      uuid.UUID
	name          string
	address       map[string]interface{} // JSONB in DB
	phone         string
	businessHours map[string]interface{} // JSONB in DB
	categories    []string
	isActive      bool
	createdAt     time.Time
	updatedAt     time.Time
	deletedAt     *time.Time
}

// NewLocation creates a new location entity
func NewLocation(clientID uuid.UUID, name string) *Location {
	now := time.Now()
	return &Location{
		id:            uuid.New(),
		clientID:      clientID,
		name:          name,
		address:       make(map[string]interface{}),
		businessHours: make(map[string]interface{}),
		categories:    []string{},
		isActive:      true,
		createdAt:     now,
		updatedAt:     now,
		deletedAt:     nil,
	}
}

// NewLocationWithID creates a location entity with a specific ID (used for reconstruction from database)
func NewLocationWithID(id, clientID uuid.UUID, name, phone string, address, businessHours map[string]interface{}, categories []string, isActive bool, createdAt, updatedAt time.Time, deletedAt *time.Time) *Location {
	if address == nil {
		address = make(map[string]interface{})
	}
	if businessHours == nil {
		businessHours = make(map[string]interface{})
	}
	if categories == nil {
		categories = []string{}
	}
	return &Location{
		id:            id,
		clientID:      clientID,
		name:          name,
		phone:         phone,
		address:       address,
		businessHours: businessHours,
		categories:    categories,
		isActive:      isActive,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
		deletedAt:     deletedAt,
	}
}

// ID returns the location ID
func (l *Location) ID() uuid.UUID {
	return l.id
}

// ClientID returns the client ID
func (l *Location) ClientID() uuid.UUID {
	return l.clientID
}

// Name returns the location name
func (l *Location) Name() string {
	return l.name
}

// Address returns the location address
func (l *Location) Address() map[string]interface{} {
	return l.address
}

// Phone returns the location phone
func (l *Location) Phone() string {
	return l.phone
}

// BusinessHours returns the business hours
func (l *Location) BusinessHours() map[string]interface{} {
	return l.businessHours
}

// Categories returns the location categories
func (l *Location) Categories() []string {
	return l.categories
}

// IsActive returns whether the location is active
func (l *Location) IsActive() bool {
	return l.isActive && !l.IsDeleted()
}

// CreatedAt returns the creation timestamp
func (l *Location) CreatedAt() time.Time {
	return l.createdAt
}

// UpdatedAt returns the update timestamp
func (l *Location) UpdatedAt() time.Time {
	return l.updatedAt
}

// DeletedAt returns the deletion timestamp
func (l *Location) DeletedAt() *time.Time {
	return l.deletedAt
}

// SetName sets the location name
func (l *Location) SetName(name string) {
	l.name = name
	l.updatedAt = time.Now()
}

// SetAddress sets the location address
func (l *Location) SetAddress(address map[string]interface{}) {
	l.address = address
	l.updatedAt = time.Now()
}

// SetPhone sets the location phone
func (l *Location) SetPhone(phone string) {
	l.phone = phone
	l.updatedAt = time.Now()
}

// SetBusinessHours sets the business hours
func (l *Location) SetBusinessHours(hours map[string]interface{}) {
	l.businessHours = hours
	l.updatedAt = time.Now()
}

// SetCategories sets the location categories
func (l *Location) SetCategories(categories []string) {
	l.categories = categories
	l.updatedAt = time.Now()
}

// SetIsActive sets whether the location is active
func (l *Location) SetIsActive(isActive bool) {
	l.isActive = isActive
	l.updatedAt = time.Now()
}

// Delete marks the location as deleted (soft delete)
func (l *Location) Delete() {
	now := time.Now()
	l.deletedAt = &now
	l.updatedAt = now
}

// IsDeleted checks if the location is deleted
func (l *Location) IsDeleted() bool {
	return l.deletedAt != nil
}

// Restore restores a deleted location
func (l *Location) Restore() {
	l.deletedAt = nil
	l.updatedAt = time.Now()
}

