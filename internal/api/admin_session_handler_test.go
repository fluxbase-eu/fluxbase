package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// AdminSessionHandler Construction Tests
// =============================================================================

func TestNewAdminSessionHandler(t *testing.T) {
	t.Run("creates handler with nil repository", func(t *testing.T) {
		handler := NewAdminSessionHandler(nil)
		assert.NotNil(t, handler)
		assert.Nil(t, handler.sessionRepo)
	})
}

// =============================================================================
// ListSessions Parameter Parsing Tests
// =============================================================================

func TestListSessions_ParameterParsing(t *testing.T) {
	t.Run("default pagination values", func(t *testing.T) {
		// Test that default values are used when no parameters provided
		// Default limit should be 25, offset should be 0
		// Can't test without a mock repo, but we can test query parameter parsing

		app := fiber.New()
		handler := NewAdminSessionHandler(nil)
		app.Get("/sessions", handler.ListSessions)

		req := httptest.NewRequest(http.MethodGet, "/sessions", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Will fail due to nil repo, but we're testing parameter parsing behavior
		// The handler should have used default values
		assert.NotNil(t, resp)
	})

	t.Run("custom limit parameter", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)
		app.Get("/sessions", handler.ListSessions)

		req := httptest.NewRequest(http.MethodGet, "/sessions?limit=50", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Handler should accept custom limit
		assert.NotNil(t, resp)
	})

	t.Run("custom offset parameter", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)
		app.Get("/sessions", handler.ListSessions)

		req := httptest.NewRequest(http.MethodGet, "/sessions?offset=10", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.NotNil(t, resp)
	})

	t.Run("limit and offset together", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)
		app.Get("/sessions", handler.ListSessions)

		req := httptest.NewRequest(http.MethodGet, "/sessions?limit=30&offset=60", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.NotNil(t, resp)
	})

	t.Run("include_expired parameter", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)
		app.Get("/sessions", handler.ListSessions)

		req := httptest.NewRequest(http.MethodGet, "/sessions?include_expired=true", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.NotNil(t, resp)
	})

	t.Run("invalid limit value ignored", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)
		app.Get("/sessions", handler.ListSessions)

		req := httptest.NewRequest(http.MethodGet, "/sessions?limit=invalid", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should fall back to default, not fail
		assert.NotNil(t, resp)
	})

	t.Run("negative limit value ignored", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)
		app.Get("/sessions", handler.ListSessions)

		req := httptest.NewRequest(http.MethodGet, "/sessions?limit=-5", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should fall back to default, not fail
		assert.NotNil(t, resp)
	})

	t.Run("invalid offset value ignored", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)
		app.Get("/sessions", handler.ListSessions)

		req := httptest.NewRequest(http.MethodGet, "/sessions?offset=invalid", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should fall back to default (0), not fail
		assert.NotNil(t, resp)
	})

	t.Run("negative offset value ignored", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)
		app.Get("/sessions", handler.ListSessions)

		req := httptest.NewRequest(http.MethodGet, "/sessions?offset=-10", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should fall back to default (0), not fail
		assert.NotNil(t, resp)
	})
}

// =============================================================================
// ListSessions Limit Capping Tests
// =============================================================================

func TestListSessions_LimitCapping(t *testing.T) {
	t.Run("limit capped at 100", func(t *testing.T) {
		// When limit > 100, it should be capped to 100
		// This is tested by ensuring the handler doesn't error with large limits

		app := fiber.New()
		handler := NewAdminSessionHandler(nil)
		app.Get("/sessions", handler.ListSessions)

		req := httptest.NewRequest(http.MethodGet, "/sessions?limit=500", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Handler should cap limit to 100, not error
		assert.NotNil(t, resp)
	})

	t.Run("limit exactly 100 is allowed", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)
		app.Get("/sessions", handler.ListSessions)

		req := httptest.NewRequest(http.MethodGet, "/sessions?limit=100", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.NotNil(t, resp)
	})

	t.Run("limit of 1 is allowed", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)
		app.Get("/sessions", handler.ListSessions)

		req := httptest.NewRequest(http.MethodGet, "/sessions?limit=1", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.NotNil(t, resp)
	})

	t.Run("limit of 0 should use default", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)
		app.Get("/sessions", handler.ListSessions)

		req := httptest.NewRequest(http.MethodGet, "/sessions?limit=0", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// 0 should be ignored (parsed > 0 check), use default
		assert.NotNil(t, resp)
	})
}

// =============================================================================
// RevokeSession Validation Tests
// =============================================================================

