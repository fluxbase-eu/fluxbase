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

func TestNewOllamaEmbeddingProviderInternal(t *testing.T) {
	t.Run("creates provider with config", func(t *testing.T) {
		config := OllamaConfig{
			Endpoint: "http://localhost:11434/",
			Model:    "nomic-embed-text",
		}

		provider, err := newOllamaEmbeddingProviderInternal(config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		// Trailing slash should be trimmed
		assert.Equal(t, "http://localhost:11434", provider.config.Endpoint)
		assert.Equal(t, "nomic-embed-text", provider.config.Model)
	})
}

func TestOllamaEmbeddingProvider_Embed(t *testing.T) {
	t.Run("returns error for empty texts", func(t *testing.T) {
		provider := &ollamaEmbeddingProvider{
			config:     OllamaConfig{},
			httpClient: http.DefaultClient,
		}

		_, err := provider.Embed(context.Background(), []string{}, "model")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no texts provided")
	})

	t.Run("uses config model when not specified", func(t *testing.T) {
		var receivedModel string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var reqBody map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&reqBody)
			require.NoError(t, err)
			receivedModel = reqBody["model"].(string)

			response := ollamaEmbeddingResponse{
				Embedding: []float64{0.1, 0.2, 0.3},
			}
			err = json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		}))
		defer server.Close()

		provider := &ollamaEmbeddingProvider{
			config: OllamaConfig{
				Endpoint: server.URL,
				Model:    "nomic-embed-text",
			},
			httpClient: server.Client(),
		}

		_, err := provider.Embed(context.Background(), []string{"test"}, "")
		require.NoError(t, err)
		assert.Equal(t, "nomic-embed-text", receivedModel)
	})

	t.Run("successful single text embedding", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/api/embeddings", r.URL.Path)

			var reqBody map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&reqBody)
			require.NoError(t, err)
			assert.Equal(t, "test model", reqBody["model"])
			assert.Equal(t, "test text", reqBody["prompt"])

			response := ollamaEmbeddingResponse{
				Embedding: []float64{0.1, 0.2, 0.3, 0.4},
			}
			err = json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		}))
		defer server.Close()

		provider := &ollamaEmbeddingProvider{
			config: OllamaConfig{
				Endpoint: server.URL,
				Model:    "nomic-embed-text",
			},
			httpClient: server.Client(),
		}

		result, err := provider.Embed(context.Background(), []string{"test text"}, "test model")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Embeddings, 1)
		assert.Len(t, result.Embeddings[0], 4)
		assert.Equal(t, float32(0.1), result.Embeddings[0][0])
		assert.Equal(t, 4, result.Dimensions)
		assert.Equal(t, "test model", result.Model)
	})

	t.Run("handles multiple texts with sequential requests", func(t *testing.T) {
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			response := ollamaEmbeddingResponse{
				Embedding: []float64{float64(requestCount) * 0.1},
			}
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		}))
		defer server.Close()

		provider := &ollamaEmbeddingProvider{
			config: OllamaConfig{
				Endpoint: server.URL,
				Model:    "model",
			},
			httpClient: server.Client(),
		}

		result, err := provider.Embed(context.Background(), []string{"text1", "text2", "text3"}, "model")
		require.NoError(t, err)
		assert.Len(t, result.Embeddings, 3)
		assert.Equal(t, 3, requestCount) // One request per text
	})

	t.Run("handles API error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			response := map[string]string{"error": "Model not found"}
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		}))
		defer server.Close()

		provider := &ollamaEmbeddingProvider{
			config: OllamaConfig{
				Endpoint: server.URL,
				Model:    "invalid-model",
			},
			httpClient: server.Client(),
		}

		_, err := provider.Embed(context.Background(), []string{"test"}, "invalid-model")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Model not found")
	})

	t.Run("estimates tokens based on text length", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := ollamaEmbeddingResponse{
				Embedding: []float64{0.1},
			}
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		}))
		defer server.Close()

		provider := &ollamaEmbeddingProvider{
			config: OllamaConfig{
				Endpoint: server.URL,
				Model:    "model",
			},
			httpClient: server.Client(),
		}

		// Text with 100 characters should have ~25 estimated tokens (100/4)
		longText := "This is a longer text that should result in a higher estimated token count for testing purposes..."
		result, err := provider.Embed(context.Background(), []string{longText}, "model")
		require.NoError(t, err)
		assert.Greater(t, result.Usage.PromptTokens, 0)
	})
}

func TestOllamaEmbeddingProvider_SupportedModels(t *testing.T) {
	provider := &ollamaEmbeddingProvider{}
	models := provider.SupportedModels()
	assert.NotEmpty(t, models)
	assert.Equal(t, OllamaEmbeddingModels, models)
}

func TestOllamaEmbeddingProvider_DefaultModel(t *testing.T) {
	t.Run("returns configured model", func(t *testing.T) {
		provider := &ollamaEmbeddingProvider{
			config: OllamaConfig{Model: "custom-embed-model"},
		}
		assert.Equal(t, "custom-embed-model", provider.DefaultModel())
	})

	t.Run("returns default when not configured", func(t *testing.T) {
		provider := &ollamaEmbeddingProvider{}
		assert.Equal(t, "nomic-embed-text", provider.DefaultModel())
	})
}

func TestOllamaEmbeddingProvider_ValidateConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		provider := &ollamaEmbeddingProvider{
			config: OllamaConfig{Model: "nomic-embed-text"},
		}
		assert.NoError(t, provider.ValidateConfig())
	})

	t.Run("missing model", func(t *testing.T) {
		provider := &ollamaEmbeddingProvider{
			config: OllamaConfig{},
		}
		err := provider.ValidateConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "model is required")
	})
}

func TestOllamaEmbeddingResponse_Struct(t *testing.T) {
	t.Run("unmarshals correctly", func(t *testing.T) {
		jsonData := `{"embedding": [0.1, 0.2, 0.3, 0.4, 0.5]}`

		var resp ollamaEmbeddingResponse
		err := json.Unmarshal([]byte(jsonData), &resp)
		require.NoError(t, err)
		assert.Len(t, resp.Embedding, 5)
		assert.Equal(t, 0.1, resp.Embedding[0])
	})
}

func TestOllamaEmbeddingProvider_Float64ToFloat32Conversion(t *testing.T) {
	t.Run("converts float64 response to float32", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Ollama returns float64 values
			response := ollamaEmbeddingResponse{
				Embedding: []float64{0.123456789, 0.987654321},
			}
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		}))
		defer server.Close()

		provider := &ollamaEmbeddingProvider{
			config: OllamaConfig{
				Endpoint: server.URL,
				Model:    "model",
			},
			httpClient: server.Client(),
		}

		result, err := provider.Embed(context.Background(), []string{"test"}, "model")
		require.NoError(t, err)
		// Verify float32 type (within precision limits)
		assert.InDelta(t, 0.123456789, float64(result.Embeddings[0][0]), 0.0001)
		assert.InDelta(t, 0.987654321, float64(result.Embeddings[0][1]), 0.0001)
	})
}
