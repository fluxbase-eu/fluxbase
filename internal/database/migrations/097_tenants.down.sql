--
-- MULTI-TENANCY: ROLLBACK CORE TENANT TABLES
--

DROP TRIGGER IF EXISTS platform_tenants_updated_at ON platform.tenants;
DROP FUNCTION IF EXISTS update_platform_tenants_updated_at();
DROP FUNCTION IF EXISTS is_instance_admin(UUID);
ALTER TABLE platform.tenants DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS platform.tenants;
DROP SCHEMA IF EXISTS platform;
