--
-- MULTI-TENANCY: ROLLBACK PLATFORM ROLES AND RLS
--

DROP POLICY IF EXISTS platform_tenants_instance_admin ON platform.tenants;
ALTER TABLE platform.tenants NO FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS platform_users_self ON platform.users;
DROP POLICY IF EXISTS platform_users_all ON platform.users;
ALTER TABLE platform.users DISABLE ROW LEVEL SECURITY;

DROP INDEX IF EXISTS idx_platform_users_role_instance_admin;

ALTER TABLE platform.users DROP CONSTRAINT IF EXISTS platform_users_role_check;
ALTER TABLE platform.users ADD CONSTRAINT dashboard_users_role_check 
    CHECK (role IN ('dashboard_admin', 'dashboard_user'));

COMMENT ON COLUMN platform.users.role IS 'User role: dashboard_admin or dashboard_user';
