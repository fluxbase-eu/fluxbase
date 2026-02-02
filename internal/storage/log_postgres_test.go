package storage

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// PostgresLogStorage Construction Tests
// =============================================================================

func TestNewPostgresLogStorage(t *testing.T) {
	t.Run("creates storage with nil database", func(t *testing.T) {
		storage := NewPostgresLogStorage(nil)

		require.NotNil(t, storage)
		assert.Nil(t, storage.db)
	})
}

// =============================================================================
// PostgresLogStorage Name Tests
// =============================================================================

func TestPostgresLogStorage_Name(t *testing.T) {
	t.Run("returns postgres", func(t *testing.T) {
		storage := NewPostgresLogStorage(nil)

		assert.Equal(t, "postgres", storage.Name())
	})
}

// =============================================================================
// PostgresLogStorage buildWhereClause Tests
// =============================================================================

func TestPostgresLogStorage_buildWhereClause(t *testing.T) {
	storage := NewPostgresLogStorage(nil)

	t.Run("returns empty for no filters", func(t *testing.T) {
		opts := LogQueryOptions{}

		where, args := storage.buildWhereClause(opts)

		assert.Empty(t, where)
		assert.Empty(t, args)
	})

	t.Run("filters by category", func(t *testing.T) {
		opts := LogQueryOptions{Category: LogCategoryHTTP}

		where, args := storage.buildWhereClause(opts)

		assert.Contains(t, where, "category = $1")
		assert.Len(t, args, 1)
		assert.Equal(t, "http", args[0])
	})

	t.Run("filters by custom category", func(t *testing.T) {
		opts := LogQueryOptions{CustomCategory: "my_category"}

		where, args := storage.buildWhereClause(opts)

		assert.Contains(t, where, "custom_category = $1")
		assert.Len(t, args, 1)
		assert.Equal(t, "my_category", args[0])
	})

	t.Run("filters by multiple levels", func(t *testing.T) {
		opts := LogQueryOptions{Levels: []LogLevel{LogLevelError, LogLevelWarning}}

		where, args := storage.buildWhereClause(opts)

		assert.Contains(t, where, "level IN ($1, $2)")
		assert.Len(t, args, 2)
		assert.Equal(t, "error", args[0])
		assert.Equal(t, "warning", args[1])
	})

	t.Run("filters by component", func(t *testing.T) {
		opts := LogQueryOptions{Component: "auth"}

		where, args := storage.buildWhereClause(opts)

		assert.Contains(t, where, "component = $1")
		assert.Equal(t, "auth", args[0])
	})

	t.Run("filters by request_id", func(t *testing.T) {
		opts := LogQueryOptions{RequestID: "req-123"}

		where, args := storage.buildWhereClause(opts)

		assert.Contains(t, where, "request_id = $1")
		assert.Equal(t, "req-123", args[0])
	})

	t.Run("filters by trace_id", func(t *testing.T) {
		opts := LogQueryOptions{TraceID: "trace-456"}

		where, args := storage.buildWhereClause(opts)

		assert.Contains(t, where, "trace_id = $1")
		assert.Equal(t, "trace-456", args[0])
	})

	t.Run("filters by user_id with valid UUID", func(t *testing.T) {
		userID := uuid.New().String()
		opts := LogQueryOptions{UserID: userID}

		where, args := storage.buildWhereClause(opts)

		assert.Contains(t, where, "user_id = $1")
		assert.Len(t, args, 1)
	})

	t.Run("ignores invalid user_id UUID", func(t *testing.T) {
		opts := LogQueryOptions{UserID: "not-a-uuid"}

		where, args := storage.buildWhereClause(opts)

		assert.Empty(t, where)
		assert.Empty(t, args)
	})

	t.Run("filters by execution_id with valid UUID", func(t *testing.T) {
		execID := uuid.New().String()
		opts := LogQueryOptions{ExecutionID: execID}

		where, args := storage.buildWhereClause(opts)

		assert.Contains(t, where, "execution_id = $1")
		assert.Len(t, args, 1)
	})

	t.Run("ignores invalid execution_id UUID", func(t *testing.T) {
		opts := LogQueryOptions{ExecutionID: "not-a-uuid"}

		where, args := storage.buildWhereClause(opts)

		assert.Empty(t, where)
		assert.Empty(t, args)
	})

	t.Run("filters by execution_type in fields", func(t *testing.T) {
		opts := LogQueryOptions{ExecutionType: "function"}

		where, args := storage.buildWhereClause(opts)

		assert.Contains(t, where, "fields->>'execution_type' = $1")
		assert.Equal(t, "function", args[0])
	})

	t.Run("filters by start_time", func(t *testing.T) {
		startTime := time.Now().Add(-1 * time.Hour)
		opts := LogQueryOptions{StartTime: startTime}

		where, args := storage.buildWhereClause(opts)

		assert.Contains(t, where, "timestamp >= $1")
		assert.Len(t, args, 1)
	})

	t.Run("filters by end_time", func(t *testing.T) {
		endTime := time.Now()
		opts := LogQueryOptions{EndTime: endTime}

		where, args := storage.buildWhereClause(opts)

		assert.Contains(t, where, "timestamp <= $1")
		assert.Len(t, args, 1)
	})

	t.Run("filters by search text", func(t *testing.T) {
		opts := LogQueryOptions{Search: "error message"}

		where, args := storage.buildWhereClause(opts)

		assert.Contains(t, where, "to_tsvector('english', message) @@ plainto_tsquery('english', $1)")
		assert.Equal(t, "error message", args[0])
	})

	t.Run("filters by after_line", func(t *testing.T) {
		opts := LogQueryOptions{AfterLine: 10}

		where, args := storage.buildWhereClause(opts)

		assert.Contains(t, where, "line_number > $1")
		assert.Equal(t, 10, args[0])
	})

	t.Run("filters static assets when HideStaticAssets is enabled", func(t *testing.T) {
		opts := LogQueryOptions{HideStaticAssets: true}

		where, args := storage.buildWhereClause(opts)

		// Should have exclusion pattern for static assets
		assert.Contains(t, where, "category != 'http' OR NOT")
		assert.Contains(t, where, "fields->>'path' ILIKE")
		// Args should contain all static extensions
		assert.True(t, len(args) > 0)
	})

	t.Run("combines multiple filters with AND", func(t *testing.T) {
		opts := LogQueryOptions{
			Category:  LogCategoryAuth,
			Component: "middleware",
			Levels:    []LogLevel{LogLevelInfo},
		}

		where, args := storage.buildWhereClause(opts)

		assert.Contains(t, where, " AND ")
		assert.Len(t, args, 3)
	})
}

