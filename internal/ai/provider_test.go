package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// NewProvider Tests
// =============================================================================

func TestNewProvider(t *testing.T) {
	t.Run("creates OpenAI provider with valid config", func(t *testing.T) {
		config := ProviderConfig{
			Name: "openai-test",
			Type: ProviderTypeOpenAI,
			Config: map[string]string{
				"api_key": "test-key",
			},
		}

		provider, err := NewProvider(config)
		require.NoError(t, err)
		require.NotNil(t, provider)
		assert.Equal(t, "openai-test", provider.Name())
		assert.Equal(t, ProviderTypeOpenAI, provider.Type())
	})

	t.Run("creates Azure provider with valid config", func(t *testing.T) {
		config := ProviderConfig{
			Name: "azure-test",
			Type: ProviderTypeAzure,
			Config: map[string]string{
				"api_key":         "test-key",
				"endpoint":        "https://example.openai.azure.com",
				"deployment_name": "my-deployment",
			},
		}

		provider, err := NewProvider(config)
		require.NoError(t, err)
		require.NotNil(t, provider)
		assert.Equal(t, "azure-test", provider.Name())
		assert.Equal(t, ProviderTypeAzure, provider.Type())
	})

	t.Run("creates Ollama provider with valid config", func(t *testing.T) {
		config := ProviderConfig{
			Name:  "ollama-test",
			Type:  ProviderTypeOllama,
			Model: "llama2",
			Config: map[string]string{
				"endpoint": "http://localhost:11434",
			},
		}

		provider, err := NewProvider(config)
		require.NoError(t, err)
		require.NotNil(t, provider)
		assert.Equal(t, "ollama-test", provider.Name())
		assert.Equal(t, ProviderTypeOllama, provider.Type())
	})

	t.Run("returns error for unsupported provider type", func(t *testing.T) {
		config := ProviderConfig{
			Name: "test",
			Type: "unsupported-provider",
		}

		provider, err := NewProvider(config)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "unsupported provider type")
	})

	t.Run("returns error for empty provider type", func(t *testing.T) {
		config := ProviderConfig{
			Name: "test",
			Type: "",
		}

		provider, err := NewProvider(config)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "unsupported provider type")
	})
}

// =============================================================================
// NewOpenAIProvider Tests
// =============================================================================

func TestNewOpenAIProvider(t *testing.T) {
	t.Run("creates provider with api_key", func(t *testing.T) {
		config := ProviderConfig{
			Name: "openai-test",
			Type: ProviderTypeOpenAI,
			Config: map[string]string{
				"api_key": "sk-test-key",
			},
		}

		provider, err := NewOpenAIProvider(config)
		require.NoError(t, err)
		require.NotNil(t, provider)
		assert.Equal(t, "openai-test", provider.Name())
		assert.Equal(t, ProviderTypeOpenAI, provider.Type())
	})

	t.Run("returns error without api_key", func(t *testing.T) {
		config := ProviderConfig{
			Name:   "openai-test",
			Type:   ProviderTypeOpenAI,
			Config: map[string]string{},
		}

		provider, err := NewOpenAIProvider(config)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "api_key is required")
	})

	t.Run("returns error with empty api_key", func(t *testing.T) {
		config := ProviderConfig{
			Name: "openai-test",
			Type: ProviderTypeOpenAI,
			Config: map[string]string{
				"api_key": "",
			},
		}

		provider, err := NewOpenAIProvider(config)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "api_key is required")
	})

	t.Run("defaults model to gpt-4-turbo when not specified", func(t *testing.T) {
		config := ProviderConfig{
			Name: "openai-test",
			Type: ProviderTypeOpenAI,
			Config: map[string]string{
				"api_key": "sk-test-key",
			},
		}

		provider, err := NewOpenAIProvider(config)
		require.NoError(t, err)
		require.NotNil(t, provider)

		// ValidateConfig should pass since model defaults
		err = provider.ValidateConfig()
		require.NoError(t, err)
	})

	t.Run("uses custom model when specified", func(t *testing.T) {
		config := ProviderConfig{
			Name:  "openai-test",
			Type:  ProviderTypeOpenAI,
			Model: "gpt-3.5-turbo",
			Config: map[string]string{
				"api_key": "sk-test-key",
			},
		}

		provider, err := NewOpenAIProvider(config)
		require.NoError(t, err)
		require.NotNil(t, provider)
	})

	t.Run("accepts optional organization_id", func(t *testing.T) {
		config := ProviderConfig{
			Name: "openai-test",
			Type: ProviderTypeOpenAI,
			Config: map[string]string{
				"api_key":         "sk-test-key",
				"organization_id": "org-12345",
			},
		}

		provider, err := NewOpenAIProvider(config)
		require.NoError(t, err)
		require.NotNil(t, provider)
	})

	t.Run("accepts optional base_url", func(t *testing.T) {
		config := ProviderConfig{
			Name: "openai-test",
			Type: ProviderTypeOpenAI,
			Config: map[string]string{
				"api_key":  "sk-test-key",
				"base_url": "https://custom-openai.example.com/v1",
			},
		}

		provider, err := NewOpenAIProvider(config)
		require.NoError(t, err)
		require.NotNil(t, provider)
	})

	t.Run("handles nil Config map", func(t *testing.T) {
		config := ProviderConfig{
			Name:   "openai-test",
			Type:   ProviderTypeOpenAI,
			Config: nil,
		}

		provider, err := NewOpenAIProvider(config)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "api_key is required")
	})
}

