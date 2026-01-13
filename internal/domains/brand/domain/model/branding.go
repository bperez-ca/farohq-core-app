package model

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DomainType represents the type of domain configuration
type DomainType string

const (
	DomainTypeSubdomain DomainType = "subdomain"
	DomainTypeCustom    DomainType = "custom"
)

// SSLStatus represents the SSL certificate status for custom domains
type SSLStatus string

const (
	SSLStatusPending SSLStatus = "pending"
	SSLStatusActive  SSLStatus = "active"
	SSLStatusFailed  SSLStatus = "failed"
)

// Branding represents branding configuration for a tenant
type Branding struct {
	agencyID                uuid.UUID
	domain                  string
	subdomain               string
	domainType              *DomainType
	website                 string
	verifiedAt              *time.Time
	logoURL                 string
	faviconURL              string
	primaryColor            string
	secondaryColor          string
	themeJSON               map[string]interface{}
	hidePoweredBy           bool
	emailDomain             string
	cloudflareZoneID        string
	domainVerificationToken string
	sslStatus               *SSLStatus
	updatedAt               time.Time
}

// NewBranding creates a new branding entity
func NewBranding(agencyID uuid.UUID, domain, subdomain string, domainType *DomainType, website, logoURL, faviconURL, primaryColor, secondaryColor string, themeJSON map[string]interface{}) *Branding {
	now := time.Now()
	if themeJSON == nil {
		themeJSON = make(map[string]interface{})
	}
	return &Branding{
		agencyID:                agencyID,
		domain:                  domain,
		subdomain:               subdomain,
		domainType:              domainType,
		website:                 website,
		verifiedAt:              nil,
		logoURL:                 logoURL,
		faviconURL:              faviconURL,
		primaryColor:            primaryColor,
		secondaryColor:          secondaryColor,
		themeJSON:               themeJSON,
		hidePoweredBy:           false,
		emailDomain:             "",
		cloudflareZoneID:        "",
		domainVerificationToken: "",
		sslStatus:               nil,
		updatedAt:               now,
	}
}

// NewBrandingWithID creates a branding entity with existing data
func NewBrandingWithID(
	agencyID uuid.UUID,
	domain, subdomain string,
	domainType *DomainType,
	website string,
	verifiedAt *time.Time,
	logoURL, faviconURL, primaryColor, secondaryColor string,
	themeJSON map[string]interface{},
	hidePoweredBy bool,
	emailDomain, cloudflareZoneID, domainVerificationToken string,
	sslStatus *SSLStatus,
	updatedAt time.Time,
) *Branding {
	if themeJSON == nil {
		themeJSON = make(map[string]interface{})
	}
	return &Branding{
		agencyID:                agencyID,
		domain:                  domain,
		subdomain:               subdomain,
		domainType:              domainType,
		website:                 website,
		verifiedAt:              verifiedAt,
		logoURL:                 logoURL,
		faviconURL:              faviconURL,
		primaryColor:            primaryColor,
		secondaryColor:          secondaryColor,
		themeJSON:               themeJSON,
		hidePoweredBy:           hidePoweredBy,
		emailDomain:             emailDomain,
		cloudflareZoneID:        cloudflareZoneID,
		domainVerificationToken: domainVerificationToken,
		sslStatus:               sslStatus,
		updatedAt:               updatedAt,
	}
}

// AgencyID returns the agency ID
func (b *Branding) AgencyID() uuid.UUID {
	return b.agencyID
}

// Domain returns the domain
func (b *Branding) Domain() string {
	return b.domain
}

// VerifiedAt returns the verification timestamp
func (b *Branding) VerifiedAt() *time.Time {
	return b.verifiedAt
}

// LogoURL returns the logo URL
func (b *Branding) LogoURL() string {
	return b.logoURL
}

// FaviconURL returns the favicon URL
func (b *Branding) FaviconURL() string {
	return b.faviconURL
}

