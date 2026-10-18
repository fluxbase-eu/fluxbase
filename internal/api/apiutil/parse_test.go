package apiutil

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBody(t *testing.T) {
	t.Run("binds a valid JSON body", func(t *testing.T) {
		app := fiber.New()
		var got struct {
			Name string `json:"name"`
		}
		app.Post("/", func(c fiber.Ctx) error {
			require.NoError(t, ParseBody(c, &got))
			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest(fiber.MethodPost, "/", bytes.NewBufferString(`{"name":"flux"}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.Equal(t, "flux", got.Name)
	})

	t.Run("returns a 400 fiber error for an invalid body", func(t *testing.T) {
		app := fiber.New()
		app.Post("/", func(c fiber.Ctx) error {
			var req struct {
				Name string `json:"name"`
			}
			return ParseBody(c, &req)
		})

		req := httptest.NewRequest(fiber.MethodPost, "/", bytes.NewBufferString(`{invalid json`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}
