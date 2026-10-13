package mcp

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/config"
)

// TestHandleGet_SSEEndToEnd runs the full notification path over a real TCP
// HTTP server: authenticated GET /mcp SSE connect, server broadcast, client
// receives the JSON-RPC notification frame, disconnect tears the
// subscription down.
func TestHandleGet_SSEEndToEnd(t *testing.T) {
	cfg := &config.MCPConfig{}
	handler := NewHandler(cfg, nil)
	handler.ssePing = 20 * time.Millisecond // fast keep-alive so client disconnects are noticed promptly

	app := fiber.New()
	// Simulate the auth middleware populating fiber locals.
	app.Get("/mcp", func(c fiber.Ctx) error {
		c.Locals("auth_type", "pat")
		uid := "user-1"
		c.Locals("user_id", &uid)
		c.Locals("user_role", "authenticated")
		return handler.HandleGet(c)
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = ln.Close() }()

	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		_ = app.Listener(ln)
	}()
	t.Cleanup(func() {
		_ = ln.Close()
		select {
		case <-serverDone:
		case <-time.After(2 * time.Second):
		}
	})

	baseURL := "http://" + ln.Addr().String()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/mcp", nil)
	require.NoError(t, err)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/event-stream")
	t.Cleanup(func() { _ = resp.Body.Close() })

	reader := bufio.NewReader(resp.Body)

	// The subscription exists while the stream is open.
	assert.Eventually(t, func() bool {
		return handler.Server().subscriberCount() == 1
	}, 2*time.Second, 5*time.Millisecond)

	// The first bytes are the connected comment.
	line, err := readSSELine(ctx, reader)
	require.NoError(t, err)
	assert.Equal(t, ": connected", line)

	// Broadcast a notification; the client must receive the framed event.
	handler.Server().SendNotification(NotificationToolsListChanged, map[string]any{"reason": "test"})
	frame := make([]string, 0, 2)
	for len(frame) < 2 {
		line, err := readSSELine(ctx, reader)
		require.NoError(t, err)
		if line == "" {
			continue // event separator
		}
		frame = append(frame, line)
	}
	assert.Equal(t, "event: message", frame[0])
	assert.True(t, strings.HasPrefix(frame[1], "data: {\"jsonrpc\":\"2.0\",\"method\":\"notifications/tools/list_changed\""),
		"notification payload must carry the JSON-RPC notification, got %q", frame[1])

	// Client disconnect tears the subscription down: the next keep-alive
	// write fails on the closed socket and the stream loop exits.
	cancel()
	assert.Eventually(t, func() bool {
		return handler.Server().subscriberCount() == 0
	}, 3*time.Second, 10*time.Millisecond, "disconnect must unsubscribe the stream")
}

// readSSELine reads one SSE line, honoring the context deadline.
func readSSELine(ctx context.Context, reader *bufio.Reader) (string, error) {
	type result struct {
		line string
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		line, err := reader.ReadString('\n')
		ch <- result{strings.TrimRight(line, "\r\n"), err}
	}()
	select {
	case r := <-ch:
		return r.line, r.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
