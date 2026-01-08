package inbound

import (
	"context"
)

// DeleteBrand is the inbound port for deleting a brand
type DeleteBrand interface {
	Execute(ctx context.Context, req *DeleteBrandRequest) (*DeleteBrandResponse, error)
}

// DeleteBrandRequest represents the request
type DeleteBrandRequest struct {
	BrandID string
}

// DeleteBrandResponse represents the response
type DeleteBrandResponse struct {
	Success bool
}

