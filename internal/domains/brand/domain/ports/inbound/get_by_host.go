package inbound

import (
	"context"

	"farohq-core-app/internal/domains/brand/domain/model"
)

// GetByHost is the inbound port for getting branding by host
type GetByHost interface {
	Execute(ctx context.Context, req *GetByHostRequest) (*GetByHostResponse, error)
}

// GetByHostRequest represents the request
type GetByHostRequest struct {
	Host string
}

// GetByHostResponse represents the response
type GetByHostResponse struct {
	Branding *model.Branding
}

