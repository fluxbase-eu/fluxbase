package ai

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewKnowledgeBaseHandler(t *testing.T) {
	t.Run("creates handler with nil dependencies", func(t *testing.T) {
		handler := NewKnowledgeBaseHandler(nil, nil)
		assert.NotNil(t, handler)
		assert.Nil(t, handler.storage)
		assert.Nil(t, handler.processor)
		assert.NotNil(t, handler.textExtractor) // Always created
	})
}

func TestNewKnowledgeBaseHandlerWithOCR(t *testing.T) {
	t.Run("creates handler with OCR service", func(t *testing.T) {
		handler := NewKnowledgeBaseHandlerWithOCR(nil, nil, nil)
		assert.NotNil(t, handler)
		assert.Nil(t, handler.storage)
		assert.Nil(t, handler.processor)
		assert.NotNil(t, handler.textExtractor)
		assert.Nil(t, handler.ocrService)
	})
}

func TestKnowledgeBaseHandler_SetStorageService(t *testing.T) {
	t.Run("sets storage service", func(t *testing.T) {
		handler := NewKnowledgeBaseHandler(nil, nil)
		assert.Nil(t, handler.storageService)

		handler.SetStorageService(nil)
		assert.Nil(t, handler.storageService)
	})
}

func TestAddDocumentRequest_Struct(t *testing.T) {
	t.Run("all fields can be set", func(t *testing.T) {
		req := AddDocumentRequest{
			Title:    "Test Document",
			Content:  "Document content here",
			Source:   "https://example.com/doc",
			MimeType: "text/plain",
			Metadata: map[string]string{
				"author":     "John Doe",
				"department": "Engineering",
			},
		}

		assert.Equal(t, "Test Document", req.Title)
		assert.Equal(t, "Document content here", req.Content)
		assert.Equal(t, "https://example.com/doc", req.Source)
		assert.Equal(t, "text/plain", req.MimeType)
		assert.Equal(t, "John Doe", req.Metadata["author"])
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{
			"title": "API Doc",
			"content": "API documentation content",
			"source": "https://api.example.com",
			"mime_type": "text/markdown",
			"metadata": {"version": "1.0"}
		}`

		var req AddDocumentRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "API Doc", req.Title)
		assert.Equal(t, "text/markdown", req.MimeType)
		assert.Equal(t, "1.0", req.Metadata["version"])
	})

	t.Run("minimal request", func(t *testing.T) {
		jsonData := `{"content": "Just the content"}`

		var req AddDocumentRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Empty(t, req.Title)
		assert.Equal(t, "Just the content", req.Content)
		assert.Empty(t, req.Source)
		assert.Nil(t, req.Metadata)
	})
}

func TestSearchKnowledgeBaseRequest_Struct(t *testing.T) {
	t.Run("all fields can be set", func(t *testing.T) {
		req := SearchKnowledgeBaseRequest{
			Query:          "search query",
			MaxChunks:      10,
			Threshold:      0.5,
			Mode:           "hybrid",
			SemanticWeight: 0.7,
		}

		assert.Equal(t, "search query", req.Query)
		assert.Equal(t, 10, req.MaxChunks)
		assert.Equal(t, 0.5, req.Threshold)
		assert.Equal(t, "hybrid", req.Mode)
		assert.Equal(t, 0.7, req.SemanticWeight)
	})

	t.Run("JSON deserialization with all modes", func(t *testing.T) {
		modes := []string{"semantic", "keyword", "hybrid"}
		for _, mode := range modes {
			jsonData := `{"query": "test", "mode": "` + mode + `"}`
			var req SearchKnowledgeBaseRequest
			err := json.Unmarshal([]byte(jsonData), &req)
			require.NoError(t, err)
			assert.Equal(t, mode, req.Mode)
		}
	})

	t.Run("default values not set", func(t *testing.T) {
		jsonData := `{"query": "test"}`
		var req SearchKnowledgeBaseRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "test", req.Query)
		assert.Equal(t, 0, req.MaxChunks)          // Will use default in handler
		assert.Equal(t, float64(0), req.Threshold) // Will use default in handler
		assert.Empty(t, req.Mode)                  // Will default to semantic
	})
}

func TestUpdateChatbotKnowledgeBaseRequest_Struct(t *testing.T) {
	t.Run("all fields", func(t *testing.T) {
		priority := 2
		maxChunks := 15
		threshold := 0.8
		enabled := false

		req := UpdateChatbotKnowledgeBaseRequest{
			Priority:            &priority,
			MaxChunks:           &maxChunks,
			SimilarityThreshold: &threshold,
			Enabled:             &enabled,
		}

		assert.Equal(t, 2, *req.Priority)
		assert.Equal(t, 15, *req.MaxChunks)
		assert.Equal(t, 0.8, *req.SimilarityThreshold)
		assert.False(t, *req.Enabled)
	})

	t.Run("partial update", func(t *testing.T) {
		enabled := true

		req := UpdateChatbotKnowledgeBaseRequest{
			Enabled: &enabled,
		}

		assert.Nil(t, req.Priority)
		assert.Nil(t, req.MaxChunks)
		assert.Nil(t, req.SimilarityThreshold)
		assert.True(t, *req.Enabled)
	})
}

func TestKnowledgeBaseCapabilities_Struct(t *testing.T) {
	t.Run("with OCR enabled", func(t *testing.T) {
		caps := KnowledgeBaseCapabilities{
			OCREnabled:         true,
			OCRAvailable:       true,
			OCRLanguages:       []string{"eng", "deu", "fra"},
			SupportedFileTypes: []string{".pdf", ".docx", ".txt", ".md"},
		}

		assert.True(t, caps.OCREnabled)
		assert.True(t, caps.OCRAvailable)
		assert.Len(t, caps.OCRLanguages, 3)
		assert.Contains(t, caps.OCRLanguages, "eng")
		assert.Len(t, caps.SupportedFileTypes, 4)
	})

	t.Run("without OCR", func(t *testing.T) {
		caps := KnowledgeBaseCapabilities{
			OCREnabled:         false,
			OCRAvailable:       false,
			OCRLanguages:       nil,
			SupportedFileTypes: []string{".txt", ".md"},
		}

		assert.False(t, caps.OCREnabled)
		assert.False(t, caps.OCRAvailable)
		assert.Nil(t, caps.OCRLanguages)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		caps := KnowledgeBaseCapabilities{
			OCREnabled:         true,
			OCRAvailable:       false,
			SupportedFileTypes: []string{".pdf"},
		}

		data, err := json.Marshal(caps)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"ocr_enabled":true`)
		assert.Contains(t, string(data), `"ocr_available":false`)
	})
}

