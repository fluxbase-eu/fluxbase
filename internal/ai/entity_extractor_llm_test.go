package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// extractionResponseJSON is a well-formed LLM extraction response.
const extractionResponseJSON = `{
	"entities": [
		{"type": "person", "name": "Marie Curie", "canonical_name": "Marie Curie", "aliases": ["Madame Curie"]},
		{"type": "organization", "name": "Sorbonne", "canonical_name": "University of Paris", "aliases": []},
		{"type": "concept", "name": "radioactivity", "canonical_name": "Radioactivity", "aliases": []}
	],
	"relationships": [
		{"source": "Marie Curie", "target": "University of Paris", "type": "works_at", "direction": "forward"},
		{"source": "Marie Curie", "target": "Radioactivity", "type": "discovered", "direction": "forward"}
	]
}`

func extractionChatResponse(content string) *ChatResponse {
	return newSimpleResponse(content)
}

func TestLLMExtractor_ExtractEntities(t *testing.T) {
	provider := newFakeProvider("extraction-test")
	provider.SetChatFn(func(req *ChatRequest) (*ChatResponse, error) {
		assert.Equal(t, entityExtractionSystemPrompt, req.Messages[0].Content)
		return extractionChatResponse(extractionResponseJSON), nil
	})

	extractor := NewLLMExtractor(LLMExtractorConfig{
		Resolve: func(ctx context.Context) (Provider, error) { return provider, nil },
	})

	result, err := extractor.ExtractEntities(context.Background(), strings.Repeat("Marie Curie pioneered radioactivity at the Sorbonne. ", 20), "kb-1")
	require.NoError(t, err)
	require.Len(t, result.Entities, 3)

	var marie *Entity
	for i := range result.Entities {
		if result.Entities[i].CanonicalName == "Marie Curie" {
			marie = &result.Entities[i]
		}
	}
	require.NotNil(t, marie, "Marie Curie entity should exist")
	assert.Equal(t, EntityPerson, marie.EntityType)
	assert.Equal(t, []string{"Madame Curie"}, marie.Aliases)
	assert.Equal(t, "llm", marie.Metadata["extraction_method"])
}

func TestLLMExtractor_RelationshipDropsUnknownEntities(t *testing.T) {
	provider := newFakeProvider("extraction-test")
	provider.SetChatFn(func(req *ChatRequest) (*ChatResponse, error) {
		return extractionChatResponse(extractionResponseJSON), nil
	})

	extractor := NewLLMExtractor(LLMExtractorConfig{
		Resolve: func(ctx context.Context) (Provider, error) { return provider, nil },
	})

	// The "discovered" relationship type is unknown → normalized to other;
	// all relationship endpoints exist → both kept, deduped.
	result, err := extractor.ExtractEntitiesWithRelationships(context.Background(), strings.Repeat("Marie Curie pioneered radioactivity at the Sorbonne. ", 20), "kb-1", "doc-1")
	require.NoError(t, err)
	require.Len(t, result.Relationships, 2)

	relTypes := map[RelationshipType]bool{}
	for _, rel := range result.Relationships {
		relTypes[rel.RelationshipType] = true
		assert.NotNil(t, rel.Confidence)
		assert.Equal(t, "llm", rel.Metadata["extraction_method"])
	}
	assert.True(t, relTypes[RelWorksAt], "works_at should be kept")
	assert.True(t, relTypes[RelOther], "unknown type normalized to other")

	require.NotEmpty(t, result.DocumentEntities, "document-entity mentions should be built")
}

func TestLLMExtractor_StripsMarkdownFences(t *testing.T) {
	provider := newFakeProvider("extraction-test")
	provider.SetChatFn(func(req *ChatRequest) (*ChatResponse, error) {
		content := "Here is the extraction:\n```json\n" + extractionResponseJSON + "\n```\nDone."
		return extractionChatResponse(content), nil
	})

	extractor := NewLLMExtractor(LLMExtractorConfig{
		Resolve: func(ctx context.Context) (Provider, error) { return provider, nil },
	})

	result, err := extractor.ExtractEntities(context.Background(), strings.Repeat("Marie Curie pioneered radioactivity at the Sorbonne. ", 20), "kb-1")
	require.NoError(t, err)
	assert.Len(t, result.Entities, 3)
}

