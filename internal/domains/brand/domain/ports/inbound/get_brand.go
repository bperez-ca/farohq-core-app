package inbound

import (
	"context"

	"farohq-core-app/internal/domains/brand/domain/model"
)

// GetBrand is the inbound port for getting a brand by ID
type GetBrand interface {
	Execute(ctx context.Context, req *GetBrandRequest) (*GetBrandResponse, error)
}

// GetBrandRequest represents the request
type GetBrandRequest struct {
	BrandID string
}

// GetBrandResponse represents the response
type GetBrandResponse struct {
	Branding *model.Branding
}

