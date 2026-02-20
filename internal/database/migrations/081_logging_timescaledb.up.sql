-- TimescaleDB Support for Logging System
--
-- This migration creates the TimescaleDB extension (if available) and documents
-- TimescaleDB support for the logging system. TimescaleDB provides enhanced
-- features for time-series log data including:
-- - Automatic time-based partitioning (hypertables)
-- - Compression of old data
-- - Automated retention policies
-- - Improved query performance for time-series data
--
-- The TimescaleDB extension will be created if it's installed on the PostgreSQL
-- server. If TimescaleDB is not available, this migration will succeed without
-- creating the extension, and the logging system will fall back to regular
-- PostgreSQL storage.
--
-- See: https://docs.timescale.com/self-hosted/latest/install/

-- Create TimescaleDB extension only if it's available on the server
-- We first check pg_available_extensions to avoid connection errors
-- when the extension is not installed
DO $$
BEGIN
    -- Check if timescaledb is available before attempting to create it
    IF EXISTS (SELECT 1 FROM pg_available_extensions WHERE name = 'timescaledb') THEN
        BEGIN
            CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;
            RAISE NOTICE 'TimescaleDB extension created successfully';
        EXCEPTION
            WHEN OTHERS THEN
                RAISE NOTICE 'TimescaleDB extension could not be created: %, logging system will use regular PostgreSQL', SQLERRM;
        END;
    ELSE
        RAISE NOTICE 'TimescaleDB extension not available on this server, logging system will use regular PostgreSQL';
    END IF;
END $$;

-- FOR DEVELOPERS:
-- The application runtime code (internal/storage/log_timescaledb.go) handles
-- TimescaleDB initialization automatically when configured with:
--   logging.backend: timescaledb or postgres-timescaledb
--
-- The application will:
-- 1. Verify the TimescaleDB extension exists (already created by this migration if available)
-- 2. Convert logging.entries to a hypertable
-- 3. Enable compression and retention policies (if configured)
--
-- WITHOUT TIMESCALEDB:
-- If TimescaleDB is not installed, the logging system will fall back to
-- regular PostgreSQL storage using native table partitioning. All logging
-- features will work correctly, just without TimescaleDB optimizations.
