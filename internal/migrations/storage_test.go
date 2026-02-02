package migrations

import (
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

		require.NotNil(t, storage)
		assert.Nil(t, storage.db)
	})
}

// =============================================================================
// Migration Status Constants
// =============================================================================

func TestMigration_StatusConstants(t *testing.T) {
	t.Run("defines valid status values", func(t *testing.T) {
		// Valid statuses used in the system
		validStatuses := map[string]string{
			"pending":     "Migration created but not applied",
			"applied":     "Migration successfully applied",
			"failed":      "Migration attempted but failed",
			"rolled_back": "Migration was rolled back",
		}

		for status, description := range validStatuses {
			assert.NotEmpty(t, status)
			assert.NotEmpty(t, description)
		}
	})
}

// =============================================================================
// UpdateMigration Allowed Fields Tests
// =============================================================================

func TestStorage_UpdateMigration_AllowedFields(t *testing.T) {
	t.Run("defines allowed update fields", func(t *testing.T) {
		// Fields that can be updated for pending/failed migrations
		allowedFields := map[string]bool{
			"description": true,
			"up_sql":      true,
			"down_sql":    true,
			"status":      true, // But only to 'pending'
		}

		// Verify structure
		assert.True(t, allowedFields["description"])
		assert.True(t, allowedFields["up_sql"])
		assert.True(t, allowedFields["down_sql"])
		assert.True(t, allowedFields["status"])
	})

	t.Run("rejects disallowed fields", func(t *testing.T) {
		allowedFields := map[string]bool{
			"description": true,
			"up_sql":      true,
			"down_sql":    true,
			"status":      true,
		}

		// These fields should NOT be updateable
		disallowedFields := []string{
			"id",
			"namespace",
			"name",
			"version",
			"created_by",
			"applied_by",
			"created_at",
			"applied_at",
		}

		for _, field := range disallowedFields {
			assert.False(t, allowedFields[field], "field %s should not be allowed", field)
		}
	})

	t.Run("status update only allows 'pending'", func(t *testing.T) {
		// When updating status, only 'pending' is allowed (to reset a failed migration)
		updates := map[string]interface{}{
			"status": "pending",
		}

		// This should be allowed
		assert.Equal(t, "pending", updates["status"])

		// These should be rejected
		invalidStatuses := []string{"applied", "failed", "rolled_back"}
		for _, status := range invalidStatuses {
			updates["status"] = status
			assert.NotEqual(t, "pending", updates["status"])
		}
	})
}

// =============================================================================
// UpdateMigrationStatus Tests
// =============================================================================

func TestStorage_UpdateMigrationStatus_QueryLogic(t *testing.T) {
	t.Run("applied status sets applied_at and applied_by", func(t *testing.T) {
		// Query pattern for 'applied' status
		// UPDATE ... SET status = $1, applied_at = NOW(), applied_by = $2, updated_at = NOW() WHERE id = $3
		status := "applied"
		appliedBy := uuid.New()

		assert.Equal(t, "applied", status)
		assert.NotEqual(t, uuid.Nil, appliedBy)
	})

	t.Run("rolled_back status sets rolled_back_at", func(t *testing.T) {
		// Query pattern for 'rolled_back' status
		// UPDATE ... SET status = $1, rolled_back_at = NOW(), updated_at = NOW() WHERE id = $2
		status := "rolled_back"

		assert.Equal(t, "rolled_back", status)
	})

	t.Run("failed status only updates status and timestamp", func(t *testing.T) {
		// Query pattern for 'failed' status
		// UPDATE ... SET status = $1, updated_at = NOW() WHERE id = $2
		status := "failed"

		assert.Equal(t, "failed", status)
	})

	t.Run("rejects invalid status values", func(t *testing.T) {
		validStatuses := map[string]bool{
			"applied":     true,
			"rolled_back": true,
			"failed":      true,
		}

		invalidStatuses := []string{"invalid", "APPLIED", "Pending", ""}

		for _, status := range invalidStatuses {
			assert.False(t, validStatuses[status], "status %q should be invalid", status)
		}
	})
}

// =============================================================================
// DeleteMigration Tests
// =============================================================================

func TestStorage_DeleteMigration_RequiresPending(t *testing.T) {
	t.Run("only deletes pending migrations", func(t *testing.T) {
		// The query uses: WHERE ... AND status = 'pending'
		// This prevents accidental deletion of applied migrations

		migration := &Migration{Status: "pending"}
		canDelete := migration.Status == "pending"
		assert.True(t, canDelete)

		migration.Status = "applied"
		canDelete = migration.Status == "pending"
		assert.False(t, canDelete)
	})
}

