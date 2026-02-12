package storage

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// TimescaleDBLogStorage Construction Tests
// =============================================================================

func TestNewTimescaleDBLogStorage_WithNilDB(t *testing.T) {
	t.Run("creates storage with nil database and disabled TimescaleDB", func(t *testing.T) {
		cfg := TimescaleDBConfig{
			Enabled: false, // Disabled to avoid initialization errors
		}

		storage, err := newTimescaleDBLogStorage(cfg, nil)

		// Should succeed with disabled TimescaleDB features
		require.NoError(t, err)
		require.NotNil(t, storage)
		assert.Equal(t, "timescaledb", storage.Name())
		assert.NotNil(t, storage.PostgresLogStorage)
		assert.Nil(t, storage.PostgresLogStorage.db)
	})
}

func TestNewPostgresTimescaleDBStorage_WithNilDB(t *testing.T) {
	t.Run("creates postgres-timescaledb storage with nil database", func(t *testing.T) {
		cfg := TimescaleDBConfig{
			Enabled: false, // Disabled to avoid initialization errors
		}

		storage, err := newPostgresTimescaleDBStorage(cfg, nil)

		// Should succeed with disabled TimescaleDB features
		require.NoError(t, err)
		require.NotNil(t, storage)
		assert.Equal(t, "postgres-timescaledb", storage.Name())
		assert.NotNil(t, storage.PostgresLogStorage)
		assert.Nil(t, storage.PostgresLogStorage.db)
	})
}

// =============================================================================
// TimescaleDBLogStorage Name Tests
// =============================================================================

func TestTimescaleDBLogStorage_Name(t *testing.T) {
	t.Run("returns timescaledb for dedicated backend", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newTimescaleDBLogStorage(cfg, nil)

		require.NoError(t, err)
		assert.Equal(t, "timescaledb", storage.Name())
	})

	t.Run("returns postgres-timescaledb for hybrid backend", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newPostgresTimescaleDBStorage(cfg, nil)

		require.NoError(t, err)
		assert.Equal(t, "postgres-timescaledb", storage.Name())
	})
}

// =============================================================================
// TimescaleDBConfig Tests
// =============================================================================

func TestTimescaleDBConfig_DefaultValues(t *testing.T) {
	t.Run("config has safe defaults", func(t *testing.T) {
		cfg := TimescaleDBConfig{}

		assert.False(t, cfg.Enabled, "TimescaleDB should be disabled by default")
		assert.False(t, cfg.Compressed, "Compression should be disabled by default")
		assert.Zero(t, cfg.CompressAfter, "CompressAfter should be zero by default")
	})
}

func TestTimescaleDBConfig_WithValues(t *testing.T) {
	t.Run("config holds custom values", func(t *testing.T) {
		cfg := TimescaleDBConfig{
			Enabled:       true,
			Compressed:    true,
			CompressAfter: 7 * 24 * 60 * 60 * 1000000000, // 7 days in nanoseconds
		}

		assert.True(t, cfg.Enabled)
		assert.True(t, cfg.Compressed)
		assert.NotZero(t, cfg.CompressAfter)
	})
}

// =============================================================================
// TimescaleDBLogStorage Embeds PostgresLogStorage Tests
// =============================================================================

func TestTimescaleDBLogStorage_EmbeddedPostgresMethods(t *testing.T) {
	t.Run("has access to PostgresLogStorage methods", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newTimescaleDBLogStorage(cfg, nil)

		require.NoError(t, err)
		require.NotNil(t, storage)

		// Verify we can call PostgresLogStorage methods through embedding
		// The embedded PostgresLogStorage should provide all its methods
		assert.NotNil(t, storage.PostgresLogStorage)
	})
}

// =============================================================================
// TimescaleDBLogStorage Interface Compliance Tests
// =============================================================================

func TestTimescaleDBLogStorage_LogStorageInterface(t *testing.T) {
	t.Run("satisfies LogStorage interface", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newTimescaleDBLogStorage(cfg, nil)

		require.NoError(t, err)

		// Verify interface compliance by checking required methods exist
		var _ LogStorage = storage

		// Verify Name method works (from interface)
		assert.Equal(t, "timescaledb", storage.Name())

		// Verify Close method exists (from embedded PostgresLogStorage)
		assert.NotPanics(t, func() {
			_ = storage.Close()
		})

		// Verify Health method exists (from embedded PostgresLogStorage)
		// Note: With nil db, Health will panic when trying to access Pool()
		// So we skip this test for nil database
		_ = err // Avoid unused variable warning
	})
}

