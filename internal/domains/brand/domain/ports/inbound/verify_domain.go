package inbound

import (
	"context"

	"farohq-core-app/internal/domains/brand/domain/model"
)

// VerifyDomain is the inbound port for verifying a custom domain
type VerifyDomain interface {
	Execute(ctx context.Context, req *VerifyDomainRequest) (*VerifyDomainResponse, error)
}

// VerifyDomainRequest represents the request
type VerifyDomainRequest struct {
	BrandID string // Agency ID (brand ID is agency_id)
	Domain  string // Custom domain to verify (Scale tier only)
}

// VerifyDomainResponse represents the response
type VerifyDomainResponse struct {
	Branding      *model.Branding
	Verified      bool
	ExpectedCNAME string // CNAME target from Vercel API (value may vary, don't hardcode)
	CurrentCNAME  string // Current DNS record (optional, for UX feedback only)
	SSLStatus     string // SSL status from Vercel API: "pending", "active", "failed"
}

// GetDomainStatus is the inbound port for getting full domain status
type GetDomainStatus interface {
	Execute(ctx context.Context, req *GetDomainStatusRequest) (*GetDomainStatusResponse, error)
}

// GetDomainStatusRequest represents the request
type GetDomainStatusRequest struct {
	BrandID string // Agency ID (brand ID is agency_id)
}

// GetDomainStatusResponse represents the response
type GetDomainStatusResponse struct {
	Branding      *model.Branding
	Verified      bool   // Domain verification status from Vercel API
	ExpectedCNAME string // Expected CNAME target from Vercel API
	SSLStatus     string // SSL status from Vercel API: "pending", "active", "failed"
}

// GetDomainInstructions is the inbound port for getting DNS setup instructions
type GetDomainInstructions interface {
	Execute(ctx context.Context, req *GetDomainInstructionsRequest) (*GetDomainInstructionsResponse, error)
}

// GetDomainInstructionsRequest represents the request
type GetDomainInstructionsRequest struct {
	BrandID string // Agency ID (brand ID is agency_id)
}

// GetDomainInstructionsResponse represents the response
type GetDomainInstructionsResponse struct {
	Domain        string // Custom domain
	CNAMETarget   string // CNAME target from Vercel API (value may vary, don't hardcode)
	Instructions  string // Human-readable instructions for DNS setup
}
