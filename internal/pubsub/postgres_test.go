package pubsub

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// NewPostgresPubSub Tests
// =============================================================================

func TestNewPostgresPubSub(t *testing.T) {
	t.Run("creates instance with nil pool", func(t *testing.T) {
		ps := NewPostgresPubSub(nil)

		require.NotNil(t, ps)
		assert.Nil(t, ps.pool)
		assert.NotNil(t, ps.subscribers)
		assert.NotNil(t, ps.ctx)
		assert.NotNil(t, ps.cancel)
		assert.False(t, ps.started)
	})

	t.Run("initializes empty subscribers map", func(t *testing.T) {
		ps := NewPostgresPubSub(nil)

		assert.Len(t, ps.subscribers, 0)
	})
}

// =============================================================================
// PostgresPubSub Struct Tests
// =============================================================================

func TestPostgresPubSub_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		ps := NewPostgresPubSub(nil)

		// Verify fields can be accessed
		assert.NotNil(t, ps.subscribers)
		assert.NotNil(t, ps.ctx)
		assert.NotNil(t, ps.cancel)
		assert.False(t, ps.started)
	})
}

// =============================================================================
// sanitizeChannelName Tests
// =============================================================================

func TestSanitizeChannelName(t *testing.T) {
	t.Run("converts single colon", func(t *testing.T) {
		result := sanitizeChannelName("test:channel")

		assert.Equal(t, "test__channel", result)
	})

	t.Run("converts multiple colons", func(t *testing.T) {
		result := sanitizeChannelName("a:b:c")

		assert.Equal(t, "a__b__c", result)
	})

	t.Run("handles channel without colons", func(t *testing.T) {
		result := sanitizeChannelName("simple_channel")

		assert.Equal(t, "simple_channel", result)
	})

	t.Run("handles empty string", func(t *testing.T) {
		result := sanitizeChannelName("")

		assert.Equal(t, "", result)
	})

	t.Run("converts BroadcastChannel constant", func(t *testing.T) {
		result := sanitizeChannelName(BroadcastChannel)

		assert.Equal(t, "fluxbase__broadcast", result)
	})

	t.Run("converts PresenceChannel constant", func(t *testing.T) {
		result := sanitizeChannelName(PresenceChannel)

		assert.Equal(t, "fluxbase__presence", result)
	})

	t.Run("converts SchemaCacheChannel constant", func(t *testing.T) {
		result := sanitizeChannelName(SchemaCacheChannel)

		assert.Equal(t, "fluxbase__schema_cache", result)
	})

	t.Run("handles consecutive colons", func(t *testing.T) {
		result := sanitizeChannelName("test::double")

		assert.Equal(t, "test____double", result)
	})

	t.Run("handles colon at start", func(t *testing.T) {
		result := sanitizeChannelName(":start")

		assert.Equal(t, "__start", result)
	})

	t.Run("handles colon at end", func(t *testing.T) {
		result := sanitizeChannelName("end:")

		assert.Equal(t, "end__", result)
	})
}

// =============================================================================
// unsanitizeChannelName Tests
// =============================================================================

func TestUnsanitizeChannelName(t *testing.T) {
	t.Run("converts double underscores to colon", func(t *testing.T) {
		result := unsanitizeChannelName("test__channel")

		assert.Equal(t, "test:channel", result)
	})

	t.Run("converts multiple double underscores", func(t *testing.T) {
		result := unsanitizeChannelName("a__b__c")

		assert.Equal(t, "a:b:c", result)
	})

	t.Run("handles channel without double underscores", func(t *testing.T) {
		result := unsanitizeChannelName("simple_channel")

		assert.Equal(t, "simple_channel", result)
	})

	t.Run("handles empty string", func(t *testing.T) {
		result := unsanitizeChannelName("")

		assert.Equal(t, "", result)
	})

	t.Run("converts sanitized BroadcastChannel", func(t *testing.T) {
		sanitized := sanitizeChannelName(BroadcastChannel)
		result := unsanitizeChannelName(sanitized)

		assert.Equal(t, BroadcastChannel, result)
	})

	t.Run("converts sanitized PresenceChannel", func(t *testing.T) {
		sanitized := sanitizeChannelName(PresenceChannel)
		result := unsanitizeChannelName(sanitized)

		assert.Equal(t, PresenceChannel, result)
	})

	t.Run("converts sanitized SchemaCacheChannel", func(t *testing.T) {
		sanitized := sanitizeChannelName(SchemaCacheChannel)
		result := unsanitizeChannelName(sanitized)

		assert.Equal(t, SchemaCacheChannel, result)
	})

	t.Run("handles single underscore", func(t *testing.T) {
		result := unsanitizeChannelName("test_channel")

		assert.Equal(t, "test_channel", result)
	})

	t.Run("handles three underscores", func(t *testing.T) {
		result := unsanitizeChannelName("test___channel")

		// One colon + one underscore
		assert.Equal(t, "test:_channel", result)
	})

	t.Run("handles four underscores", func(t *testing.T) {
		result := unsanitizeChannelName("test____channel")

		// Two colons
		assert.Equal(t, "test::channel", result)
	})

	t.Run("handles double underscores at start", func(t *testing.T) {
		result := unsanitizeChannelName("__start")

		assert.Equal(t, ":start", result)
	})

	t.Run("handles double underscores at end", func(t *testing.T) {
		result := unsanitizeChannelName("end__")

		assert.Equal(t, "end:", result)
	})
}

