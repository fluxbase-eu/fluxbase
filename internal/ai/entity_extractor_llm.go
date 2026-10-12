package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// LLMExtractor extracts entities and relationships using the configured chat
// provider. It implements EntityExtractor and is meant to be selected per
// knowledge base via StrategyEntityExtractor (strategies "llm" and "hybrid").
//
// On any provider or parse failure it falls back to the rule-based extractor
// so document ingestion and query-time graph boosting never break because of
// an LLM outage.
type LLMExtractor struct {
	// resolve returns the provider to use for extraction (typically built from
	// the effective default provider). Resolved per call so admin provider
	// changes take effect without a restart.
	resolve func(ctx context.Context) (Provider, error)

	// model overrides the provider's configured model when non-empty.
	model string

	maxTokens    int           // max_tokens for the extraction request
	minTextChars int           // below this length the LLM is skipped (rules are cheaper and sufficient)
	timeout      time.Duration // per-extraction-request timeout

	fallback *RuleBasedExtractor
}

// LLMExtractorConfig configures an LLMExtractor.
type LLMExtractorConfig struct {
	Resolve      func(ctx context.Context) (Provider, error)
	Model        string        // optional override of the provider's configured model
	MaxTokens    int           // default 2000
	MinTextChars int           // default 200
	Timeout      time.Duration // default 60s
}

// NewLLMExtractor creates an LLM-backed entity extractor. resolve must be
// non-nil; it is the wiring's responsibility to pick the right provider
// source (effective default provider, config-based provider, ...).
func NewLLMExtractor(cfg LLMExtractorConfig) *LLMExtractor {
	if cfg.MaxTokens <= 0 {
		cfg.MaxTokens = 2000
	}
	if cfg.MinTextChars <= 0 {
		cfg.MinTextChars = 200
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 60 * time.Second
	}
	return &LLMExtractor{
		resolve:      cfg.Resolve,
		model:        cfg.Model,
		maxTokens:    cfg.MaxTokens,
		minTextChars: cfg.MinTextChars,
		timeout:      cfg.Timeout,
		fallback:     NewRuleBasedExtractor(),
	}
}

// llmExtractionPayload is the JSON contract the extraction prompt asks for.
type llmExtractionPayload struct {
	Entities []struct {
		Type          string   `json:"type"`
		Name          string   `json:"name"`
		CanonicalName string   `json:"canonical_name"`
		Aliases       []string `json:"aliases"`
	} `json:"entities"`
	Relationships []struct {
		Source    string `json:"source"`
		Target    string `json:"target"`
		Type      string `json:"type"`
		Direction string `json:"direction"`
		Context   string `json:"context"`
	} `json:"relationships"`
}

const entityExtractionSystemPrompt = `You are an entity extraction engine. Extract named entities and relationships between them from the given text.

Respond with ONLY a JSON object (no markdown fences, no prose) of this shape:
{
  "entities": [
    {"type": "<type>", "name": "<exact mention in text>", "canonical_name": "<normalized name>", "aliases": ["<alternative mentions>"]}
  ],
  "relationships": [
    {"source": "<source canonical_name>", "target": "<target canonical_name>", "type": "<rel_type>", "direction": "forward"}
  ]
}

Valid entity types: person, organization, location, concept, product, event, table, url, api_endpoint, datetime, code_reference, error, other.
Valid relationship types: works_at, located_in, founded_by, owns, part_of, related_to, knows, customer_of, supplier_of, invested_in, acquired, merged_with, competitor_of, parent_of, child_of, spouse_of, sibling_of, foreign_key, depends_on, other.
Valid directions: forward, backward, bidirectional.

Rules:
- "source" and "target" of a relationship MUST be canonical_name values of entities in the "entities" array.
- Only extract entities explicitly present in or unambiguously derivable from the text.
- Extract at most 100 entities. Skip trivial or generic mentions.
- If nothing meaningful exists, return {"entities": [], "relationships": []}.`

// ExtractEntities extracts entities from text using the LLM provider.
func (l *LLMExtractor) ExtractEntities(ctx context.Context, text string, kbID string) (*EntityExtractionResult, error) {
	result, _, err := l.extract(ctx, text, kbID, "")
	if err != nil {
		// Never fail the caller: fall back to rules.
		log.Warn().Err(err).Str("kb_id", kbID).Msg("LLM entity extraction failed; falling back to rule-based extractor")
		return l.fallback.ExtractEntities(ctx, text, kbID)
	}
	return result, nil
}

