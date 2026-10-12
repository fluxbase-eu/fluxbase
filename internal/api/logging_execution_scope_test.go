package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/config"
	"github.com/nimbleflux/fluxbase/internal/logging"
	"github.com/nimbleflux/fluxbase/internal/storage"
)

// =============================================================================
// ExecutionTenantResolver fake
// =============================================================================

// fakeExecutionTenantResolver implements ExecutionTenantResolver for tests.
type fakeExecutionTenantResolver struct {
	owners map[string]string // execution ID -> owning tenant
	err    error
	calls  int
}

func (f *fakeExecutionTenantResolver) ExecutionTenant(ctx context.Context, executionID string) (string, bool, error) {
	f.calls++
	if f.err != nil {
		return "", false, f.err
	}
	owner, ok := f.owners[executionID]
	if !ok {
		return "", false, nil
	}
	return owner, true, nil
}

// =============================================================================
// Helpers
// =============================================================================

// newExecutionLogTestService creates a logging service backed by local file storage.
func newExecutionLogTestService(t *testing.T) *logging.Service {
	t.Helper()

	svc, err := logging.New(&config.LoggingConfig{
		Backend:   "local",
		LocalPath: t.TempDir(),
	}, nil, nil, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })

	return svc
}

// seedExecutionLog writes one execution log entry and flushes it to storage.
func seedExecutionLog(t *testing.T, svc *logging.Service, executionID, message string) {
	t.Helper()

	svc.Log(context.Background(), &storage.LogEntry{
		Category:    storage.LogCategoryExecution,
		Level:       storage.LogLevelInfo,
		Message:     message,
		ExecutionID: executionID,
	})
	require.NoError(t, svc.Flush(context.Background()))
}

// callGetExecutionLogs runs the handler with the given middleware locals
// simulating the tenant/auth middleware, and returns the response.
func callGetExecutionLogs(t *testing.T, handler *LoggingHandler, executionID string, locals map[string]any) *http.Response {
	t.Helper()

	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		for k, v := range locals {
			c.Locals(k, v)
		}
		return c.Next()
	})
	app.Get("/admin/logs/executions/:id", handler.GetExecutionLogs)

	req := httptest.NewRequest(http.MethodGet, "/admin/logs/executions/"+executionID, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	return resp
}

// tenantLocals simulates a caller scoped to tenantID (e.g. via JWT claim).
func tenantLocals(tenantID string) map[string]any {
	return map[string]any{
		"tenant_id":     tenantID,
		"tenant_source": "jwt",
	}
}

// instanceAdminLocals simulates an instance admin on the default tenant
// (cross-tenant context, matching the ctxForLogs bypass condition).
func instanceAdminLocals() map[string]any {
	return map[string]any{
		"tenant_id":         uuid.New().String(),
		"tenant_source":     "default",
		"is_instance_admin": true,
	}
}

// =============================================================================
// Tenant-scoped execution log reads (regression tests for issue #403)
// =============================================================================

func TestLoggingHandler_GetExecutionLogs_TenantScope(t *testing.T) {
	tenantA := uuid.New().String()
	tenantB := uuid.New().String()
	execA := uuid.New().String()
	execB := uuid.New().String()

	t.Run("returns logs for an execution owned by the caller's tenant", func(t *testing.T) {
		svc := newExecutionLogTestService(t)
		seedExecutionLog(t, svc, execA, "tenant A execution log")

		handler := &LoggingHandler{
			loggingService: svc,
			executions: &fakeExecutionTenantResolver{owners: map[string]string{
				execA: tenantA,
			}},
		}

		resp := callGetExecutionLogs(t, handler, execA, tenantLocals(tenantA))
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var body struct {
			Entries []struct {
				Message string `json:"message"`
			} `json:"entries"`
			Count int `json:"count"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
		assert.Equal(t, 1, body.Count)
		require.Len(t, body.Entries, 1)
		assert.Equal(t, "tenant A execution log", body.Entries[0].Message)
	})

	t.Run("rejects an execution owned by another tenant with 404", func(t *testing.T) {
		svc := newExecutionLogTestService(t)
		seedExecutionLog(t, svc, execB, "tenant B execution log")

		resolver := &fakeExecutionTenantResolver{owners: map[string]string{
			execB: tenantB,
		}}
		handler := &LoggingHandler{loggingService: svc, executions: resolver}

		resp := callGetExecutionLogs(t, handler, execB, tenantLocals(tenantA))
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		assert.Equal(t, 1, resolver.calls, "ownership must be verified on every tenant-scoped read")
	})

	t.Run("rejects unknown executions with 404", func(t *testing.T) {
		svc := newExecutionLogTestService(t)
		handler := &LoggingHandler{
			loggingService: svc,
			executions:     &fakeExecutionTenantResolver{owners: map[string]string{}},
		}

		resp := callGetExecutionLogs(t, handler, uuid.New().String(), tenantLocals(tenantA))
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("lets instance admins on the default tenant read any execution", func(t *testing.T) {
		svc := newExecutionLogTestService(t)
		seedExecutionLog(t, svc, execB, "tenant B execution log")

		// Resolver knows nothing: the bypass must not consult it.
		handler := &LoggingHandler{
			loggingService: svc,
			executions:     &fakeExecutionTenantResolver{owners: map[string]string{}},
		}

		resp := callGetExecutionLogs(t, handler, execB, instanceAdminLocals())
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var body struct {
			Count int `json:"count"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
		assert.Equal(t, 1, body.Count)
	})

	t.Run("fails closed when the resolver is not configured", func(t *testing.T) {
		svc := newExecutionLogTestService(t)
		handler := &LoggingHandler{loggingService: svc}

		resp := callGetExecutionLogs(t, handler, execA, tenantLocals(tenantA))
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("rejects malformed execution IDs with 400", func(t *testing.T) {
		svc := newExecutionLogTestService(t)
		handler := &LoggingHandler{loggingService: svc}

		resp := callGetExecutionLogs(t, handler, "exec-123", tenantLocals(tenantA))
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("reports resolver failures as internal errors", func(t *testing.T) {
		svc := newExecutionLogTestService(t)
		handler := &LoggingHandler{
			loggingService: svc,
			executions:     &fakeExecutionTenantResolver{err: assert.AnError},
		}

		resp := callGetExecutionLogs(t, handler, execA, tenantLocals(tenantA))
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
