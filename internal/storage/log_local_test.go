package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// NewLocalLogStorage Tests
// =============================================================================

func TestNewLocalLogStorage(t *testing.T) {
	t.Run("creates storage with default path", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer func() { _ = os.Chdir(originalDir) }()
		_ = os.Chdir(tmpDir)

		storage, err := NewLocalLogStorage("")

		require.NoError(t, err)
		require.NotNil(t, storage)
		assert.Equal(t, "./logs", storage.basePath)
	})

	t.Run("creates storage with custom path", func(t *testing.T) {
		tmpDir := t.TempDir()
		customPath := filepath.Join(tmpDir, "custom-logs")

		storage, err := NewLocalLogStorage(customPath)

		require.NoError(t, err)
		require.NotNil(t, storage)
		assert.Equal(t, customPath, storage.basePath)

		// Verify directory was created
		_, err = os.Stat(customPath)
		assert.NoError(t, err)
	})

	t.Run("returns error for invalid path", func(t *testing.T) {
		// Try to create in a location that should fail
		storage, err := NewLocalLogStorage("/proc/invalid/path/that/cannot/exist")

		assert.Error(t, err)
		assert.Nil(t, storage)
	})
}

// =============================================================================
// Name Tests
// =============================================================================

func TestLocalLogStorage_Name(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalLogStorage(tmpDir)
	require.NoError(t, err)

	assert.Equal(t, "local", storage.Name())
}

// =============================================================================
// Write Tests
// =============================================================================

func TestLocalLogStorage_Write(t *testing.T) {
	t.Run("writes empty entries without error", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		err = storage.Write(context.Background(), []*LogEntry{})

		assert.NoError(t, err)
	})

	t.Run("writes single entry", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		entry := &LogEntry{
			Timestamp: time.Now(),
			Category:  LogCategoryHTTP,
			Level:     LogLevelInfo,
			Message:   "Test log message",
		}

		err = storage.Write(context.Background(), []*LogEntry{entry})

		assert.NoError(t, err)
		// Verify ID was assigned
		assert.NotEqual(t, uuid.Nil, entry.ID)
	})

	t.Run("writes multiple entries", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		entries := []*LogEntry{
			{Category: LogCategoryHTTP, Level: LogLevelInfo, Message: "Entry 1"},
			{Category: LogCategoryHTTP, Level: LogLevelWarn, Message: "Entry 2"},
			{Category: LogCategoryAuth, Level: LogLevelError, Message: "Entry 3"},
		}

		err = storage.Write(context.Background(), entries)

		assert.NoError(t, err)
	})

	t.Run("groups execution logs by execution ID", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		execID := "exec-123"
		entries := []*LogEntry{
			{
				Timestamp:   time.Now(),
				Category:    LogCategoryExecution,
				Level:       LogLevelInfo,
				Message:     "Starting execution",
				ExecutionID: execID,
				LineNumber:  1,
			},
			{
				Timestamp:   time.Now(),
				Category:    LogCategoryExecution,
				Level:       LogLevelInfo,
				Message:     "Execution complete",
				ExecutionID: execID,
				LineNumber:  2,
			},
		}

		err = storage.Write(context.Background(), entries)

		assert.NoError(t, err)
	})

	t.Run("assigns timestamp if not set", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		entry := &LogEntry{
			Category: LogCategoryHTTP,
			Level:    LogLevelInfo,
			Message:  "No timestamp",
		}

		err = storage.Write(context.Background(), []*LogEntry{entry})

		assert.NoError(t, err)
		assert.False(t, entry.Timestamp.IsZero())
	})
}

// =============================================================================
// Query Tests
// =============================================================================

func TestLocalLogStorage_Query(t *testing.T) {
	t.Run("returns empty result for empty storage", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		result, err := storage.Query(context.Background(), LogQueryOptions{})

		require.NoError(t, err)
		assert.Empty(t, result.Entries)
		assert.Equal(t, int64(0), result.TotalCount)
	})

	t.Run("queries written entries", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		// Write entries
		entries := []*LogEntry{
			{Category: LogCategoryHTTP, Level: LogLevelInfo, Message: "Test 1"},
			{Category: LogCategoryHTTP, Level: LogLevelWarn, Message: "Test 2"},
		}
		err = storage.Write(context.Background(), entries)
		require.NoError(t, err)

		// Query
		result, err := storage.Query(context.Background(), LogQueryOptions{})

		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result.Entries), 2)
	})

	t.Run("filters by category", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		// Write entries with different categories
		entries := []*LogEntry{
			{Category: LogCategoryHTTP, Level: LogLevelInfo, Message: "HTTP entry"},
			{Category: LogCategoryAuth, Level: LogLevelInfo, Message: "Auth entry"},
		}
		err = storage.Write(context.Background(), entries)
		require.NoError(t, err)

		// Query only HTTP
		result, err := storage.Query(context.Background(), LogQueryOptions{
			Category: LogCategoryHTTP,
		})

		require.NoError(t, err)
		for _, entry := range result.Entries {
			assert.Equal(t, LogCategoryHTTP, entry.Category)
		}
	})

	t.Run("applies pagination", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		// Write multiple entries
		var entries []*LogEntry
		for i := 0; i < 10; i++ {
			entries = append(entries, &LogEntry{
				Category: LogCategoryHTTP,
				Level:    LogLevelInfo,
				Message:  "Test entry",
			})
		}
		err = storage.Write(context.Background(), entries)
		require.NoError(t, err)

		// Query with limit
		result, err := storage.Query(context.Background(), LogQueryOptions{
			Limit: 5,
		})

		require.NoError(t, err)
		assert.LessOrEqual(t, len(result.Entries), 5)
	})
}

