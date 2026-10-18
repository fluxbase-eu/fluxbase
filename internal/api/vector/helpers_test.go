package vector

import (
	"errors"
	"testing"

	"github.com/gofiber/fiber/v3"
)

// newTestApp creates a Fiber app with the standard test error handler.
// Copied from internal/api's test helpers so this package's tests are
// self-contained (Go test helpers cannot be imported across packages).
func newTestApp(t *testing.T) *fiber.App {
	t.Helper()
	return fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})
}
