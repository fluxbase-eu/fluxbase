package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
)

// EdgeFunctionPipeline executes edge function-based document transformations
type EdgeFunctionPipeline struct {
	// Runtime is not directly used here to avoid circular dependencies
	// The actual execution will be done through the functions.Handler
}

// NewEdgeFunctionPipeline creates a new edge function pipeline
func NewEdgeFunctionPipeline() *EdgeFunctionPipeline {
	return &EdgeFunctionPipeline{}
}

// EdgeFunctionRequest is the request sent to the edge function
type EdgeFunctionRequest struct {
	Event         string        `json:"event"` // "document.created"
	Document      Document      `json:"document"`
	KnowledgeBase KnowledgeBase `json:"knowledge_base"`
}

// EdgeFunctionResponse is the expected response from the edge function
type EdgeFunctionResponse struct {
	Content        string                 `json:"content"`
	Metadata       map[string]interface{} `json:"metadata"`
	ShouldChunk    bool                   `json:"should_chunk"`
	ChunkingConfig *ChunkingOverride      `json:"chunking_config,omitempty"`
}

// ChunkingOverride allows the function to customize chunking behavior
type ChunkingOverride struct {
	Strategy string `json:"strategy"` // "recursive", "sentence", "paragraph", "fixed"
	Size     int    `json:"size"`
	Overlap  int    `json:"overlap"`
}

// ExecuteTransform runs an edge function for document transformation
func (p *EdgeFunctionPipeline) ExecuteTransform(ctx context.Context, kb *KnowledgeBase, document *Document) (*TransformResult, error) {
	if kb.PipelineType != string(PipelineTypeEdgeFunction) {
		// No edge function pipeline configured
		return &TransformResult{
			Content:     document.Content,
			Metadata:    map[string]interface{}{},
			ShouldChunk: true,
		}, nil
	}

	// Get function name from pipeline config
	functionName, ok := kb.PipelineConfig["function_name"].(string)
	if !ok || functionName == "" {
		return &TransformResult{
			Content:     document.Content,
			Metadata:    map[string]interface{}{},
			ShouldChunk: true,
		}, nil
	}

	// Prepare the request for the edge function
	req := EdgeFunctionRequest{
		Event:         "document.created",
		Document:      *document,
		KnowledgeBase: *kb,
	}

	_, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal edge function request: %w", err)
	}

	// Log the request for debugging
	log.Debug().
		Str("kb_id", kb.ID).
		Str("document_id", document.ID).
		Str("function", functionName).
		Msg("Executing edge function pipeline")

	// TODO: Execute the edge function through the functions.Handler
	// This requires integration with the actual Deno runtime
	// For now, return the original content with a flag
	// The actual execution will be implemented by connecting to the functions.Handler

	return &TransformResult{
		Content: document.Content,
		Metadata: map[string]interface{}{
			"pipeline_executed": true,
			"function_name":     functionName,
		},
		ShouldChunk: true,
	}, nil
}

// EdgeFunctionInvoker handles direct edge function invocation for pipelines
// This is a placeholder for future implementation
type EdgeFunctionInvoker struct {
	functionLoader any
}

// NewEdgeFunctionInvoker creates a new invoker
func NewEdgeFunctionInvoker(loader any) *EdgeFunctionInvoker {
	return &EdgeFunctionInvoker{
		functionLoader: loader,
	}
}

// InvokeForTransformation calls an edge function for document transformation
// This is a placeholder for future implementation
func (i *EdgeFunctionInvoker) InvokeForTransformation(
	ctx context.Context,
	functionName string,
	req EdgeFunctionRequest,
) (*EdgeFunctionResponse, error) {
	// TODO: Implement actual edge function invocation
	// This will connect to the functions.Handler to load and execute the function
	log.Info().
		Str("function", functionName).
		Str("document_id", req.Document.ID).
		Msg("Edge function pipeline execution (placeholder)")

	return &EdgeFunctionResponse{
		Content:     req.Document.Content,
		Metadata:    map[string]interface{}{},
		ShouldChunk: true,
	}, nil
}
