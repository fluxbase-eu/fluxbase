package ai

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseCacheStore_Integration(t *testing.T) {
	requireDB(t)
	ctx := aiTestCtx(t)

	kbStore := NewKnowledgeBaseStorage(aiTestDB)
	cache := NewResponseCacheStore(aiTestDB)
	chatbotStore := NewStorage(aiTestDB)

	// Chatbot rows back the response_cache FK; the second chatbot verifies
	// cache scoping and that KB invalidation spares unlinked chatbots.
	cb1 := &Chatbot{Name: "cache-itest-" + uuid.New().String()[:8], Namespace: "test", Enabled: true, Source: "api"}
	require.NoError(t, chatbotStore.CreateChatbot(ctx, cb1))
	t.Cleanup(func() { _ = chatbotStore.DeleteChatbot(context.Background(), cb1.ID) })

	cb2 := &Chatbot{Name: "cache-itest-" + uuid.New().String()[:8], Namespace: "test", Enabled: true, Source: "api"}
	require.NoError(t, chatbotStore.CreateChatbot(ctx, cb2))
	t.Cleanup(func() { _ = chatbotStore.DeleteChatbot(context.Background(), cb2.ID) })

	kb := createTestKB(t, ctx, kbStore)
	_, err := kbStore.LinkChatbotKnowledgeBaseSimple(ctx, cb1.ID, kb.ID, "full", 0, 5, 0.7)
	require.NoError(t, err)

	embedding := unitVector3(1, 0, 0)
	entry := &CachedResponse{
		ChatbotID:    cb1.ID,
		Model:        "gpt-test",
		Question:     "What is the return policy?",
		QuestionHash: HashQuestion("What is the return policy?"),
		Response:     "You can return items within 30 days.",
	}
	require.NoError(t, cache.Store(ctx, entry, embedding, time.Hour))

	t.Run("exact lookup hits and records the hit", func(t *testing.T) {
		got, err := cache.LookupExact(ctx, cb1.ID, "gpt-test", HashQuestion("what is THE RETURN policy?"))
		require.NoError(t, err)
		require.NotNil(t, got, "hash normalization must make lookups case/whitespace insensitive")
		assert.Equal(t, entry.Response, got.Response)

		cache.RecordHit(ctx, got.ID)

		row, err := cache.LookupExact(ctx, cb1.ID, "gpt-test", got.QuestionHash)
		require.NoError(t, err)
		assert.Equal(t, 1, row.HitCount)
	})

	t.Run("semantic lookup hits near-duplicates only", func(t *testing.T) {
		got, err := cache.LookupSemantic(ctx, cb1.ID, "gpt-test", unitVector3(0.999, 0.04, 0))
		require.NoError(t, err)
		require.NotNil(t, got, "near-identical embedding (cosine ~0.999) must hit")

		far, err := cache.LookupSemantic(ctx, cb1.ID, "gpt-test", unitVector3(0, 1, 0))
		require.NoError(t, err)
		assert.Nil(t, far, "orthogonal question must not hit the cache")
	})

	t.Run("entries are scoped per chatbot and model", func(t *testing.T) {
		otherBot, err := cache.LookupExact(ctx, cb2.ID, "gpt-test", HashQuestion("What is the return policy?"))
		require.NoError(t, err)
		assert.Nil(t, otherBot, "another chatbot must not see this entry")

		otherModel, err := cache.LookupExact(ctx, cb1.ID, "gpt-other", HashQuestion("What is the return policy?"))
		require.NoError(t, err)
		assert.Nil(t, otherModel, "another model must not see this entry")
	})

	t.Run("expiry and purge", func(t *testing.T) {
		short := &CachedResponse{
			ChatbotID:    cb1.ID,
			Model:        "gpt-test",
			Question:     "Short lived question?",
			QuestionHash: HashQuestion("Short lived question?"),
			Response:     "short answer",
		}
		require.NoError(t, cache.Store(ctx, short, embedding, 10*time.Millisecond))
		time.Sleep(30 * time.Millisecond)

		expired, err := cache.LookupExact(ctx, cb1.ID, "gpt-test", short.QuestionHash)
		require.NoError(t, err)
		assert.Nil(t, expired, "expired entries must not be served")

		deleted, err := cache.PurgeExpired(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, deleted, int64(1), "purge removes at least the expired entry")
	})

	t.Run("chatbot invalidation drops all entries", func(t *testing.T) {
		require.NoError(t, cache.InvalidateForChatbot(ctx, cb1.ID))
		got, err := cache.LookupExact(ctx, cb1.ID, "gpt-test", HashQuestion("What is the return policy?"))
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("KB invalidation drops linked chatbots' entries", func(t *testing.T) {
		// Re-store, then invalidate via the linked knowledge base.
		require.NoError(t, cache.Store(ctx, entry, embedding, time.Hour))

		// cb2 is not linked to the KB and must be untouched.
		other := &CachedResponse{
			ChatbotID:    cb2.ID,
			Model:        "gpt-test",
			Question:     "unrelated?",
			QuestionHash: HashQuestion("unrelated?"),
			Response:     "unrelated answer",
		}
		require.NoError(t, cache.Store(ctx, other, embedding, time.Hour))

		require.NoError(t, cache.InvalidateForKnowledgeBase(ctx, kb.ID))

		linked, err := cache.LookupExact(ctx, cb1.ID, "gpt-test", entry.QuestionHash)
		require.NoError(t, err)
		assert.Nil(t, linked, "linked chatbot's entries must be invalidated")

		unlinked, err := cache.LookupExact(ctx, cb2.ID, "gpt-test", other.QuestionHash)
		require.NoError(t, err)
		require.NotNil(t, unlinked, "unlinked chatbot's entries must survive")
	})
}

func TestKnowledgeBaseFields_Integration(t *testing.T) {
	requireDB(t)
	ctx := aiTestCtx(t)

	kbStore := NewKnowledgeBaseStorage(aiTestDB)
	chatbotStore := NewStorage(aiTestDB)

	t.Run("extraction strategy and rerank flag round-trip", func(t *testing.T) {
		strategy := "hybrid"
		rerank := false
		kb := &KnowledgeBase{
			Name:                     "itest-fields-" + uuid.New().String()[:8],
			Namespace:                "test",
			Enabled:                  true,
			Source:                   "api",
			Visibility:               KBVisibilityPrivate,
			EntityExtractionStrategy: strategy,
			RerankEnabled:            rerank,
		}
		require.NoError(t, kbStore.CreateKnowledgeBase(ctx, kb))
		t.Cleanup(func() { _ = kbStore.DeleteKnowledgeBase(context.Background(), kb.ID) })

		got, err := kbStore.GetKnowledgeBase(ctx, kb.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "hybrid", got.EntityExtractionStrategy)
		assert.False(t, got.RerankEnabled, "explicit rerank_enabled=false must persist")

		// Update both fields.
		newStrategy := "llm"
		newRerank := true
		updated, err := kbStore.UpdateKnowledgeBaseByID(ctx, kb.ID, UpdateKnowledgeBaseRequest{
			EntityExtractionStrategy: &newStrategy,
			RerankEnabled:            &newRerank,
		})
		require.NoError(t, err)
		assert.Equal(t, "llm", updated.EntityExtractionStrategy)
		assert.True(t, updated.RerankEnabled)
	})

	t.Run("invalid extraction strategy is rejected", func(t *testing.T) {
		req := CreateKnowledgeBaseRequest{
			Name:                     "itest-bad-" + uuid.New().String()[:8],
			EntityExtractionStrategy: "magic",
		}
		_, err := kbStore.CreateKnowledgeBaseFromRequest(ctx, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid entity_extraction_strategy")
	})

	t.Run("ChatbotRerankEnabled reflects linked KB opt-in", func(t *testing.T) {
		cb := &Chatbot{Name: "rerank-itest-" + uuid.New().String()[:8], Namespace: "test", Enabled: true, Source: "api"}
		require.NoError(t, chatbotStore.CreateChatbot(ctx, cb))
		t.Cleanup(func() { _ = chatbotStore.DeleteChatbot(context.Background(), cb.ID) })

		kb := createTestKB(t, ctx, kbStore)
		// Default KB rerank_enabled=true → linked chatbot reranks.
		_, err := kbStore.LinkChatbotKnowledgeBaseSimple(ctx, cb.ID, kb.ID, "full", 0, 5, 0.7)
		require.NoError(t, err)

		enabled, err := kbStore.ChatbotRerankEnabled(ctx, cb.ID)
		require.NoError(t, err)
		assert.True(t, enabled, "linked KB with rerank_enabled=true opts the chatbot in")

		// Opt the KB out → chatbot no longer reranks.
		off := false
		_, err = kbStore.UpdateKnowledgeBaseByID(ctx, kb.ID, UpdateKnowledgeBaseRequest{RerankEnabled: &off})
		require.NoError(t, err)

		enabled, err = kbStore.ChatbotRerankEnabled(ctx, cb.ID)
		require.NoError(t, err)
		assert.False(t, enabled, "KB opt-out must disable the chatbot's reranking")

		// An unlinked chatbot never reranks.
		stranger := &Chatbot{Name: "rerank-stranger-" + uuid.New().String()[:8], Namespace: "test", Enabled: true, Source: "api"}
		require.NoError(t, chatbotStore.CreateChatbot(ctx, stranger))
		t.Cleanup(func() { _ = chatbotStore.DeleteChatbot(context.Background(), stranger.ID) })
		enabled, err = kbStore.ChatbotRerankEnabled(ctx, stranger.ID)
		require.NoError(t, err)
		assert.False(t, enabled)
	})
}
