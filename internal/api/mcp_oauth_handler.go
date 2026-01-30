package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/fluxbase-eu/fluxbase/internal/config"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// MCPOAuthHandler handles OAuth 2.1 endpoints for MCP authentication
type MCPOAuthHandler struct {
	db        *pgxpool.Pool
	config    *config.MCPConfig
	baseURL   string
	publicURL string
}

// NewMCPOAuthHandler creates a new MCP OAuth handler
func NewMCPOAuthHandler(db *pgxpool.Pool, mcpConfig *config.MCPConfig, baseURL, publicURL string) *MCPOAuthHandler {
	return &MCPOAuthHandler{
		db:        db,
		config:    mcpConfig,
		baseURL:   baseURL,
		publicURL: publicURL,
	}
}

// RegisterRoutes registers the MCP OAuth routes
func (h *MCPOAuthHandler) RegisterRoutes(app fiber.Router, mcpGroup fiber.Router) {
	// Discovery endpoint (public, no auth required)
	app.Get("/.well-known/oauth-authorization-server", h.handleDiscovery)

	// OAuth endpoints under MCP group
	mcpGroup.Post("/oauth/register", h.handleRegister)
	mcpGroup.Get("/oauth/authorize", h.handleAuthorize)
	mcpGroup.Post("/oauth/authorize", h.handleAuthorizeConsent)
	mcpGroup.Post("/oauth/token", h.handleToken)
	mcpGroup.Post("/oauth/revoke", h.handleRevoke)
}

