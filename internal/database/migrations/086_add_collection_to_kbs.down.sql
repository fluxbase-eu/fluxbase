-- Rollback: Remove collection_id from knowledge bases
-- Migration 087
--
ALTER TABLE ai.knowledge_bases
    DROP CONSTRAINT IF EXISTS kb_owner_or_collection;

ALTER TABLE ai.knowledge_bases
    DROP COLUMN IF EXISTS collection_id;

DROP INDEX IF EXISTS idx_ai_knowledge_bases_collection;
