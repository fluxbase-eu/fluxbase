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

func TestNewAzureProviderInternal(t *testing.T) {
	t.Run("creates provider with config", func(t *testing.T) {
		config := AzureConfig{
			APIKey:         "test-key",
			Endpoint:       "https://test.openai.azure.com/",
			DeploymentName: "gpt-4",
			APIVersion:     "2023-05-15",
		}

		provider, err := newAzureProviderInternal("azure-test", config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "azure-test", provider.name)
		assert.Equal(t, "https://test.openai.azure.com", provider.config.Endpoint)
	})
}

func TestAzureProvider_Name(t *testing.T) {
	provider := &azureProvider{name: "my-azure-provider"}
	assert.Equal(t, "my-azure-provider", provider.Name())
}

func TestAzureProvider_Type(t *testing.T) {
	provider := &azureProvider{}
	assert.Equal(t, ProviderTypeAzure, provider.Type())
}

func TestAzureProvider_ValidateConfig(t *testing.T) {
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
				DeploymentName: "gpt-4",
			},
			expectError: false,
		},
		{
			name: "missing API key",
			config: AzureConfig{
				Endpoint:       "https://test.azure.com",
				DeploymentName: "gpt-4",
			},
			expectError: true,
			errorMsg:    "api_key is required",
		},
		{
			name: "missing endpoint",
			config: AzureConfig{
				APIKey:         "key",
				DeploymentName: "gpt-4",
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
			provider := &azureProvider{config: tt.config}
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

func TestAzureProvider_Close(t *testing.T) {
	provider := &azureProvider{
		httpClient: &http.Client{},
	}
	err := provider.Close()
	assert.NoError(t, err)
}

func TestAzureProvider_GetEndpointURL(t *testing.T) {
	provider := &azureProvider{
		config: AzureConfig{
			Endpoint:       "https://test.openai.azure.com",
			DeploymentName: "gpt-4",
			APIVersion:     "2023-05-15",
		},
	}

	url := provider.getEndpointURL()
	assert.Contains(t, url, "https://test.openai.azure.com")
	assert.Contains(t, url, "/openai/deployments/gpt-4/chat/completions")
	assert.Contains(t, url, "api-version=2023-05-15")
}

func TestAzureProvider_Chat(t *testing.T) {
	t.Run("successful chat request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "test-api-key", r.Header.Get("api-key"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			response := openAIResponse{
				ID:    "chatcmpl-123",
				Model: "gpt-4",
				Choices: []openAIChoice{
					{
						Index: 0,
						Message: openAIMessage{
							Role:    "assistant",
							Content: "Hello! How can I help you?",
						},
						FinishReason: "stop",
					},
				},
				Usage: &openAIUsage{
					PromptTokens:     10,
					CompletionTokens: 8,
					TotalTokens:      18,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		}))
		defer server.Close()

		provider := &azureProvider{
			name: "test-azure",
			config: AzureConfig{
				APIKey:         "test-api-key",
				Endpoint:       server.URL,
				DeploymentName: "gpt-4",
				APIVersion:     "2023-05-15",
			},
			httpClient: server.Client(),
		}

		req := &ChatRequest{
			Messages: []Message{
				{Role: RoleUser, Content: "Hello"},
			},
		}

		resp, err := provider.Chat(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "chatcmpl-123", resp.ID)
		assert.Len(t, resp.Choices, 1)
		assert.Equal(t, "Hello! How can I help you?", resp.Choices[0].Message.Content)
		assert.Equal(t, 18, resp.Usage.TotalTokens)
	})

	t.Run("handles API error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := openAIResponse{
				Error: &openAIError{
					Message: "Invalid API key",
					Type:    "invalid_request_error",
					Code:    "invalid_api_key",
				},
			}
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		}))
		defer server.Close()

		provider := &azureProvider{
			config: AzureConfig{
				APIKey:         "invalid-key",
				Endpoint:       server.URL,
				DeploymentName: "gpt-4",
				APIVersion:     "2023-05-15",
			},
			httpClient: server.Client(),
		}

		req := &ChatRequest{
			Messages: []Message{{Role: RoleUser, Content: "Hello"}},
		}

		_, err := provider.Chat(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid API key")
	})
}