// OAuth 2.0 Authorization Server Metadata (RFC 8414)
type OAuthServerMetadata struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	RegistrationEndpoint              string   `json:"registration_endpoint,omitempty"`
	RevocationEndpoint                string   `json:"revocation_endpoint,omitempty"`
	ScopesSupported                   []string `json:"scopes_supported"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
}

// handleDiscovery handles GET /.well-known/oauth-authorization-server
func (h *MCPOAuthHandler) handleDiscovery(c *fiber.Ctx) error {
	if !h.config.OAuth.Enabled {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "oauth_disabled",
			"error_description": "OAuth is not enabled for this MCP server",
		})
	}

	baseURL := h.publicURL
	if baseURL == "" {
		baseURL = h.baseURL
	}

	metadata := OAuthServerMetadata{
		Issuer:                baseURL,
		AuthorizationEndpoint: fmt.Sprintf("%s%s/oauth/authorize", baseURL, h.config.BasePath),
		TokenEndpoint:         fmt.Sprintf("%s%s/oauth/token", baseURL, h.config.BasePath),
		RevocationEndpoint:    fmt.Sprintf("%s%s/oauth/revoke", baseURL, h.config.BasePath),
		ScopesSupported: []string{
			"read:tables", "write:tables",
			"execute:functions", "execute:rpc",
			"read:storage", "write:storage",
			"execute:jobs",
			"read:vectors",
			"read:schema",
		},
		ResponseTypesSupported:            []string{"code"},
		GrantTypesSupported:               []string{"authorization_code", "refresh_token"},
		CodeChallengeMethodsSupported:     []string{"S256"},
		TokenEndpointAuthMethodsSupported: []string{"none"},
	}

	if h.config.OAuth.DCREnabled {
		metadata.RegistrationEndpoint = fmt.Sprintf("%s%s/oauth/register", baseURL, h.config.BasePath)
	}

	return c.JSON(metadata)
}

// DCR Request (RFC 7591)
type DCRRequest struct {
	ClientName    string   `json:"client_name"`
	RedirectURIs  []string `json:"redirect_uris"`
	GrantTypes    []string `json:"grant_types,omitempty"`
	ResponseTypes []string `json:"response_types,omitempty"`
	Scope         string   `json:"scope,omitempty"`
	ClientURI     string   `json:"client_uri,omitempty"`
	LogoURI       string   `json:"logo_uri,omitempty"`
	Contacts      []string `json:"contacts,omitempty"`
	TosURI        string   `json:"tos_uri,omitempty"`
	PolicyURI     string   `json:"policy_uri,omitempty"`
}

// DCR Response
type DCRResponse struct {
	ClientID          string   `json:"client_id"`
	ClientName        string   `json:"client_name"`
	RedirectURIs      []string `json:"redirect_uris"`
	GrantTypes        []string `json:"grant_types"`
	ResponseTypes     []string `json:"response_types"`
	Scope             string   `json:"scope,omitempty"`
	ClientIDIssuedAt  int64    `json:"client_id_issued_at"`
}

// handleRegister handles POST /mcp/oauth/register (Dynamic Client Registration)
func (h *MCPOAuthHandler) handleRegister(c *fiber.Ctx) error {
	if !h.config.OAuth.Enabled || !h.config.OAuth.DCREnabled {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "registration_not_supported",
			"error_description": "Dynamic Client Registration is not enabled",
		})
	}

	var req DCRRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_client_metadata",
			"error_description": "Invalid request body",
		})
	}

	// Validate required fields
	if req.ClientName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_client_metadata",
			"error_description": "client_name is required",
		})
	}

	if len(req.RedirectURIs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_redirect_uri",
			"error_description": "At least one redirect_uri is required",
		})
	}

	// Validate redirect URIs against allowed patterns
	for _, uri := range req.RedirectURIs {
		if !h.isRedirectURIAllowed(uri) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid_redirect_uri",
				"error_description": fmt.Sprintf("Redirect URI not allowed: %s", uri),
			})
		}
	}

	// Set defaults
	grantTypes := req.GrantTypes
	if len(grantTypes) == 0 {
		grantTypes = []string{"authorization_code", "refresh_token"}
	}

	responseTypes := req.ResponseTypes
	if len(responseTypes) == 0 {
		responseTypes = []string{"code"}
	}

	// Generate client_id
	clientID := fmt.Sprintf("mcp_%s", generateRandomString(24))

	// Insert into database
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		INSERT INTO auth.mcp_oauth_clients
		(client_id, client_name, redirect_uris, grant_types, response_types, scope, client_uri, logo_uri, contacts, tos_uri, policy_uri)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, clientID, req.ClientName, req.RedirectURIs, grantTypes, responseTypes, req.Scope,
		nullIfEmpty(req.ClientURI), nullIfEmpty(req.LogoURI), req.Contacts,
		nullIfEmpty(req.TosURI), nullIfEmpty(req.PolicyURI))

	if err != nil {
		log.Error().Err(err).Msg("Failed to register MCP OAuth client")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "server_error",
			"error_description": "Failed to register client",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(DCRResponse{
		ClientID:         clientID,
		ClientName:       req.ClientName,
		RedirectURIs:     req.RedirectURIs,
		GrantTypes:       grantTypes,
		ResponseTypes:    responseTypes,
		Scope:            req.Scope,
		ClientIDIssuedAt: time.Now().Unix(),
	})
}

// handleAuthorize handles GET /mcp/oauth/authorize
func (h *MCPOAuthHandler) handleAuthorize(c *fiber.Ctx) error {
	if !h.config.OAuth.Enabled {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "oauth_disabled",
			"error_description": "OAuth is not enabled",
		})
	}

	// Parse OAuth parameters
	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	responseType := c.Query("response_type")
	scope := c.Query("scope")
	state := c.Query("state")
	codeChallenge := c.Query("code_challenge")
	codeChallengeMethod := c.Query("code_challenge_method", "S256")

	// Validate required parameters
	if clientID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_request",
			"error_description": "client_id is required",
		})
	}

	if redirectURI == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_request",
			"error_description": "redirect_uri is required",
		})
	}

	if responseType != "code" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "unsupported_response_type",
			"error_description": "Only 'code' response type is supported",
		})
	}

	// PKCE is required (OAuth 2.1)
	if codeChallenge == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_request",
			"error_description": "code_challenge is required (PKCE)",
		})
	}

	if codeChallengeMethod != "S256" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_request",
			"error_description": "Only S256 code_challenge_method is supported",
		})
	}

	// Validate client exists and redirect_uri matches
	ctx := context.Background()
	var clientName string
	var registeredURIs []string
	err := h.db.QueryRow(ctx, `
		SELECT client_name, redirect_uris FROM auth.mcp_oauth_clients WHERE client_id = $1
	`, clientID).Scan(&clientName, &registeredURIs)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_client",
			"error_description": "Client not found",
		})
	}

	// Validate redirect_uri
	uriValid := false
	for _, uri := range registeredURIs {
		if matchRedirectURI(uri, redirectURI) {
			uriValid = true
			break
		}
	}
	if !uriValid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_redirect_uri",
			"error_description": "Redirect URI does not match registered URIs",
		})
	}

	// Check if user is authenticated
	userID := c.Locals("user_id")
	if userID == nil {
		// Redirect to login with return URL
		loginURL := fmt.Sprintf("%s/login?return_to=%s", h.publicURL, c.OriginalURL())
		return c.Redirect(loginURL, fiber.StatusFound)
	}

	// Parse requested scopes
	requestedScopes := strings.Fields(scope)
	if len(requestedScopes) == 0 {
		requestedScopes = []string{"read:tables"}
	}

	// Return consent page HTML
	return c.Type("html").Send([]byte(h.renderConsentPage(
		clientName, clientID, redirectURI, requestedScopes, state, codeChallenge, codeChallengeMethod,
	)))
}

// handleAuthorizeConsent handles POST /mcp/oauth/authorize (consent form submission)
func (h *MCPOAuthHandler) handleAuthorizeConsent(c *fiber.Ctx) error {
	if !h.config.OAuth.Enabled {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "oauth_disabled",
		})
	}

	// Get form data
	clientID := c.FormValue("client_id")
	redirectURI := c.FormValue("redirect_uri")
	scope := c.FormValue("scope")
	state := c.FormValue("state")
	codeChallenge := c.FormValue("code_challenge")
	codeChallengeMethod := c.FormValue("code_challenge_method")
	action := c.FormValue("action")

	// Get user ID from auth
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Handle denial
	if action == "deny" {
		return c.Redirect(fmt.Sprintf("%s?error=access_denied&state=%s", redirectURI, state), fiber.StatusFound)
	}

	// Generate authorization code
	code := generateRandomString(32)

	// Store authorization code
	ctx := context.Background()
	expiresAt := time.Now().Add(10 * time.Minute) // Codes expire in 10 minutes

	_, err := h.db.Exec(ctx, `
		INSERT INTO auth.mcp_oauth_codes
		(code, client_id, user_id, redirect_uri, scope, code_challenge, code_challenge_method, state, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, code, clientID, userID, redirectURI, scope, codeChallenge, codeChallengeMethod, state, expiresAt)

	if err != nil {
		log.Error().Err(err).Msg("Failed to store authorization code")
		return c.Redirect(fmt.Sprintf("%s?error=server_error&state=%s", redirectURI, state), fiber.StatusFound)
	}

	// Redirect with code
	redirectURL := fmt.Sprintf("%s?code=%s", redirectURI, code)
	if state != "" {
		redirectURL += "&state=" + state
	}

	return c.Redirect(redirectURL, fiber.StatusFound)
}

