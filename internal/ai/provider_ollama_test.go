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

func TestNewOllamaProviderInternal(t *testing.T) {
	t.Run("creates provider with config", func(t *testing.T) {
		config := OllamaConfig{
			Endpoint: "http://localhost:11434/",
			Model:    "llama2",
		}

		provider, err := newOllamaProviderInternal("ollama-test", config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "ollama-test", provider.name)
		assert.Equal(t, "http://localhost:11434", provider.config.Endpoint)
	})
}

func TestOllamaProvider_Name(t *testing.T) {
	provider := &ollamaProvider{name: "my-ollama-provider"}
	assert.Equal(t, "my-ollama-provider", provider.Name())
}

func TestOllamaProvider_Type(t *testing.T) {
	provider := &ollamaProvider{}
	assert.Equal(t, ProviderTypeOllama, provider.Type())
}

func TestOllamaProvider_ValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      OllamaConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: OllamaConfig{
				Endpoint: "http://localhost:11434",
				Model:    "llama2",
			},
			expectError: false,
		},
		{
			name: "missing endpoint",
			config: OllamaConfig{
				Model: "llama2",
			},
			expectError: true,
			errorMsg:    "endpoint is required",
		},
		{
			name: "missing model",
			config: OllamaConfig{
				Endpoint: "http://localhost:11434",
			},
			expectError: true,
			errorMsg:    "model is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &ollamaProvider{config: tt.config}
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

func TestOllamaProvider_Close(t *testing.T) {
	provider := &ollamaProvider{
		httpClient: &http.Client{},
	}
	err := provider.Close()
	assert.NoError(t, err)
}

func TestOllamaProvider_Chat(t *testing.T) {
	t.Run("successful chat request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/api/chat", r.URL.Path)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			var reqBody ollamaRequest
			err := json.NewDecoder(r.Body).Decode(&reqBody)
			require.NoError(t, err)
			assert.Equal(t, "llama2", reqBody.Model)
			assert.False(t, reqBody.Stream)

			response := ollamaResponse{
				Model:     "llama2",
				CreatedAt: "2024-01-15T10:30:00Z",
				Message: ollamaMessage{
					Role:    "assistant",
					Content: "Hello! How can I help?",
				},
				Done:            true,
				DoneReason:      "stop",
				PromptEvalCount: 10,
				EvalCount:       8,
			}
			err = json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		}))
		defer server.Close()

		provider := &ollamaProvider{
			name: "test-ollama",
			config: OllamaConfig{
				Endpoint: server.URL,
				Model:    "llama2",
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
		assert.Equal(t, "llama2", resp.Model)
		assert.Len(t, resp.Choices, 1)
		assert.Equal(t, "Hello! How can I help?", resp.Choices[0].Message.Content)
		assert.Equal(t, 18, resp.Usage.TotalTokens)
	})

	t.Run("handles HTTP error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Internal server error"))
		}))
		defer server.Close()

		provider := &ollamaProvider{
			config: OllamaConfig{
				Endpoint: server.URL,
				Model:    "llama2",
			},
			httpClient: server.Client(),
		}

		req := &ChatRequest{Messages: []Message{{Role: RoleUser, Content: "Hi"}}}
		_, err := provider.Chat(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "status 500")
	})
}

