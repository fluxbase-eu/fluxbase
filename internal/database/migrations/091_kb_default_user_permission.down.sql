-- Rollback Default User Permission

DROP INDEX IF EXISTS idx_ai_knowledge_bases_default_user_permission;
ALTER TABLE ai.knowledge_bases DROP COLUMN IF EXISTS default_user_permission;