// =============================================================================
// NewAzureProvider Tests
// =============================================================================

func TestNewAzureProvider(t *testing.T) {
	t.Run("creates provider with valid config", func(t *testing.T) {
		config := ProviderConfig{
			Name: "azure-test",
			Type: ProviderTypeAzure,
			Config: map[string]string{
				"api_key":         "test-key",
				"endpoint":        "https://example.openai.azure.com",
				"deployment_name": "my-deployment",
			},
		}

		provider, err := NewAzureProvider(config)
		require.NoError(t, err)
		require.NotNil(t, provider)
		assert.Equal(t, "azure-test", provider.Name())
		assert.Equal(t, ProviderTypeAzure, provider.Type())
	})

	t.Run("returns error without api_key", func(t *testing.T) {
		config := ProviderConfig{
			Name: "azure-test",
			Type: ProviderTypeAzure,
			Config: map[string]string{
				"endpoint":        "https://example.openai.azure.com",
				"deployment_name": "my-deployment",
			},
		}

		provider, err := NewAzureProvider(config)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "api_key is required")
	})

	t.Run("returns error without endpoint", func(t *testing.T) {
		config := ProviderConfig{
			Name: "azure-test",
			Type: ProviderTypeAzure,
			Config: map[string]string{
				"api_key":         "test-key",
				"deployment_name": "my-deployment",
			},
		}

		provider, err := NewAzureProvider(config)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "endpoint is required")
	})

	t.Run("returns error without deployment_name", func(t *testing.T) {
		config := ProviderConfig{
			Name: "azure-test",
			Type: ProviderTypeAzure,
			Config: map[string]string{
				"api_key":  "test-key",
				"endpoint": "https://example.openai.azure.com",
			},
		}

		provider, err := NewAzureProvider(config)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "deployment_name is required")
	})

	t.Run("defaults api_version when not specified", func(t *testing.T) {
		config := ProviderConfig{
			Name: "azure-test",
			Type: ProviderTypeAzure,
			Config: map[string]string{
				"api_key":         "test-key",
				"endpoint":        "https://example.openai.azure.com",
				"deployment_name": "my-deployment",
			},
		}

		provider, err := NewAzureProvider(config)
		require.NoError(t, err)
		require.NotNil(t, provider)

		// Provider should be valid with default api_version
		err = provider.ValidateConfig()
		require.NoError(t, err)
	})

	t.Run("accepts custom api_version", func(t *testing.T) {
		config := ProviderConfig{
			Name: "azure-test",
			Type: ProviderTypeAzure,
			Config: map[string]string{
				"api_key":         "test-key",
				"endpoint":        "https://example.openai.azure.com",
				"deployment_name": "my-deployment",
				"api_version":     "2023-05-15",
			},
		}

		provider, err := NewAzureProvider(config)
		require.NoError(t, err)
		require.NotNil(t, provider)
	})

	t.Run("handles nil Config map", func(t *testing.T) {
		config := ProviderConfig{
			Name:   "azure-test",
			Type:   ProviderTypeAzure,
			Config: nil,
		}

		provider, err := NewAzureProvider(config)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "api_key is required")
	})

	t.Run("returns error with empty api_key", func(t *testing.T) {
		config := ProviderConfig{
			Name: "azure-test",
			Type: ProviderTypeAzure,
			Config: map[string]string{
				"api_key":         "",
				"endpoint":        "https://example.openai.azure.com",
				"deployment_name": "my-deployment",
			},
		}

		provider, err := NewAzureProvider(config)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "api_key is required")
	})

	t.Run("returns error with empty endpoint", func(t *testing.T) {
		config := ProviderConfig{
			Name: "azure-test",
			Type: ProviderTypeAzure,
			Config: map[string]string{
				"api_key":         "test-key",
				"endpoint":        "",
				"deployment_name": "my-deployment",
			},
		}

		provider, err := NewAzureProvider(config)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "endpoint is required")
	})

	t.Run("returns error with empty deployment_name", func(t *testing.T) {
		config := ProviderConfig{
			Name: "azure-test",
			Type: ProviderTypeAzure,
			Config: map[string]string{
				"api_key":         "test-key",
				"endpoint":        "https://example.openai.azure.com",
				"deployment_name": "",
			},
		}

		provider, err := NewAzureProvider(config)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "deployment_name is required")
	})
}

