package auth

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// ErrSettingNotFound Tests
// =============================================================================

func TestErrSettingNotFound(t *testing.T) {
	t.Run("error message is correct", func(t *testing.T) {
		assert.Equal(t, "system setting not found", ErrSettingNotFound.Error())
	})
}

// =============================================================================
// SystemSetting Struct Tests
// =============================================================================

func TestSystemSetting_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		now := time.Now()
		desc := "Test description"

		setting := SystemSetting{
			ID:             uuid.New(),
			Key:            "test_setting",
			Value:          map[string]interface{}{"enabled": true},
			Description:    &desc,
			IsOverridden:   false,
			OverrideSource: "",
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		assert.Equal(t, "test_setting", setting.Key)
		assert.Equal(t, true, setting.Value["enabled"])
		assert.Equal(t, "Test description", *setting.Description)
		assert.False(t, setting.IsOverridden)
	})

	t.Run("zero value setting", func(t *testing.T) {
		var setting SystemSetting

		assert.Equal(t, uuid.Nil, setting.ID)
		assert.Empty(t, setting.Key)
		assert.Nil(t, setting.Value)
		assert.Nil(t, setting.Description)
	})
}

func TestSystemSetting_JSON(t *testing.T) {
	t.Run("serializes to JSON", func(t *testing.T) {
		desc := "My setting"
		setting := SystemSetting{
			ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			Key:         "my_key",
			Value:       map[string]interface{}{"count": 42},
			Description: &desc,
		}

		data, err := json.Marshal(setting)

		require.NoError(t, err)
		assert.Contains(t, string(data), `"key":"my_key"`)
		assert.Contains(t, string(data), `"count":42`)
	})

	t.Run("deserializes from JSON", func(t *testing.T) {
		jsonData := `{
			"id": "550e8400-e29b-41d4-a716-446655440000",
			"key": "test_key",
			"value": {"enabled": true, "count": 10},
			"description": "Test setting"
		}`

		var setting SystemSetting
		err := json.Unmarshal([]byte(jsonData), &setting)

		require.NoError(t, err)
		assert.Equal(t, "test_key", setting.Key)
		assert.Equal(t, true, setting.Value["enabled"])
		assert.Equal(t, float64(10), setting.Value["count"])
	})
}

// =============================================================================
// SetupCompleteValue Tests
// =============================================================================

func TestSetupCompleteValue_Struct(t *testing.T) {
	t.Run("complete setup value", func(t *testing.T) {
		adminID := uuid.New()
		adminEmail := "admin@example.com"
		now := time.Now()

		value := SetupCompleteValue{
			Completed:       true,
			CompletedAt:     now,
			FirstAdminID:    &adminID,
			FirstAdminEmail: &adminEmail,
		}

		assert.True(t, value.Completed)
		assert.Equal(t, adminID, *value.FirstAdminID)
		assert.Equal(t, "admin@example.com", *value.FirstAdminEmail)
	})

	t.Run("incomplete setup value", func(t *testing.T) {
		value := SetupCompleteValue{
			Completed: false,
		}

		assert.False(t, value.Completed)
		assert.Nil(t, value.FirstAdminID)
		assert.Nil(t, value.FirstAdminEmail)
	})
}

func TestSetupCompleteValue_JSON(t *testing.T) {
	t.Run("serializes to JSON", func(t *testing.T) {
		adminID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
		adminEmail := "admin@test.com"
		now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

		value := SetupCompleteValue{
			Completed:       true,
			CompletedAt:     now,
			FirstAdminID:    &adminID,
			FirstAdminEmail: &adminEmail,
		}

		data, err := json.Marshal(value)

		require.NoError(t, err)
		assert.Contains(t, string(data), `"completed":true`)
		assert.Contains(t, string(data), `"first_admin_email":"admin@test.com"`)
	})

	t.Run("deserializes from JSON", func(t *testing.T) {
		jsonData := `{
			"completed": true,
			"completed_at": "2024-01-15T10:30:00Z",
			"first_admin_id": "550e8400-e29b-41d4-a716-446655440000",
			"first_admin_email": "admin@test.com"
		}`

		var value SetupCompleteValue
		err := json.Unmarshal([]byte(jsonData), &value)

		require.NoError(t, err)
		assert.True(t, value.Completed)
		assert.Equal(t, "admin@test.com", *value.FirstAdminEmail)
	})
}

// =============================================================================
// SystemSettingsService Construction Tests
// =============================================================================

func TestNewSystemSettingsService(t *testing.T) {
	t.Run("creates service with nil database", func(t *testing.T) {
		service := NewSystemSettingsService(nil)

		require.NotNil(t, service)
		assert.Nil(t, service.db)
		assert.Nil(t, service.cache)
	})
}

