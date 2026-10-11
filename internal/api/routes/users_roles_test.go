package routes

import (
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// usersNoop is a placeholder handler for route construction in tests.
func usersNoop(c fiber.Ctx) error { return c.SendString("ok") }

// TestUserMutationRoutesRequireInstanceAdmin verifies that the user-management
// routes whose underlying updates run with elevated privileges and no tenant
// predicate are restricted to instance-level admins and are not reachable by
// tenant admins (#401, post-merge rescan of #369).
func TestUserMutationRoutesRequireInstanceAdmin(t *testing.T) {
	deps := &UsersAdminDeps{
		ListUsers:           usersNoop,
		InviteUser:          usersNoop,
		DeleteUser:          usersNoop,
		UpdateUser:          usersNoop,
		UpdateUserRole:      usersNoop,
		ResetUserPassword:   usersNoop,
		ListUsersWithQuotas: usersNoop,
		GetUserQuota:        usersNoop,
		SetUserQuota:        usersNoop,
		CreateInvitation:    usersNoop,
		ListInvitations:     usersNoop,
		RevokeInvitation:    usersNoop,
	}

	group := BuildUsersAdminRoutes(deps)
	require.NotNil(t, group)

	instanceAdminOnly := []string{"admin", "instance_admin"}

	for _, route := range group.Routes {
		isMutation := route.Method == "PATCH" && route.Path == "/users/:id" ||
			route.Method == "POST" && route.Path == "/users/:id/reset-password"

		if isMutation {
			assert.Equal(t, instanceAdminOnly, route.Roles,
				"%s %s must be restricted to instance-level admins", route.Method, route.Path)
			continue
		}

		// Every id-scoped user mutation must NOT silently admit tenant_admin
		// via DefaultRoles — catches accidental re-widening of the mutation
		// routes fixed in #401. (Invitation routes are a separate concern.)
		if route.Roles == nil && strings.Contains(route.Path, "/users/:id") && route.Method != "GET" {
			t.Errorf("%s %s uses DefaultRoles; id-scoped user mutations in users_admin must set explicit instance-level roles", route.Method, route.Path)
		}
	}

	// Sanity: the two mutation routes must actually exist (guards against the
	// paths being renamed and the assertion above silently matching nothing).
	joined := ""
	for _, route := range group.Routes {
		joined += route.Method + " " + route.Path + ";"
	}
	require.Contains(t, joined, "PATCH /users/:id;")
	require.Contains(t, joined, "POST /users/:id/reset-password;")
}
