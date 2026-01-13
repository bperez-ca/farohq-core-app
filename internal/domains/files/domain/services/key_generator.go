package services

import (
	"strings"

	"github.com/google/uuid"
)

// KeyGenerator generates GCS object keys (previously S3)
// Supports tenant isolation: gs://bucket/{tenant_id}/branding/{asset_type}/{filename}
type KeyGenerator struct{}

// NewKeyGenerator creates a new key generator
func NewKeyGenerator() *KeyGenerator {
	return &KeyGenerator{}
}

// GenerateObjectKey creates the GCS object key for branding assets
// Format: {tenant_id}/branding/{asset_type}/{filename}
// Example: {uuid}/branding/logo/logo.png
func (g *KeyGenerator) GenerateObjectKey(agencyID uuid.UUID, asset string) string {
	// Asset can be "logo", "favicon", or a specific filename
	// If asset is "logo" or "favicon", we'll append a default extension later
	// For now, use the format: {tenant_id}/branding/{asset}
	return agencyID.String() + "/branding/" + asset
}

// GenerateObjectKeyWithFilename creates a GCS object key with specific filename
// Format: {tenant_id}/branding/{asset_type}/{filename}
// Example: {uuid}/branding/logo/logo.png
func (g *KeyGenerator) GenerateObjectKeyWithFilename(agencyID uuid.UUID, assetType, filename string) string {
	return agencyID.String() + "/branding/" + assetType + "/" + filename
}

// ValidateKey validates that a key doesn't contain invalid characters
func (g *KeyGenerator) ValidateKey(key string) bool {
	// Reject paths containing '..' or starting with '/'
	if strings.Contains(key, "..") || strings.HasPrefix(key, "/") {
		return false
	}
	return true
}

