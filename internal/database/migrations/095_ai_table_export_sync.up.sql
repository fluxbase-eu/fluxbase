-- Store table export sync configurations for knowledge bases
CREATE TABLE ai.table_export_sync_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    knowledge_base_id UUID NOT NULL REFERENCES ai.knowledge_bases(id) ON DELETE CASCADE,
    schema_name TEXT NOT NULL,
    table_name TEXT NOT NULL,
    columns TEXT[] DEFAULT NULL,  -- NULL means all columns
    sync_mode TEXT NOT NULL DEFAULT 'manual' CHECK (sync_mode IN ('manual', 'automatic')),
    sync_on_insert BOOLEAN DEFAULT true,
    sync_on_update BOOLEAN DEFAULT true,
    sync_on_delete BOOLEAN DEFAULT false,
    debounce_seconds INTEGER DEFAULT 60,
    include_foreign_keys BOOLEAN DEFAULT true,
    include_indexes BOOLEAN DEFAULT false,
    last_sync_at TIMESTAMPTZ,
    last_sync_status TEXT,
    last_sync_error TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(knowledge_base_id, schema_name, table_name)
);

-- Indexes for common queries
CREATE INDEX idx_table_export_sync_kb ON ai.table_export_sync_configs(knowledge_base_id);
CREATE INDEX idx_table_export_sync_table ON ai.table_export_sync_configs(schema_name, table_name);
CREATE INDEX idx_table_export_sync_mode ON ai.table_export_sync_configs(sync_mode) WHERE sync_mode = 'automatic';

-- RLS policies
ALTER TABLE ai.table_export_sync_configs ENABLE ROW LEVEL SECURITY;

-- Users can manage sync configs for knowledge bases they own or have access to
CREATE POLICY "Users can view sync configs for their knowledge bases"
    ON ai.table_export_sync_configs FOR SELECT
    USING (knowledge_base_id IN (
        SELECT id FROM ai.knowledge_bases
        WHERE owner_id = auth.uid() OR owner_id IS NULL
        OR visibility = 'public'
        OR EXISTS (
            SELECT 1 FROM ai.documents d
            JOIN ai.document_permissions dp ON dp.document_id = d.id
            WHERE d.knowledge_base_id = ai.table_export_sync_configs.knowledge_base_id
            AND dp.user_id = auth.uid()
        )
    ));

CREATE POLICY "Users can insert sync configs for their knowledge bases"
    ON ai.table_export_sync_configs FOR INSERT
    WITH CHECK (knowledge_base_id IN (
        SELECT id FROM ai.knowledge_bases
        WHERE owner_id = auth.uid() OR owner_id IS NULL
        OR EXISTS (
            SELECT 1 FROM ai.documents d
            JOIN ai.document_permissions dp ON dp.document_id = d.id
            WHERE d.knowledge_base_id = ai.table_export_sync_configs.knowledge_base_id
            AND dp.user_id = auth.uid()
            AND dp.permission = 'editor'
        )
    ));

CREATE POLICY "Users can update sync configs for their knowledge bases"
    ON ai.table_export_sync_configs FOR UPDATE
    USING (knowledge_base_id IN (
        SELECT id FROM ai.knowledge_bases
        WHERE owner_id = auth.uid() OR owner_id IS NULL
        OR EXISTS (
            SELECT 1 FROM ai.documents d
            JOIN ai.document_permissions dp ON dp.document_id = d.id
            WHERE d.knowledge_base_id = ai.table_export_sync_configs.knowledge_base_id
            AND dp.user_id = auth.uid()
            AND dp.permission = 'editor'
        )
    ));

CREATE POLICY "Users can delete sync configs for their knowledge bases"
    ON ai.table_export_sync_configs FOR DELETE
    USING (knowledge_base_id IN (
        SELECT id FROM ai.knowledge_bases
        WHERE owner_id = auth.uid() OR owner_id IS NULL
        OR EXISTS (
            SELECT 1 FROM ai.documents d
            JOIN ai.document_permissions dp ON dp.document_id = d.id
            WHERE d.knowledge_base_id = ai.table_export_sync_configs.knowledge_base_id
            AND dp.user_id = auth.uid()
            AND dp.permission = 'editor'
        )
    ));

-- Service role can manage all sync configs
CREATE POLICY "Service role can manage all sync configs"
    ON ai.table_export_sync_configs FOR ALL
    TO service_role
    USING (true)
    WITH CHECK (true);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION ai.update_table_export_sync_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_table_export_sync_updated_at
    BEFORE UPDATE ON ai.table_export_sync_configs
    FOR EACH ROW
    EXECUTE FUNCTION ai.update_table_export_sync_updated_at();