// =============================================================================
// NewOllamaProvider Tests
// =============================================================================

func TestNewOllamaProvider(t *testing.T) {
	t.Run("creates provider with model", func(t *testing.T) {
		config := ProviderConfig{
			Name:  "ollama-test",
			Type:  ProviderTypeOllama,
			Model: "llama2",
			Config: map[string]string{
				"endpoint": "http://localhost:11434",
			},
		}

		provider, err := NewOllamaProvider(config)
		require.NoError(t, err)
		require.NotNil(t, provider)
		assert.Equal(t, "ollama-test", provider.Name())
		assert.Equal(t, ProviderTypeOllama, provider.Type())
	})

	t.Run("returns error without model", func(t *testing.T) {
		config := ProviderConfig{
			Name: "ollama-test",
			Type: ProviderTypeOllama,
			Config: map[string]string{
				"endpoint": "http://localhost:11434",
			},
		}

		provider, err := NewOllamaProvider(config)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "model is required")
	})

	t.Run("defaults endpoint to localhost:11434", func(t *testing.T) {
		config := ProviderConfig{
			Name:   "ollama-test",
			Type:   ProviderTypeOllama,
			Model:  "llama2",
			Config: map[string]string{},
		}

		provider, err := NewOllamaProvider(config)
		require.NoError(t, err)
		require.NotNil(t, provider)

		// Provider should be valid with default endpoint
		err = provider.ValidateConfig()
		require.NoError(t, err)
	})

	t.Run("accepts custom endpoint", func(t *testing.T) {
		config := ProviderConfig{
			Name:  "ollama-test",
			Type:  ProviderTypeOllama,
			Model: "llama2",
			Config: map[string]string{
				"endpoint": "http://remote-ollama:11434",
			},
		}

		provider, err := NewOllamaProvider(config)
		require.NoError(t, err)
		require.NotNil(t, provider)
	})

	t.Run("handles nil Config map", func(t *testing.T) {
		config := ProviderConfig{
			Name:   "ollama-test",
			Type:   ProviderTypeOllama,
			Model:  "llama2",
			Config: nil,
		}

		provider, err := NewOllamaProvider(config)
		require.NoError(t, err) // Endpoint defaults to localhost
		require.NotNil(t, provider)
	})

	t.Run("returns error with empty model", func(t *testing.T) {
		config := ProviderConfig{
			Name:  "ollama-test",
			Type:  ProviderTypeOllama,
			Model: "",
			Config: map[string]string{
				"endpoint": "http://localhost:11434",
			},
		}

		provider, err := NewOllamaProvider(config)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "model is required")
	})
}

