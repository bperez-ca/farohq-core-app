package usecases

import (
	"context"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/model"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"

	"github.com/google/uuid"
)

// ListClients handles the use case of listing clients for an agency
type ListClients struct {
	clientRepo outbound.ClientRepository
	tenantRepo outbound.TenantRepository
}

// NewListClients creates a new ListClients use case
func NewListClients(
	clientRepo outbound.ClientRepository,
	tenantRepo outbound.TenantRepository,
) *ListClients {
	return &ListClients{
		clientRepo: clientRepo,
		tenantRepo: tenantRepo,
	}
}

// ListClientsRequest represents the request to list clients
type ListClientsRequest struct {
	AgencyID uuid.UUID
}

// ListClientsResponse represents the response from listing clients
type ListClientsResponse struct {
	Clients []*model.Client
}

// Execute executes the use case
func (uc *ListClients) Execute(ctx context.Context, req *ListClientsRequest) (*ListClientsResponse, error) {
	// Verify agency exists
	_, err := uc.tenantRepo.FindByID(ctx, req.AgencyID)
	if err != nil {
		return nil, domain.ErrTenantNotFound
	}

	// List clients
	clients, err := uc.clientRepo.ListByAgency(ctx, req.AgencyID)
	if err != nil {
		return nil, err
	}

	return &ListClientsResponse{
		Clients: clients,
	}, nil
}

