package ai

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockEmbeddingProvider implements EmbeddingProvider for testing
type mockEmbeddingProvider struct {
	embedFunc       func(ctx context.Context, texts []string, model string) (*EmbeddingResponse, error)
	supportedModels []EmbeddingModel
	defaultModel    string
	configValid     bool
}

func newMockEmbeddingProvider() *mockEmbeddingProvider {
	return &mockEmbeddingProvider{
		supportedModels: []EmbeddingModel{
			{Name: "test-model", Dimensions: 1536, MaxTokens: 8191},
		},
		defaultModel: "test-model",
		configValid:  true,
	}
}

func (m *mockEmbeddingProvider) Embed(ctx context.Context, texts []string, model string) (*EmbeddingResponse, error) {
	if m.embedFunc != nil {
		return m.embedFunc(ctx, texts, model)
	}
	embeddings := make([][]float32, len(texts))
	for i := range texts {
		embeddings[i] = make([]float32, 1536)
		for j := range embeddings[i] {
			embeddings[i][j] = float32(j) * 0.001
		}
	}
	return &EmbeddingResponse{
		Embeddings: embeddings,
		Model:      model,
		Dimensions: 1536,
		Usage: &EmbeddingUsage{
			PromptTokens: len(texts) * 10,
			TotalTokens:  len(texts) * 10,
		},
	}, nil
}

func (m *mockEmbeddingProvider) SupportedModels() []EmbeddingModel {
	return m.supportedModels
}

func (m *mockEmbeddingProvider) DefaultModel() string {
	return m.defaultModel
}

func (m *mockEmbeddingProvider) ValidateConfig() error {
	if !m.configValid {
		return assert.AnError
	}
	return nil
}

func TestNewEmbeddingService(t *testing.T) {
	t.Run("creates service with valid Ollama config", func(t *testing.T) {
		cfg := EmbeddingServiceConfig{
			Provider: ProviderConfig{
				Type:   ProviderTypeOllama,
				Model:  "nomic-embed-text",
				Config: map[string]string{},
			},
			DefaultModel: "nomic-embed-text",
		}

		service, err := NewEmbeddingService(cfg)
		require.NoError(t, err)
		assert.NotNil(t, service)
		assert.Equal(t, "nomic-embed-text", service.DefaultModel())
	})

	t.Run("uses provider default model when not specified", func(t *testing.T) {
		cfg := EmbeddingServiceConfig{
			Provider: ProviderConfig{
				Type:   ProviderTypeOllama,
				Config: map[string]string{},
			},
		}

		service, err := NewEmbeddingService(cfg)
		require.NoError(t, err)
		assert.NotNil(t, service)
		// Ollama defaults to nomic-embed-text
		assert.Equal(t, "nomic-embed-text", service.DefaultModel())
	})

	t.Run("creates service with rate limiting", func(t *testing.T) {
		cfg := EmbeddingServiceConfig{
			Provider: ProviderConfig{
				Type:   ProviderTypeOllama,
				Config: map[string]string{},
			},
			RateLimitRPM: 60,
		}

		service, err := NewEmbeddingService(cfg)
		require.NoError(t, err)
		assert.NotNil(t, service)
		assert.NotNil(t, service.rateLimiter)
	})

	t.Run("creates service with caching", func(t *testing.T) {
		cfg := EmbeddingServiceConfig{
			Provider: ProviderConfig{
				Type:   ProviderTypeOllama,
				Config: map[string]string{},
			},
			CacheEnabled: true,
			CacheTTL:     10 * time.Minute,
		}

		service, err := NewEmbeddingService(cfg)
		require.NoError(t, err)
		assert.NotNil(t, service)
		assert.True(t, service.cacheEnabled)
		assert.Equal(t, 10*time.Minute, service.cacheTTL)
	})

	t.Run("defaults cache TTL to 15 minutes", func(t *testing.T) {
		cfg := EmbeddingServiceConfig{
			Provider: ProviderConfig{
				Type:   ProviderTypeOllama,
				Config: map[string]string{},
			},
			CacheEnabled: true,
		}

		service, err := NewEmbeddingService(cfg)
		require.NoError(t, err)
		assert.Equal(t, 15*time.Minute, service.cacheTTL)
	})

	t.Run("errors on invalid provider config", func(t *testing.T) {
		cfg := EmbeddingServiceConfig{
			Provider: ProviderConfig{
				Type: "invalid-provider",
			},
		}

		service, err := NewEmbeddingService(cfg)
		require.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "unsupported embedding provider type")
	})
}

