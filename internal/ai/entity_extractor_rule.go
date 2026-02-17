package ai

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/google/uuid"
)

// EntityExtractor extracts entities from text
type EntityExtractor interface {
	ExtractEntities(text string, kbID string) (*EntityExtractionResult, error)
}

// RuleBasedExtractor uses regex patterns and rules to extract entities
type RuleBasedExtractor struct {
	patterns map[EntityType][]*regexp.Regexp
}

// NewRuleBasedExtractor creates a new rule-based entity extractor
func NewRuleBasedExtractor() *RuleBasedExtractor {
	return &RuleBasedExtractor{
		patterns: buildEntityPatterns(),
	}
}

// buildEntityPatterns builds regex patterns for entity extraction
func buildEntityPatterns() map[EntityType][]*regexp.Regexp {
	return map[EntityType][]*regexp.Regexp{
		EntityPerson: {
			// Names with title prefixes (with optional period for Mr, Mrs, Ms, Dr, Prof)
			regexp.MustCompile(`\b(?:Mr\.?|Mrs\.?|Ms\.?|Dr\.?|Prof\.?|CEO|CTO|CFO|President|VP)\s+[A-Z][a-z]+\s+[A-Z][a-z]+\b`),
			// Two-word capitalized names (e.g., "John Smith")
			regexp.MustCompile(`\b[A-Z][a-z]+\s+[A-Z][a-z]+\b`),
			// Multi-word capitalized names (3+ words, each starting with capital)
			regexp.MustCompile(`\b[A-Z][a-z]+\s+(?:[A-Z][a-z]+\s+)+[A-Z][a-z]+\b`),
		},
		EntityOrganization: {
			// Company names with suffixes
			regexp.MustCompile(`\b[A-Z][a-zA-Z0-9]*(?:Inc|Corp|LLC|Ltd|PLC|GmbH|AG|SA|S\.A\.|Pty|Ltd|Co|Company|Technologies|Industries|Labs|Studios)\b`),
			// Tech companies
			regexp.MustCompile(`\b(?:Google|Apple|Microsoft|Amazon|Meta|Tesla|Netflix|Twitter|Facebook|Intel|AMD|NVIDIA|IBM|Oracle|Salesforce|Adobe|Cisco|VMware|Red Hat|Canonical|Spotify|Uber|Lyft|Airbnb|DoorDash|Stripe|Square|PayPal)\b`),
		},
		EntityLocation: {
			// Cities and countries
			regexp.MustCompile(`\b(?:New York|Los Angeles|San Francisco|Chicago|Houston|Phoenix|Philadelphia|San Antonio|San Diego|Dallas|San Jose|Austin|Jacksonville|Fort Worth|Columbus|Charlotte|San Francisco|Seattle|Denver|Washington|Boston|El Paso|Nashville|Detroit|Portland|Memphis|Las Vegas|Baltimore|Louisville|Milwaukee|Albuquerque|Tucson|Fresno|Sacramento|Kansas City|Mesa|Atlanta|Omaha|Colorado Springs|Raleigh|Miami|Long Beach|Virginia Beach|Oakland|Minneapolis|Tulsa|Arlington|New Orleans|Wichita|Cleveland|Tampa|Bakersfield|Aurora|Anaheim|Honolulu|Santa Ana|Riverside|Corpus Christi|Lexington|Stockton|St. Louis|Saint Paul|Henderson|Pittsburgh|Cincinnati|Anchorage|Greensboro|Plano|Newark|Lincoln|Orlando|Irvine|Chandler|Fort Wayne|Jersey City|Buffalo|Gilbert|Chesapeake|Toledo|Madison|Durham|St. Petersburg|Laredo|Lubbock|Winston Salem|Glendale|Norfolk|Garland|Hialeah|Reno|Chula Vista|Scottsdale|North Las Vegas|Baton Rouge|Cedar Rapids|Worcester|Frisco|Irving|Fremont|Richmond|McKinney|El Cajon|Brownsville|Vancouver|Toronto|Montreal|Calgary|Edmonton|Ottawa|Winnipeg|Quebec City|Hamilton|Kitchener|London|Paris|Berlin|Tokyo|Sydney|Melbourne|Amsterdam|Rome|Madrid|Barcelona|Munich|Milan|Vienna|Zurich|Stockholm|Oslo|Copenhagen|Helsinki|Dublin|Brussels|Warsaw|Prague|Budapest|Bucharest|Athens|Lisbon|Moscow|St. Petersburg|Kiev|Warsaw|Sofia|Belgrade|Zagreb|Ljubljana|Bratislava|Vilnius|Riga|Tallinn|Reykjavik)\b`),
			// US States
			regexp.MustCompile(`\b(?:Alabama|Alaska|Arizona|Arkansas|California|Colorado|Connecticut|Delaware|Florida|Georgia|Hawaii|Idaho|Illinois|Indiana|Iowa|Kansas|Kentucky|Louisiana|Maine|Maryland|Massachusetts|Michigan|Minnesota|Mississippi|Missouri|Montana|Nebraska|Nevada|New Hampshire|New Jersey|New Mexico|New York|North Carolina|North Dakota|Ohio|Oklahoma|Oregon|Pennsylvania|Rhode Island|South Carolina|South Dakota|Tennessee|Texas|Utah|Vermont|Virginia|Washington|West Virginia|Wisconsin|Wyoming)\b`),
		},
		EntityProduct: {
			// Products (common tech products)
			regexp.MustCompile(`\b(?:iPhone|iPad|MacBook|iMac|AirPods|Apple Watch|Galaxy|Pixel|Surface|Kindle|Echo|Alexa|Nest|PlayStation|Xbox|Nintendo|Switch|Android|iOS|Windows|macOS|Linux|Ubuntu|Debian|Red Hat|CentOS|Docker|Kubernetes|AWS|Azure|GCP|Heroku|Vercel|Netlify|Cloudflare)\b`),
		},
	}
}

