package usecases

import (
	"context"
	"strings"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/model"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"

	"github.com/google/uuid"
)

// UpdateClient handles the use case of updating a client
type UpdateClient struct {
	clientRepo outbound.ClientRepository
}

// NewUpdateClient creates a new UpdateClient use case
func NewUpdateClient(clientRepo outbound.ClientRepository) *UpdateClient {
	return &UpdateClient{
		clientRepo: clientRepo,
	}
}

// UpdateClientRequest represents the request to update a client
type UpdateClientRequest struct {
	ClientID uuid.UUID
	Name     *string
	Slug     *string
	Status   *model.ClientStatus
}

// UpdateClientResponse represents the response from updating a client
type UpdateClientResponse struct {
	Client *model.Client
}

// Execute executes the use case
func (uc *UpdateClient) Execute(ctx context.Context, req *UpdateClientRequest) (*UpdateClientResponse, error) {
	// Get existing client
	client, err := uc.clientRepo.FindByID(ctx, req.ClientID)
	if err != nil {
		return nil, domain.ErrClientNotFound
	}

	// Update fields if provided
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, domain.ErrInvalidTenantName
		}
		client.SetName(name)
	}

	if req.Slug != nil {
		slug := strings.ToLower(strings.TrimSpace(*req.Slug))
		slug = strings.ReplaceAll(slug, " ", "-")
		client.SetSlug(slug)
	}

	if req.Status != nil {
		client.SetStatus(*req.Status)
	}

	// Save updated client
	if err := uc.clientRepo.Save(ctx, client); err != nil {
		return nil, err
	}

	return &UpdateClientResponse{
		Client: client,
	}, nil
}

