package rpc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Scheduler Construction Tests
// =============================================================================

func TestNewScheduler(t *testing.T) {
	t.Run("creates scheduler with nil dependencies", func(t *testing.T) {
		scheduler := NewScheduler(nil, nil)

		require.NotNil(t, scheduler)
		assert.Nil(t, scheduler.storage)
		assert.Nil(t, scheduler.executor)
		assert.NotNil(t, scheduler.cron)
		assert.NotNil(t, scheduler.procedureJobs)
		assert.Equal(t, 10, scheduler.maxConcurrent)
		assert.NotNil(t, scheduler.ctx)
		assert.NotNil(t, scheduler.cancel)
	})

	t.Run("initializes empty procedureJobs map", func(t *testing.T) {
		scheduler := NewScheduler(nil, nil)

		assert.Empty(t, scheduler.procedureJobs)
	})
}

// =============================================================================
// ScheduleProcedure Tests
// =============================================================================

func TestScheduler_ScheduleProcedure(t *testing.T) {
	t.Run("returns nil for nil schedule", func(t *testing.T) {
		scheduler := NewScheduler(nil, nil)

		proc := &Procedure{
			Name:      "test_proc",
			Namespace: "public",
			Schedule:  nil,
		}

		err := scheduler.ScheduleProcedure(proc)

		assert.NoError(t, err)
		assert.False(t, scheduler.IsScheduled("public", "test_proc"))
	})

	t.Run("returns nil for empty schedule", func(t *testing.T) {
		scheduler := NewScheduler(nil, nil)

		emptySchedule := ""
		proc := &Procedure{
			Name:      "test_proc",
			Namespace: "public",
			Schedule:  &emptySchedule,
		}

		err := scheduler.ScheduleProcedure(proc)

		assert.NoError(t, err)
		assert.False(t, scheduler.IsScheduled("public", "test_proc"))
	})

	t.Run("schedules valid cron expression", func(t *testing.T) {
		scheduler := NewScheduler(nil, nil)
		defer scheduler.Stop()

		schedule := "*/5 * * * *" // Every 5 minutes
		proc := &Procedure{
			Name:      "test_proc",
			Namespace: "public",
			Schedule:  &schedule,
		}

		err := scheduler.ScheduleProcedure(proc)

		assert.NoError(t, err)
		assert.True(t, scheduler.IsScheduled("public", "test_proc"))
	})

	t.Run("schedules 6-field cron with seconds", func(t *testing.T) {
		scheduler := NewScheduler(nil, nil)
		defer scheduler.Stop()

		schedule := "0 */5 * * * *" // Every 5 minutes at second 0
		proc := &Procedure{
			Name:      "test_proc",
			Namespace: "public",
			Schedule:  &schedule,
		}

		err := scheduler.ScheduleProcedure(proc)

		assert.NoError(t, err)
		assert.True(t, scheduler.IsScheduled("public", "test_proc"))
	})

	t.Run("returns error for invalid cron expression", func(t *testing.T) {
		scheduler := NewScheduler(nil, nil)

		schedule := "invalid cron"
		proc := &Procedure{
			Name:      "test_proc",
			Namespace: "public",
			Schedule:  &schedule,
		}

		err := scheduler.ScheduleProcedure(proc)

		assert.Error(t, err)
		assert.False(t, scheduler.IsScheduled("public", "test_proc"))
	})

	t.Run("replaces existing schedule", func(t *testing.T) {
		scheduler := NewScheduler(nil, nil)
		defer scheduler.Stop()

		schedule1 := "*/5 * * * *"
		proc := &Procedure{
			Name:      "test_proc",
			Namespace: "public",
			Schedule:  &schedule1,
		}

		err := scheduler.ScheduleProcedure(proc)
		require.NoError(t, err)

		// Get the first entry ID
		scheduler.jobsMu.RLock()
		firstEntryID := scheduler.procedureJobs["public/test_proc"]
		scheduler.jobsMu.RUnlock()

		// Schedule again with different schedule
		schedule2 := "*/10 * * * *"
		proc.Schedule = &schedule2

		err = scheduler.ScheduleProcedure(proc)
		require.NoError(t, err)

		// Entry ID should be different (old one removed, new one added)
		scheduler.jobsMu.RLock()
		secondEntryID := scheduler.procedureJobs["public/test_proc"]
		scheduler.jobsMu.RUnlock()

		assert.NotEqual(t, firstEntryID, secondEntryID)
	})
}

// =============================================================================
// UnscheduleProcedure Tests
// =============================================================================

