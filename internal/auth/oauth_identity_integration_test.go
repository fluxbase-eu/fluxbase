//go:build integration
// +build integration

package auth_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/auth"
	"github.com/nimbleflux/fluxbase/internal/testutil/e2e"
)

// newIdentityService builds an IdentityService around the shared e2e
// database, mirroring the production wiring in auth.NewService. The
// repository is returned alongside for direct identity setup/inspection.
func newIdentityService(t *testing.T, tc *e2e.IntegrationTestContext) (*auth.IdentityService, *auth.IdentityRepository) {
	t.Helper()
	repo := auth.NewIdentityRepository(tc.TestContext.DB)
	return auth.NewIdentityService(repo, auth.NewOAuthManager(), auth.NewStateStore()), repo
}

// signupIdentityUser registers a fresh password user for identity tests.
func signupIdentityUser(t *testing.T, tc *e2e.IntegrationTestContext) (context.Context, string) {
	t.Helper()
	ctx := context.Background()
	email := randomTestEmail()
	resp, err := newAuthServiceForIdentity(t, tc).SignUp(ctx, auth.SignUpRequest{
		Email:    email,
		Password: "TestPassword123!",
	})
	require.NoError(t, err)
	return ctx, resp.User.ID
}

// newAuthServiceForIdentity returns a minimal auth service (no MFA key needed).
func newAuthServiceForIdentity(t *testing.T, tc *e2e.IntegrationTestContext) *auth.Service {
	t.Helper()
	return createAuthService(t, tc)
}

// createOAuthIdentity links a provider identity to a user through the
// repository and returns it.
func createOAuthIdentity(t *testing.T, repo *auth.IdentityRepository, ctx context.Context, userID, provider, providerUserID string) *auth.UserIdentity {
	t.Helper()
	email := userID + "@identity.test.local"
	identity, err := repo.Create(ctx, userID, provider, providerUserID, &email, map[string]interface{}{
		"provider": provider,
	})
	require.NoError(t, err)
	return identity
}

// TestOAuth_UnlinkIdentity_Success_Integration covers the happy path: an
// identity belonging to the user is deleted, other identities survive.
func TestOAuth_UnlinkIdentity_Success_Integration(t *testing.T) {
	tc := e2e.NewIntegrationTestContextWithNamespace(t, "auth")
	defer tc.CleanupTestData()

	ctx, userID := signupIdentityUser(t, tc)
	service, repo := newIdentityService(t, tc)

	first := createOAuthIdentity(t, repo, ctx, userID, "github", "gh-123")
	second := createOAuthIdentity(t, repo, ctx, userID, "google", "gg-456")

	require.NoError(t, service.UnlinkIdentity(ctx, userID, first.ID), "unlinking an owned identity should succeed")

	remaining, err := repo.GetByUserID(ctx, userID)
	require.NoError(t, err)
	require.Len(t, remaining, 1, "exactly one identity should remain")
	assert.Equal(t, second.ID, remaining[0].ID)
	assert.Equal(t, "google", remaining[0].Provider)
}

// TestOAuth_UnlinkIdentity_NotFound_Integration covers unlinking an identity
// ID that does not exist.
func TestOAuth_UnlinkIdentity_NotFound_Integration(t *testing.T) {
	tc := e2e.NewIntegrationTestContextWithNamespace(t, "auth")
	defer tc.CleanupTestData()

	ctx, userID := signupIdentityUser(t, tc)
	service, _ := newIdentityService(t, tc)

	err := service.UnlinkIdentity(ctx, userID, uuid.New().String())
	require.Error(t, err)
	assert.ErrorIs(t, err, auth.ErrIdentityNotFound)
}

// TestOAuth_UnlinkIdentity_OwnershipCheck_Integration ensures a user cannot
// unlink another user's identity: the delete is scoped to (id, user_id), so
// the request reports not-found and the other user's link survives.
func TestOAuth_UnlinkIdentity_OwnershipCheck_Integration(t *testing.T) {
	tc := e2e.NewIntegrationTestContextWithNamespace(t, "auth")
	defer tc.CleanupTestData()

	ctx, userA := signupIdentityUser(t, tc)
	_, userB := signupIdentityUser(t, tc)
	service, repo := newIdentityService(t, tc)

	identityB := createOAuthIdentity(t, repo, ctx, userB, "github", "gh-B")

	err := service.UnlinkIdentity(ctx, userA, identityB.ID)
	require.Error(t, err, "unlinking another user's identity must fail")
	assert.ErrorIs(t, err, auth.ErrIdentityNotFound)

	remaining, err := repo.GetByUserID(ctx, userB)
	require.NoError(t, err)
	require.Len(t, remaining, 1, "user B's identity must survive user A's unlink attempt")
	assert.Equal(t, identityB.ID, remaining[0].ID)
}

// TestOAuth_UnlinkIdentity_LastIdentity_Integration pins the behavior for
// removing a user's ONLY identity.
//
// NOTE: issue #313 describes a "last identity" rule (refusing to unlink when
// nothing would remain). No such rule exists in the current implementation —
// UnlinkIdentity performs a plain (id, user_id)-scoped delete — so this test
// pins the actual contract: unlinking the last identity succeeds and leaves
// the user with zero identities.
func TestOAuth_UnlinkIdentity_LastIdentity_Integration(t *testing.T) {
	tc := e2e.NewIntegrationTestContextWithNamespace(t, "auth")
	defer tc.CleanupTestData()

	ctx, userID := signupIdentityUser(t, tc)
	service, repo := newIdentityService(t, tc)

	only := createOAuthIdentity(t, repo, ctx, userID, "github", "gh-only")

	require.NoError(t, service.UnlinkIdentity(ctx, userID, only.ID),
		"current implementation allows unlinking the last identity")

	remaining, err := repo.GetByUserID(ctx, userID)
	require.NoError(t, err)
	assert.Empty(t, remaining, "no identities should remain after unlinking the only one")
}
