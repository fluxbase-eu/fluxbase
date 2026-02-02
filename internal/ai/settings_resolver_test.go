package ai

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// parseTemplateKey Tests
// =============================================================================

func TestParseTemplateKey(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedPrefix string
		expectedKey    string
	}{
		// User prefix
		{
			name:           "user prefix",
			input:          "user:api_key",
			expectedPrefix: "user",
			expectedKey:    "api_key",
		},
		{
			name:           "user prefix with dots",
			input:          "user:service.api_key",
			expectedPrefix: "user",
			expectedKey:    "service.api_key",
		},

		// System prefix
		{
			name:           "system prefix",
			input:          "system:base_url",
			expectedPrefix: "system",
			expectedKey:    "base_url",
		},
		{
			name:           "system prefix with dots",
			input:          "system:pelias.endpoint",
			expectedPrefix: "system",
			expectedKey:    "pelias.endpoint",
		},

		// No prefix (default fallback)
		{
			name:           "no prefix simple key",
			input:          "api_key",
			expectedPrefix: "",
			expectedKey:    "api_key",
		},
		{
			name:           "no prefix dotted key",
			input:          "pelias.endpoint",
			expectedPrefix: "",
			expectedKey:    "pelias.endpoint",
		},
		{
			name:           "no prefix underscore key",
			input:          "my_setting_key",
			expectedPrefix: "",
			expectedKey:    "my_setting_key",
		},

		// Edge cases
		{
			name:           "empty string",
			input:          "",
			expectedPrefix: "",
			expectedKey:    "",
		},
		{
			name:           "user: only (no key)",
			input:          "user:",
			expectedPrefix: "user",
			expectedKey:    "",
		},
		{
			name:           "system: only (no key)",
			input:          "system:",
			expectedPrefix: "system",
			expectedKey:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix, key := parseTemplateKey(tt.input)
			assert.Equal(t, tt.expectedPrefix, prefix)
			assert.Equal(t, tt.expectedKey, key)
		})
	}
}

// =============================================================================
// templatePattern Regex Tests
// =============================================================================

func TestTemplatePattern_Matches(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string // expected matches
	}{
		{
			name:     "simple key",
			input:    "Hello {{api_key}}",
			expected: []string{"{{api_key}}"},
		},
		{
			name:     "dotted key",
			input:    "Endpoint: {{pelias.endpoint}}",
			expected: []string{"{{pelias.endpoint}}"},
		},
		{
			name:     "user prefixed key",
			input:    "Key: {{user:api_key}}",
			expected: []string{"{{user:api_key}}"},
		},
		{
			name:     "system prefixed key",
			input:    "URL: {{system:base_url}}",
			expected: []string{"{{system:base_url}}"},
		},
		{
			name:     "multiple keys",
			input:    "{{key1}} and {{key2}}",
			expected: []string{"{{key1}}", "{{key2}}"},
		},
		{
			name:     "mixed prefixed keys",
			input:    "User: {{user:name}}, System: {{system:url}}, Default: {{setting}}",
			expected: []string{"{{user:name}}", "{{system:url}}", "{{setting}}"},
		},
		{
			name:     "no matches",
			input:    "No template variables here",
			expected: nil,
		},
		{
			name:     "invalid syntax - single braces",
			input:    "{single} braces",
			expected: nil,
		},
		{
			name:     "reserved user_id key",
			input:    "User ID: {{user_id}}",
			expected: []string{"{{user_id}}"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := templatePattern.FindAllString(tt.input, -1)
			if tt.expected == nil {
				assert.Nil(t, matches)
			} else {
				assert.Equal(t, tt.expected, matches)
			}
		})
	}
}

func TestTemplatePattern_InvalidPatterns(t *testing.T) {
	invalidPatterns := []string{
		"{{}}",           // empty
		"{{123key}}",     // starts with number
		"{{-key}}",       // starts with hyphen
		"{{ space }}",    // contains spaces
		"{single}",       // single braces
		"{{key",          // unclosed
		"key}}",          // no opening
	}

	for _, pattern := range invalidPatterns {
		t.Run(pattern, func(t *testing.T) {
			matches := templatePattern.FindAllString(pattern, -1)
			// These should either not match or match partially
			if len(matches) > 0 {
				// If it matches, ensure it's not matching our invalid pattern exactly
				for _, match := range matches {
					if match == pattern {
						t.Errorf("Pattern %q should not match exactly", pattern)
					}
				}
			}
		})
	}
}

// =============================================================================
// reservedKeys Tests
// =============================================================================

func TestReservedKeys(t *testing.T) {
	t.Run("user_id is reserved", func(t *testing.T) {
		assert.True(t, reservedKeys["user_id"])
	})

	t.Run("arbitrary keys are not reserved", func(t *testing.T) {
		assert.False(t, reservedKeys["api_key"])
		assert.False(t, reservedKeys["setting"])
		assert.False(t, reservedKeys[""])
	})
}

