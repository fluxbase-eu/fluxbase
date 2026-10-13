package mcp

import (
	"bufio"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/config"
	"github.com/nimbleflux/fluxbase/internal/database"
	"github.com/nimbleflux/fluxbase/internal/middleware"
)

// rateLimiter tracks request counts per client key using a sliding window
type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time // client key -> list of request timestamps
	limit    int                    // requests per minute
}

// newRateLimiter creates a new rate limiter
func newRateLimiter(limitPerMin int) *rateLimiter {
	rl := &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limitPerMin,
	}
	// Start cleanup goroutine
	go rl.cleanup()
	return rl
}

// allow checks if a request from the given client key should be allowed
func (rl *rateLimiter) allow(clientKey string) bool {
	if rl.limit <= 0 {
		return true // Rate limiting disabled
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-time.Minute)

	// Get existing requests and filter to sliding window
	existing := rl.requests[clientKey]
	var valid []time.Time
	for _, t := range existing {
		if t.After(windowStart) {
			valid = append(valid, t)
		}
	}

	// Check if we're at limit
	if len(valid) >= rl.limit {
		rl.requests[clientKey] = valid // Update with filtered list
		return false
	}

	// Allow and record
	rl.requests[clientKey] = append(valid, now)
	return true
}

// cleanup periodically removes old entries to prevent memory growth
func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		windowStart := now.Add(-time.Minute)

		for key, times := range rl.requests {
			var valid []time.Time
			for _, t := range times {
				if t.After(windowStart) {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(rl.requests, key)
			} else {
				rl.requests[key] = valid
			}
		}
		rl.mu.Unlock()
	}
}

// Handler handles HTTP requests for the MCP server
type Handler struct {
	server      *Server
	config      *config.MCPConfig
	db          *database.Connection
	rateLimiter *rateLimiter
	// ssePingInterval is the keep-alive comment cadence for SSE GET
	// streams. Defaults to ssePingInterval; shortened in tests so teardown
	// paths don't wait 30s for the next write.
	ssePing time.Duration
}

// NewHandler creates a new MCP HTTP handler
func NewHandler(cfg *config.MCPConfig, db *database.Connection) *Handler {
	h := &Handler{
		server:      NewServer(cfg),
		config:      cfg,
		db:          db,
		rateLimiter: newRateLimiter(cfg.RateLimitPerMin),
		ssePing:     ssePingInterval,
	}
	// Enable the durable MCP audit trail when a database is available
	h.server.SetDB(db)
	return h
}

// Server returns the underlying MCP server for tool/resource registration
func (h *Handler) Server() *Server {
	return h.server
}

// HandleHealth handles health check requests
func (h *Handler) HandleHealth(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":          "healthy",
		"protocolVersion": MCPVersion,
		"serverVersion":   FluxbaseVersion,
	})
}

// HandlePost handles JSON-RPC POST requests
func (h *Handler) HandlePost(c fiber.Ctx) error {
	// Check content type
	contentType := c.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
			"error": "Content-Type must be application/json",
		})
	}

	// Check message size
	if h.config.MaxMessageSize > 0 && len(c.Body()) > h.config.MaxMessageSize {
		return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
			"error": "Request body too large",
		})
	}

	// Extract auth context from Fiber locals (set by auth middleware)
	authCtx := ExtractAuthContext(c)

	// Check rate limit using client key or user ID as the key
	rateLimitKey := rateLimitKeyFor(authCtx, c)
	if !h.rateLimiter.allow(rateLimitKey) {
		log.Warn().
			Str("client_key", rateLimitKey).
			Int("limit", h.config.RateLimitPerMin).
			Msg("MCP: Rate limit exceeded")
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error": "Rate limit exceeded. Please try again later.",
		})
	}

	// Log the request
	log.Debug().
		Str("auth_type", authCtx.AuthType).
		Str("user_role", authCtx.UserRole).
		Bool("authenticated", authCtx.IsAuthenticated()).
		Msg("MCP: Handling POST request")

	// Process the request
	response := h.server.HandleRequest(middleware.CtxWithTenant(c), c.Body(), authCtx)

	// Check Accept header for response format
	accept := c.Get("Accept")

	// For streaming responses, use SSE if client accepts it
	if strings.Contains(accept, "text/event-stream") {
		return h.sendSSEResponse(c, response)
	}

	// Default to JSON response
	return h.sendJSONResponse(c, response)
}

