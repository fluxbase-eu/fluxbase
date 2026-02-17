-- Revert secrets unique constraint fix

-- Step 1: Drop the unique indexes
DROP INDEX IF EXISTS unique_secrets_global_name;
DROP INDEX IF EXISTS unique_secrets_namespace_name;

-- Step 2: Recreate the original constraint (though it won't work properly with NULLs)
ALTER TABLE functions.secrets ADD CONSTRAINT unique_secret_name_scope UNIQUE (name, scope, namespace);