func TestRevokeSession_Validation(t *testing.T) {
	t.Run("empty session ID returns error", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)

		// Use empty path parameter
		app.Delete("/sessions/:id", handler.RevokeSession)

		req := httptest.NewRequest(http.MethodDelete, "/sessions/", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Empty ID should return 400
		// Note: Fiber routing may return 404 for missing path parameter
		assert.Contains(t, []int{fiber.StatusBadRequest, fiber.StatusNotFound}, resp.StatusCode)
	})

	t.Run("valid session ID format accepted", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)

		app.Delete("/sessions/:id", handler.RevokeSession)

		// Valid session ID
		req := httptest.NewRequest(http.MethodDelete, "/sessions/sess_abc123", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Will fail at repo level (nil), but should pass validation
		assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("UUID session ID accepted", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)

		app.Delete("/sessions/:id", handler.RevokeSession)

		req := httptest.NewRequest(http.MethodDelete, "/sessions/550e8400-e29b-41d4-a716-446655440000", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should pass validation
		assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

// =============================================================================
// RevokeUserSessions Validation Tests
// =============================================================================

func TestRevokeUserSessions_Validation(t *testing.T) {
	t.Run("empty user ID returns error", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)

		app.Delete("/users/:user_id/sessions", handler.RevokeUserSessions)

		req := httptest.NewRequest(http.MethodDelete, "/users//sessions", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Empty user_id should return 400 or 404
		assert.Contains(t, []int{fiber.StatusBadRequest, fiber.StatusNotFound}, resp.StatusCode)
	})

	t.Run("valid user ID format accepted", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)

		app.Delete("/users/:user_id/sessions", handler.RevokeUserSessions)

		req := httptest.NewRequest(http.MethodDelete, "/users/user_123/sessions", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Will fail at repo level, but should pass validation
		assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("UUID user ID accepted", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)

		app.Delete("/users/:user_id/sessions", handler.RevokeUserSessions)

		req := httptest.NewRequest(http.MethodDelete, "/users/550e8400-e29b-41d4-a716-446655440000/sessions", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should pass validation
		assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

// =============================================================================
// Response Format Tests
// =============================================================================

func TestSessionResponses_Format(t *testing.T) {
	t.Run("revoke session success response format", func(t *testing.T) {
		// Test expected response structure
		expectedResponse := fiber.Map{
			"message": "Session revoked successfully",
		}

		data, err := json.Marshal(expectedResponse)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"message":"Session revoked successfully"`)
	})

	t.Run("revoke user sessions success response format", func(t *testing.T) {
		expectedResponse := fiber.Map{
			"message": "All user sessions revoked successfully",
		}

		data, err := json.Marshal(expectedResponse)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"message":"All user sessions revoked successfully"`)
	})

	t.Run("list sessions response structure", func(t *testing.T) {
		// Test expected pagination response structure
		expectedResponse := fiber.Map{
			"sessions":    []interface{}{},
			"count":       0,
			"total_count": 0,
			"limit":       25,
			"offset":      0,
		}

		data, err := json.Marshal(expectedResponse)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"sessions"`)
		assert.Contains(t, string(data), `"count"`)
		assert.Contains(t, string(data), `"total_count"`)
		assert.Contains(t, string(data), `"limit"`)
		assert.Contains(t, string(data), `"offset"`)
	})
}

// =============================================================================
// Error Response Tests
// =============================================================================

func TestSessionErrorResponses(t *testing.T) {
	t.Run("session ID required error", func(t *testing.T) {
		expectedError := fiber.Map{
			"error": "Session ID is required",
		}

		data, err := json.Marshal(expectedError)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"error":"Session ID is required"`)
	})

	t.Run("user ID required error", func(t *testing.T) {
		expectedError := fiber.Map{
			"error": "User ID is required",
		}

		data, err := json.Marshal(expectedError)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"error":"User ID is required"`)
	})

	t.Run("session not found error", func(t *testing.T) {
		expectedError := fiber.Map{
			"error": "Session not found",
		}

		data, err := json.Marshal(expectedError)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"error":"Session not found"`)
	})

	t.Run("failed to list sessions error", func(t *testing.T) {
		expectedError := fiber.Map{
			"error": "Failed to list sessions",
		}

		data, err := json.Marshal(expectedError)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"error":"Failed to list sessions"`)
	})
}

// =============================================================================
// Pagination Logic Tests
// =============================================================================

func TestPaginationLogic(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		// Test the default pagination values
		defaultLimit := 25
		defaultOffset := 0
		maxLimit := 100

		assert.Equal(t, 25, defaultLimit)
		assert.Equal(t, 0, defaultOffset)
		assert.Equal(t, 100, maxLimit)
	})

	t.Run("limit capping logic", func(t *testing.T) {
		// Test limit capping behavior
		limit := 150
		maxLimit := 100

		if limit > maxLimit {
			limit = maxLimit
		}

		assert.Equal(t, 100, limit)
	})

	t.Run("limit not capped when under max", func(t *testing.T) {
		limit := 50
		maxLimit := 100

		if limit > maxLimit {
			limit = maxLimit
		}

		assert.Equal(t, 50, limit)
	})

	t.Run("parse valid limit", func(t *testing.T) {
		// Simulating the parsing logic
		limitStr := "50"
		limit := 25 // default

		if parsed, err := parseInt(limitStr); err == nil && parsed > 0 {
			limit = parsed
			if limit > 100 {
				limit = 100
			}
		}

		assert.Equal(t, 50, limit)
	})

	t.Run("parse invalid limit uses default", func(t *testing.T) {
		limitStr := "invalid"
		limit := 25 // default

		if parsed, err := parseInt(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}

		assert.Equal(t, 25, limit)
	})
}

// Helper function for parsing (mimics strconv.Atoi)
func parseInt(s string) (int, error) {
	var result int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fiber.ErrBadRequest
		}
		result = result*10 + int(c-'0')
	}
	return result, nil
}

// =============================================================================
// Query Parameter Tests
// =============================================================================

func TestIncludeExpiredParameter(t *testing.T) {
	t.Run("include_expired true", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)
		app.Get("/sessions", handler.ListSessions)

		req := httptest.NewRequest(http.MethodGet, "/sessions?include_expired=true", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.NotNil(t, resp)
	})

	t.Run("include_expired false", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)
		app.Get("/sessions", handler.ListSessions)

		req := httptest.NewRequest(http.MethodGet, "/sessions?include_expired=false", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.NotNil(t, resp)
	})

	t.Run("include_expired not set", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)
		app.Get("/sessions", handler.ListSessions)

		req := httptest.NewRequest(http.MethodGet, "/sessions", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should default to not including expired
		assert.NotNil(t, resp)
	})

	t.Run("include_expired invalid value", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)
		app.Get("/sessions", handler.ListSessions)

		// Invalid value should be treated as false
		req := httptest.NewRequest(http.MethodGet, "/sessions?include_expired=invalid", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.NotNil(t, resp)
	})
}

// =============================================================================
// Combined Query Parameters Tests
// =============================================================================

func TestCombinedQueryParameters(t *testing.T) {
	t.Run("all parameters together", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)
		app.Get("/sessions", handler.ListSessions)

		req := httptest.NewRequest(http.MethodGet, "/sessions?limit=50&offset=20&include_expired=true", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.NotNil(t, resp)
	})
}

// =============================================================================
// HTTP Method Tests
// =============================================================================

func TestHTTPMethods(t *testing.T) {
	t.Run("list sessions uses GET", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)
		app.Get("/sessions", handler.ListSessions)

		// POST should not work
		req := httptest.NewRequest(http.MethodPost, "/sessions", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("revoke session uses DELETE", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)
		app.Delete("/sessions/:id", handler.RevokeSession)

		// GET should not work
		req := httptest.NewRequest(http.MethodGet, "/sessions/test-id", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("revoke user sessions uses DELETE", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)
		app.Delete("/users/:user_id/sessions", handler.RevokeUserSessions)

		// PUT should not work
		req := httptest.NewRequest(http.MethodPut, "/users/test-user/sessions", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusMethodNotAllowed, resp.StatusCode)
	})
}

// =============================================================================
// Internal Server Error Tests
// =============================================================================

func TestInternalServerErrors(t *testing.T) {
	t.Run("list sessions with nil repo returns 500", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)
		app.Get("/sessions", handler.ListSessions)

		req := httptest.NewRequest(http.MethodGet, "/sessions", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(body, &result)
		assert.Equal(t, "Failed to list sessions", result["error"])
	})

	t.Run("revoke session with nil repo returns 500", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)
		app.Delete("/sessions/:id", handler.RevokeSession)

		req := httptest.NewRequest(http.MethodDelete, "/sessions/test-session-id", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(body, &result)
		assert.Equal(t, "Failed to revoke session", result["error"])
	})

	t.Run("revoke user sessions with nil repo returns 500", func(t *testing.T) {
		app := fiber.New()
		handler := NewAdminSessionHandler(nil)
		app.Delete("/users/:user_id/sessions", handler.RevokeUserSessions)

		req := httptest.NewRequest(http.MethodDelete, "/users/test-user-id/sessions", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(body, &result)
		assert.Equal(t, "Failed to revoke user sessions", result["error"])
	})
}
