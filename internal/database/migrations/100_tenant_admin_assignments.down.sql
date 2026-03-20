--
-- MULTI-TENANCY: ROLLBACK TENANT ADMIN ASSIGNMENTS
--

DROP POLICY IF EXISTS platform_tenants_assigned ON platform.tenants;
DROP FUNCTION IF EXISTS platform.user_managed_tenant_ids(UUID);

DROP POLICY IF EXISTS platform_tenant_admin_assignments_self ON platform.tenant_admin_assignments;
DROP POLICY IF EXISTS platform_tenant_admin_assignments_all ON platform.tenant_admin_assignments;

ALTER TABLE platform.tenant_admin_assignments DISABLE ROW LEVEL SECURITY;

ALTER TABLE platform.tenants DROP CONSTRAINT IF EXISTS check_platform_tenants_db_name_valid;

DROP TABLE IF EXISTS platform.tenant_admin_assignments;
