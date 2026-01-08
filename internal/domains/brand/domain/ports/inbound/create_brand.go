package inbound

import (
	"context"

	"farohq-core-app/internal/domains/brand/domain/model"
)

// CreateBrand is the inbound port for creating a brand
type CreateBrand interface {
	Execute(ctx context.Context, req *CreateBrandRequest) (*CreateBrandResponse, error)
}

// CreateBrandRequest represents the request
type CreateBrandRequest struct {
	AgencyID       string
	Domain         string
	LogoURL        string
	FaviconURL     string
	PrimaryColor   string
	SecondaryColor string
	ThemeJSON      map[string]interface{}
}

// CreateBrandResponse represents the response
type CreateBrandResponse struct {
	Branding *model.Branding
}

