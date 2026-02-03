package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// DefaultMCPOAuthRedirectURIs Tests
// =============================================================================

func TestMCP_DefaultOAuthRedirectURIs(t *testing.T) {
	t.Run("returns non-empty list", func(t *testing.T) {
		uris := DefaultMCPOAuthRedirectURIs()

		require.NotNil(t, uris)
		assert.NotEmpty(t, uris)
	})

	t.Run("includes Claude Desktop URIs", func(t *testing.T) {
		uris := DefaultMCPOAuthRedirectURIs()

		assert.Contains(t, uris, "https://claude.ai/api/mcp/auth_callback")
		assert.Contains(t, uris, "https://claude.com/api/mcp/auth_callback")
	})

	t.Run("includes Cursor URIs", func(t *testing.T) {
		uris := DefaultMCPOAuthRedirectURIs()

		assert.Contains(t, uris, "cursor://anysphere.cursor-mcp/oauth/*/callback")
		assert.Contains(t, uris, "cursor://")
	})

	t.Run("includes VS Code URIs", func(t *testing.T) {
		uris := DefaultMCPOAuthRedirectURIs()

		assert.Contains(t, uris, "http://127.0.0.1:33418")
		assert.Contains(t, uris, "https://vscode.dev/redirect")
		assert.Contains(t, uris, "vscode://")
	})

	t.Run("includes OpenCode URI", func(t *testing.T) {
		uris := DefaultMCPOAuthRedirectURIs()

		assert.Contains(t, uris, "http://127.0.0.1:19876/mcp/oauth/callback")
	})

	t.Run("includes MCP Inspector URI", func(t *testing.T) {
		uris := DefaultMCPOAuthRedirectURIs()

		assert.Contains(t, uris, "http://localhost:6274/oauth/callback")
	})

	t.Run("includes ChatGPT URI", func(t *testing.T) {
		uris := DefaultMCPOAuthRedirectURIs()

		assert.Contains(t, uris, "https://chatgpt.com/connector_platform_oauth_redirect")
	})

	t.Run("includes localhost wildcards", func(t *testing.T) {
		uris := DefaultMCPOAuthRedirectURIs()

		assert.Contains(t, uris, "http://localhost:*")
		assert.Contains(t, uris, "http://127.0.0.1:*")
	})

	t.Run("returns consistent results", func(t *testing.T) {
		uris1 := DefaultMCPOAuthRedirectURIs()
		uris2 := DefaultMCPOAuthRedirectURIs()

		assert.Equal(t, uris1, uris2)
	})
}

// =============================================================================
// MCPConfig.Validate Tests
// =============================================================================

func TestMCP_ConfigValidate(t *testing.T) {
	t.Run("returns nil when disabled", func(t *testing.T) {
		config := &MCPConfig{
			Enabled: false,
		}

		err := config.Validate()

		assert.NoError(t, err)
	})

	t.Run("returns nil for valid config", func(t *testing.T) {
		config := &MCPConfig{
			Enabled:         true,
			BasePath:        "/mcp",
			SessionTimeout:  time.Hour,
			MaxMessageSize:  1024 * 1024,
			RateLimitPerMin: 100,
		}

		err := config.Validate()

		assert.NoError(t, err)
	})

	t.Run("returns error for empty base_path", func(t *testing.T) {
		config := &MCPConfig{
			Enabled:  true,
			BasePath: "",
		}

		err := config.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "base_path cannot be empty")
	})

	t.Run("returns error for negative session_timeout", func(t *testing.T) {
		config := &MCPConfig{
			Enabled:        true,
			BasePath:       "/mcp",
			SessionTimeout: -time.Hour,
		}

		err := config.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "session_timeout cannot be negative")
	})

	t.Run("allows zero session_timeout", func(t *testing.T) {
		config := &MCPConfig{
			Enabled:        true,
			BasePath:       "/mcp",
			SessionTimeout: 0,
		}

		err := config.Validate()

		assert.NoError(t, err)
	})

	t.Run("returns error for negative max_message_size", func(t *testing.T) {
		config := &MCPConfig{
			Enabled:        true,
			BasePath:       "/mcp",
			MaxMessageSize: -1,
		}

		err := config.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "max_message_size cannot be negative")
	})

	t.Run("allows zero max_message_size", func(t *testing.T) {
		config := &MCPConfig{
			Enabled:        true,
			BasePath:       "/mcp",
			MaxMessageSize: 0,
		}

		err := config.Validate()

		assert.NoError(t, err)
	})

	t.Run("returns error for negative rate_limit_per_min", func(t *testing.T) {
		config := &MCPConfig{
			Enabled:         true,
			BasePath:        "/mcp",
			RateLimitPerMin: -10,
		}

		err := config.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "rate_limit_per_min cannot be negative")
	})

	t.Run("allows zero rate_limit_per_min", func(t *testing.T) {
		config := &MCPConfig{
			Enabled:         true,
			BasePath:        "/mcp",
			RateLimitPerMin: 0,
		}

		err := config.Validate()

		assert.NoError(t, err)
	})
}