// ExtractEntities extracts entities from text using rule-based patterns
func (r *RuleBasedExtractor) ExtractEntities(text string, kbID string) (*EntityExtractionResult, error) {
	result := &EntityExtractionResult{
		Entities:   []Entity{},
		Relationships: []EntityRelationship{},
		DocumentEntities: []DocumentEntity{},
	}

	// Track seen entities to avoid duplicates
	seen := make(map[string]bool)

	// Extract entities for each type
	for entityType, patterns := range r.patterns {
		for _, pattern := range patterns {
			matches := pattern.FindAllStringIndex(text, -1)
			for _, match := range matches {
				// Extract the matched text
				if match[0] >= len(text) || match[1] > len(text) {
					continue
				}
				matchText := text[match[0]:match[1]]

				// Create canonical name (normalized)
				canonicalName := toCanonicalName(matchText)

				// Skip if already seen
				key := string(entityType) + ":" + canonicalName
				if seen[key] {
					continue
				}
				seen[key] = true

				entity := Entity{
					ID:             uuid.New().String(),
					KnowledgeBaseID: kbID,
					EntityType:     entityType,
					Name:           matchText,
					CanonicalName:  canonicalName,
					Aliases:        []string{},
					Metadata:       map[string]interface{}{
						"extraction_method": "rule_based",
						"pattern": pattern.String(),
					},
				}

				result.Entities = append(result.Entities, entity)
			}
		}
	}

	return result, nil
}

// toCanonicalName converts a name to canonical form
func toCanonicalName(name string) string {
	// Trim whitespace
	name = strings.TrimSpace(name)

	// Convert to title case (first letter of each word capitalized)
	words := strings.Fields(name)
	for i, word := range words {
		if len(word) > 0 {
			// Capitalize first letter, lowercase the rest
			runes := []rune(word)
			runes[0] = unicode.ToUpper(runes[0])
			for j := 1; j < len(runes); j++ {
				runes[j] = unicode.ToLower(runes[j])
			}
			words[i] = string(runes)
		}
	}

	return strings.Join(words, " ")
}

// ExtractEntitiesWithRelationships extracts entities and attempts to find relationships
func (r *RuleBasedExtractor) ExtractEntitiesWithRelationships(text string, kbID string, documentID string) (*EntityExtractionResult, error) {
	result, err := r.ExtractEntities(text, kbID)
	if err != nil {
		return nil, err
	}

	// Find simple relationships based on co-occurrence and patterns
	result.Relationships = r.inferRelationships(text, kbID, result.Entities)

	// Create document-entity mentions
	result.DocumentEntities = r.createDocumentEntities(documentID, result.Entities, text)

	return result, nil
}