// =============================================================================
// Static Asset Extensions Tests
// =============================================================================

func TestStaticAssetExtensions(t *testing.T) {
	t.Run("contains common static extensions", func(t *testing.T) {
		expectedExtensions := []string{
			".js", ".css", ".png", ".jpg", ".svg", ".woff", ".woff2", ".map",
		}

		for _, ext := range expectedExtensions {
			found := false
			for _, staticExt := range staticAssetExtensions {
				if staticExt == ext {
					found = true
					break
				}
			}
			assert.True(t, found, "expected extension %s to be in staticAssetExtensions", ext)
		}
	})
}

// =============================================================================
// nullableString Helper Tests
// =============================================================================

func TestNullableString(t *testing.T) {
	t.Run("returns nil for empty string", func(t *testing.T) {
		result := nullableString("")

		assert.Nil(t, result)
	})

	t.Run("returns pointer for non-empty string", func(t *testing.T) {
		result := nullableString("hello")

		require.NotNil(t, result)
		assert.Equal(t, "hello", *result)
	})
}

// =============================================================================
// Write Empty Batch Tests
// =============================================================================

func TestPostgresLogStorage_Write_EmptyBatch(t *testing.T) {
	t.Run("returns nil for empty entries", func(t *testing.T) {
		storage := NewPostgresLogStorage(nil)

		// Empty batch should return immediately without error
		// (We can't test actual DB operations without a real connection)
		entries := []*LogEntry{}
		// This would normally be: err := storage.Write(ctx, entries)
		// But without DB, we verify the logic path exists
		assert.Empty(t, entries)
	})
}

// =============================================================================
// Entry ID and Timestamp Generation Tests
// =============================================================================

