-- Add Collection Reference to Knowledge Bases
-- Transitions KBs from user-owned to collection-based
-- Migration 086
-- ========================================================================
-- ADD COLLECTION_ID COLUMN
-- ========================================================================
-- Add collection_id as nullable initially (transition period)
ALTER TABLE ai.knowledge_bases
    ADD COLUMN IF NOT EXISTS collection_id UUID REFERENCES ai.collections(id) ON DELETE CASCADE;

-- During transition, KB can have owner_id OR collection_id (one must be set)
ALTER TABLE ai.knowledge_bases
    DROP CONSTRAINT IF EXISTS kb_owner_or_collection;

-- Add constraint to ensure KB has either owner_id or collection_id
ALTER TABLE ai.knowledge_bases
    ADD CONSTRAINT kb_owner_or_collection
    CHECK (owner_id IS NOT NULL OR collection_id IS NOT NULL);

-- ============================================================================
-- INDEX
-- ============================================================================

CREATE INDEX idx_ai_knowledge_bases_collection
    ON ai.knowledge_bases(collection_id)
    WHERE collection_id IS NOT NULL;

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON COLUMN ai.knowledge_bases.collection_id IS 'Collection this KB belongs to (replaces direct user ownership)';
