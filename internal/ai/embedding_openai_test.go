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

func TestNewOpenAIEmbeddingProviderInternal(t *testing.T) {
	t.Run("creates provider with default base URL", func(t *testing.T) {
		config := OpenAIConfig{
			APIKey: "test-key",
		}

		provider, err := newOpenAIEmbeddingProviderInternal(config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "test-key", provider.config.APIKey)
		assert.Equal(t, defaultOpenAIBaseURL, provider.config.BaseURL)
	})

	t.Run("uses custom base URL", func(t *testing.T) {
		config := OpenAIConfig{
			APIKey:  "test-key",
			BaseURL: "https://custom.api.com/",
		}

		provider, err := newOpenAIEmbeddingProviderInternal(config)
		require.NoError(t, err)
		// Trailing slash should be trimmed
		assert.Equal(t, "https://custom.api.com", provider.config.BaseURL)
	})
}

func TestOpenAIEmbeddingProvider_Embed(t *testing.T) {
	t.Run("returns error for empty texts", func(t *testing.T) {
		provider := &openAIEmbeddingProvider{
			config:     OpenAIConfig{},
			httpClient: http.DefaultClient,
		}

		_, err := provider.Embed(context.Background(), []string{}, "model")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no texts provided")
	})

	t.Run("uses default model when not specified", func(t *testing.T) {
		var receivedModel string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var reqBody map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&reqBody)
			require.NoError(t, err)
			receivedModel = reqBody["model"].(string)

			response := openAIEmbeddingResponse{
				Model: receivedModel,
				Data: []struct {
					Object    string    `json:"object"`
					Index     int       `json:"index"`
					Embedding []float32 `json:"embedding"`
				}{
					{Index: 0, Embedding: []float32{0.1}},
				},
			}
			err = json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		}))
		defer server.Close()

		provider := &openAIEmbeddingProvider{
			config: OpenAIConfig{
				APIKey:  "key",
				BaseURL: server.URL,
				Model:   "custom-model",
			},
			httpClient: server.Client(),
		}

		_, err := provider.Embed(context.Background(), []string{"test"}, "")
		require.NoError(t, err)
		assert.Equal(t, "custom-model", receivedModel)
	})

	t.Run("successful embedding with organization header", func(t *testing.T) {
		var receivedOrgHeader string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedOrgHeader = r.Header.Get("OpenAI-Organization")
			assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

			response := openAIEmbeddingResponse{
				Model: "text-embedding-3-small",
				Data: []struct {
					Object    string    `json:"object"`
					Index     int       `json:"index"`
					Embedding []float32 `json:"embedding"`
				}{
					{Index: 0, Embedding: []float32{0.1, 0.2, 0.3}},
				},
				Usage: struct {
					PromptTokens int `json:"prompt_tokens"`
					TotalTokens  int `json:"total_tokens"`
				}{
					PromptTokens: 10,
					TotalTokens:  10,
				},
			}
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		}))
		defer server.Close()

		provider := &openAIEmbeddingProvider{
			config: OpenAIConfig{
				APIKey:         "test-key",
				BaseURL:        server.URL,
				OrganizationID: "org-123",
			},
			httpClient: server.Client(),
		}

		result, err := provider.Embed(context.Background(), []string{"test"}, "text-embedding-3-small")
		require.NoError(t, err)
		assert.Equal(t, "org-123", receivedOrgHeader)
		assert.Len(t, result.Embeddings, 1)
		assert.Equal(t, 3, result.Dimensions)
		assert.Equal(t, 10, result.Usage.TotalTokens)
	})

	t.Run("handles multiple texts", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := openAIEmbeddingResponse{
				Model: "text-embedding-3-small",
				Data: []struct {
					Object    string    `json:"object"`
					Index     int       `json:"index"`
					Embedding []float32 `json:"embedding"`
				}{
					{Index: 0, Embedding: []float32{0.1, 0.2}},
					{Index: 1, Embedding: []float32{0.3, 0.4}},
					{Index: 2, Embedding: []float32{0.5, 0.6}},
				},
			}
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		}))
		defer server.Close()

		provider := &openAIEmbeddingProvider{
			config: OpenAIConfig{
				APIKey:  "key",
				BaseURL: server.URL,
			},
			httpClient: server.Client(),
		}

		result, err := provider.Embed(context.Background(), []string{"text1", "text2", "text3"}, "model")
		require.NoError(t, err)
		assert.Len(t, result.Embeddings, 3)
	})

	t.Run("handles API error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			response := map[string]interface{}{
				"error": map[string]interface{}{
					"message": "Incorrect API key provided",
					"type":    "invalid_request_error",
					"code":    "invalid_api_key",
				},
			}
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		}))
		defer server.Close()

		provider := &openAIEmbeddingProvider{
			config: OpenAIConfig{
				APIKey:  "invalid-key",
				BaseURL: server.URL,
			},
			httpClient: server.Client(),
		}

		_, err := provider.Embed(context.Background(), []string{"test"}, "model")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Incorrect API key")
	})
}

func TestOpenAIEmbeddingProvider_SupportedModels(t *testing.T) {
	provider := &openAIEmbeddingProvider{}
	models := provider.SupportedModels()
	assert.NotEmpty(t, models)
	assert.Equal(t, OpenAIEmbeddingModels, models)
}

func TestOpenAIEmbeddingProvider_DefaultModel(t *testing.T) {
	t.Run("returns configured model", func(t *testing.T) {
		provider := &openAIEmbeddingProvider{
			config: OpenAIConfig{Model: "custom-model"},
		}
		assert.Equal(t, "custom-model", provider.DefaultModel())
	})

	t.Run("returns default when not configured", func(t *testing.T) {
		provider := &openAIEmbeddingProvider{}
		assert.Equal(t, "text-embedding-3-small", provider.DefaultModel())
	})
}

func TestOpenAIEmbeddingProvider_ValidateConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		provider := &openAIEmbeddingProvider{
			config: OpenAIConfig{APIKey: "key"},
		}
		assert.NoError(t, provider.ValidateConfig())
	})

	t.Run("missing API key", func(t *testing.T) {
		provider := &openAIEmbeddingProvider{
			config: OpenAIConfig{},
		}
		err := provider.ValidateConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "api_key is required")
	})
}

func TestOpenAIEmbeddingResponse_Struct(t *testing.T) {
	t.Run("unmarshals correctly", func(t *testing.T) {
		jsonData := `{
			"object": "list",
			"model": "text-embedding-3-small",
			"data": [
				{"object": "embedding", "index": 0, "embedding": [0.1, 0.2, 0.3]}
			],
			"usage": {"prompt_tokens": 5, "total_tokens": 5}
		}`

		var resp openAIEmbeddingResponse
		err := json.Unmarshal([]byte(jsonData), &resp)
		require.NoError(t, err)
		assert.Equal(t, "list", resp.Object)
		assert.Equal(t, "text-embedding-3-small", resp.Model)
		assert.Len(t, resp.Data, 1)
		assert.Equal(t, 0, resp.Data[0].Index)
		assert.Equal(t, []float32{0.1, 0.2, 0.3}, resp.Data[0].Embedding)
		assert.Equal(t, 5, resp.Usage.PromptTokens)
	})
}
