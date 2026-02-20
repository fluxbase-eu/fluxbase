-- Rollback user quotas
BEGIN;

-- Drop quota columns from knowledge bases
ALTER TABLE ai.knowledge_bases
DROP CONSTRAINT IF EXISTS kb_quota_max_documents_positive,
DROP CONSTRAINT IF EXISTS kb_quota_max_chunks_positive,
DROP CONSTRAINT IF EXISTS kb_quota_max_storage_bytes_positive;

ALTER TABLE ai.knowledge_bases
DROP COLUMN IF EXISTS quota_max_documents,
DROP COLUMN IF EXISTS quota_max_chunks,
DROP COLUMN IF EXISTS quota_max_storage_bytes;

-- Drop index
DROP INDEX IF EXISTS idx_ai_knowledge_bases_owner_quotas;

-- Drop user quotas table
DROP TABLE IF EXISTS ai.user_quotas;

COMMIT;
