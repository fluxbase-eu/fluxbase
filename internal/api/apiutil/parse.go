// Package apiutil holds small HTTP helpers shared by the API handler
// packages. It exists so per-domain packages extracted from internal/api
// (vector, webhook, ...) can share request-parsing utilities without
// importing the god package.
package apiutil

import "github.com/gofiber/fiber/v3"

// ParseBody binds the request body to req and returns a fiber.Error on failure.
func ParseBody(c fiber.Ctx, req any) error {
	if err := c.Bind().Body(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	return nil
}
