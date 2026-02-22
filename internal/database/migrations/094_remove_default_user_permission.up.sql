-- Remove default_user_permission column (simplifies visibility model)
-- After this change:
--   - private: Owner only
--   - shared: Users with explicit permissions (viewer/editor/owner per user)
--   - public: All authenticated users get viewer access (read-only)

ALTER TABLE ai.knowledge_bases DROP COLUMN IF EXISTS default_user_permission;
