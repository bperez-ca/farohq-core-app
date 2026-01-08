-- Agency Hierarchy Migration: Create clients, locations, and client_members tables
-- Multi-tenant agency hierarchy with tier-based seat management
--
-- This migration adds:
-- 1. Tier and seat management to agencies
-- 2. Clients (SMB accounts) table
-- 3. Locations table
-- 4. Client members table
-- 5. Soft delete support for all tables
-- 6. RLS policies for data isolation

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Agencies: add tier, seat management, and soft delete
ALTER TABLE agencies ADD COLUMN IF NOT EXISTS tier TEXT CHECK (tier IN ('starter', 'growth', 'scale'));
ALTER TABLE agencies ADD COLUMN IF NOT EXISTS agency_seat_limit INTEGER DEFAULT 0;
ALTER TABLE agencies ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE agencies ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

-- Clients (SMB accounts) with soft delete
CREATE TABLE IF NOT EXISTS clients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agency_id UUID NOT NULL REFERENCES agencies(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    tier TEXT CHECK (tier IN ('starter', 'growth', 'scale')),
    status TEXT DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(agency_id, slug)
);

-- Locations with soft delete
CREATE TABLE IF NOT EXISTS locations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    address JSONB,
    phone TEXT,
    business_hours JSONB,
    categories TEXT[],
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Client Members (separate from agency members) with soft delete
CREATE TABLE IF NOT EXISTS client_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'staff', 'viewer', 'client_viewer')),
    location_id UUID REFERENCES locations(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(client_id, user_id, location_id)
);

-- Update existing tenant_members table to add soft delete and client_id
ALTER TABLE tenant_members ADD COLUMN IF NOT EXISTS client_id UUID REFERENCES clients(id) ON DELETE SET NULL;
ALTER TABLE tenant_members ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

-- Update existing tenant_invites table to add soft delete
ALTER TABLE tenant_invites ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

-- Update role constraint in tenant_members to include client_viewer
ALTER TABLE tenant_members DROP CONSTRAINT IF EXISTS tenant_members_role_check;
ALTER TABLE tenant_members ADD CONSTRAINT tenant_members_role_check 
    CHECK (role IN ('owner', 'admin', 'staff', 'viewer', 'client_viewer'));

-- Update role constraint in tenant_invites to include client_viewer
ALTER TABLE tenant_invites DROP CONSTRAINT IF EXISTS tenant_invites_role_check;
ALTER TABLE tenant_invites ADD CONSTRAINT tenant_invites_role_check 
    CHECK (role IN ('owner', 'admin', 'staff', 'viewer', 'client_viewer'));

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_clients_agency_id ON clients(agency_id);
CREATE INDEX IF NOT EXISTS idx_clients_slug ON clients(agency_id, slug);
CREATE INDEX IF NOT EXISTS idx_clients_status ON clients(agency_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_clients_deleted_at ON clients(agency_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_locations_client_id ON locations(client_id);
CREATE INDEX IF NOT EXISTS idx_locations_active ON locations(client_id, is_active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_locations_deleted_at ON locations(client_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_client_members_client_id ON client_members(client_id);
CREATE INDEX IF NOT EXISTS idx_client_members_user_id ON client_members(user_id);
CREATE INDEX IF NOT EXISTS idx_client_members_location_id ON client_members(location_id);
CREATE INDEX IF NOT EXISTS idx_client_members_deleted_at ON client_members(client_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_tenant_members_client_id ON tenant_members(client_id);
CREATE INDEX IF NOT EXISTS idx_tenant_members_deleted_at ON tenant_members(tenant_id) WHERE deleted_at IS NULL;

-- Enable Row Level Security on clients table
ALTER TABLE clients ENABLE ROW LEVEL SECURITY;

-- Create RLS policy for clients table
DROP POLICY IF EXISTS clients_tenant ON clients;
CREATE POLICY clients_tenant ON clients
    USING (agency_id = current_setting('lv.tenant_id')::uuid AND deleted_at IS NULL);

-- Enable Row Level Security on locations table
ALTER TABLE locations ENABLE ROW LEVEL SECURITY;

-- Create RLS policy for locations table
DROP POLICY IF EXISTS locations_tenant ON locations;
CREATE POLICY locations_tenant ON locations
    USING (client_id IN (
        SELECT id FROM clients 
        WHERE agency_id = current_setting('lv.tenant_id')::uuid 
        AND deleted_at IS NULL
    ) AND deleted_at IS NULL);

-- Enable Row Level Security on client_members table
ALTER TABLE client_members ENABLE ROW LEVEL SECURITY;

-- Create RLS policy for client_members table
DROP POLICY IF EXISTS client_members_tenant ON client_members;
CREATE POLICY client_members_tenant ON client_members
    USING (client_id IN (
        SELECT id FROM clients 
        WHERE agency_id = current_setting('lv.tenant_id')::uuid 
        AND deleted_at IS NULL
    ) AND deleted_at IS NULL);

-- Update RLS policy for tenant_members to include soft delete
DROP POLICY IF EXISTS tenant_members_tenant ON tenant_members;
CREATE POLICY tenant_members_tenant ON tenant_members
    USING (tenant_id = current_setting('lv.tenant_id')::uuid AND deleted_at IS NULL);

-- Update RLS policy for tenant_invites to include soft delete
DROP POLICY IF EXISTS tenant_invites_tenant ON tenant_invites;
CREATE POLICY tenant_invites_tenant ON tenant_invites
    USING (tenant_id = current_setting('lv.tenant_id')::uuid AND deleted_at IS NULL);

-- Create trigger to automatically update updated_at timestamp for agencies
CREATE OR REPLACE FUNCTION update_agencies_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_agencies_updated_at ON agencies;
CREATE TRIGGER update_agencies_updated_at
    BEFORE UPDATE ON agencies
    FOR EACH ROW
    EXECUTE FUNCTION update_agencies_updated_at();

-- Create trigger to automatically update updated_at timestamp for clients
CREATE OR REPLACE FUNCTION update_clients_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_clients_updated_at ON clients;
CREATE TRIGGER update_clients_updated_at
    BEFORE UPDATE ON clients
    FOR EACH ROW
    EXECUTE FUNCTION update_clients_updated_at();

-- Create trigger to automatically update updated_at timestamp for locations
CREATE OR REPLACE FUNCTION update_locations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_locations_updated_at ON locations;
CREATE TRIGGER update_locations_updated_at
    BEFORE UPDATE ON locations
    FOR EACH ROW
    EXECUTE FUNCTION update_locations_updated_at();

-- Create trigger to automatically update updated_at timestamp for client_members
CREATE OR REPLACE FUNCTION update_client_members_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_client_members_updated_at ON client_members;
CREATE TRIGGER update_client_members_updated_at
    BEFORE UPDATE ON client_members
    FOR EACH ROW
    EXECUTE FUNCTION update_client_members_updated_at();

-- Grant appropriate permissions
GRANT SELECT, INSERT, UPDATE, DELETE ON clients TO PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON locations TO PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON client_members TO PUBLIC;
GRANT EXECUTE ON FUNCTION update_agencies_updated_at() TO PUBLIC;
GRANT EXECUTE ON FUNCTION update_clients_updated_at() TO PUBLIC;
GRANT EXECUTE ON FUNCTION update_locations_updated_at() TO PUBLIC;
GRANT EXECUTE ON FUNCTION update_client_members_updated_at() TO PUBLIC;

