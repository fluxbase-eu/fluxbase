package jobs

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Storage Construction Tests
// =============================================================================

func TestNewStorage(t *testing.T) {
	t.Run("creates storage with nil database", func(t *testing.T) {
		storage := NewStorage(nil)
		assert.NotNil(t, storage)
		assert.Nil(t, storage.db)
	})
}

// =============================================================================
// Job Struct Tests
// =============================================================================

func TestJob_Struct(t *testing.T) {
	t.Run("basic job", func(t *testing.T) {
		job := Job{
			ID:           uuid.New(),
			Namespace:    "default",
			FunctionName: "process-data",
			Status:       JobStatusPending,
			Priority:     JobPriorityNormal,
			CreatedAt:    time.Now(),
		}

		assert.NotEqual(t, uuid.Nil, job.ID)
		assert.Equal(t, "default", job.Namespace)
		assert.Equal(t, "process-data", job.FunctionName)
		assert.Equal(t, JobStatusPending, job.Status)
		assert.Equal(t, JobPriorityNormal, job.Priority)
	})

	t.Run("job with all fields", func(t *testing.T) {
		workerID := uuid.New()
		startedAt := time.Now()
		completedAt := startedAt.Add(5 * time.Second)
		retryCount := 2
		maxRetries := 3
		params := `{"batch_size": 100}`
		result := `{"processed": 100}`

		job := Job{
			ID:           uuid.New(),
			Namespace:    "production",
			FunctionName: "batch-processor",
			Status:       JobStatusCompleted,
			Priority:     JobPriorityHigh,
			Params:       &params,
			Result:       &result,
			WorkerID:     &workerID,
			StartedAt:    &startedAt,
			CompletedAt:  &completedAt,
			RetryCount:   &retryCount,
			MaxRetries:   &maxRetries,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		assert.Equal(t, "production", job.Namespace)
		assert.Equal(t, JobStatusCompleted, job.Status)
		assert.Equal(t, JobPriorityHigh, job.Priority)
		assert.NotNil(t, job.Params)
		assert.NotNil(t, job.Result)
		assert.NotNil(t, job.WorkerID)
		assert.NotNil(t, job.StartedAt)
		assert.NotNil(t, job.CompletedAt)
		assert.Equal(t, 2, *job.RetryCount)
		assert.Equal(t, 3, *job.MaxRetries)
	})

	t.Run("job with error", func(t *testing.T) {
		errorMsg := "Connection timeout"
		errorStack := "Error: Connection timeout\n    at connect (db.ts:45)"

		job := Job{
			ID:           uuid.New(),
			FunctionName: "failing-job",
			Status:       JobStatusFailed,
			ErrorMessage: &errorMsg,
			ErrorStack:   &errorStack,
		}

		assert.Equal(t, JobStatusFailed, job.Status)
		assert.Equal(t, "Connection timeout", *job.ErrorMessage)
		assert.Contains(t, *job.ErrorStack, "connect")
	})
}

// =============================================================================
// Job Status Constants Tests
// =============================================================================

func TestJobStatus_Constants(t *testing.T) {
	statuses := []struct {
		status   JobStatus
		expected string
	}{
		{JobStatusPending, "pending"},
		{JobStatusQueued, "queued"},
		{JobStatusRunning, "running"},
		{JobStatusCompleted, "completed"},
		{JobStatusFailed, "failed"},
		{JobStatusCancelled, "cancelled"},
		{JobStatusTimeout, "timeout"},
	}

	for _, tc := range statuses {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, JobStatus(tc.expected), tc.status)
		})
	}
}

// =============================================================================
// Job Priority Constants Tests
// =============================================================================

func TestJobPriority_Constants(t *testing.T) {
	priorities := []struct {
		priority JobPriority
		expected int
	}{
		{JobPriorityLow, 1},
		{JobPriorityNormal, 5},
		{JobPriorityHigh, 10},
		{JobPriorityCritical, 20},
	}

	for _, tc := range priorities {
		t.Run(string(rune(tc.expected)), func(t *testing.T) {
			assert.Equal(t, JobPriority(tc.expected), tc.priority)
		})
	}
}

// =============================================================================
// Job JSON Serialization Tests
// =============================================================================

