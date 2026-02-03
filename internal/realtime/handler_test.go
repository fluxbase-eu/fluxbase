package realtime

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// MessageType Constants Tests
// =============================================================================

func TestRealtimeHandler_MessageTypeConstants(t *testing.T) {
	t.Run("message types have correct values", func(t *testing.T) {
		assert.Equal(t, MessageType("subscribe"), MessageTypeSubscribe)
		assert.Equal(t, MessageType("unsubscribe"), MessageTypeUnsubscribe)
		assert.Equal(t, MessageType("heartbeat"), MessageTypeHeartbeat)
		assert.Equal(t, MessageType("broadcast"), MessageTypeBroadcast)
		assert.Equal(t, MessageType("presence"), MessageTypePresence)
		assert.Equal(t, MessageType("error"), MessageTypeError)
		assert.Equal(t, MessageType("ack"), MessageTypeAck)
		assert.Equal(t, MessageType("postgres_changes"), MessageTypeChange)
		assert.Equal(t, MessageType("access_token"), MessageTypeAccessToken)
	})

	t.Run("log subscription types have correct values", func(t *testing.T) {
		assert.Equal(t, MessageType("subscribe_logs"), MessageTypeSubscribeLogs)
		assert.Equal(t, MessageType("execution_log"), MessageTypeExecutionLog)
		assert.Equal(t, MessageType("subscribe_all_logs"), MessageTypeSubscribeAllLogs)
		assert.Equal(t, MessageType("log_entry"), MessageTypeLogEntry)
	})
}

// =============================================================================
// ClientMessage Tests
// =============================================================================

func TestRealtimeClientMessage_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		msg := ClientMessage{
			Type:           MessageTypeSubscribe,
			Channel:        "realtime:public:users",
			Event:          "INSERT",
			Schema:         "public",
			Table:          "users",
			Filter:         "id=eq.123",
			SubscriptionID: "sub-123",
			MessageID:      "msg-456",
			Token:          "jwt-token",
		}

		assert.Equal(t, MessageTypeSubscribe, msg.Type)
		assert.Equal(t, "realtime:public:users", msg.Channel)
		assert.Equal(t, "INSERT", msg.Event)
		assert.Equal(t, "public", msg.Schema)
		assert.Equal(t, "users", msg.Table)
		assert.Equal(t, "id=eq.123", msg.Filter)
		assert.Equal(t, "sub-123", msg.SubscriptionID)
		assert.Equal(t, "msg-456", msg.MessageID)
		assert.Equal(t, "jwt-token", msg.Token)
	})

	t.Run("defaults to zero values", func(t *testing.T) {
		msg := ClientMessage{}

		assert.Empty(t, msg.Type)
		assert.Empty(t, msg.Channel)
		assert.Empty(t, msg.Event)
		assert.Empty(t, msg.Schema)
		assert.Empty(t, msg.Table)
		assert.Nil(t, msg.Payload)
		assert.Nil(t, msg.Config)
	})
}

