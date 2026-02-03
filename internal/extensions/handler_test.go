package extensions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHandler(t *testing.T) {
	t.Run("creates handler with service", func(t *testing.T) {
		service := &Service{}
		handler := NewHandler(service)

		assert.NotNil(t, handler)
		assert.Equal(t, service, handler.service)
	})

	t.Run("creates handler with nil service", func(t *testing.T) {
		handler := NewHandler(nil)

		assert.NotNil(t, handler)
		assert.Nil(t, handler.service)
	})
}

func TestHandler_Struct(t *testing.T) {
	// Test that handler struct is properly defined
	var h Handler
	assert.Nil(t, h.service)
}

// Note: Full handler tests require a mock fiber.Ctx and Service
// which is typically done in integration tests