// ssePingInterval keeps intermediary proxies from closing an idle
// notification stream.
const ssePingInterval = 30 * time.Second

// HandleGet handles GET requests as long-lived SSE streams carrying
// server-initiated JSON-RPC notifications (e.g. notifications/tools/
// list_changed when a custom tool is reloaded). Same auth as POST; the
// stream ends when the client disconnects.
func (h *Handler) HandleGet(c fiber.Ctx) error {
	// Check if client accepts SSE
	accept := c.Get("Accept")
	if !strings.Contains(accept, "text/event-stream") {
		return c.Status(fiber.StatusNotAcceptable).JSON(fiber.Map{
			"error": "GET requests require Accept: text/event-stream header",
		})
	}

	// Same auth as POST: middleware populates the auth context.
	authCtx := ExtractAuthContext(c)
	if !authCtx.IsAuthenticated() {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// A long-lived connection consumes a rate-limit slot like any request.
	if !h.rateLimiter.allow(rateLimitKeyFor(authCtx, c)) {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error": "Rate limit exceeded. Please try again later.",
		})
	}

	notifications, cancel := h.server.bus.Subscribe()

	c.Set(fiber.HeaderContentType, "text/event-stream")
	c.Set(fiber.HeaderCacheControl, "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no") // disable nginx buffering

	// fasthttp keeps the request context alive for the duration of the body
	// stream writer, so capturing it here is safe; the fiber.Ctx itself is
	// not (it recycles when this handler returns).
	requestCtx := c.RequestCtx()

	requestCtx.SetBodyStreamWriter(func(w *bufio.Writer) {
		defer cancel()
		streamNotifications(w, notifications, requestCtx.Done(), h.ssePing)
	})

	return nil
}

// streamNotifications writes the SSE framing for one notification stream:
// a connected comment, one "event: message" frame per notification, and
// keep-alive ping comments on the ticker. Returns when the client goes away
// (done), the subscription is cancelled (channel closed), or a write fails.
func streamNotifications(w *bufio.Writer, notifications <-chan []byte, done <-chan struct{}, pingInterval time.Duration) {
	if _, err := w.WriteString(": connected\n\n"); err != nil {
		return
	}
	if err := w.Flush(); err != nil {
		return
	}

	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case data, ok := <-notifications:
			if !ok {
				return
			}
			if _, err := w.WriteString("event: message\ndata: " + string(data) + "\n\n"); err != nil {
				return
			}
			if err := w.Flush(); err != nil {
				return
			}
		case <-ticker.C:
			if _, err := w.WriteString(": ping\n\n"); err != nil {
				return
			}
			if err := w.Flush(); err != nil {
				return
			}
		}
	}
}

// rateLimitKeyFor resolves the rate-limit key for a request: client key,
// user ID, then IP (mirrors the POST path).
func rateLimitKeyFor(authCtx *AuthContext, c fiber.Ctx) string {
	if authCtx.ClientKeyID != "" {
		return authCtx.ClientKeyID
	}
	if authCtx.UserID != nil && *authCtx.UserID != "" {
		return *authCtx.UserID
	}
	return c.IP()
}

// sendJSONResponse sends a JSON-RPC response
func (h *Handler) sendJSONResponse(c fiber.Ctx, response *Response) error {
	c.Set("Content-Type", "application/json")
	return c.JSON(response)
}

// sendSSEResponse sends a response as an SSE event
func (h *Handler) sendSSEResponse(c fiber.Ctx, response *Response) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	// Serialize the response
	data, err := h.server.SerializeResponse(response)
	if err != nil {
		log.Error().Err(err).Msg("MCP: Failed to serialize response")
		return c.SendString("data: {\"error\":\"serialization error\"}\n\n")
	}

	// Send as SSE event - single event for now (not a stream)
	// Format: "data: <json>\n\n"
	return c.SendString("data: " + string(data) + "\n\n")
}

// fiber:context-methods migrated
