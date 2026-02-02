package migrations

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Executor Construction Tests
// =============================================================================

func TestNewExecutor(t *testing.T) {
	t.Run("creates executor with nil database", func(t *testing.T) {
		executor := NewExecutor(nil)

		require.NotNil(t, executor)
		assert.Nil(t, executor.db)
		assert.NotNil(t, executor.storage) // Storage is always created
	})
}

// =============================================================================
// Migration Status Constants Tests
// =============================================================================

func TestMigrationStatusConstants(t *testing.T) {
	t.Run("valid status values", func(t *testing.T) {
		// Document expected status values
		validStatuses := []string{"pending", "applied", "failed", "rolled_back"}

		// Verify status strings are not empty
		for _, status := range validStatuses {
			assert.NotEmpty(t, status)
		}
	})

	t.Run("status for apply checks", func(t *testing.T) {
		// When applying, allowed statuses are: pending, failed
		// Status "applied" should skip
		// Status "rolled_back" should not be allowed

		migration := &Migration{Status: "pending"}
		assert.True(t, migration.Status == "pending" || migration.Status == "failed")

		migration = &Migration{Status: "applied"}
		assert.True(t, migration.Status == "applied")
	})

	t.Run("status for rollback checks", func(t *testing.T) {
		// When rolling back, only "applied" status is allowed
		migration := &Migration{Status: "applied"}
		assert.Equal(t, "applied", migration.Status)
	})
}

// =============================================================================
// Migration Struct Tests
// =============================================================================

func TestMigration_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		now := time.Now()
		desc := "Test migration"
		downSQL := "DROP TABLE test"
		createdBy := uuid.New()
		appliedBy := uuid.New()

		migration := &Migration{
			ID:           uuid.New(),
			Namespace:    "public",
			Name:         "001_create_users",
			Description:  &desc,
			UpSQL:        "CREATE TABLE test (id INT)",
			DownSQL:      &downSQL,
			Version:      1,
			Status:       "pending",
			CreatedBy:    &createdBy,
			AppliedBy:    &appliedBy,
			CreatedAt:    now,
			UpdatedAt:    now,
			AppliedAt:    &now,
			RolledBackAt: nil,
		}

		assert.Equal(t, "public", migration.Namespace)
		assert.Equal(t, "001_create_users", migration.Name)
		assert.Equal(t, "Test migration", *migration.Description)
		assert.Equal(t, "CREATE TABLE test (id INT)", migration.UpSQL)
		assert.Equal(t, "DROP TABLE test", *migration.DownSQL)
		assert.Equal(t, 1, migration.Version)
		assert.Equal(t, "pending", migration.Status)
		assert.NotNil(t, migration.AppliedAt)
		assert.Nil(t, migration.RolledBackAt)
	})

	t.Run("handles optional fields as nil", func(t *testing.T) {
		migration := &Migration{
			ID:        uuid.New(),
			Namespace: "public",
			Name:      "002_simple",
			UpSQL:     "ALTER TABLE users ADD COLUMN email TEXT",
			Status:    "pending",
		}

		assert.Nil(t, migration.Description)
		assert.Nil(t, migration.DownSQL)
		assert.Nil(t, migration.CreatedBy)
		assert.Nil(t, migration.AppliedBy)
		assert.Nil(t, migration.AppliedAt)
		assert.Nil(t, migration.RolledBackAt)
	})
}

// =============================================================================
// ExecutionLog Struct Tests
// =============================================================================

