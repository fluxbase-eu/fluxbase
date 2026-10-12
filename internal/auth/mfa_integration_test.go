//go:build integration
// +build integration

package auth_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/auth"
	"github.com/nimbleflux/fluxbase/internal/crypto"
	"github.com/nimbleflux/fluxbase/internal/testutil/e2e"
)

// testEncryptionKey is the AES-256 key (exactly 32 bytes) MFA tests configure
// on the service. TOTP secrets are only stored encrypted when a key is set —
// mirroring production, where FLUXBASE_ENCRYPTION_KEY feeds SetEncryptionKey.
var testEncryptionKey = []byte("0123456789abcdef0123456789abcdef")

// newMFAService returns the shared auth service wired with an encryption key.
func newMFAService(t *testing.T, tc *e2e.IntegrationTestContext) *auth.Service {
	t.Helper()
	service := createAuthService(t, tc)
	service.SetEncryptionKey(testEncryptionKey)
	return service
}

// signupMFAUser registers a fresh user and returns the service, user ID and
// password for MFA flow tests.
func signupMFAUser(t *testing.T, service *auth.Service) (context.Context, string, string) {
	t.Helper()
	ctx := context.Background()
	email := randomTestEmail()
	password := "TestPassword123!"
	resp, err := service.SignUp(ctx, auth.SignUpRequest{Email: email, Password: password})
	require.NoError(t, err)
	require.NotEmpty(t, resp.User.ID)
	return ctx, resp.User.ID, password
}

// enrollTOTP runs Setup + Enable with a fresh valid code and returns the
// cleartext setup secret plus the plaintext backup codes.
func enrollTOTP(t *testing.T, ctx context.Context, service *auth.Service, userID string) (string, []string) {
	t.Helper()
	setup, err := service.MFAService().SetupTOTP(ctx, userID, "Fluxbase")
	require.NoError(t, err)
	require.NotEmpty(t, setup.TOTP.Secret)

	code, err := totp.GenerateCode(setup.TOTP.Secret, time.Now())
	require.NoError(t, err)

	backupCodes, err := service.MFAService().EnableTOTP(ctx, userID, code)
	require.NoError(t, err)
	return setup.TOTP.Secret, backupCodes
}

// TestMFA_EnableTOTP_Integration covers the DB-coupled enrollment path:
// the secret is encrypted at rest, backup code hashes are stored, the flag is
// set, and the setup row is marked verified.
func TestMFA_EnableTOTP_Integration(t *testing.T) {
	tc := e2e.NewIntegrationTestContextWithNamespace(t, "auth")
	defer tc.Close()
	defer tc.CleanupTestData()

	service := newMFAService(t, tc)
	ctx, userID, _ := signupMFAUser(t, service)

	enabled, err := service.MFAService().IsTOTPEnabled(ctx, userID)
	require.NoError(t, err)
	assert.False(t, enabled, "TOTP must not be enabled before enrollment")

	secret, backupCodes := enrollTOTP(t, ctx, service, userID)

	// IsTOTPEnabled now reports true.
	enabled, err = service.MFAService().IsTOTPEnabled(ctx, userID)
	require.NoError(t, err)
	assert.True(t, enabled, "TOTP should be enabled after enrollment")

	// The stored secret is ciphertext that decrypts back to the setup secret.
	var storedSecret string
	var backupHashes []string
	err = tc.DB.Pool().QueryRow(ctx,
		`SELECT totp_secret, backup_codes FROM auth.users WHERE id = $1`, userID).
		Scan(&storedSecret, &backupHashes)
	require.NoError(t, err)
	assert.NotEqual(t, secret, storedSecret, "TOTP secret must be stored encrypted, not in cleartext")
	decrypted, err := crypto.DecryptWithBytesKey(storedSecret, testEncryptionKey)
	require.NoError(t, err)
	assert.Equal(t, secret, decrypted)

	// Ten bcrypt-hashed backup codes were persisted (not the plaintext ones).
	assert.Len(t, backupCodes, 10)
	assert.Len(t, backupHashes, 10)
	assert.NotContains(t, backupHashes, backupCodes[0], "backup codes must be stored hashed")

	// The enrollment attempt was marked verified.
	var verified bool
	err = tc.DB.Pool().QueryRow(ctx,
		`SELECT verified FROM auth.two_factor_setups WHERE user_id = $1`, userID).
		Scan(&verified)
	require.NoError(t, err)
	assert.True(t, verified, "setup row should be marked verified after enabling")
}

