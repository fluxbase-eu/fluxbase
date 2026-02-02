package functions

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Scheduler Construction Tests
// =============================================================================

func TestNewScheduler(t *testing.T) {
	t.Run("creates scheduler with nil dependencies", func(t *testing.T) {
		scheduler := NewScheduler(nil, "jwt-secret", "http://localhost", nil)
		require.NotNil(t, scheduler)
		assert.NotNil(t, scheduler.cron)
		assert.Equal(t, 10, scheduler.maxConcurrent)
		assert.Equal(t, "jwt-secret", scheduler.jwtSecret)
		assert.Equal(t, "http://localhost", scheduler.publicURL)
	})

	t.Run("initializes empty function jobs map", func(t *testing.T) {
		scheduler := NewScheduler(nil, "secret", "http://localhost", nil)
		assert.NotNil(t, scheduler.functionJobs)
		assert.Empty(t, scheduler.functionJobs)
	})

	t.Run("creates context with cancel", func(t *testing.T) {
		scheduler := NewScheduler(nil, "secret", "http://localhost", nil)
		assert.NotNil(t, scheduler.ctx)
		assert.NotNil(t, scheduler.cancel)
	})
}

// =============================================================================
// Scheduler Log Message Handling Tests
// =============================================================================

func TestScheduler_handleLogMessage(t *testing.T) {
	t.Run("handles log without counter", func(t *testing.T) {
		scheduler := NewScheduler(nil, "secret", "http://localhost", nil)
		execID := uuid.New()

		// Should not panic when no counter exists
		scheduler.handleLogMessage(execID, "info", "test message")
	})

	t.Run("increments counter when exists", func(t *testing.T) {
		scheduler := NewScheduler(nil, "secret", "http://localhost", nil)
		execID := uuid.New()

		// Set up counter
		counter := 0
		scheduler.logCounters.Store(execID, &counter)

		scheduler.handleLogMessage(execID, "info", "message 1")
		assert.Equal(t, 1, counter)

		scheduler.handleLogMessage(execID, "debug", "message 2")
		assert.Equal(t, 2, counter)

		scheduler.handleLogMessage(execID, "error", "message 3")
		assert.Equal(t, 3, counter)
	})

	t.Run("handles invalid counter type gracefully", func(t *testing.T) {
		scheduler := NewScheduler(nil, "secret", "http://localhost", nil)
		execID := uuid.New()

		// Store invalid type
		scheduler.logCounters.Store(execID, "not a pointer")

		// Should not panic
		scheduler.handleLogMessage(execID, "info", "test message")
	})

	t.Run("handles different log levels", func(t *testing.T) {
		scheduler := NewScheduler(nil, "secret", "http://localhost", nil)
		execID := uuid.New()

		levels := []string{"debug", "info", "warn", "error"}
		for _, level := range levels {
			// Should not panic for any level
			scheduler.handleLogMessage(execID, level, "test message")
		}
	})
}

// =============================================================================
// Cron Parser Tests
// =============================================================================

func TestCronParser(t *testing.T) {
	// Create parser matching the one used in NewScheduler
	parser := cron.NewParser(
		cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)

	t.Run("parses standard 5-field cron expressions", func(t *testing.T) {
		expressions := []struct {
			expr        string
			description string
		}{
			{"* * * * *", "every minute"},
			{"*/5 * * * *", "every 5 minutes"},
			{"0 * * * *", "every hour at minute 0"},
			{"0 0 * * *", "every day at midnight"},
			{"0 12 * * *", "every day at noon"},
			{"0 0 * * 0", "every Sunday at midnight"},
			{"0 0 1 * *", "first of every month"},
			{"0 0 1 1 *", "January 1st"},
			{"30 4 1,15 * *", "1st and 15th at 4:30"},
			{"0 22 * * 1-5", "weekdays at 10pm"},
		}

		for _, tc := range expressions {
			t.Run(tc.description, func(t *testing.T) {
				schedule, err := parser.Parse(tc.expr)
				require.NoError(t, err, "Failed to parse: %s", tc.expr)
				assert.NotNil(t, schedule)
			})
		}
	})

	t.Run("parses 6-field cron expressions with seconds", func(t *testing.T) {
		expressions := []struct {
			expr        string
			description string
		}{
			{"0 * * * * *", "every minute at second 0"},
			{"30 * * * * *", "every minute at second 30"},
			{"0 */5 * * * *", "every 5 minutes at second 0"},
			{"*/10 * * * * *", "every 10 seconds"},
			{"0 0 * * * *", "every hour at minute 0, second 0"},
		}

		for _, tc := range expressions {
			t.Run(tc.description, func(t *testing.T) {
				schedule, err := parser.Parse(tc.expr)
				require.NoError(t, err, "Failed to parse: %s", tc.expr)
				assert.NotNil(t, schedule)
			})
		}
	})

	t.Run("parses descriptors", func(t *testing.T) {
		descriptors := []string{
			"@yearly",
			"@annually",
			"@monthly",
			"@weekly",
			"@daily",
			"@midnight",
			"@hourly",
		}

		for _, desc := range descriptors {
			t.Run(desc, func(t *testing.T) {
				schedule, err := parser.Parse(desc)
				require.NoError(t, err, "Failed to parse: %s", desc)
				assert.NotNil(t, schedule)
			})
		}
	})

	t.Run("rejects invalid expressions", func(t *testing.T) {
		invalidExprs := []string{
			"invalid",
			"* * *",              // too few fields
			"* * * * * * *",      // too many fields
			"60 * * * *",         // invalid minute
			"* 25 * * *",         // invalid hour
			"* * 32 * *",         // invalid day
			"* * * 13 *",         // invalid month
			"* * * * 8",          // invalid day of week
		}

		for _, expr := range invalidExprs {
			t.Run(expr, func(t *testing.T) {
				_, err := parser.Parse(expr)
				assert.Error(t, err, "Should reject: %s", expr)
			})
		}
	})
}

