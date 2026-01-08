package usecases

import (
	"context"

	"farohq-core-app/internal/domains/tenants/domain/model"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"

	"github.com/google/uuid"
)

// ListLocations handles the use case of listing locations for a client
type ListLocations struct {
	locationRepo outbound.LocationRepository
}

// NewListLocations creates a new ListLocations use case
func NewListLocations(locationRepo outbound.LocationRepository) *ListLocations {
	return &ListLocations{
		locationRepo: locationRepo,
	}
}

// ListLocationsRequest represents the request to list locations
type ListLocationsRequest struct {
	ClientID uuid.UUID
}

// ListLocationsResponse represents the response from listing locations
type ListLocationsResponse struct {
	Locations []*model.Location
}

// Execute executes the use case
func (uc *ListLocations) Execute(ctx context.Context, req *ListLocationsRequest) (*ListLocationsResponse, error) {
	locations, err := uc.locationRepo.ListByClient(ctx, req.ClientID)
	if err != nil {
		return nil, err
	}

	return &ListLocationsResponse{
		Locations: locations,
	}, nil
}

