package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"farohq-core-app/internal/domains/brand/domain"
	"farohq-core-app/internal/domains/brand/domain/ports/inbound"
	"farohq-core-app/internal/platform/tenant"
)

// Handlers provides HTTP handlers for the brand domain
type Handlers struct {
	logger      zerolog.Logger
	getByDomain inbound.GetByDomain
	getByHost   inbound.GetByHost
	listBrands  inbound.ListBrands
	createBrand inbound.CreateBrand
	getBrand    inbound.GetBrand
	updateBrand inbound.UpdateBrand
	deleteBrand inbound.DeleteBrand
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
) *Handlers {
	return &Handlers{
		logger:      logger,
		getByDomain: getByDomain,
		getByHost:   getByHost,
		listBrands:  listBrands,
		createBrand: createBrand,
		getBrand:    getBrand,
		updateBrand: updateBrand,
		deleteBrand: deleteBrand,
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
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agency_id":       resp.Branding.AgencyID().String(),
		"domain":          resp.Branding.Domain(),
		"verified_at":    resp.Branding.VerifiedAt(),
		"logo_url":        resp.Branding.LogoURL(),
		"favicon_url":     resp.Branding.FaviconURL(),
		"primary_color":   resp.Branding.PrimaryColor(),
		"secondary_color": resp.Branding.SecondaryColor(),
		"theme_json":      resp.Branding.ThemeJSON(),
		"updated_at":      resp.Branding.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	})
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
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agency_id":       resp.Branding.AgencyID().String(),
		"domain":          resp.Branding.Domain(),
		"verified_at":    resp.Branding.VerifiedAt(),
		"logo_url":        resp.Branding.LogoURL(),
		"favicon_url":     resp.Branding.FaviconURL(),
		"primary_color":   resp.Branding.PrimaryColor(),
		"secondary_color": resp.Branding.SecondaryColor(),
		"theme_json":      resp.Branding.ThemeJSON(),
		"updated_at":      resp.Branding.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	})
}

// ListBrandsHandler handles GET /api/v1/brands
func (h *Handlers) ListBrandsHandler(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from context
	tenantID, ok := tenant.GetTenantFromContext(r.Context())
	if !ok {
		http.Error(w, "tenant context required", http.StatusBadRequest)
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
		var verifiedAt *string
		if branding.VerifiedAt() != nil {
			s := branding.VerifiedAt().Format("2006-01-02T15:04:05Z07:00")
			verifiedAt = &s
		}
		brands[i] = map[string]interface{}{
			"agency_id":       branding.AgencyID().String(),
			"domain":          branding.Domain(),
			"verified_at":     verifiedAt,
			"logo_url":         branding.LogoURL(),
			"favicon_url":      branding.FaviconURL(),
			"primary_color":    branding.PrimaryColor(),
			"secondary_color":  branding.SecondaryColor(),
			"theme_json":       branding.ThemeJSON(),
			"updated_at":       branding.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(brands)
}

// CreateBrandHandler handles POST /api/v1/brands
func (h *Handlers) CreateBrandHandler(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from context
	tenantID, ok := tenant.GetTenantFromContext(r.Context())
	if !ok {
		http.Error(w, "tenant context required", http.StatusBadRequest)
		return
	}

	var req struct {
		Domain         string                 `json:"domain"`
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
		Domain:         req.Domain,
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

	var verifiedAt *string
	if resp.Branding.VerifiedAt() != nil {
		s := resp.Branding.VerifiedAt().Format("2006-01-02T15:04:05Z07:00")
		verifiedAt = &s
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agency_id":       resp.Branding.AgencyID().String(),
		"domain":          resp.Branding.Domain(),
		"verified_at":    verifiedAt,
		"logo_url":        resp.Branding.LogoURL(),
		"favicon_url":     resp.Branding.FaviconURL(),
		"primary_color":   resp.Branding.PrimaryColor(),
		"secondary_color": resp.Branding.SecondaryColor(),
		"theme_json":      resp.Branding.ThemeJSON(),
		"updated_at":      resp.Branding.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	})
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

	var verifiedAt *string
	if resp.Branding.VerifiedAt() != nil {
		s := resp.Branding.VerifiedAt().Format("2006-01-02T15:04:05Z07:00")
		verifiedAt = &s
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agency_id":       resp.Branding.AgencyID().String(),
		"domain":          resp.Branding.Domain(),
		"verified_at":    verifiedAt,
		"logo_url":        resp.Branding.LogoURL(),
		"favicon_url":     resp.Branding.FaviconURL(),
		"primary_color":   resp.Branding.PrimaryColor(),
		"secondary_color": resp.Branding.SecondaryColor(),
		"theme_json":      resp.Branding.ThemeJSON(),
		"updated_at":      resp.Branding.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	})
}

// UpdateBrandHandler handles PUT /api/v1/brands/{brandId}
func (h *Handlers) UpdateBrandHandler(w http.ResponseWriter, r *http.Request) {
	brandID := chi.URLParam(r, "brandId")
	if brandID == "" {
		http.Error(w, "brand ID is required", http.StatusBadRequest)
		return
	}

	var req struct {
		Domain         *string                `json:"domain"`
		LogoURL        *string                `json:"logo_url"`
		FaviconURL     *string                `json:"favicon_url"`
		PrimaryColor   *string                `json:"primary_color"`
		SecondaryColor *string                `json:"secondary_color"`
		ThemeJSON      *map[string]interface{} `json:"theme_json"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	updateReq := &inbound.UpdateBrandRequest{
		BrandID:        brandID,
		Domain:         req.Domain,
		LogoURL:        req.LogoURL,
		FaviconURL:     req.FaviconURL,
		PrimaryColor:   req.PrimaryColor,
		SecondaryColor: req.SecondaryColor,
		ThemeJSON:      req.ThemeJSON,
	}

	resp, err := h.updateBrand.Execute(r.Context(), updateReq)
	if err != nil {
		if err == domain.ErrBrandingNotFound {
			http.Error(w, "Brand not found", http.StatusNotFound)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to update brand")
		http.Error(w, "Failed to update brand", http.StatusInternalServerError)
		return
	}

	var verifiedAt *string
	if resp.Branding.VerifiedAt() != nil {
		s := resp.Branding.VerifiedAt().Format("2006-01-02T15:04:05Z07:00")
		verifiedAt = &s
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agency_id":       resp.Branding.AgencyID().String(),
		"domain":          resp.Branding.Domain(),
		"verified_at":    verifiedAt,
		"logo_url":        resp.Branding.LogoURL(),
		"favicon_url":     resp.Branding.FaviconURL(),
		"primary_color":   resp.Branding.PrimaryColor(),
		"secondary_color": resp.Branding.SecondaryColor(),
		"theme_json":      resp.Branding.ThemeJSON(),
		"updated_at":      resp.Branding.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	})
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

