package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmbeddingModels(t *testing.T) {
	t.Run("OpenAI embedding models are defined", func(t *testing.T) {
		assert.NotEmpty(t, OpenAIEmbeddingModels)
		assert.Len(t, OpenAIEmbeddingModels, 3)

		// Check text-embedding-3-small
		found := false
		for _, m := range OpenAIEmbeddingModels {
			if m.Name == "text-embedding-3-small" {
				found = true
				assert.Equal(t, 1536, m.Dimensions)
				assert.Equal(t, 8191, m.MaxTokens)
				break
			}
		}
		assert.True(t, found, "text-embedding-3-small model should be defined")
	})

	t.Run("Azure embedding models are defined", func(t *testing.T) {
		assert.NotEmpty(t, AzureEmbeddingModels)
		assert.Len(t, AzureEmbeddingModels, 3)
	})

	t.Run("Ollama embedding models are defined", func(t *testing.T) {
		assert.NotEmpty(t, OllamaEmbeddingModels)
		assert.Len(t, OllamaEmbeddingModels, 3)

		// Check nomic-embed-text
		found := false
		for _, m := range OllamaEmbeddingModels {
			if m.Name == "nomic-embed-text" {
				found = true
				assert.Equal(t, 768, m.Dimensions)
				break
			}
		}
		assert.True(t, found, "nomic-embed-text model should be defined")
	})
}

func TestGetEmbeddingModelDimensions(t *testing.T) {
	testCases := []struct {
		model      string
		dimensions int
	}{
		{"text-embedding-3-small", 1536},
		{"text-embedding-3-large", 3072},
		{"text-embedding-ada-002", 1536},
		{"nomic-embed-text", 768},
		{"mxbai-embed-large", 1024},
		{"all-minilm", 384},
		{"unknown-model", 0},
		{"", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.model, func(t *testing.T) {
			dims := GetEmbeddingModelDimensions(tc.model)
			assert.Equal(t, tc.dimensions, dims)
		})
	}
}

func TestNewEmbeddingProvider(t *testing.T) {
	t.Run("errors on unsupported provider type", func(t *testing.T) {
		config := ProviderConfig{
			Type: "unsupported-provider",
		}

		provider, err := NewEmbeddingProvider(config)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "unsupported embedding provider type")
	})

	t.Run("errors on OpenAI without api_key", func(t *testing.T) {
		config := ProviderConfig{
			Type:   ProviderTypeOpenAI,
			Config: map[string]string{},
		}

		provider, err := NewEmbeddingProvider(config)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "api_key is required")
	})

	t.Run("errors on Azure without api_key", func(t *testing.T) {
		config := ProviderConfig{
			Type: ProviderTypeAzure,
			Config: map[string]string{
				"endpoint":        "https://example.openai.azure.com",
				"deployment_name": "my-deployment",
			},
		}

		provider, err := NewEmbeddingProvider(config)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "api_key is required")
	})

	t.Run("errors on Azure without endpoint", func(t *testing.T) {
		config := ProviderConfig{
			Type: ProviderTypeAzure,
			Config: map[string]string{
				"api_key":         "test-key",
				"deployment_name": "my-deployment",
			},
		}

		provider, err := NewEmbeddingProvider(config)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "endpoint is required")
	})

	t.Run("errors on Azure without deployment_name", func(t *testing.T) {
		config := ProviderConfig{
			Type: ProviderTypeAzure,
			Config: map[string]string{
				"api_key":  "test-key",
				"endpoint": "https://example.openai.azure.com",
			},
		}

		provider, err := NewEmbeddingProvider(config)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "deployment_name is required")
	})
}

func TestEmbeddingRequest_Struct(t *testing.T) {
	req := EmbeddingRequest{
		Texts: []string{"Hello world", "Another text"},
		Model: "text-embedding-3-small",
	}

	assert.Equal(t, 2, len(req.Texts))
	assert.Equal(t, "text-embedding-3-small", req.Model)
}

func TestEmbeddingResponse_Struct(t *testing.T) {
	resp := EmbeddingResponse{
		Embeddings: [][]float32{{0.1, 0.2, 0.3}, {0.4, 0.5, 0.6}},
		Model:      "text-embedding-3-small",
		Dimensions: 1536,
		Usage: &EmbeddingUsage{
			PromptTokens: 10,
			TotalTokens:  10,
		},
	}

	assert.Equal(t, 2, len(resp.Embeddings))
	assert.Equal(t, "text-embedding-3-small", resp.Model)
	assert.Equal(t, 1536, resp.Dimensions)
	require.NotNil(t, resp.Usage)
	assert.Equal(t, 10, resp.Usage.PromptTokens)
}

