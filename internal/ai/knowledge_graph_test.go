package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKnowledgeGraph_AddEntity(t *testing.T) {
	t.Run("adds entity successfully", func(t *testing.T) {
		entity := &Entity{
			ID:              "entity-1",
			KnowledgeBaseID: "kb-1",
			EntityType:      EntityPerson,
			Name:            "John Smith",
			CanonicalName:   "John Smith",
			Aliases:         []string{"Johnny", "J. Smith"},
			Metadata:        map[string]interface{}{"test": true},
		}

		// Verify entity structure
		assert.Equal(t, "entity-1", entity.ID)
		assert.Equal(t, EntityPerson, entity.EntityType)
		assert.Equal(t, "John Smith", entity.Name)
		assert.Equal(t, 2, len(entity.Aliases))
		assert.True(t, entity.Metadata["test"].(bool))
	})
}

func TestKnowledgeGraph_AddRelationship(t *testing.T) {
	t.Run("creates relationship structure", func(t *testing.T) {
		confidence := 0.9
		rel := &EntityRelationship{
			ID:               "rel-1",
			KnowledgeBaseID:  "kb-1",
			SourceEntityID:   "entity-1",
			TargetEntityID:   "entity-2",
			RelationshipType: RelWorksAt,
			Direction:        DirectionForward,
			Confidence:       &confidence,
			Metadata:         map[string]interface{}{"inferred": true},
		}

		// Verify relationship structure
		assert.Equal(t, "rel-1", rel.ID)
		assert.Equal(t, RelWorksAt, rel.RelationshipType)
		assert.Equal(t, DirectionForward, rel.Direction)
		assert.NotNil(t, rel.Confidence)
		assert.Equal(t, 0.9, *rel.Confidence)
		assert.True(t, rel.Metadata["inferred"].(bool))
	})
}

func TestKnowledgeGraph_RelationshipTypes(t *testing.T) {
	t.Run("all relationship types are defined", func(t *testing.T) {
		// Verify all relationship type constants are defined
		assert.Equal(t, RelationshipType("works_at"), RelWorksAt)
		assert.Equal(t, RelationshipType("located_in"), RelLocatedIn)
		assert.Equal(t, RelationshipType("founded_by"), RelFoundedBy)
		assert.Equal(t, RelationshipType("owns"), RelOwns)
		assert.Equal(t, RelationshipType("part_of"), RelPartOf)
		assert.Equal(t, RelationshipType("related_to"), RelRelatedTo)
		assert.Equal(t, RelationshipType("knows"), RelKnows)
		assert.Equal(t, RelationshipType("customer_of"), RelCustomerOf)
		assert.Equal(t, RelationshipType("supplier_of"), RelSupplierOf)
		assert.Equal(t, RelationshipType("invested_in"), RelInvestedIn)
		assert.Equal(t, RelationshipType("acquired"), RelAcquired)
		assert.Equal(t, RelationshipType("merged_with"), RelMergedWith)
		assert.Equal(t, RelationshipType("competitor_of"), RelCompetitorOf)
		assert.Equal(t, RelationshipType("parent_of"), RelParentOf)
		assert.Equal(t, RelationshipType("child_of"), RelChildOf)
		assert.Equal(t, RelationshipType("spouse_of"), RelSpouseOf)
		assert.Equal(t, RelationshipType("sibling_of"), RelSiblingOf)
		assert.Equal(t, RelationshipType("other"), RelOther)
	})
}

func TestKnowledgeGraph_EntityTypes(t *testing.T) {
	t.Run("all entity types are defined", func(t *testing.T) {
		// Verify all entity type constants are defined
		assert.Equal(t, EntityType("person"), EntityPerson)
		assert.Equal(t, EntityType("organization"), EntityOrganization)
		assert.Equal(t, EntityType("location"), EntityLocation)
		assert.Equal(t, EntityType("concept"), EntityConcept)
		assert.Equal(t, EntityType("product"), EntityProduct)
		assert.Equal(t, EntityType("event"), EntityEvent)
		assert.Equal(t, EntityType("other"), EntityOther)
	})
}

