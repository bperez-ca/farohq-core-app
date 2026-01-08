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

