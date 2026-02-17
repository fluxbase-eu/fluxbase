package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuleExtractor_ExtractPersons(t *testing.T) {
	t.Run("extracts person names with titles", func(t *testing.T) {
		extractor := NewRuleBasedExtractor()
		text := "Dr. John Smith and Mr. Jane Doe met with CEO Mary Johnson"

		result, err := extractor.ExtractEntities(text, "kb-test-1")
		require.NoError(t, err)
		assert.NotEmpty(t, result.Entities)

		// Should find at least Dr. John Smith and Mr. Jane Doe
		// (CEO Mary Johnson doesn't match because it's only 1 word after title)
		persons := filterEntitiesByType(result.Entities, EntityPerson)
		assert.GreaterOrEqual(t, len(persons), 2)
	})

	t.Run("extracts capitalized multi-word names", func(t *testing.T) {
		extractor := NewRuleBasedExtractor()
		// Use a name with 3+ capitalized words to match the pattern
		text := "Dr. Martin Luther King Jr. visited Silicon Valley"

		result, err := extractor.ExtractEntities(text, "kb-test-1")
		require.NoError(t, err)

		// Should find Dr. Martin Luther King Jr. (3+ words with title)
		// And Silicon Valley as organization
		assert.NotEmpty(t, result.Entities)

		persons := filterEntitiesByType(result.Entities, EntityPerson)
		assert.GreaterOrEqual(t, len(persons), 1)
	})
}

func TestRuleExtractor_ExtractOrganizations(t *testing.T) {
	t.Run("extracts company names with suffixes", func(t *testing.T) {
		extractor := NewRuleBasedExtractor()
		text := "Apple Inc and Microsoft Corp announced a partnership"

		result, err := extractor.ExtractEntities(text, "kb-test-1")
		require.NoError(t, err)
		assert.NotEmpty(t, result.Entities)

		orgs := filterEntitiesByType(result.Entities, EntityOrganization)
		assert.GreaterOrEqual(t, len(orgs), 2) // Apple Inc, Microsoft Corp
	})

	t.Run("extracts tech company names", func(t *testing.T) {
		extractor := NewRuleBasedExtractor()
		text := "Google and Amazon are leading cloud providers"

		result, err := extractor.ExtractEntities(text, "kb-test-1")
		require.NoError(t, err)
		assert.NotEmpty(t, result.Entities)

		orgs := filterEntitiesByType(result.Entities, EntityOrganization)
		assert.GreaterOrEqual(t, len(orgs), 2) // Google, Amazon
	})
}

func TestRuleExtractor_ExtractLocations(t *testing.T) {
	t.Run("extracts city names", func(t *testing.T) {
		extractor := NewRuleBasedExtractor()
		text := "The conference was held in San Francisco and New York"

		result, err := extractor.ExtractEntities(text, "kb-test-1")
		require.NoError(t, err)
		assert.NotEmpty(t, result.Entities)

		locations := filterEntitiesByType(result.Entities, EntityLocation)
		assert.GreaterOrEqual(t, len(locations), 2) // San Francisco, New York
	})

	t.Run("extracts US states", func(t *testing.T) {
		extractor := NewRuleBasedExtractor()
		text := "The company has offices in California and Texas"

		result, err := extractor.ExtractEntities(text, "kb-test-1")
		require.NoError(t, err)
		assert.NotEmpty(t, result.Entities)

		locations := filterEntitiesByType(result.Entities, EntityLocation)
		assert.GreaterOrEqual(t, len(locations), 2) // California, Texas
	})
}

func TestRuleExtractor_ExtractProducts(t *testing.T) {
	t.Run("extracts product names", func(t *testing.T) {
		extractor := NewRuleBasedExtractor()
		text := "The new iPhone and MacBook Pro were released alongside iOS 18"

		result, err := extractor.ExtractEntities(text, "kb-test-1")
		require.NoError(t, err)
		assert.NotEmpty(t, result.Entities)

		products := filterEntitiesByType(result.Entities, EntityProduct)
		assert.GreaterOrEqual(t, len(products), 3) // iPhone, MacBook Pro, iOS
	})
}

func TestToCanonicalName(t *testing.T) {
	t.Run("converts to title case", func(t *testing.T) {
		assert.Equal(t, "John Smith", toCanonicalName("john smith"))
		assert.Equal(t, "John Smith", toCanonicalName("JOHN SMITH"))
		assert.Equal(t, "John Smith", toCanonicalName("John Smith"))
	})

	t.Run("handles multiple words", func(t *testing.T) {
		assert.Equal(t, "New York", toCanonicalName("new york"))
		assert.Equal(t, "San Francisco", toCanonicalName("san francisco"))
	})

	t.Run("handles leading/trailing whitespace", func(t *testing.T) {
		assert.Equal(t, "Test", toCanonicalName("  test  "))
	})
}

func TestExtractEntitiesWithRelationships(t *testing.T) {
	t.Run("infers works_at relationships", func(t *testing.T) {
		extractor := NewRuleBasedExtractor()
		text := "John Smith works at Google"

		result, err := extractor.ExtractEntitiesWithRelationships(text, "kb-test-1", "doc-1")
		require.NoError(t, err)
		assert.NotEmpty(t, result.Relationships)

		// Should find a works_at relationship
		found := false
		for _, rel := range result.Relationships {
			if rel.RelationshipType == RelWorksAt {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find works_at relationship")
	})

	t.Run("infers founded_by relationships", func(t *testing.T) {
		extractor := NewRuleBasedExtractor()
		text := "Steve Jobs founded Apple Inc"

		result, err := extractor.ExtractEntitiesWithRelationships(text, "kb-test-1", "doc-1")
		require.NoError(t, err)
		assert.NotEmpty(t, result.Relationships)

		// Should find a founded_by relationship
		found := false
		for _, rel := range result.Relationships {
			if rel.RelationshipType == RelFoundedBy {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find founded_by relationship")
	})
}

func TestCreateDocumentEntities(t *testing.T) {
	t.Run("creates document-entity mentions", func(t *testing.T) {
		extractor := NewRuleBasedExtractor()
		text := "Google is a company. Google was founded in California."
		documentID := "doc-test-1"

		result, err := extractor.ExtractEntities(text, "kb-test-1")
		require.NoError(t, err)

		docEntities := extractor.createDocumentEntities(documentID, result.Entities, text)
		assert.NotEmpty(t, docEntities)

		// Google should be mentioned twice
		for _, de := range docEntities {
			if de.EntityID == result.Entities[0].ID {
				assert.Equal(t, 2, de.MentionCount)
			}
		}
	})
}

// Helper function to filter entities by type
func filterEntitiesByType(entities []Entity, entityType EntityType) []Entity {
	var filtered []Entity
	for _, e := range entities {
		if e.EntityType == entityType {
			filtered = append(filtered, e)
		}
	}
	return filtered
}
