-- Collections Table Migration
-- Creates shared collections for organizing knowledge bases
-- Migration 085
-- ========================================================================
-- TABLE: AI.COLLECTIONS
-- ========================================================================
CREATE TABLE IF NOT EXISTS ai.collections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Basic info
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    description TEXT,

    -- Creator tracking (not ownership)
    created_by UUID REFERENCES auth.users(id) ON DELETE SET NULL,

    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Unique constraint: slug is globally unique
    UNIQUE (slug)
);

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX idx_ai_collections_created_by ON ai.collections(created_by);

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE ai.collections IS 'Shared collections for organizing knowledge bases - users can be members via collection_members table';
COMMENT ON COLUMN ai.collections.slug IS 'URL-friendly identifier (globally unique)';
COMMENT ON COLUMN ai.collections.created_by IS 'User who created this collection (not owner - collection is shared)';
