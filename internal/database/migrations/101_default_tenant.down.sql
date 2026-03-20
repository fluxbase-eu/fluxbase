--
-- MULTI-TENANCY: ROLLBACK DEFAULT TENANT AND ADMIN MIGRATION
--

DO $$
BEGIN
    UPDATE platform.users
    SET role = 'dashboard_admin',
        updated_at = NOW()
    WHERE role IN ('instance_admin', 'tenant_admin')
    AND deleted_at IS NULL;

    DELETE FROM platform.tenants WHERE is_default = true;

    RAISE NOTICE 'Rolled back default tenant and admin roles';
END $$;
