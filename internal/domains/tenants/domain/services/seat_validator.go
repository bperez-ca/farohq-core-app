package services

import "farohq-core-app/internal/domains/tenants/domain"

// SeatValidator validates seat limits for agencies and clients
type SeatValidator struct{}

// NewSeatValidator creates a new seat validator
func NewSeatValidator() *SeatValidator {
	return &SeatValidator{}
}

// ValidateAgencySeats validates that adding requestedCount members won't exceed the agency seat limit
func (v *SeatValidator) ValidateAgencySeats(seatLimit, currentCount, requestedCount int) error {
	if seatLimit <= 0 {
		// No limit set, allow unlimited
		return nil
	}
	if currentCount+requestedCount > seatLimit {
		return domain.ErrAgencySeatLimitExceeded
	}
	return nil
}

// ValidateClientSeats validates that adding requestedCount members won't exceed the client seat limit
// Client seat limit = 1 base seat + 1 per location
func (v *SeatValidator) ValidateClientSeats(locationCount, currentMemberCount, requestedCount int) error {
	seatLimit := CalculateClientSeatLimit(locationCount)
	if currentMemberCount+requestedCount > seatLimit {
		return domain.ErrClientSeatLimitExceeded
	}
	return nil
}

// CalculateClientSeatLimit calculates the seat limit for a client based on location count
// Formula: 1 base seat + 1 per location
func CalculateClientSeatLimit(locationCount int) int {
	return 1 + locationCount
}

