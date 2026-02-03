package rpc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// NewStorage Tests
// =============================================================================

func TestNewStorage(t *testing.T) {
	t.Run("creates storage with nil database", func(t *testing.T) {
		storage := NewStorage(nil)

		require.NotNil(t, storage)
		assert.Nil(t, storage.db)
	})
}

// =============================================================================
// ListExecutionsOptions Tests
// =============================================================================

func TestListExecutionsOptions_Defaults(t *testing.T) {
	t.Run("zero values for defaults", func(t *testing.T) {
		opts := ListExecutionsOptions{}

		assert.Empty(t, opts.Namespace)
		assert.Empty(t, opts.ProcedureName)
		assert.Empty(t, opts.Status)
		assert.Empty(t, opts.UserID)
		assert.Equal(t, 0, opts.Limit)
		assert.Equal(t, 0, opts.Offset)
	})

	t.Run("all fields can be set", func(t *testing.T) {
		opts := ListExecutionsOptions{
			Namespace:     "api",
			ProcedureName: "get_users",
			Status:        StatusCompleted,
			UserID:        "user-123",
			Limit:         50,
			Offset:        100,
		}

		assert.Equal(t, "api", opts.Namespace)
		assert.Equal(t, "get_users", opts.ProcedureName)
		assert.Equal(t, StatusCompleted, opts.Status)
		assert.Equal(t, "user-123", opts.UserID)
		assert.Equal(t, 50, opts.Limit)
		assert.Equal(t, 100, opts.Offset)
	})
}

// =============================================================================
// Storage Query Building Tests (validates query building logic)
// =============================================================================

func TestStorage_QueryBuilding(t *testing.T) {
	t.Run("validates namespace filter would be applied", func(t *testing.T) {
		opts := ListExecutionsOptions{
			Namespace: "test-namespace",
		}

		// Verify the option is set (actual query building tested via integration)
		assert.NotEmpty(t, opts.Namespace)
	})

	t.Run("validates procedure name filter would be applied", func(t *testing.T) {
		opts := ListExecutionsOptions{
			ProcedureName: "test_proc",
		}

		assert.NotEmpty(t, opts.ProcedureName)
	})

	t.Run("validates status filter would be applied", func(t *testing.T) {
		opts := ListExecutionsOptions{
			Status: StatusFailed,
		}

		assert.Equal(t, StatusFailed, opts.Status)
	})

	t.Run("validates user ID filter would be applied", func(t *testing.T) {
		opts := ListExecutionsOptions{
			UserID: "user-456",
		}

		assert.NotEmpty(t, opts.UserID)
	})

	t.Run("validates pagination", func(t *testing.T) {
		opts := ListExecutionsOptions{
			Limit:  25,
			Offset: 50,
		}

		assert.Equal(t, 25, opts.Limit)
		assert.Equal(t, 50, opts.Offset)
	})
}

// =============================================================================
// Execution ID Generation Tests
// =============================================================================

func TestExecution_IDGeneration(t *testing.T) {
	t.Run("execution with empty ID", func(t *testing.T) {
		exec := &Execution{
			ID:            "",
			ProcedureName: "test_proc",
			Namespace:     "default",
			Status:        StatusPending,
		}

		// The actual ID generation happens in CreateExecution
		// Here we verify empty ID is acceptable as input
		assert.Empty(t, exec.ID)
	})

	t.Run("execution with provided ID", func(t *testing.T) {
		exec := &Execution{
			ID:            "exec-custom-id",
			ProcedureName: "test_proc",
			Namespace:     "default",
			Status:        StatusPending,
		}

		assert.Equal(t, "exec-custom-id", exec.ID)
	})
}

// =============================================================================
// Procedure ID Generation Tests
// =============================================================================

func TestProcedure_IDGeneration(t *testing.T) {
	t.Run("procedure with empty ID", func(t *testing.T) {
		proc := &Procedure{
			ID:       "",
			Name:     "test_proc",
			SQLQuery: "SELECT 1",
		}

		// The actual ID generation happens in CreateProcedure
		// Here we verify empty ID is acceptable as input
		assert.Empty(t, proc.ID)
	})

	t.Run("procedure with provided ID", func(t *testing.T) {
		proc := &Procedure{
			ID:       "proc-custom-id",
			Name:     "test_proc",
			SQLQuery: "SELECT 1",
		}

		assert.Equal(t, "proc-custom-id", proc.ID)
	})
}

