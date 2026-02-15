-- Collection-Knowledge Base Junction Table Migration
-- Creates many-to-many relationship between collections and knowledge bases
-- Migration 087
-- ========================================================================
-- TABLE: AI.COLLECTION_KNOWLEDGE_BASES
-- ========================================================================
CREATE TABLE IF NOT EXISTS ai.collection_knowledge_bases (
    collection_id UUID NOT NULL REFERENCES ai.collections(id) ON DELETE CASCADE,
    knowledge_base_id UUID NOT NULL REFERENCES ai.knowledge_bases(id) ON DELETE CASCADE,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (collection_id, knowledge_base_id)
);

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX idx_ai_collection_kbs_collection ON ai.collection_knowledge_bases(collection_id);
CREATE INDEX idx_ai_collection_kbs_kb ON ai.collection_knowledge_bases(knowledge_base_id);

COMMENT ON TABLE ai.collection_knowledge_bases IS 'Links knowledge bases to collections';
