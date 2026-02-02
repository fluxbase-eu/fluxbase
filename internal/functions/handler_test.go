package functions

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
// Handler Construction Tests
// =============================================================================

func TestNewHandler(t *testing.T) {
	t.Run("creates handler with nil dependencies", func(t *testing.T) {
		handler := NewHandler(nil, "/tmp/functions", config.CORSConfig{}, "secret", "http://localhost", "", "", nil, nil, nil)
		assert.NotNil(t, handler)
		assert.Equal(t, "/tmp/functions", handler.functionsDir)
		assert.Equal(t, "http://localhost", handler.publicURL)
	})

	t.Run("creates handler with custom registries", func(t *testing.T) {
		handler := NewHandler(nil, "/tmp/functions", config.CORSConfig{}, "secret", "http://localhost", "https://npm.example.com", "https://jsr.example.com", nil, nil, nil)
		assert.NotNil(t, handler)
		assert.Equal(t, "https://npm.example.com", handler.npmRegistry)
		assert.Equal(t, "https://jsr.example.com", handler.jsrRegistry)
	})
}

func TestHandler_SetScheduler(t *testing.T) {
	t.Run("sets scheduler reference", func(t *testing.T) {
		handler := NewHandler(nil, "/tmp/functions", config.CORSConfig{}, "secret", "http://localhost", "", "", nil, nil, nil)
		assert.Nil(t, handler.scheduler)

		scheduler := &Scheduler{}
		handler.SetScheduler(scheduler)
		assert.NotNil(t, handler.scheduler)
		assert.Equal(t, scheduler, handler.scheduler)
	})
}

func TestHandler_GetRuntime(t *testing.T) {
	t.Run("returns runtime instance", func(t *testing.T) {
		handler := NewHandler(nil, "/tmp/functions", config.CORSConfig{}, "secret", "http://localhost", "", "", nil, nil, nil)
		runtime := handler.GetRuntime()
		assert.NotNil(t, runtime)
	})
}

func TestHandler_GetPublicURL(t *testing.T) {
	t.Run("returns configured public URL", func(t *testing.T) {
		handler := NewHandler(nil, "/tmp/functions", config.CORSConfig{}, "secret", "https://api.example.com", "", "", nil, nil, nil)
		assert.Equal(t, "https://api.example.com", handler.GetPublicURL())
	})
}

func TestHandler_GetFunctionsDir(t *testing.T) {
	t.Run("returns configured functions directory", func(t *testing.T) {
		handler := NewHandler(nil, "/custom/path", config.CORSConfig{}, "secret", "http://localhost", "", "", nil, nil, nil)
		assert.Equal(t, "/custom/path", handler.GetFunctionsDir())
	})
}

func TestHandler_createBundler(t *testing.T) {
	t.Run("creates bundler without custom registries", func(t *testing.T) {
		handler := NewHandler(nil, "/tmp/functions", config.CORSConfig{}, "secret", "http://localhost", "", "", nil, nil, nil)
		bundler, err := handler.createBundler()
		require.NoError(t, err)
		assert.NotNil(t, bundler)
	})

	t.Run("creates bundler with npm registry", func(t *testing.T) {
		handler := NewHandler(nil, "/tmp/functions", config.CORSConfig{}, "secret", "http://localhost", "https://npm.example.com", "", nil, nil, nil)
		bundler, err := handler.createBundler()
		require.NoError(t, err)
		assert.NotNil(t, bundler)
	})

	t.Run("creates bundler with jsr registry", func(t *testing.T) {
		handler := NewHandler(nil, "/tmp/functions", config.CORSConfig{}, "secret", "http://localhost", "", "https://jsr.example.com", nil, nil, nil)
		bundler, err := handler.createBundler()
		require.NoError(t, err)
		assert.NotNil(t, bundler)
	})

	t.Run("creates bundler with both registries", func(t *testing.T) {
		handler := NewHandler(nil, "/tmp/functions", config.CORSConfig{}, "secret", "http://localhost", "https://npm.example.com", "https://jsr.example.com", nil, nil, nil)
		bundler, err := handler.createBundler()
		require.NoError(t, err)
		assert.NotNil(t, bundler)
	})
}

// =============================================================================
// Log Message Handling Tests
// =============================================================================

