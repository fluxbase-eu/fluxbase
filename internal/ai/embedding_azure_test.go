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

func TestNewAzureEmbeddingProviderInternal(t *testing.T) {
	t.Run("creates provider with config", func(t *testing.T) {
		config := AzureConfig{
			APIKey:         "test-key",
			Endpoint:       "https://test.openai.azure.com/",
			DeploymentName: "test-deployment",
			APIVersion:     "2023-05-15",
		}

		provider, err := newAzureEmbeddingProviderInternal(config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "test-key", provider.config.APIKey)
		// Endpoint should have trailing slash trimmed
		assert.Equal(t, "https://test.openai.azure.com", provider.config.Endpoint)
	})

	t.Run("trims trailing slash from endpoint", func(t *testing.T) {
		config := AzureConfig{
			Endpoint: "https://test.openai.azure.com///",
		}

		provider, err := newAzureEmbeddingProviderInternal(config)
		require.NoError(t, err)
		assert.Equal(t, "https://test.openai.azure.com//", provider.config.Endpoint)
	})
}

func TestAzureEmbeddingProvider_Embed(t *testing.T) {
	t.Run("returns error for empty texts", func(t *testing.T) {
		provider := &azureEmbeddingProvider{
			config:     AzureConfig{},
			httpClient: http.DefaultClient,
		}

		_, err := provider.Embed(context.Background(), []string{}, "model")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no texts provided")
	})

	t.Run("successful embedding request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Contains(t, r.URL.Path, "/openai/deployments/test-deployment/embeddings")
			assert.Equal(t, "test-api-key", r.Header.Get("api-key"))

			response := openAIEmbeddingResponse{
				Object: "list",
				Model:  "text-embedding-3-small",
				Data: []struct {
					Object    string    `json:"object"`
					Index     int       `json:"index"`
					Embedding []float32 `json:"embedding"`
				}{
					{Object: "embedding", Index: 0, Embedding: []float32{0.1, 0.2, 0.3}},
				},
				Usage: struct {
					PromptTokens int `json:"prompt_tokens"`
					TotalTokens  int `json:"total_tokens"`
				}{
					PromptTokens: 5,
					TotalTokens:  5,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		}))
		defer server.Close()

		provider := &azureEmbeddingProvider{
			config: AzureConfig{
				APIKey:         "test-api-key",
				Endpoint:       server.URL,
				DeploymentName: "test-deployment",
				APIVersion:     "2023-05-15",
			},
			httpClient: server.Client(),
		}

		result, err := provider.Embed(context.Background(), []string{"test text"}, "")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Embeddings, 1)
		assert.Equal(t, []float32{0.1, 0.2, 0.3}, result.Embeddings[0])
		assert.Equal(t, 3, result.Dimensions)
		assert.Equal(t, 5, result.Usage.PromptTokens)
	})

	t.Run("handles API error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			response := map[string]interface{}{
				"error": map[string]interface{}{
					"message": "Invalid request",
					"type":    "invalid_request_error",
					"code":    "invalid_api_key",
				},
			}
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		}))
		defer server.Close()

		provider := &azureEmbeddingProvider{
			config: AzureConfig{
				APIKey:         "invalid-key",
				Endpoint:       server.URL,
				DeploymentName: "test-deployment",
				APIVersion:     "2023-05-15",
			},
			httpClient: server.Client(),
		}

		_, err := provider.Embed(context.Background(), []string{"test"}, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid request")
	})
}

func TestAzureEmbeddingProvider_SupportedModels(t *testing.T) {
	provider := &azureEmbeddingProvider{}
	models := provider.SupportedModels()
	assert.NotEmpty(t, models)
	assert.Equal(t, AzureEmbeddingModels, models)
}

func TestAzureEmbeddingProvider_DefaultModel(t *testing.T) {
	provider := &azureEmbeddingProvider{}
	model := provider.DefaultModel()
	assert.Equal(t, "text-embedding-3-small", model)
}

func TestAzureEmbeddingProvider_ValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      AzureConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: AzureConfig{
				APIKey:         "key",
				Endpoint:       "https://test.azure.com",
				DeploymentName: "deployment",
			},
			expectError: false,
		},
		{
			name: "missing API key",
			config: AzureConfig{
				Endpoint:       "https://test.azure.com",
				DeploymentName: "deployment",
			},
			expectError: true,
			errorMsg:    "api_key is required",
		},
		{
			name: "missing endpoint",
			config: AzureConfig{
				APIKey:         "key",
				DeploymentName: "deployment",
			},
			expectError: true,
			errorMsg:    "endpoint is required",
		},
		{
			name: "missing deployment name",
			config: AzureConfig{
				APIKey:   "key",
				Endpoint: "https://test.azure.com",
			},
			expectError: true,
			errorMsg:    "deployment_name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &azureEmbeddingProvider{config: tt.config}
			err := provider.ValidateConfig()
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
