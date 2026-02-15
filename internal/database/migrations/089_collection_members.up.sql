-- Collection Members Table Migration
-- Manages who can access shared collections
-- Migration 089
-- ========================================================================
-- TABLE: AI.COLLECTION_MEMBERS
-- ========================================================================
CREATE TABLE IF NOT EXISTS ai.collection_members (
    collection_id UUID NOT NULL REFERENCES ai.collections(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('viewer', 'editor', 'owner')),
    added_by UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (collection_id, user_id)
);

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX idx_ai_collection_members_collection ON ai.collection_members(collection_id);
CREATE INDEX idx_ai_collection_members_user ON ai.collection_members(user_id);
CREATE INDEX idx_ai_collection_members_role ON ai.collection_members(role);

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE ai.collection_members IS 'Collection access control - manages who can view/edit shared collections';
COMMENT ON COLUMN ai.collection_members.role IS 'viewer=can view only, editor=can view and add/remove KBs, owner=full control including managing members';
