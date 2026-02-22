package ai

import (
	"github.com/gofiber/fiber/v3"
)

// RequireKBPermission creates a middleware that checks if the user has the required
// permission level on the knowledge base specified in the URL parameter.
// The middleware expects:
// - "user_id" in Locals (set by auth middleware)
// - "id" or "kb_id" URL parameter for the knowledge base ID
//
// Permission levels (hierarchical): viewer < editor < owner
// - viewer: Read access to documents
// - editor: Read + write access (can add/edit/delete documents)
// - owner: Full control (can manage permissions, delete KB)
func RequireKBPermission(storage *KnowledgeBaseStorage, requiredPermission string) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.RequestCtx()

		// Get user ID from context (set by auth middleware)
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		// Get KB ID from URL params - try "id" first, then "kb_id"
		kbID := c.Params("id")
		if kbID == "" {
			kbID = c.Params("kb_id")
		}
		if kbID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Knowledge base ID is required",
			})
		}

		// Check if user has the required permission
		hasPermission, err := storage.CheckKBPermission(ctx, kbID, userID, requiredPermission)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to check permission",
			})
		}

		if !hasPermission {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Insufficient permissions",
			})
		}

		// Store the user's permission level in context for later use
		permission, _ := storage.GetUserKBPermission(ctx, kbID, userID)
		c.Locals("kb_permission", permission)

		return c.Next()
	}
}

// RequireKBViewer is a convenience middleware for read-only access
func RequireKBViewer(storage *KnowledgeBaseStorage) fiber.Handler {
	return RequireKBPermission(storage, string(KBPermissionViewer))
}

// RequireKBEditor is a convenience middleware for write access
func RequireKBEditor(storage *KnowledgeBaseStorage) fiber.Handler {
	return RequireKBPermission(storage, string(KBPermissionEditor))
}

// RequireKBOwner is a convenience middleware for owner-level access
func RequireKBOwner(storage *KnowledgeBaseStorage) fiber.Handler {
	return RequireKBPermission(storage, string(KBPermissionOwner))
}
