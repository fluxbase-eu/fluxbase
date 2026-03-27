package routes

import (
	"github.com/gofiber/fiber/v3"
)

// PlatformServiceKeysAdminDeps contains dependencies for platform service keys admin routes.
// These routes manage keys in the platform.service_keys table which supports:
//   - anon: Anonymous keys for public access
//   - publishable: User-scoped keys
//   - tenant_service: Tenant-level service keys
//   - global_service: Instance-level (platform) service keys
//
// Auth middleware is inherited from the parent admin route group.
//
// Role Access:
//   - instance_admin: Full access to all platform service keys
//   - tenant_admin: Access to tenant-scoped service keys (RLS enforced)
type PlatformServiceKeysAdminDeps struct {
	ListPlatformServiceKeys   fiber.Handler
	GetPlatformServiceKey     fiber.Handler
	CreatePlatformServiceKey  fiber.Handler
	UpdatePlatformServiceKey  fiber.Handler
	DeletePlatformServiceKey  fiber.Handler
	DisablePlatformServiceKey fiber.Handler
	EnablePlatformServiceKey  fiber.Handler
	RotatePlatformServiceKey  fiber.Handler
}

// BuildPlatformServiceKeysAdminRoutes creates the platform service keys admin route group.
func BuildPlatformServiceKeysAdminRoutes(deps *PlatformServiceKeysAdminDeps) *RouteGroup {
	if deps == nil {
		return nil
	}

	return &RouteGroup{
		Name:         "platform_service_keys_admin",
		DefaultAuth:  AuthRequired,
		DefaultRoles: []string{"admin", "instance_admin", "tenant_admin"},
		Routes: []Route{
			{Method: "GET", Path: "/service-keys", Handler: deps.ListPlatformServiceKeys, Summary: "List platform service keys"},
			{Method: "POST", Path: "/service-keys", Handler: deps.CreatePlatformServiceKey, Summary: "Create platform service key"},
			{Method: "GET", Path: "/service-keys/:id", Handler: deps.GetPlatformServiceKey, Summary: "Get platform service key"},
			{Method: "PATCH", Path: "/service-keys/:id", Handler: deps.UpdatePlatformServiceKey, Summary: "Update platform service key"},
			{Method: "DELETE", Path: "/service-keys/:id", Handler: deps.DeletePlatformServiceKey, Summary: "Delete platform service key"},
			{Method: "POST", Path: "/service-keys/:id/disable", Handler: deps.DisablePlatformServiceKey, Summary: "Disable platform service key"},
			{Method: "POST", Path: "/service-keys/:id/enable", Handler: deps.EnablePlatformServiceKey, Summary: "Enable platform service key"},
			{Method: "POST", Path: "/service-keys/:id/rotate", Handler: deps.RotatePlatformServiceKey, Summary: "Rotate platform service key"},
		},
	}
}