// TokenRequest represents an OAuth token request
type TokenRequest struct {
	GrantType    string `json:"grant_type" form:"grant_type"`
	Code         string `json:"code" form:"code"`
	RedirectURI  string `json:"redirect_uri" form:"redirect_uri"`
	ClientID     string `json:"client_id" form:"client_id"`
	CodeVerifier string `json:"code_verifier" form:"code_verifier"`
	RefreshToken string `json:"refresh_token" form:"refresh_token"`
}

// TokenResponse represents an OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope"`
}

// handleToken handles POST /mcp/oauth/token
func (h *MCPOAuthHandler) handleToken(c *fiber.Ctx) error {
	if !h.config.OAuth.Enabled {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "oauth_disabled",
		})
	}

	var req TokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_request",
			"error_description": "Invalid request body",
		})
	}

	switch req.GrantType {
	case "authorization_code":
		return h.handleAuthorizationCodeGrant(c, &req)
	case "refresh_token":
		return h.handleRefreshTokenGrant(c, &req)
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "unsupported_grant_type",
			"error_description": "Only authorization_code and refresh_token grants are supported",
		})
	}
}

func (h *MCPOAuthHandler) handleAuthorizationCodeGrant(c *fiber.Ctx, req *TokenRequest) error {
	if req.Code == "" || req.ClientID == "" || req.RedirectURI == "" || req.CodeVerifier == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_request",
			"error_description": "Missing required parameters",
		})
	}

	// Look up and validate authorization code
	ctx := context.Background()
	var userID uuid.UUID
	var storedRedirectURI, storedScope, codeChallenge, codeChallengeMethod string
	var storedClientID string
	var expiresAt time.Time
	var usedAt *time.Time

	err := h.db.QueryRow(ctx, `
		SELECT client_id, user_id, redirect_uri, scope, code_challenge, code_challenge_method, expires_at, used_at
		FROM auth.mcp_oauth_codes WHERE code = $1
	`, req.Code).Scan(&storedClientID, &userID, &storedRedirectURI, &storedScope, &codeChallenge, &codeChallengeMethod, &expiresAt, &usedAt)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_grant",
			"error_description": "Invalid authorization code",
		})
	}

	// Check if code was already used (replay attack prevention)
	if usedAt != nil {
		// Revoke all tokens for this authorization (security measure)
		_, _ = h.db.Exec(ctx, `
			UPDATE auth.mcp_oauth_tokens SET revoked_at = NOW()
			WHERE client_id = $1 AND user_id = $2 AND revoked_at IS NULL
		`, storedClientID, userID)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_grant",
			"error_description": "Authorization code already used",
		})
	}

	// Mark code as used
	_, _ = h.db.Exec(ctx, `UPDATE auth.mcp_oauth_codes SET used_at = NOW() WHERE code = $1`, req.Code)

	// Check expiration
	if time.Now().After(expiresAt) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_grant",
			"error_description": "Authorization code expired",
		})
	}

	// Validate client_id
	if storedClientID != req.ClientID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_grant",
			"error_description": "Client ID mismatch",
		})
	}

	// Validate redirect_uri
	if storedRedirectURI != req.RedirectURI {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_grant",
			"error_description": "Redirect URI mismatch",
		})
	}

	// Verify PKCE code_verifier
	if !verifyPKCE(req.CodeVerifier, codeChallenge, codeChallengeMethod) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_grant",
			"error_description": "Invalid code_verifier",
		})
	}

	// Generate tokens
	accessToken := generateRandomString(48)
	refreshToken := generateRandomString(48)

	// Hash tokens for storage
	accessTokenHash := hashToken(accessToken)
	refreshTokenHash := hashToken(refreshToken)

	// Calculate expiry times
	accessExpiry := time.Now().Add(h.config.OAuth.TokenExpiry)
	refreshExpiry := time.Now().Add(h.config.OAuth.RefreshTokenExpiry)

	// Store tokens
	_, err = h.db.Exec(ctx, `
		INSERT INTO auth.mcp_oauth_tokens
		(client_id, user_id, access_token_hash, refresh_token_hash, scope, expires_at, refresh_expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, req.ClientID, userID, accessTokenHash, refreshTokenHash, storedScope, accessExpiry, refreshExpiry)

	if err != nil {
		log.Error().Err(err).Msg("Failed to store OAuth tokens")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "server_error",
		})
	}

	return c.JSON(TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(h.config.OAuth.TokenExpiry.Seconds()),
		RefreshToken: refreshToken,
		Scope:        storedScope,
	})
}

func (h *MCPOAuthHandler) handleRefreshTokenGrant(c *fiber.Ctx, req *TokenRequest) error {
	if req.RefreshToken == "" || req.ClientID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_request",
			"error_description": "Missing required parameters",
		})
	}

	refreshTokenHash := hashToken(req.RefreshToken)

	// Look up refresh token
	ctx := context.Background()
	var tokenID uuid.UUID
	var userID uuid.UUID
	var storedClientID, scope string
	var refreshExpiresAt time.Time
	var revokedAt *time.Time

	err := h.db.QueryRow(ctx, `
		SELECT id, client_id, user_id, scope, refresh_expires_at, revoked_at
		FROM auth.mcp_oauth_tokens WHERE refresh_token_hash = $1
	`, refreshTokenHash).Scan(&tokenID, &storedClientID, &userID, &scope, &refreshExpiresAt, &revokedAt)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_grant",
			"error_description": "Invalid refresh token",
		})
	}

	// Check if revoked
	if revokedAt != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_grant",
			"error_description": "Refresh token has been revoked",
		})
	}

	// Check expiration
	if time.Now().After(refreshExpiresAt) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_grant",
			"error_description": "Refresh token expired",
		})
	}

	// Validate client_id
	if storedClientID != req.ClientID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_grant",
			"error_description": "Client ID mismatch",
		})
	}

	// Revoke old token
	_, _ = h.db.Exec(ctx, `UPDATE auth.mcp_oauth_tokens SET revoked_at = NOW() WHERE id = $1`, tokenID)

	// Generate new tokens (token rotation)
	newAccessToken := generateRandomString(48)
	newRefreshToken := generateRandomString(48)

	accessTokenHash := hashToken(newAccessToken)
	newRefreshTokenHash := hashToken(newRefreshToken)

	accessExpiry := time.Now().Add(h.config.OAuth.TokenExpiry)
	newRefreshExpiry := time.Now().Add(h.config.OAuth.RefreshTokenExpiry)

	// Store new tokens
	_, err = h.db.Exec(ctx, `
		INSERT INTO auth.mcp_oauth_tokens
		(client_id, user_id, access_token_hash, refresh_token_hash, scope, expires_at, refresh_expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, req.ClientID, userID, accessTokenHash, newRefreshTokenHash, scope, accessExpiry, newRefreshExpiry)

	if err != nil {
		log.Error().Err(err).Msg("Failed to store refreshed OAuth tokens")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "server_error",
		})
	}

	return c.JSON(TokenResponse{
		AccessToken:  newAccessToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(h.config.OAuth.TokenExpiry.Seconds()),
		RefreshToken: newRefreshToken,
		Scope:        scope,
	})
}

