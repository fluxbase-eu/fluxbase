package mcp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/config"
)

func TestNotificationBus_SubscribeBroadcast(t *testing.T) {
	bus := newNotificationBus()

	ch, cancel := bus.Subscribe()
	defer cancel()

	bus.broadcast([]byte(`{"jsonrpc":"2.0","method":"notifications/tools/list_changed"}`))

	select {
	case data := <-ch:
		assert.Contains(t, string(data), "notifications/tools/list_changed")
	case <-time.After(time.Second):
		t.Fatal("subscriber did not receive the broadcast")
	}
}

func TestNotificationBus_CancelStopsDelivery(t *testing.T) {
	bus := newNotificationBus()

	ch, cancel := bus.Subscribe()
	cancel()
	cancel() // must be idempotent

	// Channel is closed on cancel; receiving yields immediately.
	_, open := <-ch
	assert.False(t, open, "channel should be closed after cancel")

	// Broadcast after cancel must not panic.
	assert.NotPanics(t, func() {
		bus.broadcast([]byte(`{}`))
	})
}

func TestNotificationBus_SlowSubscriberDrops(t *testing.T) {
	bus := newNotificationBus()

	ch, cancel := bus.Subscribe()
	defer cancel()

	// Fill the buffer without reading.
	for i := 0; i < notificationBuffer*2; i++ {
		bus.broadcast([]byte(`{"n":1}`))
	}

	// Drain what fit; count received.
	received := 0
	for {
		select {
		case <-ch:
			received++
			continue
		default:
		}
		break
	}
	assert.Equal(t, notificationBuffer, received, "buffer holds exactly notificationBuffer entries; overflow drops")
}

func TestNotificationBus_MultipleSubscribers(t *testing.T) {
	bus := newNotificationBus()

	ch1, cancel1 := bus.Subscribe()
	ch2, cancel2 := bus.Subscribe()
	defer cancel2()

	bus.broadcast([]byte(`{"m":1}`))

	var wg sync.WaitGroup
	wg.Add(2)
	for _, ch := range []<-chan []byte{ch1, ch2} {
		go func(ch <-chan []byte) {
			defer wg.Done()
			select {
			case <-ch:
			case <-time.After(time.Second):
				t.Error("subscriber did not receive broadcast")
			}
		}(ch)
	}
	wg.Wait()
	cancel1()
}

func TestServer_SendNotification(t *testing.T) {
	server := NewServer(&config.MCPConfig{})

	ch, cancel := server.bus.Subscribe()
	defer cancel()

	server.SendNotification(NotificationToolsListChanged, nil)

	select {
	case data := <-ch:
		var payload struct {
			JSONRPC string          `json:"jsonrpc"`
			Method  string          `json:"method"`
			Params  json.RawMessage `json:"params"`
		}
		require.NoError(t, json.Unmarshal(data, &payload))
		assert.Equal(t, "2.0", payload.JSONRPC)
		assert.Equal(t, "notifications/tools/list_changed", payload.Method)
		assert.Empty(t, payload.Params, "nil params must be omitted")
	case <-time.After(time.Second):
		t.Fatal("no notification received")
	}

	server.SendNotification(NotificationResourcesUpdated, map[string]any{"uri": "schema://public/users"})
	select {
	case data := <-ch:
		var payload struct {
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		require.NoError(t, json.Unmarshal(data, &payload))
		assert.Equal(t, "notifications/resources/updated", payload.Method)
		assert.Contains(t, string(payload.Params), "schema://public/users")
	case <-time.After(time.Second):
		t.Fatal("no notification received")
	}
}

func TestStreamNotifications(t *testing.T) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	notifications := make(chan []byte, 2)
	done := make(chan struct{})

	go func() {
		notifications <- []byte(`{"jsonrpc":"2.0","method":"notifications/tools/list_changed"}`)
		close(notifications)
	}()

	streamNotifications(w, notifications, done, time.Hour)

	flushErr := w.Flush()
	require.NoError(t, flushErr)

	out := buf.String()
	assert.Contains(t, out, ": connected\n\n")
	assert.Contains(t, out, "event: message\ndata: {\"jsonrpc\":\"2.0\",\"method\":\"notifications/tools/list_changed\"}\n\n")
}

func TestStreamNotifications_Pings(t *testing.T) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	notifications := make(chan []byte) // never sent
	done := make(chan struct{})

	go func() {
		time.Sleep(60 * time.Millisecond)
		close(done)
	}()

	streamNotifications(w, notifications, done, 20*time.Millisecond)
	require.NoError(t, w.Flush())

	assert.Contains(t, buf.String(), ": ping\n\n", "keep-alive pings should be emitted on the ticker")
}

func TestStreamNotifications_WriteErrorReturns(t *testing.T) {
	// A writer that always fails must terminate the loop instead of spinning.
	w := bufio.NewWriter(failingWriter{})

	notifications := make(chan []byte, 1)
	done := make(chan struct{})

	finished := make(chan struct{})
	go func() {
		streamNotifications(w, notifications, done, time.Hour)
		close(finished)
	}()

	notifications <- []byte(`{}`)

	select {
	case <-finished:
		// returned promptly on write error
	case <-time.After(time.Second):
		t.Fatal("streamNotifications did not return after a write error")
	}
}

type failingWriter struct{}

func (failingWriter) Write(p []byte) (int, error) { return 0, assert.AnError }
