-- Remove invite_expiry_hours column from agencies table
ALTER TABLE agencies DROP COLUMN IF EXISTS invite_expiry_hours;
