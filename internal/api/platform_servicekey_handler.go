package api

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

// PlatformServiceKeyHandler handles platform service key management requests
type PlatformServiceKeyHandler struct {
	db *pgxpool.Pool
}

// NewPlatformServiceKeyHandler creates a new platform service key handler
func NewPlatformServiceKeyHandler(db *pgxpool.Pool) *PlatformServiceKeyHandler {
	return &PlatformServiceKeyHandler{
		db: db,
	}
}

// PlatformServiceKey represents a service key in the platform.service_keys table
type PlatformServiceKey struct {
	ID                 uuid.UUID  `json:"id"`
	Name               string     `json:"name"`
	Description        *string    `json:"description,omitempty"`
	KeyType            string     `json:"key_type"`
	TenantID           *uuid.UUID `json:"tenant_id,omitempty"`
	KeyPrefix          string     `json:"key_prefix"`
	Scopes             []string   `json:"scopes"`
	AllowedNamespaces  []string   `json:"allowed_namespaces,omitempty"`
	IsActive           bool       `json:"is_active"`
	IsConfigManaged    bool       `json:"is_config_managed,omitempty"`
	RateLimitPerMinute *int       `json:"rate_limit_per_minute,omitempty"`
	CreatedBy          *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at,omitempty"`
	LastUsedAt         *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt          *time.Time `json:"expires_at,omitempty"`
	RevokedAt          *time.Time `json:"revoked_at,omitempty"`
	RevokedBy          *uuid.UUID `json:"revoked_by,omitempty"`
	RevocationReason   *string    `json:"revocation_reason,omitempty"`
	DeprecatedAt       *time.Time `json:"deprecated_at,omitempty"`
	GracePeriodEndsAt  *time.Time `json:"grace_period_ends_at,omitempty"`
	ReplacedBy         *uuid.UUID `json:"replaced_by,omitempty"`
}

// PlatformServiceKeyWithPlaintext is returned only on creation, includes the plaintext key
type PlatformServiceKeyWithPlaintext struct {
	PlatformServiceKey
	Key               string     `json:"key,omitempty"`
	GracePeriodEndsAt *time.Time `json:"grace_period_ends_at,omitempty"`
}

// CreatePlatformServiceKeyRequest represents a request to create a platform service key
type CreatePlatformServiceKeyRequest struct {
	Name               string     `json:"name"`
	Description        *string    `json:"description,omitempty"`
	KeyType            string     `json:"key_type"`
	TenantID           *uuid.UUID `json:"tenant_id,omitempty"`
	Scopes             []string   `json:"scopes,omitempty"`
	AllowedNamespaces  []string   `json:"allowed_namespaces,omitempty"`
	RateLimitPerMinute *int       `json:"rate_limit_per_minute,omitempty"`
	ExpiresAt          *time.Time `json:"expires_at,omitempty"`
}

// UpdatePlatformServiceKeyRequest represents a request to update a platform service key
type UpdatePlatformServiceKeyRequest struct {
	Name               *string  `json:"name,omitempty"`
	Description        *string  `json:"description,omitempty"`
	Scopes             []string `json:"scopes,omitempty"`
	AllowedNamespaces  []string `json:"allowed_namespaces,omitempty"`
	IsActive           *bool    `json:"is_active,omitempty"`
	RateLimitPerMinute *int     `json:"rate_limit_per_minute,omitempty"`
}

// RotatePlatformServiceKeyRequest represents a request to rotate a platform service key
type RotatePlatformServiceKeyRequest struct {
	GracePeriodHours *int     `json:"grace_period_hours,omitempty"`
	NewKeyName       *string  `json:"new_key_name,omitempty"`
	NewScopes        []string `json:"new_scopes,omitempty"`
}