func TestScheduler_UnscheduleProcedure(t *testing.T) {
	t.Run("unschedules existing procedure", func(t *testing.T) {
		scheduler := NewScheduler(nil, nil)
		defer scheduler.Stop()

		schedule := "*/5 * * * *"
		proc := &Procedure{
			Name:      "test_proc",
			Namespace: "public",
			Schedule:  &schedule,
		}

		_ = scheduler.ScheduleProcedure(proc)
		assert.True(t, scheduler.IsScheduled("public", "test_proc"))

		scheduler.UnscheduleProcedure("public", "test_proc")

		assert.False(t, scheduler.IsScheduled("public", "test_proc"))
	})

	t.Run("handles unscheduling non-existent procedure", func(t *testing.T) {
		scheduler := NewScheduler(nil, nil)

		// Should not panic
		scheduler.UnscheduleProcedure("public", "non_existent")

		assert.False(t, scheduler.IsScheduled("public", "non_existent"))
	})
}

// =============================================================================
// RescheduleProcedure Tests
// =============================================================================

func TestScheduler_RescheduleProcedure(t *testing.T) {
	t.Run("reschedules enabled procedure", func(t *testing.T) {
		scheduler := NewScheduler(nil, nil)
		defer scheduler.Stop()

		schedule := "*/5 * * * *"
		proc := &Procedure{
			Name:      "test_proc",
			Namespace: "public",
			Schedule:  &schedule,
			Enabled:   true,
		}

		_ = scheduler.ScheduleProcedure(proc)

		newSchedule := "*/10 * * * *"
		proc.Schedule = &newSchedule

		err := scheduler.RescheduleProcedure(proc)

		assert.NoError(t, err)
		assert.True(t, scheduler.IsScheduled("public", "test_proc"))
	})

	t.Run("removes schedule for disabled procedure", func(t *testing.T) {
		scheduler := NewScheduler(nil, nil)
		defer scheduler.Stop()

		schedule := "*/5 * * * *"
		proc := &Procedure{
			Name:      "test_proc",
			Namespace: "public",
			Schedule:  &schedule,
			Enabled:   true,
		}

		_ = scheduler.ScheduleProcedure(proc)
		assert.True(t, scheduler.IsScheduled("public", "test_proc"))

		// Disable the procedure
		proc.Enabled = false

		err := scheduler.RescheduleProcedure(proc)

		assert.NoError(t, err)
		assert.False(t, scheduler.IsScheduled("public", "test_proc"))
	})

	t.Run("removes schedule when schedule is nil", func(t *testing.T) {
		scheduler := NewScheduler(nil, nil)
		defer scheduler.Stop()

		schedule := "*/5 * * * *"
		proc := &Procedure{
			Name:      "test_proc",
			Namespace: "public",
			Schedule:  &schedule,
			Enabled:   true,
		}

		_ = scheduler.ScheduleProcedure(proc)
		assert.True(t, scheduler.IsScheduled("public", "test_proc"))

		// Remove schedule
		proc.Schedule = nil

		err := scheduler.RescheduleProcedure(proc)

		assert.NoError(t, err)
		assert.False(t, scheduler.IsScheduled("public", "test_proc"))
	})
}

// =============================================================================
// IsScheduled Tests
// =============================================================================

func TestScheduler_IsScheduled(t *testing.T) {
	t.Run("returns false for non-existent procedure", func(t *testing.T) {
		scheduler := NewScheduler(nil, nil)

		result := scheduler.IsScheduled("public", "non_existent")

		assert.False(t, result)
	})

	t.Run("returns true for scheduled procedure", func(t *testing.T) {
		scheduler := NewScheduler(nil, nil)
		defer scheduler.Stop()

		schedule := "*/5 * * * *"
		proc := &Procedure{
			Name:      "test_proc",
			Namespace: "public",
			Schedule:  &schedule,
		}

		_ = scheduler.ScheduleProcedure(proc)

		result := scheduler.IsScheduled("public", "test_proc")

		assert.True(t, result)
	})

	t.Run("handles different namespaces", func(t *testing.T) {
		scheduler := NewScheduler(nil, nil)
		defer scheduler.Stop()

		schedule := "*/5 * * * *"
		proc := &Procedure{
			Name:      "test_proc",
			Namespace: "namespace1",
			Schedule:  &schedule,
		}

		_ = scheduler.ScheduleProcedure(proc)

		assert.True(t, scheduler.IsScheduled("namespace1", "test_proc"))
		assert.False(t, scheduler.IsScheduled("namespace2", "test_proc"))
	})
}

// =============================================================================
// GetScheduledProcedures Tests
// =============================================================================

