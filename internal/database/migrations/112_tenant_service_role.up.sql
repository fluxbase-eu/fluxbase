-- Migration 112: Create tenant_service role and fix security warnings
-- This addresses RLS policy security warnings by:
-- 1. Creating the missing tenant_service role
-- 2. Dropping orphaned *_tenant_service policies that use USING (true)
-- 3. Fixing platform schema policies

-- ============================================================================
-- STEP 1: Create tenant_service role
-- ============================================================================

-- Create tenant_service role if it doesn't exist
-- This role is referenced by migrations 109 and 111 but was never created
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'tenant_service') THEN
        CREATE ROLE tenant_service NOLOGIN NOINHERIT NOBYPASSRLS;
    END IF;
END
$$;

COMMENT ON ROLE tenant_service IS 'Tenant-scoped service role for multi-tenant isolation - respects RLS with tenant context';

-- Grant to current user so RLS middleware can SET ROLE
GRANT tenant_service TO CURRENT_USER;

-- ============================================================================
-- STEP 2: Drop orphaned *_tenant_service policies
-- These policies use USING (true) which defeats tenant isolation
-- ============================================================================

-- AI schema
DROP POLICY IF EXISTS chunks_tenant_service ON ai.chunks;
DROP POLICY IF EXISTS messages_tenant_service ON ai.messages;
DROP POLICY IF EXISTS entities_tenant_service ON ai.entities;
DROP POLICY IF EXISTS entity_relationships_tenant_service ON ai.entity_relationships;
DROP POLICY IF EXISTS document_entities_tenant_service ON ai.document_entities;
DROP POLICY IF EXISTS query_audit_log_tenant_service ON ai.query_audit_log;
DROP POLICY IF EXISTS retrieval_log_tenant_service ON ai.retrieval_log;

-- Functions schema
DROP POLICY IF EXISTS edge_executions_tenant_service ON functions.edge_executions;
DROP POLICY IF EXISTS edge_files_tenant_service ON functions.edge_files;
DROP POLICY IF EXISTS function_dependencies_tenant_service ON functions.function_dependencies;
DROP POLICY IF EXISTS secret_versions_tenant_service ON functions.secret_versions;
DROP POLICY IF EXISTS shared_modules_tenant_service ON functions.shared_modules;

-- Jobs schema
DROP POLICY IF EXISTS jobs_function_files_tenant_service ON jobs.function_files;
DROP POLICY IF EXISTS workers_tenant_service ON jobs.workers;

-- Auth schema
DROP POLICY IF EXISTS sessions_tenant_service ON auth.sessions;
DROP POLICY IF EXISTS mfa_factors_tenant_service ON auth.mfa_factors;
DROP POLICY IF EXISTS oauth_links_tenant_service ON auth.oauth_links;
DROP POLICY IF EXISTS impersonation_sessions_tenant_service ON auth.impersonation_sessions;
DROP POLICY IF EXISTS webhook_deliveries_tenant_service ON auth.webhook_deliveries;
DROP POLICY IF EXISTS webhook_events_tenant_service ON auth.webhook_events;
DROP POLICY IF EXISTS webhook_monitored_tables_tenant_service ON auth.webhook_monitored_tables;
DROP POLICY IF EXISTS client_key_usage_tenant_service ON auth.client_key_usage;

-- Storage schema
DROP POLICY IF EXISTS chunked_upload_sessions_tenant_service ON storage.chunked_upload_sessions;

-- Logging schema
DROP POLICY IF EXISTS entries_tenant_service ON logging.entries;
DROP POLICY IF EXISTS entries_ai_tenant_service ON logging.entries_ai;
DROP POLICY IF EXISTS entries_custom_tenant_service ON logging.entries_custom;
DROP POLICY IF EXISTS entries_execution_tenant_service ON logging.entries_execution;
DROP POLICY IF EXISTS entries_http_tenant_service ON logging.entries_http;
DROP POLICY IF EXISTS entries_security_tenant_service ON logging.entries_security;
DROP POLICY IF EXISTS entries_system_tenant_service ON logging.entries_system;

-- Branching schema
DROP POLICY IF EXISTS branches_tenant_service ON branching.branches;
DROP POLICY IF EXISTS branch_access_tenant_service ON branching.branch_access;
DROP POLICY IF EXISTS github_config_tenant_service ON branching.github_config;
DROP POLICY IF EXISTS branching_activity_log_tenant_service ON branching.activity_log;
DROP POLICY IF EXISTS migration_history_tenant_service ON branching.migration_history;
DROP POLICY IF EXISTS seed_execution_log_tenant_service ON branching.seed_execution_log;

-- RPC schema
DROP POLICY IF EXISTS rpc_executions_tenant_service ON rpc.executions;
DROP POLICY IF EXISTS procedures_tenant_service ON rpc.procedures;

-- Realtime schema
DROP POLICY IF EXISTS schema_registry_tenant_service ON realtime.schema_registry;

-- ============================================================================
-- STEP 3: Fix platform.instance_settings policies
-- ============================================================================

-- Fix the overly permissive SELECT policy
DROP POLICY IF EXISTS instance_settings_select ON platform.instance_settings;
CREATE POLICY instance_settings_select ON platform.instance_settings
    FOR SELECT
    USING (
        -- Service role can always read
        current_user = 'service_role'
        -- Tenant service can read for settings cascade resolution
        OR current_user = 'tenant_service'
        -- Instance admins can read
        OR EXISTS (
            SELECT 1 FROM auth.users
            WHERE id = (current_setting('request.jwt.claims', TRUE)::JSONB->>'sub')::UUID
            AND is_instance_admin(id)
        )
        -- Authenticated users with tenant context can read (needed for settings resolution)
        OR (
            current_setting('app.current_tenant_id', TRUE) IS NOT NULL
            AND current_setting('app.current_tenant_id', TRUE) != ''
        )
    );

