package ai

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConversationState_Struct(t *testing.T) {
	userID := "user-123"
	sessionID := "session-456"
	expiresAt := time.Now().Add(24 * time.Hour)

	state := ConversationState{
		ID:                    "conv-789",
		ChatbotID:             "bot-101",
		ChatbotName:           "Test Bot",
		UserID:                &userID,
		SessionID:             &sessionID,
		Messages:              []Message{{Role: "user", Content: "Hello"}},
		TotalPromptTokens:     100,
		TotalCompletionTokens: 50,
		TurnCount:             2,
		LastAccess:            time.Now(),
		PersistToDatabase:     true,
		ExpiresAt:             &expiresAt,
	}

	assert.Equal(t, "conv-789", state.ID)
	assert.Equal(t, "bot-101", state.ChatbotID)
	assert.Equal(t, "Test Bot", state.ChatbotName)
	assert.NotNil(t, state.UserID)
	assert.Equal(t, "user-123", *state.UserID)
	assert.NotNil(t, state.SessionID)
	assert.Len(t, state.Messages, 1)
	assert.Equal(t, 100, state.TotalPromptTokens)
	assert.Equal(t, 50, state.TotalCompletionTokens)
	assert.Equal(t, 2, state.TurnCount)
	assert.True(t, state.PersistToDatabase)
	assert.NotNil(t, state.ExpiresAt)
}

func TestConversation_Struct(t *testing.T) {
	userID := "user-123"
	sessionID := "session-456"
	title := "Test Conversation"
	expiresAt := time.Now().Add(24 * time.Hour)
	now := time.Now()

	conv := Conversation{
		ID:                    "conv-789",
		ChatbotID:             "bot-101",
		UserID:                &userID,
		SessionID:             &sessionID,
		Title:                 &title,
		Status:                "active",
		TurnCount:             5,
		TotalPromptTokens:     500,
		TotalCompletionTokens: 250,
		CreatedAt:             now,
		UpdatedAt:             now,
		LastMessageAt:         now,
		ExpiresAt:             &expiresAt,
	}

	assert.Equal(t, "conv-789", conv.ID)
	assert.Equal(t, "bot-101", conv.ChatbotID)
	assert.Equal(t, "user-123", *conv.UserID)
	assert.Equal(t, "session-456", *conv.SessionID)
	assert.Equal(t, "Test Conversation", *conv.Title)
	assert.Equal(t, "active", conv.Status)
	assert.Equal(t, 5, conv.TurnCount)
	assert.Equal(t, 500, conv.TotalPromptTokens)
	assert.Equal(t, 250, conv.TotalCompletionTokens)
}

func TestConversationMessage_Struct(t *testing.T) {
	t.Run("user message", func(t *testing.T) {
		msg := ConversationMessage{
			ID:             "msg-123",
			ConversationID: "conv-456",
			Role:           "user",
			Content:        "What is the weather?",
			CreatedAt:      time.Now(),
			SequenceNumber: 1,
		}

		assert.Equal(t, "msg-123", msg.ID)
		assert.Equal(t, "conv-456", msg.ConversationID)
		assert.Equal(t, "user", msg.Role)
		assert.Equal(t, "What is the weather?", msg.Content)
		assert.Equal(t, 1, msg.SequenceNumber)
	})

	t.Run("assistant message with tool call", func(t *testing.T) {
		toolCallID := "call-789"
		toolName := "sql_query"
		executedSQL := "SELECT * FROM users"
		summary := "Found 5 users"
		rowCount := 5
		durationMs := 25

		msg := ConversationMessage{
			ID:               "msg-124",
			ConversationID:   "conv-456",
			Role:             "assistant",
			Content:          "Here are the results:",
			ToolCallID:       &toolCallID,
			ToolName:         &toolName,
			ToolInput:        map[string]interface{}{"query": "users"},
			ToolOutput:       map[string]interface{}{"success": true},
			ExecutedSQL:      &executedSQL,
			SQLResultSummary: &summary,
			SQLRowCount:      &rowCount,
			SQLDurationMs:    &durationMs,
			QueryResults: []map[string]interface{}{
				{"id": 1, "name": "Alice"},
			},
			SequenceNumber: 2,
		}

		assert.Equal(t, "assistant", msg.Role)
		assert.NotNil(t, msg.ToolCallID)
		assert.Equal(t, "call-789", *msg.ToolCallID)
		assert.NotNil(t, msg.ToolName)
		assert.Equal(t, "sql_query", *msg.ToolName)
		assert.NotNil(t, msg.ToolInput)
		assert.Equal(t, "SELECT * FROM users", *msg.ExecutedSQL)
		assert.Equal(t, 5, *msg.SQLRowCount)
		assert.Equal(t, 25, *msg.SQLDurationMs)
		assert.Len(t, msg.QueryResults, 1)
	})

	t.Run("message with token counts", func(t *testing.T) {
		promptTokens := 100
		completionTokens := 50

		msg := ConversationMessage{
			ID:               "msg-125",
			ConversationID:   "conv-456",
			Role:             "assistant",
			Content:          "Response text",
			PromptTokens:     &promptTokens,
			CompletionTokens: &completionTokens,
			SequenceNumber:   3,
		}

		assert.NotNil(t, msg.PromptTokens)
		assert.Equal(t, 100, *msg.PromptTokens)
		assert.NotNil(t, msg.CompletionTokens)
		assert.Equal(t, 50, *msg.CompletionTokens)
	})
}