// =============================================================================
// Provider ValidateConfig Tests
// =============================================================================

func TestOpenAIProvider_ValidateConfig(t *testing.T) {
	t.Run("valid config passes validation", func(t *testing.T) {
		provider, err := newOpenAIProviderInternal("test", OpenAIConfig{
			APIKey: "sk-test-key",
			Model:  "gpt-4",
		})
		require.NoError(t, err)

		err = provider.ValidateConfig()
		require.NoError(t, err)
	})

	t.Run("returns error for missing api_key", func(t *testing.T) {
		provider, err := newOpenAIProviderInternal("test", OpenAIConfig{
			APIKey: "",
			Model:  "gpt-4",
		})
		require.NoError(t, err)

		err = provider.ValidateConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "api_key is required")
	})

	t.Run("returns error for missing model", func(t *testing.T) {
		provider, err := newOpenAIProviderInternal("test", OpenAIConfig{
			APIKey: "sk-test-key",
			Model:  "",
		})
		require.NoError(t, err)

		err = provider.ValidateConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "model is required")
	})

	t.Run("returns error for both missing fields", func(t *testing.T) {
		provider, err := newOpenAIProviderInternal("test", OpenAIConfig{
			APIKey: "",
			Model:  "",
		})
		require.NoError(t, err)

		err = provider.ValidateConfig()
		require.Error(t, err)
		// Should fail on first validation error (api_key)
		assert.Contains(t, err.Error(), "api_key is required")
	})
}

