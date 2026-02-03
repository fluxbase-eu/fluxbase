package ai

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKnowledgeBase_Struct(t *testing.T) {
	t.Run("all fields can be set", func(t *testing.T) {
		createdBy := "admin"

		kb := KnowledgeBase{
			ID:                  "kb-123",
			Name:                "test-kb",
			Namespace:           "default",
			Description:         "Test knowledge base",
			EmbeddingModel:      "text-embedding-3-small",
			EmbeddingDimensions: 1536,
			ChunkSize:           512,
			ChunkOverlap:        50,
			ChunkStrategy:       "recursive",
			Enabled:             true,
			DocumentCount:       10,
			TotalChunks:         100,
			Source:              "api",
			CreatedBy:           &createdBy,
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}

		assert.Equal(t, "kb-123", kb.ID)
		assert.Equal(t, "test-kb", kb.Name)
		assert.Equal(t, "default", kb.Namespace)
		assert.Equal(t, 1536, kb.EmbeddingDimensions)
		assert.Equal(t, 512, kb.ChunkSize)
		assert.True(t, kb.Enabled)
		assert.Equal(t, 10, kb.DocumentCount)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		kb := KnowledgeBase{
			ID:                  "kb-456",
			Name:                "docs-kb",
			Namespace:           "prod",
			EmbeddingModel:      "text-embedding-ada-002",
			EmbeddingDimensions: 1536,
			ChunkSize:           256,
			ChunkOverlap:        25,
			ChunkStrategy:       "sentence",
			Enabled:             false,
			CreatedAt:           time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			UpdatedAt:           time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		}

		data, err := json.Marshal(kb)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"id":"kb-456"`)
		assert.Contains(t, string(data), `"name":"docs-kb"`)
		assert.Contains(t, string(data), `"embedding_dimensions":1536`)
		assert.Contains(t, string(data), `"chunk_size":256`)
	})

	t.Run("zero value knowledge base", func(t *testing.T) {
		var kb KnowledgeBase
		assert.Empty(t, kb.ID)
		assert.Empty(t, kb.Name)
		assert.Equal(t, 0, kb.EmbeddingDimensions)
		assert.False(t, kb.Enabled)
		assert.Nil(t, kb.CreatedBy)
	})
}

func TestKnowledgeBase_ToSummary(t *testing.T) {
	t.Run("converts all fields correctly", func(t *testing.T) {
		kb := KnowledgeBase{
			ID:            "kb-789",
			Name:          "summary-test",
			Namespace:     "default",
			Description:   "Test description",
			Enabled:       true,
			DocumentCount: 5,
			TotalChunks:   50,
			UpdatedAt:     time.Date(2024, 6, 15, 12, 30, 0, 0, time.UTC),
		}

		summary := kb.ToSummary()
		assert.Equal(t, "kb-789", summary.ID)
		assert.Equal(t, "summary-test", summary.Name)
		assert.Equal(t, "default", summary.Namespace)
		assert.Equal(t, "Test description", summary.Description)
		assert.True(t, summary.Enabled)
		assert.Equal(t, 5, summary.DocumentCount)
		assert.Equal(t, 50, summary.TotalChunks)
		assert.Contains(t, summary.UpdatedAt, "2024-06-15")
	})
}

func TestDocument_Struct(t *testing.T) {
	t.Run("all fields can be set", func(t *testing.T) {
		createdBy := "user-123"
		indexedAt := time.Now()

		doc := Document{
			ID:              "doc-123",
			KnowledgeBaseID: "kb-456",
			Title:           "Test Document",
			SourceURL:       "https://example.com/doc.pdf",
			SourceType:      "upload",
			MimeType:        "application/pdf",
			Content:         "Document content here",
			ContentHash:     "abc123hash",
			Status:          DocumentStatusIndexed,
			ChunksCount:     10,
			Tags:            []string{"important", "review"},
			CreatedBy:       &createdBy,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			IndexedAt:       &indexedAt,
		}

		assert.Equal(t, "doc-123", doc.ID)
		assert.Equal(t, "Test Document", doc.Title)
		assert.Equal(t, DocumentStatusIndexed, doc.Status)
		assert.Equal(t, 10, doc.ChunksCount)
		assert.Len(t, doc.Tags, 2)
		assert.NotNil(t, doc.IndexedAt)
	})

	t.Run("document status constants", func(t *testing.T) {
		assert.Equal(t, DocumentStatus("pending"), DocumentStatusPending)
		assert.Equal(t, DocumentStatus("processing"), DocumentStatusProcessing)
		assert.Equal(t, DocumentStatus("indexed"), DocumentStatusIndexed)
		assert.Equal(t, DocumentStatus("failed"), DocumentStatusFailed)
	})
}

func TestDocument_ToSummary(t *testing.T) {
	t.Run("converts all fields correctly", func(t *testing.T) {
		doc := Document{
			ID:          "doc-456",
			Title:       "Summary Doc",
			SourceType:  "manual",
			Status:      DocumentStatusProcessing,
			ChunksCount: 5,
			Tags:        []string{"tag1", "tag2"},
			UpdatedAt:   time.Date(2024, 7, 20, 15, 45, 0, 0, time.UTC),
		}

		summary := doc.ToSummary()
		assert.Equal(t, "doc-456", summary.ID)
		assert.Equal(t, "Summary Doc", summary.Title)
		assert.Equal(t, "manual", summary.SourceType)
		assert.Equal(t, DocumentStatusProcessing, summary.Status)
		assert.Equal(t, 5, summary.ChunksCount)
		assert.Len(t, summary.Tags, 2)
		assert.Contains(t, summary.UpdatedAt, "2024-07-20")
	})
}

func TestChunk_Struct(t *testing.T) {
	t.Run("all fields can be set", func(t *testing.T) {
		startOffset := 0
		endOffset := 500
		tokenCount := 125

		chunk := Chunk{
			ID:              "chunk-123",
			DocumentID:      "doc-456",
			KnowledgeBaseID: "kb-789",
			Content:         "Chunk content here",
			ChunkIndex:      0,
			StartOffset:     &startOffset,
			EndOffset:       &endOffset,
			TokenCount:      &tokenCount,
			Embedding:       []float32{0.1, 0.2, 0.3},
			CreatedAt:       time.Now(),
		}

		assert.Equal(t, "chunk-123", chunk.ID)
		assert.Equal(t, "doc-456", chunk.DocumentID)
		assert.Equal(t, 0, chunk.ChunkIndex)
		assert.Equal(t, 0, *chunk.StartOffset)
		assert.Equal(t, 500, *chunk.EndOffset)
		assert.Len(t, chunk.Embedding, 3)
	})

	t.Run("JSON omits embedding by default", func(t *testing.T) {
		chunk := Chunk{
			ID:        "chunk-789",
			Content:   "Test content",
			Embedding: []float32{0.1, 0.2},
		}

		data, err := json.Marshal(chunk)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"id":"chunk-789"`)
		assert.Contains(t, string(data), `"content":"Test content"`)
	})
}

