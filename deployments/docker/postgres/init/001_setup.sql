-- ProductGraph PostgreSQL Setup
-- Initial setup for RLS and extensions (schema managed by Ent)

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create service role for bypassing RLS (migrations, admin tasks)
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'service_role') THEN
        CREATE ROLE service_role NOLOGIN;
    END IF;
END
$$;

-- Grant service_role to pg user for admin operations
GRANT service_role TO pg;

-- Create app role for application connections
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'app_role') THEN
        CREATE ROLE app_role NOLOGIN;
    END IF;
END
$$;

-- Function to set current org context (called at start of each request)
CREATE OR REPLACE FUNCTION set_current_org(org_id uuid)
RETURNS void AS $$
BEGIN
    PERFORM set_config('app.current_org_id', org_id::text, false);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to get current org context
CREATE OR REPLACE FUNCTION current_org_id()
RETURNS uuid AS $$
BEGIN
    RETURN NULLIF(current_setting('app.current_org_id', true), '')::uuid;
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE;

-- Note: Tables and RLS policies are created by Ent migrations
-- This file only sets up the foundation (extensions, roles, functions)
