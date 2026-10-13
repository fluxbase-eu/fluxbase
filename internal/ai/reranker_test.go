package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- RRF fusion ---

func TestRRFScore(t *testing.T) {
	// Ranked first in both lists with 50/50 weights → perfect score.
	assert.InDelta(t, 1.0, rrfScore(0.5, 0.5, 1, 1), 0.0001)

	// First in semantic only, absent from keyword list: score = weight share.
	assert.InDelta(t, 0.7, rrfScore(0.7, 0.3, 1, 0), 0.0001)

	// Deeper ranks decay monotonically.
	second := rrfScore(0.5, 0.5, 2, 2)
	first := rrfScore(0.5, 0.5, 1, 1)
	assert.Less(t, second, first)
	assert.Greater(t, second, rrfScore(0.5, 0.5, 10, 10))

	// Absent from both lists scores zero.
	assert.Equal(t, 0.0, rrfScore(0.5, 0.5, 0, 0))

	// Weights must sum to a bounded score.
	assert.LessOrEqual(t, rrfScore(0.9, 0.1, 1, 1), 1.0)
}

func TestFuseHybridResults(t *testing.T) {
	rows := []hybridRow{
		{result: RetrievalResult{ChunkID: "both-top"}, semanticRank: 1, keywordRank: 1, vectorSim: 0.9},
		{result: RetrievalResult{ChunkID: "semantic-only"}, semanticRank: 2, keywordRank: 0, vectorSim: 0.8},
		{result: RetrievalResult{ChunkID: "keyword-strong"}, semanticRank: 5, keywordRank: 1, vectorSim: 0.6},
		{result: RetrievalResult{ChunkID: "below-threshold"}, semanticRank: 6, keywordRank: 2, vectorSim: 0.4},
	}

	t.Run("fuses and orders by RRF score", func(t *testing.T) {
		opts := HybridSearchOptions{SemanticWeight: 0.5, Limit: 10, Threshold: 0}
		fused := fuseHybridResults(rows, opts)

		require.Len(t, fused, 4)
		assert.Equal(t, "both-top", fused[0].ChunkID)
		// keyword-strong (keyword rank 1 + semantic rank 5 ≈ 0.97) beats
		// below-threshold (kw 2 + sem 6 ≈ 0.95) beats semantic-only
		// (sem rank 2, absent from keyword list ≈ 0.49) — a keyword-list
		// presence is worth roughly a semantic rank of ~30 in RRF.
		assert.Equal(t, "keyword-strong", fused[1].ChunkID)
		assert.Equal(t, "below-threshold", fused[2].ChunkID)
		assert.Equal(t, "semantic-only", fused[3].ChunkID)
	})

	t.Run("threshold filters on cosine similarity not fused score", func(t *testing.T) {
		opts := HybridSearchOptions{SemanticWeight: 0.5, Limit: 10, Threshold: 0.5}
		fused := fuseHybridResults(rows, opts)

		ids := make(map[string]bool, len(fused))
		for _, r := range fused {
			ids[r.ChunkID] = true
		}
		assert.False(t, ids["below-threshold"], "cosine 0.4 < threshold 0.5 must be dropped")
		assert.True(t, ids["keyword-strong"], "cosine 0.6 passes the threshold")
	})

	t.Run("limit truncates", func(t *testing.T) {
		opts := HybridSearchOptions{SemanticWeight: 0.5, Limit: 2, Threshold: 0}
		fused := fuseHybridResults(rows, opts)
		assert.Len(t, fused, 2)
	})
}

// --- Rerank plumbing ---

func TestRerankResults_StampsMetadataAndReorders(t *testing.T) {
	results := []RetrievalResult{
		{ChunkID: "c1", Content: "first", Similarity: 0.9},
		{ChunkID: "c2", Content: "second", Similarity: 0.8},
	}

	fake := newFakeProvider("rerank-test")
	fake.SetChatFn(func(req *ChatRequest) (*ChatResponse, error) {
		return newSimpleResponse(`{"order": [2, 1]}`), nil
	})
	reranker := NewLLMReranker(LLMRerankerConfig{
		Resolve: func(ctx context.Context) (Provider, error) { return fake, nil },
	})

	reranked := rerankResults(context.Background(), reranker, "query", results, 0, nil)

	require.Len(t, reranked, 2)
	assert.Equal(t, "c2", reranked[0].ChunkID, "model ranked candidate 2 first")

	// Original input slice must be stamped with rerank metadata too.
	var meta map[string]any
	require.NoError(t, json.Unmarshal(results[0].Metadata, &meta))
	assert.Equal(t, "llm", meta["rerank_by"])
	assert.Equal(t, float64(1), meta["rerank_rank"], "c1 moved to rank 1 (0-based)")
}