// ExtractEntitiesWithRelationships extracts entities and relationships using
// the LLM provider, with document-entity mention records built the same way
// the rule-based extractor builds them.
func (l *LLMExtractor) ExtractEntitiesWithRelationships(ctx context.Context, text string, kbID string, documentID string) (*EntityExtractionResult, error) {
	result, payloadUsed, err := l.extract(ctx, text, kbID, documentID)
	if err != nil {
		log.Warn().Err(err).Str("kb_id", kbID).Msg("LLM entity extraction failed; falling back to rule-based extractor")
		return l.fallback.ExtractEntitiesWithRelationships(ctx, text, kbID, documentID)
	}
	if !payloadUsed {
		// Extraction short-circuited (text too short for the LLM): the result
		// came from rules without relationships/document entities.
		result.Relationships = l.fallback.inferRelationships(text, kbID, result.Entities)
		result.DocumentEntities = buildDocumentEntities(documentID, result.Entities, text)
	}
	return result, nil
}

// extract runs one extraction. payloadUsed reports whether the returned
// result was built from an LLM response (as opposed to the rule-based
// short-circuit for too-short text).
func (l *LLMExtractor) extract(ctx context.Context, text string, kbID string, documentID string) (*EntityExtractionResult, bool, error) {
	trimmed := strings.TrimSpace(text)
	if len(trimmed) < l.minTextChars {
		ruleResult, err := l.fallback.ExtractEntities(ctx, text, kbID)
		if err != nil {
			return nil, false, err
		}
		return ruleResult, false, nil
	}

	extractCtx, cancel := context.WithTimeout(ctx, l.timeout)
	defer cancel()

	provider, err := l.providerFor(extractCtx)
	if err != nil {
		return nil, false, fmt.Errorf("no provider for entity extraction: %w", err)
	}

	response, err := l.chatWithRetry(extractCtx, provider, trimmed)
	if err != nil {
		return nil, false, fmt.Errorf("LLM extraction request failed: %w", err)
	}

	payload, err := parseExtractionPayload(response)
	if err != nil {
		// One retry: the model may produce valid JSON on a second attempt.
		log.Warn().Err(err).Str("kb_id", kbID).Msg("Failed to parse LLM extraction response, retrying once")
		response, err = l.chatWithRetry(extractCtx, provider, trimmed)
		if err != nil {
			return nil, false, fmt.Errorf("LLM extraction request failed: %w", err)
		}
		payload, err = parseExtractionPayload(response)
		if err != nil {
			return nil, false, fmt.Errorf("failed to parse LLM extraction response: %w", err)
		}
	}

	return l.buildResult(payload, kbID, documentID, text), true, nil
}

// providerFor resolves the provider to use.
func (l *LLMExtractor) providerFor(ctx context.Context) (Provider, error) {
	if l.resolve == nil {
		return nil, fmt.Errorf("no provider resolver configured")
	}
	provider, err := l.resolve(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return nil, fmt.Errorf("no AI provider configured")
	}
	return provider, nil
}

// chatWithRetry sends the extraction request with a single retry for
// transient provider failures (mirrors embedBatchWithRetry semantics).
func (l *LLMExtractor) chatWithRetry(ctx context.Context, provider Provider, text string) (string, error) {
	req := &ChatRequest{
		Model: l.model, // empty = provider's configured default model
		Messages: []Message{
			{Role: RoleSystem, Content: entityExtractionSystemPrompt},
			{Role: RoleUser, Content: truncateForExtraction(text, l.maxTokens*4)},
		},
		MaxTokens:   l.maxTokens,
		Temperature: 0.0, // extraction is deterministic work
	}

	do := func() (string, error) {
		resp, err := provider.Chat(ctx, req)
		if err != nil {
			return "", err
		}
		if resp == nil || len(resp.Choices) == 0 {
			return "", fmt.Errorf("provider returned no choices")
		}
		return resp.Choices[0].Message.Content, nil
	}

	content, err := do()
	if err == nil || !isTransientEmbeddingError(err) {
		return content, err
	}
	log.Warn().Err(err).Msg("Transient LLM failure during entity extraction, retrying once")
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(2 * time.Second):
	}
	return do()
}

// parseExtractionPayload extracts the JSON object from a model response,
// tolerating markdown fences and surrounding prose.
func parseExtractionPayload(content string) (*llmExtractionPayload, error) {
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start < 0 || end <= start {
		return nil, fmt.Errorf("no JSON object in response")
	}
	var payload llmExtractionPayload
	if err := json.Unmarshal([]byte(content[start:end+1]), &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON in response: %w", err)
	}
	return &payload, nil
}

