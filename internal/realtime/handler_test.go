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

func TestMessageType_Constants(t *testing.T) {
	t.Run("all message types are defined", func(t *testing.T) {
		assert.Equal(t, MessageType("subscribe"), MessageTypeSubscribe)
		assert.Equal(t, MessageType("unsubscribe"), MessageTypeUnsubscribe)
		assert.Equal(t, MessageType("heartbeat"), MessageTypeHeartbeat)
		assert.Equal(t, MessageType("broadcast"), MessageTypeBroadcast)
		assert.Equal(t, MessageType("presence"), MessageTypePresence)
		assert.Equal(t, MessageType("error"), MessageTypeError)
		assert.Equal(t, MessageType("ack"), MessageTypeAck)
		assert.Equal(t, MessageType("postgres_changes"), MessageTypeChange)
		assert.Equal(t, MessageType("access_token"), MessageTypeAccessToken)
		assert.Equal(t, MessageType("subscribe_logs"), MessageTypeSubscribeLogs)
		assert.Equal(t, MessageType("execution_log"), MessageTypeExecutionLog)
		assert.Equal(t, MessageType("subscribe_all_logs"), MessageTypeSubscribeAllLogs)
		assert.Equal(t, MessageType("log_entry"), MessageTypeLogEntry)
	})

	t.Run("message types are distinct", func(t *testing.T) {
		types := []MessageType{
			MessageTypeSubscribe,
			MessageTypeUnsubscribe,
			MessageTypeHeartbeat,
			MessageTypeBroadcast,
			MessageTypePresence,
			MessageTypeError,
			MessageTypeAck,
			MessageTypeChange,
			MessageTypeAccessToken,
			MessageTypeSubscribeLogs,
			MessageTypeExecutionLog,
			MessageTypeSubscribeAllLogs,
			MessageTypeLogEntry,
		}

		seen := make(map[MessageType]bool)
		for _, mt := range types {
			assert.False(t, seen[mt], "Duplicate message type: %s", mt)
			seen[mt] = true
		}
	})
}

// =============================================================================
// ClientMessage Tests
// =============================================================================

func TestClientMessage_Struct(t *testing.T) {
	t.Run("subscribe message", func(t *testing.T) {
		msg := ClientMessage{
			Type:    MessageTypeSubscribe,
			Channel: "test-channel",
			Event:   "INSERT",
			Schema:  "public",
			Table:   "users",
		}

		assert.Equal(t, MessageTypeSubscribe, msg.Type)
		assert.Equal(t, "test-channel", msg.Channel)
		assert.Equal(t, "INSERT", msg.Event)
		assert.Equal(t, "public", msg.Schema)
		assert.Equal(t, "users", msg.Table)
	})

	t.Run("broadcast message with payload", func(t *testing.T) {
		payload := json.RawMessage(`{"message": "hello"}`)
		msg := ClientMessage{
			Type:      MessageTypeBroadcast,
			Channel:   "room:123",
			Event:     "chat",
			Payload:   payload,
			MessageID: "msg-001",
		}

		assert.Equal(t, MessageTypeBroadcast, msg.Type)
		assert.Equal(t, "room:123", msg.Channel)
		assert.Equal(t, "msg-001", msg.MessageID)
		assert.NotNil(t, msg.Payload)
	})

	t.Run("access token message", func(t *testing.T) {
		msg := ClientMessage{
			Type:  MessageTypeAccessToken,
			Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		}

		assert.Equal(t, MessageTypeAccessToken, msg.Type)
		assert.NotEmpty(t, msg.Token)
	})
}

func TestClientMessage_JSON(t *testing.T) {
	t.Run("deserialize subscribe message", func(t *testing.T) {
		jsonData := `{
			"type": "subscribe",
			"channel": "realtime:public:users",
			"event": "*",
			"schema": "public",
			"table": "users"
		}`

		var msg ClientMessage
		err := json.Unmarshal([]byte(jsonData), &msg)

		require.NoError(t, err)
		assert.Equal(t, MessageTypeSubscribe, msg.Type)
		assert.Equal(t, "realtime:public:users", msg.Channel)
		assert.Equal(t, "*", msg.Event)
		assert.Equal(t, "public", msg.Schema)
		assert.Equal(t, "users", msg.Table)
	})

	t.Run("deserialize postgres_changes message", func(t *testing.T) {
		jsonData := `{
			"type": "postgres_changes",
			"config": {
				"event": "INSERT",
				"schema": "public",
				"table": "orders",
				"filter": "user_id=eq.123"
			}
		}`

		var msg ClientMessage
		err := json.Unmarshal([]byte(jsonData), &msg)

		require.NoError(t, err)
		assert.Equal(t, MessageTypeChange, msg.Type)
		assert.NotNil(t, msg.Config)
	})
}

// =============================================================================
// ServerMessage Tests
// =============================================================================