// =============================================================================
// Schedule Calculation Tests
// =============================================================================

func TestScheduleCalculation(t *testing.T) {
	parser := cron.NewParser(
		cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)

	t.Run("every minute schedule", func(t *testing.T) {
		schedule, err := parser.Parse("* * * * *")
		require.NoError(t, err)

		now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		next := schedule.Next(now)

		// Should be next minute
		assert.Equal(t, 2024, next.Year())
		assert.Equal(t, time.January, next.Month())
		assert.Equal(t, 15, next.Day())
		assert.Equal(t, 10, next.Hour())
		assert.Equal(t, 31, next.Minute())
	})

	t.Run("every 5 minutes schedule", func(t *testing.T) {
		schedule, err := parser.Parse("*/5 * * * *")
		require.NoError(t, err)

		now := time.Date(2024, 1, 15, 10, 32, 0, 0, time.UTC)
		next := schedule.Next(now)

		// Should be at minute 35
		assert.Equal(t, 35, next.Minute())
	})

	t.Run("daily at midnight schedule", func(t *testing.T) {
		schedule, err := parser.Parse("0 0 * * *")
		require.NoError(t, err)

		now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		next := schedule.Next(now)

		// Should be next day at midnight
		assert.Equal(t, 16, next.Day())
		assert.Equal(t, 0, next.Hour())
		assert.Equal(t, 0, next.Minute())
	})

	t.Run("weekly schedule", func(t *testing.T) {
		schedule, err := parser.Parse("@weekly")
		require.NoError(t, err)

		// Monday Jan 15, 2024
		now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		next := schedule.Next(now)

		// Should be Sunday
		assert.Equal(t, time.Sunday, next.Weekday())
	})
}

// =============================================================================
// Concurrent Execution Limits Tests
// =============================================================================

func TestConcurrentExecutionLimits(t *testing.T) {
	t.Run("default max concurrent is 10", func(t *testing.T) {
		scheduler := NewScheduler(nil, "secret", "http://localhost", nil)
		assert.Equal(t, 10, scheduler.maxConcurrent)
	})

	t.Run("active count starts at 0", func(t *testing.T) {
		scheduler := NewScheduler(nil, "secret", "http://localhost", nil)
		assert.Equal(t, 0, scheduler.activeCount)
	})
}

// =============================================================================
// Function Jobs Map Tests
// =============================================================================