func TestJob_JSONSerialization(t *testing.T) {
	t.Run("basic job serialization", func(t *testing.T) {
		job := Job{
			ID:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			Namespace:    "default",
			FunctionName: "test-job",
			Status:       JobStatusPending,
			Priority:     JobPriorityNormal,
			CreatedAt:    time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		}

		data, err := json.Marshal(job)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"namespace":"default"`)
		assert.Contains(t, string(data), `"function_name":"test-job"`)
		assert.Contains(t, string(data), `"status":"pending"`)
	})

	t.Run("job deserialization", func(t *testing.T) {
		jsonData := `{
			"id": "550e8400-e29b-41d4-a716-446655440000",
			"namespace": "test",
			"function_name": "deserialize-test",
			"status": "running",
			"priority": 10
		}`

		var job Job
		err := json.Unmarshal([]byte(jsonData), &job)
		require.NoError(t, err)

		assert.Equal(t, "test", job.Namespace)
		assert.Equal(t, "deserialize-test", job.FunctionName)
		assert.Equal(t, JobStatusRunning, job.Status)
		assert.Equal(t, JobPriorityHigh, job.Priority)
	})
}

// =============================================================================
// JobFunction Struct Tests
// =============================================================================

func TestJobFunction_Struct(t *testing.T) {
	t.Run("basic job function", func(t *testing.T) {
		fn := JobFunction{
			ID:        uuid.New(),
			Namespace: "default",
			Name:      "data-processor",
			Code:      "export default async () => { console.log('processing'); }",
			Enabled:   true,
		}

		assert.NotEqual(t, uuid.Nil, fn.ID)
		assert.Equal(t, "default", fn.Namespace)
		assert.Equal(t, "data-processor", fn.Name)
		assert.True(t, fn.Enabled)
	})

	t.Run("job function with schedule", func(t *testing.T) {
		schedule := `{"cron_expression": "*/5 * * * *"}`

		fn := JobFunction{
			ID:             uuid.New(),
			Name:           "scheduled-processor",
			Code:           "export default async () => {}",
			Enabled:        true,
			ScheduleConfig: &schedule,
		}

		assert.NotNil(t, fn.ScheduleConfig)
		assert.Contains(t, *fn.ScheduleConfig, "cron_expression")
	})

	t.Run("job function with all fields", func(t *testing.T) {
		description := "Processes batch data every hour"
		schedule := `{"cron_expression": "0 * * * *"}`
		creatorID := uuid.New()
		maxDuration := 30 * time.Minute
		maxRetries := 3
		retryDelay := time.Minute

		fn := JobFunction{
			ID:                 uuid.New(),
			Namespace:          "production",
			Name:               "hourly-batch",
			Description:        &description,
			Code:               "export default async (params) => { ... }",
			Enabled:            true,
			ScheduleConfig:     &schedule,
			TimeoutSeconds:     int(maxDuration.Seconds()),
			MaxRetries:         maxRetries,
			RetryDelaySeconds:  int(retryDelay.Seconds()),
			DisableProgressLog: false,
			CreatedBy:          &creatorID,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		assert.Equal(t, "production", fn.Namespace)
		assert.Equal(t, "Processes batch data every hour", *fn.Description)
		assert.Equal(t, 1800, fn.TimeoutSeconds)
		assert.Equal(t, 3, fn.MaxRetries)
		assert.Equal(t, 60, fn.RetryDelaySeconds)
		assert.False(t, fn.DisableProgressLog)
	})
}

// =============================================================================
// JobProgress Struct Tests
// =============================================================================

func TestJobProgress_Struct(t *testing.T) {
	t.Run("basic progress", func(t *testing.T) {
		progress := JobProgress{
			JobID:      uuid.New(),
			Progress:   50,
			Total:      100,
			Message:    "Processing items",
			RecordedAt: time.Now(),
		}

		assert.Equal(t, 50, progress.Progress)
		assert.Equal(t, 100, progress.Total)
		assert.Equal(t, "Processing items", progress.Message)
	})

	t.Run("progress with ETA", func(t *testing.T) {
		eta := 5 * time.Minute

		progress := JobProgress{
			JobID:           uuid.New(),
			Progress:        75,
			Total:           100,
			Message:         "75% complete",
			EstimatedTimeMs: int64(eta.Milliseconds()),
		}

		assert.Equal(t, 75, progress.Progress)
		assert.Equal(t, int64(300000), progress.EstimatedTimeMs)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		progress := JobProgress{
			JobID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			Progress: 30,
			Total:    100,
			Message:  "In progress",
		}

		data, err := json.Marshal(progress)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"progress":30`)
		assert.Contains(t, string(data), `"total":100`)
		assert.Contains(t, string(data), `"message":"In progress"`)
	})
}

// =============================================================================
// JobLog Struct Tests
// =============================================================================

func TestJobLog_Struct(t *testing.T) {
	t.Run("info log", func(t *testing.T) {
		log := JobLog{
			JobID:     uuid.New(),
			Level:     "info",
			Message:   "Job started",
			Timestamp: time.Now(),
		}

		assert.Equal(t, "info", log.Level)
		assert.Equal(t, "Job started", log.Message)
	})

	t.Run("error log", func(t *testing.T) {
		log := JobLog{
			JobID:     uuid.New(),
			Level:     "error",
			Message:   "Failed to connect to database",
			Timestamp: time.Now(),
		}

		assert.Equal(t, "error", log.Level)
	})

	t.Run("log with line number", func(t *testing.T) {
		lineNum := 42

		log := JobLog{
			JobID:      uuid.New(),
			Level:      "debug",
			Message:    "Processing item",
			LineNumber: &lineNum,
		}

		assert.Equal(t, 42, *log.LineNumber)
	})
}

// =============================================================================
// Job Retry Logic Tests
// =============================================================================

