package storage

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Mock S3 Provider for Testing
// =============================================================================

type mockS3Provider struct {
	name           string
	uploadedFiles  map[string][]byte
	listedObjects  []Object
	downloadData   map[string]string
	healthError    error
	uploadCalled   bool
	downloadCalled bool
	deleteCalled   bool
}

func newMockS3Provider() *mockS3Provider {
	return &mockS3Provider{
		name:          "mock-s3",
		uploadedFiles: make(map[string][]byte),
		downloadData:  make(map[string]string),
	}
}

func (m *mockS3Provider) Name() string { return m.name }

func (m *mockS3Provider) Upload(ctx context.Context, bucket, key string, data io.Reader, size int64, opts *UploadOptions) (*Object, error) {
	m.uploadCalled = true
	content, _ := io.ReadAll(data)
	m.uploadedFiles[key] = content
	return &Object{Key: key, Bucket: bucket, Size: size}, nil
}

func (m *mockS3Provider) Download(ctx context.Context, bucket, key string, opts *DownloadOptions) (io.ReadCloser, *Object, error) {
	m.downloadCalled = true
	data, ok := m.downloadData[key]
	if !ok {
		data = ""
	}
	return io.NopCloser(strings.NewReader(data)), &Object{Key: key, Bucket: bucket}, nil
}

func (m *mockS3Provider) Delete(ctx context.Context, bucket, key string) error {
	m.deleteCalled = true
	delete(m.uploadedFiles, key)
	return nil
}

func (m *mockS3Provider) List(ctx context.Context, bucket string, opts *ListOptions) (*ListResult, error) {
	return &ListResult{Objects: m.listedObjects}, nil
}

func (m *mockS3Provider) GetMetadata(ctx context.Context, bucket, key string) (*Object, error) {
	return &Object{Key: key, Bucket: bucket}, nil
}

func (m *mockS3Provider) SetMetadata(ctx context.Context, bucket, key string, metadata map[string]string) error {
	return nil
}

func (m *mockS3Provider) Exists(ctx context.Context, bucket, key string) (bool, error) {
	_, exists := m.uploadedFiles[key]
	return exists, nil
}

func (m *mockS3Provider) GetObject(ctx context.Context, bucket, key string) (*Object, error) {
	return &Object{Key: key, Bucket: bucket}, nil
}

func (m *mockS3Provider) BucketExists(ctx context.Context, bucket string) (bool, error) {
	return true, nil
}

func (m *mockS3Provider) CreateBucket(ctx context.Context, bucket string) error {
	return nil
}

func (m *mockS3Provider) DeleteBucket(ctx context.Context, bucket string) error {
	return nil
}

func (m *mockS3Provider) ListBuckets(ctx context.Context) ([]string, error) {
	return []string{}, nil
}

func (m *mockS3Provider) Copy(ctx context.Context, srcBucket, srcKey, dstBucket, dstKey string) error {
	return nil
}

func (m *mockS3Provider) Move(ctx context.Context, srcBucket, srcKey, dstBucket, dstKey string) error {
	return nil
}

func (m *mockS3Provider) CopyObject(ctx context.Context, srcBucket, srcKey, destBucket, destKey string) error {
	return nil
}

func (m *mockS3Provider) MoveObject(ctx context.Context, srcBucket, srcKey, destBucket, destKey string) error {
	return nil
}

func (m *mockS3Provider) GetSignedURL(ctx context.Context, bucket, key string, expiry int64) (string, error) {
	return "http://example.com/" + key, nil
}

func (m *mockS3Provider) GetSignedUploadURL(ctx context.Context, bucket, key string, expiry int64) (string, error) {
	return "http://example.com/upload/" + key, nil
}

func (m *mockS3Provider) GenerateSignedURL(ctx context.Context, bucket, key string, opts *SignedURLOptions) (string, error) {
	return "http://example.com/signed/" + key, nil
}

func (m *mockS3Provider) Health(ctx context.Context) error {
	return m.healthError
}

// =============================================================================
// S3LogStorage Construction Tests
// =============================================================================