func TestHandler_handleLogMessage(t *testing.T) {
	t.Run("handles log without counter", func(t *testing.T) {
		handler := NewHandler(nil, "/tmp/functions", config.CORSConfig{}, "secret", "http://localhost", "", "", nil, nil, nil)
		execID := uuid.New()

		// Should not panic when no counter exists
		handler.handleLogMessage(execID, "info", "test message")
	})

	t.Run("increments counter when exists", func(t *testing.T) {
		handler := NewHandler(nil, "/tmp/functions", config.CORSConfig{}, "secret", "http://localhost", "", "", nil, nil, nil)
		execID := uuid.New()

		// Set up counter
		counter := 0
		handler.logCounters.Store(execID, &counter)

		handler.handleLogMessage(execID, "info", "message 1")
		assert.Equal(t, 1, counter)

		handler.handleLogMessage(execID, "info", "message 2")
		assert.Equal(t, 2, counter)
	})

	t.Run("handles invalid counter type", func(t *testing.T) {
		handler := NewHandler(nil, "/tmp/functions", config.CORSConfig{}, "secret", "http://localhost", "", "", nil, nil, nil)
		execID := uuid.New()

		// Store invalid type
		handler.logCounters.Store(execID, "not a pointer")

		// Should not panic
		handler.handleLogMessage(execID, "info", "test message")
	})
}

// =============================================================================
// EdgeFunction Struct Tests
// =============================================================================