func TestFunctionJobsMap(t *testing.T) {
	t.Run("empty on initialization", func(t *testing.T) {
		scheduler := NewScheduler(nil, "secret", "http://localhost", nil)
		assert.Empty(t, scheduler.functionJobs)
	})

	t.Run("can store and retrieve job IDs", func(t *testing.T) {
		scheduler := NewScheduler(nil, "secret", "http://localhost", nil)

		// Simulate storing job entries
		scheduler.jobsMu.Lock()
		scheduler.functionJobs["test-function"] = cron.EntryID(1)
		scheduler.functionJobs["another-function"] = cron.EntryID(2)
		scheduler.jobsMu.Unlock()

		scheduler.jobsMu.RLock()
		defer scheduler.jobsMu.RUnlock()

		assert.Equal(t, cron.EntryID(1), scheduler.functionJobs["test-function"])
		assert.Equal(t, cron.EntryID(2), scheduler.functionJobs["another-function"])
	})

	t.Run("can check if function is scheduled", func(t *testing.T) {
		scheduler := NewScheduler(nil, "secret", "http://localhost", nil)

		scheduler.jobsMu.Lock()
		scheduler.functionJobs["scheduled-fn"] = cron.EntryID(1)
		scheduler.jobsMu.Unlock()

		scheduler.jobsMu.RLock()
		_, exists := scheduler.functionJobs["scheduled-fn"]
		_, notExists := scheduler.functionJobs["unscheduled-fn"]
		scheduler.jobsMu.RUnlock()

		assert.True(t, exists)
		assert.False(t, notExists)
	})

	t.Run("can remove scheduled function", func(t *testing.T) {
		scheduler := NewScheduler(nil, "secret", "http://localhost", nil)

		scheduler.jobsMu.Lock()
		scheduler.functionJobs["to-remove"] = cron.EntryID(1)
		scheduler.jobsMu.Unlock()

		// Verify it exists
		scheduler.jobsMu.RLock()
		_, exists := scheduler.functionJobs["to-remove"]
		scheduler.jobsMu.RUnlock()
		assert.True(t, exists)

		// Remove it
		scheduler.jobsMu.Lock()
		delete(scheduler.functionJobs, "to-remove")
		scheduler.jobsMu.Unlock()

		// Verify it's gone
		scheduler.jobsMu.RLock()
		_, stillExists := scheduler.functionJobs["to-remove"]
		scheduler.jobsMu.RUnlock()
		assert.False(t, stillExists)
	})
}

// =============================================================================
// Stop Tests
// =============================================================================

func TestScheduler_Stop(t *testing.T) {
	t.Run("stop cancels context", func(t *testing.T) {
		scheduler := NewScheduler(nil, "secret", "http://localhost", nil)

		// Start and then stop
		scheduler.cron.Start()
		scheduler.Stop()

		// Context should be cancelled
		select {
		case <-scheduler.ctx.Done():
			// Expected
		default:
			t.Error("Context should be cancelled after Stop()")
		}
	})
}

// =============================================================================
// Edge Function Scheduling Tests
// =============================================================================

func TestEdgeFunctionForScheduling(t *testing.T) {
	t.Run("function with cron schedule", func(t *testing.T) {
		schedule := "*/5 * * * *"
		fn := EdgeFunction{
			ID:           uuid.New(),
			Name:         "scheduled-function",
			Code:         "export default () => console.log('scheduled');",
			Enabled:      true,
			CronSchedule: &schedule,
		}

		assert.True(t, fn.Enabled)
		assert.NotNil(t, fn.CronSchedule)
		assert.Equal(t, "*/5 * * * *", *fn.CronSchedule)
	})

	t.Run("disabled function with schedule", func(t *testing.T) {
		schedule := "0 * * * *"
		fn := EdgeFunction{
			ID:           uuid.New(),
			Name:         "disabled-scheduled",
			Code:         "export default () => {};",
			Enabled:      false,
			CronSchedule: &schedule,
		}

		assert.False(t, fn.Enabled)
		assert.NotNil(t, fn.CronSchedule)
	})

	t.Run("enabled function without schedule", func(t *testing.T) {
		fn := EdgeFunction{
			ID:      uuid.New(),
			Name:    "http-only",
			Code:    "export default () => {};",
			Enabled: true,
		}

		assert.True(t, fn.Enabled)
		assert.Nil(t, fn.CronSchedule)
	})
}

// =============================================================================
// Scheduler ScheduleFunction Validation Tests
// =============================================================================

func TestScheduleFunction_Validation(t *testing.T) {
	t.Run("valid schedule", func(t *testing.T) {
		scheduler := NewScheduler(nil, "secret", "http://localhost", nil)
		schedule := "*/5 * * * *"
		fn := EdgeFunction{
			ID:           uuid.New(),
			Name:         "valid-scheduled",
			Code:         "export default () => {};",
			Enabled:      true,
			CronSchedule: &schedule,
		}

		// This will fail because storage is nil, but validates the schedule parsing
		err := scheduler.ScheduleFunction(fn)
		// We expect it to succeed in parsing but fail in execution
		// The function schedules successfully if it gets past parsing
		assert.NoError(t, err)
	})

	t.Run("invalid schedule expression", func(t *testing.T) {
		scheduler := NewScheduler(nil, "secret", "http://localhost", nil)
		invalidSchedule := "invalid cron"
		fn := EdgeFunction{
			ID:           uuid.New(),
			Name:         "invalid-scheduled",
			Code:         "export default () => {};",
			Enabled:      true,
			CronSchedule: &invalidSchedule,
		}

		err := scheduler.ScheduleFunction(fn)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid cron schedule")
	})

	t.Run("nil schedule", func(t *testing.T) {
		scheduler := NewScheduler(nil, "secret", "http://localhost", nil)
		fn := EdgeFunction{
			ID:      uuid.New(),
			Name:    "no-schedule",
			Code:    "export default () => {};",
			Enabled: true,
		}

		err := scheduler.ScheduleFunction(fn)
		assert.NoError(t, err) // Should handle nil schedule gracefully
	})

	t.Run("empty schedule string", func(t *testing.T) {
		scheduler := NewScheduler(nil, "secret", "http://localhost", nil)
		emptySchedule := ""
		fn := EdgeFunction{
			ID:           uuid.New(),
			Name:         "empty-schedule",
			Code:         "export default () => {};",
			Enabled:      true,
			CronSchedule: &emptySchedule,
		}

		err := scheduler.ScheduleFunction(fn)
		assert.NoError(t, err) // Should handle empty schedule gracefully
	})
}

