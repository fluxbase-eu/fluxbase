-- Add key_type column to auth.service_keys for database-per-tenant multi-tenancy
-- Each tenant database has its own service_keys table
-- key_type distinguishes between anon keys and service keys

ALTER TABLE auth.service_keys ADD COLUMN IF NOT EXISTS key_type TEXT NOT NULL DEFAULT 'service';

-- Add constraint for valid key types
ALTER TABLE auth.service_keys DROP CONSTRAINT IF EXISTS auth_service_keys_key_type_check;
ALTER TABLE auth.service_keys ADD CONSTRAINT auth_service_keys_key_type_check 
    CHECK (key_type IN ('anon', 'service'));

-- Update comment
COMMENT ON COLUMN auth.service_keys.key_type IS 'Type of key: anon (anonymous access) or service (elevated privileges bypassing RLS)';

-- Create index for key_type lookups
CREATE INDEX IF NOT EXISTS idx_auth_service_keys_key_type ON auth.service_keys(key_type);

-- Backfill existing keys as 'service' type (they already have service role privileges)
UPDATE auth.service_keys SET key_type = 'service' WHERE key_type IS NULL OR key_type = '';
