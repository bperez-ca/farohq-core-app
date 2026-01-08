-- Tenants Module Migration: Create tenant_members and tenant_invites tables
-- Multi-tenant member and invite management
--
-- Note: The agencies table already exists (created by brand service migration)
-- This migration adds tenant membership and invitation functionality

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create tenant_members table
-- Links users to tenants (agencies) with roles
CREATE TABLE IF NOT EXISTS tenant_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES agencies(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'staff', 'viewer')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, user_id)
);

-- Create tenant_invites table
-- Stores invitations to join tenants
CREATE TABLE IF NOT EXISTS tenant_invites (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES agencies(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'staff', 'viewer')),
    token TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    UNIQUE(tenant_id, email)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_tenant_members_tenant_id ON tenant_members(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_members_user_id ON tenant_members(user_id);
CREATE INDEX IF NOT EXISTS idx_tenant_members_tenant_user ON tenant_members(tenant_id, user_id);
CREATE INDEX IF NOT EXISTS idx_tenant_invites_tenant_id ON tenant_invites(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_invites_token ON tenant_invites(token);
CREATE INDEX IF NOT EXISTS idx_tenant_invites_email ON tenant_invites(email);
CREATE INDEX IF NOT EXISTS idx_tenant_invites_expires_at ON tenant_invites(expires_at) WHERE accepted_at IS NULL;

-- Enable Row Level Security on tenant_members table
ALTER TABLE tenant_members ENABLE ROW LEVEL SECURITY;

-- Create RLS policy for tenant_members table
CREATE POLICY tenant_members_tenant ON tenant_members
    USING (tenant_id = current_setting('lv.tenant_id')::uuid);

-- Enable Row Level Security on tenant_invites table
ALTER TABLE tenant_invites ENABLE ROW LEVEL SECURITY;

-- Create RLS policy for tenant_invites table
CREATE POLICY tenant_invites_tenant ON tenant_invites
    USING (tenant_id = current_setting('lv.tenant_id')::uuid);

-- Create trigger to automatically update updated_at timestamp for tenant_members
CREATE OR REPLACE FUNCTION update_tenant_members_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_tenant_members_updated_at ON tenant_members;
CREATE TRIGGER update_tenant_members_updated_at
    BEFORE UPDATE ON tenant_members
    FOR EACH ROW
    EXECUTE FUNCTION update_tenant_members_updated_at();

-- Grant appropriate permissions
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_members TO PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_invites TO PUBLIC;
GRANT EXECUTE ON FUNCTION update_tenant_members_updated_at() TO PUBLIC;

