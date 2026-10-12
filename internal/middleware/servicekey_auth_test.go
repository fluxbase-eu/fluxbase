package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nimbleflux/fluxbase/internal/keys"
)

// =============================================================================
// mapKeyTypetoRole Tests
// =============================================================================

func TestMapKeyTypetoRole(t *testing.T) {
	tests := []struct {
		name     string
		keyType  string
		expected string
	}{
		{"anon constant", keys.KeyTypeAnon, "anon"},
		{"tenant_service constant", keys.KeyTypeTenantService, "tenant_service"},
		{"global_service constant", keys.KeyTypeGlobalService, "service_role"},
		{"publishable constant", keys.KeyTypePublishable, "anon"},
		{"legacy service type", "service", "service_role"},
		{"unknown type defaults to anon", "unknown", "anon"},
		{"empty type defaults to anon", "", "anon"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapKeyTypetoRole(tt.keyType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// Key Prefix Routing Tests
// =============================================================================

func TestKeyPrefixRouting_PlatformKeys(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
	}{
		{"tenant_service prefix", keys.KeyPrefixTenantService},
		{"anon prefix", keys.KeyPrefixAnon},
		{"global_service prefix", keys.KeyPrefixGlobalService},
		{"publishable prefix", keys.KeyPrefixPublishable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := tt.prefix + "abcdefgh1234567890abcdefghijklmnop"
			extracted := keys.ExtractPrefix(key)
			assert.Equal(t, tt.prefix, extracted, "ExtractPrefix should return the correct prefix")
			assert.True(t, len(key) >= 8, "Key should be at least 8 chars")
		})
	}
}

func TestKeyPrefixRouting_LegacyKeys(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		isSk    bool
		isFbKey bool
	}{
		{"sk_ prefix is legacy", "sk_test_1234567890abcdef", true, false},
		{"pk_ prefix is legacy", "pk_live_1234567890abcdef", false, false},
		{"fb_tsk_ is platform key", "fb_tsk_1234567890abcdefghij", false, true},
		{"fb_anon_ is platform key", "fb_anon_1234567890abcdefghij", false, true},
		{"fb_gsk_ is platform key", "fb_gsk_1234567890abcdefghij", false, true},
		{"random prefix is unrecognized", "xx_unknown_1234567890", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isSk, len(tt.key) >= 2 && tt.key[:3] == "sk_")
			extracted := keys.ExtractPrefix(tt.key)
			assert.Equal(t, tt.isFbKey, extracted != "", "ExtractPrefix should detect platform keys")
		})
	}
}
