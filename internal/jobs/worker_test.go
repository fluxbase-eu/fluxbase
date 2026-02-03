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

// =============================================================================
// Worker Running Jobs Tests
// =============================================================================

func TestWorker_RunningJobs(t *testing.T) {
	t.Run("starts with empty running jobs map", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.NotNil(t, worker.runningJobs)
		assert.Empty(t, worker.runningJobs)
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

	t.Run("decrement does not go below zero", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		assert.Equal(t, 0, worker.currentJobCount)

		worker.decrementJobCount()
		assert.GreaterOrEqual(t, worker.currentJobCount, 0)
	})
}

// =============================================================================
// Worker HasCapacity Tests
// =============================================================================

func TestWorker_HasCapacity(t *testing.T) {
	t.Run("has capacity when no jobs running", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.True(t, worker.HasCapacity())
	})

	t.Run("has capacity when below max", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		worker.currentJobCount = 3

		assert.True(t, worker.HasCapacity())
	})

	t.Run("no capacity when at max", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		worker.currentJobCount = 5

		assert.False(t, worker.HasCapacity())
	})

	t.Run("no capacity when draining", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		worker.draining = true

		assert.False(t, worker.HasCapacity())
	})
}

// =============================================================================
// Worker Drain Tests
// =============================================================================

func TestWorker_Drain(t *testing.T) {
	t.Run("drain sets draining flag", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		assert.False(t, worker.draining)

		worker.Drain()
		assert.True(t, worker.draining)
	})

	t.Run("drain is idempotent", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		worker.Drain()
		worker.Drain()
		assert.True(t, worker.draining)
	})
}

// =============================================================================
// Worker IsDraining Tests
// =============================================================================

func TestWorker_IsDraining(t *testing.T) {
	t.Run("returns false when not draining", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.False(t, worker.IsDraining())
	})

	t.Run("returns true when draining", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		worker.Drain()

		assert.True(t, worker.IsDraining())
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
		assert.Equal(t, 0, worker.GetCurrentJobCount())

		worker.currentJobCount = 3
		assert.Equal(t, 3, worker.GetCurrentJobCount())
	})
}

// =============================================================================
// Job Execution Result Tests
// =============================================================================

func TestJobExecutionResult_Struct(t *testing.T) {
	t.Run("creates success result", func(t *testing.T) {
		result := &JobExecutionResult{
			JobID:      uuid.New(),
			Success:    true,
			Output:     "Job completed successfully",
			Error:      "",
			Duration:   5 * time.Second,
			ExitCode:   0,
			RetryCount: 0,
		}

		assert.True(t, result.Success)
		assert.Empty(t, result.Error)
		assert.Equal(t, 0, result.ExitCode)
		assert.Equal(t, "Job completed successfully", result.Output)
	})

	t.Run("creates failure result", func(t *testing.T) {
		result := &JobExecutionResult{
			JobID:      uuid.New(),
			Success:    false,
			Output:     "",
			Error:      "Process exited with code 1",
			Duration:   2 * time.Second,
			ExitCode:   1,
			RetryCount: 2,
		}

		assert.False(t, result.Success)
		assert.NotEmpty(t, result.Error)
		assert.Equal(t, 1, result.ExitCode)
		assert.Equal(t, 2, result.RetryCount)
	})

	t.Run("zero value", func(t *testing.T) {
		var result JobExecutionResult

		assert.Equal(t, uuid.Nil, result.JobID)
		assert.False(t, result.Success)
		assert.Empty(t, result.Output)
		assert.Empty(t, result.Error)
		assert.Equal(t, time.Duration(0), result.Duration)
		assert.Equal(t, 0, result.ExitCode)
	})
}

// =============================================================================
// Worker Config Tests
// =============================================================================

func TestWorker_ConfigValues(t *testing.T) {
	t.Run("stores config reference", func(t *testing.T) {
		cfg := &config.JobsConfig{
			WorkerMode:             "deno",
			MaxConcurrentPerWorker: 8,
			DefaultMaxDuration:     45 * time.Minute,
			DefaultRetries:         3,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.Equal(t, cfg, worker.Config)
		assert.Equal(t, "deno", worker.Config.WorkerMode)
		assert.Equal(t, 45*time.Minute, worker.Config.DefaultMaxDuration)
		assert.Equal(t, 3, worker.Config.DefaultRetries)
	})
}

// =============================================================================
// Worker JWT Secret Tests
// =============================================================================

func TestWorker_JWTSecret(t *testing.T) {
	t.Run("stores jwt secret", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "my-super-secret-jwt-key", "http://localhost", nil)

		assert.Equal(t, "my-super-secret-jwt-key", worker.jwtSecret)
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
				worker.HasCapacity()
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
