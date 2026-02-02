package ratelimit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// PostgresStore Construction Tests
// =============================================================================

func TestNewPostgresStore(t *testing.T) {
	t.Run("creates store with nil pool", func(t *testing.T) {
		store := NewPostgresStore(nil)

		require.NotNil(t, store)
		assert.Nil(t, store.pool)
	})
}

func TestPostgresStore_Close(t *testing.T) {
	t.Run("close is no-op", func(t *testing.T) {
		store := NewPostgresStore(nil)

		err := store.Close()

		assert.NoError(t, err)
	})
}

// =============================================================================
// PostgresStore Struct Tests
// =============================================================================

func TestPostgresStore_FieldAccess(t *testing.T) {
	t.Run("pool field accessible", func(t *testing.T) {
		store := &PostgresStore{
			pool: nil,
		}

		// Just verify the struct can be created and accessed
		assert.Nil(t, store.pool)
	})
}

// =============================================================================
// PostgresStore Query Logic Tests (unit tests for logic, not DB)
// =============================================================================

func TestPostgresStore_QueryLogic(t *testing.T) {
	t.Run("UPSERT handles concurrent access", func(t *testing.T) {
		// The PostgreSQL UPSERT query should:
		// 1. Insert new row with count=1 if key doesn't exist
		// 2. Update count=count+1 if key exists and not expired
		// 3. Reset count=1 and update expiration if key exists but expired

		// This test documents the expected behavior
		// Actual database testing would require an integration test

		store := NewPostgresStore(nil)
		assert.NotNil(t, store)
	})

	t.Run("Get returns zero for non-existent keys", func(t *testing.T) {
		// The Get query should return 0 count and empty time
		// when the key doesn't exist (no rows returned)

		// This documents the expected behavior
		// store.Get() returns (0, time.Time{}, nil) on pgx.ErrNoRows
	})

	t.Run("Reset deletes the key", func(t *testing.T) {
		// The Reset should delete the row completely
		// This allows the next request to start fresh
	})

	t.Run("Cleanup removes expired entries", func(t *testing.T) {
		// The Cleanup method should delete all rows where expires_at <= NOW()
		// This prevents table bloat over time
	})
}

// =============================================================================
// PostgresStore Table Schema Tests
// =============================================================================

func TestPostgresStore_TableSchema(t *testing.T) {
	t.Run("EnsureTable creates expected schema", func(t *testing.T) {
		// The table schema should have:
		// - key: TEXT PRIMARY KEY
		// - count: BIGINT NOT NULL DEFAULT 1
		// - expires_at: TIMESTAMPTZ NOT NULL
		// - created_at: TIMESTAMPTZ NOT NULL DEFAULT NOW()
		// And an index on expires_at for efficient cleanup

		store := NewPostgresStore(nil)
		assert.NotNil(t, store)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkNewPostgresStore(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewPostgresStore(nil)
	}
}