-- Add WITH CHECK clause to UPDATE policy
DROP POLICY IF EXISTS instance_settings_update ON platform.instance_settings;
CREATE POLICY instance_settings_update ON platform.instance_settings
    FOR UPDATE
    USING (
        current_user = 'service_role'
        OR EXISTS (
            SELECT 1 FROM auth.users
            WHERE id = (current_setting('request.jwt.claims', TRUE)::JSONB->>'sub')::UUID
            AND is_instance_admin(id)
        )
    )
    WITH CHECK (
        current_user = 'service_role'
        OR EXISTS (
            SELECT 1 FROM auth.users
            WHERE id = (current_setting('request.jwt.claims', TRUE)::JSONB->>'sub')::UUID
            AND is_instance_admin(id)
        )
    );

-- Add WITH CHECK clause to INSERT policy
DROP POLICY IF EXISTS instance_settings_insert ON platform.instance_settings;
CREATE POLICY instance_settings_insert ON platform.instance_settings
    FOR INSERT
    WITH CHECK (
        current_user = 'service_role'
        OR EXISTS (
            SELECT 1 FROM auth.users
            WHERE id = (current_setting('request.jwt.claims', TRUE)::JSONB->>'sub')::UUID
            AND is_instance_admin(id)
        )
    );

-- Add WITH CHECK clause to tenant_settings UPDATE policy
DROP POLICY IF EXISTS tenant_settings_update ON platform.tenant_settings;
CREATE POLICY tenant_settings_update ON platform.tenant_settings
    FOR UPDATE
    USING (
        current_user = 'service_role'
        OR EXISTS (
            SELECT 1 FROM auth.users
            WHERE id = (current_setting('request.jwt.claims', TRUE)::JSONB->>'sub')::UUID
            AND is_instance_admin(id)
        )
        OR auth.has_tenant_access(tenant_id)
    )
    WITH CHECK (
        current_user = 'service_role'
        OR EXISTS (
            SELECT 1 FROM auth.users
            WHERE id = (current_setting('request.jwt.claims', TRUE)::JSONB->>'sub')::UUID
            AND is_instance_admin(id)
        )
        OR auth.has_tenant_access(tenant_id)
    );

-- Add WITH CHECK clause to tenant_settings INSERT policy
DROP POLICY IF EXISTS tenant_settings_insert ON platform.tenant_settings;
CREATE POLICY tenant_settings_insert ON platform.tenant_settings
    FOR INSERT
    WITH CHECK (
        current_user = 'service_role'
        OR EXISTS (
            SELECT 1 FROM auth.users
            WHERE id = (current_setting('request.jwt.claims', TRUE)::JSONB->>'sub')::UUID
            AND is_instance_admin(id)
        )
        OR auth.has_tenant_access(tenant_id)
    );

-- ============================================================================
-- STEP 4: Fix platform.sessions policies (if table exists)
-- These policies were granting access to PUBLIC role
-- ============================================================================

-- Note: platform.sessions may not exist; these are conditional
DO $$
BEGIN
    -- Check if platform.sessions exists
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'platform' AND table_name = 'sessions'
    ) THEN
        -- Drop policies that grant to PUBLIC
        DROP POLICY IF EXISTS platform_sessions_admin ON platform.sessions;
        DROP POLICY IF EXISTS platform_sessions_self ON platform.sessions;

        -- Create proper policies
        CREATE POLICY platform_sessions_admin ON platform.sessions
            FOR ALL
            USING (
                current_user = 'service_role'
                OR EXISTS (
                    SELECT 1 FROM auth.users
                    WHERE id = (current_setting('request.jwt.claims', TRUE)::JSONB->>'sub')::UUID
                    AND is_instance_admin(id)
                )
            )
            WITH CHECK (
                current_user = 'service_role'
                OR EXISTS (
                    SELECT 1 FROM auth.users
                    WHERE id = (current_setting('request.jwt.claims', TRUE)::JSONB->>'sub')::UUID
                    AND is_instance_admin(id)
                )
            );
    END IF;
END
$$;

-- ============================================================================
-- STEP 5: Fix platform.sso_identities policies (if table exists)
-- ============================================================================

DO $$
BEGIN
    -- Check if platform.sso_identities exists
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'platform' AND table_name = 'sso_identities'
    ) THEN
        -- Drop policies that grant to PUBLIC
        DROP POLICY IF EXISTS sso_identities_admin ON platform.sso_identities;

        -- Create proper policies
        CREATE POLICY sso_identities_admin ON platform.sso_identities
            FOR ALL
            USING (
                current_user = 'service_role'
                OR EXISTS (
                    SELECT 1 FROM auth.users
                    WHERE id = (current_setting('request.jwt.claims', TRUE)::JSONB->>'sub')::UUID
                    AND is_instance_admin(id)
                )
            )
            WITH CHECK (
                current_user = 'service_role'
                OR EXISTS (
                    SELECT 1 FROM auth.users
                    WHERE id = (current_setting('request.jwt.claims', TRUE)::JSONB->>'sub')::UUID
                    AND is_instance_admin(id)
                )
            );
    END IF;
END
$$;
