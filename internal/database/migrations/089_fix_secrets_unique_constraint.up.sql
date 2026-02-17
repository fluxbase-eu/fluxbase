-- Fix secrets unique constraint to properly handle NULL namespace values
-- PostgreSQL treats NULL values in unique constraints as not equal,
-- so multiple global secrets with the same name can be created.

-- Step 1: Drop the existing constraint
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'unique_secret_name_scope'
        AND conrelid = 'functions.secrets'::regclass
    ) THEN
        ALTER TABLE functions.secrets DROP CONSTRAINT unique_secret_name_scope;
    END IF;
END $$;

-- Step 2: Create a unique index for global secrets (namespace IS NULL)
-- Uses a coalesced empty string to treat NULL consistently
CREATE UNIQUE INDEX unique_secrets_global_name
ON functions.secrets (name)
WHERE scope = 'global' AND namespace IS NULL;

-- Step 3: Create a unique index for namespace secrets (namespace IS NOT NULL)
CREATE UNIQUE INDEX unique_secrets_namespace_name
ON functions.secrets (name, namespace)
WHERE scope = 'namespace' AND namespace IS NOT NULL;
