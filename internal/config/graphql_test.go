package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGraphQLConfig_Validate(t *testing.T) {
	t.Run("disabled config needs no validation", func(t *testing.T) {
		cfg := GraphQLConfig{
			Enabled:       false,
			MaxDepth:      0,
			MaxComplexity: 0,
		}

		err := cfg.Validate()
		require.NoError(t, err)
	})

	t.Run("valid enabled config passes", func(t *testing.T) {
		cfg := GraphQLConfig{
			Enabled:       true,
			MaxDepth:      10,
			MaxComplexity: 1000,
			Introspection: true,
		}

		err := cfg.Validate()
		require.NoError(t, err)
	})

	t.Run("minimum valid values pass", func(t *testing.T) {
		cfg := GraphQLConfig{
			Enabled:       true,
			MaxDepth:      1,
			MaxComplexity: 1,
		}

		err := cfg.Validate()
		require.NoError(t, err)
	})

	t.Run("rejects zero max_depth when enabled", func(t *testing.T) {
		cfg := GraphQLConfig{
			Enabled:       true,
			MaxDepth:      0,
			MaxComplexity: 1000,
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "max_depth must be at least 1")
	})

	t.Run("rejects negative max_depth when enabled", func(t *testing.T) {
		cfg := GraphQLConfig{
			Enabled:       true,
			MaxDepth:      -5,
			MaxComplexity: 1000,
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "max_depth must be at least 1")
	})

	t.Run("rejects zero max_complexity when enabled", func(t *testing.T) {
		cfg := GraphQLConfig{
			Enabled:       true,
			MaxDepth:      10,
			MaxComplexity: 0,
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "max_complexity must be at least 1")
	})

	t.Run("rejects negative max_complexity when enabled", func(t *testing.T) {
		cfg := GraphQLConfig{
			Enabled:       true,
			MaxDepth:      10,
			MaxComplexity: -100,
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "max_complexity must be at least 1")
	})

	t.Run("error message includes actual value for max_depth", func(t *testing.T) {
		cfg := GraphQLConfig{
			Enabled:       true,
			MaxDepth:      -42,
			MaxComplexity: 1000,
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "-42")
	})

	t.Run("error message includes actual value for max_complexity", func(t *testing.T) {
		cfg := GraphQLConfig{
			Enabled:       true,
			MaxDepth:      10,
			MaxComplexity: -500,
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "-500")
	})
}

func TestGraphQLConfig_Struct(t *testing.T) {
	t.Run("all fields can be set", func(t *testing.T) {
		cfg := GraphQLConfig{
			Enabled:       true,
			MaxDepth:      15,
			MaxComplexity: 2000,
			Introspection: false,
		}

		assert.True(t, cfg.Enabled)
		assert.Equal(t, 15, cfg.MaxDepth)
		assert.Equal(t, 2000, cfg.MaxComplexity)
		assert.False(t, cfg.Introspection)
	})

	t.Run("zero value has expected defaults", func(t *testing.T) {
		var cfg GraphQLConfig

		assert.False(t, cfg.Enabled)
		assert.Zero(t, cfg.MaxDepth)
		assert.Zero(t, cfg.MaxComplexity)
		assert.False(t, cfg.Introspection)
	})
}

func TestGraphQLConfig_CommonConfigurations(t *testing.T) {
	t.Run("development configuration", func(t *testing.T) {
		cfg := GraphQLConfig{
			Enabled:       true,
			MaxDepth:      20,
			MaxComplexity: 5000,
			Introspection: true, // Enabled in development
		}

		err := cfg.Validate()
		require.NoError(t, err)
		assert.True(t, cfg.Introspection)
	})

	t.Run("production configuration", func(t *testing.T) {
		cfg := GraphQLConfig{
			Enabled:       true,
			MaxDepth:      10,
			MaxComplexity: 1000,
			Introspection: false, // Disabled in production
		}

		err := cfg.Validate()
		require.NoError(t, err)
		assert.False(t, cfg.Introspection)
	})

	t.Run("strict configuration", func(t *testing.T) {
		cfg := GraphQLConfig{
			Enabled:       true,
			MaxDepth:      5,   // Very limited depth
			MaxComplexity: 100, // Very limited complexity
			Introspection: false,
		}

		err := cfg.Validate()
		require.NoError(t, err)
	})
}
