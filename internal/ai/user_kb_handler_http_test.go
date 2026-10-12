package ai

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/testutil"
)

// newKBHandlerTestApp builds a fiber app around the user KB update/delete
// handlers with the auth locals the middleware normally provides. userID is
// stamped per request via the X-Test-User header.
func visibilityPtr() *KBVisibility {
	v := KBVisibilityPrivate
	return &v
}

func newKBHandlerTestApp(handler *UserKnowledgeBaseHandler) *fiber.App {
	app := fiber.New()
	app.Patch("/kb/:id", func(c fiber.Ctx) error {
		c.Locals("user_id", c.Get("X-Test-User"))
		return handler.UpdateMyKnowledgeBase(c)
	})
	app.Delete("/kb/:id", func(c fiber.Ctx) error {
		c.Locals("user_id", c.Get("X-Test-User"))
		return handler.DeleteMyKnowledgeBase(c)
	})
	return app
}

func TestUserKBUpdateDeleteHandlers_HTTP(t *testing.T) {
	requireDB(t)
	ctx := aiTestCtx(t)

	kbStore := NewKnowledgeBaseStorage(aiTestDB)
	handler := NewUserKnowledgeBaseHandler(kbStore)
	app := newKBHandlerTestApp(handler)

	// Real user rows: knowledge_base_permissions FKs into auth.users.
	newUser := func(name string) string {
		id := testutil.CreateTestUser(t, aiTestDB.Pool(), name+"-"+uuid.New().String()[:8]+"@itest.example")
		t.Cleanup(func() {
			_, _ = aiTestDB.Pool().Exec(context.Background(), "DELETE FROM auth.users WHERE id = $1", id)
		})
		return id
	}
	owner := newUser("owner")
	editor := newUser("editor")
	stranger := newUser("stranger")

	ownerID := owner
	kb, err := kbStore.CreateKnowledgeBaseFromRequest(ctx, CreateKnowledgeBaseRequest{
		Name:       "itest-http-" + uuid.New().String()[:8],
		OwnerID:    &ownerID,
		Visibility: visibilityPtr(),
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = kbStore.DeleteKnowledgeBase(context.Background(), kb.ID) })

	// Grant editor permission to a second user.
	_, err = kbStore.GrantKBPermission(ctx, kb.ID, editor, string(KBPermissionEditor), &owner)
	require.NoError(t, err)

	do := func(method, userID string, body string) *http.Response {
		req := httptest.NewRequest(method, "/kb/"+kb.ID, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-User", userID)
		resp, err := app.Test(req)
		require.NoError(t, err)
		t.Cleanup(func() { _ = resp.Body.Close() })
		return resp
	}

	t.Run("owner can update", func(t *testing.T) {
		resp := do(http.MethodPatch, owner, `{"description": "updated by owner"}`)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("editor can update but not delete", func(t *testing.T) {
		// Re-create the KB (the owner subtest deleted it).
		kb2, err := kbStore.CreateKnowledgeBaseFromRequest(ctx, CreateKnowledgeBaseRequest{
			Name:       "itest-http-" + uuid.New().String()[:8],
			OwnerID:    &ownerID,
			Visibility: visibilityPtr(),
		})
		require.NoError(t, err)
		t.Cleanup(func() { _ = kbStore.DeleteKnowledgeBase(context.Background(), kb2.ID) })
		_, err = kbStore.GrantKBPermission(ctx, kb2.ID, editor, string(KBPermissionEditor), &owner)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPatch, "/kb/"+kb2.ID, bytes.NewBufferString(`{"description": "editor was here"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-User", editor)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode, "editor must be allowed to update")

		req = httptest.NewRequest(http.MethodDelete, "/kb/"+kb2.ID, nil)
		req.Header.Set("X-Test-User", editor)
		resp, err = app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode, "delete must be owner-only")
	})

	t.Run("stranger is rejected on both", func(t *testing.T) {
		resp := do(http.MethodPatch, stranger, `{"description": "sneaky"}`)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)

		resp = do(http.MethodDelete, stranger, "")
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("unknown KB returns 404 on delete", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/kb/"+uuid.NewString(), nil)
		req.Header.Set("X-Test-User", owner)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("owner can delete and the KB is gone", func(t *testing.T) {
		// Runs last: deletion removes the shared KB.
		resp := do(http.MethodDelete, owner, "")
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		gone, err := kbStore.GetKnowledgeBase(ctx, kb.ID)
		require.NoError(t, err)
		assert.Nil(t, gone, "deleted KB must be gone")
	})

	t.Run("invalid extraction strategy returns 400", func(t *testing.T) {
		kb3, err := kbStore.CreateKnowledgeBaseFromRequest(ctx, CreateKnowledgeBaseRequest{
			Name:       "itest-http-" + uuid.New().String()[:8],
			OwnerID:    &ownerID,
			Visibility: visibilityPtr(),
		})
		require.NoError(t, err)
		t.Cleanup(func() { _ = kbStore.DeleteKnowledgeBase(context.Background(), kb3.ID) })

		req := httptest.NewRequest(http.MethodPatch, "/kb/"+kb3.ID, bytes.NewBufferString(`{"entity_extraction_strategy": "magic"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-User", owner)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