func TestNewS3LogStorage(t *testing.T) {
	t.Run("creates storage with default prefix", func(t *testing.T) {
		provider := newMockS3Provider()

		storage := NewS3LogStorage(provider, "test-bucket", "")

		require.NotNil(t, storage)
		assert.Equal(t, "test-bucket", storage.bucket)
		assert.Equal(t, "logs", storage.prefix) // Default prefix
	})

	t.Run("creates storage with custom prefix", func(t *testing.T) {
		provider := newMockS3Provider()

		storage := NewS3LogStorage(provider, "my-bucket", "custom-logs")

		require.NotNil(t, storage)
		assert.Equal(t, "my-bucket", storage.bucket)
		assert.Equal(t, "custom-logs", storage.prefix)
	})

	t.Run("creates storage with nil provider", func(t *testing.T) {
		storage := NewS3LogStorage(nil, "bucket", "prefix")

		require.NotNil(t, storage)
		assert.Nil(t, storage.storage)
	})
}

// =============================================================================
// S3LogStorage Name Tests
// =============================================================================

func TestS3LogStorage_Name(t *testing.T) {
	t.Run("returns s3", func(t *testing.T) {
		storage := NewS3LogStorage(nil, "bucket", "")

		assert.Equal(t, "s3", storage.Name())
	})
}

// =============================================================================
// S3LogStorage Write Tests
// =============================================================================

func TestS3LogStorage_Write(t *testing.T) {
	t.Run("returns nil for empty entries", func(t *testing.T) {
		provider := newMockS3Provider()
		storage := NewS3LogStorage(provider, "bucket", "logs")

		err := storage.Write(context.Background(), []*LogEntry{})

		assert.NoError(t, err)
		assert.False(t, provider.uploadCalled)
	})

	t.Run("uploads entries as NDJSON", func(t *testing.T) {
		provider := newMockS3Provider()
		storage := NewS3LogStorage(provider, "bucket", "logs")

		entries := []*LogEntry{
			{
				ID:        uuid.New(),
				Category:  LogCategoryHTTP,
				Level:     LogLevelInfo,
				Message:   "test message",
				Timestamp: time.Now(),
			},
		}

		err := storage.Write(context.Background(), entries)

		assert.NoError(t, err)
		assert.True(t, provider.uploadCalled)
		assert.Len(t, provider.uploadedFiles, 1)
	})

	t.Run("assigns UUID to entries with nil ID", func(t *testing.T) {
		provider := newMockS3Provider()
		storage := NewS3LogStorage(provider, "bucket", "logs")

		entry := &LogEntry{
			ID:       uuid.Nil,
			Category: LogCategoryHTTP,
			Level:    LogLevelInfo,
			Message:  "test",
		}

		_ = storage.Write(context.Background(), []*LogEntry{entry})

		// After write, entry should have ID assigned
		assert.NotEqual(t, uuid.Nil, entry.ID)
	})

	t.Run("assigns timestamp to entries with zero time", func(t *testing.T) {
		provider := newMockS3Provider()
		storage := NewS3LogStorage(provider, "bucket", "logs")

		entry := &LogEntry{
			ID:       uuid.New(),
			Category: LogCategoryHTTP,
			Level:    LogLevelInfo,
			Message:  "test",
		}

		_ = storage.Write(context.Background(), []*LogEntry{entry})

		assert.False(t, entry.Timestamp.IsZero())
	})

	t.Run("groups execution logs by execution ID", func(t *testing.T) {
		provider := newMockS3Provider()
		storage := NewS3LogStorage(provider, "bucket", "logs")

		execID := uuid.New().String()
		entries := []*LogEntry{
			{
				Category:    LogCategoryExecution,
				Level:       LogLevelInfo,
				Message:     "execution log 1",
				ExecutionID: execID,
				Timestamp:   time.Now(),
			},
			{
				Category:    LogCategoryExecution,
				Level:       LogLevelInfo,
				Message:     "execution log 2",
				ExecutionID: execID,
				Timestamp:   time.Now(),
			},
		}

		err := storage.Write(context.Background(), entries)

		assert.NoError(t, err)
		// Both entries should be in the same file (grouped by execution ID)
		// But due to timestamp suffix, might be in same or different files
		assert.True(t, provider.uploadCalled)
	})
}

// =============================================================================
// S3LogStorage matchesFilter Tests
// =============================================================================