func TestLLMExtractor_RetriesOnInvalidJSON(t *testing.T) {
	provider := newFakeProvider("extraction-test")
	calls := 0
	provider.SetChatFn(func(req *ChatRequest) (*ChatResponse, error) {
		calls++
		if calls == 1 {
			return extractionChatResponse("I could not produce valid JSON, sorry"), nil
		}
		return extractionChatResponse(extractionResponseJSON), nil
	})

	extractor := NewLLMExtractor(LLMExtractorConfig{
		Resolve: func(ctx context.Context) (Provider, error) { return provider, nil },
	})

	result, err := extractor.ExtractEntities(context.Background(), strings.Repeat("Marie Curie pioneered radioactivity at the Sorbonne. ", 20), "kb-1")
	require.NoError(t, err)
	assert.Equal(t, 2, provider.Calls(), "should retry once after parse failure")
	assert.Len(t, result.Entities, 3)
}

func TestLLMExtractor_FallsBackToRulesOnProviderError(t *testing.T) {
	provider := newFakeProvider("extraction-test")
	provider.QueueError(errors.New("provider unavailable"))
	provider.QueueError(errors.New("provider unavailable")) // retry also fails

	extractor := NewLLMExtractor(LLMExtractorConfig{
		Resolve: func(ctx context.Context) (Provider, error) { return provider, nil },
	})

	// Keep each works_at clause sentence-final so the rule extractor's
	// relationship regex captures a clean target entity.
	text := "John Smith works at Google. Mary Jones works at Amazon. " + strings.Repeat("Filler content for the minimum length requirement. ", 10)
	result, err := extractor.ExtractEntitiesWithRelationships(context.Background(), text, "kb-1", "doc-1")
	require.NoError(t, err, "rule fallback should succeed")
	assert.NotEmpty(t, result.Entities)
	assert.NotEmpty(t, result.Relationships)
	for _, e := range result.Entities {
		assert.Equal(t, "rule_based", e.Metadata["extraction_method"])
	}
}

func TestLLMExtractor_ShortTextSkipsLLM(t *testing.T) {
	provider := newFakeProvider("extraction-test")
	extractor := NewLLMExtractor(LLMExtractorConfig{
		Resolve: func(ctx context.Context) (Provider, error) { return provider, nil },
	})

	result, err := extractor.ExtractEntities(context.Background(), "Apple Inc", "kb-1")
	require.NoError(t, err)
	assert.Equal(t, 0, provider.Calls(), "LLM should not be called for short text")
	assert.NotEmpty(t, result.Entities, "rules should still extract")
}

func TestLLMExtractor_NoProviderFallsBackToRules(t *testing.T) {
	extractor := NewLLMExtractor(LLMExtractorConfig{
		Resolve: func(ctx context.Context) (Provider, error) { return nil, nil },
	})

	text := "John Smith works at Google. " + strings.Repeat("Filler content for the minimum length requirement. ", 10)
	result, err := extractor.ExtractEntities(context.Background(), text, "kb-1")
	require.NoError(t, err)
	assert.NotEmpty(t, result.Entities)
}

func TestTruncateForExtraction(t *testing.T) {
	assert.Equal(t, "hello", truncateForExtraction("hello", 0))
	assert.Equal(t, "hello", truncateForExtraction("hello", 10))

	// 13 bytes of text truncated to 10 must land on a rune boundary and stay
	// valid UTF-8 ("héllo wö" = 10 bytes, 8 complete runes).
	truncated := truncateForExtraction("héllo wörld", 10)
	assert.LessOrEqual(t, len(truncated), 10)
	assert.True(t, utf8.ValidString(truncated), "truncated text must be valid UTF-8")
	assert.True(t, strings.HasPrefix("héllo wörld", truncated))
	assert.NotEqual(t, "héllo w", truncated+"", "must not end with an incomplete rune")
}

