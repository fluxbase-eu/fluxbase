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

// Note: Full ValidateToken tests require an auth.Service
// which requires database connection
// These are covered by integration tests
