package ai

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nimbleflux/fluxbase/internal/config"
)

func TestHashQuestion(t *testing.T) {
	// Normalization: case, whitespace, and surrounding space are collapsed.
	a := HashQuestion("  What is   the RETURN Policy?  ")
	b := HashQuestion("what is the return policy?")
	assert.Equal(t, a, b)

	// Different questions hash differently.
	assert.NotEqual(t, a, HashQuestion("what is the refund policy?"))

	// Stable across calls.
	assert.Equal(t, HashQuestion("hello world"), HashQuestion("hello world"))
}

func TestChunkResponse(t *testing.T) {
	t.Run("short text is one chunk", func(t *testing.T) {
		assert.Equal(t, []string{"hello"}, chunkResponse("hello", 80))
	})

	t.Run("splits at word boundaries", func(t *testing.T) {
		chunks := chunkResponse("the quick brown fox jumps over the lazy dog again and again", 16)
		joined := strings.Join(chunks, " ")
		assert.Equal(t, "the quick brown fox jumps over the lazy dog again and again", joined)
		for _, c := range chunks {
			assert.LessOrEqual(t, len(c), 16)
		}
	})

	t.Run("handles multi-byte runes", func(t *testing.T) {
		text := strings.Repeat("héllo wörld ", 20)
		chunks := chunkResponse(text, 10)
		assert.True(t, len(chunks) > 1)
		for _, c := range chunks {
			assert.False(t, strings.ContainsRune(c, 0xFFFD), "chunk must not split runes: %q", c)
		}
	})

	t.Run("empty text yields no chunks", func(t *testing.T) {
		assert.Empty(t, chunkResponse("", 80))
	})
}

func TestResponseCacheEligible(t *testing.T) {
	optedIn := &Chatbot{ID: "cb-1", ResponseCacheEnabled: true}
	singleTurn := &ConversationState{}

	newHandler := func() *ChatHandler {
		return &ChatHandler{
			responseCache:    &ResponseCacheStore{},
			embeddingService: &EmbeddingService{},
			config:           &config.AIConfig{ResponseCacheTTL: 1_000_000},
		}
	}

	t.Run("not eligible without store", func(t *testing.T) {
		handler := &ChatHandler{embeddingService: &EmbeddingService{}}
		assert.False(t, handler.responseCacheEligible(optedIn, singleTurn))
	})

	t.Run("not eligible without embedding service", func(t *testing.T) {
		handler := &ChatHandler{responseCache: &ResponseCacheStore{}}
		assert.False(t, handler.responseCacheEligible(optedIn, singleTurn))
	})

	t.Run("not eligible without opt-in", func(t *testing.T) {
		handler := newHandler()
		assert.False(t, handler.responseCacheEligible(&Chatbot{ID: "cb-1"}, singleTurn))
	})

	t.Run("not eligible for multi-turn conversations", func(t *testing.T) {
		handler := newHandler()
		multiTurn := &ConversationState{Messages: []Message{{Role: RoleUser, Content: "hi"}}}
		assert.False(t, handler.responseCacheEligible(optedIn, multiTurn))
	})

	t.Run("not eligible when globally disabled via ttl", func(t *testing.T) {
		handler := newHandler()
		handler.config.ResponseCacheTTL = 0
		assert.False(t, handler.responseCacheEligible(optedIn, singleTurn))
	})

	t.Run("eligible when all conditions met", func(t *testing.T) {
		handler := newHandler()
		assert.True(t, handler.responseCacheEligible(optedIn, singleTurn))
	})
}

func TestResponseCacheModelKey(t *testing.T) {
	assert.Equal(t, "gpt-4", responseCacheModelKey(&Chatbot{Model: "gpt-4"}, nil))
	assert.Equal(t, "openai", responseCacheModelKey(&Chatbot{}, &fakeProvider{name: "openai"}))
	assert.Equal(t, "", responseCacheModelKey(nil, nil))
}