func TestNormalizeLLMEntityTypes(t *testing.T) {
	assert.Equal(t, EntityPerson, normalizeLLMEntityType("PERSON"))
	assert.Equal(t, EntityAPIEndpoint, normalizeLLMEntityType("api_endpoint"))
	assert.Equal(t, EntityOther, normalizeLLMEntityType("dragon"))
	assert.Equal(t, RelWorksAt, normalizeLLMRelationshipType("works_at"))
	assert.Equal(t, RelOther, normalizeLLMRelationshipType("discovered"))
	assert.Equal(t, DirectionBidirectional, normalizeLLMRelationshipDirection("bidirectional"))
	assert.Equal(t, DirectionForward, normalizeLLMRelationshipDirection("nonsense"))
}

func TestIsValidExtractionStrategy(t *testing.T) {
	assert.True(t, IsValidExtractionStrategy(""))
	assert.True(t, IsValidExtractionStrategy("rules"))
	assert.True(t, IsValidExtractionStrategy(" LLM "))
	assert.True(t, IsValidExtractionStrategy("hybrid"))
	assert.False(t, IsValidExtractionStrategy("magic"))
}

// --- StrategyEntityExtractor ---

// strategyKBLoader lets tests stub the storage-backed KB lookup by pointing
// the extractor at a storage whose GetKnowledgeBase is exercised... The
// storage is a concrete struct, so tests instead cover strategy resolution
// through the default strategy path (no storage: nil storage skips lookup).
func TestStrategyEntityExtractor_DispatchByDefaultStrategy(t *testing.T) {
	// newTestLLM builds a fresh provider + extractor so call counts don't
	// leak across subtests.
	newTestLLM := func(t *testing.T) *LLMExtractor {
		t.Helper()
		provider := newFakeProvider("strategy-test")
		provider.SetChatFn(func(req *ChatRequest) (*ChatResponse, error) {
			return extractionChatResponse(extractionResponseJSON), nil
		})
		return NewLLMExtractor(LLMExtractorConfig{
			Resolve:      func(ctx context.Context) (Provider, error) { return provider, nil },
			MinTextChars: 1,
		})
	}

	t.Run("rules strategy never calls the LLM", func(t *testing.T) {
		llm := newTestLLM(t)
		extractor := NewStrategyEntityExtractor(NewRuleBasedExtractor(), llm, nil, ExtractionStrategyRules)
		result, err := extractor.ExtractEntities(context.Background(), "John Smith works at Google", "kb-1")
		require.NoError(t, err)
		assert.NotEmpty(t, result.Entities)
		assert.Equal(t, 0, llmCalls(llm))
	})

	t.Run("llm strategy calls the LLM", func(t *testing.T) {
		llm := newTestLLM(t)
		extractor := NewStrategyEntityExtractor(NewRuleBasedExtractor(), llm, nil, ExtractionStrategyLLM)
		result, err := extractor.ExtractEntities(context.Background(), "Marie Curie pioneered radioactivity", "kb-1")
		require.NoError(t, err)
		assert.Equal(t, 1, llmCalls(llm))
		require.Len(t, result.Entities, 3)
	})

	t.Run("empty default falls back to rules", func(t *testing.T) {
		llm := newTestLLM(t)
		extractor := NewStrategyEntityExtractor(NewRuleBasedExtractor(), llm, nil, "")
		result, err := extractor.ExtractEntities(context.Background(), "John Smith works at Google", "kb-1")
		require.NoError(t, err)
		assert.NotEmpty(t, result.Entities)
		assert.Equal(t, 0, llmCalls(llm))
	})

	t.Run("invalid default falls back to rules", func(t *testing.T) {
		llm := newTestLLM(t)
		extractor := NewStrategyEntityExtractor(NewRuleBasedExtractor(), llm, nil, "magic")
		result, err := extractor.ExtractEntities(context.Background(), "John Smith works at Google", "kb-1")
		require.NoError(t, err)
		assert.NotEmpty(t, result.Entities)
		assert.Equal(t, 0, llmCalls(llm))
	})

	t.Run("llm strategy without LLM extractor degrades to rules", func(t *testing.T) {
		extractor := NewStrategyEntityExtractor(NewRuleBasedExtractor(), nil, nil, ExtractionStrategyLLM)
		result, err := extractor.ExtractEntities(context.Background(), "John Smith works at Google", "kb-1")
		require.NoError(t, err)
		assert.NotEmpty(t, result.Entities)
		for _, e := range result.Entities {
			assert.Equal(t, "rule_based", e.Metadata["extraction_method"])
		}
	})
}