// handleRevoke handles POST /mcp/oauth/revoke
func (h *MCPOAuthHandler) handleRevoke(c *fiber.Ctx) error {
	if !h.config.OAuth.Enabled {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "oauth_disabled",
		})
	}

	token := c.FormValue("token")
	tokenTypeHint := c.FormValue("token_type_hint")

	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_request",
			"error_description": "token parameter is required",
		})
	}

	tokenHash := hashToken(token)
	ctx := context.Background()

	// Try to revoke as access token or refresh token
	var result int64
	if tokenTypeHint == "refresh_token" {
		res, _ := h.db.Exec(ctx, `
			UPDATE auth.mcp_oauth_tokens SET revoked_at = NOW()
			WHERE refresh_token_hash = $1 AND revoked_at IS NULL
		`, tokenHash)
		result = res.RowsAffected()
	}

	if result == 0 {
		res, _ := h.db.Exec(ctx, `
			UPDATE auth.mcp_oauth_tokens SET revoked_at = NOW()
			WHERE access_token_hash = $1 AND revoked_at IS NULL
		`, tokenHash)
		result = res.RowsAffected()
	}

	if result == 0 && tokenTypeHint != "refresh_token" {
		res, _ := h.db.Exec(ctx, `
			UPDATE auth.mcp_oauth_tokens SET revoked_at = NOW()
			WHERE refresh_token_hash = $1 AND revoked_at IS NULL
		`, tokenHash)
		result = res.RowsAffected()
	}

	// RFC 7009: Always return 200 OK, even if token was not found
	return c.SendStatus(fiber.StatusOK)
}