func TestEdgeFunction_Struct(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		fn := EdgeFunction{
			Name: "test-function",
			Code: "export default () => 'hello'",
		}

		assert.Equal(t, "test-function", fn.Name)
		assert.Equal(t, "export default () => 'hello'", fn.Code)
		assert.False(t, fn.Enabled)
		assert.False(t, fn.AllowNet)
		assert.False(t, fn.AllowEnv)
		assert.False(t, fn.AllowUnauthenticated)
	})

	t.Run("with all fields", func(t *testing.T) {
		description := "A test function"
		cronSchedule := "*/5 * * * *"
		corsOrigins := "*"
		rateLimit := 100

		fn := EdgeFunction{
			ID:                   uuid.New(),
			Name:                 "complete-function",
			Namespace:            "default",
			Description:          &description,
			Code:                 "export default () => 'hello'",
			Version:              1,
			CronSchedule:         &cronSchedule,
			Enabled:              true,
			TimeoutSeconds:       30,
			MemoryLimitMB:        128,
			AllowNet:             true,
			AllowEnv:             true,
			AllowRead:            true,
			AllowWrite:           false,
			AllowUnauthenticated: false,
			IsPublic:             true,
			CorsOrigins:          &corsOrigins,
			RateLimitPerMinute:   &rateLimit,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
			Source:               "api",
		}

		assert.NotEqual(t, uuid.Nil, fn.ID)
		assert.Equal(t, "complete-function", fn.Name)
		assert.Equal(t, "default", fn.Namespace)
		assert.Equal(t, "A test function", *fn.Description)
		assert.Equal(t, "*/5 * * * *", *fn.CronSchedule)
		assert.True(t, fn.Enabled)
		assert.Equal(t, 30, fn.TimeoutSeconds)
		assert.Equal(t, 128, fn.MemoryLimitMB)
		assert.True(t, fn.AllowNet)
		assert.True(t, fn.AllowEnv)
		assert.Equal(t, 100, *fn.RateLimitPerMinute)
		assert.Equal(t, "api", fn.Source)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		fn := EdgeFunction{
			ID:             uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			Name:           "json-test",
			Code:           "export default () => 'hello'",
			Enabled:        true,
			TimeoutSeconds: 30,
			MemoryLimitMB:  128,
		}

		data, err := json.Marshal(fn)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"name":"json-test"`)
		assert.Contains(t, string(data), `"enabled":true`)
		assert.Contains(t, string(data), `"timeout_seconds":30`)
		assert.Contains(t, string(data), `"memory_limit_mb":128`)
	})
}

// =============================================================================
// EdgeFunctionSummary Struct Tests
// =============================================================================

func TestEdgeFunctionSummary_Struct(t *testing.T) {
	t.Run("excludes code fields", func(t *testing.T) {
		summary := EdgeFunctionSummary{
			ID:             uuid.New(),
			Name:           "summary-test",
			Enabled:        true,
			TimeoutSeconds: 30,
			Source:         "filesystem",
		}

		data, err := json.Marshal(summary)
		require.NoError(t, err)

		// Should not contain code field
		assert.NotContains(t, string(data), `"code"`)
		assert.Contains(t, string(data), `"name":"summary-test"`)
	})
}

// =============================================================================
// EdgeFunctionExecution Struct Tests
// =============================================================================

func TestEdgeFunctionExecution_Struct(t *testing.T) {
	t.Run("pending execution", func(t *testing.T) {
		exec := EdgeFunctionExecution{
			ID:          uuid.New(),
			FunctionID:  uuid.New(),
			TriggerType: "http",
			Status:      "pending",
			ExecutedAt:  time.Now(),
		}

		assert.Equal(t, "pending", exec.Status)
		assert.Equal(t, "http", exec.TriggerType)
		assert.Nil(t, exec.CompletedAt)
	})

	t.Run("completed execution", func(t *testing.T) {
		now := time.Now()
		duration := 150
		statusCode := 200
		result := `{"message": "success"}`

		exec := EdgeFunctionExecution{
			ID:          uuid.New(),
			FunctionID:  uuid.New(),
			TriggerType: "cron",
			Status:      "success",
			StatusCode:  &statusCode,
			DurationMs:  &duration,
			Result:      &result,
			ExecutedAt:  now,
			CompletedAt: &now,
		}

		assert.Equal(t, "success", exec.Status)
		assert.Equal(t, "cron", exec.TriggerType)
		assert.Equal(t, 200, *exec.StatusCode)
		assert.Equal(t, 150, *exec.DurationMs)
		assert.NotNil(t, exec.CompletedAt)
	})

	t.Run("failed execution", func(t *testing.T) {
		errorMsg := "Function timeout exceeded"
		errorStack := "Error: timeout\n    at execute (function.ts:10)"

		exec := EdgeFunctionExecution{
			ID:           uuid.New(),
			FunctionID:   uuid.New(),
			TriggerType:  "http",
			Status:       "error",
			ErrorMessage: &errorMsg,
			ErrorStack:   &errorStack,
			ExecutedAt:   time.Now(),
		}

		assert.Equal(t, "error", exec.Status)
		assert.Equal(t, "Function timeout exceeded", *exec.ErrorMessage)
		assert.Contains(t, *exec.ErrorStack, "timeout")
	})
}

// =============================================================================
// FunctionFile Struct Tests
// =============================================================================

func TestFunctionFile_Struct(t *testing.T) {
	t.Run("supporting file", func(t *testing.T) {
		file := FunctionFile{
			ID:         uuid.New(),
			FunctionID: uuid.New(),
			FilePath:   "utils/helper.ts",
			Content:    "export const add = (a, b) => a + b;",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		assert.Equal(t, "utils/helper.ts", file.FilePath)
		assert.Contains(t, file.Content, "export const add")
	})

	t.Run("JSON serialization", func(t *testing.T) {
		file := FunctionFile{
			ID:         uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			FunctionID: uuid.MustParse("660e8400-e29b-41d4-a716-446655440000"),
			FilePath:   "db.ts",
			Content:    "export const query = () => {}",
		}

		data, err := json.Marshal(file)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"file_path":"db.ts"`)
		assert.Contains(t, string(data), `"content":"export const query = () => {}"`)
	})
}

// =============================================================================
// SharedModule Struct Tests
// =============================================================================

