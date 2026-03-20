--
-- MULTI-TENANCY: CORE TENANT TABLES (DATABASE-PER-TENANT)
-- Creates the platform schema and tenants table for multi-tenant support
-- Key design: db_name = NULL means "use the main database" (backward compatibility)
--

CREATE SCHEMA IF NOT EXISTS platform;

CREATE TABLE IF NOT EXISTS platform.tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    db_name TEXT,
    is_default BOOLEAN DEFAULT false,
    status TEXT NOT NULL DEFAULT 'active',
    metadata JSONB DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT platform_tenants_status_check CHECK (status IN ('creating', 'active', 'deleting', 'error'))
);

CREATE INDEX IF NOT EXISTS idx_platform_tenants_slug ON platform.tenants(slug);
CREATE INDEX IF NOT EXISTS idx_platform_tenants_is_default ON platform.tenants(is_default) WHERE is_default = true;
CREATE INDEX IF NOT EXISTS idx_platform_tenants_status ON platform.tenants(status);
CREATE INDEX IF NOT EXISTS idx_platform_tenants_deleted_at ON platform.tenants(deleted_at) WHERE deleted_at IS NOT NULL;

COMMENT ON TABLE platform.tenants IS 'Tenant registry for database-per-tenant multi-tenancy. db_name = NULL means use main database.';
COMMENT ON COLUMN platform.tenants.id IS 'Unique identifier for the tenant';
COMMENT ON COLUMN platform.tenants.slug IS 'URL-friendly identifier for the tenant (e.g., "acme-corp")';
COMMENT ON COLUMN platform.tenants.name IS 'Display name for the tenant';
COMMENT ON COLUMN platform.tenants.db_name IS 'Database name for this tenant. NULL = use main database (backward compatibility for default tenant)';
COMMENT ON COLUMN platform.tenants.is_default IS 'True for the default tenant used for backward compatibility';
COMMENT ON COLUMN platform.tenants.status IS 'Tenant status: creating, active, deleting, error';
COMMENT ON COLUMN platform.tenants.metadata IS 'Arbitrary metadata for the tenant (plan, settings, etc.)';
COMMENT ON COLUMN platform.tenants.deleted_at IS 'Soft delete timestamp. NULL if tenant is active.';

CREATE OR REPLACE FUNCTION update_platform_tenants_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER platform_tenants_updated_at
    BEFORE UPDATE ON platform.tenants
    FOR EACH ROW
    EXECUTE FUNCTION update_platform_tenants_updated_at();

ALTER TABLE platform.tenants ENABLE ROW LEVEL SECURITY;

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
'Checks if a user is an instance-level admin with global privileges. SECURITY DEFINER to bypass RLS. References dashboard.users until schema rename migration.';
