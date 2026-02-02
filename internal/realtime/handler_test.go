package realtime

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// RealtimeHandler Construction Tests
// =============================================================================

func TestNewRealtimeHandler(t *testing.T) {
	t.Run("creates handler with nil dependencies", func(t *testing.T) {
		handler := NewRealtimeHandler(nil, nil, nil)

		require.NotNil(t, handler)
		assert.Nil(t, handler.manager)
		assert.Nil(t, handler.authService)
		assert.Nil(t, handler.subManager)
		assert.NotNil(t, handler.presenceManager) // Should always create presence manager
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkClientMessage_Unmarshal(b *testing.B) {
	data := []byte(`{"type":"subscribe","channel":"realtime:public:users","event":"*","schema":"public","table":"users"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var msg ClientMessage
		_ = json.Unmarshal(data, &msg)
	}
}

func BenchmarkServerMessage_Marshal(b *testing.B) {
	msg := ServerMessage{
		Type:    MessageTypeChange,
		Channel: "realtime:public:users",
		Payload: map[string]interface{}{
			"event": "INSERT",
			"new":   map[string]interface{}{"id": "123", "name": "Test"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(msg)
	}
}
