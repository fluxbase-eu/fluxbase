package routes

import (
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// branchNoop is a placeholder handler for route construction in tests.
func branchNoop(c fiber.Ctx) error { return c.SendString("ok") }

// TestBranchActiveRoutesRequireInstanceAdmin verifies that the routes managing
// the instance-global active branch are restricted to instance-level admins
// and are not reachable by tenant admins (SEC-09, #375).
func TestBranchActiveRoutesRequireInstanceAdmin(t *testing.T) {
	deps := &BranchDeps{
		GetActiveBranch:    branchNoop,
		SetActiveBranch:    branchNoop,
		ResetActiveBranch:  branchNoop,
		GetPoolStats:       branchNoop,
		CreateBranch:       branchNoop,
		ListBranches:       branchNoop,
		GetBranch:          branchNoop,
		DeleteBranch:       branchNoop,
		ResetBranch:        branchNoop,
		GetBranchActivity:  branchNoop,
		ListBranchAccess:   branchNoop,
		GrantBranchAccess:  branchNoop,
		RevokeBranchAccess: branchNoop,
		ListGitHubConfigs:  branchNoop,
		UpsertGitHubConfig: branchNoop,
		DeleteGitHubConfig: branchNoop,
	}

	group := BuildBranchRoutes(deps)
	require.NotNil(t, group)

	instanceAdminOnly := []string{"admin", "instance_admin"}

	found := 0
	for _, route := range group.Routes {
		if !strings.Contains(route.Path, "/branches/active") {
			continue
		}
		found++
		assert.Equal(t, instanceAdminOnly, route.Roles,
			"%s %s must be restricted to instance-level admins", route.Method, route.Path)
		assert.NotContains(t, route.Roles, "tenant_admin",
			"%s %s must not admit tenant admins", route.Method, route.Path)
	}
	assert.Equal(t, 3, found, "expected the three /branches/active routes")
}
