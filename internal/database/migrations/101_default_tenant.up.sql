--
-- MULTI-TENANCY: CREATE DEFAULT TENANT AND MIGRATE ADMINS
-- Creates default tenant for backward compatibility
-- Migrates first admin to instance_admin, others to tenant_admin
-- Key: db_name = NULL means "use the main database"
--

DO $$
DECLARE
    default_tenant_id UUID;
    first_admin_id UUID;
BEGIN
    SELECT id INTO default_tenant_id FROM platform.tenants WHERE is_default = true LIMIT 1;

    IF default_tenant_id IS NULL THEN
        INSERT INTO platform.tenants (
            id, 
            slug, 
            name, 
            db_name,
            is_default, 
            status,
            metadata, 
            created_at
        ) VALUES (
            gen_random_uuid(),
            'default',
            'Default Tenant',
            NULL,
            true,
            'active',
            '{"description": "Default tenant for backward compatibility - uses main database", "migrated": true}'::jsonb,
            NOW()
        )
        RETURNING id INTO default_tenant_id;

        RAISE NOTICE 'Created default tenant with ID: % (db_name = NULL, uses main database)', default_tenant_id;
    ELSE
        RAISE NOTICE 'Default tenant already exists with ID: %', default_tenant_id;
    END IF;

    SELECT id INTO first_admin_id
    FROM platform.users
    WHERE deleted_at IS NULL AND is_active = true
    ORDER BY created_at ASC
    LIMIT 1;

    IF first_admin_id IS NOT NULL THEN
        UPDATE platform.users
        SET role = 'instance_admin',
            updated_at = NOW()
        WHERE id = first_admin_id
        AND role != 'instance_admin';

        RAISE NOTICE 'Migrated first admin % to instance_admin', first_admin_id;
    END IF;

    UPDATE platform.users
    SET role = 'tenant_admin',
        updated_at = NOW()
    WHERE role = 'dashboard_admin'
    AND (id != first_admin_id OR first_admin_id IS NULL)
    AND deleted_at IS NULL;

    RAISE NOTICE 'Default tenant setup complete';
END $$;