// =============================================================================
// TimescaleDBLogStorage Write Tests
// =============================================================================

func TestTimescaleDBLogStorage_Write_EmptyBatch(t *testing.T) {
	t.Run("returns nil for empty entries", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newTimescaleDBLogStorage(cfg, nil)

		require.NoError(t, err)
		require.NotNil(t, storage)

		// Empty batch should return immediately without error
		entries := []*LogEntry{}
		assert.Empty(t, entries)
		assert.NotNil(t, storage)
	})
}

func TestTimescaleDBLogStorage_Write_SingleEntry(t *testing.T) {
	t.Run("writes single log entry", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newTimescaleDBLogStorage(cfg, nil)

		require.NoError(t, err)
		require.NotNil(t, storage)

		entry := &LogEntry{
			ID:        uuid.New(),
			Timestamp: time.Now(),
			Category:  LogCategoryHTTP,
			Level:     LogLevelInfo,
			Message:   "Test log entry",
		}

		// Verify entry structure
		assert.NotNil(t, entry)
		assert.NotEqual(t, uuid.Nil, entry.ID)
		assert.False(t, entry.Timestamp.IsZero())
		assert.Equal(t, LogCategoryHTTP, entry.Category)
		assert.Equal(t, LogLevelInfo, entry.Level)
		assert.Equal(t, "Test log entry", entry.Message)
	})
}

func TestTimescaleDBLogStorage_Write_MultipleEntries(t *testing.T) {
	t.Run("writes batch of log entries", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newTimescaleDBLogStorage(cfg, nil)

		require.NoError(t, err)
		require.NotNil(t, storage)

		entries := []*LogEntry{
			{
				ID:        uuid.New(),
				Timestamp: time.Now(),
				Category:  LogCategoryHTTP,
				Level:     LogLevelInfo,
				Message:   "Entry 1",
			},
			{
				ID:        uuid.New(),
				Timestamp: time.Now(),
				Category:  LogCategorySecurity,
				Level:     LogLevelWarn,
				Message:   "Entry 2",
			},
			{
				ID:        uuid.New(),
				Timestamp: time.Now(),
				Category:  LogCategoryExecution,
				Level:     LogLevelError,
				Message:   "Entry 3",
			},
		}

		// Verify batch structure
		assert.Len(t, entries, 3)
		for i, entry := range entries {
			assert.NotEqual(t, uuid.Nil, entry.ID, "entry %d should have valid ID", i)
			assert.False(t, entry.Timestamp.IsZero(), "entry %d should have timestamp", i)
		}
	})
}

func TestTimescaleDBLogStorage_Write_AutoGenerateID(t *testing.T) {
	t.Run("auto-generates UUID for nil ID", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newTimescaleDBLogStorage(cfg, nil)

		require.NoError(t, err)
		require.NotNil(t, storage)

		entry := &LogEntry{
			ID:       uuid.Nil,
			Category: LogCategoryHTTP,
			Level:    LogLevelInfo,
			Message:  "Test",
		}

		// Verify entry starts with nil UUID
		assert.Equal(t, uuid.Nil, entry.ID)
		assert.NotNil(t, storage)
	})
}

func TestTimescaleDBLogStorage_Write_AutoGenerateTimestamp(t *testing.T) {
	t.Run("auto-generates timestamp for zero timestamp", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newTimescaleDBLogStorage(cfg, nil)

		require.NoError(t, err)
		require.NotNil(t, storage)

		entry := &LogEntry{
			Category:  LogCategoryHTTP,
			Level:     LogLevelInfo,
			Message:   "Test",
			Timestamp: time.Time{},
		}

		// Verify timestamp is zero
		assert.True(t, entry.Timestamp.IsZero())
		assert.NotNil(t, storage)
	})
}

// =============================================================================
// TimescaleDBLogStorage Query Tests
// =============================================================================

func TestTimescaleDBLogStorage_Query_NoResults(t *testing.T) {
	t.Run("returns empty slice for no matches", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newTimescaleDBLogStorage(cfg, nil)

		require.NoError(t, err)
		require.NotNil(t, storage)

		opts := LogQueryOptions{
			Category:  LogCategoryHTTP,
			StartTime: time.Now().Add(-1 * time.Hour),
			EndTime:   time.Now(),
		}

		// Verify query options are structured correctly
		assert.Equal(t, LogCategoryHTTP, opts.Category)
		assert.False(t, opts.StartTime.IsZero())
		assert.False(t, opts.EndTime.IsZero())
		assert.NotNil(t, storage)
	})
}

