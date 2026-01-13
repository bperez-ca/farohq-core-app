package inbound

import (
	"context"

	"farohq-core-app/internal/domains/brand/domain/model"
)

// UpdateBrand is the inbound port for updating a brand
type UpdateBrand interface {
	Execute(ctx context.Context, req *UpdateBrandRequest) (*UpdateBrandResponse, error)
}

// UpdateBrandRequest represents the request
type UpdateBrandRequest struct {
	BrandID        string
	Domain         *string                 // Optional: Custom domain (Scale tier only)
	Website        *string                 // Optional: Agency website URL
	LogoURL        *string
	FaviconURL     *string
	PrimaryColor   *string
	SecondaryColor *string
	ThemeJSON      *map[string]interface{}
	HidePoweredBy  *bool                   // Optional: Hide "Powered by Faro" badge (Growth+ tiers only)
}

// UpdateBrandResponse represents the response
type UpdateBrandResponse struct {
	Branding *model.Branding
}