func TestEmbeddingModel_Struct(t *testing.T) {
	model := EmbeddingModel{
		Name:       "custom-model",
		Dimensions: 512,
		MaxTokens:  4096,
	}

	assert.Equal(t, "custom-model", model.Name)
	assert.Equal(t, 512, model.Dimensions)
	assert.Equal(t, 4096, model.MaxTokens)
}

// =============================================================================
// Ollama Embedding Provider Tests
// =============================================================================

func TestNewOllamaEmbeddingProvider(t *testing.T) {
	t.Run("creates provider with valid config", func(t *testing.T) {
		config := ProviderConfig{
			Type:  ProviderTypeOllama,
			Model: "nomic-embed-text",
			Config: map[string]string{
				"endpoint": "http://localhost:11434",
			},
		}

		provider, err := NewOllamaEmbeddingProvider(config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
	})

	t.Run("defaults endpoint to localhost:11434", func(t *testing.T) {
		config := ProviderConfig{
			Type:   ProviderTypeOllama,
			Model:  "nomic-embed-text",
			Config: map[string]string{},
		}

		provider, err := NewOllamaEmbeddingProvider(config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
	})

	t.Run("defaults model to nomic-embed-text", func(t *testing.T) {
		config := ProviderConfig{
			Type:   ProviderTypeOllama,
			Config: map[string]string{},
		}

		provider, err := NewOllamaEmbeddingProvider(config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "nomic-embed-text", provider.DefaultModel())
	})

	t.Run("uses custom model when specified", func(t *testing.T) {
		config := ProviderConfig{
			Type:  ProviderTypeOllama,
			Model: "mxbai-embed-large",
			Config: map[string]string{
				"endpoint": "http://localhost:11434",
			},
		}

		provider, err := NewOllamaEmbeddingProvider(config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "mxbai-embed-large", provider.DefaultModel())
	})

	t.Run("handles nil Config map", func(t *testing.T) {
		config := ProviderConfig{
			Type:   ProviderTypeOllama,
			Model:  "nomic-embed-text",
			Config: nil,
		}

		provider, err := NewOllamaEmbeddingProvider(config)
		require.NoError(t, err) // Should default endpoint
		assert.NotNil(t, provider)
	})
}

func TestOllamaEmbeddingProvider_ValidateConfig(t *testing.T) {
	t.Run("valid config passes validation", func(t *testing.T) {
		provider, err := newOllamaEmbeddingProviderInternal(OllamaConfig{
			Endpoint: "http://localhost:11434",
			Model:    "nomic-embed-text",
		})
		require.NoError(t, err)

		err = provider.ValidateConfig()
		require.NoError(t, err)
	})

	t.Run("returns error for missing model", func(t *testing.T) {
		provider, err := newOllamaEmbeddingProviderInternal(OllamaConfig{
			Endpoint: "http://localhost:11434",
			Model:    "",
		})
		require.NoError(t, err)

		err = provider.ValidateConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "model is required")
	})
}

func TestOllamaEmbeddingProvider_SupportedModels(t *testing.T) {
	provider, err := newOllamaEmbeddingProviderInternal(OllamaConfig{
		Endpoint: "http://localhost:11434",
		Model:    "nomic-embed-text",
	})
	require.NoError(t, err)

	models := provider.SupportedModels()
	assert.Equal(t, OllamaEmbeddingModels, models)
	assert.Len(t, models, 3)
}

func TestOllamaEmbeddingProvider_DefaultModel(t *testing.T) {
	t.Run("returns configured model", func(t *testing.T) {
		provider, err := newOllamaEmbeddingProviderInternal(OllamaConfig{
			Endpoint: "http://localhost:11434",
			Model:    "mxbai-embed-large",
		})
		require.NoError(t, err)

		assert.Equal(t, "mxbai-embed-large", provider.DefaultModel())
	})

	t.Run("returns nomic-embed-text when model not configured", func(t *testing.T) {
		provider, err := newOllamaEmbeddingProviderInternal(OllamaConfig{
			Endpoint: "http://localhost:11434",
			Model:    "",
		})
		require.NoError(t, err)

		assert.Equal(t, "nomic-embed-text", provider.DefaultModel())
	})
}

// =============================================================================
// OpenAI Embedding Provider ValidateConfig Tests
// =============================================================================

func TestOpenAIEmbeddingProvider_ValidateConfig(t *testing.T) {
	t.Run("valid config passes validation", func(t *testing.T) {
		provider, err := newOpenAIEmbeddingProviderInternal(OpenAIConfig{
			APIKey: "sk-test-key",
			Model:  "text-embedding-3-small",
		})
		require.NoError(t, err)

		err = provider.ValidateConfig()
		require.NoError(t, err)
	})

	t.Run("returns error for missing api_key", func(t *testing.T) {
		provider, err := newOpenAIEmbeddingProviderInternal(OpenAIConfig{
			APIKey: "",
			Model:  "text-embedding-3-small",
		})
		require.NoError(t, err)

		err = provider.ValidateConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "api_key is required")
	})
}