// =============================================================================
// SettingsResolver Constructor Tests
// =============================================================================

func TestNewSettingsResolver(t *testing.T) {
	t.Run("creates with nil secrets service", func(t *testing.T) {
		resolver := NewSettingsResolver(nil, 5*time.Minute)
		require.NotNil(t, resolver)
		assert.Nil(t, resolver.secretsService)
		assert.Equal(t, 5*time.Minute, resolver.cacheTTL)
		assert.NotNil(t, resolver.cache)
		assert.NotNil(t, resolver.cache.entries)
	})

	t.Run("creates with custom TTL", func(t *testing.T) {
		resolver := NewSettingsResolver(nil, 10*time.Minute)
		assert.Equal(t, 10*time.Minute, resolver.cacheTTL)
	})
}

// =============================================================================
// ExtractSettingKeys Tests
// =============================================================================

func TestSettingsResolver_ExtractSettingKeys(t *testing.T) {
	resolver := NewSettingsResolver(nil, 5*time.Minute)

	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "no keys",
			text:     "No template variables",
			expected: nil,
		},
		{
			name:     "single key",
			text:     "API key: {{api_key}}",
			expected: []string{"api_key"},
		},
		{
			name:     "multiple unique keys",
			text:     "{{key1}} and {{key2}}",
			expected: []string{"key1", "key2"},
		},
		{
			name:     "duplicate keys",
			text:     "{{key}} is the same as {{key}}",
			expected: []string{"key"},
		},
		{
			name:     "mixed prefixes extract key only",
			text:     "{{user:api_key}} and {{system:api_key}} and {{api_key}}",
			expected: []string{"api_key"},
		},
		{
			name:     "skips reserved user_id",
			text:     "User: {{user_id}}, Key: {{api_key}}",
			expected: []string{"api_key"},
		},
		{
			name:     "dotted keys",
			text:     "{{pelias.endpoint}} and {{service.api_key}}",
			expected: []string{"pelias.endpoint", "service.api_key"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolver.ExtractSettingKeys(tt.text)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.ElementsMatch(t, tt.expected, result)
			}
		})
	}
}

// =============================================================================
// InvalidateCache Tests
// =============================================================================

func TestSettingsResolver_InvalidateCache(t *testing.T) {
	resolver := NewSettingsResolver(nil, 5*time.Minute)

	// Add some entries to the cache
	resolver.cache.entries["system"] = &cacheEntry{
		settings:  map[string]string{"key": "value"},
		expiresAt: time.Now().Add(5 * time.Minute),
	}
	resolver.cache.entries["user-123"] = &cacheEntry{
		settings:  map[string]string{"key": "value"},
		expiresAt: time.Now().Add(5 * time.Minute),
	}

	require.Len(t, resolver.cache.entries, 2)

	resolver.InvalidateCache()

	assert.Len(t, resolver.cache.entries, 0)
}

// =============================================================================
// settingsCache Tests
// =============================================================================

func TestSettingsCache_Struct(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		cache := &settingsCache{
			entries: make(map[string]*cacheEntry),
		}
		assert.NotNil(t, cache.entries)
		assert.Len(t, cache.entries, 0)
	})
}

func TestCacheEntry_Struct(t *testing.T) {
	t.Run("valid entry", func(t *testing.T) {
		entry := &cacheEntry{
			settings:  map[string]string{"key": "value"},
			expiresAt: time.Now().Add(5 * time.Minute),
		}
		assert.NotNil(t, entry.settings)
		assert.False(t, entry.expiresAt.IsZero())
	})

	t.Run("expired check", func(t *testing.T) {
		expiredEntry := &cacheEntry{
			settings:  map[string]string{},
			expiresAt: time.Now().Add(-5 * time.Minute),
		}
		assert.True(t, expiredEntry.expiresAt.Before(time.Now()))

		validEntry := &cacheEntry{
			settings:  map[string]string{},
			expiresAt: time.Now().Add(5 * time.Minute),
		}
		assert.True(t, validEntry.expiresAt.After(time.Now()))
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkParseTemplateKey_NoPrefix(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parseTemplateKey("api_key")
	}
}

func BenchmarkParseTemplateKey_UserPrefix(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parseTemplateKey("user:api_key")
	}
}

func BenchmarkParseTemplateKey_SystemPrefix(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parseTemplateKey("system:base_url")
	}
}

func BenchmarkTemplatePattern_FindAll(b *testing.B) {
	text := "Hello {{user:name}}, your key is {{api_key}} and endpoint is {{system:url}}"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		templatePattern.FindAllString(text, -1)
	}
}

func BenchmarkExtractSettingKeys(b *testing.B) {
	resolver := NewSettingsResolver(nil, 5*time.Minute)
	text := "Hello {{user:name}}, your key is {{api_key}} and endpoint is {{system:url}}"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resolver.ExtractSettingKeys(text)
	}
}