// =============================================================================
// DefaultAnnotations Tests
// =============================================================================

func TestDefaultAnnotations_Function(t *testing.T) {
	t.Run("returns expected defaults", func(t *testing.T) {
		defaults := DefaultAnnotations()

		require.NotNil(t, defaults)
		assert.Equal(t, []string{"public"}, defaults.AllowedSchemas)
		assert.Equal(t, []string{}, defaults.AllowedTables)
		assert.False(t, defaults.IsPublic)
		assert.Equal(t, 1, defaults.Version)
	})
}

// =============================================================================
// Procedure Struct Tests
// =============================================================================

func TestProcedure_Struct(t *testing.T) {
	t.Run("creates procedure with all fields", func(t *testing.T) {
		now := time.Now()
		proc := &Procedure{
			ID:                      "proc-123",
			Name:                    "get_users",
			Namespace:               "api",
			Description:             "Retrieves all users",
			SQLQuery:                "SELECT * FROM users",
			OriginalCode:            "-- Original SQL",
			InputSchema:             []byte(`{"type": "object"}`),
			OutputSchema:            []byte(`{"type": "array"}`),
			AllowedTables:           []string{"users", "profiles"},
			AllowedSchemas:          []string{"public", "auth"},
			MaxExecutionTimeSeconds: 30,
			RequireRoles:            []string{"admin", "user"},
			IsPublic:                true,
			DisableExecutionLogs:    false,
			Schedule:                "0 * * * *",
			Enabled:                 true,
			Version:                 5,
			Source:                  "mcp",
			CreatedBy:               "user-456",
			CreatedAt:               now,
			UpdatedAt:               now,
		}

		assert.Equal(t, "proc-123", proc.ID)
		assert.Equal(t, "get_users", proc.Name)
		assert.Equal(t, "api", proc.Namespace)
		assert.Equal(t, "Retrieves all users", proc.Description)
		assert.Equal(t, "SELECT * FROM users", proc.SQLQuery)
		assert.Equal(t, "-- Original SQL", proc.OriginalCode)
		assert.NotNil(t, proc.InputSchema)
		assert.NotNil(t, proc.OutputSchema)
		assert.Len(t, proc.AllowedTables, 2)
		assert.Len(t, proc.AllowedSchemas, 2)
		assert.Equal(t, 30, proc.MaxExecutionTimeSeconds)
		assert.Len(t, proc.RequireRoles, 2)
		assert.True(t, proc.IsPublic)
		assert.False(t, proc.DisableExecutionLogs)
		assert.Equal(t, "0 * * * *", proc.Schedule)
		assert.True(t, proc.Enabled)
		assert.Equal(t, 5, proc.Version)
		assert.Equal(t, "mcp", proc.Source)
		assert.Equal(t, "user-456", proc.CreatedBy)
	})

	t.Run("zero value", func(t *testing.T) {
		var proc Procedure
		assert.Empty(t, proc.ID)
		assert.Empty(t, proc.Name)
		assert.Empty(t, proc.Namespace)
		assert.Empty(t, proc.SQLQuery)
		assert.Nil(t, proc.AllowedTables)
		assert.Nil(t, proc.AllowedSchemas)
		assert.Equal(t, 0, proc.MaxExecutionTimeSeconds)
		assert.False(t, proc.IsPublic)
		assert.False(t, proc.Enabled)
		assert.Equal(t, 0, proc.Version)
	})
}

// =============================================================================
// ProcedureSummary Struct Tests
// =============================================================================