// =============================================================================
// MCPConfig OAuth Validation Tests
// =============================================================================

func TestMCPConfig_Validate_OAuth(t *testing.T) {
	t.Run("skips OAuth validation when disabled", func(t *testing.T) {
		config := &MCPConfig{
			Enabled:  true,
			BasePath: "/mcp",
			OAuth: MCPOAuthConfig{
				Enabled:            false,
				TokenExpiry:        -time.Hour, // Would fail if validated
				RefreshTokenExpiry: -time.Hour, // Would fail if validated
			},
		}

		err := config.Validate()

		assert.NoError(t, err)
	})

	t.Run("returns error for negative token_expiry", func(t *testing.T) {
		config := &MCPConfig{
			Enabled:  true,
			BasePath: "/mcp",
			OAuth: MCPOAuthConfig{
				Enabled:     true,
				TokenExpiry: -time.Hour,
			},
		}

		err := config.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "token_expiry cannot be negative")
	})

	t.Run("allows zero token_expiry", func(t *testing.T) {
		config := &MCPConfig{
			Enabled:  true,
			BasePath: "/mcp",
			OAuth: MCPOAuthConfig{
				Enabled:     true,
				TokenExpiry: 0,
			},
		}

		err := config.Validate()

		assert.NoError(t, err)
	})

	t.Run("returns error for negative refresh_token_expiry", func(t *testing.T) {
		config := &MCPConfig{
			Enabled:  true,
			BasePath: "/mcp",
			OAuth: MCPOAuthConfig{
				Enabled:            true,
				TokenExpiry:        time.Hour,
				RefreshTokenExpiry: -time.Hour,
			},
		}

		err := config.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "refresh_token_expiry cannot be negative")
	})

	t.Run("allows zero refresh_token_expiry", func(t *testing.T) {
		config := &MCPConfig{
			Enabled:  true,
			BasePath: "/mcp",
			OAuth: MCPOAuthConfig{
				Enabled:            true,
				RefreshTokenExpiry: 0,
			},
		}

		err := config.Validate()

		assert.NoError(t, err)
	})

	t.Run("validates full OAuth config successfully", func(t *testing.T) {
		config := &MCPConfig{
			Enabled:  true,
			BasePath: "/mcp",
			OAuth: MCPOAuthConfig{
				Enabled:             true,
				DCREnabled:          true,
				AllowedRedirectURIs: []string{"https://example.com/callback"},
				TokenExpiry:         time.Hour,
				RefreshTokenExpiry:  168 * time.Hour,
			},
		}

		err := config.Validate()

		assert.NoError(t, err)
	})
}

// =============================================================================
// MCPConfig.SetOAuthDefaults Tests
// =============================================================================