func TestRealtimeClientMessage_JSONSerialization(t *testing.T) {
	t.Run("serializes to JSON", func(t *testing.T) {
		msg := ClientMessage{
			Type:    MessageTypeSubscribe,
			Channel: "realtime:public:users",
			Event:   "*",
			Schema:  "public",
			Table:   "users",
		}

		data, err := json.Marshal(msg)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"type":"subscribe"`)
		assert.Contains(t, string(data), `"channel":"realtime:public:users"`)
		assert.Contains(t, string(data), `"event":"*"`)
	})

	t.Run("deserializes from JSON", func(t *testing.T) {
		data := `{
			"type": "subscribe",
			"channel": "realtime:public:users",
			"event": "INSERT",
			"schema": "public",
			"table": "users"
		}`

		var msg ClientMessage
		err := json.Unmarshal([]byte(data), &msg)
		require.NoError(t, err)

		assert.Equal(t, MessageTypeSubscribe, msg.Type)
		assert.Equal(t, "realtime:public:users", msg.Channel)
		assert.Equal(t, "INSERT", msg.Event)
		assert.Equal(t, "public", msg.Schema)
		assert.Equal(t, "users", msg.Table)
	})

	t.Run("handles payload field", func(t *testing.T) {
		data := `{
			"type": "broadcast",
			"channel": "room:123",
			"payload": {"message": "hello", "user": "john"}
		}`

		var msg ClientMessage
		err := json.Unmarshal([]byte(data), &msg)
		require.NoError(t, err)

		assert.Equal(t, MessageTypeBroadcast, msg.Type)
		assert.NotNil(t, msg.Payload)

		var payload map[string]string
		err = json.Unmarshal(msg.Payload, &payload)
		require.NoError(t, err)
		assert.Equal(t, "hello", payload["message"])
	})
}

// =============================================================================
// ServerMessage Tests
// =============================================================================

func TestRealtimeHandler_ServerMessageStruct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		msg := ServerMessage{
			Type:    MessageTypeChange,
			Channel: "realtime:public:users",
			Payload: map[string]interface{}{"data": "test"},
			Error:   "",
		}

		assert.Equal(t, MessageTypeChange, msg.Type)
		assert.Equal(t, "realtime:public:users", msg.Channel)
		assert.NotNil(t, msg.Payload)
		assert.Empty(t, msg.Error)
	})

	t.Run("stores error message", func(t *testing.T) {
		msg := ServerMessage{
			Type:  MessageTypeError,
			Error: "subscription failed",
		}

		assert.Equal(t, MessageTypeError, msg.Type)
		assert.Equal(t, "subscription failed", msg.Error)
	})
}

func TestServerMessage_JSONSerialization(t *testing.T) {
	t.Run("serializes to JSON", func(t *testing.T) {
		msg := ServerMessage{
			Type:    MessageTypeChange,
			Channel: "realtime:public:users",
			Payload: map[string]interface{}{
				"event": "INSERT",
				"new":   map[string]interface{}{"id": 1, "name": "Test"},
			},
		}

		data, err := json.Marshal(msg)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"type":"postgres_changes"`)
		assert.Contains(t, string(data), `"channel":"realtime:public:users"`)
		assert.Contains(t, string(data), `"event":"INSERT"`)
	})

	t.Run("omits empty fields", func(t *testing.T) {
		msg := ServerMessage{
			Type: MessageTypeAck,
		}

		data, err := json.Marshal(msg)
		require.NoError(t, err)

		// Empty channel and error should be omitted
		assert.NotContains(t, string(data), `"error"`)
	})

	t.Run("deserializes from JSON", func(t *testing.T) {
		data := `{
			"type": "postgres_changes",
			"channel": "realtime:public:orders",
			"payload": {"event": "UPDATE", "old": {"status": "pending"}, "new": {"status": "completed"}}
		}`

		var msg ServerMessage
		err := json.Unmarshal([]byte(data), &msg)
		require.NoError(t, err)

		assert.Equal(t, MessageTypeChange, msg.Type)
		assert.Equal(t, "realtime:public:orders", msg.Channel)

		payload, ok := msg.Payload.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "UPDATE", payload["event"])
	})
}

// =============================================================================
// PostgresChangesConfig Tests
// =============================================================================

func TestRealtimeHandler_PostgresChangesConfigStruct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		config := PostgresChangesConfig{
			Event:  "INSERT",
			Schema: "public",
			Table:  "users",
			Filter: "id=eq.123",
		}

		assert.Equal(t, "INSERT", config.Event)
		assert.Equal(t, "public", config.Schema)
		assert.Equal(t, "users", config.Table)
		assert.Equal(t, "id=eq.123", config.Filter)
	})

	t.Run("filter is optional", func(t *testing.T) {
		config := PostgresChangesConfig{
			Event:  "*",
			Schema: "public",
			Table:  "orders",
		}

		assert.Empty(t, config.Filter)
	})
}

func TestPostgresChangesConfig_JSONSerialization(t *testing.T) {
	t.Run("serializes to JSON", func(t *testing.T) {
		config := PostgresChangesConfig{
			Event:  "INSERT",
			Schema: "public",
			Table:  "users",
			Filter: "status=eq.active",
		}

		data, err := json.Marshal(config)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"event":"INSERT"`)
		assert.Contains(t, string(data), `"schema":"public"`)
		assert.Contains(t, string(data), `"table":"users"`)
		assert.Contains(t, string(data), `"filter":"status=eq.active"`)
	})

	t.Run("deserializes from JSON", func(t *testing.T) {
		data := `{"event":"*","schema":"public","table":"messages"}`

		var config PostgresChangesConfig
		err := json.Unmarshal([]byte(data), &config)
		require.NoError(t, err)

		assert.Equal(t, "*", config.Event)
		assert.Equal(t, "public", config.Schema)
		assert.Equal(t, "messages", config.Table)
	})
}

// =============================================================================
// LogSubscriptionConfig Tests
// =============================================================================

func TestRealtimeHandler_LogSubscriptionConfigStruct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		config := LogSubscriptionConfig{
			ExecutionID: "exec-123",
			Type:        "function",
		}

		assert.Equal(t, "exec-123", config.ExecutionID)
		assert.Equal(t, "function", config.Type)
	})

	t.Run("accepts different types", func(t *testing.T) {
		types := []string{"function", "job", "rpc"}
		for _, typ := range types {
			config := LogSubscriptionConfig{Type: typ}
			assert.Equal(t, typ, config.Type)
		}
	})
}

func TestLogSubscriptionConfig_JSONSerialization(t *testing.T) {
	t.Run("serializes to JSON", func(t *testing.T) {
		config := LogSubscriptionConfig{
			ExecutionID: "exec-456",
			Type:        "job",
		}

		data, err := json.Marshal(config)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"execution_id":"exec-456"`)
		assert.Contains(t, string(data), `"type":"job"`)
	})

	t.Run("deserializes from JSON", func(t *testing.T) {
		data := `{"execution_id":"exec-789","type":"rpc"}`

		var config LogSubscriptionConfig
		err := json.Unmarshal([]byte(data), &config)
		require.NoError(t, err)

		assert.Equal(t, "exec-789", config.ExecutionID)
		assert.Equal(t, "rpc", config.Type)
	})
}

