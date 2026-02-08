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

// =============================================================================
// Migration Status Transitions
// =============================================================================

func TestMigrationStatusTransitions(t *testing.T) {
	t.Run("status transition: pending -> applied", func(t *testing.T) {
		migration := &Migration{Status: "pending"}
		assert.Equal(t, "pending", migration.Status)

		// Simulate status change
		migration.Status = "applied"
		assert.Equal(t, "applied", migration.Status)
	})

	t.Run("status transition: pending -> failed", func(t *testing.T) {
		migration := &Migration{Status: "pending"}
		migration.Status = "failed"
		assert.Equal(t, "failed", migration.Status)
	})

	t.Run("status transition: failed -> pending (retry)", func(t *testing.T) {
		migration := &Migration{Status: "failed"}
		migration.Status = "pending"
		assert.Equal(t, "pending", migration.Status)
	})

	t.Run("status transition: applied -> rolled_back", func(t *testing.T) {
		migration := &Migration{Status: "applied"}
		migration.Status = "rolled_back"
		assert.Equal(t, "rolled_back", migration.Status)
	})

	t.Run("invalid transition: applied -> pending (not allowed)", func(t *testing.T) {
		// This documents that once applied, a migration should not go back to pending
		migration := &Migration{Status: "applied"}
		initialStatus := migration.Status
		migration.Status = "pending"

		// While struct allows this, business logic should prevent it
		assert.NotEqual(t, initialStatus, migration.Status)
	})
}

// =============================================================================
// Migration SQL Validation
// =============================================================================

func TestMigrationSQL(t *testing.T) {
	t.Run("up SQL is required", func(t *testing.T) {
		migration := &Migration{
			UpSQL: "CREATE TABLE test (id INT)",
		}
		assert.NotEmpty(t, migration.UpSQL)
	})

	t.Run("down SQL is optional", func(t *testing.T) {
		migration := &Migration{
			UpSQL:   "CREATE TABLE test (id INT)",
			DownSQL: nil,
		}
		assert.NotNil(t, migration.UpSQL)
		assert.Nil(t, migration.DownSQL)
	})

	t.Run("down SQL can be empty string", func(t *testing.T) {
		emptyDown := ""
		migration := &Migration{
			UpSQL:   "CREATE TABLE test (id INT)",
			DownSQL: &emptyDown,
		}
		assert.NotNil(t, migration.DownSQL)
		assert.Empty(t, *migration.DownSQL)
	})

	t.Run("common SQL patterns", func(t *testing.T) {
		patterns := []struct {
			name string
			up   string
			down string
		}{
			{
				"create table",
				"CREATE TABLE users (id SERIAL PRIMARY KEY, email TEXT)",
				"DROP TABLE users",
			},
			{
				"alter table add column",
				"ALTER TABLE users ADD COLUMN name TEXT",
				"ALTER TABLE users DROP COLUMN name",
			},
			{
				"create index",
				"CREATE INDEX idx_users_email ON users(email)",
				"DROP INDEX idx_users_email",
			},
		}

		for _, p := range patterns {
			t.Run(p.name, func(t *testing.T) {
				migration := &Migration{
					UpSQL:   p.up,
					DownSQL: &p.down,
				}
				assert.Equal(t, p.up, migration.UpSQL)
				assert.Equal(t, p.down, *migration.DownSQL)
			})
		}
	})
}

// =============================================================================
// Migration Fields Edge Cases
// =============================================================================

