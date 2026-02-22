-- Allow NULL owner_id for system-generated documents (e.g., table exports via service role)
-- This allows service role operations to create documents without requiring a valid user owner

-- Drop the NOT NULL constraint on owner_id
ALTER TABLE ai.documents ALTER COLUMN owner_id DROP NOT NULL;

-- Update the trigger to not set owner_id for service role
CREATE OR REPLACE FUNCTION ai.set_document_owner()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.owner_id IS NULL AND auth.uid() IS NOT NULL THEN
        NEW.owner_id = auth.uid();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMENT ON COLUMN ai.documents.owner_id IS 'User who owns this document (can see and share it). NULL for system-generated documents created via service role.';
