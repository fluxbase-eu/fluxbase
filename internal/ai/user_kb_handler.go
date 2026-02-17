package ai

import (
	"github.com/gofiber/fiber/v3"
)

// UserKnowledgeBaseHandler handles user-facing KB endpoints
type UserKnowledgeBaseHandler struct {
	storage *KnowledgeBaseStorage
}

func NewUserKnowledgeBaseHandler(storage *KnowledgeBaseStorage) *UserKnowledgeBaseHandler {
	return &UserKnowledgeBaseHandler{
		storage: storage,
	}
}

// ListMyKnowledgeBases returns KBs accessible to current user
// GET /api/v1/ai/knowledge-bases
func (h *UserKnowledgeBaseHandler) ListMyKnowledgeBases(c fiber.Ctx) error {
	ctx := c.RequestCtx()
	userID := c.Locals("user_id").(string)

	kbs, err := h.storage.ListUserKnowledgeBases(ctx, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list knowledge bases",
		})
	}

	return c.JSON(fiber.Map{
		"knowledge_bases": kbs,
		"count":           len(kbs),
	})
}

// CreateMyKnowledgeBase creates a user-owned KB
// POST /api/v1/ai/knowledge-bases
func (h *UserKnowledgeBaseHandler) CreateMyKnowledgeBase(c fiber.Ctx) error {
	ctx := c.RequestCtx()
	userID := c.Locals("user_id").(string)

	var req CreateKnowledgeBaseRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Create KB with owner set to current user
	kb := &KnowledgeBase{
		Name:                req.Name,
		Namespace:           req.Namespace,
		Description:         req.Description,
		EmbeddingModel:      req.EmbeddingModel,
		EmbeddingDimensions: req.EmbeddingDimensions,
		ChunkSize:           req.ChunkSize,
		ChunkOverlap:        req.ChunkOverlap,
		ChunkStrategy:       req.ChunkStrategy,
		Enabled:             true,
		Source:              "api",
		OwnerID:             &userID,
		Visibility:          KBVisibilityPrivate,
	}

	if err := h.storage.CreateKnowledgeBase(ctx, kb); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create knowledge base",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(kb)
}

// GetMyKnowledgeBase returns a specific KB if user has access
// GET /api/v1/ai/knowledge-bases/:id
func (h *UserKnowledgeBaseHandler) GetMyKnowledgeBase(c fiber.Ctx) error {
	ctx := c.RequestCtx()
	userID := c.Locals("user_id").(string)
	kbID := c.Params("id")

	if !h.storage.CanUserAccessKB(ctx, kbID, userID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	kb, err := h.storage.GetKnowledgeBase(ctx, kbID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Knowledge base not found",
		})
	}

	return c.JSON(kb)
}

// ShareKnowledgeBase grants permission to another user
// POST /api/v1/ai/knowledge-bases/:id/share
func (h *UserKnowledgeBaseHandler) ShareKnowledgeBase(c fiber.Ctx) error {
	ctx := c.RequestCtx()
	userID := c.Locals("user_id").(string)
	kbID := c.Params("id")

	kb, err := h.storage.GetKnowledgeBase(ctx, kbID)
	if err != nil || kb.OwnerID == nil || *kb.OwnerID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only owner can share knowledge base",
		})
	}

	var req struct {
		UserID     string `json:"user_id"`
		Permission string `json:"permission"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	grant, err := h.storage.GrantKBPermission(ctx, kbID, req.UserID, req.Permission, &userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to grant permission",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(grant)
}

// ListPermissions lists permissions for a KB
// GET /api/v1/ai/knowledge-bases/:id/permissions
func (h *UserKnowledgeBaseHandler) ListPermissions(c fiber.Ctx) error {
	ctx := c.RequestCtx()
	userID := c.Locals("user_id").(string)
	kbID := c.Params("id")

	kb, err := h.storage.GetKnowledgeBase(ctx, kbID)
	if err != nil || kb.OwnerID == nil || *kb.OwnerID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only owner can view permissions",
		})
	}

	perms, err := h.storage.ListKBPermissions(ctx, kbID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list permissions",
		})
	}

	return c.JSON(perms)
}

// RevokePermission revokes a permission
// DELETE /api/v1/ai/knowledge-bases/:id/permissions/:user_id
func (h *UserKnowledgeBaseHandler) RevokePermission(c fiber.Ctx) error {
	ctx := c.RequestCtx()
	userID := c.Locals("user_id").(string)
	kbID := c.Params("id")
	targetUserID := c.Params("user_id")

	kb, err := h.storage.GetKnowledgeBase(ctx, kbID)
	if err != nil || kb.OwnerID == nil || *kb.OwnerID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only owner can revoke permissions",
		})
	}

	err = h.storage.RevokeKBPermission(ctx, kbID, targetUserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to revoke permission",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// RegisterUserKnowledgeBaseRoutes registers user-facing routes
func RegisterUserKnowledgeBaseRoutes(router fiber.Router, storage *KnowledgeBaseStorage) {
	handler := NewUserKnowledgeBaseHandler(storage)
	router.Get("/knowledge-bases", handler.ListMyKnowledgeBases)
	router.Post("/knowledge-bases", handler.CreateMyKnowledgeBase)
	router.Get("/knowledge-bases/:id", handler.GetMyKnowledgeBase)
	router.Post("/knowledge-bases/:id/share", handler.ShareKnowledgeBase)
	router.Get("/knowledge-bases/:id/permissions", handler.ListPermissions)
	router.Delete("/knowledge-bases/:id/permissions/:user_id", handler.RevokePermission)
}