func TestServerMessage_Struct(t *testing.T) {
	t.Run("success message", func(t *testing.T) {
		msg := ServerMessage{
			Type:    MessageTypeAck,
			Channel: "test-channel",
			Payload: map[string]interface{}{
				"status": "subscribed",
			},
		}

		assert.Equal(t, MessageTypeAck, msg.Type)
		assert.Equal(t, "test-channel", msg.Channel)
		assert.NotNil(t, msg.Payload)
	})

	t.Run("error message", func(t *testing.T) {
		msg := ServerMessage{
			Type:  MessageTypeError,
			Error: "Permission denied",
		}

		assert.Equal(t, MessageTypeError, msg.Type)
		assert.Equal(t, "Permission denied", msg.Error)
	})
}

func TestServerMessage_JSON(t *testing.T) {
	t.Run("serialize server message", func(t *testing.T) {
		msg := ServerMessage{
			Type:    MessageTypeChange,
			Channel: "realtime:public:users",
			Payload: map[string]interface{}{
				"event": "INSERT",
				"new": map[string]interface{}{
					"id":   "123",
					"name": "Test User",
				},
			},
		}

		data, err := json.Marshal(msg)

		require.NoError(t, err)
		assert.Contains(t, string(data), `"type":"postgres_changes"`)
		assert.Contains(t, string(data), `"channel":"realtime:public:users"`)
	})
}

// =============================================================================
// LogSubscriptionConfig Tests
// =============================================================================

func TestLogSubscriptionConfig_Struct(t *testing.T) {
	t.Run("function log subscription", func(t *testing.T) {
		config := LogSubscriptionConfig{
			ExecutionID: "exec-123",
			Type:        "function",
		}

		assert.Equal(t, "exec-123", config.ExecutionID)
		assert.Equal(t, "function", config.Type)
	})

	t.Run("job log subscription", func(t *testing.T) {
		config := LogSubscriptionConfig{
			ExecutionID: "job-456",
			Type:        "job",
		}

		assert.Equal(t, "job-456", config.ExecutionID)
		assert.Equal(t, "job", config.Type)
	})

	t.Run("rpc log subscription", func(t *testing.T) {
		config := LogSubscriptionConfig{
			ExecutionID: "rpc-789",
			Type:        "rpc",
		}

		assert.Equal(t, "rpc-789", config.ExecutionID)
		assert.Equal(t, "rpc", config.Type)
	})
}

// =============================================================================
// PostgresChangesConfig Tests
// =============================================================================

func TestPostgresChangesConfig_Struct(t *testing.T) {
	t.Run("INSERT event config", func(t *testing.T) {
		config := PostgresChangesConfig{
			Event:  "INSERT",
			Schema: "public",
			Table:  "users",
		}

		assert.Equal(t, "INSERT", config.Event)
		assert.Equal(t, "public", config.Schema)
		assert.Equal(t, "users", config.Table)
	})

	t.Run("wildcard event with filter", func(t *testing.T) {
		config := PostgresChangesConfig{
			Event:  "*",
			Schema: "public",
			Table:  "orders",
			Filter: "status=eq.pending",
		}

		assert.Equal(t, "*", config.Event)
		assert.Equal(t, "status=eq.pending", config.Filter)
	})
}

func TestPostgresChangesConfig_JSON(t *testing.T) {
	t.Run("deserialize config", func(t *testing.T) {
		jsonData := `{
			"event": "UPDATE",
			"schema": "public",
			"table": "products",
			"filter": "price=gt.100"
		}`

		var config PostgresChangesConfig
		err := json.Unmarshal([]byte(jsonData), &config)

		require.NoError(t, err)
		assert.Equal(t, "UPDATE", config.Event)
		assert.Equal(t, "public", config.Schema)
		assert.Equal(t, "products", config.Table)
		assert.Equal(t, "price=gt.100", config.Filter)
	})
}

// =============================================================================
// TokenClaims Tests
// =============================================================================

func TestTokenClaims_Struct(t *testing.T) {
	t.Run("authenticated user claims", func(t *testing.T) {
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

	t.Run("admin user claims", func(t *testing.T) {
		claims := TokenClaims{
			UserID: "admin-001",
			Role:   "dashboard_admin",
			RawClaims: map[string]interface{}{
				"sub":  "admin-001",
				"role": "dashboard_admin",
			},
		}

		assert.Equal(t, "dashboard_admin", claims.Role)
	})

	t.Run("anonymous claims", func(t *testing.T) {
		claims := TokenClaims{
			Role: "anon",
		}

		assert.Equal(t, "anon", claims.Role)
		assert.Empty(t, claims.UserID)
	})

	t.Run("custom claims for RLS", func(t *testing.T) {
		claims := TokenClaims{
			UserID: "player-123",
			Role:   "authenticated",
			RawClaims: map[string]interface{}{
				"sub":        "player-123",
				"role":       "authenticated",
				"meeting_id": "meeting-456",
				"player_id":  "player-123",
				"team":       "blue",
			},
		}

		assert.Equal(t, "meeting-456", claims.RawClaims["meeting_id"])
		assert.Equal(t, "blue", claims.RawClaims["team"])
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
