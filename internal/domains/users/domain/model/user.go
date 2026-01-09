package model

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system (synced from Clerk)
type User struct {
	id            uuid.UUID
	clerkUserID   string
	email         string
	firstName     string
	lastName      string
	fullName      string
	imageURL      string
	phoneNumbers  []string
	createdAt     time.Time
	updatedAt     time.Time
	lastSignInAt  *time.Time
}

// NewUser creates a new user entity
func NewUser(clerkUserID, email, firstName, lastName, fullName, imageURL string, phoneNumbers []string) *User {
	now := time.Now()
	if phoneNumbers == nil {
		phoneNumbers = []string{}
	}
	return &User{
		id:           uuid.New(),
		clerkUserID:  clerkUserID,
		email:        email,
		firstName:    firstName,
		lastName:     lastName,
		fullName:     fullName,
		imageURL:     imageURL,
		phoneNumbers: phoneNumbers,
		createdAt:    now,
		updatedAt:    now,
		lastSignInAt: nil,
	}
}

// NewUserWithID creates a user entity with a specific ID (used for reconstruction from database)
func NewUserWithID(id uuid.UUID, clerkUserID, email, firstName, lastName, fullName, imageURL string, phoneNumbers []string, createdAt, updatedAt time.Time, lastSignInAt *time.Time) *User {
	if phoneNumbers == nil {
		phoneNumbers = []string{}
	}
	return &User{
		id:           id,
		clerkUserID:  clerkUserID,
		email:        email,
		firstName:    firstName,
		lastName:     lastName,
		fullName:     fullName,
		imageURL:     imageURL,
		phoneNumbers: phoneNumbers,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		lastSignInAt: lastSignInAt,
	}
}

// ID returns the user ID
func (u *User) ID() uuid.UUID {
	return u.id
}

// ClerkUserID returns the Clerk user ID
func (u *User) ClerkUserID() string {
	return u.clerkUserID
}

// Email returns the user email
func (u *User) Email() string {
	return u.email
}

// FirstName returns the user first name
func (u *User) FirstName() string {
	return u.firstName
}

// LastName returns the user last name
func (u *User) LastName() string {
	return u.lastName
}

// FullName returns the user full name
func (u *User) FullName() string {
	return u.fullName
}

// ImageURL returns the user image URL
func (u *User) ImageURL() string {
	return u.imageURL
}

// PhoneNumbers returns the user phone numbers
func (u *User) PhoneNumbers() []string {
	return u.phoneNumbers
}

// CreatedAt returns the creation timestamp
func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

// UpdatedAt returns the update timestamp
func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

// LastSignInAt returns the last sign in timestamp
func (u *User) LastSignInAt() *time.Time {
	return u.lastSignInAt
}

// UpdateFromClerk updates user data from Clerk
func (u *User) UpdateFromClerk(email, firstName, lastName, fullName, imageURL string, phoneNumbers []string, lastSignInAt *time.Time) {
	u.email = email
	u.firstName = firstName
	u.lastName = lastName
	u.fullName = fullName
	u.imageURL = imageURL
	if phoneNumbers != nil {
		u.phoneNumbers = phoneNumbers
	}
	if lastSignInAt != nil {
		u.lastSignInAt = lastSignInAt
	}
	u.updatedAt = time.Now()
}
