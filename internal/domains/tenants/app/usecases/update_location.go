package usecases

import (
	"context"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/model"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"

	"github.com/google/uuid"
)

// UpdateLocation handles the use case of updating a location
type UpdateLocation struct {
	locationRepo outbound.LocationRepository
}

// NewUpdateLocation creates a new UpdateLocation use case
func NewUpdateLocation(locationRepo outbound.LocationRepository) *UpdateLocation {
	return &UpdateLocation{
		locationRepo: locationRepo,
	}
}

// UpdateLocationRequest represents the request to update a location
type UpdateLocationRequest struct {
	LocationID    uuid.UUID
	Name          *string
	Address       *map[string]interface{}
	Phone         *string
	BusinessHours *map[string]interface{}
	Categories    *[]string
	IsActive      *bool
}

// UpdateLocationResponse represents the response from updating a location
type UpdateLocationResponse struct {
	Location *model.Location
}

// Execute executes the use case
func (uc *UpdateLocation) Execute(ctx context.Context, req *UpdateLocationRequest) (*UpdateLocationResponse, error) {
	// Get existing location
	location, err := uc.locationRepo.FindByID(ctx, req.LocationID)
	if err != nil {
		return nil, domain.ErrLocationNotFound
	}

	// Update fields if provided
	if req.Name != nil {
		location.SetName(*req.Name)
	}

	if req.Address != nil {
		location.SetAddress(*req.Address)
	}

	if req.Phone != nil {
		location.SetPhone(*req.Phone)
	}

	if req.BusinessHours != nil {
		location.SetBusinessHours(*req.BusinessHours)
	}

	if req.Categories != nil {
		location.SetCategories(*req.Categories)
	}

	if req.IsActive != nil {
		location.SetIsActive(*req.IsActive)
	}

	// Save updated location
	if err := uc.locationRepo.Save(ctx, location); err != nil {
		return nil, err
	}

	return &UpdateLocationResponse{
		Location: location,
	}, nil
}

