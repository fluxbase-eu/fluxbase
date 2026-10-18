package api

import (
	"github.com/gofiber/fiber/v3"

	"github.com/nimbleflux/fluxbase/internal/api/apiutil"
)

// ParseBody binds the request body to req and returns a fiber.Error on failure.
func ParseBody(c fiber.Ctx, req any) error {
	return apiutil.ParseBody(c, req)
}

// NormalizePaginationParams validates and normalizes limit/offset pagination parameters.
// It enforces the maximum limit and ensures offset is non-negative.
// Returns the normalized (limit, offset) values.
func NormalizePaginationParams(limit, offset, defaultLimit, maxLimit int) (int, int) {
	if limit <= 0 || limit > maxLimit {
		limit = defaultLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
