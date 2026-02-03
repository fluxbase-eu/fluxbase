package realtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAuthServiceAdapter(t *testing.T) {
	t.Run("creates adapter with service", func(t *testing.T) {
		adapter := NewAuthServiceAdapter(nil)
		assert.NotNil(t, adapter)
	})
}

func TestAuthServiceAdapter_Struct(t *testing.T) {
	t.Run("stores service reference", func(t *testing.T) {
		adapter := &AuthServiceAdapter{}
		assert.Nil(t, adapter.service)
	})
}

func TestTokenClaims_Struct(t *testing.T) {
	claims := TokenClaims{
		UserID:    "user-123",
		Email:     "user@example.com",
		Role:      "authenticated",
		SessionID: "session-456",
		RawClaims: map[string]interface{}{
			"sub":   "user-123",
			"email": "user@example.com",
		},
	}

	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "user@example.com", claims.Email)
	assert.Equal(t, "authenticated", claims.Role)
	assert.Equal(t, "session-456", claims.SessionID)
	assert.NotNil(t, claims.RawClaims)
	assert.Equal(t, "user-123", claims.RawClaims["sub"])
}

// Note: Full ValidateToken tests require an auth.Service
// which requires database connection
// These are covered by integration tests
