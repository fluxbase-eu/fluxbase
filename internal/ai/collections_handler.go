package ai

import (
	"github.com/gofiber/fiber/v3"
)

type CollectionHandler struct {
	storage *CollectionStorage
}

func NewCollectionHandler(storage *CollectionStorage) *CollectionHandler {
	return &CollectionHandler{storage: storage}
}

// RegisterRoutes registers collection routes
func (h *CollectionHandler) RegisterRoutes(router fiber.Router) {
	// Collection CRUD
	router.Get("/collections", h.ListCollections)
	router.Post("/collections", h.CreateCollection)
	router.Get("/collections/:id", h.GetCollection)
	router.Put("/collections/:id", h.UpdateCollection)
	router.Delete("/collections/:id", h.DeleteCollection)

	// Collection member management
	router.Get("/collections/:id/members", h.ListCollectionMembers)
	router.Post("/collections/:id/members", h.AddCollectionMember)
	router.Put("/collections/:id/members/:user_id", h.UpdateCollectionMemberRole)
	router.Delete("/collections/:id/members/:user_id", h.RemoveCollectionMember)

	// Collection KB management
	router.Get("/collections/:id/knowledge-bases", h.ListCollectionKBs)
	router.Post("/collections/:id/knowledge-bases/link", h.LinkKB)
	router.Delete("/collections/:id/knowledge-bases/:kb_id", h.UnlinkKB)
}

// ListCollections returns user's collections
// GET /api/v1/ai/collections
func (h *CollectionHandler) ListCollections(c fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	collections, err := h.storage.ListCollections(c.RequestCtx(), userID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"collections": collections,
		"count":       len(collections),
	})
}

// CreateCollection creates new collection
// POST /api/v1/ai/collections
func (h *CollectionHandler) CreateCollection(c fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req CreateCollectionRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	collection, err := h.storage.CreateCollection(c.RequestCtx(), userID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(collection)
}

// GetCollection retrieves collection details
// GET /api/v1/ai/collections/:id
func (h *CollectionHandler) GetCollection(c fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	collectionID := c.Params("id")

	collection, err := h.storage.GetCollection(c.RequestCtx(), collectionID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Collection not found"})
	}

	return c.JSON(collection)
}

// UpdateCollection updates collection
// PUT /api/v1/ai/collections/:id
func (h *CollectionHandler) UpdateCollection(c fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	collectionID := c.Params("id")

	var req UpdateCollectionRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	collection, err := h.storage.UpdateCollection(c.RequestCtx(), collectionID, userID, req)
	if err != nil {
		return err
	}

	return c.JSON(collection)
}

// DeleteCollection deletes collection
// DELETE /api/v1/ai/collections/:id
func (h *CollectionHandler) DeleteCollection(c fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	collectionID := c.Params("id")

	if err := h.storage.DeleteCollection(c.RequestCtx(), collectionID, userID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListCollectionKBs returns KBs in collection
// GET /api/v1/ai/collections/:id/knowledge-bases
func (h *CollectionHandler) ListCollectionKBs(c fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	collectionID := c.Params("id")

	kbs, err := h.storage.ListCollectionKnowledgeBases(c.RequestCtx(), collectionID, userID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"knowledge_bases": kbs,
		"count":           len(kbs),
	})
}

// LinkKB links KB to collection
// POST /api/v1/ai/collections/:id/knowledge-bases/link
func (h *CollectionHandler) LinkKB(c fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	collectionID := c.Params("id")

	var req LinkKBToCollectionRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.storage.LinkKnowledgeBaseToCollection(c.RequestCtx(), collectionID, req.KnowledgeBaseID, userID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusCreated)
}

// UnlinkKB unlinks KB from collection
// DELETE /api/v1/ai/collections/:id/knowledge-bases/:kb_id
func (h *CollectionHandler) UnlinkKB(c fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	collectionID := c.Params("id")
	kbID := c.Params("kb_id")

	if err := h.storage.UnlinkKnowledgeBaseFromCollection(c.RequestCtx(), collectionID, kbID, userID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ============================================================================
// Collection Member Management Handlers
// ============================================================================

// ListCollectionMembers returns all members of a collection
// GET /api/v1/ai/collections/:id/members
func (h *CollectionHandler) ListCollectionMembers(c fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	collectionID := c.Params("id")

	members, err := h.storage.ListCollectionMembers(c.RequestCtx(), collectionID, userID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"members": members,
		"count":   len(members),
	})
}

// AddCollectionMemberRequest represents a request to add a member
type AddCollectionMemberRequest struct {
	UserID string `json:"user_id" validate:"required"`
	Role   string `json:"role" validate:"required,oneof=viewer editor owner"`
}

// AddCollectionMember adds a user to a collection
// POST /api/v1/ai/collections/:id/members
func (h *CollectionHandler) AddCollectionMember(c fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	collectionID := c.Params("id")

	var req AddCollectionMemberRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.storage.AddCollectionMember(c.RequestCtx(), collectionID, req.UserID, req.Role, userID); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(fiber.StatusCreated)
}

// UpdateCollectionMemberRoleRequest represents a request to update a member's role
type UpdateCollectionMemberRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=viewer editor owner"`
}

// UpdateCollectionMemberRole updates a member's role
// PUT /api/v1/ai/collections/:id/members/:user_id
func (h *CollectionHandler) UpdateCollectionMemberRole(c fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	collectionID := c.Params("id")
	targetUserID := c.Params("user_id")

	var req UpdateCollectionMemberRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.storage.UpdateCollectionMemberRole(c.RequestCtx(), collectionID, targetUserID, req.Role, userID); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(fiber.StatusOK)
}

// RemoveCollectionMember removes a member from a collection
// DELETE /api/v1/ai/collections/:id/members/:user_id
func (h *CollectionHandler) RemoveCollectionMember(c fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	collectionID := c.Params("id")
	targetUserID := c.Params("user_id")

	// Check if requester has owner role
	canManage, err := h.storage.CanUserManageCollection(c.RequestCtx(), collectionID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to check permissions"})
	}
	if !canManage {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Only owners can remove members"})
	}

	if err := h.storage.RemoveCollectionMember(c.RequestCtx(), collectionID, targetUserID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