func TestChatbotKnowledgeBase_Struct(t *testing.T) {
	t.Run("all fields can be set", func(t *testing.T) {
		link := ChatbotKnowledgeBase{
			ID:                  "link-123",
			ChatbotID:           "chatbot-456",
			KnowledgeBaseID:     "kb-789",
			Enabled:             true,
			MaxChunks:           5,
			SimilarityThreshold: 0.7,
			Priority:            1,
			CreatedAt:           time.Now(),
		}

		assert.Equal(t, "link-123", link.ID)
		assert.Equal(t, "chatbot-456", link.ChatbotID)
		assert.Equal(t, "kb-789", link.KnowledgeBaseID)
		assert.True(t, link.Enabled)
		assert.Equal(t, 5, link.MaxChunks)
		assert.Equal(t, 0.7, link.SimilarityThreshold)
		assert.Equal(t, 1, link.Priority)
	})
}

func TestRetrievalResult_Struct(t *testing.T) {
	t.Run("all fields can be set", func(t *testing.T) {
		result := RetrievalResult{
			ChunkID:           "chunk-123",
			DocumentID:        "doc-456",
			KnowledgeBaseID:   "kb-789",
			KnowledgeBaseName: "docs-kb",
			DocumentTitle:     "Test Document",
			Content:           "Retrieved chunk content",
			Similarity:        0.85,
			Tags:              []string{"relevant"},
		}

		assert.Equal(t, "chunk-123", result.ChunkID)
		assert.Equal(t, "doc-456", result.DocumentID)
		assert.Equal(t, 0.85, result.Similarity)
		assert.Equal(t, "docs-kb", result.KnowledgeBaseName)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		result := RetrievalResult{
			ChunkID:    "chunk-456",
			Content:    "Content here",
			Similarity: 0.92,
		}

		data, err := json.Marshal(result)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"chunk_id":"chunk-456"`)
		assert.Contains(t, string(data), `"similarity":0.92`)
	})
}

func TestMetadataFilter_Struct(t *testing.T) {
	t.Run("user isolation filter", func(t *testing.T) {
		userID := "user-123"
		filter := MetadataFilter{
			UserID:        &userID,
			Tags:          []string{"public", "shared"},
			IncludeGlobal: true,
			Metadata:      map[string]string{"department": "engineering"},
		}

		assert.Equal(t, "user-123", *filter.UserID)
		assert.Len(t, filter.Tags, 2)
		assert.True(t, filter.IncludeGlobal)
		assert.Equal(t, "engineering", filter.Metadata["department"])
	})

	t.Run("zero value filter", func(t *testing.T) {
		var filter MetadataFilter
		assert.Nil(t, filter.UserID)
		assert.Nil(t, filter.Tags)
		assert.False(t, filter.IncludeGlobal)
		assert.Nil(t, filter.Metadata)
	})
}

func TestVectorSearchOptions_Struct(t *testing.T) {
	t.Run("all fields can be set", func(t *testing.T) {
		userID := "user-456"
		opts := VectorSearchOptions{
			ChatbotID:      "chatbot-123",
			Query:          "search query",
			KnowledgeBases: []string{"kb-1", "kb-2"},
			Limit:          10,
			Threshold:      0.5,
			Tags:           []string{"important"},
			Metadata:       map[string]string{"type": "article"},
			UserID:         &userID,
			IsAdmin:        false,
		}

		assert.Equal(t, "chatbot-123", opts.ChatbotID)
		assert.Equal(t, "search query", opts.Query)
		assert.Len(t, opts.KnowledgeBases, 2)
		assert.Equal(t, 10, opts.Limit)
		assert.Equal(t, 0.5, opts.Threshold)
		assert.False(t, opts.IsAdmin)
	})
}

func TestRetrievalLog_Struct(t *testing.T) {
	t.Run("all fields can be set", func(t *testing.T) {
		chatbotID := "chatbot-123"
		conversationID := "conv-456"
		kbID := "kb-789"
		userID := "user-abc"

		logEntry := RetrievalLog{
			ID:                  "log-123",
			ChatbotID:           &chatbotID,
			ConversationID:      &conversationID,
			KnowledgeBaseID:     &kbID,
			UserID:              &userID,
			QueryText:           "What is AI?",
			QueryEmbeddingModel: "text-embedding-3-small",
			ChunksRetrieved:     5,
			ChunkIDs:            []string{"chunk-1", "chunk-2", "chunk-3", "chunk-4", "chunk-5"},
			SimilarityScores:    []float64{0.95, 0.90, 0.85, 0.80, 0.75},
			RetrievalDurationMs: 150,
			CreatedAt:           time.Now(),
		}

		assert.Equal(t, "log-123", logEntry.ID)
		assert.Equal(t, "chatbot-123", *logEntry.ChatbotID)
		assert.Equal(t, 5, logEntry.ChunksRetrieved)
		assert.Len(t, logEntry.ChunkIDs, 5)
		assert.Len(t, logEntry.SimilarityScores, 5)
		assert.Equal(t, 150, logEntry.RetrievalDurationMs)
	})
}

func TestCreateKnowledgeBaseRequest_Struct(t *testing.T) {
	t.Run("all fields can be set", func(t *testing.T) {
		req := CreateKnowledgeBaseRequest{
			Name:                "new-kb",
			Namespace:           "prod",
			Description:         "A new knowledge base",
			EmbeddingModel:      "text-embedding-3-large",
			EmbeddingDimensions: 3072,
			ChunkSize:           1024,
			ChunkOverlap:        100,
			ChunkStrategy:       "paragraph",
		}

		assert.Equal(t, "new-kb", req.Name)
		assert.Equal(t, "prod", req.Namespace)
		assert.Equal(t, 3072, req.EmbeddingDimensions)
		assert.Equal(t, 1024, req.ChunkSize)
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{
			"name": "test-kb",
			"namespace": "default",
			"chunk_size": 512,
			"chunk_overlap": 50
		}`

		var req CreateKnowledgeBaseRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)
		assert.Equal(t, "test-kb", req.Name)
		assert.Equal(t, "default", req.Namespace)
		assert.Equal(t, 512, req.ChunkSize)
	})
}

