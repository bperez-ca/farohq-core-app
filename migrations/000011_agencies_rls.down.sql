-- Remove RLS from agencies table
DROP POLICY IF EXISTS agencies_tenant ON agencies;
ALTER TABLE agencies DISABLE ROW LEVEL SECURITY;
