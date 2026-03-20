-- Remove key_type column from auth.service_keys

ALTER TABLE auth.service_keys DROP CONSTRAINT IF EXISTS auth_service_keys_key_type_check;
DROP INDEX IF EXISTS auth.idx_auth_service_keys_key_type;
ALTER TABLE auth.service_keys DROP COLUMN IF EXISTS key_type;
