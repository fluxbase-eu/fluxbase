-- MCP OAuth 2.1 Support
-- Enables Dynamic Client Registration (RFC 7591) and OAuth authorization for MCP clients

-- MCP OAuth clients (Dynamic Client Registration)
CREATE TABLE auth.mcp_oauth_clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id TEXT UNIQUE NOT NULL,
    client_name TEXT NOT NULL,
    redirect_uris TEXT[] NOT NULL,
    grant_types TEXT[] DEFAULT ARRAY['authorization_code', 'refresh_token'],
    response_types TEXT[] DEFAULT ARRAY['code'],
    scope TEXT,
    client_uri TEXT,
    logo_uri TEXT,
    contacts TEXT[],
    tos_uri TEXT,
    policy_uri TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for client_id lookups
CREATE INDEX idx_mcp_oauth_clients_client_id ON auth.mcp_oauth_clients(client_id);

-- MCP OAuth authorization codes (temporary, exchanged for tokens)
CREATE TABLE auth.mcp_oauth_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT UNIQUE NOT NULL,
    client_id TEXT NOT NULL REFERENCES auth.mcp_oauth_clients(client_id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    redirect_uri TEXT NOT NULL,
    scope TEXT NOT NULL,
    code_challenge TEXT,
    code_challenge_method TEXT DEFAULT 'S256',
    state TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    used_at TIMESTAMPTZ
);

-- Index for code lookups
CREATE INDEX idx_mcp_oauth_codes_code ON auth.mcp_oauth_codes(code);

-- Cleanup expired codes
CREATE INDEX idx_mcp_oauth_codes_expires_at ON auth.mcp_oauth_codes(expires_at);

-- MCP OAuth tokens (access and refresh tokens)
CREATE TABLE auth.mcp_oauth_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id TEXT NOT NULL REFERENCES auth.mcp_oauth_clients(client_id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    access_token_hash TEXT NOT NULL,
    refresh_token_hash TEXT,
    scope TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    refresh_expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

-- Index for token lookups by hash
CREATE INDEX idx_mcp_oauth_tokens_access_hash ON auth.mcp_oauth_tokens(access_token_hash);
CREATE INDEX idx_mcp_oauth_tokens_refresh_hash ON auth.mcp_oauth_tokens(refresh_token_hash);

-- Index for user's tokens (for revocation UI)
CREATE INDEX idx_mcp_oauth_tokens_user_id ON auth.mcp_oauth_tokens(user_id);

-- Index for cleanup of expired tokens
CREATE INDEX idx_mcp_oauth_tokens_expires_at ON auth.mcp_oauth_tokens(expires_at);

-- Trigger to update updated_at on clients
CREATE OR REPLACE FUNCTION auth.update_mcp_oauth_clients_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_mcp_oauth_clients_updated_at
    BEFORE UPDATE ON auth.mcp_oauth_clients
    FOR EACH ROW
    EXECUTE FUNCTION auth.update_mcp_oauth_clients_updated_at();

-- Grant permissions to authenticated role
GRANT SELECT, INSERT, UPDATE, DELETE ON auth.mcp_oauth_clients TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON auth.mcp_oauth_codes TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON auth.mcp_oauth_tokens TO authenticated;

-- Grant permissions to service role
GRANT ALL ON auth.mcp_oauth_clients TO service_role;
GRANT ALL ON auth.mcp_oauth_codes TO service_role;
GRANT ALL ON auth.mcp_oauth_tokens TO service_role;
