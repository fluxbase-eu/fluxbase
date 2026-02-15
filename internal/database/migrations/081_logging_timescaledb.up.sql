-- Enable TimescaleDB extension for the logging system
-- This migration enables TimescaleDB features for improved time-series data handling
-- including automatic partitioning, compression, and retention policies
--
-- Note: This migration is optional - if TimescaleDB is not available,
-- the logging system will fall back to regular PostgreSQL storage

DO $$
BEGIN
    -- Try to create TimescaleDB extension (if not already installed)
    -- This will fail gracefully if TimescaleDB is not available in the PostgreSQL instance
    BEGIN
        CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;
        RAISE NOTICE 'TimescaleDB extension enabled successfully';
    EXCEPTION
        WHEN OTHERS THEN
            RAISE NOTICE 'TimescaleDB extension not available, logging will use regular PostgreSQL';
    END;
END $$;

-- Note: The actual conversion to hypertable is handled at application runtime
-- by the TimescaleDBLogStorage initializer. This is necessary because:
--
-- 1. TimescaleDB hypertables are incompatible with PostgreSQL's declarative partitioning
-- 2. We need to drop existing partitions before converting to hypertable
-- 3. The application determines when/how to enable TimescaleDB based on configuration
--
-- To enable TimescaleDB features, the application will:
-- 1. Drop existing partition tables (entries_system, entries_http, etc.)
-- 2. Remove partitioning constraint from parent table
-- 3. Execute: SELECT create_hypertable('logging.entries', 'timestamp', if_not_exists => TRUE);
-- 4. Optionally enable compression: ALTER TABLE logging.entries SET (timescaledb.compress = TRUE);
-- 5. Add compression/retention policies as needed

-- Create comments for documentation (only if extension exists)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'timescaledb') THEN
        COMMENT ON EXTENSION timescaledb IS 'TimescaleDB for time-series optimization of log data';
    END IF;
END $$;
