package ai

import (
	"context"
	"encoding/json"
	"time"
)

// Reranker re-orders retrieved chunks against the query using a
// relevance model (cross-encoder, rerank API, or LLM listwise ranking).
// Implementations must be safe for concurrent use.
type Reranker interface {
	// Rerank re-orders results by relevance to the query, returning at most
	// topN entries. The returned slice may be a subset/reordering of the
	// input; implementations must not mutate the input slice.
	Rerank(ctx context.Context, query string, results []RetrievalResult, topN int) ([]RetrievalResult, error)

	// Provider identifies the rerank backend for metrics/logging
	// (e.g. "cohere", "llm").
	Provider() string
}

// rerankMetadataKey constants stamped onto result metadata after reranking.
const (
	metadataRerankScore = "rerank_score"
	metadataRerankRank  = "rerank_rank"
	metadataRerankBy    = "rerank_by"
)

// rerankResults is the shared entry point used by the RAG service and search
// handlers: applies the reranker, stamps metadata, and reports latency.
// On reranker failure the original results are returned unchanged — reranking
// is a quality optimization and must never fail a search.
func rerankResults(ctx context.Context, reranker Reranker, query string, results []RetrievalResult, topN int, record func(duration time.Duration, err error)) []RetrievalResult {
	if reranker == nil || len(results) == 0 {
		return results
	}
	if topN <= 0 || topN > len(results) {
		topN = len(results)
	}

	start := time.Now()
	reranked, err := reranker.Rerank(ctx, query, results, topN)
	duration := time.Since(start)

	if record != nil {
		record(duration, err)
	}
	if err != nil {
		return results
	}

	byID := make(map[string]int, len(reranked))
	for i, r := range reranked {
		byID[r.ChunkID] = i
	}
	for i := range results {
		pos, ok := byID[results[i].ChunkID]
		if !ok {
			continue
		}
		results[i].Similarity = reranked[pos].Similarity
		results[i].Metadata = stampRerankMetadata(results[i].Metadata, reranker.Provider(), reranked[pos].Similarity, pos)
	}

	return reranked
}

// stampRerankMetadata adds rerank diagnostics to a result's metadata JSON.
func stampRerankMetadata(metadata []byte, provider string, score float64, rank int) []byte {
	metadataMap := make(map[string]any)
	if len(metadata) > 0 {
		_ = json.Unmarshal(metadata, &metadataMap)
	}
	metadataMap[metadataRerankBy] = provider
	if score > 0 {
		metadataMap[metadataRerankScore] = score
	}
	if rank >= 0 {
		metadataMap[metadataRerankRank] = rank
	}
	metadataJSON, err := json.Marshal(metadataMap)
	if err != nil {
		return metadata
	}
	return metadataJSON
}
