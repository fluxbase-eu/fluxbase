-- Add user quotas for knowledge base resource management
BEGIN;

-- Create user quotas table
CREATE TABLE IF NOT EXISTS ai.user_quotas (
    user_id UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
    max_documents INTEGER DEFAULT 10000 NOT NULL,
    max_chunks INTEGER DEFAULT 500000 NOT NULL,
    max_storage_bytes BIGINT DEFAULT 10737418240 NOT NULL, -- 10GB default
    used_documents INTEGER DEFAULT 0 NOT NULL,
    used_chunks INTEGER DEFAULT 0 NOT NULL,
    used_storage_bytes BIGINT DEFAULT 0 NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CHECK (max_documents >= 0),
    CHECK (max_chunks >= 0),
    CHECK (max_storage_bytes >= 0),
    CHECK (used_documents >= 0),
    CHECK (used_chunks >= 0),
    CHECK (used_storage_bytes >= 0)
);

-- Add index for faster quota lookups
CREATE INDEX idx_ai_user_quotas_user_id ON ai.user_quotas(user_id);

-- Add quota columns to knowledge bases with inline CHECK constraints
ALTER TABLE ai.knowledge_bases
ADD COLUMN IF NOT EXISTS quota_max_documents INTEGER DEFAULT 1000 NOT NULL CHECK (quota_max_documents >= 0),
ADD COLUMN IF NOT EXISTS quota_max_chunks INTEGER DEFAULT 50000 NOT NULL CHECK (quota_max_chunks >= 0),
ADD COLUMN IF NOT EXISTS quota_max_storage_bytes BIGINT DEFAULT 1073741824 NOT NULL CHECK (quota_max_storage_bytes >= 0);

-- Create indexes for quota enforcement
CREATE INDEX IF NOT EXISTS idx_ai_knowledge_bases_owner_quotas ON ai.knowledge_bases(owner_id)
WHERE owner_id IS NOT NULL;

COMMENT ON TABLE ai.user_quotas IS 'Per-user resource quotas for knowledge bases';
COMMENT ON COLUMN ai.user_quotas.max_documents IS 'Maximum number of documents allowed across all user KBs';
COMMENT ON COLUMN ai.user_quotas.max_chunks IS 'Maximum number of chunks allowed across all user KBs';
COMMENT ON COLUMN ai.user_quotas.max_storage_bytes IS 'Maximum storage in bytes allowed across all user KBs';
COMMENT ON COLUMN ai.user_quotas.used_documents IS 'Current document count across all user KBs';
COMMENT ON COLUMN ai.user_quotas.used_chunks IS 'Current chunk count across all user KBs';
COMMENT ON COLUMN ai.user_quotas.used_storage_bytes IS 'Current storage in bytes across all user KBs';

COMMENT ON COLUMN ai.knowledge_bases.quota_max_documents IS 'Maximum documents allowed in this KB';
COMMENT ON COLUMN ai.knowledge_bases.quota_max_chunks IS 'Maximum chunks allowed in this KB';
COMMENT ON COLUMN ai.knowledge_bases.quota_max_storage_bytes IS 'Maximum storage in bytes allowed in this KB';

COMMIT;
