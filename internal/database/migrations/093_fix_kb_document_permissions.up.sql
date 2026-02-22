-- Fix Knowledge Base Document Permissions
-- Ensures KB permission levels are properly enforced for document operations
-- - Viewer: Can only read documents
-- - Editor/Owner: Can create, update, delete documents

-- ============================================================================
-- DROP FLAWED POLICIES
-- ============================================================================

-- Drop the buggy policy that doesn't check permission level
DROP POLICY IF EXISTS "ai_documents_manage_via_kb" ON ai.documents;

-- ============================================================================
-- NEW DOCUMENT POLICIES: KB PERMISSION ENFORCEMENT
-- ============================================================================

-- Users can READ documents if:
-- 1. They are the KB owner, OR
-- 2. They have ANY KB permission (viewer, editor, owner), OR
-- 3. The KB is public
CREATE POLICY "ai_documents_read_via_kb" ON ai.documents
    FOR SELECT
    TO authenticated
    USING (
        -- KB owner can always read
        EXISTS (
            SELECT 1 FROM ai.knowledge_bases kb
            WHERE kb.id = ai.documents.knowledge_base_id
            AND kb.owner_id = auth.current_user_id()
        )
        -- User has any KB permission (viewer, editor, or owner)
        OR EXISTS (
            SELECT 1 FROM ai.knowledge_bases kb
            JOIN ai.knowledge_base_permissions kbp ON kb.id = kbp.knowledge_base_id
            WHERE kb.id = ai.documents.knowledge_base_id
            AND kbp.user_id = auth.current_user_id()
        )
        -- KB is public
        OR EXISTS (
            SELECT 1 FROM ai.knowledge_bases kb
            WHERE kb.id = ai.documents.knowledge_base_id
            AND kb.visibility = 'public'
        )
    );

-- Users can INSERT documents if:
-- 1. They are the KB owner, OR
-- 2. They have editor or owner KB permission
CREATE POLICY "ai_documents_insert_via_kb" ON ai.documents
    FOR INSERT
    TO authenticated
    WITH CHECK (
        -- KB owner can always insert
        EXISTS (
            SELECT 1 FROM ai.knowledge_bases kb
            WHERE kb.id = ai.documents.knowledge_base_id
            AND kb.owner_id = auth.current_user_id()
        )
        -- User has editor or owner KB permission
        OR EXISTS (
            SELECT 1 FROM ai.knowledge_bases kb
            JOIN ai.knowledge_base_permissions kbp ON kb.id = kbp.knowledge_base_id
            WHERE kb.id = ai.documents.knowledge_base_id
            AND kbp.user_id = auth.current_user_id()
            AND kbp.permission IN ('editor', 'owner')
        )
    );

-- Users can UPDATE documents if:
-- 1. They are the KB owner, OR
-- 2. They have editor or owner KB permission
CREATE POLICY "ai_documents_update_via_kb" ON ai.documents
    FOR UPDATE
    TO authenticated
    USING (
        -- KB owner can always update
        EXISTS (
            SELECT 1 FROM ai.knowledge_bases kb
            WHERE kb.id = ai.documents.knowledge_base_id
            AND kb.owner_id = auth.current_user_id()
        )
        -- User has editor or owner KB permission
        OR EXISTS (
            SELECT 1 FROM ai.knowledge_bases kb
            JOIN ai.knowledge_base_permissions kbp ON kb.id = kbp.knowledge_base_id
            WHERE kb.id = ai.documents.knowledge_base_id
            AND kbp.user_id = auth.current_user_id()
            AND kbp.permission IN ('editor', 'owner')
        )
    )
    WITH CHECK (
        -- KB owner can always update
        EXISTS (
            SELECT 1 FROM ai.knowledge_bases kb
            WHERE kb.id = ai.documents.knowledge_base_id
            AND kb.owner_id = auth.current_user_id()
        )
        -- User has editor or owner KB permission
        OR EXISTS (
            SELECT 1 FROM ai.knowledge_bases kb
            JOIN ai.knowledge_base_permissions kbp ON kb.id = kbp.knowledge_base_id
            WHERE kb.id = ai.documents.knowledge_base_id
            AND kbp.user_id = auth.current_user_id()
            AND kbp.permission IN ('editor', 'owner')
        )
    );

-- Users can DELETE documents if:
-- 1. They are the KB owner, OR
-- 2. They have editor or owner KB permission
CREATE POLICY "ai_documents_delete_via_kb" ON ai.documents
    FOR DELETE
    TO authenticated
    USING (
        -- KB owner can always delete
        EXISTS (
            SELECT 1 FROM ai.knowledge_bases kb
            WHERE kb.id = ai.documents.knowledge_base_id
            AND kb.owner_id = auth.current_user_id()
        )
        -- User has editor or owner KB permission
        OR EXISTS (
            SELECT 1 FROM ai.knowledge_bases kb
            JOIN ai.knowledge_base_permissions kbp ON kb.id = kbp.knowledge_base_id
            WHERE kb.id = ai.documents.knowledge_base_id
            AND kbp.user_id = auth.current_user_id()
            AND kbp.permission IN ('editor', 'owner')
        )
    );

-- ============================================================================
-- DROP THE OVERLY PERMISSIVE INSERT POLICY FROM MIGRATION 090
-- ============================================================================

-- Migration 090 added a policy that allows all authenticated users to insert
-- This is too permissive - document creation should require KB write access
DROP POLICY IF EXISTS "ai_documents_insert" ON ai.documents;