// TestMFA_EnableTOTP_InvalidCode_Integration ensures enrollment is rejected
// when the confirmation code does not match, leaving the user unenrolled.
func TestMFA_EnableTOTP_InvalidCode_Integration(t *testing.T) {
	tc := e2e.NewIntegrationTestContextWithNamespace(t, "auth")
	defer tc.Close()
	defer tc.CleanupTestData()

	service := newMFAService(t, tc)
	ctx, userID, _ := signupMFAUser(t, service)

	_, err := service.MFAService().SetupTOTP(ctx, userID, "Fluxbase")
	require.NoError(t, err)

	_, err = service.MFAService().EnableTOTP(ctx, userID, "000000")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid TOTP code")

	enabled, err := service.MFAService().IsTOTPEnabled(ctx, userID)
	require.NoError(t, err)
	assert.False(t, enabled, "a wrong code must not enable TOTP")

	// The failed enrollment must keep the setup row unverified so a retry
	// (SetupTOTP upsert resets verified anyway) can succeed.
	var verified bool
	err = tc.DB.Pool().QueryRow(ctx,
		`SELECT verified FROM auth.two_factor_setups WHERE user_id = $1`, userID).
		Scan(&verified)
	require.NoError(t, err)
	assert.False(t, verified)
}

// TestMFA_EnableTOTP_WithoutEncryptionKey_Integration pins the guard against
// storing TOTP secrets in plaintext: without a configured key, enabling fails.
func TestMFA_EnableTOTP_WithoutEncryptionKey_Integration(t *testing.T) {
	tc := e2e.NewIntegrationTestContextWithNamespace(t, "auth")
	defer tc.Close()
	defer tc.CleanupTestData()

	// createAuthService intentionally does NOT set an encryption key here.
	service := createAuthService(t, tc)
	ctx, userID, _ := signupMFAUser(t, service)

	setup, err := service.MFAService().SetupTOTP(ctx, userID, "Fluxbase")
	require.NoError(t, err)

	_, err = service.MFAService().EnableTOTP(ctx, userID, func() string {
		code, genErr := totp.GenerateCode(setup.TOTP.Secret, time.Now())
		require.NoError(t, genErr)
		return code
	}())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "encryption key not configured")

	enabled, err := service.MFAService().IsTOTPEnabled(ctx, userID)
	require.NoError(t, err)
	assert.False(t, enabled)
}

// TestMFA_VerifyTOTP_ValidCode_Integration covers VerifyTOTPWithContext's
// happy path and its replay-protection bookkeeping (totp_last_used_step).
func TestMFA_VerifyTOTP_ValidCode_Integration(t *testing.T) {
	tc := e2e.NewIntegrationTestContextWithNamespace(t, "auth")
	defer tc.Close()
	defer tc.CleanupTestData()

	service := newMFAService(t, tc)
	ctx, userID, _ := signupMFAUser(t, service)
	secret, _ := enrollTOTP(t, ctx, service, userID)

	code, err := totp.GenerateCode(secret, time.Now())
	require.NoError(t, err)
	require.NoError(t, service.MFAService().VerifyTOTPWithContext(ctx, userID, code, "127.0.0.1", "itest"))

	var lastUsedStep int64
	err = tc.DB.Pool().QueryRow(ctx,
		`SELECT COALESCE(totp_last_used_step, 0) FROM auth.users WHERE id = $1`, userID).
		Scan(&lastUsedStep)
	require.NoError(t, err)
	assert.Greater(t, lastUsedStep, int64(0), "the consumed TOTP step must be recorded")
}

