package services

import (
	"strings"

	"github.com/google/uuid"
)

// KeyGenerator generates S3 object keys
type KeyGenerator struct{}

// NewKeyGenerator creates a new key generator
func NewKeyGenerator() *KeyGenerator {
	return &KeyGenerator{}
}

// GenerateObjectKey creates the S3 object key for branding assets
func (g *KeyGenerator) GenerateObjectKey(agencyID uuid.UUID, asset string) string {
	return "branding/" + agencyID.String() + "/" + asset
}

// ValidateKey validates that a key doesn't contain invalid characters
func (g *KeyGenerator) ValidateKey(key string) bool {
	// Reject paths containing '..' or starting with '/'
	if strings.Contains(key, "..") || strings.HasPrefix(key, "/") {
		return false
	}
	return true
}

