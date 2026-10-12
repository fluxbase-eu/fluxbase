package secrets

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/crypto"
)

var updateParamRe = regexp.MustCompile(`\$([0-9]+)`)

// maxParam returns the highest $N placeholder used in a query fragment.
func maxParam(query string) int {
	max := 0
	for _, m := range updateParamRe.FindAllStringSubmatch(query, -1) {
		n := 0
		for _, r := range m[1] {
			n = n*10 + int(r-'0')
		}
		if n > max {
			max = n
		}
	}
	return max
}

func TestBuildUpdateQuery_NoOptionalFields(t *testing.T) {
	secretID := uuid.New()
	userID := uuid.New()

	query, args := buildUpdateQuery(secretID, &userID, nil, nil, nil, "")

	assert.Contains(t, query, "SET updated_at = NOW(), updated_by = $2")
	assert.NotContains(t, query, "encrypted_value = $", "no value binding in SET clause")
	assert.NotContains(t, query, "version = version + 1")
	assert.Contains(t, query, "WHERE id = $1")
	assert.Contains(t, query, "RETURNING version, encrypted_value")

	// $1 = secret ID, $2 = user, tenant filter binds $3 (twice) and is last.
	require.Len(t, args, 3)
	assert.Equal(t, secretID, args[0])
	assert.Equal(t, &userID, args[1])
	assert.Nil(t, args[2], "empty tenant binds NULL")
	assert.Equal(t, 3, maxParam(query), "tenant filter should use the next free parameter index")
}

func TestBuildUpdateQuery_ValueOnly(t *testing.T) {
	secretID := uuid.New()
	userID := uuid.New()
	encrypted := "c2FsdGVk"

	query, args := buildUpdateQuery(secretID, &userID, &encrypted, nil, nil, "")

	assert.Contains(t, query, "encrypted_value = $3, version = version + 1")
	assert.NotContains(t, query, "description = $")
	assert.NotContains(t, query, "expires_at = $")
	require.Len(t, args, 4)
	assert.Equal(t, encrypted, args[2].(string))
	assert.Nil(t, args[3])
}

func TestBuildUpdateQuery_DescriptionOnly(t *testing.T) {
	secretID := uuid.New()
	userID := uuid.New()
	description := "rotated credential"

	query, args := buildUpdateQuery(secretID, &userID, nil, &description, nil, "")

	assert.Contains(t, query, "description = $3")
	assert.NotContains(t, query, "encrypted_value = $", "no value binding in SET clause")
	assert.NotContains(t, query, "version = version + 1", "version must not bump without a value change")
	require.Len(t, args, 4)
	assert.Equal(t, description, args[2].(string))
}

func TestBuildUpdateQuery_AllFields(t *testing.T) {
	secretID := uuid.New()
	userID := uuid.New()
	encrypted := "c2FsdGVk"
	description := "rotated credential"
	expiresAt := time.Now().Add(24 * time.Hour)

	query, args := buildUpdateQuery(secretID, &userID, &encrypted, &description, &expiresAt, "tenant-1")

	assert.Contains(t, query, "encrypted_value = $3, version = version + 1")
	assert.Contains(t, query, "description = $4")
	assert.Contains(t, query, "expires_at = $5")
	// The tenant filter uses the next free index ($6) in both comparison arms.
	assert.Contains(t, query, "tenant_id = $6 OR ($6 IS NULL AND tenant_id IS NULL)")
	require.Len(t, args, 6)
	assert.Equal(t, encrypted, args[2].(string))
	assert.Equal(t, description, args[3].(string))
	assert.Equal(t, expiresAt, args[4].(time.Time))
	assert.Equal(t, "tenant-1", args[5].(string))
}

func TestBuildUpdateQuery_TenantBindingAlwaysLast(t *testing.T) {
	secretID := uuid.New()
	userID := uuid.New()

	// The tenant binding must come after every optional field binding,
	// regardless of which combination is set.
	for _, tc := range []struct {
		name           string
		value, descr   *string
		expires        *time.Time
		wantTenantIdx  int
		wantArgsLength int
	}{
		{"none", nil, nil, nil, 3, 3},
		{"value only", strPtr("x"), nil, nil, 4, 4},
		{"description only", nil, strPtr("d"), nil, 4, 4},
		{"expiry only", nil, nil, timePtr(time.Now()), 4, 4},
		{"value+description", strPtr("x"), strPtr("d"), nil, 5, 5},
		{"value+description+expiry", strPtr("x"), strPtr("d"), timePtr(time.Now()), 6, 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			query, args := buildUpdateQuery(secretID, &userID, tc.value, tc.descr, tc.expires, "tenant-9")
			assert.Equal(t, "tenant-9", args[tc.wantTenantIdx-1].(string), "tenant binding must be last")
			assert.Len(t, args, tc.wantArgsLength)
			assert.Equal(t, tc.wantTenantIdx, maxParam(query))
			assert.True(t, strings.Contains(query, "WHERE id = $1"))
		})
	}
}

// TestBuildUpdateQuery_EncryptedValueRoundTrip verifies that the value bound
// into the query decrypts back to the plaintext with the same AES-256 key the
// storage uses — the round-trip contract UpdateSecret relies on.
func TestBuildUpdateQuery_EncryptedValueRoundTrip(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	// Use a string that is exactly 32 bytes.
	require.Len(t, key, 32)

	plain := "super-secret-value"
	encrypted, err := crypto.EncryptWithBytesKey(plain, key)
	require.NoError(t, err)

	secretID := uuid.New()
	userID := uuid.New()
	query, args := buildUpdateQuery(secretID, &userID, &encrypted, nil, nil, "")

	assert.Contains(t, query, "encrypted_value = $3")
	boundValue := args[2].(string)
	decrypted, err := crypto.DecryptWithBytesKey(boundValue, key)
	require.NoError(t, err)
	assert.Equal(t, plain, decrypted)
}

// TestUpdateEncryption_WrongKeyFails verifies that a value encrypted with one
// storage key cannot be decrypted with another — the failure mode that keeps
// encrypted secrets useless to anyone without FLUXBASE_ENCRYPTION_KEY.
func TestUpdateEncryption_WrongKeyFails(t *testing.T) {
	keyA := []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	keyB := []byte("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	require.Len(t, keyA, 32)
	require.Len(t, keyB, 32)

	encrypted, err := crypto.EncryptWithBytesKey("secret", keyA)
	require.NoError(t, err)

	_, err = crypto.DecryptWithBytesKey(encrypted, keyB)
	assert.Error(t, err, "decrypting with the wrong key must fail")

	decrypted, err := crypto.DecryptWithBytesKey(encrypted, keyA)
	require.NoError(t, err)
	assert.Equal(t, "secret", decrypted)
}