// PrimaryColor returns the primary color
func (b *Branding) PrimaryColor() string {
	return b.primaryColor
}

// SecondaryColor returns the secondary color
func (b *Branding) SecondaryColor() string {
	return b.secondaryColor
}

// ThemeJSON returns the theme JSON
func (b *Branding) ThemeJSON() map[string]interface{} {
	return b.themeJSON
}

// UpdatedAt returns the update timestamp
func (b *Branding) UpdatedAt() time.Time {
	return b.updatedAt
}

// SetDomain sets the domain
func (b *Branding) SetDomain(domain string) {
	b.domain = domain
	b.updatedAt = time.Now()
}

// SetLogoURL sets the logo URL
func (b *Branding) SetLogoURL(logoURL string) {
	b.logoURL = logoURL
	b.updatedAt = time.Now()
}

// SetFaviconURL sets the favicon URL
func (b *Branding) SetFaviconURL(faviconURL string) {
	b.faviconURL = faviconURL
	b.updatedAt = time.Now()
}

// SetPrimaryColor sets the primary color
func (b *Branding) SetPrimaryColor(color string) {
	b.primaryColor = color
	b.updatedAt = time.Now()
}

// SetSecondaryColor sets the secondary color
func (b *Branding) SetSecondaryColor(color string) {
	b.secondaryColor = color
	b.updatedAt = time.Now()
}

// SetThemeJSON sets the theme JSON
func (b *Branding) SetThemeJSON(themeJSON map[string]interface{}) {
	if themeJSON == nil {
		themeJSON = make(map[string]interface{})
	}
	b.themeJSON = themeJSON
	b.updatedAt = time.Now()
}

// Website returns the agency website URL
func (b *Branding) Website() string {
	return b.website
}

// SetWebsite sets the agency website URL
func (b *Branding) SetWebsite(website string) {
	b.website = website
	b.updatedAt = time.Now()
}

// Subdomain returns the generated subdomain
func (b *Branding) Subdomain() string {
	return b.subdomain
}

// SetSubdomain sets the subdomain
func (b *Branding) SetSubdomain(subdomain string) {
	b.subdomain = subdomain
	b.updatedAt = time.Now()
}

// DomainType returns the domain type
func (b *Branding) DomainType() *DomainType {
	return b.domainType
}

// SetDomainType sets the domain type
func (b *Branding) SetDomainType(domainType *DomainType) {
	b.domainType = domainType
	b.updatedAt = time.Now()
}

// HidePoweredBy returns whether to hide "Powered by Faro" badge
func (b *Branding) HidePoweredBy() bool {
	return b.hidePoweredBy
}

// SetHidePoweredBy sets whether to hide "Powered by Faro" badge
func (b *Branding) SetHidePoweredBy(hidePoweredBy bool) {
	b.hidePoweredBy = hidePoweredBy
	b.updatedAt = time.Now()
}

// EmailDomain returns the email domain
func (b *Branding) EmailDomain() string {
	return b.emailDomain
}

// SetEmailDomain sets the email domain
func (b *Branding) SetEmailDomain(emailDomain string) {
	b.emailDomain = emailDomain
	b.updatedAt = time.Now()
}

// CloudflareZoneID returns the Cloudflare zone ID (for custom domains)
func (b *Branding) CloudflareZoneID() string {
	return b.cloudflareZoneID
}

// SetCloudflareZoneID sets the Cloudflare zone ID
func (b *Branding) SetCloudflareZoneID(cloudflareZoneID string) {
	b.cloudflareZoneID = cloudflareZoneID
	b.updatedAt = time.Now()
}

// DomainVerificationToken returns the domain verification token
func (b *Branding) DomainVerificationToken() string {
	return b.domainVerificationToken
}

// SetDomainVerificationToken sets the domain verification token
func (b *Branding) SetDomainVerificationToken(token string) {
	b.domainVerificationToken = token
	b.updatedAt = time.Now()
}

