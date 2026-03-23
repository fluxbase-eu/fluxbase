--
-- MULTI-TENANCY: UPDATE PLATFORM ROLES AND ENABLE RLS
-- Adds instance_admin and tenant_admin roles to platform.users constraint
-- Enables RLS on control plane tables
--

ALTER TABLE platform.users DROP CONSTRAINT IF EXISTS dashboard_users_role_check;
ALTER TABLE platform.users DROP CONSTRAINT IF EXISTS platform_users_role_check;
ALTER TABLE platform.users ADD CONSTRAINT platform_users_role_check
    CHECK (role IN ('instance_admin', 'tenant_admin', 'dashboard_admin', 'dashboard_user'));

COMMENT ON COLUMN platform.users.role IS
    'User role: instance_admin (global admin managing all tenants), tenant_admin (admin for specific tenant), dashboard_admin (legacy, maps to tenant_admin), dashboard_user (limited read-only access)';

CREATE INDEX IF NOT EXISTS idx_platform_users_role_instance_admin ON platform.users(role) WHERE role = 'instance_admin';

ALTER TABLE platform.users ENABLE ROW LEVEL SECURITY;

CREATE POLICY platform_users_all ON platform.users
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY platform_users_self ON platform.users
    FOR SELECT TO authenticated
    USING (id = auth.uid());

ALTER TABLE platform.tenants FORCE ROW LEVEL SECURITY;

CREATE POLICY platform_tenants_instance_admin ON platform.tenants
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));