func TestEmbeddingService_Embed(t *testing.T) {
	createTestService := func() *EmbeddingService {
		return &EmbeddingService{
			provider:     newMockEmbeddingProvider(),
			defaultModel: "test-model",
			cacheResults: make(map[string]*cachedEmbedding),
		}
	}

	t.Run("returns embeddings for valid texts", func(t *testing.T) {
		service := createTestService()
		ctx := context.Background()

		resp, err := service.Embed(ctx, []string{"Hello world"}, "test-model")
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Embeddings, 1)
		assert.Equal(t, "test-model", resp.Model)
		assert.Equal(t, 1536, resp.Dimensions)
	})

	t.Run("returns embeddings for multiple texts", func(t *testing.T) {
		service := createTestService()
		ctx := context.Background()

		resp, err := service.Embed(ctx, []string{"Hello", "World", "Test"}, "test-model")
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Embeddings, 3)
	})

	t.Run("uses default model when not specified", func(t *testing.T) {
		service := createTestService()
		mock := service.provider.(*mockEmbeddingProvider)
		var capturedModel string
		mock.embedFunc = func(ctx context.Context, texts []string, model string) (*EmbeddingResponse, error) {
			capturedModel = model
			return &EmbeddingResponse{
				Embeddings: make([][]float32, len(texts)),
				Model:      model,
				Dimensions: 1536,
			}, nil
		}

		ctx := context.Background()
		_, err := service.Embed(ctx, []string{"test"}, "")
		require.NoError(t, err)
		assert.Equal(t, "test-model", capturedModel)
	})

	t.Run("errors on empty texts", func(t *testing.T) {
		service := createTestService()
		ctx := context.Background()

		resp, err := service.Embed(ctx, []string{}, "test-model")
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "no texts provided")
	})

	t.Run("respects rate limiting", func(t *testing.T) {
		service := createTestService()
		service.rateLimiter = &embeddingRateLimiter{
			tokens:    0, // Exhausted
			maxTokens: 1,
			lastReset: time.Now(),
			window:    time.Minute,
		}

		ctx := context.Background()
		resp, err := service.Embed(ctx, []string{"test"}, "test-model")
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "rate limit exceeded")
	})

	t.Run("returns cached embeddings when available", func(t *testing.T) {
		service := createTestService()
		service.cacheEnabled = true
		service.cacheTTL = 15 * time.Minute

		// Pre-populate cache
		cacheKey := service.cacheKey("cached text", "test-model")
		cachedEmb := make([]float32, 1536)
		cachedEmb[0] = 0.999
		service.addToCache(cacheKey, cachedEmb)

		ctx := context.Background()
		resp, err := service.Embed(ctx, []string{"cached text"}, "test-model")
		require.NoError(t, err)
		assert.Equal(t, float32(0.999), resp.Embeddings[0][0])
	})

	t.Run("handles partial cache hits", func(t *testing.T) {
		service := createTestService()
		service.cacheEnabled = true
		service.cacheTTL = 15 * time.Minute

		// Pre-populate cache for only one text
		cacheKey := service.cacheKey("cached", "test-model")
		cachedEmb := make([]float32, 1536)
		cachedEmb[0] = 0.999
		service.addToCache(cacheKey, cachedEmb)

		ctx := context.Background()
		resp, err := service.Embed(ctx, []string{"cached", "not cached"}, "test-model")
		require.NoError(t, err)
		assert.Len(t, resp.Embeddings, 2)
		// First embedding should be from cache
		assert.Equal(t, float32(0.999), resp.Embeddings[0][0])
	})
}

func TestEmbeddingService_EmbedSingle(t *testing.T) {
	createTestService := func() *EmbeddingService {
		return &EmbeddingService{
			provider:     newMockEmbeddingProvider(),
			defaultModel: "test-model",
			cacheResults: make(map[string]*cachedEmbedding),
		}
	}

	t.Run("returns single embedding", func(t *testing.T) {
		service := createTestService()
		ctx := context.Background()

		embedding, err := service.EmbedSingle(ctx, "Hello world", "test-model")
		require.NoError(t, err)
		assert.Len(t, embedding, 1536)
	})

	t.Run("uses default model when not specified", func(t *testing.T) {
		service := createTestService()
		ctx := context.Background()

		embedding, err := service.EmbedSingle(ctx, "Hello world", "")
		require.NoError(t, err)
		assert.NotNil(t, embedding)
	})
}

func TestEmbeddingService_GenerateEmbedding(t *testing.T) {
	t.Run("generates embedding using default model", func(t *testing.T) {
		service := &EmbeddingService{
			provider:     newMockEmbeddingProvider(),
			defaultModel: "test-model",
			cacheResults: make(map[string]*cachedEmbedding),
		}

		ctx := context.Background()
		embedding, err := service.GenerateEmbedding(ctx, "Hello world")
		require.NoError(t, err)
		assert.Len(t, embedding, 1536)
	})
}

func TestEmbeddingService_SupportedModels(t *testing.T) {
	t.Run("returns provider's supported models", func(t *testing.T) {
		mock := newMockEmbeddingProvider()
		mock.supportedModels = []EmbeddingModel{
			{Name: "model-a", Dimensions: 512, MaxTokens: 4096},
			{Name: "model-b", Dimensions: 768, MaxTokens: 8192},
		}

		service := &EmbeddingService{
			provider:     mock,
			defaultModel: "model-a",
		}

		models := service.SupportedModels()
		assert.Len(t, models, 2)
		assert.Equal(t, "model-a", models[0].Name)
		assert.Equal(t, "model-b", models[1].Name)
	})
}