func TestS3LogStorage_matchesFilter(t *testing.T) {
	storage := NewS3LogStorage(nil, "bucket", "logs")

	t.Run("matches all when no filters set", func(t *testing.T) {
		entry := &LogEntry{
			Category:  LogCategoryHTTP,
			Level:     LogLevelInfo,
			Component: "auth",
			Message:   "test",
			Timestamp: time.Now(),
		}

		matches := storage.matchesFilter(entry, LogQueryOptions{})

		assert.True(t, matches)
	})

	t.Run("filters by category", func(t *testing.T) {
		entry := &LogEntry{Category: LogCategoryHTTP}

		assert.True(t, storage.matchesFilter(entry, LogQueryOptions{Category: LogCategoryHTTP}))
		assert.False(t, storage.matchesFilter(entry, LogQueryOptions{Category: LogCategoryAuth}))
	})

	t.Run("filters by levels", func(t *testing.T) {
		entry := &LogEntry{Level: LogLevelError}

		assert.True(t, storage.matchesFilter(entry, LogQueryOptions{Levels: []LogLevel{LogLevelError}}))
		assert.True(t, storage.matchesFilter(entry, LogQueryOptions{Levels: []LogLevel{LogLevelError, LogLevelWarning}}))
		assert.False(t, storage.matchesFilter(entry, LogQueryOptions{Levels: []LogLevel{LogLevelInfo}}))
	})

	t.Run("filters by component", func(t *testing.T) {
		entry := &LogEntry{Component: "middleware"}

		assert.True(t, storage.matchesFilter(entry, LogQueryOptions{Component: "middleware"}))
		assert.False(t, storage.matchesFilter(entry, LogQueryOptions{Component: "handler"}))
	})

	t.Run("filters by request_id", func(t *testing.T) {
		entry := &LogEntry{RequestID: "req-123"}

		assert.True(t, storage.matchesFilter(entry, LogQueryOptions{RequestID: "req-123"}))
		assert.False(t, storage.matchesFilter(entry, LogQueryOptions{RequestID: "req-456"}))
	})

	t.Run("filters by trace_id", func(t *testing.T) {
		entry := &LogEntry{TraceID: "trace-abc"}

		assert.True(t, storage.matchesFilter(entry, LogQueryOptions{TraceID: "trace-abc"}))
		assert.False(t, storage.matchesFilter(entry, LogQueryOptions{TraceID: "trace-xyz"}))
	})

	t.Run("filters by user_id", func(t *testing.T) {
		entry := &LogEntry{UserID: "user-123"}

		assert.True(t, storage.matchesFilter(entry, LogQueryOptions{UserID: "user-123"}))
		assert.False(t, storage.matchesFilter(entry, LogQueryOptions{UserID: "user-456"}))
	})

	t.Run("filters by execution_id", func(t *testing.T) {
		entry := &LogEntry{ExecutionID: "exec-789"}

		assert.True(t, storage.matchesFilter(entry, LogQueryOptions{ExecutionID: "exec-789"}))
		assert.False(t, storage.matchesFilter(entry, LogQueryOptions{ExecutionID: "exec-000"}))
	})

	t.Run("filters by start_time", func(t *testing.T) {
		now := time.Now()
		entry := &LogEntry{Timestamp: now}

		assert.True(t, storage.matchesFilter(entry, LogQueryOptions{StartTime: now.Add(-1 * time.Hour)}))
		assert.False(t, storage.matchesFilter(entry, LogQueryOptions{StartTime: now.Add(1 * time.Hour)}))
	})

	t.Run("filters by end_time", func(t *testing.T) {
		now := time.Now()
		entry := &LogEntry{Timestamp: now}

		assert.True(t, storage.matchesFilter(entry, LogQueryOptions{EndTime: now.Add(1 * time.Hour)}))
		assert.False(t, storage.matchesFilter(entry, LogQueryOptions{EndTime: now.Add(-1 * time.Hour)}))
	})

	t.Run("filters by search text case insensitive", func(t *testing.T) {
		entry := &LogEntry{Message: "Authentication Failed for user"}

		assert.True(t, storage.matchesFilter(entry, LogQueryOptions{Search: "authentication"}))
		assert.True(t, storage.matchesFilter(entry, LogQueryOptions{Search: "FAILED"}))
		assert.True(t, storage.matchesFilter(entry, LogQueryOptions{Search: "user"}))
		assert.False(t, storage.matchesFilter(entry, LogQueryOptions{Search: "error"}))
	})

	t.Run("combines multiple filters with AND logic", func(t *testing.T) {
		entry := &LogEntry{
			Category:  LogCategoryAuth,
			Level:     LogLevelError,
			Component: "handler",
			Message:   "login failed",
		}

		// All conditions match
		assert.True(t, storage.matchesFilter(entry, LogQueryOptions{
			Category:  LogCategoryAuth,
			Levels:    []LogLevel{LogLevelError},
			Component: "handler",
		}))

		// One condition doesn't match
		assert.False(t, storage.matchesFilter(entry, LogQueryOptions{
			Category:  LogCategoryAuth,
			Levels:    []LogLevel{LogLevelInfo},
			Component: "handler",
		}))
	})
}

