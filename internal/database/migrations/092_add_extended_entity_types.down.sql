-- Rollback Migration 092: Revert to Original Entity and Relationship Types

-- Restore original entity_type check constraint
ALTER TABLE ai.entities DROP CONSTRAINT IF EXISTS entities_entity_type_check;

ALTER TABLE ai.entities
ADD CONSTRAINT entities_entity_type_check
CHECK (entity_type IN (
    'person', 'organization', 'location', 'concept', 'product', 'event', 'other'
));

-- Restore original relationship_type check constraint
ALTER TABLE ai.entity_relationships DROP CONSTRAINT IF EXISTS entity_relationships_relationship_type_check;

ALTER TABLE ai.entity_relationships
ADD CONSTRAINT entity_relationships_relationship_type_check
CHECK (relationship_type IN (
    'works_at', 'located_in', 'founded_by', 'owns', 'part_of',
    'related_to', 'knows', 'customer_of', 'supplier_of',
    'invested_in', 'acquired', 'merged_with', 'competitor_of',
    'parent_of', 'child_of', 'spouse_of', 'sibling_of', 'other'
));
