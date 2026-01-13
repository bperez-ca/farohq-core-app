package vercel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// VercelService handles domain management via Vercel API
// Vercel API is the source of truth for all domain operations
type VercelService struct {
	apiToken  string
	projectID string
	teamID    string
	baseURL   string
	client    *http.Client
	logger    zerolog.Logger
}

// NewVercelService creates a new Vercel service
func NewVercelService(apiToken, projectID, teamID string, logger zerolog.Logger) *VercelService {
	return &VercelService{
		apiToken:  apiToken,
		projectID: projectID,
		teamID:    teamID,
		baseURL:   "https://api.vercel.com",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// DomainConfig represents expected DNS configuration from Vercel API
type DomainConfig struct {
	CNAMETarget string `json:"cname_target,omitempty"`
	RecordType  string `json:"record_type,omitempty"`
	RecordValue string `json:"record_value,omitempty"`
}

// DomainStatus represents domain status from Vercel API
type DomainStatus struct {
	Domain        string    `json:"name"`
	Verified      bool      `json:"verified"`
	Verification  []struct {
		Type  string `json:"type"`
		Domain string `json:"domain"`
		Value string `json:"value"`
	} `json:"verification"`
	Config []DomainConfig `json:"config"`
	SSL    struct {
		Status string `json:"status"` // "pending", "active", "failed"
	} `json:"ssl"`
}

// AddDomainToProject adds a custom domain to Vercel project
// Returns expected DNS configuration (CNAME target - value may vary, don't hardcode)
func (s *VercelService) AddDomainToProject(ctx context.Context, domain string) (*DomainConfig, error) {
	url := fmt.Sprintf("%s/v9/projects/%s/domains", s.baseURL, s.projectID)
	if s.teamID != "" {
		url += "?teamId=" + s.teamID
	}

	reqBody := map[string]string{
		"name": domain,
	}
	reqBodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Vercel API: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		s.logger.Error().
			Int("status", resp.StatusCode).
			Str("response", string(body)).
			Msg("Vercel API error adding domain")
		return nil, fmt.Errorf("vercel API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response to get expected DNS configuration
	var domainStatus DomainStatus
	if err := json.Unmarshal(body, &domainStatus); err != nil {
		s.logger.Error().Err(err).Msg("Failed to parse Vercel API response")
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract CNAME target from config (value may vary, don't hardcode)
	var cnameTarget string
	if len(domainStatus.Config) > 0 {
		config := domainStatus.Config[0]
		if config.RecordType == "CNAME" {
			cnameTarget = config.RecordValue
		} else if config.CNAMETarget != "" {
			cnameTarget = config.CNAMETarget
		}
	}

	if cnameTarget == "" {
		// Fallback: Vercel might return CNAME target in verification array
		for _, verif := range domainStatus.Verification {
			if verif.Type == "cname" {
				cnameTarget = verif.Value
				break
			}
		}
	}

	if cnameTarget == "" {
		s.logger.Warn().Str("domain", domain).Msg("No CNAME target found in Vercel API response")
		// Return a default that Vercel typically uses, but this should be fetched from API
		cnameTarget = "cname.vercel-dns.com"
	}

	return &DomainConfig{
		CNAMETarget: cnameTarget,
		RecordType:  "CNAME",
		RecordValue: cnameTarget,
	}, nil
}

// GetExpectedDNSConfig fetches expected DNS configuration from Vercel API
// Returns CNAME target and other DNS records required by Vercel
func (s *VercelService) GetExpectedDNSConfig(ctx context.Context, domain string) (*DomainConfig, error) {
	status, err := s.GetDomainStatus(ctx, domain)
	if err != nil {
		return nil, err
	}

	// Extract CNAME target from config (value may vary, don't hardcode)
	var cnameTarget string
	if len(status.Config) > 0 {
		config := status.Config[0]
		if config.RecordType == "CNAME" {
			cnameTarget = config.RecordValue
		} else if config.CNAMETarget != "" {
			cnameTarget = config.CNAMETarget
		}
	}

	if cnameTarget == "" {
		// Fallback: Vercel might return CNAME target in verification array
		for _, verif := range status.Verification {
			if verif.Type == "cname" {
				cnameTarget = verif.Value
				break
			}
		}
	}

	if cnameTarget == "" {
		return nil, fmt.Errorf("no CNAME target found in Vercel API response for domain %s", domain)
	}

	return &DomainConfig{
		CNAMETarget: cnameTarget,
		RecordType:  "CNAME",
		RecordValue: cnameTarget,
	}, nil
}

// VerifyDomain checks domain verification status via Vercel API
// Vercel API is authoritative for verification status
func (s *VercelService) VerifyDomain(ctx context.Context, domain string) (bool, error) {
	status, err := s.GetDomainStatus(ctx, domain)
	if err != nil {
		return false, err
	}
	return status.Verified, nil
}

// GetDomainStatus gets full domain status from Vercel API
// Returns domain verification status, SSL status, and other metadata
func (s *VercelService) GetDomainStatus(ctx context.Context, domain string) (*DomainStatus, error) {
	url := fmt.Sprintf("%s/v9/projects/%s/domains/%s", s.baseURL, s.projectID, domain)
	if s.teamID != "" {
		url += "?teamId=" + s.teamID
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Vercel API: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("domain not found in Vercel project")
	}

	if resp.StatusCode != http.StatusOK {
		s.logger.Error().
			Int("status", resp.StatusCode).
			Str("response", string(body)).
			Msg("Vercel API error getting domain status")
		return nil, fmt.Errorf("vercel API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var domainStatus DomainStatus
	if err := json.Unmarshal(body, &domainStatus); err != nil {
		s.logger.Error().Err(err).Msg("Failed to parse Vercel API response")
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &domainStatus, nil
}

// GetSSLStatus checks SSL certificate status via Vercel API
// Vercel automatically provisions SSL after DNS is correct and propagated
// SSL provisioning time varies (often minutes-hours, sometimes longer)
func (s *VercelService) GetSSLStatus(ctx context.Context, domain string) (string, error) {
	status, err := s.GetDomainStatus(ctx, domain)
	if err != nil {
		return "", err
	}
	return status.SSL.Status, nil
}

// RemoveDomain removes a domain from Vercel project
func (s *VercelService) RemoveDomain(ctx context.Context, domain string) error {
	url := fmt.Sprintf("%s/v9/projects/%s/domains/%s", s.baseURL, s.projectID, domain)
	if s.teamID != "" {
		url += "?teamId=" + s.teamID
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call Vercel API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error().
			Int("status", resp.StatusCode).
			Str("response", string(body)).
			Msg("Vercel API error removing domain")
		return fmt.Errorf("vercel API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListDomains lists all domains for a project
func (s *VercelService) ListDomains(ctx context.Context) ([]DomainStatus, error) {
	url := fmt.Sprintf("%s/v9/projects/%s/domains", s.baseURL, s.projectID)
	if s.teamID != "" {
		url += "?teamId=" + s.teamID
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Vercel API: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		s.logger.Error().
			Int("status", resp.StatusCode).
			Str("response", string(body)).
			Msg("Vercel API error listing domains")
		return nil, fmt.Errorf("vercel API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Domains []DomainStatus `json:"domains"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		s.logger.Error().Err(err).Msg("Failed to parse Vercel API response")
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Domains, nil
}
