-- ============================================================================
-- GRANT PERMISSIONS FOR BRANCHING SCHEMA
-- ============================================================================
-- This migration adds the missing permissions for the branching schema.
-- Branching is an admin-only feature, so only service_role gets access.
-- ============================================================================

-- Grant schema usage to CURRENT_USER (the migration/runtime user)
GRANT USAGE, CREATE ON SCHEMA branching TO CURRENT_USER;

-- Grant schema usage to service_role for admin operations
-- Note: Only service_role has access, not authenticated or anon users
GRANT USAGE ON SCHEMA branching TO service_role;

-- Service role: Full access for admin operations
GRANT ALL ON ALL TABLES IN SCHEMA branching TO service_role;
GRANT ALL ON ALL SEQUENCES IN SCHEMA branching TO service_role;

-- Default privileges for future tables
ALTER DEFAULT PRIVILEGES IN SCHEMA branching
    GRANT ALL ON TABLES TO service_role;

ALTER DEFAULT PRIVILEGES IN SCHEMA branching
    GRANT ALL ON SEQUENCES TO service_role;
