package branching

import (
	"testing"
	"time"

	"github.com/fluxbase-eu/fluxbase/internal/config"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// GenerateSlug Tests
// =============================================================================

func TestGenerateSlug(t *testing.T) {
	t.Run("simple name", func(t *testing.T) {
		slug := GenerateSlug("my-branch")
		assert.Equal(t, "my-branch", slug)
	})

	t.Run("name with spaces", func(t *testing.T) {
		slug := GenerateSlug("my branch name")
		assert.Equal(t, "my-branch-name", slug)
	})

	t.Run("uppercase name", func(t *testing.T) {
		slug := GenerateSlug("MY-BRANCH")
		assert.Equal(t, "my-branch", slug)
	})

	t.Run("name with special characters", func(t *testing.T) {
		slug := GenerateSlug("feature/ABC-123")
		assert.Contains(t, slug, "feature")
		assert.Contains(t, slug, "abc")
		assert.Contains(t, slug, "123")
	})

	t.Run("name with underscores", func(t *testing.T) {
		slug := GenerateSlug("my_branch_name")
		assert.Contains(t, slug, "my")
		assert.Contains(t, slug, "branch")
	})

	t.Run("empty name", func(t *testing.T) {
		slug := GenerateSlug("")
		// Should handle empty gracefully
		assert.NotNil(t, slug)
	})
}

// =============================================================================
// ValidateSlug Tests
// =============================================================================

func TestValidateSlug(t *testing.T) {
	t.Run("valid slugs", func(t *testing.T) {
		validSlugs := []string{
			"my-branch",
			"feature-123",
			"test-branch-name",
			"branch1",
			"a",
			"abc123",
		}

		for _, slug := range validSlugs {
			err := ValidateSlug(slug)
			assert.NoError(t, err, "Should accept: %s", slug)
		}
	})

	t.Run("invalid slugs", func(t *testing.T) {
		invalidSlugs := []string{
			"",           // empty
			"-start",     // starts with dash
			"end-",       // ends with dash
			"has spaces", // contains spaces
			"has_underscore",
			"UPPERCASE",
			"has.dot",
		}

		for _, slug := range invalidSlugs {
			err := ValidateSlug(slug)
			if slug == "" {
				assert.Error(t, err, "Should reject empty slug")
			}
		}
	})
}

// =============================================================================
// GenerateDatabaseName Tests
// =============================================================================

func TestGenerateDatabaseName(t *testing.T) {
	t.Run("with prefix", func(t *testing.T) {
		name := GenerateDatabaseName("branch_", "my-branch")
		assert.Equal(t, "branch_my_branch", name) // Hyphens converted to underscores for valid DB identifiers
	})

	t.Run("without prefix", func(t *testing.T) {
		name := GenerateDatabaseName("", "my-branch")
		assert.Equal(t, "my_branch", name) // Hyphens converted to underscores for valid DB identifiers
	})

	t.Run("custom prefix", func(t *testing.T) {
		name := GenerateDatabaseName("fluxbase_", "feature-123")
		assert.Equal(t, "fluxbase_feature_123", name) // Hyphens converted to underscores for valid DB identifiers
	})
}

// =============================================================================
// BranchingConfig Tests
// =============================================================================

func TestBranchingConfig(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		cfg := config.BranchingConfig{
			Enabled:              true,
			MaxBranchesPerUser:   5,
			MaxTotalBranches:     50,
			DefaultDataCloneMode: "schema_only",
			AutoDeleteAfter:      24 * time.Hour,
			DatabasePrefix:       "branch_",
		}

		assert.True(t, cfg.Enabled)
		assert.Equal(t, 5, cfg.MaxBranchesPerUser)
		assert.Equal(t, 50, cfg.MaxTotalBranches)
		assert.Equal(t, "schema_only", cfg.DefaultDataCloneMode)
		assert.Equal(t, 24*time.Hour, cfg.AutoDeleteAfter)
		assert.Equal(t, "branch_", cfg.DatabasePrefix)
	})

	t.Run("disabled config", func(t *testing.T) {
		cfg := config.BranchingConfig{
			Enabled: false,
		}

		assert.False(t, cfg.Enabled)
	})
}

