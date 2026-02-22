-- Rollback KB Document Permission Fixes

-- Drop the new policies
DROP POLICY IF EXISTS "ai_documents_read_via_kb" ON ai.documents;
DROP POLICY IF EXISTS "ai_documents_insert_via_kb" ON ai.documents;
DROP POLICY IF EXISTS "ai_documents_update_via_kb" ON ai.documents;
DROP POLICY IF EXISTS "ai_documents_delete_via_kb" ON ai.documents;

-- Restore the buggy policy (matches migration 083)
CREATE POLICY "ai_documents_manage_via_kb" ON ai.documents
    FOR ALL
    TO authenticated
    USING (
        EXISTS (
            SELECT 1 FROM ai.knowledge_bases kb
            JOIN ai.knowledge_base_permissions kbp ON kb.id = kbp.knowledge_base_id
            WHERE kb.id = ai.documents.knowledge_base_id
            AND (kb.owner_id = auth.current_user_id() OR kbp.user_id = auth.current_user_id())
        )
    )
    WITH CHECK (
        EXISTS (
            SELECT 1 FROM ai.knowledge_bases kb
            JOIN ai.knowledge_base_permissions kbp ON kb.id = kbp.knowledge_base_id
            WHERE kb.id = ai.documents.knowledge_base_id
            AND (kb.owner_id = auth.current_user_id() OR kbp.user_id = auth.current_user_id())
        )
    );

-- Restore the overly permissive insert policy from migration 090
CREATE POLICY "ai_documents_insert" ON ai.documents FOR INSERT TO authenticated
WITH CHECK (true);