// =============================================================================
// TokenClaims Tests
// =============================================================================

func TestRealtimeHandler_TokenClaimsStruct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		claims := TokenClaims{
			UserID:    "user-123",
			Email:     "user@example.com",
			Role:      "authenticated",
			SessionID: "session-456",
			RawClaims: map[string]interface{}{
				"sub":   "user-123",
				"email": "user@example.com",
				"role":  "authenticated",
			},
		}

		assert.Equal(t, "user-123", claims.UserID)
		assert.Equal(t, "user@example.com", claims.Email)
		assert.Equal(t, "authenticated", claims.Role)
		assert.Equal(t, "session-456", claims.SessionID)
		assert.NotNil(t, claims.RawClaims)
	})

	t.Run("defaults to zero values", func(t *testing.T) {
		claims := TokenClaims{}

		assert.Empty(t, claims.UserID)
		assert.Empty(t, claims.Email)
		assert.Empty(t, claims.Role)
		assert.Empty(t, claims.SessionID)
		assert.Nil(t, claims.RawClaims)
	})

	t.Run("raw claims can contain custom fields", func(t *testing.T) {
		claims := TokenClaims{
			UserID: "user-123",
			RawClaims: map[string]interface{}{
				"meeting_id": "meeting-789",
				"player_id":  12345,
				"is_admin":   true,
			},
		}

		assert.Equal(t, "meeting-789", claims.RawClaims["meeting_id"])
		assert.Equal(t, 12345, claims.RawClaims["player_id"])
		assert.Equal(t, true, claims.RawClaims["is_admin"])
	})
}

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

func TestRealtimeHandler_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		handler := &RealtimeHandler{
			manager:         nil,
			authService:     nil,
			subManager:      nil,
			presenceManager: NewPresenceManager(),
		}

		assert.Nil(t, handler.manager)
		assert.Nil(t, handler.authService)
		assert.Nil(t, handler.subManager)
		assert.NotNil(t, handler.presenceManager)
	})
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestClientMessage_AllEventTypes(t *testing.T) {
	eventTypes := []struct {
		event    string
		expected string
	}{
		{"INSERT", "INSERT"},
		{"UPDATE", "UPDATE"},
		{"DELETE", "DELETE"},
		{"*", "*"},
	}

	for _, tc := range eventTypes {
		t.Run(tc.event, func(t *testing.T) {
			msg := ClientMessage{
				Type:  MessageTypeSubscribe,
				Event: tc.event,
			}
			assert.Equal(t, tc.expected, msg.Event)
		})
	}
}

func TestServerMessage_ErrorPayload(t *testing.T) {
	t.Run("error message with payload", func(t *testing.T) {
		msg := ServerMessage{
			Type:  MessageTypeError,
			Error: "max_connections_reached",
			Payload: map[string]interface{}{
				"message": "Server connection limit reached. Please try again later.",
			},
		}

		assert.Equal(t, MessageTypeError, msg.Type)
		assert.Equal(t, "max_connections_reached", msg.Error)

		payload, ok := msg.Payload.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, payload["message"], "connection limit")
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

func BenchmarkPostgresChangesConfig_Unmarshal(b *testing.B) {
	data := []byte(`{"event":"INSERT","schema":"public","table":"users","filter":"id=eq.123"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var config PostgresChangesConfig
		_ = json.Unmarshal(data, &config)
	}
}

func BenchmarkClientMessage_WithPayload_Unmarshal(b *testing.B) {
	data := []byte(`{"type":"broadcast","channel":"room:123","payload":{"message":"hello world","user":"john","timestamp":1234567890}}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var msg ClientMessage
		_ = json.Unmarshal(data, &msg)
	}
}
