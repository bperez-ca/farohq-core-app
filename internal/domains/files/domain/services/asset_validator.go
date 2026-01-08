package services

// AllowedAssets defines the allowed asset types for branding
var AllowedAssets = []string{
	"logo.png",
	"logo.svg",
	"favicon.ico",
	"favicon.png",
	"bg.jpg",
	"bg.png",
}

// AssetValidator validates asset types
type AssetValidator struct{}

// NewAssetValidator creates a new asset validator
func NewAssetValidator() *AssetValidator {
	return &AssetValidator{}
}

// IsValidAsset checks if the asset is in the allowed list
func (v *AssetValidator) IsValidAsset(asset string) bool {
	for _, allowed := range AllowedAssets {
		if asset == allowed {
			return true
		}
	}
	return false
}

