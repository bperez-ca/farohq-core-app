package model

import (
	"time"

	"github.com/google/uuid"
)

// Branding represents branding configuration for a tenant
type Branding struct {
	agencyID      uuid.UUID
	domain        string
	verifiedAt    *time.Time
	logoURL       string
	faviconURL    string
	primaryColor  string
	secondaryColor string
	themeJSON     map[string]interface{}
	updatedAt     time.Time
}

// NewBranding creates a new branding entity
func NewBranding(agencyID uuid.UUID, domain, logoURL, faviconURL, primaryColor, secondaryColor string, themeJSON map[string]interface{}) *Branding {
	now := time.Now()
	if themeJSON == nil {
		themeJSON = make(map[string]interface{})
	}
	return &Branding{
		agencyID:       agencyID,
		domain:         domain,
		verifiedAt:     nil,
		logoURL:        logoURL,
		faviconURL:     faviconURL,
		primaryColor:   primaryColor,
		secondaryColor: secondaryColor,
		themeJSON:      themeJSON,
		updatedAt:      now,
	}
}

// NewBrandingWithID creates a branding entity with existing data
func NewBrandingWithID(agencyID uuid.UUID, domain string, verifiedAt *time.Time, logoURL, faviconURL, primaryColor, secondaryColor string, themeJSON map[string]interface{}, updatedAt time.Time) *Branding {
	if themeJSON == nil {
		themeJSON = make(map[string]interface{})
	}
	return &Branding{
		agencyID:       agencyID,
		domain:         domain,
		verifiedAt:     verifiedAt,
		logoURL:        logoURL,
		faviconURL:     faviconURL,
		primaryColor:   primaryColor,
		secondaryColor: secondaryColor,
		themeJSON:      themeJSON,
		updatedAt:      updatedAt,
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

// Verify marks the branding as verified
func (b *Branding) Verify() {
	now := time.Now()
	b.verifiedAt = &now
	b.updatedAt = now
}

