-- Migration 094: Knowledge Graph for Enhanced RAG
-- Adds entities, relationships, and graph traversal capabilities

-- Entities table: extracted from documents (people, organizations, locations, concepts)
CREATE TABLE IF NOT EXISTS ai.entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    knowledge_base_id UUID NOT NULL REFERENCES ai.knowledge_bases(id) ON DELETE CASCADE,

    -- Entity details
    entity_type TEXT NOT NULL CHECK (entity_type IN ('person', 'organization', 'location', 'concept', 'product', 'event', 'other')),
    name TEXT NOT NULL,
    canonical_name TEXT, -- Normalized name (e.g., "Apple Inc." instead of "Apple")

    -- Optional: aliases/synonyms for this entity
    aliases TEXT[] DEFAULT ARRAY[]::TEXT[],

    -- Metadata about the entity (confidence, source, etc.)
    metadata JSONB DEFAULT '{}',

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Ensure unique entity per KB (same entity type + canonical name)
    CONSTRAINT entity_unique UNIQUE (knowledge_base_id, entity_type, canonical_name)
);

-- Relationships table: connections between entities
CREATE TABLE IF NOT EXISTS ai.entity_relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    knowledge_base_id UUID NOT NULL REFERENCES ai.knowledge_bases(id) ON DELETE CASCADE,

    -- Relationship endpoints
    source_entity_id UUID NOT NULL REFERENCES ai.entities(id) ON DELETE CASCADE,
    target_entity_id UUID NOT NULL REFERENCES ai.entities(id) ON DELETE CASCADE,

    -- Relationship type
    relationship_type TEXT NOT NULL CHECK (relationship_type IN (
        'works_at', 'located_in', 'founded_by', 'owns', 'part_of',
        'related_to', 'knows', 'customer_of', 'supplier_of',
        'invested_in', 'acquired', 'merged_with', 'competitor_of',
        'parent_of', 'child_of', 'spouse_of', 'sibling_of',
        'other'
    )),

    -- Relationship direction (forward, backward, or bidirectional)
    direction TEXT NOT NULL DEFAULT 'forward' CHECK (direction IN ('forward', 'backward', 'bidirectional')),

    -- Optional: confidence score for this relationship
    confidence FLOAT CHECK (confidence >= 0.0 AND confidence <= 1.0),

    -- Metadata about the relationship
    metadata JSONB DEFAULT '{}',

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Prevent duplicate relationships
    CONSTRAINT relationship_unique UNIQUE (knowledge_base_id, source_entity_id, target_entity_id, relationship_type),

    -- Prevent self-relationships
    CONSTRAINT no_self_relationship CHECK (source_entity_id != target_entity_id)
);