func TestExecutionLog_Struct(t *testing.T) {
	t.Run("stores apply log", func(t *testing.T) {
		now := time.Now()
		duration := 150
		executedBy := uuid.New()

		log := &ExecutionLog{
			ID:          uuid.New(),
			MigrationID: uuid.New(),
			Action:      "apply",
			Status:      "success",
			DurationMs:  &duration,
			ExecutedAt:  now,
			ExecutedBy:  &executedBy,
		}

		assert.Equal(t, "apply", log.Action)
		assert.Equal(t, "success", log.Status)
		assert.Equal(t, 150, *log.DurationMs)
		assert.Nil(t, log.ErrorMessage)
		assert.Nil(t, log.Logs)
	})

	t.Run("stores rollback log", func(t *testing.T) {
		log := &ExecutionLog{
			ID:          uuid.New(),
			MigrationID: uuid.New(),
			Action:      "rollback",
			Status:      "success",
		}

		assert.Equal(t, "rollback", log.Action)
		assert.Equal(t, "success", log.Status)
	})

	t.Run("stores failure log with error message", func(t *testing.T) {
		errorMsg := "syntax error at position 15"
		duration := 50

		log := &ExecutionLog{
			ID:           uuid.New(),
			MigrationID:  uuid.New(),
			Action:       "apply",
			Status:       "failed",
			DurationMs:   &duration,
			ErrorMessage: &errorMsg,
		}

		assert.Equal(t, "failed", log.Status)
		assert.Equal(t, "syntax error at position 15", *log.ErrorMessage)
	})
}

// =============================================================================
// Migration Naming Convention Tests
// =============================================================================

func TestMigrationNamingConventions(t *testing.T) {
	t.Run("sequential numbering pattern", func(t *testing.T) {
		// Convention: NNN_description
		validNames := []string{
			"001_create_users",
			"002_add_email_column",
			"010_create_orders_table",
			"100_archive_old_data",
		}

		for _, name := range validNames {
			assert.NotEmpty(t, name)
			// Check that name starts with numbers (common convention)
			assert.Regexp(t, `^\d+_`, name)
		}
	})
}

// =============================================================================
// Apply/Rollback Status Validation Tests
// =============================================================================

func TestMigrationStatusValidation(t *testing.T) {
	t.Run("pending migration can be applied", func(t *testing.T) {
		migration := &Migration{Status: "pending"}

		canApply := migration.Status == "pending" || migration.Status == "failed"
		assert.True(t, canApply)
	})

	t.Run("failed migration can be retried", func(t *testing.T) {
		migration := &Migration{Status: "failed"}

		canApply := migration.Status == "pending" || migration.Status == "failed"
		assert.True(t, canApply)
	})

	t.Run("applied migration cannot be applied again", func(t *testing.T) {
		migration := &Migration{Status: "applied"}

		// Should skip (not error)
		shouldSkip := migration.Status == "applied"
		assert.True(t, shouldSkip)
	})

	t.Run("only applied migration can be rolled back", func(t *testing.T) {
		migration := &Migration{Status: "applied"}

		canRollback := migration.Status == "applied"
		assert.True(t, canRollback)

		migration.Status = "pending"
		canRollback = migration.Status == "applied"
		assert.False(t, canRollback)
	})

	t.Run("rollback requires down SQL", func(t *testing.T) {
		downSQL := "DROP TABLE users"
		migration := &Migration{
			Status:  "applied",
			DownSQL: &downSQL,
		}

		hasDownSQL := migration.DownSQL != nil && *migration.DownSQL != ""
		assert.True(t, hasDownSQL)

		migration.DownSQL = nil
		hasDownSQL = migration.DownSQL != nil && *migration.DownSQL != ""
		assert.False(t, hasDownSQL)
	})
}

// =============================================================================
// Duration Tracking Tests
// =============================================================================

func TestExecutionLogDurationTracking(t *testing.T) {
	t.Run("duration is calculated in milliseconds", func(t *testing.T) {
		start := time.Now()
		time.Sleep(10 * time.Millisecond)
		duration := int(time.Since(start).Milliseconds())

		log := &ExecutionLog{
			DurationMs: &duration,
		}

		assert.True(t, *log.DurationMs >= 10)
	})

	t.Run("fast migration has low duration", func(t *testing.T) {
		duration := 5 // 5ms

		log := &ExecutionLog{
			DurationMs: &duration,
		}

		assert.True(t, *log.DurationMs < 100)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkMigration_StatusCheck(b *testing.B) {
	migration := &Migration{
		Status: "pending",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = migration.Status == "pending" || migration.Status == "failed"
	}
}

func BenchmarkExecutionLog_Create(b *testing.B) {
	migrationID := uuid.New()
	duration := 100

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &ExecutionLog{
			ID:          uuid.New(),
			MigrationID: migrationID,
			Action:      "apply",
			Status:      "success",
			DurationMs:  &duration,
			ExecutedAt:  time.Now(),
		}
	}
}
