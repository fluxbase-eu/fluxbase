-- Enable TimescaleDB extension for the logging system
-- This migration enables TimescaleDB features for improved time-series data handling
-- including automatic partitioning, compression, and retention policies

-- Create TimescaleDB extension (if not already installed)
-- CASCADE automatically installs dependent extensions
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

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

-- Create comments for documentation
COMMENT ON EXTENSION timescaledb IS 'TimescaleDB for time-series optimization of log data';
