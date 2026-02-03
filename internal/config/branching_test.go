package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// DataCloneMode Constants Tests
// =============================================================================

func TestBranching_DataCloneModeConstants(t *testing.T) {
	t.Run("schema_only constant value", func(t *testing.T) {
		assert.Equal(t, "schema_only", DataCloneModeSchemaOnly)
	})

	t.Run("full_clone constant value", func(t *testing.T) {
		assert.Equal(t, "full_clone", DataCloneModeFullClone)
	})

	t.Run("seed_data constant value", func(t *testing.T) {
		assert.Equal(t, "seed_data", DataCloneModeSeedData)
	})
}

// =============================================================================
// BranchingConfig.Validate Tests
// =============================================================================

func TestBranching_ConfigValidate(t *testing.T) {
	t.Run("returns nil when disabled", func(t *testing.T) {
		config := &BranchingConfig{
			Enabled: false,
		}

		err := config.Validate()

		assert.NoError(t, err)
	})

	t.Run("returns nil for valid config", func(t *testing.T) {
		config := &BranchingConfig{
			Enabled:              true,
			MaxTotalBranches:     50,
			MaxBranchesPerUser:   5,
			DefaultDataCloneMode: DataCloneModeSchemaOnly,
			AutoDeleteAfter:      24 * time.Hour,
			DatabasePrefix:       "branch_",
			SeedsPath:            "./seeds",
			DefaultBranch:        "main",
		}

		err := config.Validate()

		assert.NoError(t, err)
	})

	t.Run("returns error for negative max_total_branches", func(t *testing.T) {
		config := &BranchingConfig{
			Enabled:          true,
			MaxTotalBranches: -1,
			DatabasePrefix:   "branch_",
		}

		err := config.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "max_total_branches cannot be negative")
	})

	t.Run("allows zero max_total_branches", func(t *testing.T) {
		config := &BranchingConfig{
			Enabled:          true,
			MaxTotalBranches: 0,
			DatabasePrefix:   "branch_",
		}

		err := config.Validate()

		assert.NoError(t, err)
	})

	t.Run("returns error for negative max_branches_per_user", func(t *testing.T) {
		config := &BranchingConfig{
			Enabled:            true,
			MaxBranchesPerUser: -5,
			DatabasePrefix:     "branch_",
		}

		err := config.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "max_branches_per_user cannot be negative")
	})

	t.Run("allows zero max_branches_per_user", func(t *testing.T) {
		config := &BranchingConfig{
			Enabled:            true,
			MaxBranchesPerUser: 0,
			DatabasePrefix:     "branch_",
		}

		err := config.Validate()

		assert.NoError(t, err)
	})

	t.Run("returns error for negative auto_delete_after", func(t *testing.T) {
		config := &BranchingConfig{
			Enabled:         true,
			AutoDeleteAfter: -time.Hour,
			DatabasePrefix:  "branch_",
		}

		err := config.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "auto_delete_after cannot be negative")
	})

	t.Run("allows zero auto_delete_after", func(t *testing.T) {
		config := &BranchingConfig{
			Enabled:         true,
			AutoDeleteAfter: 0,
			DatabasePrefix:  "branch_",
		}

		err := config.Validate()

		assert.NoError(t, err)
	})

	t.Run("returns error for empty database_prefix when enabled", func(t *testing.T) {
		config := &BranchingConfig{
			Enabled:        true,
			DatabasePrefix: "",
		}

		err := config.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "database_prefix cannot be empty")
	})
}

// =============================================================================
// BranchingConfig DataCloneMode Validation Tests
// =============================================================================

func TestBranchingConfig_Validate_DataCloneMode(t *testing.T) {
	t.Run("accepts schema_only", func(t *testing.T) {
		config := &BranchingConfig{
			Enabled:              true,
			DefaultDataCloneMode: DataCloneModeSchemaOnly,
			DatabasePrefix:       "branch_",
		}

		err := config.Validate()

		assert.NoError(t, err)
	})

	t.Run("accepts full_clone", func(t *testing.T) {
		config := &BranchingConfig{
			Enabled:              true,
			DefaultDataCloneMode: DataCloneModeFullClone,
			DatabasePrefix:       "branch_",
		}

		err := config.Validate()

		assert.NoError(t, err)
	})

	t.Run("accepts seed_data", func(t *testing.T) {
		config := &BranchingConfig{
			Enabled:              true,
			DefaultDataCloneMode: DataCloneModeSeedData,
			DatabasePrefix:       "branch_",
		}

		err := config.Validate()

		assert.NoError(t, err)
	})

	t.Run("accepts empty string (defaults to schema_only)", func(t *testing.T) {
		config := &BranchingConfig{
			Enabled:              true,
			DefaultDataCloneMode: "",
			DatabasePrefix:       "branch_",
		}

		err := config.Validate()

		assert.NoError(t, err)
	})

	t.Run("returns error for invalid mode", func(t *testing.T) {
		config := &BranchingConfig{
			Enabled:              true,
			DefaultDataCloneMode: "invalid_mode",
			DatabasePrefix:       "branch_",
		}

		err := config.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "default_data_clone_mode must be one of")
		assert.Contains(t, err.Error(), DataCloneModeSchemaOnly)
		assert.Contains(t, err.Error(), DataCloneModeFullClone)
		assert.Contains(t, err.Error(), DataCloneModeSeedData)
	})

	t.Run("returns error for typo in mode", func(t *testing.T) {
		config := &BranchingConfig{
			Enabled:              true,
			DefaultDataCloneMode: "schema-only", // hyphen instead of underscore
			DatabasePrefix:       "branch_",
		}

		err := config.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "default_data_clone_mode must be one of")
	})

	t.Run("returns error for case-sensitive mode", func(t *testing.T) {
		config := &BranchingConfig{
			Enabled:              true,
			DefaultDataCloneMode: "Schema_Only", // wrong case
			DatabasePrefix:       "branch_",
		}

		err := config.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "default_data_clone_mode must be one of")
	})
}

