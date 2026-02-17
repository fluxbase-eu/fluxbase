package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEdgeFunctionPipeline_ExecuteTransform_NoPipeline(t *testing.T) {
	t.Run("bypasses when no edge function pipeline configured", func(t *testing.T) {
		pipeline := &EdgeFunctionPipeline{}

		kb := &KnowledgeBase{
			ID:           "kb-1",
			PipelineType: string(PipelineTypeNone),
		}

		document := &Document{
			ID:      "doc-1",
			Content: "Original content",
		}

		result, err := pipeline.ExecuteTransform(context.Background(), kb, document)

		require.NoError(t, err)
		assert.Equal(t, "Original content", result.Content)
		assert.True(t, result.ShouldChunk)
	})
}

func TestEdgeFunctionPipeline_ExecuteTransform_SQLType(t *testing.T) {
	t.Run("bypasses for SQL pipeline type", func(t *testing.T) {
		pipeline := &EdgeFunctionPipeline{}

		kb := &KnowledgeBase{
			ID:           "kb-1",
			PipelineType: string(PipelineTypeSQL),
		}

		document := &Document{
			ID:      "doc-1",
			Content: "Original content",
		}

		result, err := pipeline.ExecuteTransform(context.Background(), kb, document)

		require.NoError(t, err)
		assert.Equal(t, "Original content", result.Content)
	})
}

func TestEdgeFunctionPipeline_ExecuteTransform_EdgeFunction(t *testing.T) {
	t.Run("recognizes edge function pipeline type", func(t *testing.T) {
		pipeline := &EdgeFunctionPipeline{}

		kb := &KnowledgeBase{
			ID:           "kb-1",
			PipelineType: string(PipelineTypeEdgeFunction),
			PipelineConfig: map[string]interface{}{
				"function_name": "transform-document",
				"timeout":       30,
			},
		}

		document := &Document{
			ID:      "doc-1",
			Content: "Original content",
		}

		result, err := pipeline.ExecuteTransform(context.Background(), kb, document)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.ShouldChunk)
		assert.Equal(t, "Original content", result.Content) // Will be transformed when runtime is connected
	})
}

func TestEdgeFunctionRequest_Struct(t *testing.T) {
	t.Run("edge function request structure", func(t *testing.T) {
		document := Document{
			ID:      "doc-1",
			Content: "Test content",
		}

		kb := KnowledgeBase{
			ID:   "kb-1",
			Name: "Test KB",
		}

		req := EdgeFunctionRequest{
			Event:         "document.created",
			Document:      document,
			KnowledgeBase: kb,
		}

		assert.Equal(t, "document.created", req.Event)
		assert.Equal(t, "doc-1", req.Document.ID)
		assert.Equal(t, "kb-1", req.KnowledgeBase.ID)
	})
}

func TestEdgeFunctionResponse_Struct(t *testing.T) {
	t.Run("edge function response structure", func(t *testing.T) {
		response := EdgeFunctionResponse{
			Content:     "Transformed content",
			Metadata:    map[string]interface{}{"word_count": 2},
			ShouldChunk: true,
		}

		assert.Equal(t, "Transformed content", response.Content)
		assert.Equal(t, 2, response.Metadata["word_count"])
		assert.True(t, response.ShouldChunk)
	})
}

func TestEdgeFunctionResponse_WithChunkingOverride(t *testing.T) {
	t.Run("response with chunking override", func(t *testing.T) {
		response := EdgeFunctionResponse{
			Content:     "Code content",
			Metadata:    map[string]interface{}{"language": "go"},
			ShouldChunk: true,
			ChunkingConfig: &ChunkingOverride{
				Strategy: "fixed",
				Size:     256,
				Overlap:  0,
			},
		}

		assert.NotNil(t, response.ChunkingConfig)
		assert.Equal(t, "fixed", response.ChunkingConfig.Strategy)
		assert.Equal(t, 256, response.ChunkingConfig.Size)
		assert.Equal(t, 0, response.ChunkingConfig.Overlap)
	})
}

func TestChunkingOverride_Struct(t *testing.T) {
	t.Run("chunking override structure", func(t *testing.T) {
		override := &ChunkingOverride{
			Strategy: "recursive",
			Size:     512,
			Overlap:  50,
		}

		assert.Equal(t, "recursive", override.Strategy)
		assert.Equal(t, 512, override.Size)
		assert.Equal(t, 50, override.Overlap)
	})
}

func TestChunkingOverride_CanDisableChunking(t *testing.T) {
	t.Run("edge function can disable chunking", func(t *testing.T) {
		response := EdgeFunctionResponse{
			Content:     "Pre-chunked data",
			Metadata:    map[string]interface{}{},
			ShouldChunk: false,
		}

		assert.False(t, response.ShouldChunk)
	})
}

func TestEdgeFunctionPipeline_MissingFunctionName(t *testing.T) {
	t.Run("handles missing function name gracefully", func(t *testing.T) {
		pipeline := &EdgeFunctionPipeline{}

		kb := &KnowledgeBase{
			ID:           "kb-1",
			PipelineType: string(PipelineTypeEdgeFunction),
			PipelineConfig: map[string]interface{}{
				"timeout": 30, // No function_name
			},
		}

		document := &Document{
			ID:      "doc-1",
			Content: "Original content",
		}

		result, err := pipeline.ExecuteTransform(context.Background(), kb, document)

		require.NoError(t, err)
		assert.Equal(t, "Original content", result.Content)
		assert.True(t, result.ShouldChunk)
	})
}

func TestEdgeFunctionPipeline_EmptyFunctionName(t *testing.T) {
	t.Run("handles empty function name gracefully", func(t *testing.T) {
		pipeline := &EdgeFunctionPipeline{}

		kb := &KnowledgeBase{
			ID:           "kb-1",
			PipelineType: string(PipelineTypeEdgeFunction),
			PipelineConfig: map[string]interface{}{
				"function_name": "", // Empty string
			},
		}

		document := &Document{
			ID:      "doc-1",
			Content: "Original content",
		}

		result, err := pipeline.ExecuteTransform(context.Background(), kb, document)

		require.NoError(t, err)
		assert.Equal(t, "Original content", result.Content)
	})
}

func TestEdgeFunctionInvoker_Struct(t *testing.T) {
	t.Run("edge function invoker structure", func(t *testing.T) {
		mockLoader := struct{}{}
		invoker := NewEdgeFunctionInvoker(mockLoader)

		assert.NotNil(t, invoker)
		assert.NotNil(t, invoker.functionLoader)
	})
}

func TestEdgeFunctionInvoker_InvokeForTransformation_Mock(t *testing.T) {
	t.Run("mock invoker returns response", func(t *testing.T) {
		invoker := &EdgeFunctionInvoker{}

		req := EdgeFunctionRequest{
			Event: "document.created",
			Document: Document{
				ID:      "doc-1",
				Content: "Test",
			},
			KnowledgeBase: KnowledgeBase{
				ID:   "kb-1",
				Name: "Test KB",
			},
		}

		// The placeholder implementation just returns the original content
		resp, err := invoker.InvokeForTransformation(context.Background(), "test-func", req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "Test", resp.Content)
		assert.True(t, resp.ShouldChunk)
	})
}
