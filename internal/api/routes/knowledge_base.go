package routes

import (
	"github.com/gofiber/fiber/v3"
)

type KnowledgeBaseDeps struct {
	RequireAIEnabled fiber.Handler
	RequireAuth      fiber.Handler

	ListKBs          fiber.Handler
	CreateKB         fiber.Handler
	GetKB            fiber.Handler
	ShareKB          fiber.Handler
	ListPermissions  fiber.Handler
	RevokePermission fiber.Handler

	ListDocuments  fiber.Handler
	GetDocument    fiber.Handler
	AddDocument    fiber.Handler
	UploadDocument fiber.Handler
	DeleteDocument fiber.Handler
	SearchKB       fiber.Handler
}

func BuildKnowledgeBaseRoutes(deps *KnowledgeBaseDeps) *RouteGroup {
	if deps == nil {
		return nil
	}

	routes := []Route{
		{Method: "GET", Path: "/api/v1/ai/knowledge-bases", Handler: deps.ListKBs, Summary: "List user's knowledge bases", Auth: AuthRequired},
		{Method: "POST", Path: "/api/v1/ai/knowledge-bases", Handler: deps.CreateKB, Summary: "Create knowledge base", Auth: AuthRequired},
		{Method: "GET", Path: "/api/v1/ai/knowledge-bases/:id", Handler: deps.GetKB, Summary: "Get knowledge base", Auth: AuthRequired},
		{Method: "POST", Path: "/api/v1/ai/knowledge-bases/:id/share", Handler: deps.ShareKB, Summary: "Share knowledge base", Auth: AuthRequired},
		{Method: "GET", Path: "/api/v1/ai/knowledge-bases/:id/permissions", Handler: deps.ListPermissions, Summary: "List KB permissions", Auth: AuthRequired},
		{Method: "DELETE", Path: "/api/v1/ai/knowledge-bases/:id/permissions/:user_id", Handler: deps.RevokePermission, Summary: "Revoke KB permission", Auth: AuthRequired},
	}

	if deps.ListDocuments != nil {
		routes = append(routes,
			Route{Method: "GET", Path: "/api/v1/ai/knowledge-bases/:id/documents", Handler: deps.ListDocuments, Summary: "List KB documents", Auth: AuthRequired},
			Route{Method: "GET", Path: "/api/v1/ai/knowledge-bases/:id/documents/:doc_id", Handler: deps.GetDocument, Summary: "Get KB document", Auth: AuthRequired},
			Route{Method: "POST", Path: "/api/v1/ai/knowledge-bases/:id/documents", Handler: deps.AddDocument, Summary: "Add document to KB", Auth: AuthRequired},
			Route{Method: "POST", Path: "/api/v1/ai/knowledge-bases/:id/documents/upload", Handler: deps.UploadDocument, Summary: "Upload document to KB", Auth: AuthRequired},
			Route{Method: "DELETE", Path: "/api/v1/ai/knowledge-bases/:id/documents/:doc_id", Handler: deps.DeleteDocument, Summary: "Delete KB document", Auth: AuthRequired},
			Route{Method: "POST", Path: "/api/v1/ai/knowledge-bases/:id/search", Handler: deps.SearchKB, Summary: "Search knowledge base", Auth: AuthRequired},
		)
	}

	return &RouteGroup{
		Name:   "knowledge_base",
		Routes: routes,
		Middlewares: []Middleware{
			{Name: "RequireAIEnabled", Handler: deps.RequireAIEnabled},
		},
		AuthMiddlewares: &AuthMiddlewares{
			Required: deps.RequireAuth,
		},
	}
}
