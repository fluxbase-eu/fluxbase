package mcp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// NewTransport Tests
// =============================================================================

func TestNewTransport(t *testing.T) {
	t.Run("creates new transport", func(t *testing.T) {
		transport := NewTransport()

		require.NotNil(t, transport)
	})
}

// =============================================================================
// ParseRequest Tests
// =============================================================================

func TestTransport_ParseRequest(t *testing.T) {
	transport := NewTransport()

	t.Run("parses valid request", func(t *testing.T) {
		data := []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)

		req, err := transport.ParseRequest(data)

		require.NoError(t, err)
		require.NotNil(t, req)
		assert.Equal(t, JSONRPCVersion, req.JSONRPC)
		assert.Equal(t, float64(1), req.ID)
		assert.Equal(t, "tools/list", req.Method)
	})

	t.Run("parses request with string ID", func(t *testing.T) {
		data := []byte(`{"jsonrpc":"2.0","id":"request-123","method":"ping"}`)

		req, err := transport.ParseRequest(data)

		require.NoError(t, err)
		assert.Equal(t, "request-123", req.ID)
	})

	t.Run("parses request with params", func(t *testing.T) {
		data := []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"test_tool","arguments":{"key":"value"}}}`)

		req, err := transport.ParseRequest(data)

		require.NoError(t, err)
		assert.NotNil(t, req.Params)
	})

	t.Run("parses notification (no ID)", func(t *testing.T) {
		data := []byte(`{"jsonrpc":"2.0","method":"notifications/initialized"}`)

		req, err := transport.ParseRequest(data)

		require.NoError(t, err)
		assert.Nil(t, req.ID)
		assert.Equal(t, "notifications/initialized", req.Method)
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		data := []byte(`{invalid json}`)

		req, err := transport.ParseRequest(data)

		assert.Error(t, err)
		assert.Nil(t, req)
		assert.Contains(t, err.Error(), "failed to parse request")
	})

	t.Run("returns error for wrong JSON-RPC version", func(t *testing.T) {
		data := []byte(`{"jsonrpc":"1.0","id":1,"method":"test"}`)

		req, err := transport.ParseRequest(data)

		assert.Error(t, err)
		assert.Nil(t, req)
		assert.Contains(t, err.Error(), "invalid JSON-RPC version")
	})

	t.Run("returns error for missing JSON-RPC version", func(t *testing.T) {
		data := []byte(`{"id":1,"method":"test"}`)

		req, err := transport.ParseRequest(data)

		assert.Error(t, err)
		assert.Nil(t, req)
		assert.Contains(t, err.Error(), "invalid JSON-RPC version")
	})

	t.Run("returns error for missing method", func(t *testing.T) {
		data := []byte(`{"jsonrpc":"2.0","id":1}`)

		req, err := transport.ParseRequest(data)

		assert.Error(t, err)
		assert.Nil(t, req)
		assert.Contains(t, err.Error(), "method is required")
	})

	t.Run("returns error for empty method", func(t *testing.T) {
		data := []byte(`{"jsonrpc":"2.0","id":1,"method":""}`)

		req, err := transport.ParseRequest(data)

		assert.Error(t, err)
		assert.Nil(t, req)
		assert.Contains(t, err.Error(), "method is required")
	})

	t.Run("returns error for empty data", func(t *testing.T) {
		data := []byte(``)

		req, err := transport.ParseRequest(data)

		assert.Error(t, err)
		assert.Nil(t, req)
	})
}

// =============================================================================
// SerializeResponse Tests
// =============================================================================

func TestTransport_SerializeResponse(t *testing.T) {
	transport := NewTransport()

	t.Run("serializes success response", func(t *testing.T) {
		resp := &Response{
			JSONRPC: JSONRPCVersion,
			ID:      1,
			Result:  map[string]string{"status": "ok"},
		}

		data, err := transport.SerializeResponse(resp)

		require.NoError(t, err)
		assert.Contains(t, string(data), `"jsonrpc":"2.0"`)
		assert.Contains(t, string(data), `"result"`)
		assert.Contains(t, string(data), `"status":"ok"`)
	})

	t.Run("serializes error response", func(t *testing.T) {
		resp := &Response{
			JSONRPC: JSONRPCVersion,
			ID:      "test-id",
			Error: &Error{
				Code:    ErrorCodeInvalidParams,
				Message: "Invalid params",
			},
		}

		data, err := transport.SerializeResponse(resp)

		require.NoError(t, err)
		assert.Contains(t, string(data), `"error"`)
		assert.Contains(t, string(data), `"code":-32602`)
		assert.Contains(t, string(data), `"message":"Invalid params"`)
	})

	t.Run("serializes response with null ID", func(t *testing.T) {
		resp := &Response{
			JSONRPC: JSONRPCVersion,
			ID:      nil,
			Error: &Error{
				Code:    ErrorCodeParseError,
				Message: "Parse error",
			},
		}

		data, err := transport.SerializeResponse(resp)

		require.NoError(t, err)
		assert.Contains(t, string(data), `"jsonrpc":"2.0"`)
	})

	t.Run("serializes response with string ID", func(t *testing.T) {
		resp := &Response{
			JSONRPC: JSONRPCVersion,
			ID:      "request-456",
			Result:  struct{}{},
		}

		data, err := transport.SerializeResponse(resp)

		require.NoError(t, err)
		assert.Contains(t, string(data), `"id":"request-456"`)
	})
}

// =============================================================================
// ParseParams Tests
// =============================================================================

func TestParseParams(t *testing.T) {
	t.Run("parses valid params", func(t *testing.T) {
		params := json.RawMessage(`{"name":"test","value":42}`)

		type TestParams struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		}

		result, err := ParseParams[TestParams](params)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "test", result.Name)
		assert.Equal(t, 42, result.Value)
	})

	t.Run("returns nil for empty params", func(t *testing.T) {
		params := json.RawMessage(``)

		type TestParams struct {
			Name string `json:"name"`
		}

		result, err := ParseParams[TestParams](params)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("returns nil for nil params", func(t *testing.T) {
		var params json.RawMessage = nil

		type TestParams struct {
			Name string `json:"name"`
		}

		result, err := ParseParams[TestParams](params)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		params := json.RawMessage(`{invalid}`)

		type TestParams struct {
			Name string `json:"name"`
		}

		result, err := ParseParams[TestParams](params)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to parse params")
	})

	t.Run("parses nested objects", func(t *testing.T) {
		params := json.RawMessage(`{"user":{"name":"john","age":30}}`)

		type User struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		type TestParams struct {
			User User `json:"user"`
		}

		result, err := ParseParams[TestParams](params)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "john", result.User.Name)
		assert.Equal(t, 30, result.User.Age)
	})

	t.Run("parses arrays", func(t *testing.T) {
		params := json.RawMessage(`{"items":["a","b","c"]}`)

		type TestParams struct {
			Items []string `json:"items"`
		}

		result, err := ParseParams[TestParams](params)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, []string{"a", "b", "c"}, result.Items)
	})
}

// =============================================================================
// MustParseParams Tests
// =============================================================================

func TestMustParseParams(t *testing.T) {
	t.Run("parses valid params", func(t *testing.T) {
		params := json.RawMessage(`{"name":"test"}`)

		type TestParams struct {
			Name string `json:"name"`
		}

		result, err := MustParseParams[TestParams](params)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "test", result.Name)
	})

	t.Run("returns error for empty params", func(t *testing.T) {
		params := json.RawMessage(``)

		type TestParams struct {
			Name string `json:"name"`
		}

		result, err := MustParseParams[TestParams](params)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "params are required")
	})

	t.Run("returns error for nil params", func(t *testing.T) {
		var params json.RawMessage = nil

		type TestParams struct {
			Name string `json:"name"`
		}

		result, err := MustParseParams[TestParams](params)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "params are required")
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		params := json.RawMessage(`{invalid}`)

		type TestParams struct {
			Name string `json:"name"`
		}

		result, err := MustParseParams[TestParams](params)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to parse params")
	})
}

// =============================================================================
// IsNotification Tests
// =============================================================================

func TestIsNotification(t *testing.T) {
	t.Run("returns true for nil ID", func(t *testing.T) {
		req := &Request{
			JSONRPC: JSONRPCVersion,
			ID:      nil,
			Method:  "test",
		}

		result := IsNotification(req)

		assert.True(t, result)
	})

	t.Run("returns false for numeric ID", func(t *testing.T) {
		req := &Request{
			JSONRPC: JSONRPCVersion,
			ID:      float64(1),
			Method:  "test",
		}

		result := IsNotification(req)

		assert.False(t, result)
	})

	t.Run("returns false for string ID", func(t *testing.T) {
		req := &Request{
			JSONRPC: JSONRPCVersion,
			ID:      "request-123",
			Method:  "test",
		}

		result := IsNotification(req)

		assert.False(t, result)
	})

	t.Run("returns false for zero ID", func(t *testing.T) {
		req := &Request{
			JSONRPC: JSONRPCVersion,
			ID:      float64(0),
			Method:  "test",
		}

		result := IsNotification(req)

		assert.False(t, result)
	})

	t.Run("returns false for empty string ID", func(t *testing.T) {
		req := &Request{
			JSONRPC: JSONRPCVersion,
			ID:      "",
			Method:  "test",
		}

		result := IsNotification(req)

		assert.False(t, result)
	})
}

// =============================================================================
// ValidateID Tests
// =============================================================================

func TestValidateID(t *testing.T) {
	t.Run("accepts nil", func(t *testing.T) {
		assert.True(t, ValidateID(nil))
	})

	t.Run("accepts string", func(t *testing.T) {
		assert.True(t, ValidateID("request-123"))
	})

	t.Run("accepts empty string", func(t *testing.T) {
		assert.True(t, ValidateID(""))
	})

	t.Run("accepts float64", func(t *testing.T) {
		assert.True(t, ValidateID(float64(123)))
	})

	t.Run("accepts int", func(t *testing.T) {
		assert.True(t, ValidateID(42))
	})

	t.Run("accepts int64", func(t *testing.T) {
		assert.True(t, ValidateID(int64(9999)))
	})

	t.Run("rejects slice", func(t *testing.T) {
		assert.False(t, ValidateID([]int{1, 2, 3}))
	})

	t.Run("rejects map", func(t *testing.T) {
		assert.False(t, ValidateID(map[string]int{"a": 1}))
	})

	t.Run("rejects struct", func(t *testing.T) {
		type MyStruct struct{}
		assert.False(t, ValidateID(MyStruct{}))
	})

	t.Run("rejects bool", func(t *testing.T) {
		assert.False(t, ValidateID(true))
	})

	t.Run("rejects int32", func(t *testing.T) {
		assert.False(t, ValidateID(int32(100)))
	})
}

// =============================================================================
// Transport Struct Tests
// =============================================================================

func TestTransport_Struct(t *testing.T) {
	t.Run("is empty struct", func(t *testing.T) {
		transport := Transport{}
		assert.NotNil(t, &transport)
	})
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestTransport_RoundTrip(t *testing.T) {
	transport := NewTransport()

	t.Run("parse and serialize initialize request", func(t *testing.T) {
		data := []byte(`{"jsonrpc":"2.0","id":"init-1","method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"TestClient","version":"1.0"}}}`)

		req, err := transport.ParseRequest(data)
		require.NoError(t, err)
		assert.Equal(t, MethodInitialize, req.Method)

		var params InitializeParams
		err = json.Unmarshal(req.Params, &params)
		require.NoError(t, err)
		assert.Equal(t, "TestClient", params.ClientInfo.Name)
	})

	t.Run("create and serialize response", func(t *testing.T) {
		resp := NewResult("test-id", map[string]string{"status": "ok"})

		data, err := transport.SerializeResponse(resp)
		require.NoError(t, err)

		var parsed Response
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)
		assert.Equal(t, JSONRPCVersion, parsed.JSONRPC)
		assert.Equal(t, "test-id", parsed.ID)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkTransport_ParseRequest(b *testing.B) {
	transport := NewTransport()
	data := []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = transport.ParseRequest(data)
	}
}

func BenchmarkTransport_SerializeResponse(b *testing.B) {
	transport := NewTransport()
	resp := &Response{
		JSONRPC: JSONRPCVersion,
		ID:      1,
		Result:  map[string]string{"status": "ok"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = transport.SerializeResponse(resp)
	}
}

func BenchmarkParseParams(b *testing.B) {
	params := json.RawMessage(`{"name":"test","value":42}`)

	type TestParams struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseParams[TestParams](params)
	}
}

func BenchmarkValidateID(b *testing.B) {
	ids := []any{"string-id", float64(123), int(42), nil}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateID(ids[i%len(ids)])
	}
}
