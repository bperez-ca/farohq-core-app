-- Add invite_expiry_hours column to agencies table
-- Configurable per-tenant invitation expiration (in hours, max 72 hours = 3 days)
ALTER TABLE agencies ADD COLUMN IF NOT EXISTS invite_expiry_hours INTEGER DEFAULT 24 CHECK (invite_expiry_hours IS NULL OR (invite_expiry_hours > 0 AND invite_expiry_hours <= 72));

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_agencies_invite_expiry_hours ON agencies(invite_expiry_hours) WHERE invite_expiry_hours IS NOT NULL;
