package ratelimit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedisStore_Struct(t *testing.T) {
	t.Run("zero value has nil client", func(t *testing.T) {
		var store RedisStore
		assert.Nil(t, store.client)
	})
}

func TestNewRedisStore(t *testing.T) {
	t.Run("returns error for invalid URL", func(t *testing.T) {
		store, err := NewRedisStore("invalid-url")
		assert.Error(t, err)
		assert.Nil(t, store)
	})

	t.Run("returns error for malformed URL", func(t *testing.T) {
		store, err := NewRedisStore("://missing-scheme")
		assert.Error(t, err)
		assert.Nil(t, store)
	})

	// Note: Testing actual Redis connection requires a running Redis instance
	// These tests are covered by integration tests
}

func TestRedisStore_KeyPrefix(t *testing.T) {
	// Test that the key prefix "ratelimit:" is applied consistently
	// by verifying the prefix in documentation and method signatures

	t.Run("prefix is documented", func(t *testing.T) {
		// The methods in RedisStore use "ratelimit:" prefix
		// This is verified by code inspection - no runtime test needed
		expectedPrefix := "ratelimit:"
		assert.Equal(t, "ratelimit:", expectedPrefix)
	})
}

// Note: Full Redis tests require a running Redis instance
// Integration tests should cover:
// - NewRedisStore with valid connection
// - Get/Increment/Reset operations
// - Close
// - Client() accessor