func TestSharedModule_Struct(t *testing.T) {
	t.Run("basic shared module", func(t *testing.T) {
		module := SharedModule{
			ID:         uuid.New(),
			ModulePath: "_shared/cors.ts",
			Content:    "export const corsHeaders = { ... };",
			Version:    1,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		assert.Equal(t, "_shared/cors.ts", module.ModulePath)
		assert.Equal(t, 1, module.Version)
	})

	t.Run("with description and creator", func(t *testing.T) {
		description := "CORS utilities for edge functions"
		creatorID := uuid.New()

		module := SharedModule{
			ID:          uuid.New(),
			ModulePath:  "_shared/utils/http.ts",
			Content:     "export const parseBody = (req) => { ... };",
			Description: &description,
			Version:     2,
			CreatedBy:   &creatorID,
		}

		assert.Equal(t, "CORS utilities for edge functions", *module.Description)
		assert.Equal(t, 2, module.Version)
		assert.Equal(t, creatorID, *module.CreatedBy)
	})
}

// =============================================================================
// CORS Configuration Tests
// =============================================================================

func TestCORSConfig(t *testing.T) {
	t.Run("default CORS config", func(t *testing.T) {
		corsConfig := config.CORSConfig{
			AllowedOrigins:   "*",
			AllowedMethods:   "GET,POST,PUT,DELETE,OPTIONS",
			AllowedHeaders:   "Content-Type,Authorization",
			AllowCredentials: false,
			MaxAge:           3600,
		}

		assert.Equal(t, "*", corsConfig.AllowedOrigins)
		assert.Contains(t, corsConfig.AllowedMethods, "POST")
		assert.False(t, corsConfig.AllowCredentials)
		assert.Equal(t, 3600, corsConfig.MaxAge)
	})

	t.Run("restrictive CORS config", func(t *testing.T) {
		corsConfig := config.CORSConfig{
			AllowedOrigins:   "https://app.example.com,https://admin.example.com",
			AllowedMethods:   "GET,POST",
			AllowedHeaders:   "Content-Type,Authorization,X-Custom-Header",
			AllowCredentials: true,
			MaxAge:           7200,
			ExposedHeaders:   "X-Request-Id",
		}

		assert.Contains(t, corsConfig.AllowedOrigins, "https://app.example.com")
		assert.True(t, corsConfig.AllowCredentials)
		assert.Equal(t, "X-Request-Id", corsConfig.ExposedHeaders)
	})
}

// =============================================================================
// Rate Limiting Configuration Tests
// =============================================================================

func TestRateLimitConfiguration(t *testing.T) {
	t.Run("per-minute rate limit", func(t *testing.T) {
		limit := 60
		fn := EdgeFunction{
			Name:               "rate-limited",
			RateLimitPerMinute: &limit,
		}

		assert.Equal(t, 60, *fn.RateLimitPerMinute)
		assert.Nil(t, fn.RateLimitPerHour)
		assert.Nil(t, fn.RateLimitPerDay)
	})

	t.Run("multiple rate limits", func(t *testing.T) {
		perMin := 10
		perHour := 100
		perDay := 1000

		fn := EdgeFunction{
			Name:               "multi-rate-limited",
			RateLimitPerMinute: &perMin,
			RateLimitPerHour:   &perHour,
			RateLimitPerDay:    &perDay,
		}

		assert.Equal(t, 10, *fn.RateLimitPerMinute)
		assert.Equal(t, 100, *fn.RateLimitPerHour)
		assert.Equal(t, 1000, *fn.RateLimitPerDay)
	})

	t.Run("no rate limits (unlimited)", func(t *testing.T) {
		fn := EdgeFunction{
			Name: "unlimited",
		}

		assert.Nil(t, fn.RateLimitPerMinute)
		assert.Nil(t, fn.RateLimitPerHour)
		assert.Nil(t, fn.RateLimitPerDay)
	})
}

// =============================================================================
// Cron Schedule Tests
// =============================================================================

func TestCronSchedule(t *testing.T) {
	t.Run("valid cron expressions", func(t *testing.T) {
		schedules := []string{
			"* * * * *",     // Every minute
			"*/5 * * * *",   // Every 5 minutes
			"0 * * * *",     // Every hour
			"0 0 * * *",     // Every day at midnight
			"0 0 * * 0",     // Every Sunday at midnight
			"0 0 1 * *",     // First of every month
			"0 0 1 1 *",     // January 1st
			"0 */5 * * * *", // Every 5 minutes (6-field with seconds)
			"30 0 0 * * *",  // Every day at midnight + 30 seconds
		}

		for _, schedule := range schedules {
			fn := EdgeFunction{
				Name:         "scheduled-fn",
				CronSchedule: &schedule,
			}
			assert.Equal(t, schedule, *fn.CronSchedule)
		}
	})

	t.Run("function without schedule", func(t *testing.T) {
		fn := EdgeFunction{
			Name: "http-only",
		}

		assert.Nil(t, fn.CronSchedule)
	})
}

// =============================================================================
// Function Permissions Tests
// =============================================================================

func TestFunctionPermissions(t *testing.T) {
	t.Run("minimal permissions", func(t *testing.T) {
		fn := EdgeFunction{
			Name:                 "minimal",
			AllowNet:             false,
			AllowEnv:             false,
			AllowRead:            false,
			AllowWrite:           false,
			AllowUnauthenticated: false,
		}

		assert.False(t, fn.AllowNet)
		assert.False(t, fn.AllowEnv)
		assert.False(t, fn.AllowRead)
		assert.False(t, fn.AllowWrite)
		assert.False(t, fn.AllowUnauthenticated)
	})

	t.Run("full permissions", func(t *testing.T) {
		fn := EdgeFunction{
			Name:                 "full-access",
			AllowNet:             true,
			AllowEnv:             true,
			AllowRead:            true,
			AllowWrite:           true,
			AllowUnauthenticated: true,
			IsPublic:             true,
		}

		assert.True(t, fn.AllowNet)
		assert.True(t, fn.AllowEnv)
		assert.True(t, fn.AllowRead)
		assert.True(t, fn.AllowWrite)
		assert.True(t, fn.AllowUnauthenticated)
		assert.True(t, fn.IsPublic)
	})
}

// =============================================================================
// Resource Limits Tests
// =============================================================================

func TestResourceLimits(t *testing.T) {
	t.Run("default limits", func(t *testing.T) {
		fn := EdgeFunction{
			Name:           "default-limits",
			TimeoutSeconds: 30,
			MemoryLimitMB:  128,
		}

		assert.Equal(t, 30, fn.TimeoutSeconds)
		assert.Equal(t, 128, fn.MemoryLimitMB)
	})

	t.Run("custom limits", func(t *testing.T) {
		fn := EdgeFunction{
			Name:           "custom-limits",
			TimeoutSeconds: 300,
			MemoryLimitMB:  512,
		}

		assert.Equal(t, 300, fn.TimeoutSeconds)
		assert.Equal(t, 512, fn.MemoryLimitMB)
	})

	t.Run("minimal limits", func(t *testing.T) {
		fn := EdgeFunction{
			Name:           "minimal-limits",
			TimeoutSeconds: 1,
			MemoryLimitMB:  32,
		}

		assert.Equal(t, 1, fn.TimeoutSeconds)
		assert.Equal(t, 32, fn.MemoryLimitMB)
	})
}

// =============================================================================
// Trigger Types Tests
// =============================================================================

func TestTriggerTypes(t *testing.T) {
	t.Run("HTTP trigger", func(t *testing.T) {
		exec := EdgeFunctionExecution{
			TriggerType: "http",
		}
		assert.Equal(t, "http", exec.TriggerType)
	})

	t.Run("cron trigger", func(t *testing.T) {
		exec := EdgeFunctionExecution{
			TriggerType: "cron",
		}
		assert.Equal(t, "cron", exec.TriggerType)
	})

	t.Run("webhook trigger", func(t *testing.T) {
		exec := EdgeFunctionExecution{
			TriggerType: "webhook",
		}
		assert.Equal(t, "webhook", exec.TriggerType)
	})
}

// =============================================================================
// Execution Status Tests
// =============================================================================

func TestExecutionStatus(t *testing.T) {
	statuses := []string{
		"pending",
		"running",
		"success",
		"error",
		"timeout",
		"cancelled",
	}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			exec := EdgeFunctionExecution{
				Status: status,
			}
			assert.Equal(t, status, exec.Status)
		})
	}
}

// =============================================================================
// Source Types Tests
// =============================================================================

func TestSourceTypes(t *testing.T) {
	t.Run("filesystem source", func(t *testing.T) {
		fn := EdgeFunction{
			Name:   "fs-function",
			Source: "filesystem",
		}
		assert.Equal(t, "filesystem", fn.Source)
	})

	t.Run("API source", func(t *testing.T) {
		fn := EdgeFunction{
			Name:   "api-function",
			Source: "api",
		}
		assert.Equal(t, "api", fn.Source)
	})
}
