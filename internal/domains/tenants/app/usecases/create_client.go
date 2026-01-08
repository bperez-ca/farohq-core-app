package usecases

import (
	"context"
	"strings"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/model"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"
	"farohq-core-app/internal/domains/tenants/domain/services"

	"github.com/google/uuid"
)

// CreateClient handles the use case of creating a new client (SMB) under an agency
type CreateClient struct {
	clientRepo    outbound.ClientRepository
	tenantRepo    outbound.TenantRepository
	seatValidator *services.SeatValidator
}

// NewCreateClient creates a new CreateClient use case
func NewCreateClient(
	clientRepo outbound.ClientRepository,
	tenantRepo outbound.TenantRepository,
	seatValidator *services.SeatValidator,
) *CreateClient {
	return &CreateClient{
		clientRepo:    clientRepo,
		tenantRepo:    tenantRepo,
		seatValidator: seatValidator,
	}
}

// CreateClientRequest represents the request to create a client
type CreateClientRequest struct {
	AgencyID uuid.UUID
	Name     string
	Slug     string
	Tier     model.Tier
}

// CreateClientResponse represents the response from creating a client
type CreateClientResponse struct {
	Client *model.Client
}

// Execute executes the use case
func (uc *CreateClient) Execute(ctx context.Context, req *CreateClientRequest) (*CreateClientResponse, error) {
	// Validate input
	if strings.TrimSpace(req.Name) == "" {
		return nil, domain.ErrInvalidTenantName
	}

	if strings.TrimSpace(req.Slug) == "" {
		return nil, domain.ErrInvalidTenantSlug
	}

	if !model.IsValidTier(req.Tier) {
		return nil, domain.ErrInvalidRole // TODO: create ErrInvalidTier
	}

	// Verify agency exists
	agency, err := uc.tenantRepo.FindByID(ctx, req.AgencyID)
	if err != nil {
		return nil, domain.ErrTenantNotFound
	}

	// Validate tier limits
	agencyTier := agency.Tier()
	if agencyTier == nil {
		return nil, domain.ErrInvalidRole // TODO: create ErrAgencyTierNotSet
	}

	// Get current client count for this tier
	currentCount, err := uc.clientRepo.CountByAgencyAndTier(ctx, req.AgencyID, req.Tier)
	if err != nil {
		return nil, err
	}

	// Check tier limit
	tierLimit := model.TierClientLimit(*agencyTier)
	if currentCount >= tierLimit {
		return nil, domain.ErrClientAlreadyExists // TODO: create ErrTierLimitExceeded
	}

	// Normalize slug
	slug := strings.ToLower(strings.TrimSpace(req.Slug))
	slug = strings.ReplaceAll(slug, " ", "-")

	// Check if client with slug already exists
	existing, err := uc.clientRepo.FindBySlug(ctx, req.AgencyID, slug)
	if err == nil && existing != nil {
		return nil, domain.ErrClientAlreadyExists
	}

	// Create new client
	client := model.NewClient(req.AgencyID, strings.TrimSpace(req.Name), slug, req.Tier)

	// Save client
	if err := uc.clientRepo.Save(ctx, client); err != nil {
		return nil, err
	}

	return &CreateClientResponse{
		Client: client,
	}, nil
}