// =============================================================================
// matchesFilter Tests
// =============================================================================

func TestLocalLogStorage_matchesFilter(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalLogStorage(tmpDir)
	require.NoError(t, err)

	baseTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	t.Run("matches when no filters", func(t *testing.T) {
		entry := &LogEntry{
			Category:  LogCategoryHTTP,
			Level:     LogLevelInfo,
			Timestamp: baseTime,
		}

		result := storage.matchesFilter(entry, LogQueryOptions{})

		assert.True(t, result)
	})

	t.Run("filters by category", func(t *testing.T) {
		entry := &LogEntry{Category: LogCategoryHTTP}

		assert.True(t, storage.matchesFilter(entry, LogQueryOptions{Category: LogCategoryHTTP}))
		assert.False(t, storage.matchesFilter(entry, LogQueryOptions{Category: LogCategoryAuth}))
	})

	t.Run("filters by levels", func(t *testing.T) {
		entry := &LogEntry{Level: LogLevelWarn}

		assert.True(t, storage.matchesFilter(entry, LogQueryOptions{Levels: []LogLevel{LogLevelWarn, LogLevelError}}))
		assert.False(t, storage.matchesFilter(entry, LogQueryOptions{Levels: []LogLevel{LogLevelInfo}}))
	})

	t.Run("filters by component", func(t *testing.T) {
		entry := &LogEntry{Component: "auth"}

		assert.True(t, storage.matchesFilter(entry, LogQueryOptions{Component: "auth"}))
		assert.False(t, storage.matchesFilter(entry, LogQueryOptions{Component: "api"}))
	})

	t.Run("filters by request ID", func(t *testing.T) {
		entry := &LogEntry{RequestID: "req-123"}

		assert.True(t, storage.matchesFilter(entry, LogQueryOptions{RequestID: "req-123"}))
		assert.False(t, storage.matchesFilter(entry, LogQueryOptions{RequestID: "req-456"}))
	})

	t.Run("filters by user ID", func(t *testing.T) {
		entry := &LogEntry{UserID: "user-123"}

		assert.True(t, storage.matchesFilter(entry, LogQueryOptions{UserID: "user-123"}))
		assert.False(t, storage.matchesFilter(entry, LogQueryOptions{UserID: "user-456"}))
	})

	t.Run("filters by execution ID", func(t *testing.T) {
		entry := &LogEntry{ExecutionID: "exec-123"}

		assert.True(t, storage.matchesFilter(entry, LogQueryOptions{ExecutionID: "exec-123"}))
		assert.False(t, storage.matchesFilter(entry, LogQueryOptions{ExecutionID: "exec-456"}))
	})

	t.Run("filters by time range", func(t *testing.T) {
		entry := &LogEntry{Timestamp: baseTime}

		// Within range
		assert.True(t, storage.matchesFilter(entry, LogQueryOptions{
			StartTime: baseTime.Add(-time.Hour),
			EndTime:   baseTime.Add(time.Hour),
		}))

		// Before start time
		assert.False(t, storage.matchesFilter(entry, LogQueryOptions{
			StartTime: baseTime.Add(time.Hour),
		}))

		// After end time
		assert.False(t, storage.matchesFilter(entry, LogQueryOptions{
			EndTime: baseTime.Add(-time.Hour),
		}))
	})

	t.Run("filters by search term", func(t *testing.T) {
		entry := &LogEntry{Message: "User login successful"}

		assert.True(t, storage.matchesFilter(entry, LogQueryOptions{Search: "login"}))
		assert.True(t, storage.matchesFilter(entry, LogQueryOptions{Search: "LOGIN"})) // Case insensitive
		assert.False(t, storage.matchesFilter(entry, LogQueryOptions{Search: "logout"}))
	})
}

// =============================================================================
// GetExecutionLogs Tests
// =============================================================================