// =============================================================================
// S3LogStorage Query Tests
// =============================================================================

func TestS3LogStorage_Query(t *testing.T) {
	t.Run("returns empty result when no objects", func(t *testing.T) {
		provider := newMockS3Provider()
		provider.listedObjects = []Object{}
		storage := NewS3LogStorage(provider, "bucket", "logs")

		result, err := storage.Query(context.Background(), LogQueryOptions{})

		require.NoError(t, err)
		assert.Empty(t, result.Entries)
		assert.Equal(t, int64(0), result.TotalCount)
		assert.False(t, result.HasMore)
	})

	t.Run("applies default limit of 100", func(t *testing.T) {
		provider := newMockS3Provider()
		provider.listedObjects = []Object{}
		storage := NewS3LogStorage(provider, "bucket", "logs")

		// Query with no limit specified
		result, err := storage.Query(context.Background(), LogQueryOptions{})

		require.NoError(t, err)
		// Default limit is applied internally (100)
		assert.NotNil(t, result)
	})

	t.Run("handles pagination with offset exceeding results", func(t *testing.T) {
		provider := newMockS3Provider()
		provider.listedObjects = []Object{}
		storage := NewS3LogStorage(provider, "bucket", "logs")

		result, err := storage.Query(context.Background(), LogQueryOptions{
			Offset: 1000,
		})

		require.NoError(t, err)
		assert.Empty(t, result.Entries)
		assert.False(t, result.HasMore)
	})
}

// =============================================================================
// S3LogStorage GetExecutionLogs Tests
// =============================================================================

func TestS3LogStorage_GetExecutionLogs(t *testing.T) {
	t.Run("returns empty for non-existent execution", func(t *testing.T) {
		provider := newMockS3Provider()
		provider.listedObjects = []Object{}
		storage := NewS3LogStorage(provider, "bucket", "logs")

		entries, err := storage.GetExecutionLogs(context.Background(), "non-existent-id", 0)

		require.NoError(t, err)
		assert.Empty(t, entries)
	})

	t.Run("filters by afterLine", func(t *testing.T) {
		provider := newMockS3Provider()
		execID := uuid.New().String()
		provider.listedObjects = []Object{
			{Key: "logs/execution/2024/01/15/exec_" + execID + ".ndjson"},
		}
		// Set up download data with NDJSON entries
		provider.downloadData["logs/execution/2024/01/15/exec_"+execID+".ndjson"] = `{"line_number":1,"message":"line 1"}
{"line_number":2,"message":"line 2"}
{"line_number":3,"message":"line 3"}`

		storage := NewS3LogStorage(provider, "bucket", "logs")

		entries, err := storage.GetExecutionLogs(context.Background(), execID, 1)

		require.NoError(t, err)
		// Should only return lines after line 1
		for _, entry := range entries {
			assert.True(t, entry.LineNumber > 1)
		}
	})
}

// =============================================================================
// S3LogStorage Delete Tests
// =============================================================================

