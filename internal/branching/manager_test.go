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
		assert.Equal(t, "branch_my-branch", name)
	})

	t.Run("without prefix", func(t *testing.T) {
		name := GenerateDatabaseName("", "my-branch")
		assert.Equal(t, "my-branch", name)
	})

	t.Run("custom prefix", func(t *testing.T) {
		name := GenerateDatabaseName("fluxbase_", "feature-123")
		assert.Equal(t, "fluxbase_feature-123", name)
	})
}

// =============================================================================
// CreateBranchRequest Tests
// =============================================================================

func TestCreateBranchRequest_Struct(t *testing.T) {
	t.Run("minimal request", func(t *testing.T) {
		req := CreateBranchRequest{
			Name: "feature-branch",
		}

		assert.Equal(t, "feature-branch", req.Name)
		assert.Nil(t, req.ParentBranchID)
		assert.Empty(t, req.DataCloneMode)
		assert.Empty(t, req.Type)
	})

	t.Run("full request", func(t *testing.T) {
		parentID := uuid.New()
		expiresAt := time.Now().Add(24 * time.Hour)
		prNumber := 123
		prURL := "https://github.com/org/repo/pull/123"
		repo := "org/repo"

		req := CreateBranchRequest{
			Name:           "pr-123-feature",
			ParentBranchID: &parentID,
			DataCloneMode:  DataCloneModeFullClone,
			Type:           BranchTypePreview,
			GitHubPRNumber: &prNumber,
			GitHubPRURL:    &prURL,
			GitHubRepo:     &repo,
			ExpiresAt:      &expiresAt,
		}

		assert.Equal(t, "pr-123-feature", req.Name)
		assert.Equal(t, parentID, *req.ParentBranchID)
		assert.Equal(t, DataCloneModeFullClone, req.DataCloneMode)
		assert.Equal(t, BranchTypePreview, req.Type)
		assert.Equal(t, 123, *req.GitHubPRNumber)
		assert.Equal(t, "https://github.com/org/repo/pull/123", *req.GitHubPRURL)
		assert.NotNil(t, req.ExpiresAt)
	})
}

// =============================================================================
// Branch Struct Tests
// =============================================================================

func TestBranch_Struct(t *testing.T) {
	t.Run("main branch", func(t *testing.T) {
		branch := Branch{
			ID:           uuid.New(),
			Name:         "main",
			Slug:         "main",
			DatabaseName: "fluxbase",
			Status:       BranchStatusReady,
			Type:         BranchTypeMain,
			CreatedAt:    time.Now(),
		}

		assert.Equal(t, "main", branch.Name)
		assert.Equal(t, BranchTypeMain, branch.Type)
		assert.Equal(t, BranchStatusReady, branch.Status)
		assert.True(t, branch.IsMain())
	})

	t.Run("preview branch", func(t *testing.T) {
		parentID := uuid.New()

		branch := Branch{
			ID:             uuid.New(),
			Name:           "PR #123",
			Slug:           "pr-123",
			DatabaseName:   "branch_pr-123",
			Status:         BranchStatusReady,
			Type:           BranchTypePreview,
			ParentBranchID: &parentID,
			DataCloneMode:  DataCloneModeSchemaOnly,
			CreatedAt:      time.Now(),
		}

		assert.Equal(t, "PR #123", branch.Name)
		assert.Equal(t, BranchTypePreview, branch.Type)
		assert.False(t, branch.IsMain())
		assert.NotNil(t, branch.ParentBranchID)
	})

	t.Run("branch with GitHub info", func(t *testing.T) {
		prNumber := 456
		prURL := "https://github.com/org/repo/pull/456"
		repo := "org/repo"

		branch := Branch{
			ID:             uuid.New(),
			Name:           "Feature Branch",
			Slug:           "feature-branch",
			DatabaseName:   "branch_feature-branch",
			Status:         BranchStatusReady,
			Type:           BranchTypePreview,
			GitHubPRNumber: &prNumber,
			GitHubPRURL:    &prURL,
			GitHubRepo:     &repo,
		}

		assert.Equal(t, 456, *branch.GitHubPRNumber)
		assert.Equal(t, "https://github.com/org/repo/pull/456", *branch.GitHubPRURL)
		assert.Equal(t, "org/repo", *branch.GitHubRepo)
	})
}