func TestLocalLogStorage_GetExecutionLogs(t *testing.T) {
	t.Run("returns empty for non-existent execution", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		entries, err := storage.GetExecutionLogs(context.Background(), "non-existent", 0)

		require.NoError(t, err)
		assert.Empty(t, entries)
	})

	t.Run("returns entries for existing execution", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		execID := uuid.New().String()
		entries := []*LogEntry{
			{
				Category:    LogCategoryExecution,
				Level:       LogLevelInfo,
				Message:     "Line 1",
				ExecutionID: execID,
				LineNumber:  1,
			},
			{
				Category:    LogCategoryExecution,
				Level:       LogLevelInfo,
				Message:     "Line 2",
				ExecutionID: execID,
				LineNumber:  2,
			},
		}
		err = storage.Write(context.Background(), entries)
		require.NoError(t, err)

		result, err := storage.GetExecutionLogs(context.Background(), execID, 0)

		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("filters by afterLine", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		execID := uuid.New().String()
		entries := []*LogEntry{
			{Category: LogCategoryExecution, ExecutionID: execID, LineNumber: 1, Message: "Line 1"},
			{Category: LogCategoryExecution, ExecutionID: execID, LineNumber: 2, Message: "Line 2"},
			{Category: LogCategoryExecution, ExecutionID: execID, LineNumber: 3, Message: "Line 3"},
		}
		err = storage.Write(context.Background(), entries)
		require.NoError(t, err)

		result, err := storage.GetExecutionLogs(context.Background(), execID, 1)

		require.NoError(t, err)
		for _, entry := range result {
			assert.Greater(t, entry.LineNumber, 1)
		}
	})
}

// =============================================================================
// Delete Tests
// =============================================================================

func TestLocalLogStorage_Delete(t *testing.T) {
	t.Run("deletes entries matching filter", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		// Write entries
		entries := []*LogEntry{
			{Category: LogCategoryHTTP, Level: LogLevelInfo, Message: "To delete"},
		}
		err = storage.Write(context.Background(), entries)
		require.NoError(t, err)

		// Delete
		count, err := storage.Delete(context.Background(), LogQueryOptions{
			Category: LogCategoryHTTP,
		})

		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(1))
	})

	t.Run("returns zero for non-matching filter", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		count, err := storage.Delete(context.Background(), LogQueryOptions{
			Category: "nonexistent",
		})

		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

// =============================================================================
// Stats Tests
// =============================================================================

func TestLocalLogStorage_Stats(t *testing.T) {
	t.Run("returns empty stats for empty storage", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		stats, err := storage.Stats(context.Background())

		require.NoError(t, err)
		assert.Equal(t, int64(0), stats.TotalEntries)
	})

	t.Run("counts entries by category", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		entries := []*LogEntry{
			{Category: LogCategoryHTTP, Level: LogLevelInfo, Message: "HTTP 1"},
			{Category: LogCategoryHTTP, Level: LogLevelInfo, Message: "HTTP 2"},
			{Category: LogCategoryAuth, Level: LogLevelInfo, Message: "Auth 1"},
		}
		err = storage.Write(context.Background(), entries)
		require.NoError(t, err)

		stats, err := storage.Stats(context.Background())

		require.NoError(t, err)
		assert.GreaterOrEqual(t, stats.TotalEntries, int64(2))
	})
}

// =============================================================================
// Health Tests
// =============================================================================

func TestLocalLogStorage_Health(t *testing.T) {
	t.Run("returns nil for healthy storage", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		err = storage.Health(context.Background())

		assert.NoError(t, err)
	})

	t.Run("creates directory if it doesn't exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		newPath := filepath.Join(tmpDir, "new-logs")
		storage := &LocalLogStorage{basePath: newPath}

		err := storage.Health(context.Background())

		assert.NoError(t, err)
		// Verify directory was created
		_, err = os.Stat(newPath)
		assert.NoError(t, err)
	})
}

// =============================================================================
// Close Tests
// =============================================================================

func TestLocalLogStorage_Close(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalLogStorage(tmpDir)
	require.NoError(t, err)

	err = storage.Close()

	assert.NoError(t, err)
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkLocalLogStorage_Write(b *testing.B) {
	tmpDir := b.TempDir()
	storage, _ := NewLocalLogStorage(tmpDir)

	entry := &LogEntry{
		Category: LogCategoryHTTP,
		Level:    LogLevelInfo,
		Message:  "Benchmark log entry",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.Write(context.Background(), []*LogEntry{entry})
	}
}

func BenchmarkLocalLogStorage_matchesFilter(b *testing.B) {
	tmpDir := b.TempDir()
	storage, _ := NewLocalLogStorage(tmpDir)

	entry := &LogEntry{
		Category:  LogCategoryHTTP,
		Level:     LogLevelInfo,
		Message:   "Test message",
		Component: "api",
		UserID:    "user-123",
		Timestamp: time.Now(),
	}

	opts := LogQueryOptions{
		Category:  LogCategoryHTTP,
		Levels:    []LogLevel{LogLevelInfo, LogLevelWarn},
		Component: "api",
		Search:    "test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.matchesFilter(entry, opts)
	}
}