// ListPlatformServiceKeys lists all platform service keys
func (h *PlatformServiceKeyHandler) ListPlatformServiceKeys(c fiber.Ctx) error {
	if h.db == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not initialized",
		})
	}

	rows, err := h.db.Query(c.RequestCtx(), `
		SELECT id, name, description, key_type, tenant_id, key_prefix, scopes, allowed_namespaces,
		       is_active, is_config_managed, rate_limit_per_minute,
		       created_by, created_at, updated_at, last_used_at, expires_at,
		       revoked_at, revoked_by, revocation_reason, deprecated_at, grace_period_ends_at, replaced_by
		FROM platform.service_keys
		ORDER BY created_at DESC
	`)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list platform service keys")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list platform service keys",
		})
	}
	defer rows.Close()

	var keys []PlatformServiceKey
	for rows.Next() {
		var key PlatformServiceKey
		err := rows.Scan(
			&key.ID, &key.Name, &key.Description, &key.KeyType, &key.TenantID, &key.KeyPrefix,
			&key.Scopes, &key.AllowedNamespaces, &key.IsActive, &key.IsConfigManaged,
			&key.RateLimitPerMinute, &key.CreatedBy, &key.CreatedAt, &key.UpdatedAt,
			&key.LastUsedAt, &key.ExpiresAt, &key.RevokedAt, &key.RevokedBy,
			&key.RevocationReason, &key.DeprecatedAt, &key.GracePeriodEndsAt, &key.ReplacedBy,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan platform service key")
			continue
		}
		keys = append(keys, key)
	}

	if keys == nil {
		keys = []PlatformServiceKey{}
	}

	return c.JSON(keys)
}

// GetPlatformServiceKey retrieves a single platform service key
func (h *PlatformServiceKeyHandler) GetPlatformServiceKey(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid service key ID",
		})
	}

	if h.db == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not initialized",
		})
	}

	var key PlatformServiceKey
	err = h.db.QueryRow(c.RequestCtx(), `
		SELECT id, name, description, key_type, tenant_id, key_prefix, scopes, allowed_namespaces,
		       is_active, is_config_managed, rate_limit_per_minute,
		       created_by, created_at, updated_at, last_used_at, expires_at,
		       revoked_at, revoked_by, revocation_reason, deprecated_at, grace_period_ends_at, replaced_by
		FROM platform.service_keys
		WHERE id = $1
	`, id).Scan(
		&key.ID, &key.Name, &key.Description, &key.KeyType, &key.TenantID, &key.KeyPrefix,
		&key.Scopes, &key.AllowedNamespaces, &key.IsActive, &key.IsConfigManaged,
		&key.RateLimitPerMinute, &key.CreatedBy, &key.CreatedAt, &key.UpdatedAt,
		&key.LastUsedAt, &key.ExpiresAt, &key.RevokedAt, &key.RevokedBy,
		&key.RevocationReason, &key.DeprecatedAt, &key.GracePeriodEndsAt, &key.ReplacedBy,
	)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Service key not found",
		})
	}

	return c.JSON(key)
}

// CreatePlatformServiceKey creates a new platform service key
func (h *PlatformServiceKeyHandler) CreatePlatformServiceKey(c fiber.Ctx) error {
	var req CreatePlatformServiceKeyRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name is required",
		})
	}

	// Validate key_type
	validKeyTypes := map[string]bool{
		"anon": true, "publishable": true, "tenant_service": true, "global_service": true,
	}
	if req.KeyType == "" {
		req.KeyType = "tenant_service"
	}
	if !validKeyTypes[req.KeyType] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "key_type must be 'anon', 'publishable', 'tenant_service', or 'global_service'",
		})
	}

	// Validate tenant_id for tenant_service keys
	if req.KeyType == "tenant_service" && req.TenantID == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "tenant_id is required for tenant_service keys",
		})
	}

	if h.db == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not initialized",
		})
	}

	// Generate key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate key",
		})
	}

	// Determine prefix based on key type
	prefix := "sk_live_"
	switch req.KeyType {
	case "anon":
		prefix = "pk_anon_"
	case "publishable":
		prefix = "pk_live_"
	case "global_service":
		prefix = "sk_global_"
	case "tenant_service":
		prefix = "sk_tenant_"
	}

	fullKey := prefix + base64.URLEncoding.EncodeToString(keyBytes)
	keyPrefix := fullKey[:16]

	keyHash, err := bcrypt.GenerateFromPassword([]byte(fullKey), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to hash key",
		})
	}

	// Set default scopes
	var scopes []string
	if req.Scopes != nil {
		scopes = req.Scopes
	} else {
		switch req.KeyType {
		case "anon":
			scopes = []string{"read"}
		case "publishable":
			scopes = []string{"read", "write"}
		default:
			scopes = []string{"*"}
		}
	}

	userID, _ := c.Locals("user_id").(uuid.UUID)

	var keyID uuid.UUID
	err = h.db.QueryRow(c.RequestCtx(), `
		INSERT INTO platform.service_keys (name, description, key_hash, key_prefix, key_type, tenant_id, scopes, allowed_namespaces, is_active, rate_limit_per_minute, created_by, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, true, $9, $10, $11)
		RETURNING id
	`, req.Name, req.Description, string(keyHash), keyPrefix, req.KeyType, req.TenantID, scopes, req.AllowedNamespaces, req.RateLimitPerMinute, userID, req.ExpiresAt).Scan(&keyID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create platform service key")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create platform service key",
		})
	}

	log.Info().Str("key_id", keyID.String()).Str("key_type", req.KeyType).Str("name", req.Name).Msg("Platform service key created")

	return c.Status(fiber.StatusCreated).JSON(PlatformServiceKeyWithPlaintext{
		PlatformServiceKey: PlatformServiceKey{
			ID:                 keyID,
			Name:               req.Name,
			Description:        req.Description,
			KeyType:            req.KeyType,
			TenantID:           req.TenantID,
			KeyPrefix:          keyPrefix,
			Scopes:             scopes,
			AllowedNamespaces:  req.AllowedNamespaces,
			IsActive:           true,
			RateLimitPerMinute: req.RateLimitPerMinute,
			CreatedBy:          &userID,
			CreatedAt:          time.Now(),
			ExpiresAt:          req.ExpiresAt,
		},
		Key: fullKey,
	})
}

