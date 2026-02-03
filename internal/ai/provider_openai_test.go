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

func TestNewOpenAIProviderInternal(t *testing.T) {
	t.Run("creates provider with default base URL", func(t *testing.T) {
		config := OpenAIConfig{
			APIKey: "test-key",
			Model:  "gpt-4",
		}

		provider, err := newOpenAIProviderInternal("openai-test", config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "openai-test", provider.name)
		assert.Equal(t, defaultOpenAIBaseURL, provider.config.BaseURL)
	})

	t.Run("uses custom base URL", func(t *testing.T) {
		config := OpenAIConfig{
			APIKey:  "test-key",
			Model:   "gpt-4",
			BaseURL: "https://custom.api.com/",
		}

		provider, err := newOpenAIProviderInternal("custom", config)
		require.NoError(t, err)
		assert.Equal(t, "https://custom.api.com", provider.config.BaseURL)
	})
}

func TestOpenAIProvider_Name(t *testing.T) {
	provider := &openAIProvider{name: "my-openai-provider"}
	assert.Equal(t, "my-openai-provider", provider.Name())
}

func TestOpenAIProvider_Type(t *testing.T) {
	provider := &openAIProvider{}
	assert.Equal(t, ProviderTypeOpenAI, provider.Type())
}

func TestOpenAIProvider_ValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      OpenAIConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: OpenAIConfig{
				APIKey: "key",
				Model:  "gpt-4",
			},
			expectError: false,
		},
		{
			name: "missing API key",
			config: OpenAIConfig{
				Model: "gpt-4",
			},
			expectError: true,
			errorMsg:    "api_key is required",
		},
		{
			name: "missing model",
			config: OpenAIConfig{
				APIKey: "key",
			},
			expectError: true,
			errorMsg:    "model is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &openAIProvider{config: tt.config}
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

func TestOpenAIProvider_Close(t *testing.T) {
	provider := &openAIProvider{
		httpClient: &http.Client{},
	}
	err := provider.Close()
	assert.NoError(t, err)
}

func TestOpenAIProvider_Chat(t *testing.T) {
	t.Run("successful chat request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/chat/completions", r.URL.Path)
			assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

			response := openAIResponse{
				ID:    "chatcmpl-456",
				Model: "gpt-4",
				Choices: []openAIChoice{
					{
						Index: 0,
						Message: openAIMessage{
							Role:    "assistant",
							Content: "Hi there!",
						},
						FinishReason: "stop",
					},
				},
				Usage: &openAIUsage{
					PromptTokens:     5,
					CompletionTokens: 3,
					TotalTokens:      8,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		}))
		defer server.Close()

		provider := &openAIProvider{
			name: "test-openai",
			config: OpenAIConfig{
				APIKey:  "test-api-key",
				Model:   "gpt-4",
				BaseURL: server.URL,
			},
			httpClient: server.Client(),
		}

		req := &ChatRequest{
			Messages: []Message{
				{Role: RoleUser, Content: "Hi"},
			},
		}

		resp, err := provider.Chat(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "chatcmpl-456", resp.ID)
		assert.Equal(t, "Hi there!", resp.Choices[0].Message.Content)
	})

	t.Run("includes organization header when set", func(t *testing.T) {
		var receivedOrgHeader string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedOrgHeader = r.Header.Get("OpenAI-Organization")
			response := openAIResponse{
				Choices: []openAIChoice{{Message: openAIMessage{Role: "assistant"}}},
			}
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		}))
		defer server.Close()

		provider := &openAIProvider{
			config: OpenAIConfig{
				APIKey:         "key",
				Model:          "gpt-4",
				BaseURL:        server.URL,
				OrganizationID: "org-abc123",
			},
			httpClient: server.Client(),
		}

		req := &ChatRequest{Messages: []Message{{Role: RoleUser, Content: "Hi"}}}
		_, err := provider.Chat(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "org-abc123", receivedOrgHeader)
	})

	t.Run("handles API error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := openAIResponse{
				Error: &openAIError{
					Message: "Rate limit exceeded",
					Type:    "rate_limit_error",
					Code:    "rate_limit_exceeded",
				},
			}
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		}))
		defer server.Close()

		provider := &openAIProvider{
			config: OpenAIConfig{
				APIKey:  "key",
				Model:   "gpt-4",
				BaseURL: server.URL,
			},
			httpClient: server.Client(),
		}

		req := &ChatRequest{Messages: []Message{{Role: RoleUser, Content: "Hi"}}}
		_, err := provider.Chat(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Rate limit exceeded")
	})
}

