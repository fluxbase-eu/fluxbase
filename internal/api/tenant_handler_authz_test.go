package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/middleware"
	"github.com/nimbleflux/fluxbase/internal/tenantdb"
)

// =============================================================================
// SEC-01 (#369) regression tests: admin handlers acting on the :id path
// parameter must deny principals that are not assigned to that tenant.
//
// These tests use a fake tenantdb.DB so the guarded handlers run for real:
// the fake serves the membership lookup (platform.tenant_admin_assignments)
// and the tenant row (platform.tenants) queries.
// =============================================================================

// fakeTenantRow is a pgx.Row backed by a scan closure.
type fakeTenantRow struct {
	scan func(dest ...any) error
}

func (r fakeTenantRow) Scan(dest ...any) error { return r.scan(dest...) }

// fakeTenantDB implements tenantdb.DB for the queries exercised by the
// tenant admin handlers (membership check, GetTenant, UpdateTenant).
type fakeTenantDB struct {
	assignments map[string]bool // "userID:tenantID" -> assigned
	tenants     map[string]*tenantdb.Tenant
	execErr     error // returned by Exec (UpdateTenant)
	updated     bool  // set when an UPDATE platform.tenants is executed
}

func (f *fakeTenantDB) QueryRow(_ context.Context, sql string, args ...any) pgx.Row {
	// IsUserAssignedToTenant: SELECT EXISTS(... tenant_admin_assignments ...)
	if strings.Contains(sql, "platform.tenant_admin_assignments") {
		tenantID, _ := args[0].(string)
		userID, _ := args[1].(string)
		exists := f.assignments[userID+":"+tenantID]
		return fakeTenantRow{scan: func(dest ...any) error {
			*dest[0].(*bool) = exists
			return nil
		}}
	}
	// GetTenant: SELECT ... FROM platform.tenants WHERE id = $1 ...
	if strings.Contains(sql, "FROM platform.tenants") {
		id, _ := args[0].(string)
		t, ok := f.tenants[id]
		return fakeTenantRow{scan: func(dest ...any) error {
			if !ok || t == nil {
				return tenantdb.ErrTenantNotFound
			}
			*dest[0].(*string) = t.ID
			*dest[1].(*string) = t.Slug
			*dest[2].(*string) = t.Name
			*dest[3].(**string) = t.DBName
			*dest[4].(*bool) = t.IsDefault
			*dest[5].(*tenantdb.TenantStatus) = t.Status
			*dest[6].(*[]byte) = []byte("{}")
			*dest[7].(*time.Time) = t.CreatedAt
			*dest[8].(*time.Time) = t.UpdatedAt
			*dest[9].(**time.Time) = t.DeletedAt
			return nil
		}}
	}
	return fakeTenantRow{scan: func(dest ...any) error {
		return fmt.Errorf("fakeTenantDB: unexpected query: %s", sql)
	}}
}

func (f *fakeTenantDB) Query(_ context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, fmt.Errorf("fakeTenantDB: unexpected query: %s (%v)", sql, args)
}

func (f *fakeTenantDB) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if f.execErr != nil {
		return pgconn.CommandTag{}, f.execErr
	}
	if strings.Contains(sql, "UPDATE platform.tenants") {
		f.updated = true
		// Apply the update so a subsequent GetTenant observes it.
		if len(args) == 2 {
			if id, ok := args[1].(string); ok {
				if t, found := f.tenants[id]; found {
					switch v := args[0].(type) {
					case string:
						t.Name = v
					case []byte:
						var m map[string]any
						if json.Unmarshal(v, &m) == nil {
							t.Metadata = m
						}
					}
					t.UpdatedAt = time.Now()
				}
			}
		}
		return pgconn.NewCommandTag("UPDATE 1"), nil
	}
	return pgconn.NewCommandTag("EXEC 1"), nil
}

// newAuthzTenantHandler builds a TenantHandler backed by the fake DB.
func newAuthzTenantHandler(db *fakeTenantDB) *TenantHandler {
	return NewTenantHandler(nil, nil, tenantdb.NewRepository(db), nil, nil, nil)
}

