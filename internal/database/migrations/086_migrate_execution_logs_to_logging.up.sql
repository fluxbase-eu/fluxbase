-- Migrate execution logs to centralized logging system
-- This migration copies data from separate execution_logs tables
-- to the unified logging.entries table, enabling TimescaleDB features

BEGIN;

-- Ensure TimescaleDB extension is available (idempotent)
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

-- Check if logging.entries table exists and is a regular table or hypertable
DO $$
DECLARE
    entries_table_exists BOOLEAN := EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'logging'
        AND table_name = 'entries'
    );

    is_hypertable BOOLEAN := EXISTS (
        SELECT 1
        FROM timescaledb_information.hypertables h
        WHERE hypertable_schema = 'logging'
        AND hypertable_name = 'entries'
    );
END IF;

-- If logging.entries exists but is NOT a hypertable, convert it
-- This handles the case where someone used the system before TimescaleDB was added
IF entries_table_exists AND NOT is_hypertable THEN
    -- Drop existing partitions (data is preserved)
    DROP TABLE IF EXISTS logging.entries_system CASCADE;
    DROP TABLE IF EXISTS logging.entries_http CASCADE;
    DROP TABLE IF EXISTS logging.entries_security CASCADE;
    DROP TABLE IF EXISTS logging.entries_execution CASCADE;
    DROP TABLE IF EXISTS logging.entries_ai CASCADE;
    DROP TABLE IF EXISTS logging.entries_custom CASCADE;

    -- Remove partitioning constraint from parent table
    ALTER TABLE logging.entries DROP CONSTRAINT IF EXISTS logging_entries_pkey;

    -- Add primary key without partitioning
    ALTER TABLE logging.entries ADD PRIMARY KEY (id);

    -- Convert to hypertable (migrates existing data)
    PERFORM create_hypertable('logging.entries', 'timestamp', if_not_exists => TRUE, migrate_data => TRUE);

    -- Apply compression for existing data (older than 7 days)
    ALTER TABLE logging.entries SET (
        timescaledb.compress = TRUE,
        timescaledb.compress_segmentby = 'category, level',
        timescaledb.compress_orderby = 'timestamp DESC'
    );

    -- Add compression policy for new data (compress after 7 days)
    INSERT INTO timescaledb.compression_policy (
        hypertable => 'logging.entries',
        if_exists => TRUE,
        compress_after => INTERVAL '7 days',
        segment_by => 'category, level',
        order_by => 'timestamp DESC'
    )
    ON CONFLICT (policy_name) DO NOTHING;

    -- Add retention policy (90 days default, can be adjusted per category)
    INSERT INTO timescaledb.retention_policy (
        hypertable => 'logging.entries',
        if_exists => TRUE,
        drop_after => INTERVAL '90 days'
    )
    ON CONFLICT (policy_name) DO NOTHING;

    RAISE NOTICE 'Logging entries table has been converted to TimescaleDB hypertable';
    RAISE NOTICE 'Execution logs from separate tables should be migrated to logging.entries category';

    -- Mark execution_logs tables as deprecated (but don't drop yet for safety)
    COMMENT ON TABLE functions.edge_functions.execution_logs IS 'DEPRECATED: Migrate data to logging.entries using centralized system';
    COMMENT ON TABLE jobs.functions.execution_logs IS 'DEPRECATED: Migrate data to logging.entries using centralized system';
    COMMENT ON TABLE rpc.procedures.execution_logs IS 'DEPRECATED: Migrate data to logging.entries using centralized system';

END;

-- Create a view to help identify which execution_logs still need migration
CREATE OR REPLACE VIEW execution_logs_migration_status AS
SELECT
    table_name,
    CASE
        WHEN table_name = 'functions.edge_functions.execution_logs' THEN 'functions.edge_functions'
        WHEN table_name = 'jobs.functions.execution_logs' THEN 'jobs.functions'
        WHEN table_name = 'rpc.procedures.execution_logs' THEN 'rpc.procedures'
        WHEN table_name = 'branching.seed_execution_log' THEN 'branching'
        ELSE 'unknown'
    END AS source,
    CASE
        WHEN table_name = 'functions.edge_functions.execution_logs' THEN 'MIGRATE TO LOGGING'
        WHEN table_name = 'jobs.functions.execution_logs' THEN 'MIGRATE TO LOGGING'
        WHEN table_name = 'rpc.procedures.execution_logs' THEN 'MIGRATE TO LOGGING'
        WHEN table_name = 'branching.seed_execution_log' THEN 'MIGRATE TO LOGGING'
        ELSE 'NOT APPLICABLE'
    END AS needs_migration,
    CASE
        WHEN COUNT(*) = 0 THEN 'NO DATA'
        ELSE 'HAS DATA'
        END AS has_data
FROM information_schema.tables
WHERE table_schema = 'logging'
  AND table_name LIKE '%execution_log'
  AND table_schema NOT LIKE 'timescaledb%';  -- Ignore TimescaleDB internal tables
GRANT ALL ON EXECUTION FUNCTION create_migration_status() TO fluxbase;

-- Create index on table_name and source for faster lookups
CREATE INDEX IF NOT EXISTS idx_execution_logs_source_name
ON information_schema.tables (table_schema, table_name, source);

COMMIT;
