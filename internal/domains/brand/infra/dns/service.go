package dns

import (
	"context"
	"net"

	"github.com/rs/zerolog"
)

// DNSService provides optional DNS lookup for UX feedback only
// DNS lookups are optional - only used to provide better UI feedback
// Vercel API is the source of truth - never use DNS lookups for verification decisions
type DNSService struct {
	logger zerolog.Logger
}

// NewDNSService creates a new DNS service for optional DNS lookups
func NewDNSService(logger zerolog.Logger) *DNSService {
	return &DNSService{
		logger: logger,
	}
}

// LookupCNAME performs optional DNS lookup using net.LookupCNAME
// Returns current CNAME record from DNS
// Used to compare with expected CNAME target (from Vercel API) for UX feedback
// Helps show "DNS record looks correct" vs "DNS record not found" in UI
// Never use this result for verification decisions - only for UI feedback
func (s *DNSService) LookupCNAME(ctx context.Context, domain string) (string, error) {
	cname, err := net.LookupCNAME(domain)
	if err != nil {
		s.logger.Debug().Err(err).Str("domain", domain).Msg("DNS lookup failed (optional, for UX only)")
		return "", err
	}

	// Remove trailing dot if present
	if len(cname) > 0 && cname[len(cname)-1] == '.' {
		cname = cname[:len(cname)-1]
	}

	return cname, nil
}

// CheckDNSPropagation verifies DNS changes have propagated (optional helper for UI feedback)
// Shows progress indicator while DNS propagates
// This is for UX feedback only, not for verification decisions
func (s *DNSService) CheckDNSPropagation(ctx context.Context, domain, expectedTarget string) (bool, error) {
	currentCNAME, err := s.LookupCNAME(ctx, domain)
	if err != nil {
		return false, err
	}

	// Compare current DNS record with expected target (for UX feedback)
	// Normalize both for comparison (lowercase, remove trailing dots)
	normalize := func(s string) string {
		result := s
		if len(result) > 0 && result[len(result)-1] == '.' {
			result = result[:len(result)-1]
		}
		return result
	}

	currentNormalized := normalize(currentCNAME)
	expectedNormalized := normalize(expectedTarget)

	return currentNormalized == expectedNormalized, nil
}
