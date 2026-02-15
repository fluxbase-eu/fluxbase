-- Rollback: Remove collection_ids from chatbots
-- Migration 090
--
ALTER TABLE ai.chatbots
    DROP COLUMN IF EXISTS collection_ids;

DROP INDEX IF EXISTS idx_chatbot_collections;
