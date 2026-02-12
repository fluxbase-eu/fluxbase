-- Rollback TimescaleDB extension
-- Warning: This will fail if the logging.entries table has been converted to a hypertable
-- The table must be converted back to a regular table before dropping TimescaleDB

-- Drop the TimescaleDB extension ( CASCADE removes dependent objects )
DROP EXTENSION IF EXISTS timescaledb CASCADE;
