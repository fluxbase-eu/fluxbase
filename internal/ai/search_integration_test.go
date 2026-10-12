package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedHybridFixtures provisions one KB with three chunks exercising distinct
// semantic/keyword rank combinations:
//
//	A "reset instructions here"     cos 1.0  keyword: 1 term           (sem 1, kw 2)
//	B "password reset guide"        cos 0.31 keyword: exact phrase + 3 (sem 3, kw 1)
//	C "unrelated cooking recipe"    cos 1.0  keyword: absent           (sem 2, kw —)
func seedHybridFixtures(t *testing.T, ctx context.Context, storage *KnowledgeBaseStorage) *KnowledgeBase {
	t.Helper()
	kb := createTestKB(t, ctx, storage)

	qVec := unitVector3(1, 0, 0)
	offVec := unitVector3(0.31, 0.95, 0) // cosine 0.31 vs qVec

	docA := createTestDocument(t, ctx, storage, kb.ID, "reset instructions here", "")
	createTestChunk(t, ctx, storage, kb.ID, docA.ID, "reset instructions here", 0, qVec)

	docB := createTestDocument(t, ctx, storage, kb.ID, "password reset guide", "")
	createTestChunk(t, ctx, storage, kb.ID, docB.ID, "password reset guide", 0, offVec)

	docC := createTestDocument(t, ctx, storage, kb.ID, "completely unrelated cooking recipe", "")
	createTestChunk(t, ctx, storage, kb.ID, docC.ID, "completely unrelated cooking recipe", 0, qVec)

	return kb
}

func TestSearchChunksHybrid_RRF(t *testing.T) {
	requireDB(t)
	ctx := aiTestCtx(t)
	storage := NewKnowledgeBaseStorage(aiTestDB)
	kb := seedHybridFixtures(t, ctx, storage)

	query := "password reset guide"
	qVec := unitVector3(1, 0, 0)

	t.Run("keyword-list presence outranks semantic-only under RRF", func(t *testing.T) {
		results, err := storage.SearchChunksHybrid(ctx, kb.ID, HybridSearchOptions{
			Query:          query,
			QueryEmbedding: qVec,
			Limit:          10,
			Mode:           SearchModeHybrid,
			SemanticWeight: 0.5,
			KeywordBoost:   0.3,
		})
		require.NoError(t, err)
		require.Len(t, results, 3, "all three chunks should pass threshold 0 (default)")

		byID := map[string]float64{}
		for _, r := range results {
			byID[r.Content] = r.Similarity
		}

		// Only B matches all query terms (plainto_tsquery ANDs them), so the
		// keyword ranked list contains exactly B at rank 1.
		// B: kw rank 1 + sem rank 3 → (0.5/61 + 0.5/63) * 61 ≈ 0.984
		assert.InDelta(t, 0.9841, byID["password reset guide"], 0.01)
		// A: sem rank 1 only → 0.5
		assert.InDelta(t, 0.5, byID["reset instructions here"], 0.01)
		// C: sem rank 2 only → 0.5 * 61/62 ≈ 0.492
		assert.InDelta(t, 0.4919, byID["completely unrelated cooking recipe"], 0.01)

		assert.Greater(t, byID["password reset guide"], byID["reset instructions here"],
			"RRF must rank keyword-list presence above semantic rank 1")
		assert.Equal(t, "password reset guide", results[0].Content, "highest fused score first")
	})

	t.Run("threshold filters on cosine similarity", func(t *testing.T) {
		results, err := storage.SearchChunksHybrid(ctx, kb.ID, HybridSearchOptions{
			Query:          query,
			QueryEmbedding: qVec,
			Limit:          10,
			Threshold:      0.5,
			Mode:           SearchModeHybrid,
			SemanticWeight: 0.5,
		})
		require.NoError(t, err)
		// B's cosine is 0.31 < 0.5 → dropped even though it would fuse highest.
		for _, r := range results {
			assert.NotEqual(t, "password reset guide", r.Content,
				"cosine below threshold must be filtered before fusion")
		}
		assert.Len(t, results, 2)
	})

	t.Run("user isolation filter scopes results", func(t *testing.T) {
		userVec := unitVector3(0, 1, 0)
		docU := createTestDocument(t, ctx, storage, kb.ID, "private roadmap planning", "user-1")
		createTestChunk(t, ctx, storage, kb.ID, docU.ID, "private roadmap planning", 0, userVec)

		// user-1 with IncludeGlobal=false: only their document.
		strict, err := storage.SearchChunksHybrid(ctx, kb.ID, HybridSearchOptions{
			Query:          "private roadmap planning",
			QueryEmbedding: userVec,
			Limit:          10,
			Mode:           SearchModeHybrid,
			SemanticWeight: 0.5,
			Filter:         &MetadataFilter{UserID: strPtr("user-1"), IncludeGlobal: false},
		})
		require.NoError(t, err)
		require.Len(t, strict, 1)
		assert.Equal(t, "private roadmap planning", strict[0].Content)

		// user-1 querying the public phrase must not see user-1's unrelated
		// private document just because the isolation filter matches.
		loose, err := storage.SearchChunksHybrid(ctx, kb.ID, HybridSearchOptions{
			Query:          "password reset guide",
			QueryEmbedding: qVec,
			Limit:          10,
			Mode:           SearchModeHybrid,
			SemanticWeight: 0.5,
			Filter:         &MetadataFilter{UserID: strPtr("user-1"), IncludeGlobal: true},
		})
		require.NoError(t, err)
		for _, r := range loose {
			assert.NotEqual(t, "private roadmap planning", r.Content,
				"private doc must not leak into a different query's results")
		}
	})

	t.Run("keyword-only mode works without embeddings", func(t *testing.T) {
		results, err := storage.SearchChunksHybrid(ctx, kb.ID, HybridSearchOptions{
			Query: "password reset guide",
			Limit: 10,
			Mode:  SearchModeKeyword,
		})
		require.NoError(t, err)
		require.NotEmpty(t, results)
		assert.Equal(t, "password reset guide", results[0].Content,
			"exact phrase match ranks first in keyword-only mode")
	})
}
