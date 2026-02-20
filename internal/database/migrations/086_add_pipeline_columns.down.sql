-- Rollback pipeline columns
BEGIN;

ALTER TABLE ai.knowledge_bases
DROP CONSTRAINT IF EXISTS kb_pipeline_type_valid,
DROP COLUMN IF EXISTS transformation_function,
DROP COLUMN IF EXISTS pipeline_config,
DROP COLUMN IF EXISTS pipeline_type;

COMMIT;
