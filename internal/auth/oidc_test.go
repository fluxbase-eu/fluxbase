package auth

import (
	"testing"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeIssuerURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "base URL without trailing slash",
			input:    "https://auth.domain.com",
			expected: "https://auth.domain.com/.well-known/openid-configuration",
		},
		{
			name:     "base URL with trailing slash",
			input:    "https://auth.domain.com/",
			expected: "https://auth.domain.com/.well-known/openid-configuration",
		},
		{
			name:     "base URL with path",
			input:    "https://example.com/auth",
			expected: "https://example.com/auth/.well-known/openid-configuration",
		},
		{
			name:     "base URL with path and trailing slash",
			input:    "https://example.com/auth/",
			expected: "https://example.com/auth/.well-known/openid-configuration",
		},
		{
			name:     "already contains .well-known endpoint",
			input:    "https://auth.domain.com/.well-known/openid-configuration",
			expected: "https://auth.domain.com/.well-known/openid-configuration",
		},
		{
			name:     "custom .well-known path",
			input:    "https://auth.domain.com/.well-known/custom-oidc",
			expected: "https://auth.domain.com/.well-known/custom-oidc",
		},
		{
			name:     "Keycloak-style URL",
			input:    "https://keycloak.example.com/realms/myrealm",
			expected: "https://keycloak.example.com/realms/myrealm/.well-known/openid-configuration",
		},
		{
			name:     "Auth0-style URL",
			input:    "https://tenant.auth0.com",
			expected: "https://tenant.auth0.com/.well-known/openid-configuration",
		},
		{
			name:     "localhost for development",
			input:    "http://localhost:8080",
			expected: "http://localhost:8080/.well-known/openid-configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeIssuerURL(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeIssuerURL(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestWellKnownIssuers(t *testing.T) {
	t.Run("google issuer is configured", func(t *testing.T) {
		issuer, ok := wellKnownIssuers["google"]
		assert.True(t, ok)
		assert.Equal(t, "https://accounts.google.com", issuer)
	})

	t.Run("apple issuer is configured", func(t *testing.T) {
		issuer, ok := wellKnownIssuers["apple"]
		assert.True(t, ok)
		assert.Equal(t, "https://appleid.apple.com", issuer)
	})

	t.Run("microsoft issuer is configured", func(t *testing.T) {
		issuer, ok := wellKnownIssuers["microsoft"]
		assert.True(t, ok)
		assert.Equal(t, "https://login.microsoftonline.com/common/v2.0", issuer)
	})

	t.Run("unknown provider returns false", func(t *testing.T) {
		_, ok := wellKnownIssuers["unknown"]
		assert.False(t, ok)
	})
}

func TestIDTokenClaims(t *testing.T) {
	t.Run("creates claims with all fields", func(t *testing.T) {
		claims := &IDTokenClaims{
			Subject:       "user-123",
			Email:         "test@example.com",
			EmailVerified: true,
			Name:          "Test User",
			Picture:       "https://example.com/pic.jpg",
			Nonce:         "nonce-abc",
		}

		assert.Equal(t, "user-123", claims.Subject)
		assert.Equal(t, "test@example.com", claims.Email)
		assert.True(t, claims.EmailVerified)
		assert.Equal(t, "Test User", claims.Name)
		assert.Equal(t, "https://example.com/pic.jpg", claims.Picture)
		assert.Equal(t, "nonce-abc", claims.Nonce)
	})

	t.Run("creates minimal claims", func(t *testing.T) {
		claims := &IDTokenClaims{
			Subject: "user-456",
			Email:   "minimal@example.com",
		}

		assert.Equal(t, "user-456", claims.Subject)
		assert.Equal(t, "minimal@example.com", claims.Email)
		assert.False(t, claims.EmailVerified)
		assert.Empty(t, claims.Name)
		assert.Empty(t, claims.Picture)
		assert.Empty(t, claims.Nonce)
	})

	t.Run("zero value", func(t *testing.T) {
		var claims IDTokenClaims
		assert.Empty(t, claims.Subject)
		assert.Empty(t, claims.Email)
		assert.False(t, claims.EmailVerified)
	})
}

