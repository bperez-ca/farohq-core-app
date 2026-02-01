-- Add RLS to agencies table so all access is scoped to current tenant context.
-- Application always sets lv.tenant_id before querying; this prevents leakage if agencies is queried directly.

ALTER TABLE agencies ENABLE ROW LEVEL SECURITY;

-- Only allow access to the agency row matching the current tenant context.
-- When lv.tenant_id is not set, no rows are visible.
DROP POLICY IF EXISTS agencies_tenant ON agencies;
CREATE POLICY agencies_tenant ON agencies
    USING (
        current_setting('lv.tenant_id', true) <> ''
        AND id = current_setting('lv.tenant_id', true)::uuid
    );
