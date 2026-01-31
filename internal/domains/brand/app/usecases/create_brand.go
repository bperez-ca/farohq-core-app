package usecases

import (
	"context"

	"farohq-core-app/internal/domains/brand/domain/model"
	"farohq-core-app/internal/domains/brand/domain/ports/inbound"
	"farohq-core-app/internal/domains/brand/domain/ports/outbound"
	tenants_outbound "farohq-core-app/internal/domains/tenants/domain/ports/outbound"
	tenants_model "farohq-core-app/internal/domains/tenants/domain/model"

	"github.com/google/uuid"
)

// CreateBrand implements the CreateBrand inbound port
type CreateBrand struct {
	brandRepo  outbound.BrandRepository
	tenantRepo tenants_outbound.TenantRepository
}

// NewCreateBrand creates a new CreateBrand use case
func NewCreateBrand(brandRepo outbound.BrandRepository, tenantRepo tenants_outbound.TenantRepository) inbound.CreateBrand {
	return &CreateBrand{
		brandRepo:  brandRepo,
		tenantRepo: tenantRepo,
	}
}

// Execute executes the use case
func (uc *CreateBrand) Execute(ctx context.Context, req *inbound.CreateBrandRequest) (*inbound.CreateBrandResponse, error) {
	agencyID, err := uuid.Parse(req.AgencyID)
	if err != nil {
		return nil, err
	}

	// Get tenant/agency to check tier
	tenant, err := uc.tenantRepo.FindByID(ctx, agencyID)
	if err != nil {
		return nil, err
	}

	// Determine domain configuration based on tier
	var domainType *model.DomainType
	var subdomain string
	var customDomain string

	tier := tenant.Tier()
	if tier != nil && *tier == tenants_model.TierScale {
		// Scale tier: Can use custom domain OR subdomain
		if req.Domain != "" {
			// Custom domain provided
			dt := model.DomainTypeCustom
			domainType = &dt
			customDomain = req.Domain
			// Generate subdomain as fallback (optional)
			subdomain = model.GenerateSubdomain(tenant.Slug())
		} else {
			// No custom domain: Generate subdomain as fallback
			dt := model.DomainTypeSubdomain
			domainType = &dt
			subdomain = model.GenerateSubdomainWithFallback(tenant.Slug(), func(s string) bool {
				exists, _ := uc.brandRepo.CheckSubdomainExists(ctx, s)
				return exists
			})
		}
	} else {
		// Lower tiers (Starter, Growth): Always use subdomain
		dt := model.DomainTypeSubdomain
		domainType = &dt
		// Use website to generate subdomain if provided, otherwise use tenant slug
		if req.Website != "" {
			subdomain = model.GenerateSubdomainFromWebsite(req.Website)
			if subdomain == "" {
				// Fallback to tenant slug if website parsing fails
				subdomain = model.GenerateSubdomainWithFallback(tenant.Slug(), func(s string) bool {
					exists, _ := uc.brandRepo.CheckSubdomainExists(ctx, s)
					return exists
				})
			} else {
				// Check uniqueness and add fallback if needed
				exists, _ := uc.brandRepo.CheckSubdomainExists(ctx, subdomain)
				if exists {
					subdomain = model.GenerateSubdomainWithFallback(tenant.Slug(), func(s string) bool {
						exists, _ := uc.brandRepo.CheckSubdomainExists(ctx, s)
						return exists
					})
				}
			}
		} else {
			subdomain = model.GenerateSubdomainWithFallback(tenant.Slug(), func(s string) bool {
				exists, _ := uc.brandRepo.CheckSubdomainExists(ctx, s)
				return exists
			})
		}
		// Lower tiers cannot configure custom domains
		// Use empty string which will be converted to NULL in the repository
		customDomain = ""
	}

	// Convert empty string to empty string (will be handled as NULL in repository)
	// The repository layer will convert empty strings to NULL for database storage
	branding := model.NewBranding(
		agencyID,
		customDomain,
		subdomain,
		domainType,
		req.Website, // Optional website (may be empty)
		req.LogoURL,
		req.FaviconURL,
		req.PrimaryColor,
		req.SecondaryColor,
		req.ThemeJSON,
	)

	if err := uc.brandRepo.Save(ctx, branding); err != nil {
		return nil, err
	}

	return &inbound.CreateBrandResponse{
		Branding: branding,
	}, nil
}

