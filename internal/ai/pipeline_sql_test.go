package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPipelineType_Values(t *testing.T) {
	t.Run("has correct pipeline type values", func(t *testing.T) {
		assert.Equal(t, "none", string(PipelineTypeNone))
		assert.Equal(t, "sql", string(PipelineTypeSQL))
		assert.Equal(t, "edge_function", string(PipelineTypeEdgeFunction))
		assert.Equal(t, "webhook", string(PipelineTypeWebhook))
	})
}

func TestTransformResult_Struct(t *testing.T) {
	t.Run("transform result structure", func(t *testing.T) {
		result := &TransformResult{
			Content:     "Cleaned content",
			Metadata:    map[string]interface{}{"word_count": 2},
			ShouldChunk: true,
		}

		assert.Equal(t, "Cleaned content", result.Content)
		assert.Equal(t, 2, result.Metadata["word_count"])
		assert.True(t, result.ShouldChunk)
	})
}

func TestSQLPipeline_ExecuteTransform_NoPipeline(t *testing.T) {
	t.Run("bypasses when no pipeline configured", func(t *testing.T) {
		pipeline := &SQLPipeline{}

		kb := &KnowledgeBase{
			ID:           "kb-1",
			PipelineType: string(PipelineTypeNone),
		}

		document := &Document{
			ID:      "doc-1",
			Content: "Original content",
		}

		result, err := pipeline.ExecuteTransform(context.Background(), kb, document)

		assert.NoError(t, err)
		assert.Equal(t, "Original content", result.Content)
		assert.True(t, result.ShouldChunk)
	})
}

func TestSQLPipeline_ExecuteTransform_SQLType(t *testing.T) {
	t.Run("SQL pipeline type structure", func(t *testing.T) {
		kb := &KnowledgeBase{
			ID:                     "kb-1",
			PipelineType:           string(PipelineTypeSQL),
			TransformationFunction: strPtr("ai.clean_document"),
		}

		assert.Equal(t, string(PipelineTypeSQL), kb.PipelineType)
		assert.NotNil(t, kb.TransformationFunction)
		assert.Equal(t, "ai.clean_document", *kb.TransformationFunction)
	})
}

func TestTransformResult_CanDisableChunking(t *testing.T) {
	t.Run("transform can disable chunking", func(t *testing.T) {
		result := &TransformResult{
			Content:     "Pre-chunked content",
			Metadata:    map[string]interface{}{},
			ShouldChunk: false,
		}

		assert.False(t, result.ShouldChunk)
		assert.Equal(t, "Pre-chunked content", result.Content)
	})
}

func TestTransformResult_MetadataHandling(t *testing.T) {
	t.Run("handles complex metadata", func(t *testing.T) {
		result := &TransformResult{
			Content: "Content with metadata",
			Metadata: map[string]interface{}{
				"word_count":      3,
				"language":        "en",
				"has_code":        true,
				"processing_time": 0.5,
			},
			ShouldChunk: true,
		}

		assert.Equal(t, 3, result.Metadata["word_count"])
		assert.Equal(t, "en", result.Metadata["language"])
		assert.True(t, result.Metadata["has_code"].(bool))
		assert.InDelta(t, 0.5, result.Metadata["processing_time"].(float64), 0.01)
	})
}

func TestPipelineType_EdgeFunction(t *testing.T) {
	t.Run("edge function pipeline type", func(t *testing.T) {
		kb := &KnowledgeBase{
			ID:           "kb-1",
			PipelineType: string(PipelineTypeEdgeFunction),
		}

		assert.Equal(t, "edge_function", kb.PipelineType)
	})
}

func TestPipelineType_Webhook(t *testing.T) {
	t.Run("webhook pipeline type", func(t *testing.T) {
		kb := &KnowledgeBase{
			ID:           "kb-1",
			PipelineType: string(PipelineTypeWebhook),
		}

		assert.Equal(t, "webhook", kb.PipelineType)
	})
}

func TestKnowledgeBase_PipelineFields(t *testing.T) {
	t.Run("knowledge base has pipeline fields", func(t *testing.T) {
		functionName := "ai.transform"
		kb := &KnowledgeBase{
			ID:                     "kb-1",
			PipelineType:           string(PipelineTypeSQL),
			PipelineConfig:         map[string]interface{}{"timeout": 30},
			TransformationFunction: &functionName,
		}

		assert.Equal(t, "sql", kb.PipelineType)
		assert.Equal(t, 30, kb.PipelineConfig["timeout"])
		assert.Equal(t, "ai.transform", *kb.TransformationFunction)
	})
}
