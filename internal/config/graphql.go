package config

import "fmt"

// GraphQLConfig contains GraphQL API settings
type GraphQLConfig struct {
	Enabled         bool `mapstructure:"enabled"`            // Enable GraphQL API endpoint
	MaxDepth        int  `mapstructure:"max_depth"`          // Maximum query depth (default: 10)
	MaxComplexity   int  `mapstructure:"max_complexity"`     // Maximum query complexity score (default: 1000)
	Introspection   bool `mapstructure:"introspection"`      // Enable GraphQL introspection (default: true in dev, false in prod)
	AllowFragments  bool `mapstructure:"allow_fragments"`    // Allow fragment spreads in queries (default: false for security)
	MaxFieldsPerLvl int  `mapstructure:"max_fields_per_lvl"` // Maximum unique fields per query level (default: 50)
}

// Validate validates GraphQL configuration
func (gc *GraphQLConfig) Validate() error {
	if !gc.Enabled {
		return nil // No validation needed if disabled
	}

	if gc.MaxDepth < 1 {
		return fmt.Errorf("graphql max_depth must be at least 1, got: %d", gc.MaxDepth)
	}

	if gc.MaxComplexity < 1 {
		return fmt.Errorf("graphql max_complexity must be at least 1, got: %d", gc.MaxComplexity)
	}

	if gc.MaxFieldsPerLvl < 1 {
		return fmt.Errorf("graphql max_fields_per_lvl must be at least 1, got: %d", gc.MaxFieldsPerLvl)
	}

	return nil
}
