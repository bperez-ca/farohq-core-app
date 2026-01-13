-- Rollback Branding Features Migration
-- Remove all added fields and indexes

-- Drop indexes
DROP INDEX IF EXISTS idx_branding_subdomain;
DROP INDEX IF EXISTS idx_branding_domain_type;
DROP INDEX IF EXISTS idx_branding_website;
DROP INDEX IF EXISTS idx_branding_hide_powered_by;
DROP INDEX IF EXISTS idx_branding_domain_non_null;

-- Drop columns
ALTER TABLE branding DROP COLUMN IF EXISTS ssl_status;
ALTER TABLE branding DROP COLUMN IF EXISTS domain_verification_token;
ALTER TABLE branding DROP COLUMN IF EXISTS cloudflare_zone_id;
ALTER TABLE branding DROP COLUMN IF EXISTS email_domain;
ALTER TABLE branding DROP COLUMN IF EXISTS hide_powered_by;
ALTER TABLE branding DROP COLUMN IF EXISTS domain_type;
ALTER TABLE branding DROP COLUMN IF EXISTS subdomain;
ALTER TABLE branding DROP COLUMN IF EXISTS website;
