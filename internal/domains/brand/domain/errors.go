package domain

import "errors"

var (
	// ErrBrandingNotFound is returned when branding is not found
	ErrBrandingNotFound = errors.New("branding not found")

	// ErrInvalidDomain is returned when domain is invalid
	ErrInvalidDomain = errors.New("invalid domain")
)

