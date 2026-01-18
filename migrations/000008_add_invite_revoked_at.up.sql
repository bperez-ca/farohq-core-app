-- Add revoked_at column to tenant_invites table for invite revocation support

-- Add revoked_at column
ALTER TABLE tenant_invites ADD COLUMN IF NOT EXISTS revoked_at TIMESTAMPTZ;

-- Create index for efficient querying of active vs revoked invites
CREATE INDEX IF NOT EXISTS idx_tenant_invites_revoked_at ON tenant_invites(tenant_id, revoked_at) WHERE deleted_at IS NULL;

-- Update existing queries to handle revoked_at (no data migration needed as existing invites are null)