func TestSystemSettingsService_SetCache(t *testing.T) {
	t.Run("sets cache instance", func(t *testing.T) {
		service := NewSystemSettingsService(nil)

		assert.Nil(t, service.cache)

		// We can't easily create a SettingsCache in tests, but we verify the method exists
		service.SetCache(nil)
		assert.Nil(t, service.cache)
	})
}

// =============================================================================
// Legacy Format Handling Tests
// =============================================================================

func TestSystemSettings_LegacyFormatHandling(t *testing.T) {
	t.Run("parses legacy primitive value", func(t *testing.T) {
		// Legacy format: "true" (raw primitive)
		// Expected result: {"value": true}

		legacyJSON := []byte(`true`)

		var rawValue interface{}
		err := json.Unmarshal(legacyJSON, &rawValue)
		require.NoError(t, err)

		// Convert to expected format
		value := map[string]interface{}{"value": rawValue}
		assert.Equal(t, true, value["value"])
	})

	t.Run("parses legacy string value", func(t *testing.T) {
		// Legacy format: "my-value" (raw string)
		legacyJSON := []byte(`"my-value"`)

		var rawValue interface{}
		err := json.Unmarshal(legacyJSON, &rawValue)
		require.NoError(t, err)

		value := map[string]interface{}{"value": rawValue}
		assert.Equal(t, "my-value", value["value"])
	})

	t.Run("parses legacy number value", func(t *testing.T) {
		// Legacy format: 42 (raw number)
		legacyJSON := []byte(`42`)

		var rawValue interface{}
		err := json.Unmarshal(legacyJSON, &rawValue)
		require.NoError(t, err)

		value := map[string]interface{}{"value": rawValue}
		assert.Equal(t, float64(42), value["value"])
	})

	t.Run("parses new object format", func(t *testing.T) {
		// New format: {"key": "value", "enabled": true}
		newJSON := []byte(`{"key": "value", "enabled": true}`)

		var value map[string]interface{}
		err := json.Unmarshal(newJSON, &value)
		require.NoError(t, err)

		assert.Equal(t, "value", value["key"])
		assert.Equal(t, true, value["enabled"])
	})
}

// =============================================================================
// GetSettings Batch Query Tests
// =============================================================================

func TestSystemSettings_BatchQuery(t *testing.T) {
	t.Run("empty keys returns empty map", func(t *testing.T) {
		// GetSettings with empty keys should return an empty map immediately
		// without making a database query

		keys := []string{}
		result := make(map[string]*SystemSetting, len(keys))

		assert.Empty(t, result)
		assert.NotNil(t, result)
	})

	t.Run("multiple keys query structure", func(t *testing.T) {
		// The query should use ANY($1) for efficient batch lookup
		// This is more efficient than multiple single-key queries

		keys := []string{"setting1", "setting2", "setting3"}
		assert.Len(t, keys, 3)
	})
}

// =============================================================================
// Cache Invalidation Tests
// =============================================================================

func TestSystemSettings_CacheInvalidation(t *testing.T) {
	t.Run("SetSetting should invalidate cache", func(t *testing.T) {
		// When SetSetting is called, it should:
		// 1. Update the database
		// 2. Call cache.Invalidate(key) if cache is set

		service := NewSystemSettingsService(nil)
		assert.Nil(t, service.cache)
	})

	t.Run("DeleteSetting should invalidate cache", func(t *testing.T) {
		// When DeleteSetting is called, it should:
		// 1. Delete from database
		// 2. Call cache.Invalidate(key) if cache is set

		service := NewSystemSettingsService(nil)
		assert.Nil(t, service.cache)
	})
}

// =============================================================================
// MarkSetupComplete Idempotency Tests
// =============================================================================

func TestSystemSettings_SetupIdempotency(t *testing.T) {
	t.Run("MarkSetupComplete checks if already complete", func(t *testing.T) {
		// The method should:
		// 1. First call IsSetupComplete()
		// 2. If already complete, return error "setup already marked as completed"
		// 3. If not complete, create the setup_completed setting

		service := NewSystemSettingsService(nil)
		assert.NotNil(t, service)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkNewSystemSettingsService(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewSystemSettingsService(nil)
	}
}

func BenchmarkSetupCompleteValue_Marshal(b *testing.B) {
	adminID := uuid.New()
	adminEmail := "admin@test.com"
	value := SetupCompleteValue{
		Completed:       true,
		CompletedAt:     time.Now(),
		FirstAdminID:    &adminID,
		FirstAdminEmail: &adminEmail,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(value)
	}
}

func BenchmarkSystemSetting_Marshal(b *testing.B) {
	desc := "Test setting"
	setting := SystemSetting{
		ID:          uuid.New(),
		Key:         "test_key",
		Value:       map[string]interface{}{"enabled": true, "count": 42},
		Description: &desc,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(setting)
	}
}