func TestNewConversationManager(t *testing.T) {
	t.Run("creates manager with specified parameters", func(t *testing.T) {
		manager := NewConversationManager(nil, 30*time.Minute, 50)
		assert.NotNil(t, manager)
		assert.Equal(t, 30*time.Minute, manager.cacheTTL)
		assert.Equal(t, 50, manager.maxTurns)
		assert.NotNil(t, manager.cache)

		// Close cleanup goroutine
		close(manager.cleanupDone)
	})
}

func TestConversationManager_CacheOperations(t *testing.T) {
	manager := &ConversationManager{
		cache:       make(map[string]*ConversationState),
		cacheTTL:    30 * time.Minute,
		maxTurns:    50,
		cleanupDone: make(chan struct{}),
	}
	defer close(manager.cleanupDone)

	t.Run("cache stores and retrieves state", func(t *testing.T) {
		state := &ConversationState{
			ID:          "test-conv",
			ChatbotName: "Test Bot",
			LastAccess:  time.Now(),
		}

		manager.cacheMu.Lock()
		manager.cache["test-conv"] = state
		manager.cacheMu.Unlock()

		manager.cacheMu.RLock()
		retrieved, exists := manager.cache["test-conv"]
		manager.cacheMu.RUnlock()

		assert.True(t, exists)
		assert.Equal(t, "test-conv", retrieved.ID)
		assert.Equal(t, "Test Bot", retrieved.ChatbotName)
	})

	t.Run("cache handles missing keys", func(t *testing.T) {
		manager.cacheMu.RLock()
		_, exists := manager.cache["nonexistent"]
		manager.cacheMu.RUnlock()

		assert.False(t, exists)
	})
}

func TestConversationState_UpdateLastAccess(t *testing.T) {
	state := &ConversationState{
		ID:         "test-conv",
		LastAccess: time.Now().Add(-1 * time.Hour),
	}

	oldAccess := state.LastAccess
	state.LastAccess = time.Now()

	assert.True(t, state.LastAccess.After(oldAccess))
}

func TestConversationState_MessageManagement(t *testing.T) {
	state := &ConversationState{
		ID:       "test-conv",
		Messages: []Message{},
	}

	t.Run("append messages", func(t *testing.T) {
		state.Messages = append(state.Messages, Message{Role: "user", Content: "Hello"})
		state.Messages = append(state.Messages, Message{Role: "assistant", Content: "Hi there!"})

		assert.Len(t, state.Messages, 2)
		assert.Equal(t, "user", state.Messages[0].Role)
		assert.Equal(t, "assistant", state.Messages[1].Role)
	})

	t.Run("update turn count", func(t *testing.T) {
		state.TurnCount = len(state.Messages)
		assert.Equal(t, 2, state.TurnCount)
	})

	t.Run("update token counts", func(t *testing.T) {
		state.TotalPromptTokens += 50
		state.TotalCompletionTokens += 30

		assert.Equal(t, 50, state.TotalPromptTokens)
		assert.Equal(t, 30, state.TotalCompletionTokens)
	})
}
