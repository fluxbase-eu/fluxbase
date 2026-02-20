-- Rollback: Migrate execution logs to centralized logging system

BEGIN;

-- Drop the migration status view
DROP VIEW IF EXISTS logging.execution_logs_migration_status;

-- Remove comments marking tables as deprecated (only if tables exist)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'functions' AND table_name = 'execution_logs') THEN
        EXECUTE 'COMMENT ON TABLE functions.execution_logs IS NULL';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'jobs' AND table_name = 'execution_logs') THEN
        EXECUTE 'COMMENT ON TABLE jobs.execution_logs IS NULL';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'rpc' AND table_name = 'execution_logs') THEN
        EXECUTE 'COMMENT ON TABLE rpc.execution_logs IS NULL';
    END IF;
END $$;

COMMIT;
