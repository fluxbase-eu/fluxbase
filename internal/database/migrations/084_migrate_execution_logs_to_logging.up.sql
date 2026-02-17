-- Migrate execution logs to centralized logging system
-- This migration copies data from separate execution_logs tables
-- to the unified logging.entries table, enabling TimescaleDB features if available

BEGIN;

-- Try to enable TimescaleDB (optional - fails gracefully if not available)
DO $$
BEGIN
    BEGIN
        CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;
        RAISE NOTICE 'TimescaleDB extension available';
    EXCEPTION
        WHEN OTHERS THEN
            RAISE NOTICE 'TimescaleDB extension not available, using regular PostgreSQL';
    END;
END $$;

-- Check if TimescaleDB is available
DO $$
DECLARE
    timescaledb_available BOOLEAN := EXISTS (
        SELECT 1
        FROM pg_extension
        WHERE extname = 'timescaledb'
    );

    entries_table_exists BOOLEAN := EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'logging'
        AND table_name = 'entries'
    );

    is_hypertable BOOLEAN := FALSE;
BEGIN
    -- Only proceed with TimescaleDB-specific logic if extension is available
    IF timescaledb_available AND entries_table_exists THEN
        -- Check if logging.entries is a hypertable
        SELECT EXISTS (
            SELECT 1
            FROM timescaledb_information.hypertables h
            WHERE hypertable_schema = 'logging'
            AND hypertable_name = 'entries'
        ) INTO is_hypertable;

        -- If logging.entries exists but is NOT a hypertable, convert it
        IF NOT is_hypertable THEN
            -- Drop existing partitions (data is preserved)
            DROP TABLE IF EXISTS logging.entries_system CASCADE;
            DROP TABLE IF EXISTS logging.entries_http CASCADE;
            DROP TABLE IF EXISTS logging.entries_security CASCADE;
            DROP TABLE IF EXISTS logging.entries_execution CASCADE;
            DROP TABLE IF EXISTS logging.entries_ai CASCADE;
            DROP TABLE IF EXISTS logging.entries_custom CASCADE;

            -- Remove partitioning constraint from parent table (if exists)
            -- Note: Can't use DROP CONSTRAINT IF EXISTS in older PostgreSQL versions
            IF EXISTS (
                SELECT 1 FROM pg_constraint
                WHERE conname = 'logging_entries_pkey'
                AND conrelid = 'logging.entries'::regclass
            ) THEN
                ALTER TABLE logging.entries DROP CONSTRAINT logging_entries_pkey;
            END IF;

            -- Add primary key without partitioning
            ALTER TABLE logging.entries ADD PRIMARY KEY (id);

            -- Convert to hypertable (migrates existing data)
            PERFORM create_hypertable('logging.entries', 'timestamp', if_not_exists := TRUE, migrate_data := TRUE);

            -- Apply compression for existing data (older than 7 days)
            ALTER TABLE logging.entries SET (
                timescaledb.compress = TRUE,
                timescaledb.compress_segmentby = 'category, level',
                timescaledb.compress_orderby = 'timestamp DESC'
            );

            -- Add compression policy for new data (compress after 7 days)
            PERFORM add_compression_policy('logging.entries', INTERVAL '7 days',
                compress_after := INTERVAL '7 days',
                if_not_exists := TRUE
            );

            -- Add retention policy (90 days default, can be adjusted per category)
            PERFORM add_retention_policy('logging.entries', INTERVAL '90 days',
                if_not_exists := TRUE
            );

            RAISE NOTICE 'Logging entries table has been converted to TimescaleDB hypertable';
        END IF;
    END IF;

    RAISE NOTICE 'Execution logs from separate tables should be migrated to logging.entries category';

    -- Mark execution_logs tables as deprecated (but don't drop yet for safety)
    -- Only add comments if tables exist
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'functions' AND table_name = 'execution_logs') THEN
        EXECUTE 'COMMENT ON TABLE functions.execution_logs IS ''DEPRECATED: Migrate data to logging.entries using centralized system''';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'jobs' AND table_name = 'execution_logs') THEN
        EXECUTE 'COMMENT ON TABLE jobs.execution_logs IS ''DEPRECATED: Migrate data to logging.entries using centralized system''';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'rpc' AND table_name = 'execution_logs') THEN
        EXECUTE 'COMMENT ON TABLE rpc.execution_logs IS ''DEPRECATED: Migrate data to logging.entries using centralized system''';
    END IF;
END $$;

-- Create a view to help identify which execution_logs still need migration
CREATE OR REPLACE VIEW logging.execution_logs_migration_status AS
SELECT
    table_schema,
    table_name,
    CASE
        WHEN table_schema = 'functions' AND table_name = 'execution_logs' THEN 'functions.edge_functions'
        WHEN table_schema = 'jobs' AND table_name = 'execution_logs' THEN 'jobs.functions'
        WHEN table_schema = 'rpc' AND table_name = 'execution_logs' THEN 'rpc.procedures'
        WHEN table_schema = 'branching' AND table_name = 'seed_execution_log' THEN 'branching'
        ELSE table_schema || '.' || table_name
    END AS source,
    CASE
        WHEN table_schema IN ('functions', 'jobs', 'rpc') AND table_name = 'execution_logs' THEN 'MIGRATE TO LOGGING'
        WHEN table_schema = 'branching' AND table_name = 'seed_execution_log' THEN 'MIGRATE TO LOGGING'
        ELSE 'NOT APPLICABLE'
    END AS needs_migration
FROM information_schema.tables
WHERE (table_schema, table_name) IN (
    ('functions', 'execution_logs'),
    ('jobs', 'execution_logs'),
    ('rpc', 'execution_logs'),
    ('branching', 'seed_execution_log')
);

COMMIT;
