package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"

	apperrors "github.com/nimbleflux/fluxbase/internal/errors"
	"github.com/nimbleflux/fluxbase/internal/settings"
)

// RequireAdmin middleware restricts access to admin users only
// Allows: service_role (from service keys or service_role JWT) and instance_admin users
// This should be used after authentication middleware (RequireAuthOrServiceKey)
func RequireAdmin() fiber.Handler {
	return func(c fiber.Ctx) error {
		authType, _ := c.Locals("auth_type").(string)
		role, _ := c.Locals("user_role").(string)

		// Service keys always have service_role
		if authType == "service_key" {
			log.Debug().
				Str("auth_type", authType).
				Msg("Admin access granted - service key")
			return c.Next()
		}

		// Check for admin roles
		if role == "service_role" || role == "instance_admin" || role == "tenant_service" {
			log.Debug().
				Str("auth_type", authType).
				Str("role", role).
				Msg("Admin access granted")
			return c.Next()
		}

		log.Warn().
			Str("auth_type", authType).
			Str("role", role).
			Str("path", c.Path()).
			Msg("Admin access denied - requires service_role or instance_admin")

		return apperrors.SendAdminRequired(c)
	}
}

// RequireAdminIfClientKeysDisabled middleware conditionally requires admin access
// when the 'app.auth.allow_user_client_keys' setting is disabled.
// If the setting is enabled (default), allows regular users through.
// If the setting is disabled, requires admin access (service_role or instance_admin).
func RequireAdminIfClientKeysDisabled(settingsCache *settings.SettingsCache) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Check if user client keys are allowed
		allowUserKeys := settingsCache.GetBool(c.RequestCtx(), "app.auth.allow_user_client_keys", true)

		if allowUserKeys {
			// Setting is enabled - allow regular users to manage their own keys
			return c.Next()
		}

		// Setting is disabled - require admin access
		return RequireAdmin()(c)
	}
}

// fiber:context-methods migrated
