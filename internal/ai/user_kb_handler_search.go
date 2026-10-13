package ai

import (
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/middleware"
)

// SearchMyKB searches a knowledge base (requires viewer permission)
// POST /api/v1/ai/knowledge-bases/:id/search
func (h *UserKnowledgeBaseHandler) SearchMyKB(c fiber.Ctx) error {
	ctx := middleware.CtxWithTenant(c)
	userID := middleware.GetUserID(c)
	kbID := c.Params("id")

	// Check read permission
	hasPermission, err := h.storage.CheckKBPermission(ctx, kbID, userID, string(KBPermissionViewer))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check permission",
		})
	}
	if !hasPermission {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	var req SearchRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Query is required",
		})
	}

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 10
	}

	// Perform search using hybrid search (keyword-only if embeddings not available)
	// Per-user isolation: viewers only see their own documents plus global
	// (user_id-less) content — same semantics as the chat RAG path.
	opts := HybridSearchOptions{
		Query:  req.Query,
		Limit:  req.Limit,
		Mode:   SearchModeKeyword, // Default to keyword search for user endpoint
		Filter: &MetadataFilter{UserID: &userID, IncludeGlobal: true},
	}

	// Request-level search options (aligned with the admin search endpoint).
	// "semantic"/"hybrid" require an embedding service; fall back to keyword
	// when unavailable.
	switch req.Mode {
	case "semantic", "hybrid":
		if h.processor != nil && h.processor.embeddingService != nil {
			if embedding, err := h.processor.embeddingService.EmbedSingle(ctx, req.Query, ""); err == nil && len(embedding) > 0 {
				opts.QueryEmbedding = embedding
				opts.Mode = SearchModeHybrid
				if req.Mode == "semantic" {
					opts.SemanticWeight = 1.0
				} else {
					opts.SemanticWeight = 0.7
				}
			}
		}
	}
	if req.SemanticWeight > 0 {
		opts.SemanticWeight = req.SemanticWeight
	}
	if req.Threshold > 0 {
		opts.Threshold = req.Threshold
	}

	// If processor has embedding service and no explicit keyword mode, use hybrid search
	if opts.Mode == SearchModeKeyword && req.Mode == "" && h.processor != nil && h.processor.embeddingService != nil {
		if embedding, err := h.processor.embeddingService.EmbedSingle(ctx, req.Query, ""); err == nil && len(embedding) > 0 {
			opts.QueryEmbedding = embedding
			opts.Mode = SearchModeHybrid
			opts.SemanticWeight = 0.7
		}
	}

	results, err := h.storage.SearchChunksHybrid(ctx, kbID, opts)
	if err != nil {
		log.Error().Err(err).Str("kb_id", kbID).Msg("Search failed")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Search failed",
		})
	}

	// Optional rerank: request flag overrides the KB's rerank_enabled opt-in.
	if h.reranker != nil && len(results) > 1 {
		shouldRerank := false
		if req.Rerank != nil {
			shouldRerank = *req.Rerank
		} else if kb, kbErr := h.storage.GetKnowledgeBase(ctx, kbID); kbErr == nil && kb != nil {
			shouldRerank = kb.RerankEnabled
		}
		if shouldRerank {
			results = rerankResults(ctx, h.reranker, req.Query, results, req.Limit, nil)
		}
	}

	return c.JSON(fiber.Map{
		"results": results,
		"query":   req.Query,
		"limit":   req.Limit,
		"count":   len(results),
	})
}

// DebugSearchMyKB performs a debug search with detailed diagnostic information
// POST /api/v1/ai/knowledge-bases/:id/debug-search
func (h *UserKnowledgeBaseHandler) DebugSearchMyKB(c fiber.Ctx) error {
	ctx := middleware.CtxWithTenant(c)
	userID := middleware.GetUserID(c)
	kbID := c.Params("id")

	// Check viewer permission
	hasPermission, err := h.storage.CheckKBPermission(ctx, kbID, userID, string(KBPermissionViewer))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check permission",
		})
	}
	if !hasPermission {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Viewer permission required",
		})
	}

	var req struct {
		Query string `json:"query"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Query is required",
		})
	}

	// Perform search with debug info. Per-user isolation matches SearchMyKB
	// (caller's documents + global). Defaults to keyword mode since no
	// embedding is required; upgrades to hybrid when one can be generated.
	opts := HybridSearchOptions{
		Query:          req.Query,
		Limit:          10,
		SemanticWeight: 0.7,
		Mode:           SearchModeKeyword,
		Filter:         &MetadataFilter{UserID: &userID, IncludeGlobal: true},
	}

	if h.processor != nil && h.processor.embeddingService != nil {
		if embedding, embErr := h.processor.embeddingService.EmbedSingle(ctx, req.Query, ""); embErr == nil && len(embedding) > 0 {
			opts.QueryEmbedding = embedding
			opts.Mode = SearchModeHybrid
		}
	}

	results, err := h.storage.SearchChunksHybrid(ctx, kbID, opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Search failed",
		})
	}

	// Get KB info for context (checked — unknown KB must 404, not nil-deref)
	kb, err := h.storage.GetKnowledgeBase(ctx, kbID)
	if err != nil || kb == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Knowledge base not found",
		})
	}

	return c.JSON(fiber.Map{
		"query":          req.Query,
		"results":        results,
		"result_count":   len(results),
		"search_options": opts,
		"knowledge_base": fiber.Map{
			"id":   kbID,
			"name": kb.Name,
		},
		"debug_info": fiber.Map{
			"search_type":      "hybrid",
			"semantic_weight":  opts.SemanticWeight,
			"keyword_weight":   1 - opts.SemanticWeight,
			"embedding_status": "available",
		},
	})
}

// SearchRequest represents a search request for the user KB search endpoint
type SearchRequest struct {
	Query          string  `json:"query"`
	Limit          int     `json:"limit,omitempty"`
	Mode           string  `json:"mode,omitempty"`            // "semantic", "keyword", or "hybrid" (default: hybrid when embeddings available, else keyword)
	SemanticWeight float64 `json:"semantic_weight,omitempty"` // For hybrid mode: 0-1, default 0.7
	Threshold      float64 `json:"threshold,omitempty"`       // Minimum semantic similarity, default 0 (no floor)
	Rerank         *bool   `json:"rerank,omitempty"`          // Override KB rerank opt-in (requires a configured reranker)
}
