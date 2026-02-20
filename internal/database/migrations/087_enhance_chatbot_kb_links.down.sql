-- Rollback Migration 093: Enhanced Chatbot-Knowledge Base Links

-- Drop RLS policies
DROP POLICY IF EXISTS chatbot_kb_links_service_all ON ai.chatbot_knowledge_bases;
DROP POLICY IF EXISTS chatbot_kb_links_admin_all ON ai.chatbot_knowledge_bases;

-- Disable RLS
ALTER TABLE ai.chatbot_knowledge_bases DISABLE ROW LEVEL SECURITY;

-- Drop trigger and function
DROP TRIGGER IF EXISTS chatbot_kb_link_updated_at ON ai.chatbot_knowledge_bases;
DROP FUNCTION IF EXISTS ai.update_chatbot_kb_link_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS ai.execution_logs_trace_idx;
DROP INDEX IF EXISTS ai.chatbot_kb_links_priority_idx;
DROP INDEX IF EXISTS ai.chatbot_kb_links_kb_idx;
DROP INDEX IF EXISTS ai.chatbot_kb_links_chatbot_idx;

-- Drop trace ID columns from execution_logs (only if table exists)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'ai' AND table_name = 'execution_logs') THEN
        ALTER TABLE ai.execution_logs DROP COLUMN IF EXISTS span_id;
        ALTER TABLE ai.execution_logs DROP COLUMN IF EXISTS trace_id;
    END IF;
END $$;

-- Drop helper functions
DROP FUNCTION IF EXISTS ai.generate_span_id();
DROP FUNCTION IF EXISTS ai.generate_trace_id();

-- Recreate simple table (original state)
DROP TABLE IF EXISTS ai.chatbot_knowledge_bases CASCADE;

CREATE TABLE ai.chatbot_knowledge_bases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chatbot_id UUID NOT NULL REFERENCES ai.chatbots(id) ON DELETE CASCADE,
    knowledge_base_id UUID NOT NULL REFERENCES ai.knowledge_bases(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT true,
    max_chunks INTEGER NOT NULL DEFAULT 5,
    similarity_threshold FLOAT NOT NULL DEFAULT 0.7,
    priority INTEGER NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chatbot_kb_unique UNIQUE (chatbot_id, knowledge_base_id)
);

CREATE INDEX chatbot_kb_links_chatbot_idx ON ai.chatbot_knowledge_bases(chatbot_id);
CREATE INDEX chatbot_kb_links_kb_idx ON ai.chatbot_knowledge_bases(knowledge_base_id);