func TestDebugSearchRequest_Struct(t *testing.T) {
	t.Run("basic request", func(t *testing.T) {
		req := DebugSearchRequest{
			Query: "debug search query",
		}

		assert.Equal(t, "debug search query", req.Query)
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{"query": "test debug"}`
		var req DebugSearchRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)
		assert.Equal(t, "test debug", req.Query)
	})
}

func TestDebugSearchResponse_Struct(t *testing.T) {
	t.Run("successful debug response", func(t *testing.T) {
		resp := DebugSearchResponse{
			Query:                  "test query",
			QueryEmbeddingPreview:  []float32{0.1, 0.2, 0.3},
			QueryEmbeddingDims:     1536,
			StoredEmbeddingPreview: []float32{0.15, 0.25, 0.35},
			RawSimilarities:        []float64{0.95, 0.90, 0.85},
			EmbeddingModel:         "text-embedding-3-small",
			KBEmbeddingModel:       "text-embedding-3-small",
			ChunksFound:            3,
			TopChunkContentPreview: "This is the top chunk content...",
			TotalChunks:            100,
			ChunksWithEmbedding:    100,
			ChunksWithoutEmbedding: 0,
		}

		assert.Equal(t, "test query", resp.Query)
		assert.Len(t, resp.QueryEmbeddingPreview, 3)
		assert.Equal(t, 1536, resp.QueryEmbeddingDims)
		assert.Len(t, resp.RawSimilarities, 3)
		assert.Equal(t, 0.95, resp.RawSimilarities[0])
		assert.Equal(t, 3, resp.ChunksFound)
		assert.Empty(t, resp.ErrorMessage)
	})

	t.Run("error response", func(t *testing.T) {
		resp := DebugSearchResponse{
			Query:                  "test",
			QueryEmbeddingDims:     1536,
			TotalChunks:            50,
			ChunksWithEmbedding:    0,
			ChunksWithoutEmbedding: 50,
			ErrorMessage:           "All chunks have NULL embeddings",
		}

		assert.Equal(t, 0, resp.ChunksWithEmbedding)
		assert.Equal(t, 50, resp.ChunksWithoutEmbedding)
		assert.NotEmpty(t, resp.ErrorMessage)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		resp := DebugSearchResponse{
			Query:              "test",
			QueryEmbeddingDims: 1536,
			ChunksFound:        5,
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"query":"test"`)
		assert.Contains(t, string(data), `"query_embedding_dims":1536`)
		assert.Contains(t, string(data), `"chunks_found":5`)
	})
}
