--
-- MULTI-TENANCY: TENANT ADMIN ASSIGNMENTS
-- Creates table for assigning platform users as tenant admins
-- Adds check constraint for db_name validity
-- Adds RLS policies for tenant admin access
--

CREATE TABLE IF NOT EXISTS platform.tenant_admin_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES platform.tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES platform.users(id) ON DELETE CASCADE,
    assigned_by UUID REFERENCES platform.users(id) ON DELETE SET NULL,
    assigned_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT platform_tenant_admin_assignments_unique UNIQUE(tenant_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_platform_tenant_admin_assignments_user_id ON platform.tenant_admin_assignments(user_id);
CREATE INDEX IF NOT EXISTS idx_platform_tenant_admin_assignments_tenant_id ON platform.tenant_admin_assignments(tenant_id);

COMMENT ON TABLE platform.tenant_admin_assignments IS 'Maps platform users to tenants as tenant administrators';
COMMENT ON COLUMN platform.tenant_admin_assignments.user_id IS 'Reference to the platform.users table (platform admin)';
COMMENT ON COLUMN platform.tenant_admin_assignments.tenant_id IS 'Reference to the platform.tenants table';
COMMENT ON COLUMN platform.tenant_admin_assignments.assigned_by IS 'Platform user who assigned this admin role';
COMMENT ON COLUMN platform.tenant_admin_assignments.assigned_at IS 'Timestamp when the admin assignment was created';

ALTER TABLE platform.tenants
ADD CONSTRAINT check_platform_tenants_db_name_valid
CHECK (
    (is_default = true AND db_name IS NULL)
    OR
    (is_default = false AND db_name IS NOT NULL AND db_name != '')
);

ALTER TABLE platform.tenant_admin_assignments ENABLE ROW LEVEL SECURITY;

CREATE POLICY platform_tenant_admin_assignments_all ON platform.tenant_admin_assignments
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY platform_tenant_admin_assignments_self ON platform.tenant_admin_assignments
    FOR SELECT TO authenticated
    USING (
        is_instance_admin(auth.uid())
        OR user_id = auth.uid()
    );

CREATE POLICY platform_tenants_assigned ON platform.tenants
    FOR SELECT TO authenticated
    USING (
        is_instance_admin(auth.uid())
        OR EXISTS (
            SELECT 1 FROM platform.tenant_admin_assignments taa
            WHERE taa.tenant_id = platform.tenants.id
            AND taa.user_id = auth.uid()
        )
    );

CREATE OR REPLACE FUNCTION platform.user_managed_tenant_ids(p_user_id UUID) RETURNS UUID[] AS $$
BEGIN
    IF p_user_id IS NULL THEN
        RETURN '{}'::UUID[];
    END IF;
    
    IF EXISTS (
        SELECT 1 FROM platform.users
        WHERE id = p_user_id
        AND role = 'instance_admin'
        AND deleted_at IS NULL
        AND is_active = true
    ) THEN
        RETURN ARRAY(
            SELECT id FROM platform.tenants WHERE deleted_at IS NULL
        );
    END IF;
    
    RETURN ARRAY(
        SELECT tenant_id 
        FROM platform.tenant_admin_assignments 
        WHERE user_id = p_user_id
    );
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

COMMENT ON FUNCTION platform.user_managed_tenant_ids(UUID) IS
'Returns array of tenant IDs that a platform user can manage. Instance admins get all tenants; others get only their assigned tenants. SECURITY DEFINER to bypass RLS.';
