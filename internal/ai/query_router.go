package ai

import (
	"context"
	"fmt"
	"strings"
)

// ChatbotKBStorage is the interface needed by the query router
type ChatbotKBStorage interface {
	GetChatbotKnowledgeBaseLinks(ctx context.Context, chatbotID string) ([]ChatbotKnowledgeBase, error)
}

// QueryRouter handles intelligent routing of queries to appropriate knowledge bases
type QueryRouter struct {
	storage ChatbotKBStorage
}

// NewQueryRouter creates a new query router
func NewQueryRouter(storage ChatbotKBStorage) *QueryRouter {
	return &QueryRouter{
		storage: storage,
	}
}

// RouteQuery determines which knowledge bases should be queried based on intent
type RouteQuery struct {
	ChatbotID      string
	QueryText      string
	ConversationID string // Optional: for conversation context
	UserID         string // Optional: for personalization
}

// RouteResult contains the routing decision
type RouteResult struct {
	SelectedKBs    []SelectedKnowledgeBase `json:"selected_kbs"`
	FallbackToAll  bool                    `json:"fallback_to_all"` // True if no intent match
	MatchedIntents []string                `json:"matched_intents"` // Keywords that matched
	TraceID        string                  `json:"trace_id"`        // For observability
}

// SelectedKnowledgeBase represents a KB selected for querying with its config
type SelectedKnowledgeBase struct {
	KnowledgeBaseID     string                 `json:"knowledge_base_id"`
	KnowledgeBaseName   string                 `json:"knowledge_base_name"`
	AccessLevel         AccessLevel            `json:"access_level"`
	ContextWeight       float64                `json:"context_weight"`
	Priority            int                    `json:"priority"`
	FilterExpression    map[string]interface{} `json:"filter_expression,omitempty"`
	MaxChunks           *int                   `json:"max_chunks,omitempty"`
	SimilarityThreshold *float64               `json:"similarity_threshold,omitempty"`
}

// Route selects appropriate knowledge bases for a query based on:
// 1. Intent keyword matching
// 2. Priority ordering (for tiered access)
// 3. Context weighting
func (r *QueryRouter) Route(ctx context.Context, query RouteQuery) (*RouteResult, error) {
	// Generate trace ID for observability
	traceID := NewTraceIDGenerator().GenerateTraceID()

	// Get all linked KBs for this chatbot
	links, err := r.storage.GetChatbotKnowledgeBaseLinks(ctx, query.ChatbotID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chatbot KB links: %w", err)
	}

	// Filter enabled links only
	var enabledLinks []ChatbotKnowledgeBase
	for _, link := range links {
		if link.Enabled {
			enabledLinks = append(enabledLinks, link)
		}
	}

	// Try intent-based routing first
	if result := r.tryIntentRouting(query.QueryText, enabledLinks, traceID); result != nil {
		return result, nil
	}

	// No intent match - use all enabled KBs
	return r.buildFallbackResult(enabledLinks, traceID), nil
}

// tryIntentRouting attempts to match query against intent keywords
func (r *QueryRouter) tryIntentRouting(queryText string, links []ChatbotKnowledgeBase, traceID string) *RouteResult {
	queryLower := strings.ToLower(queryText)

	var selected []SelectedKnowledgeBase
	var matchedIntents []string

	// Find KBs with matching intent keywords
	for _, link := range links {
		if len(link.IntentKeywords) == 0 {
			// No intent keywords defined - will be handled in fallback
			continue
		}

		// Check if any keyword matches
		for _, keyword := range link.IntentKeywords {
			if strings.Contains(queryLower, strings.ToLower(keyword)) {
				selected = append(selected, SelectedKnowledgeBase{
					KnowledgeBaseID:     link.KnowledgeBaseID,
					KnowledgeBaseName:   link.KnowledgeBaseName,
					AccessLevel:         AccessLevel(link.AccessLevel),
					ContextWeight:       link.ContextWeight,
					Priority:            link.Priority,
					FilterExpression:    link.FilterExpression,
					MaxChunks:           link.MaxChunks,
					SimilarityThreshold: link.SimilarityThreshold,
				})
				matchedIntents = append(matchedIntents, keyword)
				break // KB matched, don't check other keywords
			}
		}
	}

	// If we found matches, return them
	if len(selected) > 0 {
		// Sort by context weight (descending) and priority (ascending)
		r.sortSelectedKBs(selected)

		return &RouteResult{
			SelectedKBs:    selected,
			FallbackToAll:  false,
			MatchedIntents: matchedIntents,
			TraceID:        traceID,
		}
	}

	// No intent matches
	return nil
}

// buildFallbackResult creates a result using all enabled KBs
func (r *QueryRouter) buildFallbackResult(links []ChatbotKnowledgeBase, traceID string) *RouteResult {
	selected := make([]SelectedKnowledgeBase, 0, len(links))

	for _, link := range links {
		selected = append(selected, SelectedKnowledgeBase{
			KnowledgeBaseID:     link.KnowledgeBaseID,
			KnowledgeBaseName:   link.KnowledgeBaseName,
			AccessLevel:         AccessLevel(link.AccessLevel),
			ContextWeight:       link.ContextWeight,
			Priority:            link.Priority,
			FilterExpression:    link.FilterExpression,
			MaxChunks:           link.MaxChunks,
			SimilarityThreshold: link.SimilarityThreshold,
		})
	}

	// Sort by context weight (descending) and priority (ascending)
	r.sortSelectedKBs(selected)

	return &RouteResult{
		SelectedKBs:    selected,
		FallbackToAll:  true,
		MatchedIntents: []string{},
		TraceID:        traceID,
	}
}

// sortSelectedKBs sorts KBs by context weight (desc) and priority (asc)
func (r *QueryRouter) sortSelectedKBs(kbs []SelectedKnowledgeBase) {
	// Simple bubble sort - good enough for small lists
	n := len(kbs)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			// Higher context weight comes first
			if kbs[j].ContextWeight < kbs[j+1].ContextWeight {
				kbs[j], kbs[j+1] = kbs[j+1], kbs[j]
			} else if kbs[j].ContextWeight == kbs[j+1].ContextWeight {
				// Same weight, lower priority comes first
				if kbs[j].Priority > kbs[j+1].Priority {
					kbs[j], kbs[j+1] = kbs[j+1], kbs[j]
				}
			}
		}
	}
}

// SelectKBsByEntityType is a placeholder for future entity-based routing
// This will be used in Phase 6 (Knowledge Graph) for entity-centric routing
func (r *QueryRouter) SelectKBsByEntityType(ctx context.Context, chatbotID string, entityType string, entityValue string) ([]SelectedKnowledgeBase, error) {
	// TODO: Implement entity-based routing when knowledge graph is available
	// For now, fall back to getting all linked KBs
	links, err := r.storage.GetChatbotKnowledgeBaseLinks(ctx, chatbotID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chatbot KB links: %w", err)
	}

	var selected []SelectedKnowledgeBase
	for _, link := range links {
		if link.Enabled {
			selected = append(selected, SelectedKnowledgeBase{
				KnowledgeBaseID:     link.KnowledgeBaseID,
				KnowledgeBaseName:   link.KnowledgeBaseName,
				AccessLevel:         AccessLevel(link.AccessLevel),
				ContextWeight:       link.ContextWeight,
				Priority:            link.Priority,
				FilterExpression:    link.FilterExpression,
				MaxChunks:           link.MaxChunks,
				SimilarityThreshold: link.SimilarityThreshold,
			})
		}
	}

	return selected, nil
}