func TestOpenAIEmbeddingProvider_SupportedModels(t *testing.T) {
	provider, err := newOpenAIEmbeddingProviderInternal(OpenAIConfig{
		APIKey: "sk-test-key",
		Model:  "text-embedding-3-small",
	})
	require.NoError(t, err)

	models := provider.SupportedModels()
	assert.Equal(t, OpenAIEmbeddingModels, models)
	assert.Len(t, models, 3)
}

func TestOpenAIEmbeddingProvider_DefaultModel(t *testing.T) {
	t.Run("returns configured model", func(t *testing.T) {
		provider, err := newOpenAIEmbeddingProviderInternal(OpenAIConfig{
			APIKey: "sk-test-key",
			Model:  "text-embedding-3-large",
		})
		require.NoError(t, err)

		assert.Equal(t, "text-embedding-3-large", provider.DefaultModel())
	})

	t.Run("returns text-embedding-3-small when model not configured", func(t *testing.T) {
		provider, err := newOpenAIEmbeddingProviderInternal(OpenAIConfig{
			APIKey: "sk-test-key",
			Model:  "",
		})
		require.NoError(t, err)

		assert.Equal(t, "text-embedding-3-small", provider.DefaultModel())
	})
}

// =============================================================================
// Azure Embedding Provider ValidateConfig Tests
// =============================================================================

func TestAzureEmbeddingProvider_ValidateConfig(t *testing.T) {
	t.Run("valid config passes validation", func(t *testing.T) {
		provider, err := newAzureEmbeddingProviderInternal(AzureConfig{
			APIKey:         "test-key",
			Endpoint:       "https://example.openai.azure.com",
			DeploymentName: "my-deployment",
		})
		require.NoError(t, err)

		err = provider.ValidateConfig()
		require.NoError(t, err)
	})

	t.Run("returns error for missing api_key", func(t *testing.T) {
		provider, err := newAzureEmbeddingProviderInternal(AzureConfig{
			APIKey:         "",
			Endpoint:       "https://example.openai.azure.com",
			DeploymentName: "my-deployment",
		})
		require.NoError(t, err)

		err = provider.ValidateConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "api_key is required")
	})

	t.Run("returns error for missing endpoint", func(t *testing.T) {
		provider, err := newAzureEmbeddingProviderInternal(AzureConfig{
			APIKey:         "test-key",
			Endpoint:       "",
			DeploymentName: "my-deployment",
		})
		require.NoError(t, err)

		err = provider.ValidateConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "endpoint is required")
	})

	t.Run("returns error for missing deployment_name", func(t *testing.T) {
		provider, err := newAzureEmbeddingProviderInternal(AzureConfig{
			APIKey:         "test-key",
			Endpoint:       "https://example.openai.azure.com",
			DeploymentName: "",
		})
		require.NoError(t, err)

		err = provider.ValidateConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "deployment_name is required")
	})
}

func TestAzureEmbeddingProvider_SupportedModels(t *testing.T) {
	provider, err := newAzureEmbeddingProviderInternal(AzureConfig{
		APIKey:         "test-key",
		Endpoint:       "https://example.openai.azure.com",
		DeploymentName: "my-deployment",
	})
	require.NoError(t, err)

	models := provider.SupportedModels()
	assert.Equal(t, AzureEmbeddingModels, models)
	assert.Len(t, models, 3)
}

func TestAzureEmbeddingProvider_DefaultModel(t *testing.T) {
	provider, err := newAzureEmbeddingProviderInternal(AzureConfig{
		APIKey:         "test-key",
		Endpoint:       "https://example.openai.azure.com",
		DeploymentName: "my-deployment",
	})
	require.NoError(t, err)

	// Azure always returns text-embedding-3-small as default
	assert.Equal(t, "text-embedding-3-small", provider.DefaultModel())
}

// =============================================================================
// EmbeddingUsage Struct Tests
// =============================================================================

func TestEmbeddingUsage_Struct(t *testing.T) {
	t.Run("zero value has expected defaults", func(t *testing.T) {
		var usage EmbeddingUsage

		assert.Zero(t, usage.PromptTokens)
		assert.Zero(t, usage.TotalTokens)
	})

	t.Run("all fields can be set", func(t *testing.T) {
		usage := EmbeddingUsage{
			PromptTokens: 100,
			TotalTokens:  100,
		}

		assert.Equal(t, 100, usage.PromptTokens)
		assert.Equal(t, 100, usage.TotalTokens)
	})
}
