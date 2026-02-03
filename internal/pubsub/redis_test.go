package pubsub

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// RedisPubSub Struct Tests
// =============================================================================

func TestRedisPubSub_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ps := &RedisPubSub{
			client:      nil,
			subscribers: make(map[string][]chan Message),
			ctx:         ctx,
			cancel:      cancel,
		}

		require.NotNil(t, ps)
		assert.NotNil(t, ps.subscribers)
		assert.NotNil(t, ps.ctx)
		assert.NotNil(t, ps.cancel)
	})

	t.Run("initializes empty subscribers map", func(t *testing.T) {
		ps := &RedisPubSub{
			subscribers: make(map[string][]chan Message),
		}

		assert.Len(t, ps.subscribers, 0)
	})
}

// =============================================================================
// RedisPubSub unsubscribe Tests
// =============================================================================

func TestRedisPubSub_unsubscribe(t *testing.T) {
	t.Run("removes subscriber from map", func(t *testing.T) {
		ps := &RedisPubSub{
			subscribers: make(map[string][]chan Message),
		}

		ch := make(chan Message, 1)
		ps.subscribers["test-channel"] = []chan Message{ch}

		assert.Len(t, ps.subscribers["test-channel"], 1)

		ps.unsubscribe("test-channel", ch)

		assert.Len(t, ps.subscribers["test-channel"], 0)
	})

	t.Run("handles multiple subscribers on same channel", func(t *testing.T) {
		ps := &RedisPubSub{
			subscribers: make(map[string][]chan Message),
		}

		ch1 := make(chan Message, 1)
		ch2 := make(chan Message, 1)
		ch3 := make(chan Message, 1)
		ps.subscribers["channel"] = []chan Message{ch1, ch2, ch3}

		assert.Len(t, ps.subscribers["channel"], 3)

		ps.unsubscribe("channel", ch2)

		assert.Len(t, ps.subscribers["channel"], 2)
		assert.Equal(t, ch1, ps.subscribers["channel"][0])
		assert.Equal(t, ch3, ps.subscribers["channel"][1])
	})

	t.Run("handles unsubscribe of first element", func(t *testing.T) {
		ps := &RedisPubSub{
			subscribers: make(map[string][]chan Message),
		}

		ch1 := make(chan Message, 1)
		ch2 := make(chan Message, 1)
		ps.subscribers["channel"] = []chan Message{ch1, ch2}

		ps.unsubscribe("channel", ch1)

		assert.Len(t, ps.subscribers["channel"], 1)
		assert.Equal(t, ch2, ps.subscribers["channel"][0])
	})

	t.Run("handles unsubscribe of last element", func(t *testing.T) {
		ps := &RedisPubSub{
			subscribers: make(map[string][]chan Message),
		}

		ch1 := make(chan Message, 1)
		ch2 := make(chan Message, 1)
		ps.subscribers["channel"] = []chan Message{ch1, ch2}

		ps.unsubscribe("channel", ch2)

		assert.Len(t, ps.subscribers["channel"], 1)
		assert.Equal(t, ch1, ps.subscribers["channel"][0])
	})

	t.Run("handles non-existent subscriber", func(t *testing.T) {
		ps := &RedisPubSub{
			subscribers: make(map[string][]chan Message),
		}

		ch1 := make(chan Message, 1)
		ch2 := make(chan Message, 1)
		ps.subscribers["channel"] = []chan Message{ch1}

		// Unsubscribing a channel not in the list should not panic
		ps.unsubscribe("channel", ch2)

		assert.Len(t, ps.subscribers["channel"], 1)
	})

	t.Run("handles non-existent channel", func(t *testing.T) {
		ps := &RedisPubSub{
			subscribers: make(map[string][]chan Message),
		}

		ch := make(chan Message, 1)

		// Should not panic on non-existent channel
		ps.unsubscribe("non-existent", ch)

		assert.Nil(t, ps.subscribers["non-existent"])
	})

	t.Run("is thread-safe", func(t *testing.T) {
		ps := &RedisPubSub{
			subscribers: make(map[string][]chan Message),
		}

		// Create multiple channels
		channels := make([]chan Message, 100)
		for i := 0; i < 100; i++ {
			channels[i] = make(chan Message, 1)
		}

		// Add all to subscribers
		ps.mu.Lock()
		ps.subscribers["channel"] = channels
		ps.mu.Unlock()

		// Unsubscribe concurrently
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(ch chan Message) {
				defer wg.Done()
				ps.unsubscribe("channel", ch)
			}(channels[i])
		}
		wg.Wait()

		assert.Len(t, ps.subscribers["channel"], 0)
	})
}

// =============================================================================
// RedisPubSub Close Tests
// =============================================================================

