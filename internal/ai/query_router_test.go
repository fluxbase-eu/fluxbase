package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock storage for testing
type mockQueryRouterStorage struct {
	links []ChatbotKnowledgeBase
}

func (m *mockQueryRouterStorage) GetChatbotKnowledgeBaseLinks(ctx context.Context, chatbotID string) ([]ChatbotKnowledgeBase, error) {
	return m.links, nil
}

func TestQueryRouter_SelectKB_ByIntent(t *testing.T) {
	t.Run("selects KB when query matches intent keyword", func(t *testing.T) {
		storage := &mockQueryRouterStorage{
			links: []ChatbotKnowledgeBase{
				{
					ID:              "link-1",
					ChatbotID:       "chatbot-1",
					KnowledgeBaseID: "kb-1",
					KnowledgeBaseName: "Technical Docs",
					AccessLevel:     "full",
					ContextWeight:   1.0,
					Priority:        100,
					IntentKeywords:  []string{"technical", "api", "troubleshooting"},
					Enabled:         true,
				},
				{
					ID:              "link-2",
					ChatbotID:       "chatbot-1",
					KnowledgeBaseID: "kb-2",
					KnowledgeBaseName: "FAQ",
					AccessLevel:     "full",
					ContextWeight:   0.8,
					Priority:        200,
					IntentKeywords:  []string{"faq", "help", "support"},
					Enabled:         true,
				},
			},
		}

		router := NewQueryRouter(storage)
		result, err := router.Route(context.Background(), RouteQuery{
			ChatbotID: "chatbot-1",
			QueryText: "How do I troubleshoot the API connection?",
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 1, len(result.SelectedKBs))
		assert.Equal(t, "Technical Docs", result.SelectedKBs[0].KnowledgeBaseName)
		// Query contains "API" which matches the "api" keyword
		assert.Contains(t, result.MatchedIntents, "api")
		assert.False(t, result.FallbackToAll)
		assert.NotEmpty(t, result.TraceID)
	})
}

func TestQueryRouter_SelectKB_ByEntityType(t *testing.T) {
	t.Run("entity type selection placeholder returns all KBs", func(t *testing.T) {
		storage := &mockQueryRouterStorage{
			links: []ChatbotKnowledgeBase{
				{
					ID:              "link-1",
					ChatbotID:       "chatbot-1",
					KnowledgeBaseID: "kb-1",
					KnowledgeBaseName: "KB 1",
					AccessLevel:     "full",
					ContextWeight:   1.0,
					Priority:        100,
					IntentKeywords:  []string{},
					Enabled:         true,
				},
			},
		}

		router := NewQueryRouter(storage)
		kbs, err := router.SelectKBsByEntityType(context.Background(), "chatbot-1", "person", "John Doe")

		require.NoError(t, err)
		assert.Equal(t, 1, len(kbs))
		assert.Equal(t, "KB 1", kbs[0].KnowledgeBaseName)
	})
}

func TestQueryRouter_SelectKB_Fallback(t *testing.T) {
	t.Run("falls back to all KBs when no intent match", func(t *testing.T) {
		storage := &mockQueryRouterStorage{
			links: []ChatbotKnowledgeBase{
				{
					ID:              "link-1",
					ChatbotID:       "chatbot-1",
					KnowledgeBaseID: "kb-1",
					KnowledgeBaseName: "KB 1",
					AccessLevel:     "full",
					ContextWeight:   1.0,
					Priority:        100,
					IntentKeywords:  []string{"technical"},
					Enabled:         true,
				},
			},
		}

		router := NewQueryRouter(storage)
		result, err := router.Route(context.Background(), RouteQuery{
			ChatbotID: "chatbot-1",
			QueryText: "What is the weather today?",
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 1, len(result.SelectedKBs))
		assert.True(t, result.FallbackToAll)
		assert.Empty(t, result.MatchedIntents)
	})
}

func TestQueryRouter_PriorityOrdering(t *testing.T) {
	t.Run("sorts by context weight and priority", func(t *testing.T) {
		storage := &mockQueryRouterStorage{
			links: []ChatbotKnowledgeBase{
				{
					ID:              "link-1",
					ChatbotID:       "chatbot-1",
					KnowledgeBaseID: "kb-1",
					KnowledgeBaseName: "Medium Weight",
					AccessLevel:     "full",
					ContextWeight:   0.8,
					Priority:        100,
					IntentKeywords:  []string{},
					Enabled:         true,
				},
				{
					ID:              "link-2",
					ChatbotID:       "chatbot-1",
					KnowledgeBaseID: "kb-2",
					KnowledgeBaseName: "High Weight",
					AccessLevel:     "full",
					ContextWeight:   1.0,
					Priority:        200,
					IntentKeywords:  []string{},
					Enabled:         true,
				},
				{
					ID:              "link-3",
					ChatbotID:       "chatbot-1",
					KnowledgeBaseID: "kb-3",
					KnowledgeBaseName: "High Weight Low Priority",
					AccessLevel:     "full",
					ContextWeight:   1.0,
					Priority:        50,
					IntentKeywords:  []string{},
					Enabled:         true,
				},
			},
		}

		router := NewQueryRouter(storage)
		result, err := router.Route(context.Background(), RouteQuery{
			ChatbotID: "chatbot-1",
			QueryText: "general query",
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 3, len(result.SelectedKBs))
		// High weight, low priority comes first, then high weight high priority, then medium
		assert.Equal(t, "High Weight Low Priority", result.SelectedKBs[0].KnowledgeBaseName)
		assert.Equal(t, "High Weight", result.SelectedKBs[1].KnowledgeBaseName)
		assert.Equal(t, "Medium Weight", result.SelectedKBs[2].KnowledgeBaseName)
	})
}

func TestQueryRouter_DisabledKBs(t *testing.T) {
	t.Run("filters out disabled KBs", func(t *testing.T) {
		storage := &mockQueryRouterStorage{
			links: []ChatbotKnowledgeBase{
				{
					ID:              "link-1",
					ChatbotID:       "chatbot-1",
					KnowledgeBaseID: "kb-1",
					KnowledgeBaseName: "Enabled KB",
					AccessLevel:     "full",
					ContextWeight:   1.0,
					Priority:        100,
					IntentKeywords:  []string{},
					Enabled:         true,
				},
				{
					ID:              "link-2",
					ChatbotID:       "chatbot-1",
					KnowledgeBaseID: "kb-2",
					KnowledgeBaseName: "Disabled KB",
					AccessLevel:     "full",
					ContextWeight:   1.0,
					Priority:        50,
					IntentKeywords:  []string{},
					Enabled:         false,
				},
			},
		}

		router := NewQueryRouter(storage)
		result, err := router.Route(context.Background(), RouteQuery{
			ChatbotID: "chatbot-1",
			QueryText: "test query",
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 1, len(result.SelectedKBs))
		assert.Equal(t, "Enabled KB", result.SelectedKBs[0].KnowledgeBaseName)
	})
}

func TestQueryRouter_TraceID(t *testing.T) {
	t.Run("generates unique trace IDs", func(t *testing.T) {
		storage := &mockQueryRouterStorage{
			links: []ChatbotKnowledgeBase{
				{
					ID:              "link-1",
					ChatbotID:       "chatbot-1",
					KnowledgeBaseID: "kb-1",
					KnowledgeBaseName: "KB 1",
					AccessLevel:     "full",
					ContextWeight:   1.0,
					Priority:        100,
					IntentKeywords:  []string{},
					Enabled:         true,
				},
			},
		}

		router := NewQueryRouter(storage)
		result1, _ := router.Route(context.Background(), RouteQuery{
			ChatbotID: "chatbot-1",
			QueryText: "query 1",
		})
		result2, _ := router.Route(context.Background(), RouteQuery{
			ChatbotID: "chatbot-1",
			QueryText: "query 2",
		})

		assert.NotEmpty(t, result1.TraceID)
		assert.NotEmpty(t, result2.TraceID)
		assert.NotEqual(t, result1.TraceID, result2.TraceID)
	})
}
