-- Add Default User Permission to Knowledge Bases
-- Allows setting a default permission level for all authenticated users

-- Add the column
ALTER TABLE ai.knowledge_bases ADD COLUMN IF NOT EXISTS default_user_permission TEXT CHECK (default_user_permission IN ('viewer', 'editor', 'owner')) DEFAULT 'viewer';

-- Add index for filtering
CREATE INDEX IF NOT EXISTS idx_ai_knowledge_bases_default_user_permission ON ai.knowledge_bases(default_user_permission);

COMMENT ON COLUMN ai.knowledge_bases.default_user_permission IS 'Default permission level for all authenticated users (viewer/editor/owner)';