// inferRelationships infers relationships between entities based on text patterns
func (r *RuleBasedExtractor) inferRelationships(text string, kbID string, entities []Entity) []EntityRelationship {
	var relationships []EntityRelationship

	// Build entity lookup by name
	entityMap := make(map[string]*Entity)
	for i := range entities {
		entityMap[entities[i].CanonicalName] = &entities[i]
		entityMap[entities[i].Name] = &entities[i]
	}

	// Pattern: "X works at Y" or "X is employed by Y"
	worksAtPattern := regexp.MustCompile(`(?i)\b([A-Z][a-zA-Z]+(?:\s+[A-Z][a-zA-Z]+)?)\s+(?:works at|is employed by|joined|works for)\s+([A-Z][a-zA-Z]+(?:\s+[A-Z][a-zA-Z]+)?)\b`)
	for _, match := range worksAtPattern.FindAllStringSubmatch(text, -1) {
		if len(match) >= 3 {
			sourceName := toCanonicalName(match[1])
			targetName := toCanonicalName(match[2])

			if source, ok := entityMap[sourceName]; ok {
				if target, ok := entityMap[targetName]; ok {
					relationships = append(relationships, EntityRelationship{
						ID:               uuid.New().String(),
						KnowledgeBaseID:  kbID,
						SourceEntityID:   source.ID,
						TargetEntityID:   target.ID,
						RelationshipType: RelWorksAt,
						Direction:        DirectionForward,
						Metadata: map[string]interface{}{
							"extraction_method": "rule_based",
							"context":            match[0],
						},
					})
				}
			}
		}
	}

	// Pattern: "X founded Y" or "X co-founded Y"
	foundedPattern := regexp.MustCompile(`(?i)\b([A-Z][a-zA-Z]+(?:\s+[A-Z][a-zA-Z]+)?)\s+(?:founded|co-founded|cofounded|started|created)\s+([A-Z][a-zA-Z]+(?:\s+[A-Z][a-zA-Z]+)?)\b`)
	for _, match := range foundedPattern.FindAllStringSubmatch(text, -1) {
		if len(match) >= 3 {
			sourceName := toCanonicalName(match[1])
			targetName := toCanonicalName(match[2])

			if source, ok := entityMap[sourceName]; ok {
				if target, ok := entityMap[targetName]; ok {
					relationships = append(relationships, EntityRelationship{
						ID:               uuid.New().String(),
						KnowledgeBaseID:  kbID,
						SourceEntityID:   source.ID,
						TargetEntityID:   target.ID,
						RelationshipType: RelFoundedBy,
						Direction:        DirectionForward,
						Metadata: map[string]interface{}{
							"extraction_method": "rule_based",
							"context":            match[0],
						},
					})
				}
			}
		}
	}

	return relationships
}

// createDocumentEntities creates document-entity mention records
func (r *RuleBasedExtractor) createDocumentEntities(documentID string, entities []Entity, text string) []DocumentEntity {
	var docEntities []DocumentEntity

	for _, entity := range entities {
		// Count mentions in text
		mentionCount := strings.Count(text, entity.Name)
		if entity.CanonicalName != "" && entity.CanonicalName != entity.Name {
			mentionCount += strings.Count(text, entity.CanonicalName)
		}

		// Find first mention offset
		firstOffset := strings.Index(text, entity.Name)
		if firstOffset < 0 && entity.CanonicalName != "" {
			firstOffset = strings.Index(text, entity.CanonicalName)
		}

		// Extract context around first mention
		context := ""
		if firstOffset >= 0 {
			start := firstOffset - 50
			end := firstOffset + len(entity.Name) + 50
			if start < 0 {
				start = 0
			}
			if end > len(text) {
				end = len(text)
			}
			context = "..." + text[start:end] + "..."
		}

		docEntities = append(docEntities, DocumentEntity{
			ID:                uuid.New().String(),
			DocumentID:        documentID,
			EntityID:          entity.ID,
			MentionCount:      mentionCount,
			FirstMentionOffset: &firstOffset,
			Salience:          0.5, // Default salience, could be calculated based on frequency/position
			Context:           context,
		})
	}

	return docEntities
}