func TestUpdateKnowledgeBaseRequest_Struct(t *testing.T) {
	t.Run("partial update with pointers", func(t *testing.T) {
		name := "updated-name"
		enabled := false
		chunkSize := 256

		req := UpdateKnowledgeBaseRequest{
			Name:      &name,
			Enabled:   &enabled,
			ChunkSize: &chunkSize,
		}

		assert.Equal(t, "updated-name", *req.Name)
		assert.False(t, *req.Enabled)
		assert.Equal(t, 256, *req.ChunkSize)
		assert.Nil(t, req.Description)
		assert.Nil(t, req.EmbeddingModel)
	})
}

func TestCreateDocumentRequest_Struct(t *testing.T) {
	t.Run("all fields can be set", func(t *testing.T) {
		req := CreateDocumentRequest{
			Title:            "Test Document",
			Content:          "Document content here",
			SourceURL:        "https://example.com/doc",
			SourceType:       "upload",
			MimeType:         "text/plain",
			Metadata:         map[string]string{"author": "John"},
			Tags:             []string{"important"},
			OriginalFilename: "doc.txt",
		}

		assert.Equal(t, "Test Document", req.Title)
		assert.Equal(t, "upload", req.SourceType)
		assert.Equal(t, "John", req.Metadata["author"])
		assert.Equal(t, "doc.txt", req.OriginalFilename)
	})
}

