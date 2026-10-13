package ai

import (
	"context"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// responseCacheChunkSize is the size of the pseudo-streaming deltas used to
// replay a cached answer, small enough that clients render it like a normal
// stream.
const responseCacheChunkSize = 80

// responseCacheEligible reports whether this turn may be served from (or
// stored into) the response cache: the chatbot opted in via the
// @fluxbase:response-cache annotation, the cache store and embedding service
// are wired, and the conversation has no prior turns (multi-turn answers
// depend on history that the cache key doesn't capture).
func (h *ChatHandler) responseCacheEligible(chatbot *Chatbot, state *ConversationState) bool {
	if h.responseCache == nil || h.embeddingService == nil || chatbot == nil {
		return false
	}
	if !chatbot.ResponseCacheEnabled {
		return false
	}
	if h.config != nil && h.config.ResponseCacheTTL <= 0 {
		return false // globally disabled via ttl = 0
	}
	return len(state.Messages) == 0
}

// responseCacheModelKey is the model half of the cache key: the chatbot's
// configured model, or the provider's name when the chatbot delegates model
// choice to the provider, so answers from different models never mix.
func responseCacheModelKey(chatbot *Chatbot, provider Provider) string {
	if chatbot != nil && chatbot.Model != "" {
		return chatbot.Model
	}
	if provider != nil {
		return provider.Name()
	}
	return ""
}

// chunkResponse splits text into replay deltas of at most size bytes,
// preferring word boundaries so multi-byte runes and words stay intact.
func chunkResponse(text string, size int) []string {
	var chunks []string
	for len(text) > 0 {
		chunk := text
		if len(chunk) > size {
			cut := size
			if spaceIdx := strings.LastIndex(text[:cut], " "); spaceIdx > 0 {
				cut = spaceIdx
			}
			chunk = text[:cut]
		}
		text = strings.TrimPrefix(text[len(chunk):], " ")
		chunks = append(chunks, chunk)
	}
	return chunks
}

// tryServeCachedResponse looks up the cache for a single-turn question and,
// on a hit, replays the answer as pseudo-streamed content events followed by
// the done event. Returns true when the turn was served from the cache.
func (h *ChatHandler) tryServeCachedResponse(ctx context.Context, chatCtx *ChatContext, chatbot *Chatbot, msg *ClientMessage) bool {
	questionHash := HashQuestion(msg.Content)

	embedding, err := h.embeddingService.EmbedSingle(ctx, msg.Content, "")
	if err != nil {
		log.Warn().Err(err).Str("chatbot_id", chatbot.ID).Msg("Response cache: failed to embed question; skipping cache")
		return false
	}

	// Exact hash first, then semantic fallback.
	entry, err := h.responseCache.LookupExact(ctx, chatbot.ID, chatbot.Model, questionHash)
	if err != nil {
		log.Warn().Err(err).Msg("Response cache: exact lookup failed")
	}
	if entry == nil {
		entry, err = h.responseCache.LookupSemantic(ctx, chatbot.ID, chatbot.Model, embedding)
		if err != nil {
			log.Warn().Err(err).Msg("Response cache: semantic lookup failed")
		}
	}
	if entry == nil || strings.TrimSpace(entry.Response) == "" {
		if h.metrics != nil {
			h.metrics.RecordAIResponseCache(chatbot.Name, false)
		}
		return false
	}

	if h.metrics != nil {
		h.metrics.RecordAIResponseCache(chatbot.Name, true)
	}
	h.responseCache.RecordHit(ctx, entry.ID)

	// Replay the cached answer as streamed content deltas.
	for _, chunk := range chunkResponse(entry.Response, responseCacheChunkSize) {
		h.send(chatCtx, ServerMessage{
			Type:           "content",
			ConversationID: msg.ConversationID,
			Delta:          chunk,
		})
	}

	// Persist the assistant turn so the conversation keeps working normally
	// from here on (the next message is no longer single-turn).
	assistantMsg := Message{Role: RoleAssistant, Content: entry.Response}
	_ = h.conversations.AddMessage(ctx, msg.ConversationID, assistantMsg, entry.PromptTokens, entry.CompletionTokens)

	usage := UsageStats{
		PromptTokens:     entry.PromptTokens,
		CompletionTokens: entry.CompletionTokens,
		TotalTokens:      entry.PromptTokens + entry.CompletionTokens,
	}
	h.send(chatCtx, ServerMessage{
		Type:           "done",
		ConversationID: msg.ConversationID,
		Usage:          &usage,
		Cached:         true,
	})

	if h.metrics != nil {
		h.metrics.RecordAIChatRequest(chatbot.Name, "success", 0)
	}

	log.Debug().
		Str("chatbot_id", chatbot.ID).
		Str("conversation_id", msg.ConversationID).
		Int("hit_count", entry.HitCount+1).
		Msg("Served response from cache")

	return true
}

// storeCachedResponse saves a plain-text single-turn answer to the response
// cache (async, best-effort). usedTools answers are skipped: they depend on
// tool-time data the cache key doesn't capture.
func (h *ChatHandler) storeCachedResponse(ctx context.Context, chatbot *Chatbot, modelKey string, question, response string, usage UsageStats) {
	if h.responseCache == nil || h.embeddingService == nil || !chatbot.ResponseCacheEnabled {
		return
	}
	if strings.TrimSpace(response) == "" {
		return
	}
	ttl := time.Hour
	if h.config != nil && h.config.ResponseCacheTTL > 0 {
		ttl = h.config.ResponseCacheTTL
	}

	embedCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
	go func() {
		defer cancel()
		embedding, err := h.embeddingService.EmbedSingle(embedCtx, question, "")
		if err != nil {
			log.Warn().Err(err).Str("chatbot_id", chatbot.ID).Msg("Response cache: failed to embed question for store")
			return
		}
		entry := &CachedResponse{
			ChatbotID:        chatbot.ID,
			Model:            modelKey,
			Question:         question,
			Response:         response,
			PromptTokens:     usage.PromptTokens,
			CompletionTokens: usage.CompletionTokens,
		}
		entry.QuestionHash = HashQuestion(question)
		if err := h.responseCache.Store(embedCtx, entry, embedding, ttl); err != nil {
			log.Warn().Err(err).Str("chatbot_id", chatbot.ID).Msg("Response cache: store failed")
		}
	}()
}
