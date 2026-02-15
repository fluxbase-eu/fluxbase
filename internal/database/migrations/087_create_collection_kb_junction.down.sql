-- Rollback: Drop collection-knowledge base junction table
-- Migration 089
--
DROP TABLE IF EXISTS ai.collection_knowledge_bases CASCADE;
DROP INDEX IF EXISTS idx_ai_collection_kbs_kb;
DROP INDEX IF EXISTS idx_ai_collection_kbs_collection;
