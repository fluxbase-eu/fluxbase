-- Knowledge Base Ownership Migration
-- Adds ownership, visibility, and permissions for user-scoped knowledge bases
-- Note: ai schema is created in 002_schemas

-- ============================================================================
-- KB OWNERSHIP AND VISIBILITY
-- ============================================================================

-- Add visibility column (private=owner only, shared=explicit permissions, public=all users)
ALTER TABLE ai.knowledge_bases
    ADD COLUMN IF NOT EXISTS visibility TEXT DEFAULT 'private'
    CHECK (visibility IN ('private', 'shared', 'public'));

-- Add owner_id column
ALTER TABLE ai.knowledge_bases
    ADD COLUMN IF NOT EXISTS owner_id UUID REFERENCES auth.users(id) ON DELETE SET NULL;

-- Index for user's KBs
CREATE INDEX IF NOT EXISTS idx_ai_knowledge_bases_owner
    ON ai.knowledge_bases(owner_id) WHERE owner_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_ai_knowledge_bases_visibility
    ON ai.knowledge_bases(visibility) WHERE visibility != 'private';

-- ============================================================================
-- KB PERMISSIONS TABLE
-- ============================================================================

-- KB permissions table (for shared KBs)
CREATE TABLE IF NOT EXISTS ai.knowledge_base_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    knowledge_base_id UUID NOT NULL REFERENCES ai.knowledge_bases(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,

    -- Permission level
    permission TEXT NOT NULL CHECK (permission IN ('viewer', 'editor', 'owner')),

    -- Grant metadata
    granted_by UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    granted_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT unique_kb_user_permission UNIQUE (knowledge_base_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_ai_kb_permissions_kb
    ON ai.knowledge_base_permissions(knowledge_base_id);

CREATE INDEX IF NOT EXISTS idx_ai_kb_permissions_user
    ON ai.knowledge_base_permissions(user_id);

COMMENT ON TABLE ai.knowledge_base_permissions IS 'Granular permissions for shared knowledge bases';
COMMENT ON COLUMN ai.knowledge_bases.visibility IS 'private=owner only, shared=explicit permissions, public=all authenticated users';
COMMENT ON COLUMN ai.knowledge_base_permissions.permission IS 'viewer=read only, editor=read+write, owner=full control+manage permissions';
