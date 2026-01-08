package usecases

import (
	"context"

	"farohq-core-app/internal/domains/tenants/domain/services"
	"farohq-core-app/internal/domains/tenants/domain/ports/outbound"

	"github.com/google/uuid"
)

// GetSeatUsage handles the use case of getting seat usage for an agency or client
type GetSeatUsage struct {
	tenantRepo       outbound.TenantRepository
	clientRepo       outbound.ClientRepository
	clientMemberRepo outbound.ClientMemberRepository
	locationRepo     outbound.LocationRepository
}

// NewGetSeatUsage creates a new GetSeatUsage use case
func NewGetSeatUsage(
	tenantRepo outbound.TenantRepository,
	clientRepo outbound.ClientRepository,
	clientMemberRepo outbound.ClientMemberRepository,
	locationRepo outbound.LocationRepository,
) *GetSeatUsage {
	return &GetSeatUsage{
		tenantRepo:       tenantRepo,
		clientRepo:       clientRepo,
		clientMemberRepo: clientMemberRepo,
		locationRepo:     locationRepo,
	}
}

// GetSeatUsageRequest represents the request to get seat usage
type GetSeatUsageRequest struct {
	AgencyID *uuid.UUID // optional - if provided, get agency seat usage
	ClientID *uuid.UUID // optional - if provided, get client seat usage
}

// SeatUsage represents seat usage information
type SeatUsage struct {
	AgencySeatsUsed  int
	AgencySeatsLimit int
	ClientSeatsUsed  int
	ClientSeatsLimit int
	TotalClients     int
	TotalLocations   int
	ClientBreakdown  []ClientSeatInfo
}

// ClientSeatInfo represents seat information for a specific client
type ClientSeatInfo struct {
	ClientID   uuid.UUID
	ClientName string
	Locations  int
	Members    int
	SeatLimit  int
	SeatsUsed  int
}

// GetSeatUsageResponse represents the response from getting seat usage
type GetSeatUsageResponse struct {
	Usage *SeatUsage
}

// Execute executes the use case
func (uc *GetSeatUsage) Execute(ctx context.Context, req *GetSeatUsageRequest) (*GetSeatUsageResponse, error) {
	usage := &SeatUsage{
		ClientBreakdown: []ClientSeatInfo{},
	}

	if req.AgencyID != nil {
		// Get agency seat usage
		agency, err := uc.tenantRepo.FindByID(ctx, *req.AgencyID)
		if err != nil {
			return nil, err
		}

		usage.AgencySeatsLimit = agency.AgencySeatLimit()
		// TODO: count agency members (need to add method to tenant member repo)
		// For now, we'll leave it at 0

		// Get client breakdown
		clients, err := uc.clientRepo.ListByAgency(ctx, *req.AgencyID)
		if err != nil {
			return nil, err
		}

		usage.TotalClients = len(clients)

		for _, client := range clients {
			locationCount, err := uc.locationRepo.CountByClient(ctx, client.ID())
			if err != nil {
				return nil, err
			}

			memberCount, err := uc.clientMemberRepo.CountByClient(ctx, client.ID())
			if err != nil {
				return nil, err
			}

			seatLimit := services.CalculateClientSeatLimit(locationCount)
			usage.TotalLocations += locationCount
			usage.ClientSeatsUsed += memberCount

			usage.ClientBreakdown = append(usage.ClientBreakdown, ClientSeatInfo{
				ClientID:   client.ID(),
				ClientName: client.Name(),
				Locations:  locationCount,
				Members:    memberCount,
				SeatLimit:  seatLimit,
				SeatsUsed:  memberCount,
			})
		}
	}

	if req.ClientID != nil {
		// Get client-specific seat usage
		locationCount, err := uc.locationRepo.CountByClient(ctx, *req.ClientID)
		if err != nil {
			return nil, err
		}

		memberCount, err := uc.clientMemberRepo.CountByClient(ctx, *req.ClientID)
		if err != nil {
			return nil, err
		}

		usage.ClientSeatsLimit = services.CalculateClientSeatLimit(locationCount)
		usage.ClientSeatsUsed = memberCount
		usage.TotalLocations = locationCount
	}

	return &GetSeatUsageResponse{
		Usage: usage,
	}, nil
}

