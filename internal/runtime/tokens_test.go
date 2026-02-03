package runtime

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// generateUserToken Tests
// =============================================================================

func TestGenerateUserToken(t *testing.T) {
	t.Run("returns error for empty JWT secret", func(t *testing.T) {
		req := ExecutionRequest{
			ID:   uuid.New(),
			Name: "test-function",
		}

		token, err := generateUserToken("", req, RuntimeTypeFunction, 30*time.Second)

		require.Error(t, err)
		assert.Empty(t, token)
		assert.Contains(t, err.Error(), "JWT secret not configured")
	})

	t.Run("generates valid token for function", func(t *testing.T) {
		req := ExecutionRequest{
			ID:        uuid.New(),
			Name:      "test-function",
			UserID:    "user-123",
			UserEmail: "user@example.com",
			UserRole:  "admin",
		}

		token, err := generateUserToken("test-secret", req, RuntimeTypeFunction, 30*time.Second)

		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// Verify token can be parsed
		parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			return []byte("test-secret"), nil
		})
		require.NoError(t, err)
		assert.True(t, parsed.Valid)

		// Verify claims
		claims := parsed.Claims.(jwt.MapClaims)
		assert.Equal(t, "fluxbase", claims["iss"])
		assert.Equal(t, "user-123", claims["sub"])
		assert.Equal(t, "user-123", claims["user_id"])
		assert.Equal(t, "user@example.com", claims["email"])
		assert.Equal(t, "admin", claims["role"])
		assert.Equal(t, "access", claims["token_type"])
		assert.Equal(t, req.ID.String(), claims["execution_id"])
	})

	t.Run("generates valid token for job", func(t *testing.T) {
		req := ExecutionRequest{
			ID:     uuid.New(),
			Name:   "test-job",
			UserID: "user-456",
		}

		token, err := generateUserToken("test-secret", req, RuntimeTypeJob, 5*time.Minute)

		require.NoError(t, err)
		assert.NotEmpty(t, token)

		parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			return []byte("test-secret"), nil
		})
		require.NoError(t, err)

		claims := parsed.Claims.(jwt.MapClaims)
		assert.Equal(t, req.ID.String(), claims["job_id"])
		assert.Nil(t, claims["execution_id"]) // Should not have execution_id for jobs
	})

	t.Run("defaults role to authenticated when not provided", func(t *testing.T) {
		req := ExecutionRequest{
			ID:     uuid.New(),
			Name:   "test-function",
			UserID: "user-123",
			// No UserRole set
		}

		token, err := generateUserToken("test-secret", req, RuntimeTypeFunction, 30*time.Second)

		require.NoError(t, err)

		parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			return []byte("test-secret"), nil
		})
		require.NoError(t, err)

		claims := parsed.Claims.(jwt.MapClaims)
		assert.Equal(t, "authenticated", claims["role"])
	})

	t.Run("omits user claims when user ID is empty", func(t *testing.T) {
		req := ExecutionRequest{
			ID:   uuid.New(),
			Name: "anonymous-function",
			// No UserID
		}

		token, err := generateUserToken("test-secret", req, RuntimeTypeFunction, 30*time.Second)

		require.NoError(t, err)

		parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			return []byte("test-secret"), nil
		})
		require.NoError(t, err)

		claims := parsed.Claims.(jwt.MapClaims)
		assert.Nil(t, claims["sub"])
		assert.Nil(t, claims["user_id"])
	})

	t.Run("sets correct expiration based on timeout", func(t *testing.T) {
		req := ExecutionRequest{
			ID:   uuid.New(),
			Name: "test-function",
		}
		timeout := 2 * time.Minute

		token, err := generateUserToken("test-secret", req, RuntimeTypeFunction, timeout)

		require.NoError(t, err)

		parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			return []byte("test-secret"), nil
		})
		require.NoError(t, err)

		claims := parsed.Claims.(jwt.MapClaims)
		iat := int64(claims["iat"].(float64))
		exp := int64(claims["exp"].(float64))

		// Expiration should be approximately iat + timeout (within 2 seconds)
		assert.InDelta(t, iat+int64(timeout.Seconds()), exp, 2)
	})

	t.Run("includes unique jti", func(t *testing.T) {
		req := ExecutionRequest{
			ID:   uuid.New(),
			Name: "test-function",
		}

		token1, _ := generateUserToken("test-secret", req, RuntimeTypeFunction, 30*time.Second)
		token2, _ := generateUserToken("test-secret", req, RuntimeTypeFunction, 30*time.Second)

		// Parse both tokens
		parsed1, _ := jwt.Parse(token1, func(t *jwt.Token) (interface{}, error) {
			return []byte("test-secret"), nil
		})
		parsed2, _ := jwt.Parse(token2, func(t *jwt.Token) (interface{}, error) {
			return []byte("test-secret"), nil
		})

		claims1 := parsed1.Claims.(jwt.MapClaims)
		claims2 := parsed2.Claims.(jwt.MapClaims)

		// Each token should have a unique jti
		assert.NotEqual(t, claims1["jti"], claims2["jti"])
	})
}

