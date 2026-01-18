-- Rollback: Remove revoked_at column from tenant_invites table

-- Drop index
DROP INDEX IF EXISTS idx_tenant_invites_revoked_at;

-- Drop revoked_at column
ALTER TABLE tenant_invites DROP COLUMN IF EXISTS revoked_at;