func TestMCP_ConfigSetOAuthDefaults(t *testing.T) {
	t.Run("sets default token_expiry", func(t *testing.T) {
		config := &MCPConfig{}

		config.SetOAuthDefaults()

		assert.Equal(t, time.Hour, config.OAuth.TokenExpiry)
	})

	t.Run("preserves non-zero token_expiry", func(t *testing.T) {
		config := &MCPConfig{
			OAuth: MCPOAuthConfig{
				TokenExpiry: 2 * time.Hour,
			},
		}

		config.SetOAuthDefaults()

		assert.Equal(t, 2*time.Hour, config.OAuth.TokenExpiry)
	})

	t.Run("sets default refresh_token_expiry", func(t *testing.T) {
		config := &MCPConfig{}

		config.SetOAuthDefaults()

		assert.Equal(t, 168*time.Hour, config.OAuth.RefreshTokenExpiry)
	})

	t.Run("preserves non-zero refresh_token_expiry", func(t *testing.T) {
		config := &MCPConfig{
			OAuth: MCPOAuthConfig{
				RefreshTokenExpiry: 24 * time.Hour,
			},
		}

		config.SetOAuthDefaults()

		assert.Equal(t, 24*time.Hour, config.OAuth.RefreshTokenExpiry)
	})

	t.Run("sets default redirect URIs when empty", func(t *testing.T) {
		config := &MCPConfig{}

		config.SetOAuthDefaults()

		assert.NotEmpty(t, config.OAuth.AllowedRedirectURIs)
		assert.Equal(t, DefaultMCPOAuthRedirectURIs(), config.OAuth.AllowedRedirectURIs)
	})

	t.Run("preserves non-empty redirect URIs", func(t *testing.T) {
		customURIs := []string{"https://custom.example.com/callback"}
		config := &MCPConfig{
			OAuth: MCPOAuthConfig{
				AllowedRedirectURIs: customURIs,
			},
		}

		config.SetOAuthDefaults()

		assert.Equal(t, customURIs, config.OAuth.AllowedRedirectURIs)
	})

	t.Run("sets all defaults simultaneously", func(t *testing.T) {
		config := &MCPConfig{}

		config.SetOAuthDefaults()

		assert.Equal(t, time.Hour, config.OAuth.TokenExpiry)
		assert.Equal(t, 168*time.Hour, config.OAuth.RefreshTokenExpiry)
		assert.Equal(t, DefaultMCPOAuthRedirectURIs(), config.OAuth.AllowedRedirectURIs)
	})

	t.Run("is idempotent", func(t *testing.T) {
		config := &MCPConfig{}

		config.SetOAuthDefaults()
		first := config.OAuth

		config.SetOAuthDefaults()
		second := config.OAuth

		assert.Equal(t, first.TokenExpiry, second.TokenExpiry)
		assert.Equal(t, first.RefreshTokenExpiry, second.RefreshTokenExpiry)
		assert.Equal(t, first.AllowedRedirectURIs, second.AllowedRedirectURIs)
	})
}

// =============================================================================
// MCPOAuthConfig Struct Tests
// =============================================================================

func TestMCPOAuthConfig_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		config := MCPOAuthConfig{
			Enabled:             true,
			DCREnabled:          true,
			AllowedRedirectURIs: []string{"https://example.com"},
			TokenExpiry:         time.Hour,
			RefreshTokenExpiry:  24 * time.Hour,
		}

		assert.True(t, config.Enabled)
		assert.True(t, config.DCREnabled)
		assert.Equal(t, []string{"https://example.com"}, config.AllowedRedirectURIs)
		assert.Equal(t, time.Hour, config.TokenExpiry)
		assert.Equal(t, 24*time.Hour, config.RefreshTokenExpiry)
	})

	t.Run("defaults to false for booleans", func(t *testing.T) {
		config := MCPOAuthConfig{}

		assert.False(t, config.Enabled)
		assert.False(t, config.DCREnabled)
	})

	t.Run("defaults to nil for slice", func(t *testing.T) {
		config := MCPOAuthConfig{}

		assert.Nil(t, config.AllowedRedirectURIs)
	})
}

