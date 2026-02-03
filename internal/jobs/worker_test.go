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

func TestWorker_setDrainingState(t *testing.T) {
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

// =============================================================================
// Worker Running Jobs Tests
// =============================================================================

func TestWorker_CurrentJobs(t *testing.T) {
	t.Run("starts with empty current jobs map", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		// currentJobs is a sync.Map - verify it's usable
		count := 0
		worker.currentJobs.Range(func(k, v interface{}) bool {
			count++
			return true
		})
		assert.Equal(t, 0, count)
	})
}

// =============================================================================
// Worker Job Count Operations Tests
// =============================================================================

func TestWorker_JobCountOperations(t *testing.T) {
	t.Run("increment job count", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		assert.Equal(t, 0, worker.currentJobCount)

		worker.incrementJobCount()
		assert.Equal(t, 1, worker.currentJobCount)

		worker.incrementJobCount()
		assert.Equal(t, 2, worker.currentJobCount)
	})

	t.Run("decrement job count", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		worker.currentJobCount = 3

		worker.decrementJobCount()
		assert.Equal(t, 2, worker.currentJobCount)

		worker.decrementJobCount()
		assert.Equal(t, 1, worker.currentJobCount)
	})

	t.Run("decrement can go below zero", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		assert.Equal(t, 0, worker.currentJobCount)

		worker.decrementJobCount()
		// Note: The implementation doesn't prevent going negative
		assert.Equal(t, -1, worker.currentJobCount)
	})
}

// =============================================================================
// Worker hasCapacity Tests
// =============================================================================

func TestWorker_hasCapacity(t *testing.T) {
	t.Run("has capacity when no jobs running", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.True(t, worker.hasCapacity())
	})

	t.Run("has capacity when below max", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		worker.currentJobCount = 3

		assert.True(t, worker.hasCapacity())
	})

	t.Run("no capacity when at max", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		worker.currentJobCount = 5

		assert.False(t, worker.hasCapacity())
	})

	t.Run("hasCapacity ignores draining state", func(t *testing.T) {
		// Note: hasCapacity only checks job count vs max concurrent
		// It does not check draining state - that's checked separately in the poll loop
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		worker.draining = true

		// hasCapacity returns true because it only checks count < max
		assert.True(t, worker.hasCapacity())
	})
}

// =============================================================================
// Worker setDraining Tests
// =============================================================================

func TestWorker_setDraining(t *testing.T) {
	t.Run("drain sets draining flag", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		assert.False(t, worker.draining)

		worker.setDraining(true)
		assert.True(t, worker.draining)
	})

	t.Run("drain is idempotent", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		worker.setDraining(true)
		worker.setDraining(true)
		assert.True(t, worker.draining)
	})
}

// =============================================================================
// Worker isDraining Tests
// =============================================================================

func TestWorker_isDraining(t *testing.T) {
	t.Run("returns false when not draining", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.False(t, worker.isDraining())
	})

	t.Run("returns true when draining", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		worker.setDraining(true)

		assert.True(t, worker.isDraining())
	})
}

// =============================================================================
// Worker GetCurrentJobCount Tests
// =============================================================================

func TestWorker_GetCurrentJobCount(t *testing.T) {
	t.Run("returns current job count", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		assert.Equal(t, 0, worker.getCurrentJobCount())

		worker.currentJobCount = 3
		assert.Equal(t, 3, worker.getCurrentJobCount())
	})
}

// =============================================================================
// Worker Concurrent Access Tests
// =============================================================================

func TestWorker_ConcurrentJobCountAccess(t *testing.T) {
	cfg := &config.JobsConfig{
		MaxConcurrentPerWorker: 100,
	}

	worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

	// Run concurrent increments and decrements
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 50; j++ {
				worker.incrementJobCount()
			}
			done <- true
		}()
		go func() {
			for j := 0; j < 50; j++ {
				worker.hasCapacity()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Should complete without panics or data races
	assert.GreaterOrEqual(t, worker.currentJobCount, 0)
}

// =============================================================================
// Worker Shutdown Channel Tests
// =============================================================================

func TestWorker_ShutdownChannels(t *testing.T) {
	t.Run("creates shutdown channels", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.NotNil(t, worker.shutdownChan)
		assert.NotNil(t, worker.shutdownComplete)
	})
}

// =============================================================================
// Worker Mode Tests
// =============================================================================

func TestWorker_Modes(t *testing.T) {
	testCases := []struct {
		mode string
	}{
		{"deno"},
		{"docker"},
		{"process"},
	}

	for _, tc := range testCases {
		t.Run("mode "+tc.mode, func(t *testing.T) {
			cfg := &config.JobsConfig{
				WorkerMode:             tc.mode,
				MaxConcurrentPerWorker: 5,
			}

			worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
			assert.Equal(t, tc.mode, worker.Config.WorkerMode)
		})
	}
}