func TestKnowledgeGraph_DirectionTypes(t *testing.T) {
	t.Run("all direction types are defined", func(t *testing.T) {
		assert.Equal(t, RelationshipDirection("forward"), DirectionForward)
		assert.Equal(t, RelationshipDirection("backward"), DirectionBackward)
		assert.Equal(t, RelationshipDirection("bidirectional"), DirectionBidirectional)
	})
}

func TestKnowledgeGraph_FindRelatedEntities_Struct(t *testing.T) {
	t.Run("related entity structure is correct", func(t *testing.T) {
		related := &RelatedEntity{
			EntityID:         "entity-2",
			EntityType:       "organization",
			Name:             "Google",
			CanonicalName:    "Google",
			RelationshipType: "works_at",
			Depth:            1,
			Path:             []string{"entity-1", "entity-2"},
		}

		assert.Equal(t, "entity-2", related.EntityID)
		assert.Equal(t, "organization", related.EntityType)
		assert.Equal(t, "Google", related.Name)
		assert.Equal(t, "works_at", related.RelationshipType)
		assert.Equal(t, 1, related.Depth)
		assert.Equal(t, 2, len(related.Path))
		assert.Equal(t, "entity-1", related.Path[0])
		assert.Equal(t, "entity-2", related.Path[1])
	})
}

func TestKnowledgeGraph_EntityExtractionResult_Struct(t *testing.T) {
	t.Run("extraction result structure is correct", func(t *testing.T) {
		result := &EntityExtractionResult{
			DocumentID: "doc-1",
			Entities: []Entity{
				{
					ID:              "entity-1",
					KnowledgeBaseID: "kb-1",
					EntityType:      EntityPerson,
					Name:            "John Smith",
					CanonicalName:   "John Smith",
				},
			},
			Relationships: []EntityRelationship{
				{
					ID:               "rel-1",
					KnowledgeBaseID:  "kb-1",
					SourceEntityID:   "entity-1",
					TargetEntityID:   "entity-2",
					RelationshipType: RelWorksAt,
				},
			},
			DocumentEntities: []DocumentEntity{
				{
					ID:           "de-1",
					DocumentID:   "doc-1",
					EntityID:     "entity-1",
					MentionCount: 3,
					Salience:     0.8,
					Context:      "...context...",
				},
			},
		}

		assert.Equal(t, "doc-1", result.DocumentID)
		assert.Len(t, result.Entities, 1)
		assert.Len(t, result.Relationships, 1)
		assert.Len(t, result.DocumentEntities, 1)
		assert.Equal(t, EntityPerson, result.Entities[0].EntityType)
		assert.Equal(t, RelWorksAt, result.Relationships[0].RelationshipType)
		assert.Equal(t, 3, result.DocumentEntities[0].MentionCount)
	})
}

func TestKnowledgeGraph_DocumentEntity_Struct(t *testing.T) {
	t.Run("document entity structure is correct", func(t *testing.T) {
		offset := 100
		docEntity := &DocumentEntity{
			ID:                 "de-1",
			DocumentID:         "doc-1",
			EntityID:           "entity-1",
			MentionCount:       5,
			FirstMentionOffset: &offset,
			Salience:           0.9,
			Context:            "...John Smith worked at Google...",
		}

		assert.Equal(t, "de-1", docEntity.ID)
		assert.Equal(t, "doc-1", docEntity.DocumentID)
		assert.Equal(t, "entity-1", docEntity.EntityID)
		assert.Equal(t, 5, docEntity.MentionCount)
		assert.Equal(t, 100, *docEntity.FirstMentionOffset)
		assert.Equal(t, 0.9, docEntity.Salience)
		assert.Contains(t, docEntity.Context, "John Smith")
	})
}
