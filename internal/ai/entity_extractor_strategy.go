package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

// Entity extraction strategies. Configured globally via
// ai.entity_extraction_strategy and overridable per knowledge base via the
// entity_extraction_strategy column (empty value inherits the global default).
const (
	ExtractionStrategyRules  = "rules"  // regex/rule-based only (default)
	ExtractionStrategyLLM    = "llm"    // LLM-backed, falls back to rules on failure
	ExtractionStrategyHybrid = "hybrid" // rules + LLM merged, deduped by canonical name
)

// IsValidExtractionStrategy reports whether s is a usable strategy value
// (empty is valid: it means "inherit the global default").
func IsValidExtractionStrategy(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", ExtractionStrategyRules, ExtractionStrategyLLM, ExtractionStrategyHybrid:
		return true
	default:
		return false
	}
}

// StrategyEntityExtractor dispatches to the rule-based or LLM extractor based
// on the owning knowledge base's entity_extraction_strategy. It implements
// EntityExtractor and is safe for concurrent use.
type StrategyEntityExtractor struct {
	rule            *RuleBasedExtractor
	llm             *LLMExtractor
	storage         *KnowledgeBaseStorage
	defaultStrategy string
}

// NewStrategyEntityExtractor creates the dispatching extractor. llm may be nil
// when no chat provider is configured — "llm"/"hybrid" KBs then degrade to
// rules with a warning instead of failing extraction.
func NewStrategyEntityExtractor(rule *RuleBasedExtractor, llm *LLMExtractor, storage *KnowledgeBaseStorage, defaultStrategy string) *StrategyEntityExtractor {
	if !IsValidExtractionStrategy(defaultStrategy) || defaultStrategy == "" {
		defaultStrategy = ExtractionStrategyRules
	}
	return &StrategyEntityExtractor{
		rule:            rule,
		llm:             llm,
		storage:         storage,
		defaultStrategy: strings.ToLower(strings.TrimSpace(defaultStrategy)),
	}
}

// ExtractEntities extracts entities using the KB's configured strategy.
func (s *StrategyEntityExtractor) ExtractEntities(ctx context.Context, text string, kbID string) (*EntityExtractionResult, error) {
	strategy, err := s.strategyFor(ctx, kbID)
	if err != nil {
		return s.rule.ExtractEntities(ctx, text, kbID)
	}

	switch strategy {
	case ExtractionStrategyLLM:
		if s.llm == nil {
			return s.rule.ExtractEntities(ctx, text, kbID)
		}
		return s.llm.ExtractEntities(ctx, text, kbID)
	case ExtractionStrategyHybrid:
		return s.extractHybrid(ctx, text, kbID)
	default:
		return s.rule.ExtractEntities(ctx, text, kbID)
	}
}

// ExtractEntitiesWithRelationships extracts entities and relationships using
// the KB's configured strategy.
func (s *StrategyEntityExtractor) ExtractEntitiesWithRelationships(ctx context.Context, text string, kbID string, documentID string) (*EntityExtractionResult, error) {
	strategy, err := s.strategyFor(ctx, kbID)
	if err != nil {
		return s.rule.ExtractEntitiesWithRelationships(ctx, text, kbID, documentID)
	}

	switch strategy {
	case ExtractionStrategyLLM:
		if s.llm == nil {
			return s.rule.ExtractEntitiesWithRelationships(ctx, text, kbID, documentID)
		}
		return s.llm.ExtractEntitiesWithRelationships(ctx, text, kbID, documentID)
	case ExtractionStrategyHybrid:
		return s.extractHybridWithRelationships(ctx, text, kbID, documentID)
	default:
		return s.rule.ExtractEntitiesWithRelationships(ctx, text, kbID, documentID)
	}
}

// strategyFor resolves the effective strategy for a KB: the KB column value
// when set, otherwise the global default. Any lookup error falls back to the
// global default (extraction must never fail because of strategy resolution).
func (s *StrategyEntityExtractor) strategyFor(ctx context.Context, kbID string) (string, error) {
	strategy := s.defaultStrategy
	if s.storage != nil {
		kb, err := s.storage.GetKnowledgeBase(ctx, kbID)
		if err != nil {
			log.Warn().Err(err).Str("kb_id", kbID).Msg("Failed to load KB for extraction strategy; using global default")
			return "", err
		}
		if kb != nil {
			kbStrategy := strings.ToLower(strings.TrimSpace(kb.EntityExtractionStrategy))
			if kbStrategy != "" {
				strategy = kbStrategy
			}
		}
	}
	if !IsValidExtractionStrategy(strategy) {
		log.Warn().Str("kb_id", kbID).Str("strategy", strategy).Msg("Unknown entity extraction strategy; using rules")
		strategy = ExtractionStrategyRules
	}
	return strategy, nil
}