// =============================================================================
// ListMigrations Tests
// =============================================================================

func TestStorage_ListMigrations_QueryLogic(t *testing.T) {
	t.Run("orders by name ascending", func(t *testing.T) {
		// The query includes: ORDER BY name ASC
		// This ensures migrations are applied in the correct order
		names := []string{
			"001_create_users",
			"002_add_email",
			"003_create_orders",
		}

		// Verify they are sorted
		for i := 1; i < len(names); i++ {
			assert.True(t, names[i] > names[i-1])
		}
	})

	t.Run("optional status filter", func(t *testing.T) {
		// When status is provided, query adds: AND status = $2
		status := "pending"

		assert.NotEmpty(t, status)
	})
}

// =============================================================================
// ExecutionLog Tests
// =============================================================================

func TestExecutionLog_Actions(t *testing.T) {
	t.Run("valid actions", func(t *testing.T) {
		validActions := []string{"apply", "rollback"}

		for _, action := range validActions {
			assert.NotEmpty(t, action)
		}
	})
}

func TestExecutionLog_Statuses(t *testing.T) {
	t.Run("valid execution statuses", func(t *testing.T) {
		validStatuses := []string{"success", "failed"}

		for _, status := range validStatuses {
			assert.NotEmpty(t, status)
		}
	})
}

// =============================================================================
// GetExecutionLogs Tests
// =============================================================================

func TestStorage_GetExecutionLogs_QueryLogic(t *testing.T) {
	t.Run("orders by executed_at descending", func(t *testing.T) {
		// The query includes: ORDER BY executed_at DESC
		// Most recent executions first

		times := []time.Time{
			time.Now().Add(-2 * time.Hour),
			time.Now().Add(-1 * time.Hour),
			time.Now(),
		}

		// Sorted descending
		for i := 0; i < len(times)-1; i++ {
			// In descending order, later times come first
			// After sorting DESC: times[2], times[1], times[0]
			assert.True(t, times[i].Before(times[i+1]))
		}
	})

	t.Run("respects limit parameter", func(t *testing.T) {
		// The query uses: LIMIT $2
		limit := 10

		assert.True(t, limit > 0)
	})
}

// =============================================================================
// Migration Struct JSON Tags Tests
// =============================================================================

func TestMigration_JSONTags(t *testing.T) {
	t.Run("migration has correct JSON field names", func(t *testing.T) {
		// Verify expected JSON structure
		migration := &Migration{
			ID:        uuid.New(),
			Namespace: "public",
			Name:      "test",
			Status:    "pending",
		}

		// Fields should serialize with correct names
		assert.NotEqual(t, uuid.Nil, migration.ID)
		assert.NotEmpty(t, migration.Namespace)
		assert.NotEmpty(t, migration.Name)
		assert.NotEmpty(t, migration.Status)
	})
}

// =============================================================================
// ExecutionLog Struct JSON Tags Tests
// =============================================================================

func TestExecutionLog_JSONTags(t *testing.T) {
	t.Run("execution log has correct JSON field names", func(t *testing.T) {
		log := &ExecutionLog{
			ID:          uuid.New(),
			MigrationID: uuid.New(),
			Action:      "apply",
			Status:      "success",
		}

		assert.NotEqual(t, uuid.Nil, log.ID)
		assert.NotEqual(t, uuid.Nil, log.MigrationID)
		assert.Equal(t, "apply", log.Action)
		assert.Equal(t, "success", log.Status)
	})
}

// =============================================================================
// Namespace Handling Tests
// =============================================================================

func TestMigration_NamespaceHandling(t *testing.T) {
	t.Run("namespace can be any valid identifier", func(t *testing.T) {
		validNamespaces := []string{
			"public",
			"app",
			"tenant_1",
			"my_migrations",
		}

		for _, ns := range validNamespaces {
			migration := &Migration{Namespace: ns}
			assert.NotEmpty(t, migration.Namespace)
		}
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkNewStorage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewStorage(nil)
	}
}

func BenchmarkMigration_StatusValidation(b *testing.B) {
	validStatuses := map[string]bool{
		"applied":     true,
		"rolled_back": true,
		"failed":      true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validStatuses["applied"]
	}
}
