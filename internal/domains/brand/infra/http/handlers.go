package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"farohq-core-app/internal/domains/brand/domain"
	"farohq-core-app/internal/domains/brand/domain/model"
	"farohq-core-app/internal/domains/brand/domain/ports/inbound"
	tenants_model "farohq-core-app/internal/domains/tenants/domain/model"
	tenants_outbound "farohq-core-app/internal/domains/tenants/domain/ports/outbound"
	"farohq-core-app/internal/platform/tenant"
)

// Handlers provides HTTP handlers for the brand domain
type Handlers struct {
	logger              zerolog.Logger
	getByDomain         inbound.GetByDomain
	getByHost           inbound.GetByHost
	listBrands          inbound.ListBrands
	createBrand         inbound.CreateBrand
	getBrand            inbound.GetBrand
	updateBrand         inbound.UpdateBrand
	deleteBrand         inbound.DeleteBrand
	verifyDomain        inbound.VerifyDomain
	getDomainStatus     inbound.GetDomainStatus
	getDomainInstructions inbound.GetDomainInstructions
	tenantRepo          tenants_outbound.TenantRepository // For tier-based flags in responses
}

// NewHandlers creates new brand HTTP handlers
func NewHandlers(
	logger zerolog.Logger,
	getByDomain inbound.GetByDomain,
	getByHost inbound.GetByHost,
	listBrands inbound.ListBrands,
	createBrand inbound.CreateBrand,
	getBrand inbound.GetBrand,
	updateBrand inbound.UpdateBrand,
	deleteBrand inbound.DeleteBrand,
	verifyDomain inbound.VerifyDomain,
	getDomainStatus inbound.GetDomainStatus,
	getDomainInstructions inbound.GetDomainInstructions,
	tenantRepo tenants_outbound.TenantRepository,
) *Handlers {
	return &Handlers{
		logger:              logger,
		getByDomain:         getByDomain,
		getByHost:           getByHost,
		listBrands:          listBrands,
		createBrand:         createBrand,
		getBrand:            getBrand,
		updateBrand:         updateBrand,
		deleteBrand:         deleteBrand,
		verifyDomain:        verifyDomain,
		getDomainStatus:     getDomainStatus,
		getDomainInstructions: getDomainInstructions,
		tenantRepo:          tenantRepo,
	}
}