func TestLinkKnowledgeBaseRequest_Struct(t *testing.T) {
	t.Run("all fields can be set", func(t *testing.T) {
		maxChunks := 10
		threshold := 0.6
		priority := 2

		req := LinkKnowledgeBaseRequest{
			KnowledgeBaseID:     "kb-123",
			MaxChunks:           &maxChunks,
			SimilarityThreshold: &threshold,
			Priority:            &priority,
		}

		assert.Equal(t, "kb-123", req.KnowledgeBaseID)
		assert.Equal(t, 10, *req.MaxChunks)
		assert.Equal(t, 0.6, *req.SimilarityThreshold)
		assert.Equal(t, 2, *req.Priority)
	})
}

func TestChunkingStrategy_Constants(t *testing.T) {
	t.Run("all strategies defined", func(t *testing.T) {
		assert.Equal(t, ChunkingStrategy("recursive"), ChunkingStrategyRecursive)
		assert.Equal(t, ChunkingStrategy("sentence"), ChunkingStrategySentence)
		assert.Equal(t, ChunkingStrategy("paragraph"), ChunkingStrategyParagraph)
		assert.Equal(t, ChunkingStrategy("fixed"), ChunkingStrategyFixed)
	})
}

func TestDefaultKnowledgeBaseConfig(t *testing.T) {
	t.Run("returns sensible defaults", func(t *testing.T) {
		defaults := DefaultKnowledgeBaseConfig()

		assert.Equal(t, "default", defaults.Namespace)
		assert.Equal(t, "text-embedding-3-small", defaults.EmbeddingModel)
		assert.Equal(t, 1536, defaults.EmbeddingDimensions)
		assert.Equal(t, 512, defaults.ChunkSize)
		assert.Equal(t, 50, defaults.ChunkOverlap)
		assert.Equal(t, string(ChunkingStrategyRecursive), defaults.ChunkStrategy)
	})

	t.Run("name and description are empty", func(t *testing.T) {
		defaults := DefaultKnowledgeBaseConfig()

		assert.Empty(t, defaults.Name)
		assert.Empty(t, defaults.Description)
	})
}

func TestVectorSearchResult_Struct(t *testing.T) {
	t.Run("all fields can be set", func(t *testing.T) {
		result := VectorSearchResult{
			ChunkID:           "chunk-123",
			DocumentID:        "doc-456",
			DocumentTitle:     "Test Doc",
			KnowledgeBaseName: "kb-name",
			Content:           "Search result content",
			Similarity:        0.88,
			Tags:              []string{"tag1", "tag2"},
		}

		assert.Equal(t, "chunk-123", result.ChunkID)
		assert.Equal(t, "Test Doc", result.DocumentTitle)
		assert.Equal(t, 0.88, result.Similarity)
		assert.Len(t, result.Tags, 2)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		result := VectorSearchResult{
			ChunkID:           "chunk-789",
			KnowledgeBaseName: "test-kb",
			Similarity:        0.95,
		}

		data, err := json.Marshal(result)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"chunk_id":"chunk-789"`)
		assert.Contains(t, string(data), `"knowledge_base_name":"test-kb"`)
		assert.Contains(t, string(data), `"similarity":0.95`)
	})
}
