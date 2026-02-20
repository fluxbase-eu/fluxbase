package ai

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRAGService(t *testing.T) {
	t.Run("creates service with storage and embedding service", func(t *testing.T) {
		service := NewRAGService(nil, nil, nil, nil)
		assert.NotNil(t, service)
	})
}

func TestRAGService_FormatContext(t *testing.T) {
	service := &RAGService{}

	t.Run("returns empty string for empty chunks", func(t *testing.T) {
		result := service.formatContext([]RetrievalResult{})
		assert.Empty(t, result)
	})

	t.Run("returns empty string for nil chunks", func(t *testing.T) {
		result := service.formatContext(nil)
		assert.Empty(t, result)
	})

	t.Run("formats single chunk", func(t *testing.T) {
		chunks := []RetrievalResult{
			{
				ChunkID:           "chunk-1",
				Content:           "This is the content of the first chunk.",
				DocumentTitle:     "Test Document",
				KnowledgeBaseName: "Test KB",
				Similarity:        0.95,
			},
		}

		result := service.formatContext(chunks)
		assert.Contains(t, result, "## Relevant Knowledge")
		assert.Contains(t, result, "Source 1")
		assert.Contains(t, result, "Test Document")
		assert.Contains(t, result, "Test KB")
		assert.Contains(t, result, "0.95")
		assert.Contains(t, result, "This is the content of the first chunk.")
	})

	t.Run("formats multiple chunks", func(t *testing.T) {
		chunks := []RetrievalResult{
			{ChunkID: "1", Content: "First chunk", DocumentTitle: "Doc 1", Similarity: 0.95},
			{ChunkID: "2", Content: "Second chunk", DocumentTitle: "Doc 2", Similarity: 0.85},
			{ChunkID: "3", Content: "Third chunk", DocumentTitle: "Doc 3", Similarity: 0.75},
		}

		result := service.formatContext(chunks)
		assert.Contains(t, result, "Source 1")
		assert.Contains(t, result, "Source 2")
		assert.Contains(t, result, "Source 3")
		assert.Contains(t, result, "First chunk")
		assert.Contains(t, result, "Second chunk")
		assert.Contains(t, result, "Third chunk")
	})

	t.Run("handles chunk without document title", func(t *testing.T) {
		chunks := []RetrievalResult{
			{ChunkID: "1", Content: "Content", Similarity: 0.9},
		}

		result := service.formatContext(chunks)
		assert.Contains(t, result, "Source 1")
		assert.Contains(t, result, "Content")
		assert.NotContains(t, result, ": :") // No double colon from missing title
	})

	t.Run("handles chunk without knowledge base name", func(t *testing.T) {
		chunks := []RetrievalResult{
			{ChunkID: "1", Content: "Content", DocumentTitle: "Doc", Similarity: 0.9},
		}

		result := service.formatContext(chunks)
		assert.Contains(t, result, "Doc")
		assert.NotContains(t, result, "(from )") // No empty "from" clause
	})
}

func TestSortVectorSearchResults(t *testing.T) {
	t.Run("sorts by similarity descending", func(t *testing.T) {
		results := []VectorSearchResult{
			{ChunkID: "1", Similarity: 0.5},
			{ChunkID: "2", Similarity: 0.9},
			{ChunkID: "3", Similarity: 0.7},
		}

		sortVectorSearchResults(results)

		assert.Equal(t, "2", results[0].ChunkID) // 0.9
		assert.Equal(t, "3", results[1].ChunkID) // 0.7
		assert.Equal(t, "1", results[2].ChunkID) // 0.5
	})

	t.Run("handles empty slice", func(t *testing.T) {
		results := []VectorSearchResult{}
		sortVectorSearchResults(results)
		assert.Empty(t, results)
	})

	t.Run("handles single element", func(t *testing.T) {
		results := []VectorSearchResult{
			{ChunkID: "1", Similarity: 0.8},
		}
		sortVectorSearchResults(results)
		assert.Len(t, results, 1)
	})

	t.Run("handles already sorted slice", func(t *testing.T) {
		results := []VectorSearchResult{
			{ChunkID: "1", Similarity: 0.9},
			{ChunkID: "2", Similarity: 0.8},
			{ChunkID: "3", Similarity: 0.7},
		}

		sortVectorSearchResults(results)

		assert.Equal(t, "1", results[0].ChunkID)
		assert.Equal(t, "2", results[1].ChunkID)
		assert.Equal(t, "3", results[2].ChunkID)
	})

	t.Run("handles equal similarities", func(t *testing.T) {
		results := []VectorSearchResult{
			{ChunkID: "1", Similarity: 0.8},
			{ChunkID: "2", Similarity: 0.8},
		}

		sortVectorSearchResults(results)
		// Both should still be present
		assert.Len(t, results, 2)
		assert.Equal(t, 0.8, results[0].Similarity)
		assert.Equal(t, 0.8, results[1].Similarity)
	})
}

