-- Branding Features Migration: Add white-label branding enhancements
-- Multi-tenant white-label branding with subdomain and custom domain support
--
-- This migration adds:
-- 1. Website field (optional, captured during onboarding)
-- 2. Subdomain field (generated for lower tiers: {slug}.portal.farohq.com)
-- 3. Domain type field ('subdomain', 'custom', NULL)
-- 4. Hide "Powered by Faro" badge field (tier-based)
-- 5. Domain verification fields (for custom domains only)
-- 6. SSL status tracking (for custom domains only)
-- 7. Email domain field (for future use, separate from A2)

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Add website field (optional - some agencies may not have a website yet)
ALTER TABLE branding ADD COLUMN IF NOT EXISTS website TEXT;

-- Add subdomain field (generated for lower tiers)
-- Format: {agency-slug}.portal.farohq.com
ALTER TABLE branding ADD COLUMN IF NOT EXISTS subdomain TEXT;

-- Add domain_type field ('subdomain', 'custom', NULL)
-- 'subdomain': Lower tiers using {slug}.portal.farohq.com
-- 'custom': Scale tier using custom domain via Vercel
-- NULL: Not configured yet
ALTER TABLE branding ADD COLUMN IF NOT EXISTS domain_type TEXT CHECK (domain_type IN ('subdomain', 'custom'));

-- Add hide_powered_by field (tier-based: Growth+ tiers only)
ALTER TABLE branding ADD COLUMN IF NOT EXISTS hide_powered_by BOOLEAN DEFAULT false;

-- Add email_domain field (nullable, separate from A2 but useful for future)
ALTER TABLE branding ADD COLUMN IF NOT EXISTS email_domain TEXT;

-- Add domain verification fields (for custom domains only, Scale tier)
-- These are only used when domain_type = 'custom'
ALTER TABLE branding ADD COLUMN IF NOT EXISTS cloudflare_zone_id TEXT;
ALTER TABLE branding ADD COLUMN IF NOT EXISTS domain_verification_token TEXT;

-- Add SSL status field (for custom domains only, Scale tier)
-- 'pending', 'active', 'failed' - tracked via Vercel API
ALTER TABLE branding ADD COLUMN IF NOT EXISTS ssl_status TEXT CHECK (ssl_status IN ('pending', 'active', 'failed'));

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_branding_subdomain ON branding(subdomain) WHERE subdomain IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_branding_domain_type ON branding(domain_type) WHERE domain_type IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_branding_website ON branding(website) WHERE website IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_branding_hide_powered_by ON branding(hide_powered_by) WHERE hide_powered_by = true;

-- Update existing domain index to be conditional (only non-null domains)
-- Note: Existing index idx_branding_domain already exists, but we ensure it works correctly
CREATE INDEX IF NOT EXISTS idx_branding_domain_non_null ON branding(domain) WHERE domain IS NOT NULL;

-- Grant appropriate permissions (RLS policies already exist from migration 000001)
-- No additional grants needed as branding table already has RLS enabled