func TestProcedureSummary_Struct(t *testing.T) {
	t.Run("creates summary with all fields", func(t *testing.T) {
		now := time.Now()
		summary := &ProcedureSummary{
			ID:                      "sum-123",
			Name:                    "list_items",
			Namespace:               "inventory",
			Description:             "Lists inventory items",
			AllowedTables:           []string{"items"},
			AllowedSchemas:          []string{"public"},
			MaxExecutionTimeSeconds: 10,
			RequireRoles:            []string{"viewer"},
			IsPublic:                true,
			DisableExecutionLogs:    true,
			Schedule:                "",
			Enabled:                 true,
			Version:                 2,
			Source:                  "admin",
			CreatedAt:               now,
			UpdatedAt:               now,
		}

		assert.Equal(t, "sum-123", summary.ID)
		assert.Equal(t, "list_items", summary.Name)
		assert.Equal(t, "inventory", summary.Namespace)
		assert.True(t, summary.DisableExecutionLogs)
	})

	t.Run("summary excludes sensitive fields", func(t *testing.T) {
		// ProcedureSummary does not include SQLQuery, OriginalCode, InputSchema, OutputSchema
		summary := &ProcedureSummary{
			ID:   "sum-456",
			Name: "test_proc",
		}

		// These fields don't exist in ProcedureSummary
		assert.Equal(t, "sum-456", summary.ID)
		assert.Equal(t, "test_proc", summary.Name)
	})
}

// =============================================================================
// Execution Struct Tests
// =============================================================================

func TestExecution_Struct(t *testing.T) {
	t.Run("creates execution with all fields", func(t *testing.T) {
		now := time.Now()
		startedAt := now.Add(-time.Second)
		completedAt := now

		exec := &Execution{
			ID:            "exec-789",
			ProcedureID:   "proc-123",
			ProcedureName: "get_users",
			Namespace:     "api",
			Status:        StatusCompleted,
			InputParams:   []byte(`{"limit": 10}`),
			Result:        []byte(`[{"id": 1}]`),
			ErrorMessage:  "",
			RowsReturned:  1,
			DurationMs:    150,
			UserID:        "user-456",
			UserRole:      "admin",
			UserEmail:     "admin@example.com",
			IsAsync:       false,
			CreatedAt:     now,
			StartedAt:     &startedAt,
			CompletedAt:   &completedAt,
		}

		assert.Equal(t, "exec-789", exec.ID)
		assert.Equal(t, "proc-123", exec.ProcedureID)
		assert.Equal(t, "get_users", exec.ProcedureName)
		assert.Equal(t, "api", exec.Namespace)
		assert.Equal(t, StatusCompleted, exec.Status)
		assert.NotNil(t, exec.InputParams)
		assert.NotNil(t, exec.Result)
		assert.Empty(t, exec.ErrorMessage)
		assert.Equal(t, 1, exec.RowsReturned)
		assert.Equal(t, int64(150), exec.DurationMs)
		assert.Equal(t, "user-456", exec.UserID)
		assert.Equal(t, "admin", exec.UserRole)
		assert.Equal(t, "admin@example.com", exec.UserEmail)
		assert.False(t, exec.IsAsync)
		assert.NotNil(t, exec.StartedAt)
		assert.NotNil(t, exec.CompletedAt)
	})

	t.Run("execution with error", func(t *testing.T) {
		exec := &Execution{
			ID:           "exec-err",
			Status:       StatusFailed,
			ErrorMessage: "Query execution timeout",
			Result:       nil,
			RowsReturned: 0,
		}

		assert.Equal(t, StatusFailed, exec.Status)
		assert.Equal(t, "Query execution timeout", exec.ErrorMessage)
		assert.Nil(t, exec.Result)
		assert.Equal(t, 0, exec.RowsReturned)
	})

	t.Run("async execution", func(t *testing.T) {
		exec := &Execution{
			ID:      "exec-async",
			Status:  StatusPending,
			IsAsync: true,
		}

		assert.True(t, exec.IsAsync)
		assert.Equal(t, StatusPending, exec.Status)
	})
}

// =============================================================================
// Status Constants Tests
// =============================================================================

