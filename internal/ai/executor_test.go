package ai

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExecuteRequest_Struct(t *testing.T) {
	req := ExecuteRequest{
		ChatbotName:       "test-bot",
		ChatbotID:         "bot-123",
		ConversationID:    "conv-456",
		UserID:            "user-789",
		Role:              "authenticated",
		SQL:               "SELECT * FROM users",
		Description:       "Get all users",
		AllowedSchemas:    []string{"public"},
		AllowedTables:     []string{"users", "profiles"},
		AllowedOperations: []string{"SELECT"},
	}

	assert.Equal(t, "test-bot", req.ChatbotName)
	assert.Equal(t, "bot-123", req.ChatbotID)
	assert.Equal(t, "conv-456", req.ConversationID)
	assert.Equal(t, "user-789", req.UserID)
	assert.Equal(t, "authenticated", req.Role)
	assert.Equal(t, "SELECT * FROM users", req.SQL)
	assert.Equal(t, []string{"public"}, req.AllowedSchemas)
	assert.Equal(t, []string{"users", "profiles"}, req.AllowedTables)
}

func TestExecuteResult_Struct(t *testing.T) {
	t.Run("successful result", func(t *testing.T) {
		result := ExecuteResult{
			Success:  true,
			RowCount: 5,
			Columns:  []string{"id", "name"},
			Rows: []map[string]any{
				{"id": 1, "name": "Alice"},
				{"id": 2, "name": "Bob"},
			},
			Summary:        "Query returned 5 rows",
			DurationMs:     25,
			TablesAccessed: []string{"users"},
			OperationsUsed: []string{"SELECT"},
		}

		assert.True(t, result.Success)
		assert.Equal(t, 5, result.RowCount)
		assert.Equal(t, []string{"id", "name"}, result.Columns)
		assert.Len(t, result.Rows, 2)
		assert.Equal(t, int64(25), result.DurationMs)
	})

	t.Run("error result", func(t *testing.T) {
		result := ExecuteResult{
			Success:        false,
			Error:          "Permission denied",
			Summary:        "Query was rejected",
			TablesAccessed: []string{"secrets"},
		}

		assert.False(t, result.Success)
		assert.Equal(t, "Permission denied", result.Error)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		result := ExecuteResult{
			Success:        true,
			RowCount:       1,
			Columns:        []string{"count"},
			Rows:           []map[string]any{{"count": 42}},
			Summary:        "Count query",
			DurationMs:     10,
			TablesAccessed: []string{"items"},
			OperationsUsed: []string{"SELECT"},
		}

		data, err := json.Marshal(result)
		assert.NoError(t, err)

		var decoded ExecuteResult
		err = json.Unmarshal(data, &decoded)
		assert.NoError(t, err)
		assert.True(t, decoded.Success)
		assert.Equal(t, 1, decoded.RowCount)
	})
}

func TestAI_ConvertValue(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		result := convertValue(nil)
		assert.Nil(t, result)
	})

	t.Run("byte slice as JSON", func(t *testing.T) {
		jsonBytes := []byte(`{"key":"value"}`)
		result := convertValue(jsonBytes)

		converted, ok := result.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "value", converted["key"])
	})

	t.Run("byte slice as string", func(t *testing.T) {
		textBytes := []byte("plain text")
		result := convertValue(textBytes)
		assert.Equal(t, "plain text", result)
	})

	t.Run("time value", func(t *testing.T) {
		testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		result := convertValue(testTime)
		assert.Equal(t, "2024-01-15T10:30:00Z", result)
	})

	t.Run("string value", func(t *testing.T) {
		result := convertValue("hello")
		assert.Equal(t, "hello", result)
	})

	t.Run("int value", func(t *testing.T) {
		result := convertValue(42)
		assert.Equal(t, 42, result)
	})

	t.Run("float value", func(t *testing.T) {
		result := convertValue(3.14)
		assert.Equal(t, 3.14, result)
	})

	t.Run("bool value", func(t *testing.T) {
		result := convertValue(true)
		assert.Equal(t, true, result)
	})

	t.Run("JSON array bytes", func(t *testing.T) {
		jsonBytes := []byte(`[1, 2, 3]`)
		result := convertValue(jsonBytes)

		arr, ok := result.([]interface{})
		assert.True(t, ok)
		assert.Len(t, arr, 3)
	})

	t.Run("invalid JSON bytes falls back to string", func(t *testing.T) {
		invalidJSON := []byte(`{invalid`)
		result := convertValue(invalidJSON)
		assert.Equal(t, "{invalid", result)
	})
}

func TestMin(t *testing.T) {
	t.Run("first is smaller", func(t *testing.T) {
		assert.Equal(t, 5, min(5, 10))
	})

	t.Run("second is smaller", func(t *testing.T) {
		assert.Equal(t, 3, min(10, 3))
	})

	t.Run("equal values", func(t *testing.T) {
		assert.Equal(t, 7, min(7, 7))
	})

	t.Run("negative values", func(t *testing.T) {
		assert.Equal(t, -10, min(-10, -5))
	})

	t.Run("zero values", func(t *testing.T) {
		assert.Equal(t, 0, min(0, 5))
		assert.Equal(t, 0, min(5, 0))
	})
}

func TestExecutor_BuildSummary(t *testing.T) {
	executor := &Executor{
		maxRows: 100,
	}

	t.Run("empty result", func(t *testing.T) {
		req := &ExecuteRequest{ChatbotName: "test"}
		result := &ExecuteResult{
			RowCount:       0,
			TablesAccessed: []string{"users"},
		}

		summary := executor.buildSummary(req, result)
		assert.Contains(t, summary, "returned 0 rows")
		assert.Contains(t, summary, "no data matches")
		assert.Contains(t, summary, "users")
	})

	t.Run("normal result", func(t *testing.T) {
		req := &ExecuteRequest{ChatbotName: "test"}
		result := &ExecuteResult{
			RowCount:       5,
			Columns:        []string{"name", "email"},
			Rows:           []map[string]any{{"name": "Alice"}, {"name": "Bob"}, {"name": "Charlie"}},
			TablesAccessed: []string{"users"},
		}

		summary := executor.buildSummary(req, result)
		assert.Contains(t, summary, "returned 5 row(s)")
		assert.Contains(t, summary, "users")
		assert.Contains(t, summary, "Sample name values")
	})

	t.Run("limited result", func(t *testing.T) {
		executor := &Executor{maxRows: 10}
		req := &ExecuteRequest{ChatbotName: "test"}
		result := &ExecuteResult{
			RowCount:       10,
			Columns:        []string{"id"},
			Rows:           make([]map[string]any, 10),
			TablesAccessed: []string{"items"},
		}

		summary := executor.buildSummary(req, result)
		assert.Contains(t, summary, "limited to 10")
	})

	t.Run("result with no columns", func(t *testing.T) {
		req := &ExecuteRequest{ChatbotName: "test"}
		result := &ExecuteResult{
			RowCount:       3,
			Columns:        []string{},
			Rows:           []map[string]any{},
			TablesAccessed: []string{"data"},
		}

		summary := executor.buildSummary(req, result)
		assert.Contains(t, summary, "returned 3 row(s)")
		assert.NotContains(t, summary, "Sample")
	})
}

func TestNewExecutor(t *testing.T) {
	t.Run("creates executor with specified values", func(t *testing.T) {
		executor := NewExecutor(nil, nil, 500, 30*time.Second)
		assert.NotNil(t, executor)
		assert.Equal(t, 500, executor.maxRows)
		assert.Equal(t, 30*time.Second, executor.timeout)
	})
}
