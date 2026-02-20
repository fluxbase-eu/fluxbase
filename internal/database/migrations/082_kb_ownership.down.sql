-- Knowledge Base Ownership Migration Down

-- Drop KB permissions table
DROP TABLE IF EXISTS ai.knowledge_base_permissions CASCADE;

-- Remove indexes
DROP INDEX IF EXISTS idx_ai_knowledge_bases_visibility;
DROP INDEX IF EXISTS idx_ai_knowledge_bases_owner;

-- Remove columns from knowledge_bases
ALTER TABLE ai.knowledge_bases DROP COLUMN IF EXISTS owner_id;
ALTER TABLE ai.knowledge_bases DROP COLUMN IF EXISTS visibility;
