-- Add pipeline configuration to knowledge bases for document transformations
BEGIN;

-- Add pipeline configuration columns with inline CHECK constraint
ALTER TABLE ai.knowledge_bases
ADD COLUMN IF NOT EXISTS pipeline_type TEXT DEFAULT 'none' NOT NULL CHECK (pipeline_type IN ('none', 'sql', 'edge_function', 'webhook')),
ADD COLUMN IF NOT EXISTS pipeline_config JSONB DEFAULT '{}' NOT NULL,
ADD COLUMN IF NOT EXISTS transformation_function TEXT;

COMMENT ON COLUMN ai.knowledge_bases.pipeline_type IS 'Type of transformation pipeline: none, sql, edge_function, or webhook';
COMMENT ON COLUMN ai.knowledge_bases.pipeline_config IS 'Configuration for the pipeline (function name, webhook URL, etc.)';
COMMENT ON COLUMN ai.knowledge_bases.transformation_function IS 'Name of SQL transformation function (for pipeline_type=sql)';

COMMIT;