func TestEmbeddingService_SetProvider(t *testing.T) {
	t.Run("updates provider successfully", func(t *testing.T) {
		service := &EmbeddingService{
			provider:     newMockEmbeddingProvider(),
			defaultModel: "old-model",
		}

		err := service.SetProvider(ProviderConfig{
			Type:   ProviderTypeOllama,
			Model:  "new-model",
			Config: map[string]string{},
		})
		require.NoError(t, err)
		assert.Equal(t, "new-model", service.provider.DefaultModel())
	})

	t.Run("errors on invalid provider config", func(t *testing.T) {
		service := &EmbeddingService{
			provider:     newMockEmbeddingProvider(),
			defaultModel: "old-model",
		}

		err := service.SetProvider(ProviderConfig{
			Type: "invalid",
		})
		require.Error(t, err)
	})
}

func TestEmbeddingService_IsConfigured(t *testing.T) {
	t.Run("returns true when provider is set", func(t *testing.T) {
		service := &EmbeddingService{
			provider: newMockEmbeddingProvider(),
		}

		assert.True(t, service.IsConfigured())
	})

	t.Run("returns false when provider is nil", func(t *testing.T) {
		service := &EmbeddingService{
			provider: nil,
		}

		assert.False(t, service.IsConfigured())
	})
}

func TestEmbeddingService_CacheKey(t *testing.T) {
	service := &EmbeddingService{}

	t.Run("generates unique cache key for text and model", func(t *testing.T) {
		key1 := service.cacheKey("text1", "model-a")
		key2 := service.cacheKey("text2", "model-a")
		key3 := service.cacheKey("text1", "model-b")

		assert.NotEqual(t, key1, key2)
		assert.NotEqual(t, key1, key3)
		assert.Contains(t, key1, "model-a")
		assert.Contains(t, key1, "text1")
	})
}

func TestEmbeddingService_Cache(t *testing.T) {
	t.Run("getFromCache returns nil for missing key", func(t *testing.T) {
		service := &EmbeddingService{
			cacheResults: make(map[string]*cachedEmbedding),
		}

		result := service.getFromCache("nonexistent")
		assert.Nil(t, result)
	})

	t.Run("getFromCache returns nil for expired entry", func(t *testing.T) {
		service := &EmbeddingService{
			cacheResults: make(map[string]*cachedEmbedding),
		}

		// Add expired entry
		service.cacheResults["expired"] = &cachedEmbedding{
			embedding: []float32{0.1, 0.2},
			expiresAt: time.Now().Add(-1 * time.Hour),
		}

		result := service.getFromCache("expired")
		assert.Nil(t, result)
	})

	t.Run("getFromCache returns valid cached entry", func(t *testing.T) {
		service := &EmbeddingService{
			cacheResults: make(map[string]*cachedEmbedding),
		}

		expected := []float32{0.1, 0.2, 0.3}
		service.cacheResults["valid"] = &cachedEmbedding{
			embedding: expected,
			expiresAt: time.Now().Add(1 * time.Hour),
		}

		result := service.getFromCache("valid")
		assert.Equal(t, expected, result)
	})

	t.Run("addToCache stores embedding with TTL", func(t *testing.T) {
		service := &EmbeddingService{
			cacheResults: make(map[string]*cachedEmbedding),
			cacheTTL:     30 * time.Minute,
		}

		embedding := []float32{0.5, 0.6, 0.7}
		service.addToCache("new-key", embedding)

		cached, exists := service.cacheResults["new-key"]
		assert.True(t, exists)
		assert.Equal(t, embedding, cached.embedding)
		assert.True(t, cached.expiresAt.After(time.Now().Add(29*time.Minute)))
	})
}

func TestEmbeddingRateLimiter(t *testing.T) {
	t.Run("allows requests when tokens available", func(t *testing.T) {
		rl := &embeddingRateLimiter{
			tokens:    5,
			maxTokens: 10,
			lastReset: time.Now(),
			window:    time.Minute,
		}

		assert.True(t, rl.allow())
		assert.Equal(t, 4, rl.tokens)
	})

	t.Run("denies requests when no tokens", func(t *testing.T) {
		rl := &embeddingRateLimiter{
			tokens:    0,
			maxTokens: 10,
			lastReset: time.Now(),
			window:    time.Minute,
		}

		assert.False(t, rl.allow())
	})

	t.Run("resets tokens after window expires", func(t *testing.T) {
		rl := &embeddingRateLimiter{
			tokens:    0,
			maxTokens: 10,
			lastReset: time.Now().Add(-2 * time.Minute),
			window:    time.Minute,
		}

		assert.True(t, rl.allow())
		assert.Equal(t, 9, rl.tokens) // 10 - 1 after reset and allow
	})
}

func TestEmbeddingServiceFromConfig(t *testing.T) {
	t.Run("returns error suggesting NewEmbeddingService", func(t *testing.T) {
		service, err := EmbeddingServiceFromConfig(nil)
		require.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "use NewEmbeddingService")
	})
}
