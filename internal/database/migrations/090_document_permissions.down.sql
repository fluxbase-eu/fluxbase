-- Rollback Document-Level Permissions

-- Drop helper function
DROP FUNCTION IF EXISTS ai.can_access_document;

-- Drop trigger
DROP TRIGGER IF EXISTS documents_set_owner ON ai.documents;
DROP FUNCTION IF EXISTS ai.set_document_owner;

-- Drop document permissions table
DROP TABLE IF EXISTS ai.document_permissions;

-- Remove owner_id column from documents
ALTER TABLE ai.documents DROP COLUMN IF EXISTS owner_id;

-- Drop indexes
DROP INDEX IF EXISTS idx_ai_documents_owner;
DROP INDEX IF EXISTS idx_ai_document_permissions_document;
DROP INDEX IF EXISTS idx_ai_document_permissions_user;

-- Restore original document RLS policies
DROP POLICY IF EXISTS "ai_documents_service_all" ON ai.documents;
DROP POLICY IF EXISTS "ai_documents_dashboard_admin" ON ai.documents;
DROP POLICY IF EXISTS "ai_documents_read_own" ON ai.documents;
DROP POLICY IF EXISTS "ai_documents_read_shared" ON ai.documents;
DROP POLICY IF EXISTS "ai_documents_insert" ON ai.documents;
DROP POLICY IF EXISTS "ai_documents_update_own" ON ai.documents;
DROP POLICY IF EXISTS "ai_documents_update_shared" ON ai.documents;
DROP POLICY IF EXISTS "ai_documents_delete_own" ON ai.documents;

-- Restore original policies
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE schemaname = 'ai' AND tablename = 'documents' AND policyname = 'ai_documents_service_all') THEN
        CREATE POLICY "ai_documents_service_all" ON ai.documents FOR ALL TO service_role USING (true);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE schemaname = 'ai' AND tablename = 'documents' AND policyname = 'ai_documents_dashboard_admin') THEN
        CREATE POLICY "ai_documents_dashboard_admin" ON ai.documents FOR ALL TO authenticated USING (auth.role() = 'dashboard_admin');
    END IF;
END $$;

-- Restore original chunk RLS policies
DROP POLICY IF EXISTS "ai_chunks_service_all" ON ai.chunks;
DROP POLICY IF EXISTS "ai_chunks_dashboard_admin" ON ai.chunks;
DROP POLICY IF EXISTS "ai_chunks_read_own_docs" ON ai.chunks;
DROP POLICY IF EXISTS "ai_chunks_read_shared_docs" ON ai.chunks;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE schemaname = 'ai' AND tablename = 'chunks' AND policyname = 'ai_chunks_service_all') THEN
        CREATE POLICY "ai_chunks_service_all" ON ai.chunks FOR ALL TO service_role USING (true);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE schemaname = 'ai' AND tablename = 'chunks' AND policyname = 'ai_chunks_dashboard_admin') THEN
        CREATE POLICY "ai_chunks_dashboard_admin" ON ai.chunks FOR ALL TO authenticated USING (auth.role() = 'dashboard_admin');
    END IF;
END $$;