// UpdatePlatformServiceKey updates a platform service key
func (h *PlatformServiceKeyHandler) UpdatePlatformServiceKey(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid service key ID",
		})
	}

	var req UpdatePlatformServiceKeyRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == nil && req.Description == nil && req.Scopes == nil &&
		req.AllowedNamespaces == nil && req.IsActive == nil && req.RateLimitPerMinute == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No fields to update",
		})
	}

	if h.db == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not initialized",
		})
	}

	result, err := h.db.Exec(c.RequestCtx(), `
		UPDATE platform.service_keys
		SET name = COALESCE($1, name),
		    description = COALESCE($2, description),
		    scopes = COALESCE($3, scopes),
		    allowed_namespaces = COALESCE($4, allowed_namespaces),
		    is_active = COALESCE($5, is_active),
		    rate_limit_per_minute = COALESCE($6, rate_limit_per_minute),
		    updated_at = NOW()
		WHERE id = $7
	`, req.Name, req.Description, req.Scopes, req.AllowedNamespaces, req.IsActive, req.RateLimitPerMinute, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update platform service key")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update platform service key",
		})
	}

	if result.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Service key not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Service key updated successfully",
	})
}

// DeletePlatformServiceKey deletes a platform service key
func (h *PlatformServiceKeyHandler) DeletePlatformServiceKey(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid service key ID",
		})
	}

	if h.db == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not initialized",
		})
	}

	result, err := h.db.Exec(c.RequestCtx(), `DELETE FROM platform.service_keys WHERE id = $1`, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete platform service key")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete platform service key",
		})
	}

	if result.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Service key not found",
		})
	}

	log.Info().Str("key_id", id.String()).Msg("Platform service key deleted")

	return c.SendStatus(fiber.StatusNoContent)
}

// DisablePlatformServiceKey disables a platform service key
func (h *PlatformServiceKeyHandler) DisablePlatformServiceKey(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid service key ID",
		})
	}

	if h.db == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not initialized",
		})
	}

	result, err := h.db.Exec(c.RequestCtx(), `UPDATE platform.service_keys SET is_active = false, updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to disable platform service key")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to disable platform service key",
		})
	}

	if result.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Service key not found",
		})
	}

	log.Info().Str("key_id", id.String()).Msg("Platform service key disabled")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Service key disabled successfully",
	})
}

// EnablePlatformServiceKey enables a platform service key
func (h *PlatformServiceKeyHandler) EnablePlatformServiceKey(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid service key ID",
		})
	}

	if h.db == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not initialized",
		})
	}

	result, err := h.db.Exec(c.RequestCtx(), `UPDATE platform.service_keys SET is_active = true, updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to enable platform service key")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to enable platform service key",
		})
	}

	if result.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Service key not found",
		})
	}

	log.Info().Str("key_id", id.String()).Msg("Platform service key enabled")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Service key enabled successfully",
	})
}