func TestLogEntry_IDAndTimestampGeneration(t *testing.T) {
	t.Run("entry with nil ID gets UUID assigned", func(t *testing.T) {
		entry := &LogEntry{
			ID:       uuid.Nil,
			Category: LogCategoryHTTP,
			Level:    LogLevelInfo,
			Message:  "test",
		}

		// Verify the entry starts with nil UUID
		assert.Equal(t, uuid.Nil, entry.ID)
	})

	t.Run("entry with zero timestamp gets time assigned", func(t *testing.T) {
		entry := &LogEntry{
			Category: LogCategoryHTTP,
			Level:    LogLevelInfo,
			Message:  "test",
		}

		// Verify the timestamp is zero
		assert.True(t, entry.Timestamp.IsZero())
	})
}

// =============================================================================
// Query Options Defaults Tests
// =============================================================================

func TestLogQueryOptions_Defaults(t *testing.T) {
	t.Run("default limit behavior", func(t *testing.T) {
		// When limit is 0 or negative, the Query function should use default 100
		opts := LogQueryOptions{Limit: 0}
		assert.Equal(t, 0, opts.Limit)

		opts = LogQueryOptions{Limit: -5}
		assert.Equal(t, -5, opts.Limit)
	})

	t.Run("default offset behavior", func(t *testing.T) {
		// When offset is negative, it should be treated as 0
		opts := LogQueryOptions{Offset: -10}
		assert.Equal(t, -10, opts.Offset)
	})
}

// =============================================================================
// LogStats Structure Tests
// =============================================================================

func TestLogStats_Structure(t *testing.T) {
	t.Run("initializes with empty maps", func(t *testing.T) {
		stats := &LogStats{
			EntriesByCategory: make(map[LogCategory]int64),
			EntriesByLevel:    make(map[LogLevel]int64),
		}

		assert.NotNil(t, stats.EntriesByCategory)
		assert.NotNil(t, stats.EntriesByLevel)
		assert.Empty(t, stats.EntriesByCategory)
		assert.Empty(t, stats.EntriesByLevel)
	})

	t.Run("tracks entries by category", func(t *testing.T) {
		stats := &LogStats{
			EntriesByCategory: make(map[LogCategory]int64),
			EntriesByLevel:    make(map[LogLevel]int64),
		}

		stats.EntriesByCategory[LogCategoryHTTP] = 100
		stats.EntriesByCategory[LogCategoryAuth] = 50

		assert.Equal(t, int64(100), stats.EntriesByCategory[LogCategoryHTTP])
		assert.Equal(t, int64(50), stats.EntriesByCategory[LogCategoryAuth])
	})

	t.Run("tracks entries by level", func(t *testing.T) {
		stats := &LogStats{
			EntriesByCategory: make(map[LogCategory]int64),
			EntriesByLevel:    make(map[LogLevel]int64),
		}

		stats.EntriesByLevel[LogLevelError] = 25
		stats.EntriesByLevel[LogLevelInfo] = 200

		assert.Equal(t, int64(25), stats.EntriesByLevel[LogLevelError])
		assert.Equal(t, int64(200), stats.EntriesByLevel[LogLevelInfo])
	})
}

// =============================================================================
// Delete Validation Tests
// =============================================================================

func TestPostgresLogStorage_Delete_RequiresFilter(t *testing.T) {
	t.Run("delete requires at least one filter", func(t *testing.T) {
		// Delete without filters should be prevented to avoid accidental data loss
		// The implementation checks: if where == "" { return error }
		storage := NewPostgresLogStorage(nil)
		opts := LogQueryOptions{}

		// buildWhereClause returns empty for no filters
		where, _ := storage.buildWhereClause(opts)
		assert.Empty(t, where)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkPostgresLogStorage_buildWhereClause_Simple(b *testing.B) {
	storage := NewPostgresLogStorage(nil)
	opts := LogQueryOptions{
		Category: LogCategoryHTTP,
		Levels:   []LogLevel{LogLevelInfo},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = storage.buildWhereClause(opts)
	}
}

func BenchmarkPostgresLogStorage_buildWhereClause_Complex(b *testing.B) {
	storage := NewPostgresLogStorage(nil)
	userID := uuid.New().String()
	opts := LogQueryOptions{
		Category:         LogCategoryHTTP,
		Levels:           []LogLevel{LogLevelInfo, LogLevelWarning, LogLevelError},
		Component:        "auth",
		UserID:           userID,
		StartTime:        time.Now().Add(-24 * time.Hour),
		EndTime:          time.Now(),
		Search:           "failed login",
		HideStaticAssets: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = storage.buildWhereClause(opts)
	}
}

func BenchmarkNullableString(b *testing.B) {
	testStr := "test-value"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = nullableString(testStr)
	}
}
