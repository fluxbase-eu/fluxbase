package mcp

import (
	"encoding/json"
	"sync"

	"github.com/rs/zerolog/log"
)

// Server-initiated notification methods (MCP spec).
const (
	NotificationToolsListChanged     = "notifications/tools/list_changed"
	NotificationResourcesListChanged = "notifications/resources/list_changed"
	NotificationResourcesUpdated     = "notifications/resources/updated"
	NotificationProgress             = "notifications/progress"
)

// notificationBuffer is the per-subscriber channel capacity. A subscriber
// that falls further behind drops notifications (they are advisory); this
// keeps one slow SSE client from stalling the server.
const notificationBuffer = 16

// notificationBus fans server-initiated JSON-RPC notifications out to
// connected SSE GET streams. Subscribers are anonymous: notifications carry
// only list-changed hints, so one broadcast reaches all connected clients
// and each client refetches with its own auth.
type notificationBus struct {
	mu          sync.RWMutex
	subscribers map[int]chan []byte
	nextID      int
}

func newNotificationBus() *notificationBus {
	return &notificationBus{subscribers: make(map[int]chan []byte)}
}

// Subscribe registers a subscriber. The returned cancel func unregisters it
// and is safe to call multiple times.
func (b *notificationBus) Subscribe() (<-chan []byte, func()) {
	ch := make(chan []byte, notificationBuffer)

	b.mu.Lock()
	id := b.nextID
	b.nextID++
	b.subscribers[id] = ch
	b.mu.Unlock()

	cancel := func() {
		b.mu.Lock()
		if _, ok := b.subscribers[id]; ok {
			delete(b.subscribers, id)
			close(ch)
		}
		b.mu.Unlock()
	}
	return ch, cancel
}

// broadcast delivers a serialized notification to all subscribers without
// blocking: full subscribers drop the message.
func (b *notificationBus) broadcast(data []byte) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subscribers {
		select {
		case ch <- data:
		default:
			log.Debug().Int("dropped_for", len(b.subscribers)).Msg("MCP: notification dropped for slow subscriber")
		}
	}
}

// SendNotification broadcasts a JSON-RPC 2.0 notification (no id) with the
// given method and optional params to all connected SSE streams.
func (s *Server) SendNotification(method string, params any) {
	payload := struct {
		JSONRPC string          `json:"jsonrpc"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params,omitempty"`
	}{JSONRPC: JSONRPCVersion, Method: method}

	if params != nil {
		raw, err := json.Marshal(params)
		if err != nil {
			log.Warn().Err(err).Str("method", method).Msg("MCP: failed to marshal notification params")
			return
		}
		payload.Params = raw
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Warn().Err(err).Str("method", method).Msg("MCP: failed to marshal notification")
		return
	}

	s.bus.broadcast(data)
}

// subscriberCount reports the number of connected notification streams
// (exposed for tests and diagnostics).
func (s *Server) subscriberCount() int {
	s.bus.mu.RLock()
	defer s.bus.mu.RUnlock()
	return len(s.bus.subscribers)
}
