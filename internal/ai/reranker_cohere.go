package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// CohereReranker re-ranks retrieval results via the Cohere Rerank API
// (https://api.cohere.com/v2/rerank). It implements the Reranker interface.
type CohereReranker struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

// CohereRerankerConfig configures a CohereReranker.
type CohereRerankerConfig struct {
	APIKey  string
	Model   string        // default "rerank-v3.5"
	BaseURL string        // default "https://api.cohere.com" (override for self-hosted/compatible APIs)
	Timeout time.Duration // default 10s
}

// DefaultCohereRerankModel is the rerank model used when none is configured.
const DefaultCohereRerankModel = "rerank-v3.5"

// NewCohereReranker creates a Cohere Rerank API client.
func NewCohereReranker(cfg CohereRerankerConfig) *CohereReranker {
	if cfg.Model == "" {
		cfg.Model = DefaultCohereRerankModel
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.cohere.com"
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Second
	}
	return &CohereReranker{
		apiKey:  cfg.APIKey,
		model:   cfg.Model,
		baseURL: cfg.BaseURL,
		client:  &http.Client{Timeout: cfg.Timeout},
	}
}

// Provider identifies the rerank backend for metrics.
func (c *CohereReranker) Provider() string { return "cohere" }

// cohereRerankRequest is the Rerank v2 API request body.
type cohereRerankRequest struct {
	Model     string   `json:"model"`
	Query     string   `json:"query"`
	Documents []string `json:"documents"`
	TopN      int      `json:"top_n,omitempty"`
}

// cohereRerankResponse is the Rerank v2 API response body.
type cohereRerankResponse struct {
	Results []struct {
		Index          int     `json:"index"`
		RelevanceScore float64 `json:"relevance_score"`
	} `json:"results"`
}

// Rerank re-orders results by relevance to the query.
func (c *CohereReranker) Rerank(ctx context.Context, query string, results []RetrievalResult, topN int) ([]RetrievalResult, error) {
	if len(results) == 0 {
		return results, nil
	}
	if topN <= 0 || topN > len(results) {
		topN = len(results)
	}

	documents := make([]string, len(results))
	for i, r := range results {
		documents[i] = truncateForExtraction(r.Content, 2000)
	}

	reqBody, err := json.Marshal(cohereRerankRequest{
		Model:     c.model,
		Query:     query,
		Documents: documents,
		TopN:      topN,
	})
	if err != nil {
		return nil, fmt.Errorf("cohere rerank: encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v2/rerank", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("cohere rerank: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cohere rerank: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("cohere rerank: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cohere rerank: status %d: %s", resp.StatusCode, truncateForExtraction(string(body), 300))
	}

	var parsed cohereRerankResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("cohere rerank: decode response: %w", err)
	}
	if len(parsed.Results) == 0 {
		return nil, fmt.Errorf("cohere rerank: empty results")
	}

	reranked := make([]RetrievalResult, 0, len(parsed.Results))
	for _, rr := range parsed.Results {
		if rr.Index < 0 || rr.Index >= len(results) {
			continue
		}
		r := results[rr.Index]
		r.Similarity = rr.RelevanceScore
		reranked = append(reranked, r)
	}
	return reranked, nil
}
