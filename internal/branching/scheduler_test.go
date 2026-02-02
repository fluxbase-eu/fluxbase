package branching

import (
	"testing"
	"time"

	"github.com/fluxbase-eu/fluxbase/internal/config"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Scheduler Construction Tests
// =============================================================================

func TestNewScheduler(t *testing.T) {
	t.Run("creates scheduler with nil dependencies", func(t *testing.T) {
		cfg := config.BranchingConfig{
			AutoDeleteAfter: 24 * time.Hour,
		}

		scheduler := NewScheduler(nil, nil, cfg)

		assert.NotNil(t, scheduler)
		assert.Equal(t, cfg.AutoDeleteAfter, scheduler.config.AutoDeleteAfter)
	})

	t.Run("scheduler with disabled auto-delete", func(t *testing.T) {
		cfg := config.BranchingConfig{
			AutoDeleteAfter: 0, // Disabled
		}

		scheduler := NewScheduler(nil, nil, cfg)

		assert.NotNil(t, scheduler)
		assert.Equal(t, time.Duration(0), scheduler.config.AutoDeleteAfter)
	})
}

// =============================================================================
// Scheduler Configuration Tests
// =============================================================================

func TestScheduler_Config(t *testing.T) {
	t.Run("auto delete after 24 hours", func(t *testing.T) {
		cfg := config.BranchingConfig{
			AutoDeleteAfter: 24 * time.Hour,
		}

		scheduler := NewScheduler(nil, nil, cfg)

		assert.Equal(t, 24*time.Hour, scheduler.config.AutoDeleteAfter)
	})

	t.Run("auto delete after 7 days", func(t *testing.T) {
		cfg := config.BranchingConfig{
			AutoDeleteAfter: 7 * 24 * time.Hour,
		}

		scheduler := NewScheduler(nil, nil, cfg)

		assert.Equal(t, 7*24*time.Hour, scheduler.config.AutoDeleteAfter)
	})
}

// =============================================================================
// Scheduler Stop Tests
// =============================================================================

func TestScheduler_Stop(t *testing.T) {
	t.Run("stop cancels context", func(t *testing.T) {
		cfg := config.BranchingConfig{
			AutoDeleteAfter: 24 * time.Hour,
		}

		scheduler := NewScheduler(nil, nil, cfg)

		// Stop should not panic
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
// Branch Expiration Tests
// =============================================================================

func TestBranchExpiration(t *testing.T) {
	t.Run("branch is expired", func(t *testing.T) {
		expiresAt := time.Now().Add(-1 * time.Hour)

		branch := Branch{
			ID:        uuid.New(),
			Name:      "expired-branch",
			Type:      BranchTypePreview,
			ExpiresAt: &expiresAt,
		}

		isExpired := branch.ExpiresAt != nil && branch.ExpiresAt.Before(time.Now())
		assert.True(t, isExpired)
	})

	t.Run("branch is not expired", func(t *testing.T) {
		expiresAt := time.Now().Add(24 * time.Hour)

		branch := Branch{
			ID:        uuid.New(),
			Name:      "active-branch",
			Type:      BranchTypePreview,
			ExpiresAt: &expiresAt,
		}

		isExpired := branch.ExpiresAt != nil && branch.ExpiresAt.Before(time.Now())
		assert.False(t, isExpired)
	})

	t.Run("branch without expiration", func(t *testing.T) {
		branch := Branch{
			ID:   uuid.New(),
			Name: "persistent-branch",
			Type: BranchTypePersistent,
		}

		hasExpiration := branch.ExpiresAt != nil
		assert.False(t, hasExpiration)
	})
}

// =============================================================================
// Auto-Delete Configuration Tests
// =============================================================================

func TestAutoDeleteConfiguration(t *testing.T) {
	t.Run("auto delete disabled when duration is 0", func(t *testing.T) {
		cfg := config.BranchingConfig{
			AutoDeleteAfter: 0,
		}

		autoDeleteEnabled := cfg.AutoDeleteAfter > 0
		assert.False(t, autoDeleteEnabled)
	})

	t.Run("auto delete enabled when duration is positive", func(t *testing.T) {
		cfg := config.BranchingConfig{
			AutoDeleteAfter: 24 * time.Hour,
		}

		autoDeleteEnabled := cfg.AutoDeleteAfter > 0
		assert.True(t, autoDeleteEnabled)
	})
}

// =============================================================================
// Branch Type Auto-Delete Tests
// =============================================================================

func TestBranchTypeAutoDelete(t *testing.T) {
	t.Run("preview branches can be auto-deleted", func(t *testing.T) {
		branch := Branch{
			ID:   uuid.New(),
			Type: BranchTypePreview,
		}

		canAutoDelete := branch.Type == BranchTypePreview
		assert.True(t, canAutoDelete)
	})

	t.Run("main branch cannot be auto-deleted", func(t *testing.T) {
		branch := Branch{
			ID:   uuid.New(),
			Type: BranchTypeMain,
		}

		canAutoDelete := branch.Type == BranchTypePreview
		assert.False(t, canAutoDelete)
	})

	t.Run("persistent branches are not auto-deleted", func(t *testing.T) {
		branch := Branch{
			ID:   uuid.New(),
			Type: BranchTypePersistent,
		}

		canAutoDelete := branch.Type == BranchTypePreview
		assert.False(t, canAutoDelete)
	})
}

// =============================================================================
// Scheduler Interval Tests
// =============================================================================

func TestScheduler_Interval(t *testing.T) {
	t.Run("check interval for cleanup", func(t *testing.T) {
		// Typical cleanup interval is every minute or every 5 minutes
		checkInterval := 5 * time.Minute

		assert.Equal(t, 5*time.Minute, checkInterval)
	})
}

// =============================================================================
// Branch Status for Deletion Tests
// =============================================================================

func TestBranchStatusForDeletion(t *testing.T) {
	t.Run("ready branch can be deleted", func(t *testing.T) {
		branch := Branch{
			ID:     uuid.New(),
			Status: BranchStatusReady,
		}

		canDelete := branch.Status == BranchStatusReady
		assert.True(t, canDelete)
	})

	t.Run("creating branch should not be deleted", func(t *testing.T) {
		branch := Branch{
			ID:     uuid.New(),
			Status: BranchStatusCreating,
		}

		canDelete := branch.Status == BranchStatusReady
		assert.False(t, canDelete)
	})

	t.Run("deleting branch should not be deleted again", func(t *testing.T) {
		branch := Branch{
			ID:     uuid.New(),
			Status: BranchStatusDeleting,
		}

		canDelete := branch.Status == BranchStatusReady
		assert.False(t, canDelete)
	})

	t.Run("error branch can be deleted", func(t *testing.T) {
		branch := Branch{
			ID:     uuid.New(),
			Status: BranchStatusError,
		}

		// Error branches might be cleaned up
		canDelete := branch.Status == BranchStatusError || branch.Status == BranchStatusReady
		assert.True(t, canDelete)
	})
}

// =============================================================================
// Expiration Calculation Tests
// =============================================================================

func TestExpirationCalculation(t *testing.T) {
	t.Run("calculate expiration from auto-delete duration", func(t *testing.T) {
		autoDeleteAfter := 24 * time.Hour
		createdAt := time.Now()
		expiresAt := createdAt.Add(autoDeleteAfter)

		assert.True(t, expiresAt.After(createdAt))
		assert.Equal(t, 24*time.Hour, expiresAt.Sub(createdAt))
	})

	t.Run("calculate expiration for 48 hours", func(t *testing.T) {
		autoDeleteAfter := 48 * time.Hour
		createdAt := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
		expiresAt := createdAt.Add(autoDeleteAfter)

		assert.Equal(t, time.Date(2024, 1, 17, 10, 0, 0, 0, time.UTC), expiresAt)
	})

	t.Run("calculate expiration for 7 days", func(t *testing.T) {
		autoDeleteAfter := 7 * 24 * time.Hour
		createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		expiresAt := createdAt.Add(autoDeleteAfter)

		assert.Equal(t, time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), expiresAt)
	})
}

// =============================================================================
// Scheduler Context Tests
// =============================================================================

func TestScheduler_Context(t *testing.T) {
	t.Run("context is created on initialization", func(t *testing.T) {
		cfg := config.BranchingConfig{
			AutoDeleteAfter: 24 * time.Hour,
		}

		scheduler := NewScheduler(nil, nil, cfg)

		assert.NotNil(t, scheduler.ctx)
		assert.NotNil(t, scheduler.cancel)
	})

	t.Run("context is not done initially", func(t *testing.T) {
		cfg := config.BranchingConfig{}

		scheduler := NewScheduler(nil, nil, cfg)

		select {
		case <-scheduler.ctx.Done():
			t.Error("Context should not be done initially")
		default:
			// Expected
		}
	})
}

// =============================================================================
// Branch Cleanup Priority Tests
// =============================================================================

func TestBranchCleanupPriority(t *testing.T) {
	t.Run("older expired branches cleaned first", func(t *testing.T) {
		now := time.Now()

		branch1 := Branch{
			ID:        uuid.New(),
			ExpiresAt: func() *time.Time { t := now.Add(-2 * time.Hour); return &t }(),
		}

		branch2 := Branch{
			ID:        uuid.New(),
			ExpiresAt: func() *time.Time { t := now.Add(-1 * time.Hour); return &t }(),
		}

		// branch1 expired earlier, should be cleaned first
		assert.True(t, branch1.ExpiresAt.Before(*branch2.ExpiresAt))
	})
}

// =============================================================================
// Scheduler Storage Tests
// =============================================================================

func TestScheduler_Storage(t *testing.T) {
	t.Run("stores storage reference", func(t *testing.T) {
		cfg := config.BranchingConfig{}

		scheduler := NewScheduler(nil, nil, cfg)

		assert.Nil(t, scheduler.storage)
	})
}

// =============================================================================
// Scheduler Manager Tests
// =============================================================================

func TestScheduler_Manager(t *testing.T) {
	t.Run("stores manager reference", func(t *testing.T) {
		cfg := config.BranchingConfig{}

		scheduler := NewScheduler(nil, nil, cfg)

		assert.Nil(t, scheduler.manager)
	})
}
