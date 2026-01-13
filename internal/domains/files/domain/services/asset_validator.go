package services

// AllowedAssets defines the allowed asset types for branding
var AllowedAssets = []string{
	"logo",
	"favicon",
	"logo.png",
	"logo.svg",
	"favicon.ico",
	"favicon.png",
	"bg.jpg",
	"bg.png",
}

// AssetValidator validates asset types, dimensions, and file formats
type AssetValidator struct{}

// NewAssetValidator creates a new asset validator
func NewAssetValidator() *AssetValidator {
	return &AssetValidator{}
}

// IsValidAsset checks if the asset type is in the allowed list
// Supports both specific filenames and asset types (logo, favicon)
func (v *AssetValidator) IsValidAsset(asset string) bool {
	// Check for exact match (e.g., "logo.png", "favicon.ico")
	for _, allowed := range AllowedAssets {
		if asset == allowed {
			return true
		}
	}

	// Check for asset type match (e.g., "logo", "favicon")
	// These are used when generating keys: branding/{tenant_id}/{asset_type}/{filename}
	if asset == "logo" || asset == "favicon" {
		return true
	}

	return false
}

// ValidateImageDimensions validates image dimensions
// Max 2048x2048px, min 64x64px for logos
// Max 512x512px, min 16x16px for favicons
func (v *AssetValidator) ValidateImageDimensions(width, height int, assetType string) bool {
	if assetType == "logo" {
		// Logo: min 64x64, max 2048x2048
		return width >= 64 && height >= 64 && width <= 2048 && height <= 2048
	} else if assetType == "favicon" {
		// Favicon: min 16x16, max 512x512 (typically square)
		return width >= 16 && height >= 16 && width <= 512 && height <= 512
	}
	return true
}

// ValidateAspectRatio validates aspect ratio for logos
// Logos: 1:1 to 4:1 (square to wide)
// Favicons: 1:1 (square only)
func (v *AssetValidator) ValidateAspectRatio(width, height int, assetType string) bool {
	if height == 0 {
		return false
	}
	ratio := float64(width) / float64(height)

	if assetType == "logo" {
		// Logo: 1:1 to 4:1
		return ratio >= 1.0 && ratio <= 4.0
	} else if assetType == "favicon" {
		// Favicon: 1:1 (square only, allow small tolerance)
		return ratio >= 0.95 && ratio <= 1.05
	}
	return true
}

// ValidateFileSize validates file size
// Max 2MB for logos, 1MB for favicons
func (v *AssetValidator) ValidateFileSize(size int64, assetType string) bool {
	if assetType == "logo" {
		return size <= 2*1024*1024 // 2MB
	} else if assetType == "favicon" {
		return size <= 1*1024*1024 // 1MB
	}
	return size <= 2*1024*1024 // Default: 2MB
}

// ValidateSVGContent performs basic SVG validation
// Checks for embedded scripts to prevent XSS attacks
// Note: This is a basic check; full SVG validation would require a proper parser
func (v *AssetValidator) ValidateSVGContent(content []byte) bool {
	contentStr := string(content)

	// Reject SVG files with embedded scripts (prevent XSS)
	// Check for common script patterns
	scriptPatterns := []string{
		"<script",
		"javascript:",
		"onerror=",
		"onload=",
		"<iframe",
		"<embed",
	}

	for _, pattern := range scriptPatterns {
		// Case-insensitive check
		for i := 0; i <= len(contentStr)-len(pattern); i++ {
			if len(contentStr) >= i+len(pattern) {
				substr := contentStr[i : i+len(pattern)]
				// Simple case-insensitive check
				if substr == pattern || (len(substr) == len(pattern) && isCaseInsensitiveMatch(substr, pattern)) {
					return false
				}
			}
		}
	}

	return true
}

// isCaseInsensitiveMatch checks if two strings match case-insensitively
func isCaseInsensitiveMatch(s1, s2 string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := 0; i < len(s1); i++ {
		c1 := s1[i]
		c2 := s2[i]
		// Convert to lowercase for comparison
		if c1 >= 'A' && c1 <= 'Z' {
			c1 += 32
		}
		if c2 >= 'A' && c2 <= 'Z' {
			c2 += 32
		}
		if c1 != c2 {
			return false
		}
	}
	return true
}