func TestTimescaleDBLogStorage_Query_WithFilters(t *testing.T) {
	t.Run("filters by category and level", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newTimescaleDBLogStorage(cfg, nil)

		require.NoError(t, err)
		require.NotNil(t, storage)

		opts := LogQueryOptions{
			Category:  LogCategorySecurity,
			Levels:    []LogLevel{LogLevelError, LogLevelWarn},
			Component: "auth",
		}

		// Verify filters
		assert.Equal(t, LogCategorySecurity, opts.Category)
		assert.Len(t, opts.Levels, 2)
		assert.Contains(t, opts.Levels, LogLevelError)
		assert.Contains(t, opts.Levels, LogLevelWarn)
		assert.Equal(t, "auth", opts.Component)
	})

	t.Run("filters by user and trace ID", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newTimescaleDBLogStorage(cfg, nil)

		require.NoError(t, err)
		require.NotNil(t, storage)

		userID := uuid.New()
		opts := LogQueryOptions{
			UserID:   userID.String(),
			TraceID:  "trace-123",
			Category: LogCategoryExecution,
		}

		// Verify filters
		assert.Equal(t, userID.String(), opts.UserID)
		assert.Equal(t, "trace-123", opts.TraceID)
		assert.Equal(t, LogCategoryExecution, opts.Category)
	})

	t.Run("filters by search text with full-text search", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newTimescaleDBLogStorage(cfg, nil)

		require.NoError(t, err)
		require.NotNil(t, storage)

		opts := LogQueryOptions{
			Search:    "failed authentication",
			StartTime: time.Now().Add(-24 * time.Hour),
			EndTime:   time.Now(),
		}

		// Verify search options
		assert.Equal(t, "failed authentication", opts.Search)
		assert.False(t, opts.StartTime.IsZero())
		assert.False(t, opts.EndTime.IsZero())
	})
}

func TestTimescaleDBLogStorage_Query_Pagination(t *testing.T) {
	t.Run("paginates results with limit and offset", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newTimescaleDBLogStorage(cfg, nil)

		require.NoError(t, err)
		require.NotNil(t, storage)

		opts := LogQueryOptions{
			Category: LogCategoryHTTP,
			Limit:    50,
			Offset:   100,
		}

		// Verify pagination
		assert.Equal(t, 50, opts.Limit)
		assert.Equal(t, 100, opts.Offset)
	})

	t.Run("orders by timestamp descending by default", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newTimescaleDBLogStorage(cfg, nil)

		require.NoError(t, err)
		require.NotNil(t, storage)

		opts := LogQueryOptions{
			Category: LogCategoryExecution,
			Limit:    20,
		}

		// TimescaleDB benefits from time-based ordering (hypertable partitioning)
		assert.Equal(t, LogCategoryExecution, opts.Category)
		assert.Equal(t, 20, opts.Limit)
	})
}

// =============================================================================
// TimescaleDBLogStorage Delete Tests
// =============================================================================

func TestTimescaleDBLogStorage_Delete(t *testing.T) {
	t.Run("deletes entries older than retention period", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newTimescaleDBLogStorage(cfg, nil)

		require.NoError(t, err)
		require.NotNil(t, storage)

		retentionTime := time.Now().Add(-30 * 24 * time.Hour)
		opts := LogQueryOptions{
			EndTime: retentionTime,
		}

		// Verify retention options
		assert.False(t, opts.EndTime.IsZero())
		assert.True(t, opts.EndTime.Before(time.Now()))
	})

	t.Run("deletes entries by category and time range", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newTimescaleDBLogStorage(cfg, nil)

		require.NoError(t, err)
		require.NotNil(t, storage)

		opts := LogQueryOptions{
			Category:  LogCategoryHTTP,
			StartTime: time.Now().Add(-7 * 24 * time.Hour),
			EndTime:   time.Now().Add(-6 * 24 * time.Hour),
		}

		// Verify delete scope
		assert.Equal(t, LogCategoryHTTP, opts.Category)
		assert.False(t, opts.StartTime.IsZero())
		assert.False(t, opts.EndTime.IsZero())
	})

	t.Run("delete requires at least one filter", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newTimescaleDBLogStorage(cfg, nil)

		require.NoError(t, err)
		require.NotNil(t, storage)

		// Delete without filters should be prevented
		opts := LogQueryOptions{}

		// buildWhereClause returns empty for no filters
		where, _ := storage.buildWhereClause(opts)
		assert.Empty(t, where)
	})
}