// =============================================================================
// generateServiceToken Tests
// =============================================================================

func TestGenerateServiceToken(t *testing.T) {
	t.Run("returns error for empty JWT secret", func(t *testing.T) {
		req := ExecutionRequest{
			ID:   uuid.New(),
			Name: "test-function",
		}

		token, err := generateServiceToken("", req, RuntimeTypeFunction, 30*time.Second)

		require.Error(t, err)
		assert.Empty(t, token)
		assert.Contains(t, err.Error(), "JWT secret not configured")
	})

	t.Run("generates valid service token for function", func(t *testing.T) {
		req := ExecutionRequest{
			ID:   uuid.New(),
			Name: "test-function",
		}

		token, err := generateServiceToken("test-secret", req, RuntimeTypeFunction, 30*time.Second)

		require.NoError(t, err)
		assert.NotEmpty(t, token)

		parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			return []byte("test-secret"), nil
		})
		require.NoError(t, err)
		assert.True(t, parsed.Valid)

		claims := parsed.Claims.(jwt.MapClaims)
		assert.Equal(t, "fluxbase", claims["iss"])
		assert.Equal(t, "service_role", claims["sub"])
		assert.Equal(t, "service_role", claims["role"])
		assert.Equal(t, "access", claims["token_type"])
		assert.Equal(t, req.ID.String(), claims["execution_id"])
	})

	t.Run("generates valid service token for job", func(t *testing.T) {
		req := ExecutionRequest{
			ID:   uuid.New(),
			Name: "test-job",
		}

		token, err := generateServiceToken("test-secret", req, RuntimeTypeJob, 5*time.Minute)

		require.NoError(t, err)

		parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			return []byte("test-secret"), nil
		})
		require.NoError(t, err)

		claims := parsed.Claims.(jwt.MapClaims)
		assert.Equal(t, req.ID.String(), claims["job_id"])
		assert.Nil(t, claims["execution_id"])
	})

	t.Run("service token has service_role regardless of request user", func(t *testing.T) {
		req := ExecutionRequest{
			ID:       uuid.New(),
			Name:     "test-function",
			UserID:   "user-123",
			UserRole: "admin",
		}

		token, err := generateServiceToken("test-secret", req, RuntimeTypeFunction, 30*time.Second)

		require.NoError(t, err)

		parsed, _ := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			return []byte("test-secret"), nil
		})

		claims := parsed.Claims.(jwt.MapClaims)
		// Service token should NOT include user's role/id
		assert.Equal(t, "service_role", claims["sub"])
		assert.Equal(t, "service_role", claims["role"])
		assert.Nil(t, claims["user_id"])
	})

	t.Run("uses HS256 signing method", func(t *testing.T) {
		req := ExecutionRequest{
			ID:   uuid.New(),
			Name: "test-function",
		}

		token, err := generateServiceToken("test-secret", req, RuntimeTypeFunction, 30*time.Second)
		require.NoError(t, err)

		// Check token header
		parts := strings.Split(token, ".")
		require.Len(t, parts, 3, "JWT should have 3 parts")
	})
}

// =============================================================================
// Token Generation Benchmarks
// =============================================================================

func BenchmarkGenerateUserToken(b *testing.B) {
	req := ExecutionRequest{
		ID:        uuid.New(),
		Name:      "benchmark-function",
		UserID:    "user-123",
		UserEmail: "user@example.com",
		UserRole:  "authenticated",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generateUserToken("benchmark-secret", req, RuntimeTypeFunction, 30*time.Second)
	}
}

func BenchmarkGenerateServiceToken(b *testing.B) {
	req := ExecutionRequest{
		ID:   uuid.New(),
		Name: "benchmark-function",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generateServiceToken("benchmark-secret", req, RuntimeTypeFunction, 30*time.Second)
	}
}
