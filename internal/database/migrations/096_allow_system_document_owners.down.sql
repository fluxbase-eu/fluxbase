-- Revert: make owner_id NOT NULL again
ALTER TABLE ai.documents ALTER COLUMN owner_id SET NOT NULL;

-- Revert the trigger to original behavior
CREATE OR REPLACE FUNCTION ai.set_document_owner()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.owner_id IS NULL THEN
        NEW.owner_id = auth.uid();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
