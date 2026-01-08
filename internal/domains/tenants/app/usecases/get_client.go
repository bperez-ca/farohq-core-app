package usecases

import (
	"context"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/model"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"

	"github.com/google/uuid"
)

// GetClient handles the use case of getting a client by ID
type GetClient struct {
	clientRepo outbound.ClientRepository
}

// NewGetClient creates a new GetClient use case
func NewGetClient(clientRepo outbound.ClientRepository) *GetClient {
	return &GetClient{
		clientRepo: clientRepo,
	}
}

// GetClientRequest represents the request to get a client
type GetClientRequest struct {
	ClientID uuid.UUID
}

// GetClientResponse represents the response from getting a client
type GetClientResponse struct {
	Client *model.Client
}

// Execute executes the use case
func (uc *GetClient) Execute(ctx context.Context, req *GetClientRequest) (*GetClientResponse, error) {
	client, err := uc.clientRepo.FindByID(ctx, req.ClientID)
	if err != nil {
		return nil, domain.ErrClientNotFound
	}

	return &GetClientResponse{
		Client: client,
	}, nil
}