// extractHybrid runs both extractors and merges the results, deduplicating
// entities by canonical name (LLM wins over rules for duplicates) and
// relationships by (source, target, type).
func (s *StrategyEntityExtractor) extractHybrid(ctx context.Context, text string, kbID string) (*EntityExtractionResult, error) {
	return s.mergeHybrid(ctx, text, kbID, "", func(ctx context.Context, text, kbID, documentID string) (*EntityExtractionResult, error) {
		return s.extractorForHybrid().ExtractEntities(ctx, text, kbID)
	})
}

func (s *StrategyEntityExtractor) extractHybridWithRelationships(ctx context.Context, text string, kbID string, documentID string) (*EntityExtractionResult, error) {
	return s.mergeHybrid(ctx, text, kbID, documentID, s.extractorForHybrid().ExtractEntitiesWithRelationships)
}

// extractorForHybrid picks which extractor leads the hybrid merge. The LLM
// extractor leads when available because its fallback already wraps rules —
// so an LLM outage still yields rule output for the LLM half of the merge.
func (s *StrategyEntityExtractor) extractorForHybrid() EntityExtractor {
	if s.llm != nil {
		return s.llm
	}
	return s.rule
}

// mergeHybrid runs the lead extractor plus the complementary rule extractor
// and merges the two results.
func (s *StrategyEntityExtractor) mergeHybrid(
	ctx context.Context,
	text string,
	kbID string,
	documentID string,
	lead func(ctx context.Context, text, kbID, documentID string) (*EntityExtractionResult, error),
) (*EntityExtractionResult, error) {
	primary, err := lead(ctx, text, kbID, documentID)
	if err != nil {
		return nil, err
	}

	secondary, err := s.rule.ExtractEntitiesWithRelationships(ctx, text, kbID, documentID)
	if err != nil {
		log.Warn().Err(err).Str("kb_id", kbID).Msg("Rule-based half of hybrid extraction failed; using primary only")
		return primary, nil
	}

	return mergeExtractionResults(primary, secondary, kbID, documentID, text), nil
}

// mergeExtractionResults merges two extraction results. Entities dedupe on
// lowercase canonical name with primary taking precedence; relationships
// dedupe on (source, target, type); document-entity mentions are rebuilt from
// the merged entity set.
func mergeExtractionResults(primary, secondary *EntityExtractionResult, kbID string, documentID string, text string) *EntityExtractionResult {
	merged := &EntityExtractionResult{
		DocumentID:       documentID,
		Entities:         []Entity{},
		Relationships:    []EntityRelationship{},
		DocumentEntities: []DocumentEntity{},
	}
	if primary == nil && secondary == nil {
		return merged
	}

	canonicalKey := func(e Entity) string {
		key := strings.ToLower(e.CanonicalName)
		if key == "" {
			key = strings.ToLower(toCanonicalName(e.Name))
		}
		return key
	}

	// Start from the secondary (rules) set, then let primary (LLM) entities
	// replace duplicates in place and add new ones — primary wins conflicts.
	seenEntities := make(map[string]int, len(secondary.Entities)) // canonical key -> index in merged.Entities
	idMapping := make(map[string]string)                          // pre-merge entity ID -> merged entity ID
	for _, e := range secondary.Entities {
		seenEntities[canonicalKey(e)] = len(merged.Entities)
		idMapping[e.ID] = e.ID
		merged.Entities = append(merged.Entities, e)
	}
	for _, e := range primary.Entities {
		if idx, dup := seenEntities[canonicalKey(e)]; dup {
			merged.Entities[idx] = e
			idMapping[e.ID] = e.ID
			continue
		}
		seenEntities[canonicalKey(e)] = len(merged.Entities)
		idMapping[e.ID] = e.ID
		merged.Entities = append(merged.Entities, e)
	}

	seenRelationships := make(map[string]bool)
	addRelationships := func(results []EntityRelationship) {
		for _, rel := range results {
			sourceID, ok := idMapping[rel.SourceEntityID]
			if !ok {
				continue
			}
			targetID, ok := idMapping[rel.TargetEntityID]
			if !ok || sourceID == targetID {
				continue
			}
			key := fmt.Sprintf("%s|%s|%s", sourceID, targetID, rel.RelationshipType)
			if seenRelationships[key] {
				continue
			}
			seenRelationships[key] = true
			rel.SourceEntityID = sourceID
			rel.TargetEntityID = targetID
			merged.Relationships = append(merged.Relationships, rel)
		}
	}
	addRelationships(primary.Relationships)
	addRelationships(secondary.Relationships)

	merged.DocumentEntities = buildDocumentEntities(documentID, merged.Entities, text)

	return merged
}
