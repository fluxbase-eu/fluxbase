--
-- MULTI-TENANCY: RENAME DASHBOARD SCHEMA TO PLATFORM
-- Renames dashboard.* to platform.* and updates all FK references
--

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = 'dashboard')
       AND NOT EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = 'platform' AND schema_name != 'dashboard')
    THEN
        ALTER SCHEMA dashboard RENAME TO platform;
    END IF;
END $$;

CREATE OR REPLACE FUNCTION is_instance_admin(p_user_id UUID) RETURNS BOOLEAN AS $$
BEGIN
    IF p_user_id IS NULL THEN
        RETURN false;
    END IF;
    
    RETURN EXISTS (
        SELECT 1 FROM platform.users pu
        WHERE pu.id = p_user_id
        AND pu.role = 'instance_admin'
        AND pu.deleted_at IS NULL
        AND pu.is_active = true
    );
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

COMMENT ON FUNCTION is_instance_admin(UUID) IS
'Checks if a user is an instance-level admin with global privileges. SECURITY DEFINER to bypass RLS.';

COMMENT ON COLUMN auth.mcp_oauth_codes.user_id IS 'Platform user who authorized this code (references platform.users)';
COMMENT ON COLUMN auth.mcp_oauth_tokens.user_id IS 'Platform user this token represents (references platform.users)';