func TestOllamaProvider_BuildRequest(t *testing.T) {
	provider := &ollamaProvider{
		config: OllamaConfig{Model: "llama2"},
	}

	t.Run("uses request model over config model", func(t *testing.T) {
		req := &ChatRequest{
			Model:    "mistral",
			Messages: []Message{{Role: RoleUser, Content: "Hi"}},
		}

		ollamaReq := provider.buildRequest(req)
		assert.Equal(t, "mistral", ollamaReq.Model)
	})

	t.Run("falls back to config model", func(t *testing.T) {
		req := &ChatRequest{
			Messages: []Message{{Role: RoleUser, Content: "Hi"}},
		}

		ollamaReq := provider.buildRequest(req)
		assert.Equal(t, "llama2", ollamaReq.Model)
	})

	t.Run("sets options when provided", func(t *testing.T) {
		req := &ChatRequest{
			Messages:    []Message{{Role: RoleUser, Content: "Hi"}},
			MaxTokens:   500,
			Temperature: 0.8,
		}

		ollamaReq := provider.buildRequest(req)
		assert.NotNil(t, ollamaReq.Options)
		assert.Equal(t, 500, ollamaReq.Options.NumPredict)
		assert.Equal(t, 0.8, ollamaReq.Options.Temperature)
	})

	t.Run("does not set options when not needed", func(t *testing.T) {
		req := &ChatRequest{
			Messages: []Message{{Role: RoleUser, Content: "Hi"}},
		}

		ollamaReq := provider.buildRequest(req)
		assert.Nil(t, ollamaReq.Options)
	})

	t.Run("converts tools correctly", func(t *testing.T) {
		req := &ChatRequest{
			Messages: []Message{{Role: RoleUser, Content: "Search"}},
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

		ollamaReq := provider.buildRequest(req)
		assert.Len(t, ollamaReq.Tools, 1)
		assert.Equal(t, "search", ollamaReq.Tools[0].Function.Name)
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

		ollamaReq := provider.buildRequest(req)
		assert.Len(t, ollamaReq.Messages[0].ToolCalls, 1)
		assert.Equal(t, "get_weather", ollamaReq.Messages[0].ToolCalls[0].Function.Name)
		assert.Equal(t, "NYC", ollamaReq.Messages[0].ToolCalls[0].Function.Arguments["location"])
	})
}

func TestOllamaProvider_ConvertResponse(t *testing.T) {
	provider := &ollamaProvider{}

	t.Run("converts basic response", func(t *testing.T) {
		ollamaResp := &ollamaResponse{
			Model:     "llama2",
			CreatedAt: "2024-01-15T10:30:00Z",
			Message: ollamaMessage{
				Role:    "assistant",
				Content: "Hello!",
			},
			Done:            true,
			PromptEvalCount: 5,
			EvalCount:       2,
		}

		resp := provider.convertResponse(ollamaResp)
		assert.Equal(t, "llama2", resp.Model)
		assert.Len(t, resp.Choices, 1)
		assert.Equal(t, RoleAssistant, resp.Choices[0].Message.Role)
		assert.Equal(t, "Hello!", resp.Choices[0].Message.Content)
		assert.Equal(t, 7, resp.Usage.TotalTokens)
	})

	t.Run("converts response with native tool calls", func(t *testing.T) {
		ollamaResp := &ollamaResponse{
			Model: "llama2",
			Message: ollamaMessage{
				Role: "assistant",
				ToolCalls: []ollamaToolCall{
					{
						Function: ollamaFunctionCall{
							Name:      "search",
							Arguments: map[string]interface{}{"query": "weather"},
						},
					},
				},
			},
			Done: true,
		}

		resp := provider.convertResponse(ollamaResp)
		assert.Len(t, resp.Choices[0].Message.ToolCalls, 1)
		assert.Equal(t, "call_0", resp.Choices[0].Message.ToolCalls[0].ID)
		assert.Equal(t, "search", resp.Choices[0].Message.ToolCalls[0].Function.Name)
	})

	t.Run("uses done_reason when provided", func(t *testing.T) {
		ollamaResp := &ollamaResponse{
			Model:      "llama2",
			Message:    ollamaMessage{Role: "assistant", Content: "Done"},
			Done:       true,
			DoneReason: "length",
		}

		resp := provider.convertResponse(ollamaResp)
		assert.Equal(t, "length", resp.Choices[0].FinishReason)
	})
}

func TestTryParseToolCallFromContent(t *testing.T) {
	t.Run("parses valid execute_sql tool call", func(t *testing.T) {
		content := `{"name": "execute_sql", "arguments": {"sql": "SELECT * FROM users"}}`
		tc := tryParseToolCallFromContent(content)
		require.NotNil(t, tc)
		assert.Equal(t, "execute_sql", tc.Function.Name)
		assert.Equal(t, "SELECT * FROM users", tc.Function.Arguments["sql"])
	})

	t.Run("parses valid http_request tool call", func(t *testing.T) {
		content := `{"name": "http_request", "arguments": {"url": "https://api.example.com"}}`
		tc := tryParseToolCallFromContent(content)
		require.NotNil(t, tc)
		assert.Equal(t, "http_request", tc.Function.Name)
	})

	t.Run("returns nil for non-JSON content", func(t *testing.T) {
		content := "This is just regular text response"
		tc := tryParseToolCallFromContent(content)
		assert.Nil(t, tc)
	})

	t.Run("returns nil for unknown tool name", func(t *testing.T) {
		content := `{"name": "unknown_tool", "arguments": {"key": "value"}}`
		tc := tryParseToolCallFromContent(content)
		assert.Nil(t, tc)
	})

	t.Run("returns nil for missing name", func(t *testing.T) {
		content := `{"arguments": {"key": "value"}}`
		tc := tryParseToolCallFromContent(content)
		assert.Nil(t, tc)
	})

	t.Run("returns nil for missing arguments", func(t *testing.T) {
		content := `{"name": "execute_sql"}`
		tc := tryParseToolCallFromContent(content)
		assert.Nil(t, tc)
	})

	t.Run("handles whitespace around JSON", func(t *testing.T) {
		content := `  {"name": "execute_sql", "arguments": {"sql": "SELECT 1"}}  `
		tc := tryParseToolCallFromContent(content)
		require.NotNil(t, tc)
		assert.Equal(t, "execute_sql", tc.Function.Name)
	})

	t.Run("returns nil for partial JSON", func(t *testing.T) {
		content := `{"name": "execute_sql"`
		tc := tryParseToolCallFromContent(content)
		assert.Nil(t, tc)
	})
}

func TestOllamaRequest_Struct(t *testing.T) {
	t.Run("marshals correctly", func(t *testing.T) {
		req := ollamaRequest{
			Model: "llama2",
			Messages: []ollamaMessage{
				{Role: "user", Content: "Hello"},
			},
			Stream: true,
			Options: &ollamaOptions{
				Temperature: 0.7,
				NumPredict:  100,
			},
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"model":"llama2"`)
		assert.Contains(t, string(data), `"stream":true`)
		assert.Contains(t, string(data), `"temperature":0.7`)
	})
}

func TestOllamaResponse_Struct(t *testing.T) {
	t.Run("unmarshals complete response", func(t *testing.T) {
		jsonData := `{
			"model": "llama2",
			"created_at": "2024-01-15T10:30:00Z",
			"message": {"role": "assistant", "content": "Hello!"},
			"done": true,
			"done_reason": "stop",
			"total_duration": 1000000000,
			"load_duration": 100000000,
			"prompt_eval_count": 10,
			"prompt_eval_duration": 200000000,
			"eval_count": 5,
			"eval_duration": 500000000
		}`

		var resp ollamaResponse
		err := json.Unmarshal([]byte(jsonData), &resp)
		require.NoError(t, err)
		assert.Equal(t, "llama2", resp.Model)
		assert.True(t, resp.Done)
		assert.Equal(t, "stop", resp.DoneReason)
		assert.Equal(t, 10, resp.PromptEvalCount)
		assert.Equal(t, 5, resp.EvalCount)
	})
}

func TestOllamaMessage_Struct(t *testing.T) {
	t.Run("marshals message with tool calls", func(t *testing.T) {
		msg := ollamaMessage{
			Role:    "assistant",
			Content: "",
			ToolCalls: []ollamaToolCall{
				{
					Function: ollamaFunctionCall{
						Name:      "search",
						Arguments: map[string]interface{}{"query": "weather"},
					},
				},
			},
		}

		data, err := json.Marshal(msg)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"tool_calls"`)
		assert.Contains(t, string(data), `"search"`)
	})
}

func TestOllamaTool_Struct(t *testing.T) {
	t.Run("marshals tool definition", func(t *testing.T) {
		tool := ollamaTool{
			Type: "function",
			Function: ollamaToolFunction{
				Name:        "get_weather",
				Description: "Get current weather",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]interface{}{"type": "string"},
					},
				},
			},
		}

		data, err := json.Marshal(tool)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"get_weather"`)
		assert.Contains(t, string(data), `"Get current weather"`)
	})
}
