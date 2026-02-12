package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/fluxbase-eu/fluxbase/internal/database"
)

// TimescaleDBLogStorage implements LogStorage using TimescaleDB (PostgreSQL extension).
// TimescaleDB provides hypertables with automatic partitioning, compression, and retention policies.
type TimescaleDBLogStorage struct {
	*PostgresLogStorage
	backendName string // "timescaledb" or "postgres-timescaledb"
}

// TimescaleDBConfig contains TimescaleDB-specific configuration.
type TimescaleDBConfig struct {
	// Enabled enables TimescaleDB features (hypertable conversion, compression)
	Enabled bool
	// Compressed enables compression of old data
	Compressed bool
	// CompressAfter specifies how long to wait before compressing data
	CompressAfter time.Duration
}

// newTimescaleDBLogStorage creates a new TimescaleDB-backed log storage.
// This is used when backend is explicitly "timescaledb" (requires a separate database).
func newTimescaleDBLogStorage(cfg TimescaleDBConfig, db *database.Connection) (*TimescaleDBLogStorage, error) {
	postgres := NewPostgresLogStorage(db)

	tsdb := &TimescaleDBLogStorage{
		PostgresLogStorage: postgres,
		backendName:        "timescaledb",
	}

	if err := tsdb.enableTimescaleDB(context.Background(), cfg); err != nil {
		return nil, fmt.Errorf("failed to enable TimescaleDB: %w", err)
	}

	return tsdb, nil
}

// newPostgresTimescaleDBStorage creates a TimescaleDB log storage using the main database connection.
// This is used when backend is "postgres-timescaledb" (uses main database with TimescaleDB extension).
func newPostgresTimescaleDBStorage(cfg TimescaleDBConfig, db *database.Connection) (*TimescaleDBLogStorage, error) {
	postgres := NewPostgresLogStorage(db)

	tsdb := &TimescaleDBLogStorage{
		PostgresLogStorage: postgres,
		backendName:        "postgres-timescaledb",
	}

	if err := tsdb.enableTimescaleDB(context.Background(), cfg); err != nil {
		return nil, fmt.Errorf("failed to enable TimescaleDB: %w", err)
	}

	return tsdb, nil
}

// Name returns the backend identifier.
func (s *TimescaleDBLogStorage) Name() string {
	return s.backendName
}

// enableTimescaleDB enables TimescaleDB features on the logging.entries table.
func (s *TimescaleDBLogStorage) enableTimescaleDB(ctx context.Context, cfg TimescaleDBConfig) error {
	if !cfg.Enabled {
		return nil
	}

	// Convert to hypertable (if not already)
	// Note: TimescaleDB hypertables are incompatible with PostgreSQL declarative partitioning
	// We need to check if the table is already a hypertable first
	var isHypertable bool
	err := s.db.Pool().QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM timescaledb_information.hypertables
			WHERE hypertable_schema = 'logging' AND hypertable_name = 'entries'
		)
	`).Scan(&isHypertable)
	if err != nil {
		return fmt.Errorf("failed to check if table is already a hypertable: %w", err)
	}

	if !isHypertable {
		// Drop existing partitions first (TimescaleDB is incompatible with native partitioning)
		// This is a one-time migration operation
		_, err = s.db.Pool().Exec(ctx, `
			-- Drop partitions (data is preserved in the parent table)
			DROP TABLE IF EXISTS logging.entries_system CASCADE;
			DROP TABLE IF EXISTS logging.entries_http CASCADE;
			DROP TABLE IF EXISTS logging.entries_security CASCADE;
			DROP TABLE IF EXISTS logging.entries_execution CASCADE;
			DROP TABLE IF EXISTS logging.entries_ai CASCADE;
			DROP TABLE IF EXISTS logging.entries_custom CASCADE;

			-- Remove partitioning from parent table
			ALTER TABLE logging.entries DROP CONSTRAINT logging_entries_pkey;
			ALTER TABLE logging.entries DROP CONSTRAINT IF EXISTS valid_category;

			-- Add primary key without partitioning
			ALTER TABLE logging.entries ADD PRIMARY KEY (id);

			-- Convert to hypertable
			SELECT create_hypertable('logging.entries', 'timestamp',
				if_not_exists => TRUE,
				migrate_data => TRUE
			);
		`)
		if err != nil {
			return fmt.Errorf("failed to convert table to hypertable: %w", err)
		}
	}

	// Enable compression if configured
	if cfg.Compressed {
		// Set compression on the hypertable
		_, err = s.db.Pool().Exec(ctx, `
			ALTER TABLE logging.entries SET (
				timescaledb.compress = TRUE,
				timescaledb.compress_segmentby = 'category, level',
				timescaledb.compress_orderby = 'timestamp DESC'
			);
		`)
		if err != nil {
			return fmt.Errorf("failed to enable compression: %w", err)
		}

		// Add compression policy if compress_after is specified
		if cfg.CompressAfter > 0 {
			// First remove existing compression policy if any
			_, err = s.db.Pool().Exec(ctx, `
				SELECT remove_compression_policy('logging.entries', if_exists => TRUE);
			`)
			if err != nil {
				return fmt.Errorf("failed to remove existing compression policy: %w", err)
			}

			// Add new compression policy
			compressInterval := fmt.Sprintf("INTERVAL '%d seconds'", int64(cfg.CompressAfter.Seconds()))
			_, err = s.db.Pool().Exec(ctx, `
				SELECT add_compression_policy('logging.entries', $1::interval)
			`, compressInterval)
			if err != nil {
				return fmt.Errorf("failed to add compression policy: %w", err)
			}
		}
	}

	return nil
}