// TestMFA_VerifyTOTP_ReplayRejected_Integration ensures a code cannot be used
// twice: the second presentation of the same step is rejected.
func TestMFA_VerifyTOTP_ReplayRejected_Integration(t *testing.T) {
	tc := e2e.NewIntegrationTestContextWithNamespace(t, "auth")
	defer tc.Close()
	defer tc.CleanupTestData()

	service := newMFAService(t, tc)
	ctx, userID, _ := signupMFAUser(t, service)
	secret, _ := enrollTOTP(t, ctx, service, userID)

	code, err := totp.GenerateCode(secret, time.Now())
	require.NoError(t, err)
	require.NoError(t, service.MFAService().VerifyTOTPWithContext(ctx, userID, code, "", ""))

	err = service.MFAService().VerifyTOTPWithContext(ctx, userID, code, "", "")
	require.Error(t, err, "replaying the same TOTP code must fail")
	assert.Contains(t, err.Error(), "already been used")
}

// TestMFA_VerifyTOTP_WrongCode_Integration covers the failure path for a
// well-formed but incorrect code.
func TestMFA_VerifyTOTP_WrongCode_Integration(t *testing.T) {
	tc := e2e.NewIntegrationTestContextWithNamespace(t, "auth")
	defer tc.Close()
	defer tc.CleanupTestData()

	service := newMFAService(t, tc)
	ctx, userID, _ := signupMFAUser(t, service)
	enrollTOTP(t, ctx, service, userID)

	err := service.MFAService().VerifyTOTPWithContext(ctx, userID, "000000", "", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid 2FA code")
}

// TestMFA_VerifyTOTP_BackupCode_Integration covers backup-code recovery:
// a backup code verifies once, is consumed from the stored list, and is
// rejected on reuse.
func TestMFA_VerifyTOTP_BackupCode_Integration(t *testing.T) {
	tc := e2e.NewIntegrationTestContextWithNamespace(t, "auth")
	defer tc.Close()
	defer tc.CleanupTestData()

	service := newMFAService(t, tc)
	ctx, userID, _ := signupMFAUser(t, service)
	_, backupCodes := enrollTOTP(t, ctx, service, userID)
	require.NotEmpty(t, backupCodes)

	backupCode := backupCodes[0]
	require.NoError(t, service.MFAService().VerifyTOTPWithContext(ctx, userID, backupCode, "", ""),
		"an unused backup code should verify")

	// The consumed code was removed from the stored (hashed) list.
	var remaining []string
	err := tc.DB.Pool().QueryRow(ctx,
		`SELECT backup_codes FROM auth.users WHERE id = $1`, userID).
		Scan(&remaining)
	require.NoError(t, err)
	assert.Len(t, remaining, len(backupCodes)-1, "consumed backup code should be removed")

	err = service.MFAService().VerifyTOTPWithContext(ctx, userID, backupCode, "", "")
	require.Error(t, err, "a consumed backup code must be rejected")
}

// TestMFA_VerifyTOTP_NotEnabled_Integration ensures verification fails for a
// user who never enrolled in TOTP.
func TestMFA_VerifyTOTP_NotEnabled_Integration(t *testing.T) {
	tc := e2e.NewIntegrationTestContextWithNamespace(t, "auth")
	defer tc.Close()
	defer tc.CleanupTestData()

	service := newMFAService(t, tc)
	ctx, userID, _ := signupMFAUser(t, service)

	err := service.MFAService().VerifyTOTPWithContext(ctx, userID, "123456", "", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "2FA not enabled")
}

// TestMFA_DisableTOTP_Integration covers the password re-verification gate on
// disable, and the full cleanup of the stored factor.
func TestMFA_DisableTOTP_Integration(t *testing.T) {
	tc := e2e.NewIntegrationTestContextWithNamespace(t, "auth")
	defer tc.Close()
	defer tc.CleanupTestData()

	service := newMFAService(t, tc)
	ctx, userID, password := signupMFAUser(t, service)
	enrollTOTP(t, ctx, service, userID)

	// Wrong password is rejected and the factor survives.
	err := service.MFAService().DisableTOTP(ctx, userID, "WrongPassword!")
	require.Error(t, err, "disabling with the wrong password must fail")
	assert.Contains(t, err.Error(), "invalid password")
	enabled, err := service.MFAService().IsTOTPEnabled(ctx, userID)
	require.NoError(t, err)
	assert.True(t, enabled, "TOTP must stay enabled after a failed disable")

	// Correct password disables and clears the stored material.
	require.NoError(t, service.MFAService().DisableTOTP(ctx, userID, password))

	enabled, err = service.MFAService().IsTOTPEnabled(ctx, userID)
	require.NoError(t, err)
	assert.False(t, enabled, "TOTP should be disabled after DisableTOTP")

	var storedSecret *string
	var backupHashes *[]string
	var lastUsedStep int64
	err = tc.DB.Pool().QueryRow(ctx,
		`SELECT totp_secret, backup_codes, COALESCE(totp_last_used_step, 0) FROM auth.users WHERE id = $1`, userID).
		Scan(&storedSecret, &backupHashes, &lastUsedStep)
	require.NoError(t, err)
	assert.Nil(t, storedSecret, "totp_secret must be cleared on disable")
	assert.Nil(t, backupHashes, "backup_codes must be cleared on disable")
	assert.Zero(t, lastUsedStep, "totp_last_used_step must be reset on disable")

	// The setup row is removed as part of disabling.
	var count int
	err = tc.DB.Pool().QueryRow(ctx,
		`SELECT COUNT(*) FROM auth.two_factor_setups WHERE user_id = $1`, userID).
		Scan(&count)
	require.NoError(t, err)
	assert.Zero(t, count, "two_factor_setups row should be deleted on disable")
}

// TestMFA_GenerateTokensForUser_Integration covers session creation outside a
// normal sign-in (used e.g. after MFA challenge completion): tokens must be
// valid JWTs backed by a stored session row.
func TestMFA_GenerateTokensForUser_Integration(t *testing.T) {
	tc := e2e.NewIntegrationTestContextWithNamespace(t, "auth")
	defer tc.Close()
	defer tc.CleanupTestData()

	service := newMFAService(t, tc)
	ctx, userID, _ := signupMFAUser(t, service)

	// Sessions store token hashes, so verify session creation by count: the
	// signup flow already created one session for this user.
	var sessionsBefore int
	err := tc.DB.Pool().QueryRow(ctx,
		`SELECT COUNT(*) FROM auth.sessions WHERE user_id = $1`, userID).Scan(&sessionsBefore)
	require.NoError(t, err)

	resp, err := service.MFAService().GenerateTokensForUser(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, userID, resp.User.ID)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Greater(t, resp.ExpiresIn, int64(0))

	// The access token validates and carries the user identity.
	claims, err := service.JWTManager().ValidateToken(resp.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)

	var sessionsAfter int
	err = tc.DB.Pool().QueryRow(ctx,
		`SELECT COUNT(*) FROM auth.sessions WHERE user_id = $1`, userID).Scan(&sessionsAfter)
	require.NoError(t, err)
	assert.Equal(t, sessionsBefore+1, sessionsAfter, "one session row should be added per GenerateTokensForUser call")
}

// TestMFA_GenerateTokensForUser_UnknownUser_Integration pins the error path.
func TestMFA_GenerateTokensForUser_UnknownUser_Integration(t *testing.T) {
	tc := e2e.NewIntegrationTestContextWithNamespace(t, "auth")
	defer tc.Close()
	defer tc.CleanupTestData()

	service := newMFAService(t, tc)
	ctx := context.Background()

	resp, err := service.MFAService().GenerateTokensForUser(ctx, "00000000-0000-0000-0000-000000000000")
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, strings.Contains(err.Error(), "user not found"))
}
