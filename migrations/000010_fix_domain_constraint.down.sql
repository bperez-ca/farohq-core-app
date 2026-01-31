-- Rollback migration for domain constraint fix

-- Step 1: Drop the partial unique index
DROP INDEX IF EXISTS idx_branding_domain_unique;

-- Step 2: Restore the UNIQUE constraint on domain column
-- Note: This will fail if there are duplicate domain values (other than NULL)
-- which is expected behavior - the constraint should only allow unique non-NULL values
ALTER TABLE branding
ADD CONSTRAINT branding_domain_key UNIQUE (domain);

-- Step 3: Convert NULL domains back to empty strings (optional - for rollback compatibility)
-- UPDATE branding
-- SET domain = ''
-- WHERE domain IS NULL;
