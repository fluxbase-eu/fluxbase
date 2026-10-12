package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// RequireAdmin Tests
// =============================================================================

func TestRequireAdmin_ServiceKey(t *testing.T) {
	app := fiber.New()

	app.Use(func(c fiber.Ctx) error {
		c.Locals("auth_type", "service_key")
		c.Locals("user_role", "service_role")
		return c.Next()
	})
	app.Use(RequireAdmin())
	app.Get("/admin", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRequireAdmin_ServiceRoleJWT(t *testing.T) {
	app := fiber.New()

	app.Use(func(c fiber.Ctx) error {
		c.Locals("auth_type", "service_role_jwt")
		c.Locals("user_role", "service_role")
		return c.Next()
	})
	app.Use(RequireAdmin())
	app.Get("/admin", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRequireAdmin_DashboardAdmin(t *testing.T) {
	app := fiber.New()

	app.Use(func(c fiber.Ctx) error {
		c.Locals("auth_type", "jwt")
		c.Locals("user_role", "instance_admin")
		return c.Next()
	})
	app.Use(RequireAdmin())
	app.Get("/admin", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRequireAdmin_RegularUser(t *testing.T) {
	app := fiber.New()

	app.Use(func(c fiber.Ctx) error {
		c.Locals("auth_type", "jwt")
		c.Locals("user_role", "authenticated")
		return c.Next()
	})
	app.Use(RequireAdmin())
	app.Get("/admin", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Admin role required")
}

func TestRequireAdmin_AnonUser(t *testing.T) {
	app := fiber.New()

	app.Use(func(c fiber.Ctx) error {
		c.Locals("auth_type", "service_role_jwt")
		c.Locals("user_role", "anon")
		return c.Next()
	})
	app.Use(RequireAdmin())
	app.Get("/admin", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestRequireAdmin_NoAuth(t *testing.T) {
	app := fiber.New()

	// No auth locals set
	app.Use(RequireAdmin())
	app.Get("/admin", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func BenchmarkRequireAdmin(b *testing.B) {
	app := fiber.New()

	app.Use(func(c fiber.Ctx) error {
		c.Locals("auth_type", "service_key")
		c.Locals("user_role", "service_role")
		return c.Next()
	})
	app.Use(RequireAdmin())
	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = app.Test(req)
	}
}