func TestAzureProvider_BuildRequest(t *testing.T) {
	provider := &azureProvider{
		config: AzureConfig{
			DeploymentName: "gpt-4",
		},
	}

	t.Run("converts messages correctly", func(t *testing.T) {
		req := &ChatRequest{
			Messages: []Message{
				{Role: RoleSystem, Content: "You are helpful"},
				{Role: RoleUser, Content: "Hello"},
			},
			MaxTokens:   100,
			Temperature: 0.7,
		}

		azureReq := provider.buildRequest(req)
		assert.Len(t, azureReq.Messages, 2)
		assert.Equal(t, "system", azureReq.Messages[0].Role)
		assert.Equal(t, "user", azureReq.Messages[1].Role)
		assert.Equal(t, 100, azureReq.MaxTokens)
		assert.Equal(t, 0.7, azureReq.Temperature)
	})

	t.Run("converts tools correctly", func(t *testing.T) {
		req := &ChatRequest{
			Messages: []Message{{Role: RoleUser, Content: "Search for info"}},
			Tools: []Tool{
				{
					Type: "function",
					Function: ToolFunction{
						Name:        "search",
						Description: "Search the web",
						Parameters:  map[string]interface{}{"type": "object"},
					},
				},
			},
		}

		azureReq := provider.buildRequest(req)
		assert.Len(t, azureReq.Tools, 1)
		assert.Equal(t, "function", azureReq.Tools[0].Type)
		assert.Equal(t, "search", azureReq.Tools[0].Function.Name)
		assert.NotNil(t, azureReq.ParallelToolCalls)
		assert.True(t, *azureReq.ParallelToolCalls)
	})

	t.Run("converts tool calls in messages", func(t *testing.T) {
		req := &ChatRequest{
			Messages: []Message{
				{
					Role: RoleAssistant,
					ToolCalls: []ToolCall{
						{
							ID:   "call_1",
							Type: "function",
							Function: FunctionCall{
								Name:      "get_weather",
								Arguments: `{"location": "NYC"}`,
							},
						},
					},
				},
			},
		}

		azureReq := provider.buildRequest(req)
		assert.Len(t, azureReq.Messages[0].ToolCalls, 1)
		assert.Equal(t, "call_1", azureReq.Messages[0].ToolCalls[0].ID)
		assert.Equal(t, "get_weather", azureReq.Messages[0].ToolCalls[0].Function.Name)
	})
}

func TestAzureProvider_ConvertResponse(t *testing.T) {
	provider := &azureProvider{}

	t.Run("converts basic response", func(t *testing.T) {
		openaiResp := &openAIResponse{
			ID:    "resp-123",
			Model: "gpt-4",
			Choices: []openAIChoice{
				{
					Index: 0,
					Message: openAIMessage{
						Role:    "assistant",
						Content: "Hello!",
					},
					FinishReason: "stop",
				},
			},
			Usage: &openAIUsage{
				PromptTokens:     5,
				CompletionTokens: 2,
				TotalTokens:      7,
			},
		}

		resp := provider.convertResponse(openaiResp)
		assert.Equal(t, "resp-123", resp.ID)
		assert.Equal(t, "gpt-4", resp.Model)
		assert.Len(t, resp.Choices, 1)
		assert.Equal(t, RoleAssistant, resp.Choices[0].Message.Role)
		assert.Equal(t, "Hello!", resp.Choices[0].Message.Content)
		assert.Equal(t, 7, resp.Usage.TotalTokens)
	})

	t.Run("converts response with tool calls", func(t *testing.T) {
		openaiResp := &openAIResponse{
			ID:    "resp-456",
			Model: "gpt-4",
			Choices: []openAIChoice{
				{
					Index: 0,
					Message: openAIMessage{
						Role: "assistant",
						ToolCalls: []openAIToolCall{
							{
								ID:   "call_abc",
								Type: "function",
								Function: openAIFunctionCall{
									Name:      "search",
									Arguments: `{"query": "weather"}`,
								},
							},
						},
					},
					FinishReason: "tool_calls",
				},
			},
		}

		resp := provider.convertResponse(openaiResp)
		assert.Len(t, resp.Choices[0].Message.ToolCalls, 1)
		assert.Equal(t, "call_abc", resp.Choices[0].Message.ToolCalls[0].ID)
		assert.Equal(t, "search", resp.Choices[0].Message.ToolCalls[0].Function.Name)
	})
}

func TestAzureRequest_Struct(t *testing.T) {
	t.Run("marshals correctly", func(t *testing.T) {
		req := azureRequest{
			Messages: []openAIMessage{
				{Role: "user", Content: "Hello"},
			},
			MaxTokens:   100,
			Temperature: 0.5,
			Stream:      true,
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var decoded map[string]interface{}
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)
		assert.Equal(t, float64(100), decoded["max_tokens"])
		assert.Equal(t, true, decoded["stream"])
	})
}

func TestAzureStreamOptions_Struct(t *testing.T) {
	t.Run("marshals include_usage", func(t *testing.T) {
		opts := azureStreamOptions{IncludeUsage: true}
		data, err := json.Marshal(opts)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"include_usage":true`)
	})
}
