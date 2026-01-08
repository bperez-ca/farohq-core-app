package inbound

import (
	"context"

	"farohq-core-app/internal/domains/brand/domain/model"
)

// ListBrands is the inbound port for listing brands
type ListBrands interface {
	Execute(ctx context.Context, req *ListBrandsRequest) (*ListBrandsResponse, error)
}

// ListBrandsRequest represents the request
type ListBrandsRequest struct {
	AgencyID string
}

// ListBrandsResponse represents the response
type ListBrandsResponse struct {
	Brands []*model.Branding
}

