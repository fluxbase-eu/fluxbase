package jobs

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/fluxbase-eu/fluxbase/internal/config"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Worker Construction Tests
// =============================================================================

func TestNewWorker(t *testing.T) {
	t.Run("creates worker with default config", func(t *testing.T) {
		cfg := &config.JobsConfig{
			WorkerMode:             "deno",
			MaxConcurrentPerWorker: 5,
			DefaultMaxDuration:     30 * time.Minute,
		}

		worker := NewWorker(cfg, nil, "jwt-secret", "http://localhost", nil)

		require.NotNil(t, worker)
		assert.NotEqual(t, uuid.Nil, worker.ID)
		assert.Contains(t, worker.Name, "worker-")
		assert.Equal(t, cfg, worker.Config)
		assert.Equal(t, 5, worker.MaxConcurrent)
		assert.NotNil(t, worker.Runtime)
		assert.NotNil(t, worker.shutdownChan)
		assert.NotNil(t, worker.shutdownComplete)
	})

	t.Run("generates unique worker ID", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker1 := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		worker2 := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.NotEqual(t, worker1.ID, worker2.ID)
	})

	t.Run("includes hostname in worker name", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.Contains(t, worker.Name, "@")
	})
}

// =============================================================================
// Worker State Tests
// =============================================================================

func TestWorker_DrainState(t *testing.T) {
	t.Run("starts not draining", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.False(t, worker.draining)
	})
}

func TestWorker_JobCount(t *testing.T) {
	t.Run("starts with zero jobs", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.Equal(t, 0, worker.currentJobCount)
	})
}

// =============================================================================
// WorkerRecord Struct Tests
// =============================================================================

func TestWorkerRecord_Struct(t *testing.T) {
	t.Run("basic worker record", func(t *testing.T) {
		hostname := "worker-host"
		name := "worker-1"

		record := WorkerRecord{
			ID:                uuid.New(),
			Name:              &name,
			Hostname:          &hostname,
			Status:            WorkerStatusActive,
			MaxConcurrentJobs: 5,
			CurrentJobCount:   2,
		}

		assert.NotEqual(t, uuid.Nil, record.ID)
		assert.Equal(t, "worker-1", *record.Name)
		assert.Equal(t, "worker-host", *record.Hostname)
		assert.Equal(t, WorkerStatusActive, record.Status)
		assert.Equal(t, 5, record.MaxConcurrentJobs)
		assert.Equal(t, 2, record.CurrentJobCount)
	})

	t.Run("worker record with metadata", func(t *testing.T) {
		metadata := map[string]interface{}{
			"hostname": "prod-worker-1",
			"pid":      12345,
			"version":  "1.0.0",
		}
		metadataJSON, _ := json.Marshal(metadata)
		metadataStr := string(metadataJSON)

		record := WorkerRecord{
			ID:       uuid.New(),
			Status:   WorkerStatusActive,
			Metadata: &metadataStr,
		}

		assert.NotNil(t, record.Metadata)
		assert.Contains(t, *record.Metadata, "hostname")
		assert.Contains(t, *record.Metadata, "pid")
	})

	t.Run("JSON serialization", func(t *testing.T) {
		name := "test-worker"
		hostname := "test-host"

		record := WorkerRecord{
			ID:                uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			Name:              &name,
			Hostname:          &hostname,
			Status:            WorkerStatusActive,
			MaxConcurrentJobs: 10,
			CurrentJobCount:   3,
		}

		data, err := json.Marshal(record)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"status":"active"`)
		assert.Contains(t, string(data), `"max_concurrent_jobs":10`)
		assert.Contains(t, string(data), `"current_job_count":3`)
	})
}

// =============================================================================
// Worker Status Constants Tests
// =============================================================================

func TestWorkerStatus_Constants(t *testing.T) {
	t.Run("active status", func(t *testing.T) {
		assert.Equal(t, WorkerStatus("active"), WorkerStatusActive)
	})

	t.Run("draining status", func(t *testing.T) {
		assert.Equal(t, WorkerStatus("draining"), WorkerStatusDraining)
	})

	t.Run("stopped status", func(t *testing.T) {
		assert.Equal(t, WorkerStatus("stopped"), WorkerStatusStopped)
	})
}

// =============================================================================
// Worker Concurrent Job Handling Tests
// =============================================================================

func TestWorker_MaxConcurrent(t *testing.T) {
	t.Run("respects max concurrent config", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 10,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.Equal(t, 10, worker.MaxConcurrent)
	})

	t.Run("single concurrent job", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 1,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.Equal(t, 1, worker.MaxConcurrent)
	})
}

// =============================================================================
// Worker Stop Tests
// =============================================================================

func TestWorker_Stop(t *testing.T) {
	t.Run("stop signals shutdown", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		// Stop should close the channel
		worker.Stop()

		// Channel should be closed
		select {
		case <-worker.shutdownChan:
			// Expected
		default:
			t.Error("Shutdown channel should be closed after Stop()")
		}
	})
}

// =============================================================================
// Worker Cancel Job Tests
// =============================================================================

func TestWorker_CancelJob(t *testing.T) {
	t.Run("cancel non-existent job does not panic", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		// Should not panic
		worker.cancelJob(uuid.New())
	})
}

// =============================================================================
// Worker Runtime Tests
// =============================================================================

func TestWorker_Runtime(t *testing.T) {
	t.Run("creates deno runtime", func(t *testing.T) {
		cfg := &config.JobsConfig{
			WorkerMode:             "deno",
			MaxConcurrentPerWorker: 5,
			DefaultMaxDuration:     30 * time.Minute,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.NotNil(t, worker.Runtime)
	})
}

// =============================================================================
// Worker Public URL Tests
// =============================================================================

func TestWorker_PublicURL(t *testing.T) {
	t.Run("stores public URL", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "https://api.example.com", nil)

		assert.Equal(t, "https://api.example.com", worker.publicURL)
	})
}
