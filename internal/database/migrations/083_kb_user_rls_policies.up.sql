-- User-Scoped RLS Policies for Knowledge Bases
-- Replaces admin-only policies with user-scoped access control

-- ============================================================================
-- DROP OLD POLICIES (make migration idempotent)
-- ============================================================================

-- Drop knowledge base policies
DROP POLICY IF EXISTS "ai_kb_read" ON ai.knowledge_bases;
DROP POLICY IF EXISTS "ai_kb_dashboard_admin" ON ai.knowledge_bases;
DROP POLICY IF EXISTS "ai_kb_manage_own" ON ai.knowledge_bases;
DROP POLICY IF EXISTS "ai_kb_read_public" ON ai.knowledge_bases;
DROP POLICY IF EXISTS "ai_kb_read_shared" ON ai.knowledge_bases;
DROP POLICY IF EXISTS "ai_kb_service_all" ON ai.knowledge_bases;
DROP POLICY IF EXISTS "ai_kb_admin_all" ON ai.knowledge_bases;

-- Drop document policies
DROP POLICY IF EXISTS "ai_documents_manage_via_kb" ON ai.documents;
DROP POLICY IF EXISTS "ai_documents_read_public" ON ai.documents;

-- ============================================================================
-- NEW POLICIES: USER-SCOPED ACCESS CONTROL
-- ============================================================================

-- Users can manage their own KBs (private or shared with editor/owner permission)
CREATE POLICY "ai_kb_manage_own" ON ai.knowledge_bases
    FOR ALL
    TO authenticated
    USING (
        owner_id = auth.current_user_id()
        OR EXISTS (
            SELECT 1 FROM ai.knowledge_base_permissions
            WHERE knowledge_base_id = ai.knowledge_bases.id
            AND user_id = auth.current_user_id()
            AND permission IN ('editor', 'owner')
        )
    )
    WITH CHECK (
        owner_id = auth.current_user_id()
        OR EXISTS (
            SELECT 1 FROM ai.knowledge_base_permissions
            WHERE knowledge_base_id = ai.knowledge_bases.id
            AND user_id = auth.current_user_id()
            AND permission IN ('editor', 'owner')
        )
    );

-- Users can read public KBs
CREATE POLICY "ai_kb_read_public" ON ai.knowledge_bases
    FOR SELECT
    TO authenticated
    USING (visibility = 'public');

-- Users can read shared KBs if granted permission
CREATE POLICY "ai_kb_read_shared" ON ai.knowledge_bases
    FOR SELECT
    TO authenticated
    USING (
        visibility = 'shared'
        AND EXISTS (
            SELECT 1 FROM ai.knowledge_base_permissions
            WHERE knowledge_base_id = ai.knowledge_bases.id
            AND user_id = auth.current_user_id()
        )
    );

-- Service role bypasses all
CREATE POLICY "ai_kb_service_all" ON ai.knowledge_bases
    FOR ALL
    TO authenticated
    USING (auth.current_user_role() = 'service_role')
    WITH CHECK (auth.current_user_role() = 'service_role');

-- Dashboard admin bypasses all
CREATE POLICY "ai_kb_admin_all" ON ai.knowledge_bases
    FOR ALL
    TO authenticated
    USING (auth.current_user_role() = 'dashboard_admin')
    WITH CHECK (auth.current_user_role() = 'dashboard_admin');

-- ============================================================================
-- DOCUMENTS: INHERIT VISIBILITY FROM PARENT KB
-- ============================================================================

-- Documents inherit visibility from parent KB
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

CREATE POLICY "ai_documents_read_public" ON ai.documents
    FOR SELECT
    TO authenticated
    USING (
        EXISTS (
            SELECT 1 FROM ai.knowledge_bases kb
            WHERE kb.id = ai.documents.knowledge_base_id
            AND kb.visibility = 'public'
        )
    );
