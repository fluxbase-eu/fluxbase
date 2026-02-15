-- Chatbot Collection Links Migration
-- Adds collections to chatbot knowledge sources
-- Migration 088
-- ========================================================================
-- ADD COLLECTION_IDS COLUMN
-- ========================================================================
-- Add collection_ids as nullable initially (array of collection references)
ALTER TABLE ai.chatbots
    ADD COLUMN IF NOT EXISTS collection_ids UUID[] DEFAULT '{}';

CREATE INDEX IF NOT EXISTS idx_chatbot_collections ON ai.chatbots USING GIN (collection_ids);

COMMENT ON COLUMN ai.chatbots.collection_ids IS 'Knowledge base collections linked to chatbot (supplements individual KB selection)';
