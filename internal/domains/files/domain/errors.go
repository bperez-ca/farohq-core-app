package domain

import "errors"

var (
	// ErrFileNotFound is returned when a file is not found
	ErrFileNotFound = errors.New("file not found")

	// ErrInvalidAsset is returned when an asset type is invalid
	ErrInvalidAsset = errors.New("invalid asset type")

	// ErrInvalidAgencyID is returned when agency ID is invalid
	ErrInvalidAgencyID = errors.New("invalid agency ID")
)