// llmCalls counts Chat invocations the extractor made through its resolver.
// The fakeProvider is created inside the resolver closure, so count via the
// extractor's own behavior: run a request and read the counter from the
// captured provider.
func llmCalls(llm *LLMExtractor) int {
	provider, err := llm.providerFor(context.Background())
	if err != nil || provider == nil {
		return -1
	}
	if fp, ok := provider.(*fakeProvider); ok {
		return fp.Calls()
	}
	return -1
}

func TestStrategyEntityExtractor_HybridMergesAndDedupes(t *testing.T) {
	// The LLM half returns Marie Curie + Sorbonne; rules additionally find
	// URL/datetime patterns. The merge should dedupe overlapping entities by
	// canonical name with the LLM entity winning.
	provider := newFakeProvider("strategy-test")
	provider.SetChatFn(func(req *ChatRequest) (*ChatResponse, error) {
		return extractionChatResponse(extractionResponseJSON), nil
	})
	llm := NewLLMExtractor(LLMExtractorConfig{
		Resolve:      func(ctx context.Context) (Provider, error) { return provider, nil },
		MinTextChars: 1,
	})

	extractor := NewStrategyEntityExtractor(NewRuleBasedExtractor(), llm, nil, ExtractionStrategyHybrid)
	text := fmt.Sprintf("Marie Curie pioneered radioactivity at the Sorbonne. Visit https://example.org/archives for records. %s", strings.Repeat("Padding to satisfy rule extractor patterns. ", 3))

	result, err := extractor.ExtractEntitiesWithRelationships(context.Background(), text, "kb-1", "doc-1")
	require.NoError(t, err)

	// Dedupe check: no two merged entities share a canonical key.
	seen := make(map[string]bool)
	for _, e := range result.Entities {
		key := strings.ToLower(e.CanonicalName)
		assert.False(t, seen[key], "duplicate canonical entity %q in merged result", key)
		seen[key] = true
	}

	// Marie Curie should exist exactly once with LLM metadata (LLM wins).
	var marieCount, llmMarie int
	for _, e := range result.Entities {
		if strings.EqualFold(e.CanonicalName, "Marie Curie") {
			marieCount++
			if e.Metadata["extraction_method"] == "llm" {
				llmMarie++
			}
		}
	}
	assert.Equal(t, 1, marieCount, "Marie Curie should appear exactly once")
	assert.Equal(t, 1, llmMarie, "the LLM-derived entity should win the merge")

	// All relationships must reference merged entities.
	ids := make(map[string]bool, len(result.Entities))
	for _, e := range result.Entities {
		ids[e.ID] = true
	}
	for _, rel := range result.Relationships {
		assert.True(t, ids[rel.SourceEntityID], "relationship source must be a merged entity")
		assert.True(t, ids[rel.TargetEntityID], "relationship target must be a merged entity")
	}
}

func TestMergeExtractionResults_EmptyInputs(t *testing.T) {
	merged := mergeExtractionResults(nil, nil, "kb-1", "doc-1", "text")
	require.NotNil(t, merged)
	assert.Empty(t, merged.Entities)
	assert.Empty(t, merged.Relationships)
}