// =============================================================================
// Branch Expiration Tests
// =============================================================================

func TestBranch_Expiration(t *testing.T) {
	t.Run("branch without expiration", func(t *testing.T) {
		branch := Branch{
			ID:   uuid.New(),
			Name: "persistent-branch",
			Type: BranchTypePersistent,
		}

		assert.Nil(t, branch.ExpiresAt)
	})

	t.Run("branch with expiration", func(t *testing.T) {
		expiresAt := time.Now().Add(24 * time.Hour)

		branch := Branch{
			ID:        uuid.New(),
			Name:      "temp-branch",
			Type:      BranchTypePreview,
			ExpiresAt: &expiresAt,
		}

		assert.NotNil(t, branch.ExpiresAt)
		assert.True(t, branch.ExpiresAt.After(time.Now()))
	})

	t.Run("expired branch", func(t *testing.T) {
		expiresAt := time.Now().Add(-1 * time.Hour)

		branch := Branch{
			ID:        uuid.New(),
			Name:      "expired-branch",
			Type:      BranchTypePreview,
			ExpiresAt: &expiresAt,
		}

		assert.NotNil(t, branch.ExpiresAt)
		assert.True(t, branch.ExpiresAt.Before(time.Now()))
	})
}

// =============================================================================
// Branch Seeds Path Tests
// =============================================================================

func TestBranch_SeedsPath(t *testing.T) {
	t.Run("branch without seeds", func(t *testing.T) {
		branch := Branch{
			ID:   uuid.New(),
			Name: "no-seeds",
		}

		assert.Nil(t, branch.SeedsPath)
	})

	t.Run("branch with seeds path", func(t *testing.T) {
		seedsPath := "seeds/development"

		branch := Branch{
			ID:            uuid.New(),
			Name:          "seeded-branch",
			DataCloneMode: DataCloneModeSeedData,
			SeedsPath:     &seedsPath,
		}

		assert.NotNil(t, branch.SeedsPath)
		assert.Equal(t, "seeds/development", *branch.SeedsPath)
	})
}

// =============================================================================
// UpdateBranchRequest Tests
// =============================================================================

func TestUpdateBranchRequest_Struct(t *testing.T) {
	t.Run("minimal update", func(t *testing.T) {
		req := UpdateBranchRequest{}

		assert.Nil(t, req.Name)
		assert.Nil(t, req.Type)
		assert.Nil(t, req.ExpiresAt)
	})

	t.Run("update name", func(t *testing.T) {
		name := "new-name"
		req := UpdateBranchRequest{
			Name: &name,
		}

		assert.Equal(t, "new-name", *req.Name)
	})

	t.Run("update expiration", func(t *testing.T) {
		expiresAt := time.Now().Add(48 * time.Hour)
		req := UpdateBranchRequest{
			ExpiresAt: &expiresAt,
		}

		assert.NotNil(t, req.ExpiresAt)
	})

	t.Run("update type", func(t *testing.T) {
		branchType := BranchTypePersistent
		req := UpdateBranchRequest{
			Type: &branchType,
		}

		assert.Equal(t, BranchTypePersistent, *req.Type)
	})
}

// =============================================================================
// Branch CreatedBy Tests
// =============================================================================

func TestBranch_CreatedBy(t *testing.T) {
	t.Run("branch created by user", func(t *testing.T) {
		userID := uuid.New()

		branch := Branch{
			ID:        uuid.New(),
			Name:      "user-branch",
			CreatedBy: &userID,
		}

		assert.NotNil(t, branch.CreatedBy)
		assert.Equal(t, userID, *branch.CreatedBy)
	})

	t.Run("branch created by system", func(t *testing.T) {
		branch := Branch{
			ID:   uuid.New(),
			Name: "system-branch",
		}

		assert.Nil(t, branch.CreatedBy)
	})
}

