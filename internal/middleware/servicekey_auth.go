package middleware

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"

	"github.com/nimbleflux/fluxbase/internal/keys"
)

// serviceKeyInfo holds the result of a service key lookup from either key table.
type serviceKeyInfo struct {
	keyID              string
	keyName            string
	keyType            string
	scopes             []string
	allowedNamespaces  *[]string
	isActive           bool
	expiresAt          *time.Time
	rateLimitPerMinute *int
	tenantID           *string
	lookupTable        string // "auth" or "platform"
}

// serviceKeyAuthenticator groups the service key lookup helpers so they can be
// backed by a stub in tests without changing their call signatures.
type serviceKeyAuthenticator struct{}

// validateServiceKey validates a service key against auth.service_keys or platform.service_keys.
// It routes to the correct table based on key prefix and uses SET LOCAL ROLE service_role
// to bypass RLS during key lookup (necessary since no auth context exists yet).
// Returns true if valid, false otherwise.
func validateServiceKey(c fiber.Ctx, db *pgxpool.Pool, serviceKey string) bool {
	if len(serviceKey) < 8 {
		return false
	}

	// Determine which table and prefix to use based on key format
	prefix := keys.ExtractPrefix(serviceKey)
	var keyInfo *serviceKeyInfo
	var keyHash string
	authenticator := serviceKeyAuthenticator{}

	if prefix != "" {
		// New-style key (fb_tsk_, fb_anon_, fb_gsk_, fb_pk_) -> platform.service_keys
		info, hash, err := authenticator.lookupPlatformServiceKey(c, db, serviceKey)
		if err != nil {
			log.Debug().Err(err).Str("prefix", prefix).Msg("Platform service key lookup failed")
			return false
		}
		keyInfo = info
		keyHash = hash
	} else if strings.HasPrefix(serviceKey, "sk_") || strings.HasPrefix(serviceKey, "pk_") {
		// Legacy key (sk_, pk_) -> auth.service_keys
		info, hash, err := authenticator.lookupAuthServiceKey(c, db, serviceKey)
		if err != nil {
			log.Debug().Err(err).Str("key_prefix", serviceKey[:min(16, len(serviceKey))]).Msg("Auth service key lookup failed")
			return false
		}
		keyInfo = info
		keyHash = hash
	} else {
		return false
	}

	// Verify the key is active
	if !keyInfo.isActive {
		log.Debug().Str("key_id", keyInfo.keyID).Msg("Service key is disabled")
		return false
	}

	// Verify not expired
	if keyInfo.expiresAt != nil && keyInfo.expiresAt.Before(time.Now()) {
		log.Debug().Str("key_id", keyInfo.keyID).Msg("Service key has expired")
		return false
	}

	// Verify the key hash
	if err := bcrypt.CompareHashAndPassword([]byte(keyHash), []byte(serviceKey)); err != nil {
		log.Debug().Err(err).Str("key_id", keyInfo.keyID).Msg("Invalid service key hash")
		return false
	}

	// Map key type to application role
	role := mapKeyTypetoRole(keyInfo.keyType)

	// Fire-and-forget last_used_at update
	updateLastUsedAt(c, db, keyInfo)

	// Store service key information in context
	c.Locals("service_key_id", keyInfo.keyID)
	c.Locals("service_key_name", keyInfo.keyName)
	c.Locals("service_key_scopes", keyInfo.scopes)
	c.Locals("service_key_type", keyInfo.keyType)
	c.Locals("auth_type", "service_key")
	c.Locals("user_role", role)

	// Store rate limits in context
	c.Locals("service_key_rate_limit_per_minute", keyInfo.rateLimitPerMinute)

	// Store allowed namespaces if present
	if keyInfo.allowedNamespaces != nil {
		c.Locals("allowed_namespaces", *keyInfo.allowedNamespaces)
	}

	// Store tenant_id from key if present (needed for tenant_service keys)
	if keyInfo.tenantID != nil {
		c.Locals("service_key_tenant_id", *keyInfo.tenantID)
	}

	// For RLS context
	c.Locals("rls_role", role)
	c.Locals("rls_user_id", nil)

	log.Debug().
		Str("key_id", keyInfo.keyID).
		Str("key_name", keyInfo.keyName).
		Str("key_type", keyInfo.keyType).
		Str("role", role).
		Str("lookup_table", keyInfo.lookupTable).
		Interface("rate_limit_per_minute", keyInfo.rateLimitPerMinute).
		Msg("Authenticated with service key")

	return true
}

// lookupPlatformServiceKey looks up a key in platform.service_keys using SET LOCAL ROLE service_role
// to bypass RLS (key validation happens before auth context is set).
func (a serviceKeyAuthenticator) lookupPlatformServiceKey(c fiber.Ctx, db *pgxpool.Pool, serviceKey string) (*serviceKeyInfo, string, error) {
	keyPrefix := keys.ExtractPrefix(serviceKey)
	if keyPrefix == "" || len(serviceKey) < len(keyPrefix)+8 {
		return nil, "", errors.New("invalid platform key format")
	}
	prefixForLookup := serviceKey[:len(keyPrefix)+8]

	// Use tenant pool if available (database-per-tenant), otherwise use main pool
	pool := GetTenantPool(c)
	if pool == nil {
		pool = db
	}

	var info serviceKeyInfo
	var keyHash string

	err := a.queryWithServiceRole(c.RequestCtx(), pool, func(tx pgx.Tx) error {
		return tx.QueryRow(
			c.RequestCtx(),
			`SELECT id, name, key_hash, key_type, scopes, allowed_namespaces,
			        is_active, expires_at, rate_limit_per_minute, tenant_id
			 FROM platform.service_keys
			 WHERE key_prefix = $1`,
			prefixForLookup,
		).Scan(&info.keyID, &info.keyName, &keyHash, &info.keyType, &info.scopes,
			&info.allowedNamespaces, &info.isActive, &info.expiresAt,
			&info.rateLimitPerMinute, &info.tenantID)
	})
	if err != nil {
		return nil, "", fmt.Errorf("platform key lookup failed: %w", err)
	}

	info.lookupTable = "platform"
	return &info, keyHash, nil
}

