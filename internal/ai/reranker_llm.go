package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// LLMReranker re-ranks retrieval results listwise using the configured chat
// provider: the model sees the query plus numbered candidates and returns the
// candidate indices ordered by relevance. Zero additional vendor dependency —
// it reuses whatever default provider the instance already has.
type LLMReranker struct {
	resolve func(ctx context.Context) (Provider, error)

	model       string        // optional override of the provider's configured model
	maxTokens   int           // completion budget for the ordering response
	timeout     time.Duration // per-request timeout
	maxDocs     int           // candidates sent to the model per request
	docMaxChars int           // per-candidate truncation
}

// LLMRerankerConfig configures an LLMReranker.
type LLMRerankerConfig struct {
	Resolve     func(ctx context.Context) (Provider, error)
	Model       string
	MaxTokens   int           // default 500
	Timeout     time.Duration // default 15s
	MaxDocs     int           // default 30
	DocMaxChars int           // default 600
}

// NewLLMReranker creates a provider-backed listwise reranker.
func NewLLMReranker(cfg LLMRerankerConfig) *LLMReranker {
	if cfg.MaxTokens <= 0 {
		cfg.MaxTokens = 500
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 15 * time.Second
	}
	if cfg.MaxDocs <= 0 {
		cfg.MaxDocs = 30
	}
	if cfg.DocMaxChars <= 0 {
		cfg.DocMaxChars = 600
	}
	return &LLMReranker{
		resolve:     cfg.Resolve,
		model:       cfg.Model,
		maxTokens:   cfg.MaxTokens,
		timeout:     cfg.Timeout,
		maxDocs:     cfg.MaxDocs,
		docMaxChars: cfg.DocMaxChars,
	}
}

// Provider identifies the rerank backend for metrics.
func (l *LLMReranker) Provider() string { return "llm" }

// rerankPromptTemplate asks the model for a relevance-ordered index list.
const rerankPromptTemplate = `You are a search result reranker. Given a query and numbered document candidates, order the candidate numbers from most to least relevant to the query.

Respond with ONLY a JSON object (no markdown fences, no prose):
{"order": [<candidate number>, <candidate number>, ...]}

Rules:
- Include every candidate number exactly once.
- Candidates that do not answer the query go last, in their original order.

Query: %s

Candidates:
%s`

// Rerank re-orders results by relevance to the query. On any provider or
// parse failure the caller falls back to the original order.
func (l *LLMReranker) Rerank(ctx context.Context, query string, results []RetrievalResult, topN int) ([]RetrievalResult, error) {
	if len(results) == 0 {
		return results, nil
	}
	if topN <= 0 || topN > len(results) {
		topN = len(results)
	}

	if l.resolve == nil {
		return nil, fmt.Errorf("llm rerank: no provider resolver configured")
	}

	rerankCtx, cancel := context.WithTimeout(ctx, l.timeout)
	defer cancel()

	provider, err := l.resolve(rerankCtx)
	if err != nil {
		return nil, fmt.Errorf("llm rerank: resolve provider: %w", err)
	}
	if provider == nil {
		return nil, fmt.Errorf("llm rerank: no AI provider configured")
	}

	docs := results
	if len(docs) > l.maxDocs {
		docs = docs[:l.maxDocs]
	}

	var candidates strings.Builder
	for i, r := range docs {
		fmt.Fprintf(&candidates, "[%d] %s\n\n", i+1, truncateForExtraction(r.Content, l.docMaxChars))
	}

	req := &ChatRequest{
		Model: l.model, // empty = provider's configured default model
		Messages: []Message{
			{Role: RoleUser, Content: fmt.Sprintf(rerankPromptTemplate, query, candidates.String())},
		},
		MaxTokens:   l.maxTokens,
		Temperature: 0.0,
	}

	resp, err := provider.Chat(rerankCtx, req)
	if err != nil {
		return nil, fmt.Errorf("llm rerank: request failed: %w", err)
	}
	if resp == nil || len(resp.Choices) == 0 {
		return nil, fmt.Errorf("llm rerank: provider returned no choices")
	}

	order, err := parseRerankOrder(resp.Choices[0].Message.Content, len(docs))
	if err != nil {
		return nil, err
	}

	// Rank-position score: the model's ordering mapped onto a 0-1 scale so
	// downstream consumers see a monotone relevance value.
	reranked := make([]RetrievalResult, 0, len(order))
	used := make(map[int]bool, len(order))
	for _, idx := range order {
		if idx < 0 || idx >= len(docs) || used[idx] {
			continue
		}
		used[idx] = true
		r := docs[idx]
		r.Similarity = 1 - (float64(len(reranked)) / float64(len(docs)+1))
		reranked = append(reranked, r)
		if len(reranked) >= topN {
			return reranked, nil
		}
	}

	// Append candidates the model omitted, keeping their original relative
	// order, until topN is satisfied.
	for i, r := range docs {
		if used[i] {
			continue
		}
		r.Similarity = 1 - (float64(len(reranked)) / float64(len(docs)+1))
		reranked = append(reranked, r)
		if len(reranked) >= topN {
			break
		}
	}

	return reranked, nil
}

// parseRerankOrder extracts {"order": [1, 3, 2, ...]} from a model response,
// tolerating markdown fences and surrounding prose. Indices are returned
// 0-based in the model's relevance order; duplicates and out-of-range entries
// are skipped by the caller.
func parseRerankOrder(content string, docCount int) ([]int, error) {
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start < 0 || end <= start {
		return nil, fmt.Errorf("llm rerank: no JSON object in response")
	}

	var payload struct {
		Order []json.Number `json:"order"`
	}
	if err := json.Unmarshal([]byte(content[start:end+1]), &payload); err != nil {
		return nil, fmt.Errorf("llm rerank: invalid JSON in response: %w", err)
	}
	if len(payload.Order) == 0 {
		return nil, fmt.Errorf("llm rerank: empty order list")
	}

	order := make([]int, 0, len(payload.Order))
	for _, n := range payload.Order {
		idx, err := strconv.Atoi(strings.TrimSpace(n.String()))
		if err != nil {
			continue
		}
		order = append(order, idx-1) // candidates are 1-numbered in the prompt
	}
	if len(order) == 0 {
		return nil, fmt.Errorf("llm rerank: no usable indices in order list")
	}

	valid := 0
	for _, idx := range order {
		if idx >= 0 && idx < docCount {
			valid++
		}
	}
	if valid == 0 {
		return nil, fmt.Errorf("llm rerank: all indices out of range")
	}
	return order, nil
}
