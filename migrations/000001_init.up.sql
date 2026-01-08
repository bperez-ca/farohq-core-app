-- Core App Migration 0001: Initialize agencies and branding tables
-- Multi-tenant white-label SaaS brand management

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create agencies table
CREATE TABLE IF NOT EXISTS agencies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    status TEXT DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended')),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create branding table with agency relationship
CREATE TABLE IF NOT EXISTS branding (
    agency_id UUID PRIMARY KEY REFERENCES agencies(id) ON DELETE CASCADE,
    domain TEXT UNIQUE,
    verified_at TIMESTAMPTZ,
    logo_url TEXT,
    favicon_url TEXT,
    primary_color TEXT CHECK (primary_color ~ '^#[0-9A-Fa-f]{6}$' OR primary_color IS NULL),
    secondary_color TEXT CHECK (secondary_color ~ '^#[0-9A-Fa-f]{6}$' OR secondary_color IS NULL),
    theme_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_agencies_slug ON agencies(slug);
CREATE INDEX IF NOT EXISTS idx_agencies_status ON agencies(status);
CREATE INDEX IF NOT EXISTS idx_branding_domain ON branding(domain);
CREATE INDEX IF NOT EXISTS idx_branding_agency_id ON branding(agency_id);

-- Enable Row Level Security on branding table
ALTER TABLE branding ENABLE ROW LEVEL SECURITY;

-- Create RLS policy for branding table
-- This policy ensures tenants can only access their own branding data
CREATE POLICY branding_tenant ON branding
    USING (agency_id = current_setting('lv.tenant_id')::uuid);

-- Create function to set tenant context
-- This function will be called by the gateway to set the current tenant
CREATE OR REPLACE FUNCTION set_tenant_context(tenant_uuid UUID)
RETURNS VOID AS $$
BEGIN
    -- Set the tenant ID in the session for RLS policies
    PERFORM set_config('lv.tenant_id', tenant_uuid::text, true);
END;
$$ LANGUAGE plpgsql;

-- Create function to clear tenant context
CREATE OR REPLACE FUNCTION clear_tenant_context()
RETURNS VOID AS $$
BEGIN
    -- Clear the tenant ID from the session
    PERFORM set_config('lv.tenant_id', '', true);
END;
$$ LANGUAGE plpgsql;

-- Create function to get current tenant
CREATE OR REPLACE FUNCTION get_current_tenant()
RETURNS UUID AS $$
BEGIN
    RETURN current_setting('lv.tenant_id')::uuid;
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql;