// =============================================================================
// MCPConfig Struct Tests
// =============================================================================

func TestMCPConfig_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		config := MCPConfig{
			Enabled:          true,
			BasePath:         "/mcp",
			SessionTimeout:   time.Minute * 30,
			MaxMessageSize:   1024 * 1024,
			AllowedTools:     []string{"query", "storage"},
			AllowedResources: []string{"schema://", "storage://"},
			RateLimitPerMin:  100,
			ToolsDir:         "/app/mcp-tools",
			AutoLoadOnBoot:   true,
			OAuth: MCPOAuthConfig{
				Enabled: true,
			},
		}

		assert.True(t, config.Enabled)
		assert.Equal(t, "/mcp", config.BasePath)
		assert.Equal(t, time.Minute*30, config.SessionTimeout)
		assert.Equal(t, 1024*1024, config.MaxMessageSize)
		assert.Equal(t, []string{"query", "storage"}, config.AllowedTools)
		assert.Equal(t, []string{"schema://", "storage://"}, config.AllowedResources)
		assert.Equal(t, 100, config.RateLimitPerMin)
		assert.Equal(t, "/app/mcp-tools", config.ToolsDir)
		assert.True(t, config.AutoLoadOnBoot)
		assert.True(t, config.OAuth.Enabled)
	})

	t.Run("defaults to zero values", func(t *testing.T) {
		config := MCPConfig{}

		assert.False(t, config.Enabled)
		assert.Empty(t, config.BasePath)
		assert.Zero(t, config.SessionTimeout)
		assert.Zero(t, config.MaxMessageSize)
		assert.Nil(t, config.AllowedTools)
		assert.Nil(t, config.AllowedResources)
		assert.Zero(t, config.RateLimitPerMin)
		assert.Empty(t, config.ToolsDir)
		assert.False(t, config.AutoLoadOnBoot)
	})
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func TestMCPConfig_EdgeCases(t *testing.T) {
	t.Run("validates with empty allowed lists", func(t *testing.T) {
		config := &MCPConfig{
			Enabled:          true,
			BasePath:         "/mcp",
			AllowedTools:     []string{},
			AllowedResources: []string{},
		}

		err := config.Validate()

		assert.NoError(t, err)
	})

	t.Run("validates with whitespace base_path", func(t *testing.T) {
		config := &MCPConfig{
			Enabled:  true,
			BasePath: "   ", // whitespace only
		}

		err := config.Validate()

		// Current implementation accepts whitespace - no validation for this
		assert.NoError(t, err)
	})

	t.Run("validates with very large values", func(t *testing.T) {
		config := &MCPConfig{
			Enabled:         true,
			BasePath:        "/mcp",
			SessionTimeout:  time.Hour * 24 * 365, // 1 year
			MaxMessageSize:  1024 * 1024 * 100,    // 100 MB
			RateLimitPerMin: 1000000,
		}

		err := config.Validate()

		assert.NoError(t, err)
	})

	t.Run("validates multiple errors independently", func(t *testing.T) {
		config := &MCPConfig{
			Enabled:  true,
			BasePath: "", // First error
			// Other negative values would be checked after BasePath
		}

		err := config.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "base_path")
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkMCPConfig_Validate(b *testing.B) {
	config := &MCPConfig{
		Enabled:         true,
		BasePath:        "/mcp",
		SessionTimeout:  time.Hour,
		MaxMessageSize:  1024 * 1024,
		RateLimitPerMin: 100,
		OAuth: MCPOAuthConfig{
			Enabled:            true,
			TokenExpiry:        time.Hour,
			RefreshTokenExpiry: 168 * time.Hour,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}

func BenchmarkMCPConfig_SetOAuthDefaults(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config := &MCPConfig{}
		config.SetOAuthDefaults()
	}
}

func BenchmarkDefaultMCPOAuthRedirectURIs(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DefaultMCPOAuthRedirectURIs()
	}
}
