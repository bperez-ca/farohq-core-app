package model

// Tier represents the subscription tier of an agency
type Tier string

const (
	TierStarter Tier = "starter"
	TierGrowth  Tier = "growth"
	TierScale   Tier = "scale"
)

// TierClientLimit returns the maximum number of clients allowed for a tier
func TierClientLimit(tier Tier) int {
	switch tier {
	case TierStarter:
		return 15
	case TierGrowth:
		return 50
	case TierScale:
		return 200
	default:
		return 0
	}
}

// IsValidTier checks if a tier is valid
func IsValidTier(tier Tier) bool {
	return tier == TierStarter || tier == TierGrowth || tier == TierScale
}

// String returns the string representation of the tier
func (t Tier) String() string {
	return string(t)
}

// TierSupportsCustomDomain checks if a tier supports custom domain configuration
// Only Scale tier can configure custom domains
func TierSupportsCustomDomain(tier *Tier) bool {
	if tier == nil {
		return false
	}
	return *tier == TierScale
}

// TierCanHidePoweredBy checks if a tier can hide "Powered by Faro" badge
// Growth+ tiers (growth, scale) can hide the badge
func TierCanHidePoweredBy(tier *Tier) bool {
	if tier == nil {
		return false
	}
	return *tier == TierGrowth || *tier == TierScale
}// TierUsesSubdomain checks if a tier uses subdomain for portal access
// Lower tiers (Starter, Growth) always use subdomain
// Scale tier can use custom domain OR subdomain
func TierUsesSubdomain(tier *Tier) bool {
	if tier == nil {
		return true // Default to subdomain if tier not set
	}
	return *tier == TierStarter || *tier == TierGrowth
}