// RotatePlatformServiceKey rotates a platform service key
func (h *PlatformServiceKeyHandler) RotatePlatformServiceKey(c fiber.Ctx) error {
	idStr := c.Params("id")
	oldID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid service key ID",
		})
	}

	var req RotatePlatformServiceKeyRequest
	_ = c.Bind().Body(&req) // Optional body

	gracePeriodHours := 24
	if req.GracePeriodHours != nil && *req.GracePeriodHours > 0 {
		gracePeriodHours = *req.GracePeriodHours
	}

	if h.db == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not initialized",
		})
	}

	// Get old key details
	var oldKey PlatformServiceKey
	err = h.db.QueryRow(c.RequestCtx(), `
		SELECT id, name, description, key_type, tenant_id, scopes, allowed_namespaces, rate_limit_per_minute, expires_at
		FROM platform.service_keys
		WHERE id = $1
	`, oldID).Scan(
		&oldKey.ID, &oldKey.Name, &oldKey.Description, &oldKey.KeyType, &oldKey.TenantID,
		&oldKey.Scopes, &oldKey.AllowedNamespaces, &oldKey.RateLimitPerMinute, &oldKey.ExpiresAt,
	)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Service key not found",
		})
	}

	// Generate new key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate key",
		})
	}

	// Determine prefix based on key type
	prefix := "sk_live_"
	switch oldKey.KeyType {
	case "anon":
		prefix = "pk_anon_"
	case "publishable":
		prefix = "pk_live_"
	case "global_service":
		prefix = "sk_global_"
	case "tenant_service":
		prefix = "sk_tenant_"
	}

	fullKey := prefix + base64.URLEncoding.EncodeToString(keyBytes)
	keyPrefix := fullKey[:16]

	keyHash, err := bcrypt.GenerateFromPassword([]byte(fullKey), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to hash key",
		})
	}

	userID, _ := c.Locals("user_id").(uuid.UUID)

	// Determine new key name
	newName := oldKey.Name
	if req.NewKeyName != nil && *req.NewKeyName != "" {
		newName = *req.NewKeyName
	}

	// Determine new scopes
	newScopes := oldKey.Scopes
	if req.NewScopes != nil {
		newScopes = req.NewScopes
	}

	tx, err := h.db.Begin(c.RequestCtx())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to begin transaction",
		})
	}
	defer func() { _ = tx.Rollback(c.RequestCtx()) }()

	var newID uuid.UUID
	err = tx.QueryRow(c.RequestCtx(), `
		INSERT INTO platform.service_keys (name, description, key_hash, key_prefix, key_type, tenant_id, scopes, allowed_namespaces, is_active, rate_limit_per_minute, created_by, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, true, $9, $10, $11)
		RETURNING id
	`, newName, oldKey.Description, string(keyHash), keyPrefix, oldKey.KeyType, oldKey.TenantID,
		newScopes, oldKey.AllowedNamespaces, oldKey.RateLimitPerMinute, userID, oldKey.ExpiresAt).Scan(&newID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create rotated platform service key")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create rotated key",
		})
	}

	gracePeriodEndsAt := time.Now().Add(time.Duration(gracePeriodHours) * time.Hour)

	_, err = tx.Exec(c.RequestCtx(), `
		UPDATE platform.service_keys
		SET deprecated_at = NOW(),
		    grace_period_ends_at = $1,
		    replaced_by = $2,
		    updated_at = NOW()
		WHERE id = $3
	`, gracePeriodEndsAt, newID, oldID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to deprecate old platform service key")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to deprecate old key",
		})
	}

	if err := tx.Commit(c.RequestCtx()); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to commit transaction",
		})
	}

	log.Warn().Str("old_key_id", oldID.String()).Str("new_key_id", newID.String()).Msg("Platform service key rotated")

	return c.Status(fiber.StatusCreated).JSON(PlatformServiceKeyWithPlaintext{
		PlatformServiceKey: PlatformServiceKey{
			ID:                 newID,
			Name:               newName,
			Description:        oldKey.Description,
			KeyType:            oldKey.KeyType,
			TenantID:           oldKey.TenantID,
			KeyPrefix:          keyPrefix,
			Scopes:             newScopes,
			AllowedNamespaces:  oldKey.AllowedNamespaces,
			IsActive:           true,
			RateLimitPerMinute: oldKey.RateLimitPerMinute,
			CreatedBy:          &userID,
			CreatedAt:          time.Now(),
			ExpiresAt:          oldKey.ExpiresAt,
		},
		Key:               fullKey,
		GracePeriodEndsAt: &gracePeriodEndsAt,
	})
}