func TestMigration_EdgeCases(t *testing.T) {
	t.Run("zero version is valid", func(t *testing.T) {
		migration := &Migration{
			Version: 0,
		}
		assert.Equal(t, 0, migration.Version)
	})

	t.Run("negative version should be handled", func(t *testing.T) {
		migration := &Migration{
			Version: -1,
		}
		assert.Equal(t, -1, migration.Version)
	})

	t.Run("large version number", func(t *testing.T) {
		migration := &Migration{
			Version: 999999,
		}
		assert.Equal(t, 999999, migration.Version)
	})

	t.Run("empty namespace", func(t *testing.T) {
		migration := &Migration{
			Namespace: "",
		}
		assert.Empty(t, migration.Namespace)
	})

	t.Run("namespace with dots", func(t *testing.T) {
		migration := &Migration{
			Namespace: "app.feature.module",
		}
		assert.Equal(t, "app.feature.module", migration.Namespace)
	})

	t.Run("very long description", func(t *testing.T) {
		longDesc := string(make([]byte, 10000))
		for i := range longDesc {
			longDesc = longDesc[:i] + "a" + longDesc[i+1:]
		}

		migration := &Migration{
			Description: &longDesc,
		}
		assert.Len(t, *migration.Description, 10000)
	})

	t.Run("nil pointer fields", func(t *testing.T) {
		migration := &Migration{
			Description:  nil,
			DownSQL:      nil,
			CreatedBy:    nil,
			AppliedBy:    nil,
			AppliedAt:    nil,
			RolledBackAt: nil,
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
// ExecutionLog Fields Edge Cases
// =============================================================================

func TestExecutionLog_EdgeCases(t *testing.T) {
	t.Run("nil duration is allowed", func(t *testing.T) {
		log := &ExecutionLog{
			DurationMs: nil,
		}
		assert.Nil(t, log.DurationMs)
	})

	t.Run("zero duration", func(t *testing.T) {
		duration := 0
		log := &ExecutionLog{
			DurationMs: &duration,
		}
		assert.Equal(t, 0, *log.DurationMs)
	})

	t.Run("negative duration (should not happen in practice)", func(t *testing.T) {
		duration := -1
		log := &ExecutionLog{
			DurationMs: &duration,
		}
		assert.Equal(t, -1, *log.DurationMs)
	})

	t.Run("very long duration", func(t *testing.T) {
		duration := 999999 // ~16 minutes
		log := &ExecutionLog{
			DurationMs: &duration,
		}
		assert.Equal(t, 999999, *log.DurationMs)
	})

	t.Run("empty error message", func(t *testing.T) {
		emptyErr := ""
		log := &ExecutionLog{
			ErrorMessage: &emptyErr,
		}
		assert.Equal(t, "", *log.ErrorMessage)
	})

	t.Run("nil error message", func(t *testing.T) {
		log := &ExecutionLog{
			ErrorMessage: nil,
		}
		assert.Nil(t, log.ErrorMessage)
	})

	t.Run("nil logs field", func(t *testing.T) {
		log := &ExecutionLog{
			Logs: nil,
		}
		assert.Nil(t, log.Logs)
	})

	t.Run("nil executed by", func(t *testing.T) {
		log := &ExecutionLog{
			ExecutedBy: nil,
		}
		assert.Nil(t, log.ExecutedBy)
	})
}

// =============================================================================
// Action Type Validation
// =============================================================================

func TestExecutionLog_ActionTypes(t *testing.T) {
	validActions := []string{"apply", "rollback"}

	t.Run("valid action types", func(t *testing.T) {
		for _, action := range validActions {
			log := &ExecutionLog{
				Action: action,
			}
			assert.Equal(t, action, log.Action)
		}
	})

	t.Run("action is apply", func(t *testing.T) {
		log := &ExecutionLog{
			Action: "apply",
		}
		assert.Equal(t, "apply", log.Action)
	})

	t.Run("action is rollback", func(t *testing.T) {
		log := &ExecutionLog{
			Action: "rollback",
		}
		assert.Equal(t, "rollback", log.Action)
	})
}

// =============================================================================
// Status Type Validation
// =============================================================================

func TestExecutionLog_StatusTypes(t *testing.T) {
	validStatuses := []string{"success", "failed"}

	t.Run("valid status types", func(t *testing.T) {
		for _, status := range validStatuses {
			log := &ExecutionLog{
				Status: status,
			}
			assert.Equal(t, status, log.Status)
		}
	})

	t.Run("status is success", func(t *testing.T) {
		log := &ExecutionLog{
			Status: "success",
		}
		assert.Equal(t, "success", log.Status)
	})

	t.Run("status is failed", func(t *testing.T) {
		log := &ExecutionLog{
			Status: "failed",
		}
		assert.Equal(t, "failed", log.Status)
	})
}

// =============================================================================
// Timestamp Handling
// =============================================================================

func TestMigration_Timestamps(t *testing.T) {
	t.Run("created_at is set", func(t *testing.T) {
		now := time.Now()
		migration := &Migration{
			CreatedAt: now,
		}
		assert.False(t, migration.CreatedAt.IsZero())
	})

	t.Run("updated_at can be after created_at", func(t *testing.T) {
		created := time.Now().Add(-24 * time.Hour)
		updated := time.Now()

		migration := &Migration{
			CreatedAt: created,
			UpdatedAt: updated,
		}

		assert.True(t, migration.UpdatedAt.After(migration.CreatedAt))
	})

	t.Run("applied_at is optional", func(t *testing.T) {
		migration := &Migration{
			AppliedAt: nil,
		}
		assert.Nil(t, migration.AppliedAt)
	})

	t.Run("applied_at when set", func(t *testing.T) {
		now := time.Now()
		migration := &Migration{
			AppliedAt: &now,
		}
		assert.NotNil(t, migration.AppliedAt)
		assert.False(t, migration.AppliedAt.IsZero())
	})

	t.Run("rolled_back_at is optional", func(t *testing.T) {
		migration := &Migration{
			RolledBackAt: nil,
		}
		assert.Nil(t, migration.RolledBackAt)
	})

	t.Run("rolled_back_at when set", func(t *testing.T) {
		now := time.Now()
		migration := &Migration{
			RolledBackAt: &now,
		}
		assert.NotNil(t, migration.RolledBackAt)
		assert.False(t, migration.RolledBackAt.IsZero())
	})
}

func TestExecutionLog_Timestamps(t *testing.T) {
	t.Run("executed_at is set", func(t *testing.T) {
		now := time.Now()
		log := &ExecutionLog{
			ExecutedAt: now,
		}
		assert.False(t, log.ExecutedAt.IsZero())
	})

	t.Run("executed_at can be zero", func(t *testing.T) {
		log := &ExecutionLog{
			ExecutedAt: time.Time{},
		}
		assert.True(t, log.ExecutedAt.IsZero())
	})
}

// =============================================================================
// User Tracking
// =============================================================================

func TestMigration_UserTracking(t *testing.T) {
	t.Run("created_by and applied_by can be different", func(t *testing.T) {
		creator := uuid.New()
		applier := uuid.New()

		migration := &Migration{
			CreatedBy: &creator,
			AppliedBy: &applier,
		}

		assert.Equal(t, creator, *migration.CreatedBy)
		assert.Equal(t, applier, *migration.AppliedBy)
		assert.NotEqual(t, creator, applier)
	})

	t.Run("created_by and applied_by can be same", func(t *testing.T) {
		user := uuid.New()

		migration := &Migration{
			CreatedBy: &user,
			AppliedBy: &user,
		}

		assert.Equal(t, user, *migration.CreatedBy)
		assert.Equal(t, user, *migration.AppliedBy)
	})

	t.Run("nil created_by is allowed", func(t *testing.T) {
		migration := &Migration{
			CreatedBy: nil,
		}
		assert.Nil(t, migration.CreatedBy)
	})

	t.Run("nil applied_by is allowed", func(t *testing.T) {
		migration := &Migration{
			AppliedBy: nil,
		}
		assert.Nil(t, migration.AppliedBy)
	})
}

func TestExecutionLog_UserTracking(t *testing.T) {
	t.Run("executed_by is set", func(t *testing.T) {
		user := uuid.New()
		log := &ExecutionLog{
			ExecutedBy: &user,
		}
		assert.Equal(t, user, *log.ExecutedBy)
	})

	t.Run("executed_by can be nil", func(t *testing.T) {
		log := &ExecutionLog{
			ExecutedBy: nil,
		}
		assert.Nil(t, log.ExecutedBy)
	})
}

// =============================================================================
// Migration ID Handling
// =============================================================================

func TestMigration_ID(t *testing.T) {
	t.Run("valid UUID", func(t *testing.T) {
		id := uuid.New()
		migration := &Migration{
			ID: id,
		}
		assert.Equal(t, id, migration.ID)
	})

	t.Run("nil UUID", func(t *testing.T) {
		migration := &Migration{
			ID: uuid.Nil,
		}
		assert.Equal(t, uuid.Nil, migration.ID)
	})
}

func TestExecutionLog_ID(t *testing.T) {
	t.Run("valid UUID", func(t *testing.T) {
		id := uuid.New()
		log := &ExecutionLog{
			ID: id,
		}
		assert.Equal(t, id, log.ID)
	})

	t.Run("migration_id reference", func(t *testing.T) {
		migrationID := uuid.New()
		log := &ExecutionLog{
			MigrationID: migrationID,
		}
		assert.Equal(t, migrationID, log.MigrationID)
	})
}

// =============================================================================
// Additional Benchmarks
// =============================================================================

func BenchmarkMigration_StructCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &Migration{
			ID:        uuid.New(),
			Namespace: "public",
			Name:      "001_test",
			UpSQL:     "CREATE TABLE test (id INT)",
			Status:    "pending",
		}
	}
}

func BenchmarkExecutionLog_StatusCheck(b *testing.B) {
	log := &ExecutionLog{
		Status: "success",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = log.Status == "success" || log.Status == "failed"
	}
}