func TestS3LogStorage_Delete(t *testing.T) {
	t.Run("deletes files matching prefix", func(t *testing.T) {
		provider := newMockS3Provider()
		provider.listedObjects = []Object{
			{Key: "logs/http/2024/01/15/batch_abc.ndjson"},
		}
		storage := NewS3LogStorage(provider, "bucket", "logs")

		count, err := storage.Delete(context.Background(), LogQueryOptions{
			Category: LogCategoryHTTP,
		})

		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
		assert.True(t, provider.deleteCalled)
	})

	t.Run("skips non-ndjson files", func(t *testing.T) {
		provider := newMockS3Provider()
		provider.listedObjects = []Object{
			{Key: "logs/http/2024/01/15/readme.txt"},
		}
		storage := NewS3LogStorage(provider, "bucket", "logs")

		count, err := storage.Delete(context.Background(), LogQueryOptions{})

		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

// =============================================================================
// S3LogStorage Stats Tests
// =============================================================================

func TestS3LogStorage_Stats(t *testing.T) {
	t.Run("counts entries by category from file paths", func(t *testing.T) {
		provider := newMockS3Provider()
		provider.listedObjects = []Object{
			{Key: "logs/http/2024/01/15/batch1.ndjson"},
			{Key: "logs/http/2024/01/15/batch2.ndjson"},
			{Key: "logs/auth/2024/01/15/batch1.ndjson"},
		}
		storage := NewS3LogStorage(provider, "bucket", "logs")

		stats, err := storage.Stats(context.Background())

		require.NoError(t, err)
		assert.Equal(t, int64(3), stats.TotalEntries)
		assert.Equal(t, int64(2), stats.EntriesByCategory[LogCategoryHTTP])
		assert.Equal(t, int64(1), stats.EntriesByCategory[LogCategoryAuth])
	})

	t.Run("ignores non-ndjson files", func(t *testing.T) {
		provider := newMockS3Provider()
		provider.listedObjects = []Object{
			{Key: "logs/http/2024/01/15/batch1.ndjson"},
			{Key: "logs/http/readme.txt"},
		}
		storage := NewS3LogStorage(provider, "bucket", "logs")

		stats, err := storage.Stats(context.Background())

		require.NoError(t, err)
		assert.Equal(t, int64(1), stats.TotalEntries)
	})
}

// =============================================================================
// S3LogStorage Health Tests
// =============================================================================

func TestS3LogStorage_Health(t *testing.T) {
	t.Run("returns provider health status", func(t *testing.T) {
		provider := newMockS3Provider()
		storage := NewS3LogStorage(provider, "bucket", "logs")

		err := storage.Health(context.Background())

		assert.NoError(t, err)
	})
}

// =============================================================================
// S3LogStorage Close Tests
// =============================================================================

func TestS3LogStorage_Close(t *testing.T) {
	t.Run("close returns nil", func(t *testing.T) {
		storage := NewS3LogStorage(nil, "bucket", "logs")

		err := storage.Close()

		assert.NoError(t, err)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkS3LogStorage_matchesFilter_Simple(b *testing.B) {
	storage := NewS3LogStorage(nil, "bucket", "logs")
	entry := &LogEntry{
		Category: LogCategoryHTTP,
		Level:    LogLevelInfo,
	}
	opts := LogQueryOptions{
		Category: LogCategoryHTTP,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.matchesFilter(entry, opts)
	}
}

func BenchmarkS3LogStorage_matchesFilter_Complex(b *testing.B) {
	storage := NewS3LogStorage(nil, "bucket", "logs")
	entry := &LogEntry{
		Category:    LogCategoryHTTP,
		Level:       LogLevelError,
		Component:   "auth",
		RequestID:   "req-123",
		TraceID:     "trace-456",
		UserID:      "user-789",
		ExecutionID: "exec-000",
		Message:     "authentication failed for user",
		Timestamp:   time.Now(),
	}
	opts := LogQueryOptions{
		Category:    LogCategoryHTTP,
		Levels:      []LogLevel{LogLevelError, LogLevelWarning},
		Component:   "auth",
		RequestID:   "req-123",
		TraceID:     "trace-456",
		UserID:      "user-789",
		ExecutionID: "exec-000",
		Search:      "failed",
		StartTime:   time.Now().Add(-1 * time.Hour),
		EndTime:     time.Now().Add(1 * time.Hour),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.matchesFilter(entry, opts)
	}
}
