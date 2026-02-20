-- Document-Level Permissions
-- Adds ability to share individual documents with specific users
-- Users can only see their own documents + documents explicitly shared with them

-- ============================================================================
-- ADD OWNER TO DOCUMENTS
-- ============================================================================

-- Add owner_id column to documents table
ALTER TABLE ai.documents ADD COLUMN IF NOT EXISTS owner_id UUID REFERENCES auth.users(id) ON DELETE SET NULL;

-- Create index for owner lookups
CREATE INDEX IF NOT EXISTS idx_ai_documents_owner ON ai.documents(owner_id);

-- For existing documents, set owner_id to created_by if null
UPDATE ai.documents
SET owner_id = created_by
WHERE owner_id IS NULL AND created_by IS NOT NULL;

-- Make owner_id NOT NULL with a default
ALTER TABLE ai.documents ALTER COLUMN owner_id SET NOT NULL;
ALTER TABLE ai.documents ALTER COLUMN owner_id SET DEFAULT auth.uid();

COMMENT ON COLUMN ai.documents.owner_id IS 'User who owns this document (can see and share it)';

-- ============================================================================
-- DOCUMENT PERMISSIONS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS ai.document_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES ai.documents(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    permission TEXT NOT NULL CHECK (permission IN ('viewer', 'editor')),
    granted_by UUID NOT NULL REFERENCES auth.users(id) ON DELETE SET NULL,
    granted_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT unique_document_user_permission UNIQUE (document_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_ai_document_permissions_document ON ai.document_permissions(document_id);
CREATE INDEX IF NOT EXISTS idx_ai_document_permissions_user ON ai.document_permissions(user_id);

COMMENT ON TABLE ai.document_permissions IS 'Permissions for sharing individual documents with specific users';
COMMENT ON COLUMN ai.document_permissions.permission IS 'viewer: can view, editor: can view and edit';

-- ============================================================================
-- ROW LEVEL SECURITY FOR DOCUMENT PERMISSIONS
-- ============================================================================

ALTER TABLE ai.document_permissions ENABLE ROW LEVEL SECURITY;

-- Service role can do everything
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE schemaname = 'ai' AND tablename = 'document_permissions' AND policyname = 'ai_doc_perms_service_all') THEN
        CREATE POLICY "ai_doc_perms_service_all" ON ai.document_permissions FOR ALL TO service_role USING (true);
    END IF;
END $$;

-- Document owners can manage permissions on their documents
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE schemaname = 'ai' AND tablename = 'document_permissions' AND policyname = 'ai_doc_perms_owner_manage') THEN
        CREATE POLICY "ai_doc_perms_owner_manage" ON ai.document_permissions FOR ALL TO authenticated
        USING (
            EXISTS (
                SELECT 1 FROM ai.documents d
                WHERE d.id = document_permissions.document_id
                AND d.owner_id = auth.uid()
            )
        );
    END IF;
END $$;

-- Users can read permissions granted to them
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE schemaname = 'ai' AND tablename = 'document_permissions' AND policyname = 'ai_doc_perms_user_read') THEN
        CREATE POLICY "ai_doc_perms_user_read" ON ai.document_permissions FOR SELECT TO authenticated
        USING (user_id = auth.uid());
    END IF;
END $$;

-- Dashboard admins can manage all permissions
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE schemaname = 'ai' AND tablename = 'document_permissions' AND policyname = 'ai_doc_perms_dashboard_admin') THEN
        CREATE POLICY "ai_doc_perms_dashboard_admin" ON ai.document_permissions FOR ALL TO authenticated
        USING (auth.role() = 'dashboard_admin');
    END IF;
END $$;

-- ============================================================================
-- UPDATE DOCUMENT RLS POLICIES FOR DOCUMENT-LEVEL ACCESS
-- ============================================================================

-- Drop existing document read policies
DROP POLICY IF EXISTS "ai_documents_service_all" ON ai.documents;
DROP POLICY IF EXISTS "ai_documents_dashboard_admin" ON ai.documents;

-- Service role can do everything
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE schemaname = 'ai' AND tablename = 'documents' AND policyname = 'ai_documents_service_all') THEN
        CREATE POLICY "ai_documents_service_all" ON ai.documents FOR ALL TO service_role USING (true);
    END IF;
END $$;

-- Dashboard admins can do everything
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE schemaname = 'ai' AND tablename = 'documents' AND policyname = 'ai_documents_dashboard_admin') THEN
        CREATE POLICY "ai_documents_dashboard_admin" ON ai.documents FOR ALL TO authenticated
        USING (auth.role() = 'dashboard_admin');
    END IF;
END $$;

-- Users can read their own documents
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE schemaname = 'ai' AND tablename = 'documents' AND policyname = 'ai_documents_read_own') THEN
        CREATE POLICY "ai_documents_read_own" ON ai.documents FOR SELECT TO authenticated
        USING (owner_id = auth.uid());
    END IF;
END $$;

-- Users can read documents shared with them
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE schemaname = 'ai' AND tablename = 'documents' AND policyname = 'ai_documents_read_shared') THEN
        CREATE POLICY "ai_documents_read_shared" ON ai.documents FOR SELECT TO authenticated
        USING (
            EXISTS (
                SELECT 1 FROM ai.document_permissions dp
                WHERE dp.document_id = documents.id
                AND dp.user_id = auth.uid()
            )
        );
    END IF;
