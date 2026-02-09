-- Emergency token revocation table for service_role tokens
-- This provides a mechanism to revoke compromised service_role tokens immediately
-- without waiting for token expiry

CREATE TABLE IF NOT EXISTS auth.emergency_revocation (
    id BIGSERIAL PRIMARY KEY,
    revoked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_by TEXT NOT NULL, -- Admin user ID who initiated revocation
    reason TEXT, -- Reason for emergency revocation (e.g., "key leak", "security incident")
    revokes_all BOOLEAN NOT NULL DEFAULT FALSE, -- If true, revokes ALL service_role tokens
    revoked_jti TEXT UNIQUE, -- Specific JWT ID to revoke (if revokes_all is false)
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '7 days') -- Auto-cleanup after 7 days
);

-- Index for quick lookups of specific token revocations
CREATE INDEX IF NOT EXISTS idx_emergency_revocation_jti ON auth.emergency_revocation(revoked_jti) WHERE revoked_jti IS NOT NULL;

-- Index for checking if emergency revocation is active
CREATE INDEX IF NOT EXISTS idx_emergency_revocation_active ON auth.emergency_revocation(expires_at) WHERE expires_at > NOW();

-- Index for quick lookups of global revocation status
CREATE INDEX IF NOT EXISTS idx_emergency_revocation_all ON auth.emergency_revocation(revokes_all) WHERE revokes_all = TRUE AND expires_at > NOW();

-- Grant access to authenticated role for checking revocations (read-only)
GRANT SELECT ON auth.emergency_revocation TO authenticated;

-- Grant access to service_role for checking revocations (read-only)
GRANT SELECT ON auth.emergency_revocation TO service_role;

-- Grant full access to admin role for managing emergency revocations
GRANT ALL ON auth.emergency_revocation TO admin;

-- Grant usage on sequence
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA auth TO admin;

-- Comment for documentation
COMMENT ON TABLE auth.emergency_revocation IS 'Emergency revocation table for service_role tokens. Allows immediate revocation of compromised service keys without waiting for expiry.';
COMMENT ON COLUMN auth.emergency_revocation.revokes_all IS 'When true, all service_role tokens are considered revoked. Used for security incidents requiring immediate global revocation.';
COMMENT ON COLUMN auth.emergency_revocation.revoked_jti IS 'Specific JWT ID to revoke. Only used when revokes_all is false.';
COMMENT ON COLUMN auth.emergency_revocation.expires_at IS 'Records auto-expire after 7 days for cleanup. Active revocations have expires_at > NOW().';
