-- Drop trigger and function
DROP TRIGGER IF EXISTS trigger_update_table_export_sync_updated_at ON ai.table_export_sync_configs;
DROP FUNCTION IF EXISTS ai.update_table_export_sync_updated_at();

-- Drop table
DROP TABLE IF EXISTS ai.table_export_sync_configs;