// =============================================================================
// BranchingConfig SeedsPath Default Tests
// =============================================================================

func TestBranchingConfig_Validate_SeedsPath(t *testing.T) {
	t.Run("sets default seeds_path when empty", func(t *testing.T) {
		config := &BranchingConfig{
			Enabled:        true,
			SeedsPath:      "",
			DatabasePrefix: "branch_",
		}

		err := config.Validate()

		assert.NoError(t, err)
		assert.Equal(t, "./seeds", config.SeedsPath)
	})

	t.Run("preserves non-empty seeds_path", func(t *testing.T) {
		config := &BranchingConfig{
			Enabled:        true,
			SeedsPath:      "/custom/path/seeds",
			DatabasePrefix: "branch_",
		}

		err := config.Validate()

		assert.NoError(t, err)
		assert.Equal(t, "/custom/path/seeds", config.SeedsPath)
	})

	t.Run("does not set seeds_path when disabled", func(t *testing.T) {
		config := &BranchingConfig{
			Enabled:   false,
			SeedsPath: "",
		}

		err := config.Validate()

		assert.NoError(t, err)
		assert.Empty(t, config.SeedsPath)
	})
}

// =============================================================================
// BranchingConfig Struct Tests
// =============================================================================

func TestBranchingConfig_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		config := BranchingConfig{
			Enabled:              true,
			MaxTotalBranches:     100,
			MaxBranchesPerUser:   10,
			DefaultDataCloneMode: DataCloneModeFullClone,
			AutoDeleteAfter:      48 * time.Hour,
			DatabasePrefix:       "dev_branch_",
			SeedsPath:            "/data/seeds",
			DefaultBranch:        "develop",
		}

		assert.True(t, config.Enabled)
		assert.Equal(t, 100, config.MaxTotalBranches)
		assert.Equal(t, 10, config.MaxBranchesPerUser)
		assert.Equal(t, DataCloneModeFullClone, config.DefaultDataCloneMode)
		assert.Equal(t, 48*time.Hour, config.AutoDeleteAfter)
		assert.Equal(t, "dev_branch_", config.DatabasePrefix)
		assert.Equal(t, "/data/seeds", config.SeedsPath)
		assert.Equal(t, "develop", config.DefaultBranch)
	})

	t.Run("defaults to zero values", func(t *testing.T) {
		config := BranchingConfig{}

		assert.False(t, config.Enabled)
		assert.Zero(t, config.MaxTotalBranches)
		assert.Zero(t, config.MaxBranchesPerUser)
		assert.Empty(t, config.DefaultDataCloneMode)
		assert.Zero(t, config.AutoDeleteAfter)
		assert.Empty(t, config.DatabasePrefix)
		assert.Empty(t, config.SeedsPath)
		assert.Empty(t, config.DefaultBranch)
	})
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func TestBranchingConfig_EdgeCases(t *testing.T) {
	t.Run("validation order: negative values before prefix", func(t *testing.T) {
		config := &BranchingConfig{
			Enabled:          true,
			MaxTotalBranches: -10, // First error
			DatabasePrefix:   "",  // Would also fail
		}

		err := config.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "max_total_branches")
	})

	t.Run("validates with very large branch limits", func(t *testing.T) {
		config := &BranchingConfig{
			Enabled:            true,
			MaxTotalBranches:   10000,
			MaxBranchesPerUser: 1000,
			DatabasePrefix:     "branch_",
		}

		err := config.Validate()

		assert.NoError(t, err)
	})

	t.Run("validates with very long duration", func(t *testing.T) {
		config := &BranchingConfig{
			Enabled:         true,
			AutoDeleteAfter: time.Hour * 24 * 365, // 1 year
			DatabasePrefix:  "branch_",
		}

		err := config.Validate()

		assert.NoError(t, err)
	})

	t.Run("validates with whitespace prefix", func(t *testing.T) {
		config := &BranchingConfig{
			Enabled:        true,
			DatabasePrefix: "   ", // whitespace only - still not empty
		}

		err := config.Validate()

		assert.NoError(t, err)
	})

	t.Run("validates full production config", func(t *testing.T) {
		config := &BranchingConfig{
			Enabled:              true,
			MaxTotalBranches:     50,
			MaxBranchesPerUser:   5,
			DefaultDataCloneMode: DataCloneModeSchemaOnly,
			AutoDeleteAfter:      24 * time.Hour,
			DatabasePrefix:       "branch_",
			SeedsPath:            "./seeds",
			DefaultBranch:        "main",
		}

		err := config.Validate()

		assert.NoError(t, err)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkBranchingConfig_Validate_Disabled(b *testing.B) {
	config := &BranchingConfig{
		Enabled: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}

func BenchmarkBranchingConfig_Validate_Enabled(b *testing.B) {
	config := &BranchingConfig{
		Enabled:              true,
		MaxTotalBranches:     50,
		MaxBranchesPerUser:   5,
		DefaultDataCloneMode: DataCloneModeSchemaOnly,
		AutoDeleteAfter:      24 * time.Hour,
		DatabasePrefix:       "branch_",
		SeedsPath:            "./seeds",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}