func TestOptString(t *testing.T) {
	t.Run("returns nil for empty string", func(t *testing.T) {
		result := optString("")
		assert.Nil(t, result)
	})

	t.Run("returns pointer to non-empty string", func(t *testing.T) {
		result := optString("hello")
		assert.NotNil(t, result)
		assert.Equal(t, "hello", *result)
	})

	t.Run("returns pointer to whitespace string", func(t *testing.T) {
		result := optString("  ")
		assert.NotNil(t, result)
		assert.Equal(t, "  ", *result)
	})
}

func TestRetrieveContextOptions_Struct(t *testing.T) {
	opts := RetrieveContextOptions{
		ChatbotID:      "bot-123",
		ConversationID: "conv-456",
		UserID:         "user-789",
		Query:          "What is the meaning of life?",
		MaxChunks:      5,
		Threshold:      0.7,
	}

	assert.Equal(t, "bot-123", opts.ChatbotID)
	assert.Equal(t, "conv-456", opts.ConversationID)
	assert.Equal(t, "user-789", opts.UserID)
	assert.Equal(t, "What is the meaning of life?", opts.Query)
	assert.Equal(t, 5, opts.MaxChunks)
	assert.Equal(t, 0.7, opts.Threshold)
}

func TestRetrieveContextResult_Struct(t *testing.T) {
	chunks := []RetrievalResult{
		{ChunkID: "1", Content: "Content 1"},
	}

	result := RetrieveContextResult{
		Chunks:           chunks,
		FormattedContext: "## Context\n\nContent 1",
		TotalRetrieved:   1,
		DurationMs:       25,
		EmbeddingModel:   "text-embedding-3-small",
	}

	assert.Len(t, result.Chunks, 1)
	assert.Contains(t, result.FormattedContext, "Context")
	assert.Equal(t, 1, result.TotalRetrieved)
	assert.Equal(t, int64(25), result.DurationMs)
	assert.Equal(t, "text-embedding-3-small", result.EmbeddingModel)
}

func TestChatbotRAGConfig_Struct(t *testing.T) {
	config := ChatbotRAGConfig{
		Enabled: true,
		KnowledgeBases: []KnowledgeBaseSummary{
			{ID: "kb-1", Name: "Knowledge Base 1"},
		},
		TotalMaxChunks: 10,
	}

	assert.True(t, config.Enabled)
	assert.Len(t, config.KnowledgeBases, 1)
	assert.Equal(t, "kb-1", config.KnowledgeBases[0].ID)
	assert.Equal(t, 10, config.TotalMaxChunks)
}

func TestKnowledgeBaseStats_Struct(t *testing.T) {
	now := time.Now()
	stats := KnowledgeBaseStats{
		ID:             "kb-123",
		Name:           "Test KB",
		DocumentCount:  10,
		TotalChunks:    50,
		PendingDocs:    2,
		IndexedDocs:    7,
		FailedDocs:     1,
		EmbeddingModel: "text-embedding-3-small",
		ChunkSize:      512,
		ChunkOverlap:   64,
		ChunkStrategy:  "recursive",
		Enabled:        true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	assert.Equal(t, "kb-123", stats.ID)
	assert.Equal(t, "Test KB", stats.Name)
	assert.Equal(t, 10, stats.DocumentCount)
	assert.Equal(t, 50, stats.TotalChunks)
	assert.Equal(t, 2, stats.PendingDocs)
	assert.Equal(t, 7, stats.IndexedDocs)
	assert.Equal(t, 1, stats.FailedDocs)
	assert.Equal(t, "text-embedding-3-small", stats.EmbeddingModel)
	assert.Equal(t, 512, stats.ChunkSize)
	assert.Equal(t, 64, stats.ChunkOverlap)
	assert.Equal(t, "recursive", stats.ChunkStrategy)
	assert.True(t, stats.Enabled)
}
