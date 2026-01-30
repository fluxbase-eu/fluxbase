-- Revert MCP OAuth 2.1 Support

-- Drop trigger first
DROP TRIGGER IF EXISTS trigger_mcp_oauth_clients_updated_at ON auth.mcp_oauth_clients;
DROP FUNCTION IF EXISTS auth.update_mcp_oauth_clients_updated_at();

-- Drop tables (order matters due to foreign keys)
DROP TABLE IF EXISTS auth.mcp_oauth_tokens;
DROP TABLE IF EXISTS auth.mcp_oauth_codes;
DROP TABLE IF EXISTS auth.mcp_oauth_clients;