// =============================================================================
// Branch IsMain Tests
// =============================================================================

func TestBranch_IsMain(t *testing.T) {
	t.Run("main type is main", func(t *testing.T) {
		branch := Branch{Type: BranchTypeMain}
		assert.True(t, branch.IsMain())
	})

	t.Run("preview type is not main", func(t *testing.T) {
		branch := Branch{Type: BranchTypePreview}
		assert.False(t, branch.IsMain())
	})

	t.Run("persistent type is not main", func(t *testing.T) {
		branch := Branch{Type: BranchTypePersistent}
		assert.False(t, branch.IsMain())
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
// DataCloneMode Tests
// =============================================================================

func TestDataCloneMode_Constants(t *testing.T) {
	t.Run("schema only mode", func(t *testing.T) {
		assert.Equal(t, DataCloneMode("schema_only"), DataCloneModeSchemaOnly)
	})

	t.Run("full clone mode", func(t *testing.T) {
		assert.Equal(t, DataCloneMode("full_clone"), DataCloneModeFullClone)
	})

	t.Run("seed data mode", func(t *testing.T) {
		assert.Equal(t, DataCloneMode("seed_data"), DataCloneModeSeedData)
	})
}

// =============================================================================
// BranchType Tests
// =============================================================================

func TestBranchType_Constants(t *testing.T) {
	t.Run("main type", func(t *testing.T) {
		assert.Equal(t, BranchType("main"), BranchTypeMain)
	})

	t.Run("preview type", func(t *testing.T) {
		assert.Equal(t, BranchType("preview"), BranchTypePreview)
	})

	t.Run("persistent type", func(t *testing.T) {
		assert.Equal(t, BranchType("persistent"), BranchTypePersistent)
	})
}

// =============================================================================
// BranchStatus Tests
// =============================================================================

func TestBranchStatus_Constants(t *testing.T) {
	statuses := []struct {
		status   BranchStatus
		expected string
	}{
		{BranchStatusCreating, "creating"},
		{BranchStatusReady, "ready"},
		{BranchStatusDeleting, "deleting"},
		{BranchStatusError, "error"},
	}

	for _, tc := range statuses {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, BranchStatus(tc.expected), tc.status)
		})
	}
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

func TestBranchAccessRule_Struct(t *testing.T) {
	t.Run("read-only access", func(t *testing.T) {
		userID := uuid.New()
		branchID := uuid.New()

		rule := BranchAccessRule{
			ID:       uuid.New(),
			BranchID: branchID,
			UserID:   userID,
			CanRead:  true,
			CanWrite: false,
			CanAdmin: false,
		}

		assert.Equal(t, branchID, rule.BranchID)
		assert.Equal(t, userID, rule.UserID)
		assert.True(t, rule.CanRead)
		assert.False(t, rule.CanWrite)
		assert.False(t, rule.CanAdmin)
	})

	t.Run("full access", func(t *testing.T) {
		rule := BranchAccessRule{
			ID:       uuid.New(),
			BranchID: uuid.New(),
			UserID:   uuid.New(),
			CanRead:  true,
			CanWrite: true,
			CanAdmin: true,
		}

		assert.True(t, rule.CanRead)
		assert.True(t, rule.CanWrite)
		assert.True(t, rule.CanAdmin)
	})
}

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

func TestBranchConnectionInfo_Struct(t *testing.T) {
	t.Run("connection info", func(t *testing.T) {
		info := BranchConnectionInfo{
			Host:         "localhost",
			Port:         5432,
			DatabaseName: "branch_feature-123",
			Username:     "fluxbase_app",
		}

		assert.Equal(t, "localhost", info.Host)
		assert.Equal(t, 5432, info.Port)
		assert.Equal(t, "branch_feature-123", info.DatabaseName)
		assert.Equal(t, "fluxbase_app", info.Username)
	})
}

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
		assert.Equal(t, "dev_my-feature", dbName)
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