func TestScheduler_GetScheduledProcedures(t *testing.T) {
	t.Run("returns empty slice when no procedures scheduled", func(t *testing.T) {
		scheduler := NewScheduler(nil, nil)

		procs := scheduler.GetScheduledProcedures()

		assert.Empty(t, procs)
	})

	t.Run("returns all scheduled procedures", func(t *testing.T) {
		scheduler := NewScheduler(nil, nil)
		defer scheduler.Stop()

		schedule := "*/5 * * * *"
		proc1 := &Procedure{Name: "proc1", Namespace: "public", Schedule: &schedule}
		proc2 := &Procedure{Name: "proc2", Namespace: "public", Schedule: &schedule}

		_ = scheduler.ScheduleProcedure(proc1)
		_ = scheduler.ScheduleProcedure(proc2)

		procs := scheduler.GetScheduledProcedures()

		assert.Len(t, procs, 2)
	})

	t.Run("includes next run time", func(t *testing.T) {
		scheduler := NewScheduler(nil, nil)
		defer scheduler.Stop()
		scheduler.cron.Start()

		schedule := "*/1 * * * *" // Every minute
		proc := &Procedure{Name: "test_proc", Namespace: "public", Schedule: &schedule}

		_ = scheduler.ScheduleProcedure(proc)

		procs := scheduler.GetScheduledProcedures()

		require.Len(t, procs, 1)
		assert.Equal(t, "public/test_proc", procs[0].Key)
		assert.False(t, procs[0].NextRun.IsZero())
	})
}

// =============================================================================
// ScheduledProcedureInfo Tests
// =============================================================================

func TestScheduledProcedureInfo(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		now := time.Now()
		nextRun := now.Add(1 * time.Hour)
		prevRun := now.Add(-1 * time.Hour)

		info := ScheduledProcedureInfo{
			Key:     "namespace/proc_name",
			EntryID: 123,
			NextRun: nextRun,
			PrevRun: prevRun,
		}

		assert.Equal(t, "namespace/proc_name", info.Key)
		assert.Equal(t, 123, info.EntryID)
		assert.Equal(t, nextRun, info.NextRun)
		assert.Equal(t, prevRun, info.PrevRun)
	})
}

// =============================================================================
// Stop Tests
// =============================================================================

func TestScheduler_Stop(t *testing.T) {
	t.Run("stops scheduler gracefully", func(t *testing.T) {
		scheduler := NewScheduler(nil, nil)

		schedule := "*/5 * * * *"
		proc := &Procedure{Name: "test_proc", Namespace: "public", Schedule: &schedule}
		_ = scheduler.ScheduleProcedure(proc)

		// Stop should not panic
		scheduler.Stop()

		// Context should be cancelled
		select {
		case <-scheduler.ctx.Done():
			// Expected
		default:
			t.Error("context should be cancelled after stop")
		}
	})
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

func TestScheduler_ConcurrentAccess(t *testing.T) {
	t.Run("handles concurrent schedule/unschedule", func(t *testing.T) {
		scheduler := NewScheduler(nil, nil)
		defer scheduler.Stop()

		schedule := "*/5 * * * *"
		done := make(chan bool)

		// Schedule concurrently
		for i := 0; i < 10; i++ {
			go func(idx int) {
				proc := &Procedure{
					Name:      "test_proc",
					Namespace: "public",
					Schedule:  &schedule,
				}
				_ = scheduler.ScheduleProcedure(proc)
				done <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		// Should still be scheduled (only one entry)
		assert.True(t, scheduler.IsScheduled("public", "test_proc"))
	})
}

// =============================================================================
// Cron Expression Tests
// =============================================================================

func TestScheduler_CronExpressions(t *testing.T) {
	scheduler := NewScheduler(nil, nil)
	defer scheduler.Stop()

	testCases := []struct {
		name     string
		schedule string
		valid    bool
	}{
		{"every minute", "* * * * *", true},
		{"every 5 minutes", "*/5 * * * *", true},
		{"every hour", "0 * * * *", true},
		{"every day at midnight", "0 0 * * *", true},
		{"with seconds", "0 */5 * * * *", true},
		{"every monday", "0 0 * * MON", true},
		{"@hourly descriptor", "@hourly", true},
		{"@daily descriptor", "@daily", true},
		{"@weekly descriptor", "@weekly", true},
		{"invalid expression", "invalid", false},
		{"too few fields", "* * *", false},
		{"invalid minute", "60 * * * *", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			proc := &Procedure{
				Name:      "test_proc_" + tc.name,
				Namespace: "public",
				Schedule:  &tc.schedule,
			}

			err := scheduler.ScheduleProcedure(proc)

			if tc.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}

			// Cleanup
			scheduler.UnscheduleProcedure("public", proc.Name)
		})
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkScheduler_ScheduleProcedure(b *testing.B) {
	scheduler := NewScheduler(nil, nil)
	defer scheduler.Stop()

	schedule := "*/5 * * * *"
	proc := &Procedure{
		Name:      "bench_proc",
		Namespace: "public",
		Schedule:  &schedule,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scheduler.ScheduleProcedure(proc)
	}
}

func BenchmarkScheduler_IsScheduled(b *testing.B) {
	scheduler := NewScheduler(nil, nil)
	defer scheduler.Stop()

	schedule := "*/5 * * * *"
	proc := &Procedure{
		Name:      "bench_proc",
		Namespace: "public",
		Schedule:  &schedule,
	}
	_ = scheduler.ScheduleProcedure(proc)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scheduler.IsScheduled("public", "bench_proc")
	}
}
