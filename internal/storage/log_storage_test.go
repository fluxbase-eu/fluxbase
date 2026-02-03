package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultLogStorageConfig(t *testing.T) {
	t.Run("returns sensible defaults", func(t *testing.T) {
		cfg := DefaultLogStorageConfig()

		assert.Equal(t, "postgres", cfg.Backend)
		assert.Equal(t, "logs", cfg.S3Prefix)
		assert.Equal(t, "./logs", cfg.LocalPath)
		assert.Equal(t, 100, cfg.BatchSize)
		assert.Equal(t, 1000, cfg.FlushInterval)
		assert.Equal(t, 10000, cfg.BufferSize)
	})

	t.Run("S3Bucket is empty by default", func(t *testing.T) {
		cfg := DefaultLogStorageConfig()
		assert.Empty(t, cfg.S3Bucket)
	})
}

func TestLogStorageConfig_Struct(t *testing.T) {
	t.Run("all fields can be set", func(t *testing.T) {
		cfg := LogStorageConfig{
			Backend:       "s3",
			S3Bucket:      "my-logs-bucket",
			S3Prefix:      "app/logs",
			LocalPath:     "/var/log/app",
			BatchSize:     500,
			FlushInterval: 5000,
			BufferSize:    50000,
		}

		assert.Equal(t, "s3", cfg.Backend)
		assert.Equal(t, "my-logs-bucket", cfg.S3Bucket)
		assert.Equal(t, "app/logs", cfg.S3Prefix)
		assert.Equal(t, "/var/log/app", cfg.LocalPath)
		assert.Equal(t, 500, cfg.BatchSize)
		assert.Equal(t, 5000, cfg.FlushInterval)
		assert.Equal(t, 50000, cfg.BufferSize)
	})

	t.Run("postgres backend config", func(t *testing.T) {
		cfg := LogStorageConfig{
			Backend:       "postgres",
			BatchSize:     200,
			FlushInterval: 2000,
			BufferSize:    20000,
		}

		assert.Equal(t, "postgres", cfg.Backend)
		// S3 and Local settings should be empty/zero
		assert.Empty(t, cfg.S3Bucket)
		assert.Empty(t, cfg.S3Prefix)
		assert.Empty(t, cfg.LocalPath)
	})

	t.Run("local backend config", func(t *testing.T) {
		cfg := LogStorageConfig{
			Backend:       "local",
			LocalPath:     "/tmp/logs",
			BatchSize:     50,
			FlushInterval: 500,
		}

		assert.Equal(t, "local", cfg.Backend)
		assert.Equal(t, "/tmp/logs", cfg.LocalPath)
	})

	t.Run("zero values", func(t *testing.T) {
		var cfg LogStorageConfig

		assert.Empty(t, cfg.Backend)
		assert.Empty(t, cfg.S3Bucket)
		assert.Empty(t, cfg.S3Prefix)
		assert.Empty(t, cfg.LocalPath)
		assert.Equal(t, 0, cfg.BatchSize)
		assert.Equal(t, 0, cfg.FlushInterval)
		assert.Equal(t, 0, cfg.BufferSize)
	})
}

func TestLogStorageConfig_BackendTypes(t *testing.T) {
	t.Run("supported backend types", func(t *testing.T) {
		backends := []string{"postgres", "s3", "local"}

		for _, backend := range backends {
			cfg := LogStorageConfig{Backend: backend}
			assert.Equal(t, backend, cfg.Backend)
		}
	})
}