END $$;

-- Users can create documents (will be owned by them via trigger)
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE schemaname = 'ai' AND tablename = 'documents' AND policyname = 'ai_documents_insert') THEN
        CREATE POLICY "ai_documents_insert" ON ai.documents FOR INSERT TO authenticated
        WITH CHECK (true);
    END IF;
END $$;

-- Users can update their own documents
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE schemaname = 'ai' AND tablename = 'documents' AND policyname = 'ai_documents_update_own') THEN
        CREATE POLICY "ai_documents_update_own" ON ai.documents FOR UPDATE TO authenticated
        USING (owner_id = auth.uid());
    END IF;
END $$;

-- Users can update documents shared with them as editor
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE schemaname = 'ai' AND tablename = 'documents' AND policyname = 'ai_documents_update_shared') THEN
        CREATE POLICY "ai_documents_update_shared" ON ai.documents FOR UPDATE TO authenticated
        USING (
            EXISTS (
                SELECT 1 FROM ai.document_permissions dp
                WHERE dp.document_id = documents.id
                AND dp.user_id = auth.uid()
                AND dp.permission = 'editor'
            )
        );
    END IF;
END $$;

-- Users can delete their own documents
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE schemaname = 'ai' AND tablename = 'documents' AND policyname = 'ai_documents_delete_own') THEN
        CREATE POLICY "ai_documents_delete_own" ON ai.documents FOR DELETE TO authenticated
        USING (owner_id = auth.uid());
    END IF;
END $$;

-- ============================================================================
-- UPDATE CHUNKS RLS POLICIES TO RESPECT DOCUMENT PERMISSIONS
-- ============================================================================

-- Drop existing chunk policies
DROP POLICY IF EXISTS "ai_chunks_service_all" ON ai.chunks;
DROP POLICY IF EXISTS "ai_chunks_dashboard_admin" ON ai.chunks;

-- Service role can do everything
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE schemaname = 'ai' AND tablename = 'chunks' AND policyname = 'ai_chunks_service_all') THEN
        CREATE POLICY "ai_chunks_service_all" ON ai.chunks FOR ALL TO service_role USING (true);
    END IF;
END $$;

-- Dashboard admins can do everything
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE schemaname = 'ai' AND tablename = 'chunks' AND policyname = 'ai_chunks_dashboard_admin') THEN
        CREATE POLICY "ai_chunks_dashboard_admin" ON ai.chunks FOR ALL TO authenticated
        USING (auth.role() = 'dashboard_admin');
    END IF;
END $$;

-- Users can read chunks from their own documents
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE schemaname = 'ai' AND tablename = 'chunks' AND policyname = 'ai_chunks_read_own_docs') THEN
        CREATE POLICY "ai_chunks_read_own_docs" ON ai.chunks FOR SELECT TO authenticated
        USING (
            EXISTS (
                SELECT 1 FROM ai.documents d
                WHERE d.id = chunks.document_id
                AND d.owner_id = auth.uid()
            )
        );
    END IF;
END $$;

-- Users can read chunks from documents shared with them
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE schemaname = 'ai' AND tablename = 'chunks' AND policyname = 'ai_chunks_read_shared_docs') THEN
        CREATE POLICY "ai_chunks_read_shared_docs" ON ai.chunks FOR SELECT TO authenticated
        USING (
            EXISTS (
                SELECT 1 FROM ai.documents d
                JOIN ai.document_permissions dp ON dp.document_id = d.id
                WHERE d.id = chunks.document_id
                AND dp.user_id = auth.uid()
            )
        );
    END IF;
END $$;

-- ============================================================================
-- TRIGGER TO SET DOCUMENT OWNER ON INSERT
-- ============================================================================

CREATE OR REPLACE FUNCTION ai.set_document_owner()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.owner_id IS NULL THEN
        NEW.owner_id = auth.uid();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

DROP TRIGGER IF EXISTS documents_set_owner ON ai.documents;
CREATE TRIGGER documents_set_owner
BEFORE INSERT ON ai.documents
FOR EACH ROW EXECUTE FUNCTION ai.set_document_owner();

-- ============================================================================
-- GRANTS
-- ============================================================================

GRANT SELECT ON ai.document_permissions TO authenticated;
GRANT ALL ON ai.document_permissions TO service_role;

-- ============================================================================
-- HELPER FUNCTION TO CHECK DOCUMENT ACCESS
-- ============================================================================

CREATE OR REPLACE FUNCTION ai.can_access_document(p_document_id UUID, p_user_id UUID)
RETURNS BOOLEAN AS $$
BEGIN
    -- User owns the document
    IF EXISTS (
        SELECT 1 FROM ai.documents
        WHERE id = p_document_id
        AND owner_id = p_user_id
    ) THEN
        RETURN true;
    END IF;

    -- Document is shared with user
    IF EXISTS (
        SELECT 1 FROM ai.document_permissions
        WHERE document_id = p_document_id
        AND user_id = p_user_id
    ) THEN
        RETURN true;
    END IF;

    RETURN false;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMENT ON FUNCTION ai.can_access_document IS 'Check if a user can access a document (owns it or has been granted permission)';
