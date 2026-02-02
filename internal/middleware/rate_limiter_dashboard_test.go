package middleware

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractRoleFromToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name: "valid token with dashboard_admin role",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
				base64EncodeForJWT([]byte(`{"role":"dashboard_admin","sub":"user-123"}`)) +
				".signature",
			expected: "dashboard_admin",
		},
		{
			name: "valid token with admin role",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
				base64EncodeForJWT([]byte(`{"role":"admin","sub":"user-123"}`)) +
				".signature",
			expected: "admin",
		},
		{
			name: "valid token with service_role",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
				base64EncodeForJWT([]byte(`{"role":"service_role","sub":"user-123"}`)) +
				".signature",
			expected: "service_role",
		},
		{
			name: "valid token with regular user role",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
				base64EncodeForJWT([]byte(`{"role":"authenticated","sub":"user-123"}`)) +
				".signature",
			expected: "authenticated",
		},
		{
			name: "token with no role field",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
				base64EncodeForJWT([]byte(`{"sub":"user-123","email":"user@example.com"}`)) +
				".signature",
			expected: "",
		},
		{
			name:     "invalid token format - missing parts",
			token:    "invalid.token",
			expected: "",
		},
		{
			name:     "invalid token format - empty",
			token:    "",
			expected: "",
		},
		{
			name:     "invalid base64 payload",
			token:    "header.invalid-base64!@#$.signature",
			expected: "",
		},
		{
			name: "valid base64 but invalid JSON",
			token: "header." +
				base64EncodeForJWT([]byte("not json")) +
				".signature",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRoleFromToken(tt.token)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractRoleFromToken_RealWorldJWT(t *testing.T) {
	// Test with a more realistic JWT payload structure
	payload := map[string]interface{}{
		"role":     "dashboard_admin",
		"sub":      "550e8400-e29b-41d4-a716-446655440000",
		"email":    "admin@example.com",
		"iat":      1234567890,
		"exp":      1234567890 + 3600,
		"aud":      "authenticated",
		"iss":      "fluxbase",
	}

	payloadBytes, err := json.Marshal(payload)
	assert.NoError(t, err)

	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
		base64EncodeForJWT(payloadBytes) +
		".SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

	role := extractRoleFromToken(token)
	assert.Equal(t, "dashboard_admin", role)
}

// Helper function to encode payload for JWT (base64url encoding without padding)
func base64EncodeForJWT(data []byte) string {
	// JWT uses base64url encoding (RFC 4648) without padding
	return base64.RawURLEncoding.EncodeToString(data)
}
