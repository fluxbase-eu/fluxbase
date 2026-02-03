package jobs

import (
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

		// Simulate worker completion by closing shutdownComplete in a goroutine
		// (normally this would be done by the worker's goroutines when they finish)
		go func() {
			<-worker.shutdownChan // Wait for Stop() to close shutdownChan
			close(worker.shutdownComplete)
		}()

		// Stop should close the shutdown channel and wait for completion
		worker.Stop()

		// Verify shutdown channel was closed
		select {
		case <-worker.shutdownChan:
			// Expected - channel is closed
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
