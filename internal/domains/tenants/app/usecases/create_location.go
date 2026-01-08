package usecases

import (
	"context"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/model"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"

	"github.com/google/uuid"
)

// CreateLocation handles the use case of creating a location for a client
type CreateLocation struct {
	locationRepo outbound.LocationRepository
	clientRepo   outbound.ClientRepository
}

// NewCreateLocation creates a new CreateLocation use case
func NewCreateLocation(
	locationRepo outbound.LocationRepository,
	clientRepo outbound.ClientRepository,
) *CreateLocation {
	return &CreateLocation{
		locationRepo: locationRepo,
		clientRepo:   clientRepo,
	}
}

// CreateLocationRequest represents the request to create a location
type CreateLocationRequest struct {
	ClientID uuid.UUID
	Name     string
}

// CreateLocationResponse represents the response from creating a location
type CreateLocationResponse struct {
	Location *model.Location
}

// Execute executes the use case
func (uc *CreateLocation) Execute(ctx context.Context, req *CreateLocationRequest) (*CreateLocationResponse, error) {
	// Validate client exists and is active
	client, err := uc.clientRepo.FindByID(ctx, req.ClientID)
	if err != nil {
		return nil, domain.ErrClientNotFound
	}

	if !client.IsActive() {
		return nil, domain.ErrClientNotFound // TODO: create ErrClientInactive
	}

	// Create new location
	location := model.NewLocation(req.ClientID, req.Name)

	// Save location
	if err := uc.locationRepo.Save(ctx, location); err != nil {
		return nil, err
	}

	return &CreateLocationResponse{
		Location: location,
	}, nil
}

