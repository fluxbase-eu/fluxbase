-- Migration 093: Enhanced Chatbot-Knowledge Base Links
-- Adds tiered access, query routing, context weighting, and trace IDs

-- Drop existing simple table (we're in development, no production data)
DROP TABLE IF EXISTS ai.chatbot_knowledge_bases CASCADE;

-- Create enhanced chatbot-KB link table with all new features
CREATE TABLE ai.chatbot_knowledge_bases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chatbot_id UUID NOT NULL REFERENCES ai.chatbots(id) ON DELETE CASCADE,
    knowledge_base_id UUID NOT NULL REFERENCES ai.knowledge_bases(id) ON DELETE CASCADE,

    -- Access level: full (all chunks), filtered (with expression), tiered (by priority)
    access_level TEXT NOT NULL DEFAULT 'full' CHECK (access_level IN ('full', 'filtered', 'tiered')),

    -- Filter expression (JSONB) for filtered access
    -- Example: {"metadata": {"category": "support"}}
    filter_expression JSONB DEFAULT '{}',

    -- Context weight (0.0-1.0) for priority ordering
    -- Higher weight = higher priority in retrieval
    context_weight FLOAT NOT NULL DEFAULT 1.0 CHECK (context_weight >= 0.0 AND context_weight <= 1.0),

    -- Priority for tiered access (lower = higher priority)
    -- Only used when access_level = 'tiered'
    priority INTEGER DEFAULT 100 CHECK (priority >= 1 AND priority <= 1000),

    -- Intent routing: KB is selected only when query matches these keywords
    -- Example: ['technical', 'api', 'troubleshooting']
    intent_keywords TEXT[] DEFAULT ARRAY[]::TEXT[],

    -- Max chunks to retrieve from this KB (overrides chatbot default)
    -- NULL = use chatbot default
    max_chunks INTEGER CHECK (max_chunks > 0 OR max_chunks IS NULL),

    -- Similarity threshold for this KB (overrides chatbot default)
    -- NULL = use chatbot default
    similarity_threshold FLOAT CHECK (similarity_threshold >= 0.0 AND similarity_threshold <= 1.0 OR similarity_threshold IS NULL),

    -- Is this link enabled?
    enabled BOOLEAN NOT NULL DEFAULT true,

    -- Metadata for extensibility
    metadata JSONB DEFAULT '{}',

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Ensure unique chatbot-KB pairs
    CONSTRAINT chatbot_kb_unique UNIQUE (chatbot_id, knowledge_base_id)
);

-- Create index for efficient querying
CREATE INDEX chatbot_kb_links_chatbot_idx ON ai.chatbot_knowledge_bases(chatbot_id) WHERE enabled = true;
CREATE INDEX chatbot_kb_links_kb_idx ON ai.chatbot_knowledge_bases(knowledge_base_id) WHERE enabled = true;
CREATE INDEX chatbot_kb_links_priority_idx ON ai.chatbot_knowledge_bases(chatbot_id, priority) WHERE enabled = true AND access_level = 'tiered';

-- Add trace_id column to execution logs for Langfuse integration
-- Note: ai.execution_logs was migrated to logging.entries, so this only runs if the table still exists
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'ai' AND table_name = 'execution_logs') THEN
        ALTER TABLE ai.execution_logs
        ADD COLUMN IF NOT EXISTS trace_id TEXT,
        ADD COLUMN IF NOT EXISTS span_id TEXT;

        CREATE INDEX IF NOT EXISTS execution_logs_trace_idx ON ai.execution_logs(trace_id) WHERE trace_id IS NOT NULL;
    END IF;
END $$;

-- Create function to generate trace IDs
CREATE OR REPLACE FUNCTION ai.generate_trace_id()
RETURNS TEXT AS $$
BEGIN
    RETURN 'trace_' || encode(gen_random_bytes(16), 'hex');
END;
$$ LANGUAGE plpgsql;

-- Create function to generate span IDs
CREATE OR REPLACE FUNCTION ai.generate_span_id()
RETURNS TEXT AS $$
BEGIN
    RETURN 'span_' || encode(gen_random_bytes(8), 'hex');
END;
$$ LANGUAGE plpgsql;

-- Add updated_at trigger
CREATE OR REPLACE FUNCTION ai.update_chatbot_kb_link_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER chatbot_kb_link_updated_at
    BEFORE UPDATE ON ai.chatbot_knowledge_bases
    FOR EACH ROW
    EXECUTE FUNCTION ai.update_chatbot_kb_link_updated_at();

-- Add RLS policies
ALTER TABLE ai.chatbot_knowledge_bases ENABLE ROW LEVEL SECURITY;

-- Admins can do everything
CREATE POLICY chatbot_kb_links_admin_all ON ai.chatbot_knowledge_bases
    FOR ALL
    TO authenticated
    USING (
        EXISTS (
            SELECT 1 FROM auth.users
            WHERE auth.users.id = auth.current_user_id()
            AND auth.users.role = 'admin'
        )
    )
    WITH CHECK (
        EXISTS (
            SELECT 1 FROM auth.users
            WHERE auth.users.id = auth.current_user_id()
            AND auth.users.role = 'admin'
        )
    );

-- Service role can do everything
CREATE POLICY chatbot_kb_links_service_all ON ai.chatbot_knowledge_bases
    FOR ALL
    TO authenticated
    USING (
        EXISTS (
            SELECT 1 FROM auth.users
            WHERE auth.users.id = auth.current_user_id()
            AND auth.users.role = 'service_role'
        )
    )
    WITH CHECK (
        EXISTS (
            SELECT 1 FROM auth.users
            WHERE auth.users.id = auth.current_user_id()
            AND auth.users.role = 'service_role'
        )
    );
