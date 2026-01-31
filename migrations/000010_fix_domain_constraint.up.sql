-- Fix domain unique constraint to allow NULL values
-- This migration:
-- 1. Updates existing rows with empty string domain to NULL
-- 2. Drops the existing UNIQUE constraint
-- 3. Creates a partial unique index that only enforces uniqueness on non-NULL, non-empty domain values

-- Step 1: Update all existing rows with empty string domain to NULL
UPDATE branding
SET domain = NULL
WHERE domain = '';

-- Step 2: Drop the existing UNIQUE constraint on domain column
-- The constraint name is automatically generated, so we need to find and drop it
DO $$
DECLARE
    constraint_name TEXT;
BEGIN
    -- Find the unique constraint name on the domain column
    SELECT conname INTO constraint_name
    FROM pg_constraint
    WHERE conrelid = 'branding'::regclass
      AND contype = 'u'
      AND conkey = ARRAY(
          SELECT attnum
          FROM pg_attribute
          WHERE attrelid = 'branding'::regclass
            AND attname = 'domain'
      )
    LIMIT 1;

    -- Drop the constraint if it exists
    IF constraint_name IS NOT NULL THEN
        EXECUTE 'ALTER TABLE branding DROP CONSTRAINT ' || constraint_name;
    END IF;
END $$;

-- Step 3: Create a partial unique index that only enforces uniqueness on non-NULL, non-empty domain values
-- This allows multiple rows to have NULL domain, but enforces uniqueness on actual domain values
CREATE UNIQUE INDEX IF NOT EXISTS idx_branding_domain_unique
ON branding (domain)
WHERE domain IS NOT NULL AND domain != '';

-- Note: The existing idx_branding_domain index (if it exists) can remain for query performance
-- This new index will handle the uniqueness constraint for non-NULL values