// GetByDomainHandler handles GET /api/v1/brand/by-domain
func (h *Handlers) GetByDomainHandler(w http.ResponseWriter, r *http.Request) {
	domainParam := r.URL.Query().Get("domain")
	if domainParam == "" {
		http.Error(w, "domain parameter is required", http.StatusBadRequest)
		return
	}

	req := &inbound.GetByDomainRequest{
		Domain: domainParam,
	}

	resp, err := h.getByDomain.Execute(r.Context(), req)
	if err != nil {
		if err == domain.ErrBrandingNotFound {
			http.Error(w, "Branding not found", http.StatusNotFound)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to get branding by domain")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.buildBrandResponse(r.Context(), resp.Branding))
}

// GetByHostHandler handles GET /api/v1/brand/by-host
func (h *Handlers) GetByHostHandler(w http.ResponseWriter, r *http.Request) {
	host := r.URL.Query().Get("host")
	if host == "" {
		http.Error(w, "host parameter is required", http.StatusBadRequest)
		return
	}

	req := &inbound.GetByHostRequest{
		Host: host,
	}

	resp, err := h.getByHost.Execute(r.Context(), req)
	if err != nil {
		if err == domain.ErrBrandingNotFound {
			http.Error(w, "Branding not found", http.StatusNotFound)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to get branding by host")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.buildBrandResponse(r.Context(), resp.Branding))
}

// buildBrandResponse converts a Branding entity to a map for JSON encoding with tier-based flags
func (h *Handlers) buildBrandResponse(ctx context.Context, branding *model.Branding) map[string]interface{} {
	var verifiedAt *string
	if branding.VerifiedAt() != nil {
		s := branding.VerifiedAt().Format("2006-01-02T15:04:05Z07:00")
		verifiedAt = &s
	}

	var domainType *string
	if branding.DomainType() != nil {
		dt := string(*branding.DomainType())
		domainType = &dt
	}

	var sslStatus *string
	if branding.SSLStatus() != nil {
		ss := string(*branding.SSLStatus())
		sslStatus = &ss
	}

	// Get tier information for tier-based flags
	var canHidePoweredBy, canConfigureDomain bool
	tenant, err := h.tenantRepo.FindByID(ctx, branding.AgencyID())
	if err == nil && tenant != nil {
		tier := tenant.Tier()
		canHidePoweredBy = tenants_model.TierCanHidePoweredBy(tier)
		canConfigureDomain = tenants_model.TierSupportsCustomDomain(tier)
	}

	response := map[string]interface{}{
		"agency_id":            branding.AgencyID().String(),
		"domain":               branding.Domain(),
		"subdomain":            branding.Subdomain(),
		"domain_type":          domainType,
		"website":              branding.Website(),
		"verified_at":          verifiedAt,
		"logo_url":             branding.LogoURL(),
		"favicon_url":          branding.FaviconURL(),
		"primary_color":        branding.PrimaryColor(),
		"secondary_color":      branding.SecondaryColor(),
		"theme_json":           branding.ThemeJSON(),
		"hide_powered_by":      branding.HidePoweredBy(),
		"can_hide_powered_by":  canHidePoweredBy,
		"can_configure_domain": canConfigureDomain,
		"email_domain":         branding.EmailDomain(),
		"ssl_status":           sslStatus,
		"updated_at":           branding.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	return response
}

// ListBrandsHandler handles GET /api/v1/brands
func (h *Handlers) ListBrandsHandler(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from context
	tenantID, ok := tenant.GetTenantFromContext(r.Context())
	if !ok {
		http.Error(w, "Failed to resolve tenant. Provide X-Tenant-ID header or use a tenant domain.", http.StatusBadRequest)
		return
	}

	req := &inbound.ListBrandsRequest{
		AgencyID: tenantID,
	}

	resp, err := h.listBrands.Execute(r.Context(), req)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list brands")
		http.Error(w, "Failed to query brands", http.StatusInternalServerError)
		return
	}

	brands := make([]map[string]interface{}, len(resp.Brands))
	for i, branding := range resp.Brands {
		brands[i] = h.buildBrandResponse(r.Context(), branding)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(brands)
}

// CreateBrandHandler handles POST /api/v1/brands
func (h *Handlers) CreateBrandHandler(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from context
	tenantID, ok := tenant.GetTenantFromContext(r.Context())
	if !ok {
		http.Error(w, "Failed to resolve tenant. Provide X-Tenant-ID header or use a tenant domain.", http.StatusBadRequest)
		return
	}

	var req struct {
		Domain         string                 `json:"domain"`          // Optional: Custom domain (Scale tier only)
		Website        string                 `json:"website"`         // Optional: Agency website URL
		LogoURL        string                 `json:"logo_url"`
		FaviconURL     string                 `json:"favicon_url"`
		PrimaryColor   string                 `json:"primary_color"`
		SecondaryColor string                 `json:"secondary_color"`
		ThemeJSON      map[string]interface{} `json:"theme_json"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	createReq := &inbound.CreateBrandRequest{
		AgencyID:       tenantID,
		Domain:         req.Domain,  // Optional: Only used for Scale tier
		Website:        req.Website, // Optional: Captured for future custom domain integration
		LogoURL:        req.LogoURL,
		FaviconURL:     req.FaviconURL,
		PrimaryColor:   req.PrimaryColor,
		SecondaryColor: req.SecondaryColor,
		ThemeJSON:      req.ThemeJSON,
	}

	resp, err := h.createBrand.Execute(r.Context(), createReq)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create/update brand")
		http.Error(w, "Failed to create brand", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(h.buildBrandResponse(r.Context(), resp.Branding))
}

// GetBrandHandler handles GET /api/v1/brands/{brandId}
func (h *Handlers) GetBrandHandler(w http.ResponseWriter, r *http.Request) {
	brandID := chi.URLParam(r, "brandId")
	if brandID == "" {
		http.Error(w, "brand ID is required", http.StatusBadRequest)
		return
	}

	req := &inbound.GetBrandRequest{
		BrandID: brandID,
	}

	resp, err := h.getBrand.Execute(r.Context(), req)
	if err != nil {
		if err == domain.ErrBrandingNotFound {
			http.Error(w, "Brand not found", http.StatusNotFound)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to get brand")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.buildBrandResponse(r.Context(), resp.Branding))
}

// UpdateBrandHandler handles PUT /api/v1/brands/{brandId}
func (h *Handlers) UpdateBrandHandler(w http.ResponseWriter, r *http.Request) {
	brandID := chi.URLParam(r, "brandId")
	if brandID == "" {
		http.Error(w, "brand ID is required", http.StatusBadRequest)
		return
	}

	var req struct {
		Domain         *string                `json:"domain"`          // Optional: Custom domain (Scale tier only)
		Website        *string                `json:"website"`         // Optional: Agency website URL
		LogoURL        *string                `json:"logo_url"`
		FaviconURL     *string                `json:"favicon_url"`
		PrimaryColor   *string                `json:"primary_color"`
		SecondaryColor *string                `json:"secondary_color"`
		ThemeJSON      *map[string]interface{} `json:"theme_json"`
		HidePoweredBy  *bool                   `json:"hide_powered_by"` // Optional: Hide "Powered by Faro" badge (Growth+ tiers only)
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	updateReq := &inbound.UpdateBrandRequest{
		BrandID:        brandID,
		Domain:         req.Domain,
		Website:        req.Website,
		LogoURL:        req.LogoURL,
		FaviconURL:     req.FaviconURL,
		PrimaryColor:   req.PrimaryColor,
		SecondaryColor: req.SecondaryColor,
		ThemeJSON:      req.ThemeJSON,
		HidePoweredBy:  req.HidePoweredBy,
	}

	resp, err := h.updateBrand.Execute(r.Context(), updateReq)
	if err != nil {
		if err == domain.ErrBrandingNotFound {
			http.Error(w, "Brand not found", http.StatusNotFound)
			return
		}
		// Check for tier-related errors
		errMsg := err.Error()
		if errMsg == "Custom domain support is only available for Scale tier" {
			http.Error(w, errMsg, http.StatusForbidden)
			return
		}
		if errMsg == "Hide 'Powered by Faro' badge is only available for Growth+ tiers" {
			http.Error(w, errMsg, http.StatusForbidden)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to update brand")
		http.Error(w, "Failed to update brand", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.buildBrandResponse(r.Context(), resp.Branding))
}

// DeleteBrandHandler handles DELETE /api/v1/brands/{brandId}
func (h *Handlers) DeleteBrandHandler(w http.ResponseWriter, r *http.Request) {
	brandID := chi.URLParam(r, "brandId")
	if brandID == "" {
		http.Error(w, "brand ID is required", http.StatusBadRequest)
		return
	}

	req := &inbound.DeleteBrandRequest{
		BrandID: brandID,
	}

	resp, err := h.deleteBrand.Execute(r.Context(), req)
	if err != nil {
		if err == domain.ErrBrandingNotFound {
			http.Error(w, "Brand not found", http.StatusNotFound)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to delete brand")
		http.Error(w, "Failed to delete brand", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": resp.Success,
	})
}

// VerifyDomainHandler handles POST /api/v1/brands/{brandId}/verify-domain
func (h *Handlers) VerifyDomainHandler(w http.ResponseWriter, r *http.Request) {
	brandID := chi.URLParam(r, "brandId")
	if brandID == "" {
		http.Error(w, "brand ID is required", http.StatusBadRequest)
		return
	}

	var req struct {
		Domain string `json:"domain"` // Optional: Custom domain to verify (if not provided, uses domain from brand)
	}

	// Allow empty body (use domain from brand if not provided)
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
	}

	verifyReq := &inbound.VerifyDomainRequest{
		BrandID: brandID,
		Domain:  req.Domain,
	}

	resp, err := h.verifyDomain.Execute(r.Context(), verifyReq)
	if err != nil {
		if err == domain.ErrBrandingNotFound {
			http.Error(w, "Brand not found", http.StatusNotFound)
			return
		}
		// Check for tier-related errors
		if err.Error() == "Custom domain support is only available for Scale tier" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to verify domain")
		http.Error(w, "Failed to verify domain", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"branding":       h.buildBrandResponse(r.Context(), resp.Branding),
		"verified":       resp.Verified,
		"expected_cname": resp.ExpectedCNAME,
		"current_cname":  resp.CurrentCNAME, // Optional, for UX feedback only
		"ssl_status":     resp.SSLStatus,
	})
}

// GetDomainStatusHandler handles GET /api/v1/brands/{brandId}/domain-status
func (h *Handlers) GetDomainStatusHandler(w http.ResponseWriter, r *http.Request) {
	brandID := chi.URLParam(r, "brandId")
	if brandID == "" {
		http.Error(w, "brand ID is required", http.StatusBadRequest)
		return
	}

	req := &inbound.GetDomainStatusRequest{
		BrandID: brandID,
	}

	resp, err := h.getDomainStatus.Execute(r.Context(), req)
	if err != nil {
		if err == domain.ErrBrandingNotFound {
			http.Error(w, "Brand not found", http.StatusNotFound)
			return
		}
		// Check for tier-related errors
		if err.Error() == "Custom domain support is only available for Scale tier" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to get domain status")
		http.Error(w, "Failed to get domain status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"branding":       h.buildBrandResponse(r.Context(), resp.Branding),
		"verified":       resp.Verified,
		"expected_cname": resp.ExpectedCNAME,
		"ssl_status":     resp.SSLStatus,
	})
}

// GetDomainInstructionsHandler handles GET /api/v1/brands/{brandId}/domain-instructions
func (h *Handlers) GetDomainInstructionsHandler(w http.ResponseWriter, r *http.Request) {
	brandID := chi.URLParam(r, "brandId")
	if brandID == "" {
		http.Error(w, "brand ID is required", http.StatusBadRequest)
		return
	}

	req := &inbound.GetDomainInstructionsRequest{
		BrandID: brandID,
	}

	resp, err := h.getDomainInstructions.Execute(r.Context(), req)
	if err != nil {
		if err == domain.ErrBrandingNotFound {
			http.Error(w, "Brand not found", http.StatusNotFound)
			return
		}
		// Check for tier-related errors
		if err.Error() == "Custom domain support is only available for Scale tier" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to get domain instructions")
		http.Error(w, "Failed to get domain instructions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"domain":        resp.Domain,
		"cname_target":  resp.CNAMETarget,
		"instructions":  resp.Instructions,
	})
}

// GetBySubdomainHandler handles GET /api/v1/brands/by-subdomain?subdomain={subdomain}
func (h *Handlers) GetBySubdomainHandler(w http.ResponseWriter, r *http.Request) {
	subdomain := r.URL.Query().Get("subdomain")
	if subdomain == "" {
		http.Error(w, "subdomain parameter is required", http.StatusBadRequest)
		return
	}

	// Note: This endpoint would need a new use case for subdomain lookup
	// For now, we can use getByHost which handles subdomain resolution
	req := &inbound.GetByHostRequest{
		Host: subdomain,
	}

	resp, err := h.getByHost.Execute(r.Context(), req)
	if err != nil {
		if err == domain.ErrBrandingNotFound {
			http.Error(w, "Branding not found", http.StatusNotFound)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to get branding by subdomain")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.buildBrandResponse(r.Context(), resp.Branding))
}

// GetSSLStatusHandler handles GET /api/v1/brands/{brandId}/ssl-status
func (h *Handlers) GetSSLStatusHandler(w http.ResponseWriter, r *http.Request) {
	// GetDomainStatus already returns SSL status, so we can reuse it
	// Or we can call GetDomainStatus and extract SSL status
	h.GetDomainStatusHandler(w, r)
}