// Helper functions

func (h *MCPOAuthHandler) isRedirectURIAllowed(uri string) bool {
	for _, pattern := range h.config.OAuth.AllowedRedirectURIs {
		if matchRedirectURI(pattern, uri) {
			return true
		}
	}
	return false
}

func matchRedirectURI(pattern, uri string) bool {
	// Exact match
	if pattern == uri {
		return true
	}

	// Wildcard matching for localhost ports
	if strings.HasSuffix(pattern, ":*") {
		prefix := strings.TrimSuffix(pattern, "*")
		if strings.HasPrefix(uri, prefix) {
			return true
		}
	}

	// Wildcard matching for paths (e.g., cursor://anysphere.cursor-mcp/oauth/*/callback)
	if strings.Contains(pattern, "*") {
		// Convert glob pattern to prefix/suffix matching
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			if strings.HasPrefix(uri, parts[0]) && strings.HasSuffix(uri, parts[1]) {
				return true
			}
		}
	}

	return false
}

func generateRandomString(length int) string {
	bytes := make([]byte, length)
	_, _ = rand.Read(bytes)
	return base64.RawURLEncoding.EncodeToString(bytes)[:length]
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func verifyPKCE(codeVerifier, codeChallenge, method string) bool {
	if method != "S256" {
		return false
	}

	// S256: BASE64URL(SHA256(code_verifier)) == code_challenge
	hash := sha256.Sum256([]byte(codeVerifier))
	computed := base64.RawURLEncoding.EncodeToString(hash[:])

	return computed == codeChallenge
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (h *MCPOAuthHandler) renderConsentPage(clientName, clientID, redirectURI string, scopes []string, state, codeChallenge, codeChallengeMethod string) string {
	scopeList := ""
	for _, s := range scopes {
		scopeList += fmt.Sprintf("<li>%s</li>", s)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Authorize %s - Fluxbase</title>
    <style>
        * { box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #0f172a; color: #e2e8f0; margin: 0; padding: 20px;
            min-height: 100vh; display: flex; align-items: center; justify-content: center;
        }
        .container { max-width: 400px; width: 100%%; }
        .card { background: #1e293b; border-radius: 12px; padding: 32px; }
        h1 { margin: 0 0 8px; font-size: 24px; font-weight: 600; }
        .subtitle { color: #94a3b8; margin: 0 0 24px; }
        .client-name { color: #60a5fa; font-weight: 500; }
        h2 { font-size: 14px; font-weight: 500; margin: 0 0 12px; color: #94a3b8; }
        ul { margin: 0 0 24px; padding-left: 20px; }
        li { margin: 8px 0; color: #e2e8f0; }
        .buttons { display: flex; gap: 12px; }
        button {
            flex: 1; padding: 12px 20px; border: none; border-radius: 8px;
            font-size: 14px; font-weight: 500; cursor: pointer; transition: all 0.2s;
        }
        .allow { background: #3b82f6; color: white; }
        .allow:hover { background: #2563eb; }
        .deny { background: #334155; color: #e2e8f0; }
        .deny:hover { background: #475569; }
    </style>
</head>
<body>
    <div class="container">
        <div class="card">
            <h1>Authorization Request</h1>
            <p class="subtitle"><span class="client-name">%s</span> wants to access your Fluxbase data</p>

            <h2>Requested permissions:</h2>
            <ul>%s</ul>

            <form method="POST" action="%s/oauth/authorize">
                <input type="hidden" name="client_id" value="%s">
                <input type="hidden" name="redirect_uri" value="%s">
                <input type="hidden" name="scope" value="%s">
                <input type="hidden" name="state" value="%s">
                <input type="hidden" name="code_challenge" value="%s">
                <input type="hidden" name="code_challenge_method" value="%s">

                <div class="buttons">
                    <button type="submit" name="action" value="deny" class="deny">Deny</button>
                    <button type="submit" name="action" value="allow" class="allow">Allow</button>
                </div>
            </form>
        </div>
    </div>
</body>
</html>`, clientName, clientName, scopeList, h.config.BasePath, clientID, redirectURI,
		strings.Join(scopes, " "), state, codeChallenge, codeChallengeMethod)
}