func TestJob_RetryLogic(t *testing.T) {
	t.Run("job with no retries", func(t *testing.T) {
		job := Job{
			ID:           uuid.New(),
			FunctionName: "no-retry",
			Status:       JobStatusFailed,
		}

		assert.Nil(t, job.RetryCount)
		assert.Nil(t, job.MaxRetries)
	})

	t.Run("job with retry configuration", func(t *testing.T) {
		retryCount := 0
		maxRetries := 3

		job := Job{
			ID:           uuid.New(),
			FunctionName: "retry-job",
			Status:       JobStatusPending,
			RetryCount:   &retryCount,
			MaxRetries:   &maxRetries,
		}

		assert.Equal(t, 0, *job.RetryCount)
		assert.Equal(t, 3, *job.MaxRetries)
	})

	t.Run("job after retry", func(t *testing.T) {
		retryCount := 2
		maxRetries := 3

		job := Job{
			ID:           uuid.New(),
			FunctionName: "retrying-job",
			Status:       JobStatusQueued,
			RetryCount:   &retryCount,
			MaxRetries:   &maxRetries,
		}

		// Check if more retries available
		hasMoreRetries := *job.RetryCount < *job.MaxRetries
		assert.True(t, hasMoreRetries)
	})

	t.Run("job exhausted retries", func(t *testing.T) {
		retryCount := 3
		maxRetries := 3
		errorMsg := "Max retries exceeded"

		job := Job{
			ID:           uuid.New(),
			FunctionName: "exhausted-retries",
			Status:       JobStatusFailed,
			RetryCount:   &retryCount,
			MaxRetries:   &maxRetries,
			ErrorMessage: &errorMsg,
		}

		// Check if more retries available
		hasMoreRetries := *job.RetryCount < *job.MaxRetries
		assert.False(t, hasMoreRetries)
		assert.Equal(t, JobStatusFailed, job.Status)
	})
}

// =============================================================================
// Job Scheduling Tests
// =============================================================================

func TestJob_Scheduling(t *testing.T) {
	t.Run("immediate job (no schedule)", func(t *testing.T) {
		job := Job{
			ID:           uuid.New(),
			FunctionName: "immediate-job",
			Status:       JobStatusPending,
		}

		assert.Nil(t, job.ScheduledFor)
	})

	t.Run("scheduled job", func(t *testing.T) {
		scheduledFor := time.Now().Add(time.Hour)

		job := Job{
			ID:           uuid.New(),
			FunctionName: "scheduled-job",
			Status:       JobStatusPending,
			ScheduledFor: &scheduledFor,
		}

		assert.NotNil(t, job.ScheduledFor)
		assert.True(t, job.ScheduledFor.After(time.Now()))
	})
}

// =============================================================================
// Job Duration Tests
// =============================================================================

func TestJob_Duration(t *testing.T) {
	t.Run("calculate job duration", func(t *testing.T) {
		startedAt := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
		completedAt := time.Date(2024, 1, 15, 10, 5, 30, 0, time.UTC)

		job := Job{
			ID:          uuid.New(),
			Status:      JobStatusCompleted,
			StartedAt:   &startedAt,
			CompletedAt: &completedAt,
		}

		duration := job.CompletedAt.Sub(*job.StartedAt)
		assert.Equal(t, 5*time.Minute+30*time.Second, duration)
	})

	t.Run("running job (no completion time)", func(t *testing.T) {
		startedAt := time.Now().Add(-2 * time.Minute)

		job := Job{
			ID:        uuid.New(),
			Status:    JobStatusRunning,
			StartedAt: &startedAt,
		}

		assert.NotNil(t, job.StartedAt)
		assert.Nil(t, job.CompletedAt)
	})
}

// =============================================================================
// Job Namespace Tests
// =============================================================================

func TestJob_Namespace(t *testing.T) {
	t.Run("default namespace", func(t *testing.T) {
		job := Job{
			ID:           uuid.New(),
			Namespace:    "default",
			FunctionName: "test",
		}

		assert.Equal(t, "default", job.Namespace)
	})

	t.Run("custom namespace", func(t *testing.T) {
		job := Job{
			ID:           uuid.New(),
			Namespace:    "production",
			FunctionName: "test",
		}

		assert.Equal(t, "production", job.Namespace)
	})

	t.Run("empty namespace", func(t *testing.T) {
		job := Job{
			ID:           uuid.New(),
			Namespace:    "",
			FunctionName: "test",
		}

		assert.Empty(t, job.Namespace)
	})
}

// =============================================================================
// JobFunction Timeout Tests
// =============================================================================

func TestJobFunction_Timeout(t *testing.T) {
	t.Run("default timeout", func(t *testing.T) {
		fn := JobFunction{
			Name:           "default-timeout",
			TimeoutSeconds: 1800, // 30 minutes
		}

		assert.Equal(t, 1800, fn.TimeoutSeconds)
	})

	t.Run("custom timeout", func(t *testing.T) {
		fn := JobFunction{
			Name:           "quick-job",
			TimeoutSeconds: 60, // 1 minute
		}

		assert.Equal(t, 60, fn.TimeoutSeconds)
	})

	t.Run("long running job timeout", func(t *testing.T) {
		fn := JobFunction{
			Name:           "long-job",
			TimeoutSeconds: 86400, // 24 hours
		}

		assert.Equal(t, 86400, fn.TimeoutSeconds)
	})
}
