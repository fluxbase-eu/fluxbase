-- ============================================================================
-- STORAGE BUCKET EXISTS FUNCTION
-- ============================================================================
-- This creates a SECURITY DEFINER function to check if a bucket exists.
-- The function bypasses RLS policies to allow the storage handler to verify
-- bucket existence regardless of the current user's role.
--
-- This is needed because:
-- 1. storage.buckets has FORCE ROW LEVEL SECURITY enabled
-- 2. The storage handler needs to verify a bucket exists before upload
-- 3. The handler should be able to see all buckets, not just public ones
-- ============================================================================

-- SECURITY DEFINER function to check if bucket exists
-- This bypasses RLS to allow bucket existence checks by the storage handler
CREATE OR REPLACE FUNCTION storage.bucket_exists(bucket_name TEXT)
RETURNS BOOLEAN
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path = public, storage
AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM storage.buckets
        WHERE name = bucket_name
    );
END;
$$;

COMMENT ON FUNCTION storage.bucket_exists(TEXT) IS 'SECURITY DEFINER function to check if a bucket exists, bypassing RLS. Used by storage handler to validate bucket existence before upload.';

-- SECURITY DEFINER function to get bucket settings
-- This bypasses RLS to allow the storage handler to fetch bucket settings
CREATE OR REPLACE FUNCTION storage.get_bucket_settings(bucket_name TEXT)
RETURNS TABLE (
    max_file_size BIGINT,
    allowed_mime_types TEXT[]
)
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path = public, storage
AS $$
BEGIN
    RETURN QUERY
    SELECT b.max_file_size, b.allowed_mime_types
    FROM storage.buckets b
    WHERE b.name = bucket_name;
END;
$$;

COMMENT ON FUNCTION storage.get_bucket_settings(TEXT) IS 'SECURITY DEFINER function to get bucket settings, bypassing RLS. Used by storage handler to validate upload constraints.';

-- Grant execute permissions on these functions to all roles
GRANT EXECUTE ON FUNCTION storage.bucket_exists(TEXT) TO anon, authenticated, service_role;
GRANT EXECUTE ON FUNCTION storage.get_bucket_settings(TEXT) TO anon, authenticated, service_role;