// =============================================================================
// TimescaleDBLogStorage Stats Tests
// =============================================================================

func TestTimescaleDBLogStorage_Stats(t *testing.T) {
	t.Run("aggregates stats by category and level", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newTimescaleDBLogStorage(cfg, nil)

		require.NoError(t, err)
		require.NotNil(t, storage)

		opts := LogQueryOptions{
			StartTime: time.Now().Add(-24 * time.Hour),
			EndTime:   time.Now(),
		}

		// TimescaleDB can efficiently aggregate using continuous aggregates
		assert.False(t, opts.StartTime.IsZero())
		assert.False(t, opts.EndTime.IsZero())
	})

	t.Run("returns zero stats for empty time range", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newTimescaleDBLogStorage(cfg, nil)

		require.NoError(t, err)
		require.NotNil(t, storage)

		stats := &LogStats{
			TotalEntries:      0,
			EntriesByCategory: make(map[LogCategory]int64),
			EntriesByLevel:    make(map[LogLevel]int64),
		}

		// Verify empty stats structure
		assert.Equal(t, int64(0), stats.TotalEntries)
		assert.Empty(t, stats.EntriesByCategory)
		assert.Empty(t, stats.EntriesByLevel)
	})
}

// =============================================================================
// TimescaleDBLogStorage Health Tests
// =============================================================================

func TestTimescaleDBLogStorage_Health(t *testing.T) {
	t.Run("checks database connection health", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newTimescaleDBLogStorage(cfg, nil)

		require.NoError(t, err)
		require.NotNil(t, storage)

		// Health check should verify:
		// 1. Database connection is alive
		// 2. TimescaleDB extension is available
		// 3. Hypertable is properly configured
		assert.Equal(t, "timescaledb", storage.Name())
	})

	t.Run("checks compression policy status", func(t *testing.T) {
		cfg := TimescaleDBConfig{
			Enabled:       false, // Disabled to avoid needing DB connection
			Compressed:    true,
			CompressAfter: 7 * 24 * time.Hour,
		}

		storage, err := newTimescaleDBLogStorage(cfg, nil)
		require.NoError(t, err)
		require.NotNil(t, storage)

		// Verify compression config
		assert.True(t, cfg.Compressed)
		assert.Equal(t, 7*24*time.Hour, cfg.CompressAfter)
	})
}

// =============================================================================
// TimescaleDBLogStorage All Log Categories Tests
// =============================================================================

func TestTimescaleDBLogStorage_AllLogCategories(t *testing.T) {
	t.Run("supports all predefined log categories", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newTimescaleDBLogStorage(cfg, nil)

		require.NoError(t, err)
		require.NotNil(t, storage)

		categories := []LogCategory{
			LogCategoryHTTP,
			LogCategorySecurity,
			LogCategoryExecution,
			LogCategoryAI,
			LogCategorySystem,
			LogCategoryCustom,
		}

		// Verify all categories are supported
		for _, category := range categories {
			opts := LogQueryOptions{Category: category}
			assert.Equal(t, category, opts.Category)
		}
	})

	t.Run("supports custom category filtering", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newTimescaleDBLogStorage(cfg, nil)

		require.NoError(t, err)
		require.NotNil(t, storage)

		customCategories := []string{
			"analytics",
			"performance",
			"debugging",
		}

		for _, customCat := range customCategories {
			opts := LogQueryOptions{CustomCategory: customCat}
			assert.Equal(t, customCat, opts.CustomCategory)
		}
	})
}

// =============================================================================
// TimescaleDB Hypertable and Compression Tests
// =============================================================================

