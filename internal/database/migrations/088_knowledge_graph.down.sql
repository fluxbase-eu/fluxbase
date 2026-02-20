-- Rollback Migration 094: Knowledge Graph

-- Drop RLS policies
DROP POLICY IF EXISTS document_entities_service_all ON ai.document_entities;
DROP POLICY IF EXISTS document_entities_admin_all ON ai.document_entities;
DROP POLICY IF EXISTS relationships_service_all ON ai.entity_relationships;
DROP POLICY IF EXISTS relationships_admin_all ON ai.entity_relationships;
DROP POLICY IF EXISTS entities_service_all ON ai.entities;
DROP POLICY IF EXISTS entities_admin_all ON ai.entities;

-- Disable RLS
ALTER TABLE ai.document_entities DISABLE ROW LEVEL SECURITY;
ALTER TABLE ai.entity_relationships DISABLE ROW LEVEL SECURITY;
ALTER TABLE ai.entities DISABLE ROW LEVEL SECURITY;

-- Drop trigger and function
DROP TRIGGER IF EXISTS entities_updated_at ON ai.entities;
DROP FUNCTION IF EXISTS ai.update_entities_updated_at();

-- Drop functions
DROP FUNCTION IF EXISTS ai.search_entities(UUID, TEXT, TEXT[], INTEGER);
DROP FUNCTION IF EXISTS ai.find_related_entities(UUID, UUID, INTEGER, TEXT[]);

-- Drop indexes
DROP INDEX IF EXISTS ai.document_entities_salience_idx;
DROP INDEX IF EXISTS ai.document_entities_entity_idx;
DROP INDEX IF EXISTS ai.document_entities_doc_idx;
DROP INDEX IF EXISTS ai.relationships_type_idx;
DROP INDEX IF EXISTS ai.relationships_target_idx;
DROP INDEX IF EXISTS ai.relationships_source_idx;
DROP INDEX IF EXISTS ai.relationships_kb_idx;
DROP INDEX IF EXISTS ai.entities_name_gin_idx;
DROP INDEX IF EXISTS ai.entities_name_idx;
DROP INDEX IF EXISTS ai.entities_type_idx;
DROP INDEX IF EXISTS ai.entities_kb_idx;

-- Drop tables (in order of dependencies)
DROP TABLE IF EXISTS ai.document_entities CASCADE;
DROP TABLE IF EXISTS ai.entity_relationships CASCADE;
DROP TABLE IF EXISTS ai.entities CASCADE;
