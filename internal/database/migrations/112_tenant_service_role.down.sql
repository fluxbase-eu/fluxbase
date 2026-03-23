-- Migration 112: Revert tenant_service role and policy changes

-- ============================================================================
-- STEP 1: Revert platform.sso_identities policies
-- ============================================================================

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'platform' AND table_name = 'sso_identities'
    ) THEN
        DROP POLICY IF EXISTS sso_identities_admin ON platform.sso_identities;

        -- Restore original policy (granted to PUBLIC)
        CREATE POLICY sso_identities_admin ON platform.sso_identities
            FOR ALL
            USING (true);
    END IF;
END
$$;

-- ============================================================================
-- STEP 2: Revert platform.sessions policies
-- ============================================================================

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'platform' AND table_name = 'sessions'
    ) THEN
        DROP POLICY IF EXISTS platform_sessions_admin ON platform.sessions;
        DROP POLICY IF EXISTS platform_sessions_self ON platform.sessions;

        -- Restore original policies (granted to PUBLIC)
        CREATE POLICY platform_sessions_admin ON platform.sessions
            FOR ALL
            USING (true);

        CREATE POLICY platform_sessions_self ON platform.sessions
            FOR ALL
            USING (true);
    END IF;
END
$$;

-- ============================================================================
-- STEP 3: Revert tenant_settings policies
-- ============================================================================

DROP POLICY IF EXISTS tenant_settings_insert ON platform.tenant_settings;
DROP POLICY IF EXISTS tenant_settings_update ON platform.tenant_settings;

-- Restore original policies (without WITH CHECK)
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
    );

-- ============================================================================
-- STEP 4: Revert instance_settings policies
-- ============================================================================

DROP POLICY IF EXISTS instance_settings_insert ON platform.instance_settings;
DROP POLICY IF EXISTS instance_settings_update ON platform.instance_settings;
DROP POLICY IF EXISTS instance_settings_select ON platform.instance_settings;

-- Restore original policies (USING (TRUE) for SELECT, no WITH CHECK for UPDATE)
CREATE POLICY instance_settings_select ON platform.instance_settings
    FOR SELECT
    USING (TRUE);

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

CREATE POLICY instance_settings_update ON platform.instance_settings
    FOR UPDATE
    USING (
        current_user = 'service_role'
        OR EXISTS (
            SELECT 1 FROM auth.users
            WHERE id = (current_setting('request.jwt.claims', TRUE)::JSONB->>'sub')::UUID
            AND is_instance_admin(id)
        )
    );

-- ============================================================================
-- STEP 5: Note about orphaned policies
-- We do NOT recreate the orphaned *_tenant_service policies since they were
-- never part of any migration and should not exist.
-- ============================================================================

-- ============================================================================
-- STEP 6: Note about tenant_service role
-- We do NOT drop the tenant_service role since it may be in use.
-- If needed, it can be manually dropped after revoking all grants:
-- REVOKE ALL ON ALL TABLES IN SCHEMA auth FROM tenant_service;
-- REVOKE ALL ON ALL TABLES IN SCHEMA platform FROM tenant_service;
-- DROP ROLE tenant_service;
-- ============================================================================