func TestAzureProvider_ValidateConfig(t *testing.T) {
	t.Run("valid config passes validation", func(t *testing.T) {
		provider, err := newAzureProviderInternal("test", AzureConfig{
			APIKey:         "test-key",
			Endpoint:       "https://example.openai.azure.com",
			DeploymentName: "my-deployment",
			APIVersion:     "2024-02-15-preview",
		})
		require.NoError(t, err)

		err = provider.ValidateConfig()
		require.NoError(t, err)
	})

	t.Run("returns error for missing api_key", func(t *testing.T) {
		provider, err := newAzureProviderInternal("test", AzureConfig{
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
		provider, err := newAzureProviderInternal("test", AzureConfig{
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
		provider, err := newAzureProviderInternal("test", AzureConfig{
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

func TestOllamaProvider_ValidateConfig(t *testing.T) {
	t.Run("valid config passes validation", func(t *testing.T) {
		provider, err := newOllamaProviderInternal("test", OllamaConfig{
			Endpoint: "http://localhost:11434",
			Model:    "llama2",
		})
		require.NoError(t, err)

		err = provider.ValidateConfig()
		require.NoError(t, err)
	})

	t.Run("returns error for missing endpoint", func(t *testing.T) {
		provider, err := newOllamaProviderInternal("test", OllamaConfig{
			Endpoint: "",
			Model:    "llama2",
		})
		require.NoError(t, err)

		err = provider.ValidateConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "endpoint is required")
	})

	t.Run("returns error for missing model", func(t *testing.T) {
		provider, err := newOllamaProviderInternal("test", OllamaConfig{
			Endpoint: "http://localhost:11434",
			Model:    "",
		})
		require.NoError(t, err)

		err = provider.ValidateConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "model is required")
	})
}

// =============================================================================
// Provider Struct Tests
// =============================================================================

func TestProviderConfig_Struct(t *testing.T) {
	t.Run("zero value has expected defaults", func(t *testing.T) {
		var config ProviderConfig

		assert.Empty(t, config.Name)
		assert.Empty(t, config.DisplayName)
		assert.Empty(t, config.Type)
		assert.Empty(t, config.Model)
		assert.Nil(t, config.Config)
	})

	t.Run("all fields can be set", func(t *testing.T) {
		config := ProviderConfig{
			Name:        "test-provider",
			DisplayName: "Test Provider",
			Type:        ProviderTypeOpenAI,
			Model:       "gpt-4",
			Config: map[string]string{
				"api_key": "test-key",
			},
		}

		assert.Equal(t, "test-provider", config.Name)
		assert.Equal(t, "Test Provider", config.DisplayName)
		assert.Equal(t, ProviderTypeOpenAI, config.Type)
		assert.Equal(t, "gpt-4", config.Model)
		assert.Equal(t, "test-key", config.Config["api_key"])
	})
}

func TestOpenAIConfig_Struct(t *testing.T) {
	t.Run("zero value has expected defaults", func(t *testing.T) {
		var config OpenAIConfig

		assert.Empty(t, config.APIKey)
		assert.Empty(t, config.Model)
		assert.Empty(t, config.OrganizationID)
		assert.Empty(t, config.BaseURL)
	})

	t.Run("all fields can be set", func(t *testing.T) {
		config := OpenAIConfig{
			APIKey:         "sk-test-key",
			Model:          "gpt-4",
			OrganizationID: "org-12345",
			BaseURL:        "https://custom.example.com/v1",
		}

		assert.Equal(t, "sk-test-key", config.APIKey)
		assert.Equal(t, "gpt-4", config.Model)
		assert.Equal(t, "org-12345", config.OrganizationID)
		assert.Equal(t, "https://custom.example.com/v1", config.BaseURL)
	})
}

func TestAzureConfig_Struct(t *testing.T) {
	t.Run("zero value has expected defaults", func(t *testing.T) {
		var config AzureConfig

		assert.Empty(t, config.APIKey)
		assert.Empty(t, config.Endpoint)
		assert.Empty(t, config.DeploymentName)
		assert.Empty(t, config.APIVersion)
	})

	t.Run("all fields can be set", func(t *testing.T) {
		config := AzureConfig{
			APIKey:         "test-key",
			Endpoint:       "https://example.openai.azure.com",
			DeploymentName: "my-deployment",
			APIVersion:     "2024-02-15-preview",
		}

		assert.Equal(t, "test-key", config.APIKey)
		assert.Equal(t, "https://example.openai.azure.com", config.Endpoint)
		assert.Equal(t, "my-deployment", config.DeploymentName)
		assert.Equal(t, "2024-02-15-preview", config.APIVersion)
	})
}

func TestOllamaConfig_Struct(t *testing.T) {
	t.Run("zero value has expected defaults", func(t *testing.T) {
		var config OllamaConfig

		assert.Empty(t, config.Endpoint)
		assert.Empty(t, config.Model)
	})

	t.Run("all fields can be set", func(t *testing.T) {
		config := OllamaConfig{
			Endpoint: "http://localhost:11434",
			Model:    "llama2",
		}

		assert.Equal(t, "http://localhost:11434", config.Endpoint)
		assert.Equal(t, "llama2", config.Model)
	})
}

// =============================================================================
// Provider Type Constants Tests
// =============================================================================

func TestProviderType_Constants(t *testing.T) {
	t.Run("provider types have expected values", func(t *testing.T) {
		assert.Equal(t, ProviderType("openai"), ProviderTypeOpenAI)
		assert.Equal(t, ProviderType("azure"), ProviderTypeAzure)
		assert.Equal(t, ProviderType("ollama"), ProviderTypeOllama)
	})

	t.Run("provider types are unique", func(t *testing.T) {
		types := []ProviderType{
			ProviderTypeOpenAI,
			ProviderTypeAzure,
			ProviderTypeOllama,
		}

		seen := make(map[ProviderType]bool)
		for _, pt := range types {
			assert.False(t, seen[pt], "duplicate provider type: %s", pt)
			seen[pt] = true
		}
	})
}

// =============================================================================
// Role Constants Tests
// =============================================================================

func TestRole_Constants(t *testing.T) {
	t.Run("roles have expected values", func(t *testing.T) {
		assert.Equal(t, Role("system"), RoleSystem)
		assert.Equal(t, Role("user"), RoleUser)
		assert.Equal(t, Role("assistant"), RoleAssistant)
		assert.Equal(t, Role("tool"), RoleTool)
	})

	t.Run("roles are unique", func(t *testing.T) {
		roles := []Role{
			RoleSystem,
			RoleUser,
			RoleAssistant,
			RoleTool,
		}

		seen := make(map[Role]bool)
		for _, r := range roles {
			assert.False(t, seen[r], "duplicate role: %s", r)
			seen[r] = true
		}
	})
}

// =============================================================================
// Provider Interface Tests
// =============================================================================

func TestProvider_Interface(t *testing.T) {
	t.Run("openAIProvider implements Provider interface", func(t *testing.T) {
		provider, err := newOpenAIProviderInternal("test", OpenAIConfig{
			APIKey: "test-key",
			Model:  "gpt-4",
		})
		require.NoError(t, err)

		var _ Provider = provider // Compile-time check
	})

	t.Run("azureProvider implements Provider interface", func(t *testing.T) {
		provider, err := newAzureProviderInternal("test", AzureConfig{
			APIKey:         "test-key",
			Endpoint:       "https://example.openai.azure.com",
			DeploymentName: "my-deployment",
		})
		require.NoError(t, err)

		var _ Provider = provider // Compile-time check
	})

	t.Run("ollamaProvider implements Provider interface", func(t *testing.T) {
		provider, err := newOllamaProviderInternal("test", OllamaConfig{
			Endpoint: "http://localhost:11434",
			Model:    "llama2",
		})
		require.NoError(t, err)

		var _ Provider = provider // Compile-time check
	})
}

// =============================================================================
// Provider Close Tests
// =============================================================================

func TestProvider_Close(t *testing.T) {
	t.Run("OpenAI provider Close returns nil", func(t *testing.T) {
		provider, err := newOpenAIProviderInternal("test", OpenAIConfig{
			APIKey: "test-key",
			Model:  "gpt-4",
		})
		require.NoError(t, err)

		err = provider.Close()
		assert.NoError(t, err)
	})

	t.Run("Azure provider Close returns nil", func(t *testing.T) {
		provider, err := newAzureProviderInternal("test", AzureConfig{
			APIKey:         "test-key",
			Endpoint:       "https://example.openai.azure.com",
			DeploymentName: "my-deployment",
		})
		require.NoError(t, err)

		err = provider.Close()
		assert.NoError(t, err)
	})

	t.Run("Ollama provider Close returns nil", func(t *testing.T) {
		provider, err := newOllamaProviderInternal("test", OllamaConfig{
			Endpoint: "http://localhost:11434",
			Model:    "llama2",
		})
		require.NoError(t, err)

		err = provider.Close()
		assert.NoError(t, err)
	})
}

// =============================================================================
// ExecuteSQLTool Tests
// =============================================================================

func TestExecuteSQLTool(t *testing.T) {
	t.Run("has expected type", func(t *testing.T) {
		assert.Equal(t, "function", ExecuteSQLTool.Type)
	})

	t.Run("has expected function name", func(t *testing.T) {
		assert.Equal(t, "execute_sql", ExecuteSQLTool.Function.Name)
	})

	t.Run("has description", func(t *testing.T) {
		assert.NotEmpty(t, ExecuteSQLTool.Function.Description)
		assert.Contains(t, ExecuteSQLTool.Function.Description, "SQL")
	})

	t.Run("has required parameters", func(t *testing.T) {
		params := ExecuteSQLTool.Function.Parameters
		assert.NotNil(t, params)

		// Check type
		assert.Equal(t, "object", params["type"])

		// Check properties exist
		props, ok := params["properties"].(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, props, "sql")
		assert.Contains(t, props, "description")

		// Check required
		required, ok := params["required"].([]string)
		require.True(t, ok)
		assert.Contains(t, required, "sql")
		assert.Contains(t, required, "description")
	})
}

// =============================================================================
// ReadCloserWrapper Tests
// =============================================================================

func TestReadCloserWrapper(t *testing.T) {
	t.Run("Close returns nil", func(t *testing.T) {
		wrapper := &ReadCloserWrapper{}
		err := wrapper.Close()
		assert.NoError(t, err)
	})
}
