-- User-Scoped RLS Policies for Knowledge Bases Down

-- ============================================================================
-- DROP NEW POLICIES
-- ============================================================================

DROP POLICY IF EXISTS "ai_kb_manage_own" ON ai.knowledge_bases;
DROP POLICY IF EXISTS "ai_kb_read_public" ON ai.knowledge_bases;
DROP POLICY IF EXISTS "ai_kb_read_shared" ON ai.knowledge_bases;
DROP POLICY IF EXISTS "ai_kb_service_all" ON ai.knowledge_bases;
DROP POLICY IF EXISTS "ai_kb_admin_all" ON ai.knowledge_bases;

DROP POLICY IF EXISTS "ai_documents_manage_via_kb" ON ai.documents;
DROP POLICY IF EXISTS "ai_documents_read_public" ON ai.documents;

-- ============================================================================
-- RECREATE OLD POLICIES
-- ============================================================================

CREATE POLICY "ai_kb_read" ON ai.knowledge_bases
    FOR SELECT
    TO authenticated
    USING (true);

CREATE POLICY "ai_kb_dashboard_admin" ON ai.knowledge_bases
    FOR ALL
    TO authenticated
    USING (auth.current_user_role() = 'dashboard_admin')
    WITH CHECK (auth.current_user_role() = 'dashboard_admin');