// lookupAuthServiceKey looks up a key in auth.service_keys using SET LOCAL ROLE service_role
// to bypass RLS (key validation happens before auth context is set).
func (a serviceKeyAuthenticator) lookupAuthServiceKey(c fiber.Ctx, db *pgxpool.Pool, serviceKey string) (*serviceKeyInfo, string, error) {
	if len(serviceKey) < 16 {
		return nil, "", errors.New("auth key too short")
	}
	keyPrefix := serviceKey[:16]

	// Use tenant pool if available (database-per-tenant), otherwise use main pool
	pool := GetTenantPool(c)
	if pool == nil {
		pool = db
	}

	var info serviceKeyInfo
	var keyHash string
	var revokedAt *time.Time
	var rateLimitPerHour *int

	err := a.queryWithServiceRole(c.RequestCtx(), pool, func(tx pgx.Tx) error {
		return tx.QueryRow(
			c.RequestCtx(),
			`SELECT id, name, key_hash, COALESCE(key_type, 'service'), scopes, allowed_namespaces,
			        enabled, expires_at, rate_limit_per_minute, rate_limit_per_hour,
			        revoked_at, tenant_id
			 FROM auth.service_keys
			 WHERE key_prefix = $1`,
			keyPrefix,
		).Scan(&info.keyID, &info.keyName, &keyHash, &info.keyType, &info.scopes,
			&info.allowedNamespaces, &info.isActive, &info.expiresAt,
			&info.rateLimitPerMinute, &rateLimitPerHour, &revokedAt, &info.tenantID)
	})
	if err != nil {
		return nil, "", fmt.Errorf("auth key lookup failed: %w", err)
	}

	// Check revoked (handled inline for legacy table)
	if revokedAt != nil {
		info.isActive = false
	}

	info.lookupTable = "auth"
	return &info, keyHash, nil
}

// queryWithServiceRole executes a read query within a transaction with SET LOCAL ROLE service_role.
// This bypasses RLS for key lookup operations that happen before auth context is established.
func (serviceKeyAuthenticator) queryWithServiceRole(ctx context.Context, pool *pgxpool.Pool, fn func(tx pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `SET LOCAL ROLE "service_role"`)
	if err != nil {
		return fmt.Errorf("failed to SET LOCAL ROLE service_role: %w", err)
	}

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// mapKeyTypetoRole maps a service key type to the application role used for RLS.
func mapKeyTypetoRole(keyType string) string {
	switch keyType {
	case keys.KeyTypeAnon:
		return "anon"
	case keys.KeyTypeTenantService:
		return "tenant_service"
	case keys.KeyTypeGlobalService, "service":
		return "service_role"
	case keys.KeyTypePublishable:
		// Publishable keys carry no user identity, so they map to the
		// anonymous role. Policies keyed on the authenticated role require a
		// real user identity (rls_user_id) which these keys never provide.
		return "anon"
	default:
		log.Warn().Str("key_type", keyType).Msg("mapKeyTypetoRole: unrecognized key type, defaulting to anon")
		return "anon"
	}
}

// updateLastUsedAt fires a background goroutine to update the last_used_at timestamp.
// Uses SET LOCAL ROLE service_role to bypass RLS on both key tables.
func updateLastUsedAt(c fiber.Ctx, db *pgxpool.Pool, info *serviceKeyInfo) {
	table := "auth.service_keys"
	if info.lookupTable == "platform" {
		table = "platform.service_keys"
	}

	keyID := info.keyID
	// Capture pool reference before spawning goroutine to avoid data race
	// on fiber.Ctx (the context is not safe for concurrent access).
	pool := GetTenantPool(c)
	if pool == nil {
		pool = db
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tx, err := pool.Begin(ctx)
		if err != nil {
			log.Debug().Err(err).Str("key_id", keyID).Str("table", table).Msg("Failed to begin transaction for last_used_at update")
			return
		}
		defer func() { _ = tx.Rollback(ctx) }()

		if _, err := tx.Exec(ctx, `SET LOCAL ROLE "service_role"`); err != nil {
			log.Debug().Err(err).Str("key_id", keyID).Str("table", table).Msg("Failed to set service role for last_used_at update")
			return
		}

		if _, err := tx.Exec(
			ctx,
			`UPDATE `+table+` SET last_used_at = NOW() WHERE id = $1`,
			keyID,
		); err != nil {
			log.Debug().Err(err).Str("key_id", keyID).Str("table", table).Msg("Failed to update last_used_at")
			return
		}

		if err := tx.Commit(ctx); err != nil {
			log.Debug().Err(err).Str("key_id", keyID).Str("table", table).Msg("Failed to commit last_used_at update")
		}
	}()
}
