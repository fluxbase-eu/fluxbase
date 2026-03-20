--
-- MULTI-TENANCY: ROLLBACK DASHBOARD SCHEMA RENAME
-- Renames platform.* back to dashboard.* and updates function references
--

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = 'platform')
       AND NOT EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = 'dashboard')
    THEN
        ALTER SCHEMA platform RENAME TO dashboard;
    END IF;
END $$;

CREATE OR REPLACE FUNCTION is_instance_admin(p_user_id UUID) RETURNS BOOLEAN AS $$
BEGIN
    IF p_user_id IS NULL THEN
        RETURN false;
    END IF;
    
    RETURN EXISTS (
        SELECT 1 FROM dashboard.users du
        WHERE du.id = p_user_id
        AND du.role = 'instance_admin'
        AND (du.deleted_at IS NULL OR du.is_active = true)
    );
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

COMMENT ON FUNCTION is_instance_admin(UUID) IS
'Checks if a user is an instance-level admin with global privileges. SECURITY DEFINER to bypass RLS.';

COMMENT ON COLUMN auth.mcp_oauth_codes.user_id IS 'Dashboard user who authorized this code (references dashboard.users)';
COMMENT ON COLUMN auth.mcp_oauth_tokens.user_id IS 'Dashboard user this token represents (references dashboard.users)';
