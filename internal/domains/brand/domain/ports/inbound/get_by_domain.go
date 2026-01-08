package inbound

import (
	"context"

	"farohq-core-app/internal/domains/brand/domain/model"
)

// GetByDomain is the inbound port for getting branding by domain
type GetByDomain interface {
	Execute(ctx context.Context, req *GetByDomainRequest) (*GetByDomainResponse, error)
}

// GetByDomainRequest represents the request
type GetByDomainRequest struct {
	Domain string
}

// GetByDomainResponse represents the response
type GetByDomainResponse struct {
	Branding *model.Branding
}