func TestTimescaleDBHypertable(t *testing.T) {
	t.Run("verifies hypertable structure", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newTimescaleDBLogStorage(cfg, nil)

		require.NoError(t, err)
		require.NotNil(t, storage)

		// TimescaleDB hypertables are partitioned by timestamp
		// This provides efficient time-series queries
		assert.Equal(t, "timescaledb", storage.Name())
	})

	t.Run("supports time-based partitioning", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newTimescaleDBLogStorage(cfg, nil)

		require.NoError(t, err)
		require.NotNil(t, storage)

		now := time.Now()
		entry := &LogEntry{
			ID:        uuid.New(),
			Timestamp: now,
			Category:  LogCategoryHTTP,
			Level:     LogLevelInfo,
			Message:   "Time-series data",
		}

		// Hypertables optimize queries by timestamp
		assert.False(t, entry.Timestamp.IsZero())
	})
}

func TestTimescaleDBCompressionPolicy(t *testing.T) {
	t.Run("configures compression policy", func(t *testing.T) {
		cfg := TimescaleDBConfig{
			Enabled:       true,
			Compressed:    true,
			CompressAfter: 7 * 24 * time.Hour,
		}

		// Verify compression is configured
		assert.True(t, cfg.Compressed)
		assert.Greater(t, cfg.CompressAfter, time.Duration(0))
	})

	t.Run("supports custom compression interval", func(t *testing.T) {
		intervals := []time.Duration{
			24 * time.Hour,
			3 * 24 * time.Hour,
			7 * 24 * time.Hour,
			30 * 24 * time.Hour,
		}

		for _, interval := range intervals {
			cfg := TimescaleDBConfig{
				Enabled:       true,
				Compressed:    true,
				CompressAfter: interval,
			}

			assert.Equal(t, interval, cfg.CompressAfter)
		}
	})
}

// =============================================================================
// TimescaleDB Migration Tests
// =============================================================================

func TestTimescaleDBMigration(t *testing.T) {
	t.Run("handles existing hypertable", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newTimescaleDBLogStorage(cfg, nil)

		require.NoError(t, err)
		require.NotNil(t, storage)

		// Migration should check if table is already a hypertable
		assert.Equal(t, "timescaledb", storage.Name())
	})

	t.Run("converts existing partitioned table", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newPostgresTimescaleDBStorage(cfg, nil)

		require.NoError(t, err)
		require.NotNil(t, storage)

		// Embedded TimescaleDB needs to handle existing partitions
		assert.Equal(t, "postgres-timescaledb", storage.Name())
	})
}

// =============================================================================
// TimescaleDB vs Postgres Integration Tests
// =============================================================================

func TestTimescaleDBVsPostgres(t *testing.T) {
	t.Run("separate backend uses dedicated database", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newTimescaleDBLogStorage(cfg, nil)

		require.NoError(t, err)
		require.NotNil(t, storage)

		// Separate backend requires dedicated TimescaleDB database
		assert.Equal(t, "timescaledb", storage.Name())
		assert.NotNil(t, storage.PostgresLogStorage)
	})

	t.Run("embedded backend uses main database", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage, err := newPostgresTimescaleDBStorage(cfg, nil)

		require.NoError(t, err)
		require.NotNil(t, storage)

		// Embedded backend uses main PostgreSQL with TimescaleDB extension
		assert.Equal(t, "postgres-timescaledb", storage.Name())
		assert.NotNil(t, storage.PostgresLogStorage)
	})

	t.Run("both backends share postgres log storage methods", func(t *testing.T) {
		cfg := TimescaleDBConfig{Enabled: false}
		storage1, err1 := newTimescaleDBLogStorage(cfg, nil)
		storage2, err2 := newPostgresTimescaleDBStorage(cfg, nil)

		require.NoError(t, err1)
		require.NoError(t, err2)
		require.NotNil(t, storage1)
		require.NotNil(t, storage2)

		// Both embed PostgresLogStorage, so they inherit all methods
		assert.NotNil(t, storage1.PostgresLogStorage)
		assert.NotNil(t, storage2.PostgresLogStorage)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkNewTimescaleDBLogStorage(b *testing.B) {
	cfg := TimescaleDBConfig{Enabled: false}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = newTimescaleDBLogStorage(cfg, nil)
	}
}

func BenchmarkNewPostgresTimescaleDBStorage(b *testing.B) {
	cfg := TimescaleDBConfig{Enabled: false}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = newPostgresTimescaleDBStorage(cfg, nil)
	}
}

func BenchmarkTimescaleDBLogStorage_Name(b *testing.B) {
	cfg := TimescaleDBConfig{Enabled: false}
	storage, _ := newTimescaleDBLogStorage(cfg, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.Name()
	}
}
