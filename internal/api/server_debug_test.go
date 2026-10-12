package api

import (
	"net/http"
	httptest "net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubUnauthorizedAuth stands in for the real auth middleware: it rejects
// every request, which is enough to prove the routes exist and are gated.
func stubUnauthorizedAuth(c fiber.Ctx) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
}

func TestDebugPprofRoutes_DisabledByDefault(t *testing.T) {
	s := &Server{
		app:    fiber.New(),
		config: testConfig(),
	}
	// testConfig leaves DebugPprof false; requireAuth stays nil because the
	// flag gate must prevent the routes from being registered at all.
	s.registerDebugPprofRoutes()

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
	resp, err := s.app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode, "pprof routes must not exist when server.debug_pprof is off")
}

func TestDebugPprofRoutes_EnabledRequiresAdminAuth(t *testing.T) {
	cfg := testConfig()
	cfg.Server.DebugPprof = true
	s := &Server{
		app:         fiber.New(),
		config:      cfg,
		requireAuth: stubUnauthorizedAuth,
	}
	s.registerDebugPprofRoutes()

	for _, path := range []string{
		"/debug/pprof",
		"/debug/pprof/",
		"/debug/pprof/heap?debug=1",
		"/debug/pprof/goroutine",
		"/debug/pprof/cmdline",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		resp, err := s.app.Test(req)
		require.NoError(t, err, path)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode, "unauthenticated %s must be rejected", path)
	}
}
