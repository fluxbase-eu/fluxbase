-- ============================================================================
-- API SCHEMA PERMISSIONS
-- ============================================================================
-- This migration grants permissions on the api schema and its tables.
-- The api schema was created in migration 064 but lacked permission grants,
-- causing "permission denied" errors during idempotency key cleanup.
-- ============================================================================

-- Grant schema usage to service_role
GRANT USAGE ON SCHEMA api TO service_role;

-- Grant permissions on existing tables (api.idempotency_keys)
GRANT ALL ON ALL TABLES IN SCHEMA api TO service_role;
GRANT ALL ON ALL SEQUENCES IN SCHEMA api TO service_role;

-- Set default privileges for future tables in the api schema
-- This ensures any new tables automatically get correct permissions
ALTER DEFAULT PRIVILEGES IN SCHEMA api
    GRANT ALL ON TABLES TO service_role;

ALTER DEFAULT PRIVILEGES IN SCHEMA api
    GRANT ALL ON SEQUENCES TO service_role;

-- Add comment
COMMENT ON SCHEMA api IS 'API management features (idempotency, rate limiting, etc.)';