func TestOpenAIProvider_BuildRequest(t *testing.T) {
	provider := &openAIProvider{
		config: OpenAIConfig{Model: "gpt-4"},
	}

	t.Run("uses request model over config model", func(t *testing.T) {
		req := &ChatRequest{
			Model:    "gpt-3.5-turbo",
			Messages: []Message{{Role: RoleUser, Content: "Hi"}},
		}

		openaiReq := provider.buildRequest(req)
		assert.Equal(t, "gpt-3.5-turbo", openaiReq.Model)
	})

	t.Run("falls back to config model", func(t *testing.T) {
		req := &ChatRequest{
			Messages: []Message{{Role: RoleUser, Content: "Hi"}},
		}

		openaiReq := provider.buildRequest(req)
		assert.Equal(t, "gpt-4", openaiReq.Model)
	})

	t.Run("converts all message fields", func(t *testing.T) {
		req := &ChatRequest{
			Messages: []Message{
				{Role: RoleSystem, Content: "System prompt"},
				{Role: RoleUser, Content: "User message"},
				{Role: RoleTool, Content: "Tool result", ToolCallID: "call_123", Name: "search"},
			},
		}

		openaiReq := provider.buildRequest(req)
		assert.Len(t, openaiReq.Messages, 3)
		assert.Equal(t, "tool", openaiReq.Messages[2].Role)
		assert.Equal(t, "call_123", openaiReq.Messages[2].ToolCallID)
		assert.Equal(t, "search", openaiReq.Messages[2].Name)
	})

	t.Run("sets parallel tool calls for tools", func(t *testing.T) {
		req := &ChatRequest{
			Messages: []Message{{Role: RoleUser, Content: "Hi"}},
			Tools: []Tool{
				{Type: "function", Function: ToolFunction{Name: "test"}},
			},
		}

		openaiReq := provider.buildRequest(req)
		assert.NotNil(t, openaiReq.ParallelToolCalls)
		assert.True(t, *openaiReq.ParallelToolCalls)
	})

	t.Run("does not set parallel tool calls without tools", func(t *testing.T) {
		req := &ChatRequest{
			Messages: []Message{{Role: RoleUser, Content: "Hi"}},
		}

		openaiReq := provider.buildRequest(req)
		assert.Nil(t, openaiReq.ParallelToolCalls)
	})
}

func TestOpenAIProvider_ConvertResponse(t *testing.T) {
	provider := &openAIProvider{}

	t.Run("converts response without usage", func(t *testing.T) {
		openaiResp := &openAIResponse{
			ID:    "resp-789",
			Model: "gpt-4",
			Choices: []openAIChoice{
				{
					Index:        0,
					Message:      openAIMessage{Role: "assistant", Content: "Response"},
					FinishReason: "stop",
				},
			},
			Usage: nil,
		}

		resp := provider.convertResponse(openaiResp)
		assert.Equal(t, "resp-789", resp.ID)
		assert.Nil(t, resp.Usage)
	})
}

func TestOpenAIRequest_Struct(t *testing.T) {
	t.Run("marshals all fields", func(t *testing.T) {
		parallel := true
		req := openAIRequest{
			Model: "gpt-4",
			Messages: []openAIMessage{
				{Role: "user", Content: "Hello"},
			},
			Tools: []openAITool{
				{Type: "function", Function: openAIToolFunc{Name: "test", Description: "Test func"}},
			},
			MaxTokens:         100,
			Temperature:       0.7,
			Stream:            true,
			StreamOptions:     &openAIStreamOptions{IncludeUsage: true},
			ToolChoice:        "auto",
			ParallelToolCalls: &parallel,
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"model":"gpt-4"`)
		assert.Contains(t, string(data), `"parallel_tool_calls":true`)
	})
}

func TestOpenAIMessage_Struct(t *testing.T) {
	t.Run("marshals tool call message", func(t *testing.T) {
		msg := openAIMessage{
			Role: "assistant",
			ToolCalls: []openAIToolCall{
				{
					Index: 0,
					ID:    "call_1",
					Type:  "function",
					Function: openAIFunctionCall{
						Name:      "search",
						Arguments: `{"query": "test"}`,
					},
				},
			},
		}

		data, err := json.Marshal(msg)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"tool_calls"`)
		assert.Contains(t, string(data), `"call_1"`)
	})
}

func TestOpenAIResponse_Struct(t *testing.T) {
	t.Run("unmarshals complete response", func(t *testing.T) {
		jsonData := `{
			"id": "chatcmpl-123",
			"object": "chat.completion",
			"created": 1677652288,
			"model": "gpt-4",
			"choices": [{
				"index": 0,
				"message": {"role": "assistant", "content": "Hello!"},
				"finish_reason": "stop"
			}],
			"usage": {
				"prompt_tokens": 10,
				"completion_tokens": 5,
				"total_tokens": 15
			}
		}`

		var resp openAIResponse
		err := json.Unmarshal([]byte(jsonData), &resp)
		require.NoError(t, err)
		assert.Equal(t, "chatcmpl-123", resp.ID)
		assert.Equal(t, "chat.completion", resp.Object)
		assert.Equal(t, int64(1677652288), resp.Created)
		assert.Len(t, resp.Choices, 1)
		assert.Equal(t, 15, resp.Usage.TotalTokens)
	})
}

func TestOpenAIStreamChunk_Struct(t *testing.T) {
	t.Run("unmarshals streaming chunk", func(t *testing.T) {
		jsonData := `{
			"id": "chatcmpl-123",
			"object": "chat.completion.chunk",
			"created": 1677652288,
			"model": "gpt-4",
			"choices": [{
				"index": 0,
				"delta": {"content": "Hello"},
				"finish_reason": null
			}]
		}`

		var chunk openAIStreamChunk
		err := json.Unmarshal([]byte(jsonData), &chunk)
		require.NoError(t, err)
		assert.Equal(t, "chatcmpl-123", chunk.ID)
		assert.Equal(t, "Hello", chunk.Choices[0].Delta.Content)
	})
}

func TestOpenAIError_Struct(t *testing.T) {
	t.Run("unmarshals error response", func(t *testing.T) {
		jsonData := `{
			"error": {
				"message": "Invalid API key",
				"type": "invalid_request_error",
				"code": "invalid_api_key"
			}
		}`

		var resp openAIResponse
		err := json.Unmarshal([]byte(jsonData), &resp)
		require.NoError(t, err)
		assert.NotNil(t, resp.Error)
		assert.Equal(t, "Invalid API key", resp.Error.Message)
		assert.Equal(t, "invalid_request_error", resp.Error.Type)
		assert.Equal(t, "invalid_api_key", resp.Error.Code)
	})
}