// authzLocals returns middleware that fakes an authenticated principal.
func authzLocals(userID string, role string, instanceAdmin bool) fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Locals("user_id", userID)
		c.Locals("user_role", role)
		if instanceAdmin {
			c.Locals("is_instance_admin", true)
		}
		return c.Next()
	}
}

// mustReadErrorResponse decodes a JSON error response body.
func mustReadErrorResponse(t *testing.T, resp *http.Response) ErrorResponse {
	t.Helper()
	var e ErrorResponse
	err := json.NewDecoder(resp.Body).Decode(&e)
	require.NoError(t, err)
	return e
}

// =============================================================================
// RequireTenantParamAccess unit tests
// =============================================================================

func TestTenantParamAccess(t *testing.T) {
	userA := "11111111-1111-1111-1111-111111111111"
	tenantA := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	tenantB := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

	db := &fakeTenantDB{
		assignments: map[string]bool{userA + ":" + tenantA: true},
		tenants: map[string]*tenantdb.Tenant{
			tenantA: {ID: tenantA, Slug: "tenant-a", Name: "Tenant A", Status: tenantdb.TenantStatusActive},
			tenantB: {ID: tenantB, Slug: "tenant-b", Name: "Tenant B", Status: tenantdb.TenantStatusActive},
		},
	}
	checker := tenantdb.NewRepository(db)

	tests := []struct {
		name     string
		locals   func(c fiber.Ctx)
		tenantID string
		expected bool
	}{
		{
			name:     "assigned tenant admin is allowed",
			locals:   func(c fiber.Ctx) { c.Locals("user_id", userA); c.Locals("user_role", "tenant_admin") },
			tenantID: tenantA,
			expected: true,
		},
		{
			name:     "unassigned tenant admin is denied",
			locals:   func(c fiber.Ctx) { c.Locals("user_id", userA); c.Locals("user_role", "tenant_admin") },
			tenantID: tenantB,
			expected: false,
		},
		{
			name:     "instance admin local bypasses membership",
			locals:   func(c fiber.Ctx) { c.Locals("user_id", userA); c.Locals("is_instance_admin", true) },
			tenantID: tenantB,
			expected: true,
		},
		{
			name:     "dashboard instance_admin role bypasses membership",
			locals:   func(c fiber.Ctx) { c.Locals("user_id", userA); c.Locals("user_role", "instance_admin") },
			tenantID: tenantB,
			expected: true,
		},
		{
			name:     "missing user id is denied",
			locals:   func(c fiber.Ctx) {},
			tenantID: tenantA,
			expected: false,
		},
		{
			name:     "empty tenant id is denied",
			locals:   func(c fiber.Ctx) { c.Locals("user_id", userA) },
			tenantID: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestApp(t)
			var allowed bool
			app.Use(func(c fiber.Ctx) error {
				tt.locals(c)
				allowed = middleware.RequireTenantParamAccess(c, checker, tt.tenantID)
				return c.Next()
			})
			app.Get("/probe", func(c fiber.Ctx) error { return c.SendString("ok") })

			req := httptest.NewRequest(http.MethodGet, "/probe", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, tt.expected, allowed)
		})
	}

	t.Run("nil checker is denied", func(t *testing.T) {
		app := newTestApp(t)
		var allowed bool
		app.Use(func(c fiber.Ctx) error {
			c.Locals("user_id", userA)
			allowed = middleware.RequireTenantParamAccess(c, nil, tenantA)
			return c.Next()
		})
		app.Get("/probe", func(c fiber.Ctx) error { return c.SendString("ok") })

		req := httptest.NewRequest(http.MethodGet, "/probe", nil)
		_, err := app.Test(req)
		require.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("database failure denies access", func(t *testing.T) {
		app := newTestApp(t)
		var allowed bool
		app.Use(func(c fiber.Ctx) error {
			c.Locals("user_id", userA)
			allowed = middleware.RequireTenantParamAccess(c, tenantdb.NewRepository(brokenFakeTenantDB{}), tenantA)
			return c.Next()
		})
		app.Get("/probe", func(c fiber.Ctx) error { return c.SendString("ok") })

		req := httptest.NewRequest(http.MethodGet, "/probe", nil)
		_, err := app.Test(req)
		require.NoError(t, err)
		assert.False(t, allowed)
	})
}

// brokenFakeTenantDB always fails queries.
type brokenFakeTenantDB struct{}

func (brokenFakeTenantDB) QueryRow(context.Context, string, ...any) pgx.Row {
	return fakeTenantRow{scan: func(dest ...any) error { return errors.New("db down") }}
}

func (brokenFakeTenantDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("db down")
}

func (brokenFakeTenantDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("db down")
}

// =============================================================================
// UpdateTenant handler tests (full success path runs on the fake repository)
// =============================================================================

func TestUpdateTenant_CrossTenantForbidden(t *testing.T) {
	userA := "11111111-1111-1111-1111-111111111111"
	tenantA := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	tenantB := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

	newDB := func() *fakeTenantDB {
		return &fakeTenantDB{
			assignments: map[string]bool{userA + ":" + tenantA: true},
			tenants: map[string]*tenantdb.Tenant{
				tenantA: {ID: tenantA, Slug: "tenant-a", Name: "Tenant A", Status: tenantdb.TenantStatusActive},
				tenantB: {ID: tenantB, Slug: "tenant-b", Name: "Tenant B", Status: tenantdb.TenantStatusActive},
			},
		}
	}

	t.Run("tenant admin of tenant A is denied on tenant B", func(t *testing.T) {
		app := newTestApp(t)
		db := newDB()
		handler := newAuthzTenantHandler(db)
		app.Use(authzLocals(userA, "tenant_admin", false))
		app.Patch("/tenants/:id", handler.UpdateTenant)

		req := httptest.NewRequest(http.MethodPatch, "/tenants/"+tenantB, strings.NewReader(`{"name":"Hijacked"}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
		e := mustReadErrorResponse(t, resp)
		assert.Equal(t, ErrCodeAccessDenied, e.Code)
		assert.False(t, db.updated, "tenant B must not be modified")
	})

	t.Run("tenant admin of tenant A succeeds on tenant A", func(t *testing.T) {
		app := newTestApp(t)
		db := newDB()
		handler := newAuthzTenantHandler(db)
		app.Use(authzLocals(userA, "tenant_admin", false))
		app.Patch("/tenants/:id", handler.UpdateTenant)

		req := httptest.NewRequest(http.MethodPatch, "/tenants/"+tenantA, strings.NewReader(`{"name":"Renamed"}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.True(t, db.updated, "tenant A should be modified")

		var tenant TenantResponse
		err = json.NewDecoder(resp.Body).Decode(&tenant)
		require.NoError(t, err)
		assert.Equal(t, tenantA, tenant.ID)
		assert.Equal(t, "Renamed", tenant.Name)
	})

	t.Run("instance admin succeeds on tenant B", func(t *testing.T) {
		app := newTestApp(t)
		db := newDB()
		handler := newAuthzTenantHandler(db)
		app.Use(authzLocals(userA, "instance_admin", true))
		app.Patch("/tenants/:id", handler.UpdateTenant)

		req := httptest.NewRequest(http.MethodPatch, "/tenants/"+tenantB, strings.NewReader(`{"name":"Renamed B"}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.True(t, db.updated, "tenant B should be modified by an instance admin")
	})
}

// =============================================================================
// AssignAdmin / RemoveAdmin handler tests (denial path; the success path is
// covered by TestTenantParamAccess and TestUpdateTenant_CrossTenantForbidden
// because AssignAdmin/RemoveAdmin need a live main pool beyond the guard)
// =============================================================================

func TestAssignAdmin_CrossTenantForbidden(t *testing.T) {
	userA := "11111111-1111-1111-1111-111111111111"
	tenantB := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

	app := newTestApp(t)
	handler := newAuthzTenantHandler(&fakeTenantDB{
		assignments: map[string]bool{},
		tenants:     map[string]*tenantdb.Tenant{},
	})
	app.Use(authzLocals(userA, "tenant_admin", false))
	app.Post("/tenants/:id/admins", handler.AssignAdmin)

	req := httptest.NewRequest(http.MethodPost, "/tenants/"+tenantB+"/admins",
		strings.NewReader(`{"user_id":"`+userA+`"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	e := mustReadErrorResponse(t, resp)
	assert.Equal(t, ErrCodeAccessDenied, e.Code)
}

func TestRemoveAdmin_CrossTenantForbidden(t *testing.T) {
	userA := "11111111-1111-1111-1111-111111111111"
	tenantB := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

	app := newTestApp(t)
	handler := newAuthzTenantHandler(&fakeTenantDB{
		assignments: map[string]bool{},
		tenants:     map[string]*tenantdb.Tenant{},
	})
	app.Use(authzLocals(userA, "tenant_admin", false))
	app.Delete("/tenants/:id/admins/:user_id", handler.RemoveAdmin)

	req := httptest.NewRequest(http.MethodDelete, "/tenants/"+tenantB+"/admins/"+userA, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	e := mustReadErrorResponse(t, resp)
	assert.Equal(t, ErrCodeAccessDenied, e.Code)
}

// =============================================================================
// Tenant schema handler tests
// =============================================================================

func TestTenantSchemaHandlers_CrossTenantForbidden(t *testing.T) {
	userA := "11111111-1111-1111-1111-111111111111"
	tenantB := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

	tests := []struct {
		method  string
		path    string
		handler func(h *TenantHandler) fiber.Handler
	}{
		{http.MethodGet, "/tenants/:id/schema", func(h *TenantHandler) fiber.Handler { return h.GetTenantSchemaStatus }},
		{http.MethodPost, "/tenants/:id/schema/apply", func(h *TenantHandler) fiber.Handler { return h.ApplyTenantSchema }},
		{http.MethodGet, "/tenants/:id/schema/content", func(h *TenantHandler) fiber.Handler { return h.GetStoredSchema }},
		{http.MethodPost, "/tenants/:id/schema/content", func(h *TenantHandler) fiber.Handler { return h.UploadTenantSchema }},
		{http.MethodPost, "/tenants/:id/schema/content/apply", func(h *TenantHandler) fiber.Handler { return h.ApplyUploadedTenantSchema }},
		{http.MethodDelete, "/tenants/:id/schema/content", func(h *TenantHandler) fiber.Handler { return h.DeleteStoredSchema }},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			app := newTestApp(t)
			handler := newAuthzTenantHandler(&fakeTenantDB{
				assignments: map[string]bool{},
				tenants:     map[string]*tenantdb.Tenant{},
			})
			app.Use(authzLocals(userA, "tenant_admin", false))
			app.Add([]string{tt.method}, tt.path, tt.handler(handler))

			url := strings.ReplaceAll(strings.ReplaceAll(tt.path, ":id", tenantB), "/", "/")
			req := httptest.NewRequest(tt.method, url, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
			e := mustReadErrorResponse(t, resp)
			assert.Equal(t, ErrCodeAccessDenied, e.Code)
		})
	}
}

// =============================================================================
// Tenant settings handler tests
// =============================================================================

func TestTenantSettingsHandlers_CrossTenantForbidden(t *testing.T) {
	userA := "11111111-1111-1111-1111-111111111111"
	tenantB := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

	repo := tenantdb.NewRepository(&fakeTenantDB{
		assignments: map[string]bool{},
		tenants:     map[string]*tenantdb.Tenant{},
	})
	settingsHandler := NewTenantSettingsHandler(nil, repo)

	tests := []struct {
		method  string
		path    string
		handler fiber.Handler
	}{
		{http.MethodGet, "/tenants/:id/settings", settingsHandler.GetTenantSettings},
		{http.MethodPatch, "/tenants/:id/settings", settingsHandler.UpdateTenantSettings},
		{http.MethodGet, "/tenants/:id/settings/*", settingsHandler.GetTenantSetting},
		{http.MethodDelete, "/tenants/:id/settings/*", settingsHandler.DeleteTenantSetting},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			app := newTestApp(t)
			app.Use(authzLocals(userA, "tenant_admin", false))
			app.Add([]string{tt.method}, tt.path, tt.handler)

			url := strings.ReplaceAll(tt.path, ":id", tenantB)
			req := httptest.NewRequest(tt.method, url, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
			e := mustReadErrorResponse(t, resp)
			assert.Equal(t, ErrCodeAccessDenied, e.Code)
		})
	}
}
