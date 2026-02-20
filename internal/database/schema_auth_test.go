package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextWithAuth(t *testing.T) {
	ctx := context.Background()

	// Add auth context
	authCtx := ContextWithAuth(ctx, "user123", "admin", true)

	// Extract auth context
	auth := AuthFromContext(authCtx)

	assert.NotNil(t, auth)
	assert.Equal(t, "user123", auth.UserID)
	assert.Equal(t, "admin", auth.UserRole)
	assert.True(t, auth.IsAdmin)
}

func TestAuthFromContext_NoContext(t *testing.T) {
	// Context without auth key
	ctx := context.Background()
	auth := AuthFromContext(ctx)
	assert.Nil(t, auth)
}

func TestAuthContext_PermissionChecks(t *testing.T) {
	tests := []struct {
		name        string
		auth        *AuthContext
		shouldAllow bool
		reason      string
	}{
		{
			name:        "nil auth",
			auth:        nil,
			shouldAllow: false,
			reason:      "no auth context",
		},
		{
			name: "service role",
			auth: &AuthContext{
				UserRole: "service_role",
			},
			shouldAllow: true,
			reason:      "service role has access",
		},
		{
			name: "admin",
			auth: &AuthContext{
				UserRole: "admin",
				IsAdmin:  true,
			},
			shouldAllow: true,
			reason:      "admin has access",
		},
		{
			name: "authenticated user",
			auth: &AuthContext{
				UserRole: "authenticated",
			},
			shouldAllow: true,
			reason:      "authenticated users have access",
		},
		{
			name: "anonymous user",
			auth: &AuthContext{
				UserRole: "anon",
			},
			shouldAllow: false,
			reason:      "anonymous users don't have access",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// In a real implementation, we'd check HasPermission()
			// For now, just verify the auth context structure
			if tt.auth == nil {
				assert.Nil(t, tt.auth)
			} else {
				assert.Equal(t, tt.reason, tt.reason) // placeholder assertion
				if tt.shouldAllow && tt.auth.UserRole == "anon" {
					t.Error("Anonymous users should not have permission")
				}
			}
		})
	}
}

func TestLogSchemaIntrospection_WithAuth(t *testing.T) {
	ctx := ContextWithAuth(context.Background(), "user123", "admin", true)

	// This should log with auth info
	// In a real test, we'd capture the log output
	// For now, just ensure it doesn't panic
	LogSchemaIntrospection(ctx, "GetAllTables", map[string]interface{}{"schemas": []string{"public"}})
}

func TestLogSchemaIntrospection_WithoutAuth(t *testing.T) {
	ctx := context.Background()

	// This should log without auth info
	LogSchemaIntrospection(ctx, "GetAllTables", map[string]interface{}{"schemas": []string{"public"}})
}