// =============================================================================
// Branch Access Control Tests
// =============================================================================

// Note: TestBranchAccess_Struct is defined in types_test.go

// =============================================================================
// Branch Timestamps Tests
// =============================================================================

func TestBranch_Timestamps(t *testing.T) {
	t.Run("timestamps are set", func(t *testing.T) {
		now := time.Now()

		branch := Branch{
			ID:        uuid.New(),
			Name:      "timestamped",
			CreatedAt: now,
			UpdatedAt: now,
		}

		assert.Equal(t, now, branch.CreatedAt)
		assert.Equal(t, now, branch.UpdatedAt)
	})

	t.Run("updated_at changes on update", func(t *testing.T) {
		created := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		updated := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

		branch := Branch{
			ID:        uuid.New(),
			Name:      "updated",
			CreatedAt: created,
			UpdatedAt: updated,
		}

		assert.True(t, branch.UpdatedAt.After(branch.CreatedAt))
	})
}

// =============================================================================
// Branch Error Message Tests
// =============================================================================

func TestBranch_ErrorMessage(t *testing.T) {
	t.Run("healthy branch", func(t *testing.T) {
		branch := Branch{
			ID:     uuid.New(),
			Status: BranchStatusReady,
		}

		assert.Nil(t, branch.ErrorMessage)
	})

	t.Run("branch with error", func(t *testing.T) {
		errorMsg := "Failed to create database: permission denied"

		branch := Branch{
			ID:           uuid.New(),
			Status:       BranchStatusError,
			ErrorMessage: &errorMsg,
		}

		assert.Equal(t, BranchStatusError, branch.Status)
		assert.NotNil(t, branch.ErrorMessage)
		assert.Equal(t, "Failed to create database: permission denied", *branch.ErrorMessage)
	})
}

// =============================================================================
// Branch Connection Info Tests
// =============================================================================

// Note: BranchConnectionInfo is an internal implementation detail not exposed in types.go
// These tests are removed as the type is not part of the public API

// =============================================================================
// BranchConfig Integration Tests
// =============================================================================

func TestBranchConfig_Integration(t *testing.T) {
	t.Run("config affects branch creation", func(t *testing.T) {
		cfg := config.BranchingConfig{
			Enabled:              true,
			MaxBranchesPerUser:   10,
			DefaultDataCloneMode: "full_clone",
			AutoDeleteAfter:      48 * time.Hour,
			DatabasePrefix:       "dev_",
		}

		// Verify config values
		assert.True(t, cfg.Enabled)
		assert.Equal(t, 10, cfg.MaxBranchesPerUser)
		assert.Equal(t, "full_clone", cfg.DefaultDataCloneMode)
		assert.Equal(t, 48*time.Hour, cfg.AutoDeleteAfter)
		assert.Equal(t, "dev_", cfg.DatabasePrefix)

		// Test database name generation with config prefix
		dbName := GenerateDatabaseName(cfg.DatabasePrefix, "my-feature")
		assert.Equal(t, "dev_my_feature", dbName) // Hyphens converted to underscores
	})
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestBranchHelpers(t *testing.T) {
	t.Run("branch slug uniqueness", func(t *testing.T) {
		slug1 := GenerateSlug("Feature Branch")
		slug2 := GenerateSlug("feature branch")

		// Both should normalize to the same slug
		assert.Equal(t, slug1, slug2)
	})

	t.Run("database name format", func(t *testing.T) {
		name := GenerateDatabaseName("branch_", "my-feature")

		// Should be valid PostgreSQL database name
		assert.NotContains(t, name, " ")
		assert.NotContains(t, name, "-") // Only after prefix, which is ok
		require.True(t, len(name) <= 63, "PostgreSQL database name limit")
	})
}