// SSLStatus returns the SSL status (for custom domains)
func (b *Branding) SSLStatus() *SSLStatus {
	return b.sslStatus
}

// SetSSLStatus sets the SSL status
func (b *Branding) SetSSLStatus(status *SSLStatus) {
	b.sslStatus = status
	b.updatedAt = time.Now()
}

// Verify marks the branding as verified
func (b *Branding) Verify() {
	now := time.Now()
	b.verifiedAt = &now
	b.updatedAt = now
}

// GenerateSubdomain generates a subdomain from agency slug
// Format: {slugified-agency-slug}.app.farohq.com
func GenerateSubdomain(agencySlug string) string {
	// Ensure slug is lowercase and URL-safe
	slug := strings.ToLower(agencySlug)
	slug = strings.TrimSpace(slug)
	// Replace spaces and special characters with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	// Keep only alphanumeric and hyphens
	var builder strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			builder.WriteRune(r)
		}
	}
	slug = builder.String()
	// Remove consecutive hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	// Trim leading/trailing hyphens
	slug = strings.Trim(slug, "-")
	// Ensure non-empty
	if slug == "" {
		slug = "agency"
	}
	return slug + ".app.farohq.com"
}// GenerateSubdomainWithFallback generates a subdomain with uniqueness check fallback
// If base subdomain exists, appends a number (e.g., agency-2.app.farohq.com)
func GenerateSubdomainWithFallback(agencySlug string, checkExists func(subdomain string) bool) string {
	baseSubdomain := GenerateSubdomain(agencySlug)
	if !checkExists(baseSubdomain) {
		return baseSubdomain
	}
	// Extract base slug (without .app.farohq.com)
	baseSlug := strings.TrimSuffix(baseSubdomain, ".app.farohq.com")
	// Try with numbers (2-1000)
	for i := 2; i <= 1000; i++ {
		candidate := baseSlug + "-" + strconv.Itoa(i) + ".app.farohq.com"
		if !checkExists(candidate) {
			return candidate
		}
	}
	// Fallback: append timestamp if all numbers are taken (unlikely)
	return baseSlug + "-" + strconv.FormatInt(time.Now().Unix(), 10) + ".app.farohq.com"
}

// GenerateSubdomainFromWebsite extracts domain from website URL and generates subdomain
// Example: "https://www.example.com" -> "example.app.farohq.com"
func GenerateSubdomainFromWebsite(websiteURL string) string {
	if websiteURL == "" {
		return ""
	}

	// Parse URL
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		// If parsing fails, try to extract domain manually
		domain := websiteURL
		domain = strings.TrimPrefix(domain, "http://")
		domain = strings.TrimPrefix(domain, "https://")
		domain = strings.TrimPrefix(domain, "www.")
		domain = strings.Split(domain, "/")[0]
		domain = strings.Split(domain, ":")[0]
		slug := generateSlugFromText(domain)
		if slug == "" {
			return ""
		}
		return slug + ".app.farohq.com"
	}

	host := parsedURL.Hostname()
	if host == "" {
		return ""
	}

	// Remove www. prefix
	host = strings.TrimPrefix(host, "www.")

	// Generate slug from domain
	slug := generateSlugFromText(host)
	if slug == "" {
		return ""
	}

	return slug + ".app.farohq.com"
}

// generateSlugFromText generates a URL-safe slug from text
func generateSlugFromText(text string) string {
	slug := strings.ToLower(text)
	slug = strings.TrimSpace(slug)
	slug = strings.ReplaceAll(slug, " ", "-")
	
	// Keep only alphanumeric and hyphens
	var builder strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			builder.WriteRune(r)
		}
	}
	slug = builder.String()

	// Remove consecutive hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Trim leading/trailing hyphens
	slug = strings.Trim(slug, "-")

	// Ensure non-empty
	if slug == "" {
		return ""
	}
	return slug
}