// buildResult converts the parsed payload into an EntityExtractionResult.
// Unknown entity/relationship types are normalized; relationships referencing
// unknown entities are dropped. rawText is the original (untruncated) text,
// used for mention counting.
func (l *LLMExtractor) buildResult(payload *llmExtractionPayload, kbID string, documentID string, rawText string) *EntityExtractionResult {
	result := &EntityExtractionResult{
		Entities:         []Entity{},
		Relationships:    []EntityRelationship{},
		DocumentEntities: []DocumentEntity{},
	}

	byCanonical := make(map[string]*Entity, len(payload.Entities))
	for _, e := range payload.Entities {
		name := strings.TrimSpace(e.Name)
		if name == "" {
			continue
		}
		canonical := strings.TrimSpace(e.CanonicalName)
		if canonical == "" {
			canonical = toCanonicalName(name)
		}
		key := strings.ToLower(canonical)
		if _, dup := byCanonical[key]; dup {
			continue
		}

		entity := Entity{
			ID:              uuid.New().String(),
			KnowledgeBaseID: kbID,
			EntityType:      normalizeLLMEntityType(e.Type),
			Name:            name,
			CanonicalName:   canonical,
			Aliases:         e.Aliases,
			Metadata: map[string]interface{}{
				"extraction_method": "llm",
			},
		}
		if entity.Aliases == nil {
			entity.Aliases = []string{}
		}
		byCanonical[key] = &entity
		result.Entities = append(result.Entities, entity)
	}

	confidence := 0.85
	for _, r := range payload.Relationships {
		sourceKey := strings.ToLower(strings.TrimSpace(r.Source))
		targetKey := strings.ToLower(strings.TrimSpace(r.Target))
		source, ok := byCanonical[sourceKey]
		if !ok {
			continue
		}
		target, ok := byCanonical[targetKey]
		if !ok || target == source {
			continue
		}
		relationship := EntityRelationship{
			ID:               uuid.New().String(),
			KnowledgeBaseID:  kbID,
			SourceEntityID:   source.ID,
			TargetEntityID:   target.ID,
			RelationshipType: normalizeLLMRelationshipType(r.Type),
			Direction:        normalizeLLMRelationshipDirection(r.Direction),
			Confidence:       &confidence,
			Metadata: map[string]interface{}{
				"extraction_method": "llm",
			},
		}
		if r.Context != "" {
			relationship.Metadata["context"] = r.Context
		}
		result.Relationships = append(result.Relationships, relationship)
	}

	result.DocumentEntities = buildDocumentEntities(documentID, result.Entities, rawText)

	return result
}

// normalizeLLMEntityType maps an LLM-provided type to a known EntityType,
// falling back to EntityOther.
func normalizeLLMEntityType(t string) EntityType {
	switch EntityType(strings.ToLower(strings.TrimSpace(t))) {
	case EntityPerson:
		return EntityPerson
	case EntityOrganization:
		return EntityOrganization
	case EntityLocation:
		return EntityLocation
	case EntityConcept:
		return EntityConcept
	case EntityProduct:
		return EntityProduct
	case EntityEvent:
		return EntityEvent
	case EntityTable:
		return EntityTable
	case EntityURL:
		return EntityURL
	case EntityAPIEndpoint:
		return EntityAPIEndpoint
	case EntityDateTime:
		return EntityDateTime
	case EntityCodeReference:
		return EntityCodeReference
	case EntityError:
		return EntityError
	default:
		return EntityOther
	}
}

// normalizeLLMRelationshipType maps an LLM-provided type to a known
// RelationshipType, falling back to RelOther.
func normalizeLLMRelationshipType(t string) RelationshipType {
	switch RelationshipType(strings.ToLower(strings.TrimSpace(t))) {
	case RelWorksAt:
		return RelWorksAt
	case RelLocatedIn:
		return RelLocatedIn
	case RelFoundedBy:
		return RelFoundedBy
	case RelOwns:
		return RelOwns
	case RelPartOf:
		return RelPartOf
	case RelKnows:
		return RelKnows
	case RelCustomerOf:
		return RelCustomerOf
	case RelSupplierOf:
		return RelSupplierOf
	case RelInvestedIn:
		return RelInvestedIn
	case RelAcquired:
		return RelAcquired
	case RelMergedWith:
		return RelMergedWith
	case RelCompetitorOf:
		return RelCompetitorOf
	case RelParentOf:
		return RelParentOf
	case RelChildOf:
		return RelChildOf
	case RelSpouseOf:
		return RelSpouseOf
	case RelSiblingOf:
		return RelSiblingOf
	case RelForeignKey:
		return RelForeignKey
	case RelDependsOn:
		return RelDependsOn
	default:
		return RelOther
	}
}

// normalizeLLMRelationshipDirection maps an LLM-provided direction, defaulting
// to forward.
func normalizeLLMRelationshipDirection(d string) RelationshipDirection {
	switch RelationshipDirection(strings.ToLower(strings.TrimSpace(d))) {
	case DirectionBackward:
		return DirectionBackward
	case DirectionBidirectional:
		return DirectionBidirectional
	default:
		return DirectionForward
	}
}

// truncateForExtraction truncates text to maxChars bytes on a UTF-8 rune
// boundary, keeping the prompt within the token budget. maxChars <= 0 returns
// the text unchanged.
func truncateForExtraction(text string, maxChars int) string {
	if maxChars <= 0 || len(text) <= maxChars {
		return text
	}
	cut := text[:maxChars]
	// Walk back over any incomplete trailing rune.
	for len(cut) > 0 {
		if r, size := utf8.DecodeLastRuneInString(cut); r != utf8.RuneError || size != 1 {
			break
		}
		cut = cut[:len(cut)-1]
	}
	return cut
}