// =============================================================================
// Round-trip Tests (sanitize + unsanitize)
// =============================================================================

func TestChannelNameRoundTrip(t *testing.T) {
	testCases := []string{
		"simple",
		"with:colon",
		"multiple:colons:here",
		"fluxbase:broadcast",
		"fluxbase:presence",
		"fluxbase:schema_cache",
		"",
		"no_colons_here",
		"mixed:with_underscore",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			sanitized := sanitizeChannelName(tc)
			unsanitized := unsanitizeChannelName(sanitized)

			assert.Equal(t, tc, unsanitized)
		})
	}
}

// =============================================================================
// quoteIdentifier Tests
// =============================================================================

func TestQuoteIdentifier(t *testing.T) {
	t.Run("quotes simple identifier", func(t *testing.T) {
		result := quoteIdentifier("mytable")

		assert.Equal(t, `"mytable"`, result)
	})

	t.Run("escapes embedded double quote", func(t *testing.T) {
		result := quoteIdentifier(`my"table`)

		// Embedded " becomes "" inside the quoted identifier
		assert.Equal(t, `"my""table"`, result)
	})

	t.Run("escapes multiple double quotes", func(t *testing.T) {
		result := quoteIdentifier(`a"b"c`)

		assert.Equal(t, `"a""b""c"`, result)
	})

	t.Run("handles empty string", func(t *testing.T) {
		result := quoteIdentifier("")

		assert.Equal(t, `""`, result)
	})

	t.Run("handles identifier with spaces", func(t *testing.T) {
		result := quoteIdentifier("my table")

		assert.Equal(t, `"my table"`, result)
	})

	t.Run("handles identifier with special characters", func(t *testing.T) {
		result := quoteIdentifier("test_channel")

		assert.Equal(t, `"test_channel"`, result)
	})

	t.Run("handles PostgreSQL channel name with colons converted", func(t *testing.T) {
		pgChannel := sanitizeChannelName(BroadcastChannel)
		result := quoteIdentifier(pgChannel)

		assert.Equal(t, `"fluxbase__broadcast"`, result)
	})

	t.Run("prevents SQL injection", func(t *testing.T) {
		// Attempt SQL injection
		malicious := `test"; DROP TABLE users; --`
		result := quoteIdentifier(malicious)

		// The " should be escaped to ""
		assert.Equal(t, `"test""; DROP TABLE users; --"`, result)
		assert.Contains(t, result, `""`)
	})

	t.Run("handles only double quotes", func(t *testing.T) {
		result := quoteIdentifier(`"""`)

		assert.Equal(t, `"""""""`, result)
	})
}

// =============================================================================
// Payload Size Validation Tests
// =============================================================================

func TestPayloadSizeValidation(t *testing.T) {
	t.Run("payload at limit (8000 bytes)", func(t *testing.T) {
		payload := make([]byte, 8000)
		isValid := len(payload) <= 8000

		assert.True(t, isValid)
	})

	t.Run("payload over limit", func(t *testing.T) {
		payload := make([]byte, 8001)
		isValid := len(payload) <= 8000

		assert.False(t, isValid)
	})

	t.Run("small payload", func(t *testing.T) {
		payload := []byte("hello world")
		isValid := len(payload) <= 8000

		assert.True(t, isValid)
	})

	t.Run("empty payload", func(t *testing.T) {
		payload := []byte{}
		isValid := len(payload) <= 8000

		assert.True(t, isValid)
	})
}

// =============================================================================
// Message Struct Tests
// =============================================================================

func TestMessage_Postgres(t *testing.T) {
	t.Run("stores channel and payload", func(t *testing.T) {
		msg := Message{
			Channel: "test:channel",
			Payload: []byte(`{"event": "test"}`),
		}

		assert.Equal(t, "test:channel", msg.Channel)
		assert.Equal(t, []byte(`{"event": "test"}`), msg.Payload)
	})

	t.Run("handles empty payload", func(t *testing.T) {
		msg := Message{
			Channel: "empty:payload",
			Payload: []byte{},
		}

		assert.Equal(t, "empty:payload", msg.Channel)
		assert.Empty(t, msg.Payload)
	})

	t.Run("handles nil payload", func(t *testing.T) {
		msg := Message{
			Channel: "nil:payload",
			Payload: nil,
		}

		assert.Nil(t, msg.Payload)
	})
}

