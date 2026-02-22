-- Re-add default_user_permission column
ALTER TABLE ai.knowledge_bases ADD COLUMN default_user_permission TEXT
    CHECK (default_user_permission IN ('viewer', 'editor', 'owner')) DEFAULT 'viewer';

COMMENT ON COLUMN ai.knowledge_bases.default_user_permission IS 'Default permission level for all authenticated users (viewer/editor/owner)';