func TestRerankResults_ErrorReturnsOriginal(t *testing.T) {
	results := []RetrievalResult{
		{ChunkID: "c1", Content: "first", Similarity: 0.9},
		{ChunkID: "c2", Content: "second", Similarity: 0.8},
	}

	fake := newFakeProvider("rerank-test")
	fake.QueueError(assert.AnError)
	fake.QueueError(assert.AnError)
	reranker := NewLLMReranker(LLMRerankerConfig{
		Resolve: func(ctx context.Context) (Provider, error) { return fake, nil },
	})

	reranked := rerankResults(context.Background(), reranker, "query", results, 0, nil)

	assert.Equal(t, "c1", reranked[0].ChunkID, "original order preserved on rerank failure")
	assert.InDelta(t, 0.9, reranked[0].Similarity, 0.0001)
}

func TestRerankResults_NilRerankerIsNoop(t *testing.T) {
	results := []RetrievalResult{{ChunkID: "c1", Similarity: 0.9}}
	reranked := rerankResults(context.Background(), nil, "query", results, 0, nil)
	assert.Equal(t, results, reranked)
}

// --- LLM reranker ---

func TestLLMReranker_ParsesOrder(t *testing.T) {
	order, err := parseRerankOrder("Sure! ```json\n{\"order\": [3, 1, 2]}\n``` done", 3)
	require.NoError(t, err)
	assert.Equal(t, []int{2, 0, 1}, order) // 0-based
}

func TestLLMReranker_RejectsGarbage(t *testing.T) {
	_, err := parseRerankOrder("I cannot do that", 3)
	assert.Error(t, err)

	_, err = parseRerankOrder(`{"order": [99, 100]}`, 3)
	assert.Error(t, err, "all indices out of range must fail")
}

func TestLLMReranker_EndToEnd(t *testing.T) {
	results := []RetrievalResult{
		{ChunkID: "a", Content: "apples"},
		{ChunkID: "b", Content: "oranges"},
		{ChunkID: "c", Content: "fruit"},
	}

	fake := newFakeProvider("rerank-test")
	fake.SetChatFn(func(req *ChatRequest) (*ChatResponse, error) {
		// The prompt must contain the query and numbered candidates.
		prompt := req.Messages[0].Content
		assert.Contains(t, prompt, "tropical fruit")
		assert.Contains(t, prompt, "[1] apples")
		return newSimpleResponse(`{"order": [3, 1]}`), nil
	})
	reranker := NewLLMReranker(LLMRerankerConfig{
		Resolve: func(ctx context.Context) (Provider, error) { return fake, nil },
	})

	reranked, err := reranker.Rerank(context.Background(), "tropical fruit", results, 0)
	require.NoError(t, err)

	// Candidate 3 first; candidate 2 was omitted by the model and appended.
	require.Len(t, reranked, 3)
	assert.Equal(t, "c", reranked[0].ChunkID)
	assert.Equal(t, "a", reranked[1].ChunkID)
	assert.Equal(t, "b", reranked[2].ChunkID)
	// Scores are monotone by rank position.
	assert.Greater(t, reranked[0].Similarity, reranked[1].Similarity)
	assert.Greater(t, reranked[1].Similarity, reranked[2].Similarity)
}

func TestLLMReranker_NoProvider(t *testing.T) {
	reranker := NewLLMReranker(LLMRerankerConfig{
		Resolve: func(ctx context.Context) (Provider, error) { return nil, nil },
	})
	_, err := reranker.Rerank(context.Background(), "q", []RetrievalResult{{ChunkID: "a", Content: "x"}}, 0)
	assert.Error(t, err)
}

// --- Cohere reranker ---

func TestCohereReranker_EndToEnd(t *testing.T) {
	var gotAuth, gotPath string
	var gotBody cohereRerankRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results": [{"index": 1, "relevance_score": 0.97}, {"index": 0, "relevance_score": 0.42}]}`))
	}))
	defer server.Close()

	reranker := NewCohereReranker(CohereRerankerConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "rerank-test",
	})

	results := []RetrievalResult{
		{ChunkID: "c1", Content: "less relevant"},
		{ChunkID: "c2", Content: "most relevant"},
	}

	reranked, err := reranker.Rerank(context.Background(), "query", results, 0)
	require.NoError(t, err)

	assert.Equal(t, "Bearer test-key", gotAuth)
	assert.Equal(t, "/v2/rerank", gotPath)
	assert.Equal(t, "rerank-test", gotBody.Model)
	assert.Equal(t, "query", gotBody.Query)
	assert.Len(t, gotBody.Documents, 2)

	require.Len(t, reranked, 2)
	assert.Equal(t, "c2", reranked[0].ChunkID)
	assert.InDelta(t, 0.97, reranked[0].Similarity, 0.0001)
	assert.Equal(t, "c1", reranked[1].ChunkID)
	assert.InDelta(t, 0.42, reranked[1].Similarity, 0.0001)
}

func TestCohereReranker_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message": "invalid api key"}`))
	}))
	defer server.Close()

	reranker := NewCohereReranker(CohereRerankerConfig{APIKey: "bad", BaseURL: server.URL})
	_, err := reranker.Rerank(context.Background(), "q", []RetrievalResult{{ChunkID: "a", Content: "x"}}, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 401")
}
