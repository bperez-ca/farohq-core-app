package domain

import "errors"

var (
	// ErrTenantNotFound is returned when a tenant is not found
	ErrTenantNotFound = errors.New("tenant not found")

	// ErrTenantAlreadyExists is returned when a tenant with the same slug already exists
	ErrTenantAlreadyExists = errors.New("tenant already exists")

	// ErrInvalidTenantName is returned when tenant name is invalid
	ErrInvalidTenantName = errors.New("invalid tenant name")

	// ErrInvalidTenantSlug is returned when tenant slug is invalid
	ErrInvalidTenantSlug = errors.New("invalid tenant slug")

	// ErrMemberNotFound is returned when a member is not found
	ErrMemberNotFound = errors.New("member not found")

	// ErrMemberAlreadyExists is returned when a user is already a member
	ErrMemberAlreadyExists = errors.New("member already exists")

	// ErrInvalidRole is returned when an invalid role is provided
	ErrInvalidRole = errors.New("invalid role")

	// ErrInviteNotFound is returned when an invite is not found
	ErrInviteNotFound = errors.New("invite not found")

	// ErrInviteExpired is returned when an invite has expired
	ErrInviteExpired = errors.New("invite expired")

	// ErrInviteAlreadyAccepted is returned when an invite has already been accepted
	ErrInviteAlreadyAccepted = errors.New("invite already accepted")

	// ErrInviteRevoked is returned when an invite has been revoked
	ErrInviteRevoked = errors.New("invite revoked")

	// ErrPendingInviteExists is returned when a pending invite already exists for the email/tenant
	ErrPendingInviteExists = errors.New("a pending invitation already exists for this email")

	// ErrInvalidEmail is returned when an email is invalid
	ErrInvalidEmail = errors.New("invalid email")

	// ErrUnauthorized is returned when a user is not authorized to perform an action
	ErrUnauthorized = errors.New("unauthorized")

	// ErrClientNotFound is returned when a client is not found
	ErrClientNotFound = errors.New("client not found")

	// ErrClientAlreadyExists is returned when a client with the same slug already exists
	ErrClientAlreadyExists = errors.New("client already exists")

	// ErrLocationNotFound is returned when a location is not found
	ErrLocationNotFound = errors.New("location not found")

	// ErrClientMemberNotFound is returned when a client member is not found
	ErrClientMemberNotFound = errors.New("client member not found")

	// ErrAgencySeatLimitExceeded is returned when agency seat limit is exceeded
	ErrAgencySeatLimitExceeded = errors.New("agency seat limit exceeded")

	// ErrClientSeatLimitExceeded is returned when client seat limit is exceeded
	ErrClientSeatLimitExceeded = errors.New("client seat limit exceeded")
)