// =============================================================================
// Scheduler UnscheduleFunction Tests
// =============================================================================

func TestUnscheduleFunction(t *testing.T) {
	t.Run("unschedule existing function", func(t *testing.T) {
		scheduler := NewScheduler(nil, "secret", "http://localhost", nil)

		// First schedule a function
		schedule := "*/5 * * * *"
		fn := EdgeFunction{
			ID:           uuid.New(),
			Name:         "to-unschedule",
			Code:         "export default () => {};",
			Enabled:      true,
			CronSchedule: &schedule,
		}

		err := scheduler.ScheduleFunction(fn)
		require.NoError(t, err)

		// Verify it's scheduled
		scheduler.jobsMu.RLock()
		_, exists := scheduler.functionJobs[fn.Name]
		scheduler.jobsMu.RUnlock()
		assert.True(t, exists)

		// Unschedule it
		scheduler.UnscheduleFunction(fn.Name)

		// Verify it's removed
		scheduler.jobsMu.RLock()
		_, stillExists := scheduler.functionJobs[fn.Name]
		scheduler.jobsMu.RUnlock()
		assert.False(t, stillExists)
	})

	t.Run("unschedule non-existent function", func(t *testing.T) {
		scheduler := NewScheduler(nil, "secret", "http://localhost", nil)

		// Should not panic
		scheduler.UnscheduleFunction("non-existent")
	})
}

// =============================================================================
// Scheduler IsScheduled Tests
// =============================================================================

func TestIsScheduled(t *testing.T) {
	t.Run("returns true for scheduled function", func(t *testing.T) {
		scheduler := NewScheduler(nil, "secret", "http://localhost", nil)

		schedule := "*/5 * * * *"
		fn := EdgeFunction{
			ID:           uuid.New(),
			Name:         "scheduled-check",
			Code:         "export default () => {};",
			Enabled:      true,
			CronSchedule: &schedule,
		}

		err := scheduler.ScheduleFunction(fn)
		require.NoError(t, err)

		assert.True(t, scheduler.IsScheduled(fn.Name))
	})

	t.Run("returns false for unscheduled function", func(t *testing.T) {
		scheduler := NewScheduler(nil, "secret", "http://localhost", nil)
		assert.False(t, scheduler.IsScheduled("not-scheduled"))
	})
}

// =============================================================================
// Scheduler GetScheduledFunctions Tests
// =============================================================================

func TestGetScheduledFunctions(t *testing.T) {
	t.Run("returns empty list initially", func(t *testing.T) {
		scheduler := NewScheduler(nil, "secret", "http://localhost", nil)
		functions := scheduler.GetScheduledFunctions()
		assert.Empty(t, functions)
	})

	t.Run("returns scheduled function names", func(t *testing.T) {
		scheduler := NewScheduler(nil, "secret", "http://localhost", nil)

		schedules := []struct {
			name string
			cron string
		}{
			{"func-1", "*/5 * * * *"},
			{"func-2", "0 * * * *"},
			{"func-3", "0 0 * * *"},
		}

		for _, s := range schedules {
			fn := EdgeFunction{
				ID:           uuid.New(),
				Name:         s.name,
				Code:         "export default () => {};",
				Enabled:      true,
				CronSchedule: &s.cron,
			}
			err := scheduler.ScheduleFunction(fn)
			require.NoError(t, err)
		}

		functions := scheduler.GetScheduledFunctions()
		assert.Len(t, functions, 3)
		assert.Contains(t, functions, "func-1")
		assert.Contains(t, functions, "func-2")
		assert.Contains(t, functions, "func-3")
	})
}