func TestOIDCVerifier_Struct(t *testing.T) {
	t.Run("creates empty verifier", func(t *testing.T) {
		v := &OIDCVerifier{
			verifiers: make(map[string]*oidc.IDTokenVerifier),
			providers: make(map[string]*oidc.Provider),
			clientIDs: make(map[string]string),
		}

		assert.NotNil(t, v.verifiers)
		assert.NotNil(t, v.providers)
		assert.NotNil(t, v.clientIDs)
		assert.Empty(t, v.verifiers)
		assert.Empty(t, v.providers)
		assert.Empty(t, v.clientIDs)
	})
}

func TestOIDCVerifier_IsProviderConfigured(t *testing.T) {
	t.Run("returns false for unconfigured provider", func(t *testing.T) {
		v := &OIDCVerifier{
			verifiers: make(map[string]*oidc.IDTokenVerifier),
		}

		assert.False(t, v.IsProviderConfigured("google"))
		assert.False(t, v.IsProviderConfigured("apple"))
		assert.False(t, v.IsProviderConfigured("unknown"))
	})

	t.Run("returns true for configured provider", func(t *testing.T) {
		v := &OIDCVerifier{
			verifiers: map[string]*oidc.IDTokenVerifier{
				"google": nil, // nil verifier is still "configured"
			},
		}

		assert.True(t, v.IsProviderConfigured("google"))
		assert.True(t, v.IsProviderConfigured("Google"))  // case insensitive
		assert.True(t, v.IsProviderConfigured("GOOGLE"))  // case insensitive
		assert.False(t, v.IsProviderConfigured("apple"))
	})
}

func TestOIDCVerifier_ListProviders(t *testing.T) {
	t.Run("returns empty list when no providers configured", func(t *testing.T) {
		v := &OIDCVerifier{
			verifiers: make(map[string]*oidc.IDTokenVerifier),
		}

		providers := v.ListProviders()
		assert.Empty(t, providers)
	})

	t.Run("returns configured providers", func(t *testing.T) {
		v := &OIDCVerifier{
			verifiers: map[string]*oidc.IDTokenVerifier{
				"google": nil,
				"apple":  nil,
			},
		}

		providers := v.ListProviders()
		assert.Len(t, providers, 2)
		assert.Contains(t, providers, "google")
		assert.Contains(t, providers, "apple")
	})

	t.Run("returns single provider", func(t *testing.T) {
		v := &OIDCVerifier{
			verifiers: map[string]*oidc.IDTokenVerifier{
				"microsoft": nil,
			},
		}

		providers := v.ListProviders()
		assert.Len(t, providers, 1)
		assert.Equal(t, "microsoft", providers[0])
	})
}

func TestOIDCVerifier_Verify_Errors(t *testing.T) {
	t.Run("returns error for unconfigured provider", func(t *testing.T) {
		v := &OIDCVerifier{
			verifiers: make(map[string]*oidc.IDTokenVerifier),
		}

		claims, err := v.Verify(nil, "unknown", "token", "")
		assert.Nil(t, claims)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not configured")
	})

	t.Run("handles case insensitive provider names", func(t *testing.T) {
		v := &OIDCVerifier{
			verifiers: make(map[string]*oidc.IDTokenVerifier),
		}

		// All of these should return "not configured" error (case-normalized)
		_, err1 := v.Verify(nil, "Google", "token", "")
		_, err2 := v.Verify(nil, "GOOGLE", "token", "")
		_, err3 := v.Verify(nil, "google", "token", "")

		assert.Contains(t, err1.Error(), "not configured")
		assert.Contains(t, err2.Error(), "not configured")
		assert.Contains(t, err3.Error(), "not configured")
	})
}

func TestOIDCVerifier_ConcurrentAccess(t *testing.T) {
	v := &OIDCVerifier{
		verifiers: map[string]*oidc.IDTokenVerifier{
			"google": nil,
			"apple":  nil,
		},
		providers: make(map[string]*oidc.Provider),
		clientIDs: make(map[string]string),
	}

	// Test concurrent reads don't panic
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				v.IsProviderConfigured("google")
				v.ListProviders()
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Should complete without panics
	assert.True(t, v.IsProviderConfigured("google"))
}