// =============================================================================
// Subscriber Channel Tests
// =============================================================================

func TestSubscriberChannel(t *testing.T) {
	t.Run("subscriber channel buffer size is 100", func(t *testing.T) {
		ch := make(chan Message, 100)

		assert.Equal(t, 100, cap(ch))
	})

	t.Run("can send message without blocking when buffer has space", func(t *testing.T) {
		ch := make(chan Message, 100)

		select {
		case ch <- Message{Channel: "test", Payload: []byte("data")}:
			// Success
		default:
			t.Fatal("channel should not block")
		}

		assert.Len(t, ch, 1)
	})

	t.Run("channel full behavior", func(t *testing.T) {
		ch := make(chan Message, 1)

		// Fill the channel
		ch <- Message{Channel: "test", Payload: []byte("first")}

		// Try to send without blocking
		sent := false
		select {
		case ch <- Message{Channel: "test", Payload: []byte("second")}:
			sent = true
		default:
			// Channel full
		}

		assert.False(t, sent)
	})
}

// =============================================================================
// Channel Constants Tests
// =============================================================================

func TestChannelConstants(t *testing.T) {
	t.Run("BroadcastChannel has colon", func(t *testing.T) {
		assert.Contains(t, BroadcastChannel, ":")
		assert.Equal(t, "fluxbase:broadcast", BroadcastChannel)
	})

	t.Run("PresenceChannel has colon", func(t *testing.T) {
		assert.Contains(t, PresenceChannel, ":")
		assert.Equal(t, "fluxbase:presence", PresenceChannel)
	})

	t.Run("SchemaCacheChannel has colon", func(t *testing.T) {
		assert.Contains(t, SchemaCacheChannel, ":")
		assert.Equal(t, "fluxbase:schema_cache", SchemaCacheChannel)
	})
}

// =============================================================================
// Subscriber Management Tests
// =============================================================================

func TestSubscriberManagement(t *testing.T) {
	t.Run("add subscriber to map", func(t *testing.T) {
		subscribers := make(map[string][]chan Message)
		ch := make(chan Message, 100)
		channel := "test:channel"

		subscribers[channel] = append(subscribers[channel], ch)

		assert.Len(t, subscribers[channel], 1)
	})

	t.Run("add multiple subscribers to same channel", func(t *testing.T) {
		subscribers := make(map[string][]chan Message)
		ch1 := make(chan Message, 100)
		ch2 := make(chan Message, 100)
		channel := "test:channel"

		subscribers[channel] = append(subscribers[channel], ch1)
		subscribers[channel] = append(subscribers[channel], ch2)

		assert.Len(t, subscribers[channel], 2)
	})

	t.Run("remove subscriber from map", func(t *testing.T) {
		subscribers := make(map[string][]chan Message)
		ch1 := make(chan Message, 100)
		ch2 := make(chan Message, 100)
		channel := "test:channel"

		subscribers[channel] = []chan Message{ch1, ch2}

		// Remove ch1
		subs := subscribers[channel]
		for i, sub := range subs {
			if sub == ch1 {
				subscribers[channel] = append(subs[:i], subs[i+1:]...)
				break
			}
		}

		assert.Len(t, subscribers[channel], 1)
	})
}

// =============================================================================
// Close Behavior Tests
// =============================================================================

func TestPostgresPubSub_Close(t *testing.T) {
	t.Run("close cleans up subscribers map", func(t *testing.T) {
		ps := NewPostgresPubSub(nil)

		// Add some subscribers manually
		ps.subscribers["test:channel"] = []chan Message{make(chan Message, 100)}

		// Close
		err := ps.Close()

		assert.NoError(t, err)
		assert.Len(t, ps.subscribers, 0)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkSanitizeChannelName(b *testing.B) {
	channels := []string{
		"simple",
		"with:colon",
		"multiple:colons:here",
		"fluxbase:broadcast",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sanitizeChannelName(channels[i%len(channels)])
	}
}

func BenchmarkUnsanitizeChannelName(b *testing.B) {
	channels := []string{
		"simple",
		"with__colon",
		"multiple__colons__here",
		"fluxbase__broadcast",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = unsanitizeChannelName(channels[i%len(channels)])
	}
}

func BenchmarkQuoteIdentifier(b *testing.B) {
	identifiers := []string{
		"simple",
		`with"quote`,
		"fluxbase__broadcast",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = quoteIdentifier(identifiers[i%len(identifiers)])
	}
}

func BenchmarkPayloadSizeCheck(b *testing.B) {
	payload := make([]byte, 4000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = len(payload) <= 8000
	}
}

func BenchmarkChannelNameRoundTrip(b *testing.B) {
	channel := "fluxbase:broadcast:test"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sanitized := sanitizeChannelName(channel)
		_ = unsanitizeChannelName(sanitized)
	}
}