-- Document-Entity mentions: tracks which entities appear in which documents
CREATE TABLE IF NOT EXISTS ai.document_entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES ai.documents(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES ai.entities(id) ON DELETE CASCADE,

    -- Mention details
    mention_count INTEGER DEFAULT 1, -- How many times entity appears in document
    first_mention_offset INTEGER,   -- Character offset of first mention
    salience FLOAT DEFAULT 0.0,      -- Importance/relevance score (0-1)

    -- Context snippet where entity was found
    context TEXT,

    -- Timestamp
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Prevent duplicate document-entity pairs
    CONSTRAINT document_entity_unique UNIQUE (document_id, entity_id)
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS entities_kb_idx ON ai.entities(knowledge_base_id);
CREATE INDEX IF NOT EXISTS entities_type_idx ON ai.entities(entity_type);
CREATE INDEX IF NOT EXISTS entities_name_idx ON ai.entities(canonical_name);
CREATE INDEX IF NOT EXISTS entities_name_gin_idx ON ai.entities USING gin(to_tsvector('english', canonical_name));

CREATE INDEX IF NOT EXISTS relationships_kb_idx ON ai.entity_relationships(knowledge_base_id);
CREATE INDEX IF NOT EXISTS relationships_source_idx ON ai.entity_relationships(source_entity_id);
CREATE INDEX IF NOT EXISTS relationships_target_idx ON ai.entity_relationships(target_entity_id);
CREATE INDEX IF NOT EXISTS relationships_type_idx ON ai.entity_relationships(relationship_type);

CREATE INDEX IF NOT EXISTS document_entities_doc_idx ON ai.document_entities(document_id);
CREATE INDEX IF NOT EXISTS document_entities_entity_idx ON ai.document_entities(entity_id);
CREATE INDEX IF NOT EXISTS document_entities_salience_idx ON ai.document_entities(salience DESC);

-- Function to find related entities (graph traversal)
CREATE OR REPLACE FUNCTION ai.find_related_entities(
    p_kb_id UUID,
    p_entity_id UUID,
    p_max_depth INTEGER DEFAULT 2,
    p_relationship_types TEXT[] DEFAULT NULL
)
RETURNS TABLE (
    entity_id UUID,
    entity_type TEXT,
    name TEXT,
    canonical_name TEXT,
    relationship_type TEXT,
    depth INTEGER,
    path TEXT[] -- Array of entity IDs showing the traversal path
) AS $$
DECLARE
    v_max_depth INTEGER := GREATEST(LEAST(p_max_depth, 5), 1); -- Limit to depth 5
BEGIN
    RETURN QUERY
    WITH RECURSIVE graph_traversal AS (
        -- Base case: direct relationships
        SELECT
            e.id,
            e.entity_type,
            e.name,
            e.canonical_name,
            r.relationship_type,
            1::INTEGER as depth,
            ARRAY[p_entity_id, e.id]::UUID[] as path
        FROM ai.entity_relationships r
        JOIN ai.entities e ON e.id = r.target_entity_id
        WHERE r.source_entity_id = p_entity_id
            AND r.knowledge_base_id = p_kb_id
            AND (p_relationship_types IS NULL OR r.relationship_type = ANY(p_relationship_types))

        UNION ALL

        -- Recursive case: traverse to depth N
        SELECT
            e.id,
            e.entity_type,
            e.name,
            e.canonical_name,
            r.relationship_type,
            gt.depth + 1,
            gt.path || e.id
        FROM ai.entity_relationships r
        JOIN ai.entities e ON e.id = r.target_entity_id
        JOIN graph_traversal gt ON gt.entity_id = r.source_entity_id
        WHERE r.knowledge_base_id = p_kb_id
            AND gt.depth < v_max_depth
            AND (p_relationship_types IS NULL OR r.relationship_type = ANY(p_relationship_types))
            AND NOT (e.id = ANY(gt.path)) -- Prevent cycles
    )
    SELECT * FROM graph_traversal
    ORDER BY depth, relationship_type, name;
END;
$$ LANGUAGE plpgsql;

-- Function to search entities by name (fuzzy matching)
CREATE OR REPLACE FUNCTION ai.search_entities(
    p_kb_id UUID,
    p_query TEXT,
    p_entity_types TEXT[] DEFAULT NULL,
    p_limit INTEGER DEFAULT 20
)
RETURNS TABLE (
    entity_id UUID,
    entity_type TEXT,
    name TEXT,
    canonical_name TEXT,
    aliases TEXT[],
    rank REAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        e.id,
        e.entity_type,
        e.name,
        e.canonical_name,
        e.aliases,
        CASE
            WHEN e.canonical_name ILIKE p_query || '%' THEN 1.0
            WHEN e.name ILIKE p_query || '%' THEN 0.9
            WHEN e.canonical_name ILIKE '%' || p_query || '%' THEN 0.7
            WHEN EXISTS (SELECT 1 FROM unnest(e.aliases) alias WHERE alias ILIKE '%' || p_query || '%') THEN 0.6
            ELSE ts_rank(to_tsvector('english', e.canonical_name), plainto_tsquery('english', p_query)) * 0.5
        END::REAL as rank
    FROM ai.entities e
    WHERE e.knowledge_base_id = p_kb_id
        AND (p_entity_types IS NULL OR e.entity_type = ANY(p_entity_types))
        AND (
            e.canonical_name ILIKE '%' || p_query || '%'
            OR e.name ILIKE '%' || p_query || '%'
            OR EXISTS (SELECT 1 FROM unnest(e.aliases) alias WHERE alias ILIKE '%' || p_query || '%')
            OR to_tsvector('english', e.canonical_name) @@ plainto_tsquery('english', p_query)
        )
    ORDER BY rank DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION ai.update_entities_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER entities_updated_at
    BEFORE UPDATE ON ai.entities
    FOR EACH ROW
    EXECUTE FUNCTION ai.update_entities_updated_at();

-- RLS policies
ALTER TABLE ai.entities ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai.entity_relationships ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai.document_entities ENABLE ROW LEVEL SECURITY;

-- Admins can do everything
CREATE POLICY entities_admin_all ON ai.entities
    FOR ALL TO authenticated
    USING (EXISTS (SELECT 1 FROM auth.users WHERE auth.users.id = auth.current_user_id() AND auth.users.role = 'admin'))
    WITH CHECK (EXISTS (SELECT 1 FROM auth.users WHERE auth.users.id = auth.current_user_id() AND auth.users.role = 'admin'));

CREATE POLICY relationships_admin_all ON ai.entity_relationships
    FOR ALL TO authenticated
    USING (EXISTS (SELECT 1 FROM auth.users WHERE auth.users.id = auth.current_user_id() AND auth.users.role = 'admin'))
    WITH CHECK (EXISTS (SELECT 1 FROM auth.users WHERE auth.users.id = auth.current_user_id() AND auth.users.role = 'admin'));

CREATE POLICY document_entities_admin_all ON ai.document_entities
    FOR ALL TO authenticated
    USING (EXISTS (SELECT 1 FROM auth.users WHERE auth.users.id = auth.current_user_id() AND auth.users.role = 'admin'))
    WITH CHECK (EXISTS (SELECT 1 FROM auth.users WHERE auth.users.id = auth.current_user_id() AND auth.users.role = 'admin'));

-- Service role can do everything
CREATE POLICY entities_service_all ON ai.entities
    FOR ALL TO authenticated
    USING (EXISTS (SELECT 1 FROM auth.users WHERE auth.users.id = auth.current_user_id() AND auth.users.role = 'service_role'))
    WITH CHECK (EXISTS (SELECT 1 FROM auth.users WHERE auth.users.id = auth.current_user_id() AND auth.users.role = 'service_role'));

CREATE POLICY relationships_service_all ON ai.entity_relationships
    FOR ALL TO authenticated
    USING (EXISTS (SELECT 1 FROM auth.users WHERE auth.users.id = auth.current_user_id() AND auth.users.role = 'service_role'))
    WITH CHECK (EXISTS (SELECT 1 FROM auth.users WHERE auth.users.id = auth.current_user_id() AND auth.users.role = 'service_role'));

CREATE POLICY document_entities_service_all ON ai.document_entities
    FOR ALL TO authenticated
    USING (EXISTS (SELECT 1 FROM auth.users WHERE auth.users.id = auth.current_user_id() AND auth.users.role = 'service_role'))
    WITH CHECK (EXISTS (SELECT 1 FROM auth.users WHERE auth.users.id = auth.current_user_id() AND auth.users.role = 'service_role'));