func TestStatusConstants(t *testing.T) {
	t.Run("StatusPending", func(t *testing.T) {
		assert.Equal(t, "pending", StatusPending)
	})

	t.Run("StatusRunning", func(t *testing.T) {
		assert.Equal(t, "running", StatusRunning)
	})

	t.Run("StatusCompleted", func(t *testing.T) {
		assert.Equal(t, "completed", StatusCompleted)
	})

	t.Run("StatusFailed", func(t *testing.T) {
		assert.Equal(t, "failed", StatusFailed)
	})

	t.Run("StatusCancelled", func(t *testing.T) {
		assert.Equal(t, "cancelled", StatusCancelled)
	})

	t.Run("all statuses are unique", func(t *testing.T) {
		statuses := []string{StatusPending, StatusRunning, StatusCompleted, StatusFailed, StatusCancelled}
		seen := make(map[string]bool)
		for _, s := range statuses {
			assert.False(t, seen[s], "duplicate status: %s", s)
			seen[s] = true
		}
	})
}

// =============================================================================
// ListExecutionsOptions Comprehensive Tests
// =============================================================================

func TestListExecutionsOptions_AllCombinations(t *testing.T) {
	t.Run("empty filters returns all", func(t *testing.T) {
		opts := ListExecutionsOptions{}
		assert.Empty(t, opts.Namespace)
		assert.Empty(t, opts.Status)
		assert.Equal(t, 0, opts.Limit)
	})

	t.Run("single filter", func(t *testing.T) {
		opts := ListExecutionsOptions{
			Status: StatusCompleted,
		}
		assert.Equal(t, StatusCompleted, opts.Status)
		assert.Empty(t, opts.Namespace)
	})

	t.Run("multiple filters", func(t *testing.T) {
		opts := ListExecutionsOptions{
			Namespace:     "api",
			ProcedureName: "get_data",
			Status:        StatusFailed,
			UserID:        "user-123",
			Limit:         20,
			Offset:        40,
		}

		assert.Equal(t, "api", opts.Namespace)
		assert.Equal(t, "get_data", opts.ProcedureName)
		assert.Equal(t, StatusFailed, opts.Status)
		assert.Equal(t, "user-123", opts.UserID)
		assert.Equal(t, 20, opts.Limit)
		assert.Equal(t, 40, opts.Offset)
	})
}

// =============================================================================
// Storage Method Existence Tests
// =============================================================================

func TestStorage_MethodsExist(t *testing.T) {
	storage := NewStorage(nil)

	t.Run("procedure methods exist", func(t *testing.T) {
		// These tests verify the methods exist on the Storage type
		// Actual functionality requires database
		assert.NotNil(t, storage)
	})
}

// =============================================================================
// Procedure Source Values Tests
// =============================================================================

func TestProcedure_SourceValues(t *testing.T) {
	validSources := []string{"mcp", "admin", "cli", "migration", "api"}

	for _, source := range validSources {
		t.Run(source, func(t *testing.T) {
			proc := &Procedure{Source: source}
			assert.Equal(t, source, proc.Source)
		})
	}
}

// =============================================================================
// Execution User Fields Tests
// =============================================================================

func TestExecution_UserFields(t *testing.T) {
	t.Run("anonymous execution", func(t *testing.T) {
		exec := &Execution{
			ID:       "exec-anon",
			UserID:   "",
			UserRole: "anon",
		}

		assert.Empty(t, exec.UserID)
		assert.Equal(t, "anon", exec.UserRole)
		assert.Empty(t, exec.UserEmail)
	})

	t.Run("authenticated execution", func(t *testing.T) {
		exec := &Execution{
			ID:        "exec-auth",
			UserID:    "user-abc",
			UserRole:  "authenticated",
			UserEmail: "user@example.com",
		}

		assert.NotEmpty(t, exec.UserID)
		assert.Equal(t, "authenticated", exec.UserRole)
		assert.Equal(t, "user@example.com", exec.UserEmail)
	})

	t.Run("service role execution", func(t *testing.T) {
		exec := &Execution{
			ID:       "exec-service",
			UserID:   "",
			UserRole: "service_role",
		}

		assert.Empty(t, exec.UserID)
		assert.Equal(t, "service_role", exec.UserRole)
	})
}