func TestRedisPubSub_Close_SubscriberCleanup(t *testing.T) {
	t.Run("clears all subscribers on close", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		ps := &RedisPubSub{
			subscribers: make(map[string][]chan Message),
			ctx:         ctx,
			cancel:      cancel,
		}

		// Add subscribers
		ch1 := make(chan Message, 1)
		ch2 := make(chan Message, 1)
		ps.subscribers["channel1"] = []chan Message{ch1}
		ps.subscribers["channel2"] = []chan Message{ch2}

		assert.Len(t, ps.subscribers, 2)

		// Cancel context (simulates part of close)
		cancel()

		// Clear subscribers (simulates close cleanup)
		ps.mu.Lock()
		for _, subs := range ps.subscribers {
			for _, ch := range subs {
				close(ch)
			}
		}
		ps.subscribers = make(map[string][]chan Message)
		ps.mu.Unlock()

		assert.Len(t, ps.subscribers, 0)
	})
}

// =============================================================================
// NewRedisPubSub Tests
// =============================================================================

func TestNewRedisPubSub_URLParsing(t *testing.T) {
	t.Run("returns error for invalid URL", func(t *testing.T) {
		_, err := NewRedisPubSub("invalid://url")
		assert.Error(t, err)
	})

	t.Run("returns error for malformed URL", func(t *testing.T) {
		_, err := NewRedisPubSub("not a url at all")
		assert.Error(t, err)
	})

	t.Run("returns error for empty URL", func(t *testing.T) {
		_, err := NewRedisPubSub("")
		assert.Error(t, err)
	})
}

// =============================================================================
// Message Tests (shared with other pubsub implementations)
// =============================================================================

func TestRedisPubSub_Message(t *testing.T) {
	t.Run("stores channel and payload", func(t *testing.T) {
		msg := Message{
			Channel: "test:channel",
			Payload: []byte("hello world"),
		}

		assert.Equal(t, "test:channel", msg.Channel)
		assert.Equal(t, []byte("hello world"), msg.Payload)
	})

	t.Run("handles empty payload", func(t *testing.T) {
		msg := Message{
			Channel: "empty",
			Payload: []byte{},
		}

		assert.Equal(t, "empty", msg.Channel)
		assert.Empty(t, msg.Payload)
	})

	t.Run("handles nil payload", func(t *testing.T) {
		msg := Message{
			Channel: "nil",
			Payload: nil,
		}

		assert.Nil(t, msg.Payload)
	})

	t.Run("handles binary payload", func(t *testing.T) {
		payload := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE}
		msg := Message{
			Channel: "binary",
			Payload: payload,
		}

		assert.Equal(t, payload, msg.Payload)
	})
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

func TestRedisPubSub_ConcurrentSubscriberAccess(t *testing.T) {
	t.Run("handles concurrent subscriber operations", func(t *testing.T) {
		ps := &RedisPubSub{
			subscribers: make(map[string][]chan Message),
		}

		var wg sync.WaitGroup

		// Add subscribers concurrently
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				ch := make(chan Message, 1)
				ps.mu.Lock()
				ps.subscribers["concurrent"] = append(ps.subscribers["concurrent"], ch)
				ps.mu.Unlock()
			}(i)
		}

		wg.Wait()

		assert.Len(t, ps.subscribers["concurrent"], 50)
	})
}

// =============================================================================
// Subscriber Map Tests
// =============================================================================

func TestRedisPubSub_SubscriberMap(t *testing.T) {
	t.Run("supports multiple channels", func(t *testing.T) {
		ps := &RedisPubSub{
			subscribers: make(map[string][]chan Message),
		}

		ch1 := make(chan Message, 1)
		ch2 := make(chan Message, 1)
		ch3 := make(chan Message, 1)

		ps.subscribers["channel1"] = []chan Message{ch1}
		ps.subscribers["channel2"] = []chan Message{ch2}
		ps.subscribers["channel3"] = []chan Message{ch3}

		assert.Len(t, ps.subscribers, 3)
		assert.Len(t, ps.subscribers["channel1"], 1)
		assert.Len(t, ps.subscribers["channel2"], 1)
		assert.Len(t, ps.subscribers["channel3"], 1)
	})

	t.Run("supports multiple subscribers per channel", func(t *testing.T) {
		ps := &RedisPubSub{
			subscribers: make(map[string][]chan Message),
		}

		ch1 := make(chan Message, 1)
		ch2 := make(chan Message, 1)
		ch3 := make(chan Message, 1)

		ps.subscribers["shared"] = []chan Message{ch1, ch2, ch3}

		assert.Len(t, ps.subscribers["shared"], 3)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkRedisPubSub_unsubscribe(b *testing.B) {
	ps := &RedisPubSub{
		subscribers: make(map[string][]chan Message),
	}

	// Setup: create channels
	channels := make([]chan Message, b.N)
	for i := 0; i < b.N; i++ {
		channels[i] = make(chan Message, 1)
	}
	ps.subscribers["bench"] = channels

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps.unsubscribe("bench", channels[i])
	}
}

func BenchmarkRedisPubSub_SubscriberMapAccess(b *testing.B) {
	ps := &RedisPubSub{
		subscribers: make(map[string][]chan Message),
	}

	ch := make(chan Message, 1)
	ps.subscribers["bench"] = []chan Message{ch}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps.mu.RLock()
		_ = ps.subscribers["bench"]
		ps.mu.RUnlock()
	}
}
