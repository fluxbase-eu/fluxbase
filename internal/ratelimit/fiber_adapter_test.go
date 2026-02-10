package ratelimit

import (
	"context"
	"encoding/binary"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Mock Store
// =============================================================================

type mockStore struct {
	count           int64
	expiresAt       time.Time
	incrementBy     int64
	resetCalled     bool
	resetAllCalled  bool
	resetAllPattern string
	closeCalled     bool
}

func (m *mockStore) Get(ctx context.Context, key string) (int64, time.Time, error) {
	return m.count, m.expiresAt, nil
}

func (m *mockStore) Increment(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	m.incrementBy++
	m.count++
	return m.count, nil
}

func (m *mockStore) Reset(ctx context.Context, key string) error {
	m.resetCalled = true
	m.count = 0
	return nil
}

func (m *mockStore) ResetAll(ctx context.Context, pattern string) error {
	m.resetAllCalled = true
	m.resetAllPattern = pattern
	m.count = 0
	return nil
}

func (m *mockStore) Close() error {
	m.closeCalled = true
	return nil
}

// =============================================================================
// FiberAdapter Tests
// =============================================================================

func TestNewFiberAdapter(t *testing.T) {
	t.Run("creates adapter with store", func(t *testing.T) {
		store := &mockStore{}
		adapter := NewFiberAdapter(store)

		require.NotNil(t, adapter)
		assert.Equal(t, store, adapter.store)
	})
}

func TestFiberAdapter_Get(t *testing.T) {
	t.Run("returns encoded count", func(t *testing.T) {
		store := &mockStore{count: 5}
		adapter := NewFiberAdapter(store)

		result, err := adapter.Get("test-key")

		assert.NoError(t, err)
		require.NotNil(t, result)

		// Decode the result
		count := int64(binary.BigEndian.Uint64(result))
		assert.Equal(t, int64(5), count)
	})

	t.Run("returns nil for zero count", func(t *testing.T) {
		store := &mockStore{count: 0}
		adapter := NewFiberAdapter(store)

		result, err := adapter.Get("test-key")

		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestFiberAdapter_Set(t *testing.T) {
	t.Run("set is a no-op", func(t *testing.T) {
		store := &mockStore{}
		adapter := NewFiberAdapter(store)

		// Set should do nothing since our store handles increment differently
		err := adapter.Set("test-key", []byte{1, 2, 3, 4, 5, 6, 7, 8}, time.Minute)

		assert.NoError(t, err)
	})
}

func TestFiberAdapter_Delete(t *testing.T) {
	t.Run("calls reset on store", func(t *testing.T) {
		store := &mockStore{}
		adapter := NewFiberAdapter(store)

		err := adapter.Delete("test-key")

		assert.NoError(t, err)
		assert.True(t, store.resetCalled)
	})
}

func TestFiberAdapter_Reset(t *testing.T) {
	t.Run("reset returns nil for distributed stores", func(t *testing.T) {
		store := &mockStore{}
		adapter := NewFiberAdapter(store)

		err := adapter.Reset()

		assert.NoError(t, err)
	})
}

func TestFiberAdapter_Close(t *testing.T) {
	t.Run("closes underlying store", func(t *testing.T) {
		store := &mockStore{}
		adapter := NewFiberAdapter(store)

		err := adapter.Close()

		assert.NoError(t, err)
		assert.True(t, store.closeCalled)
	})
}

// =============================================================================
// IncrementAdapter Tests
// =============================================================================

func TestNewIncrementAdapter(t *testing.T) {
	t.Run("creates adapter with store and expiration", func(t *testing.T) {
		store := &mockStore{}
		expiration := 5 * time.Minute

		adapter := NewIncrementAdapter(store, expiration)

		require.NotNil(t, adapter)
		assert.Equal(t, store, adapter.store)
		assert.Equal(t, expiration, adapter.expiration)
	})
}

func TestIncrementAdapter_Get(t *testing.T) {
	t.Run("returns count minus one after increment", func(t *testing.T) {
		store := &mockStore{count: 0} // Will become 1 after increment
		adapter := NewIncrementAdapter(store, time.Minute)

		result, err := adapter.Get("test-key")

		assert.NoError(t, err)
		require.NotNil(t, result)

		// Decode: after increment, count is 1, so result should be 0 (count - 1)
		count := int64(binary.BigEndian.Uint64(result))
		assert.Equal(t, int64(0), count)
	})

	t.Run("correctly reports pre-increment count for rate limiting", func(t *testing.T) {
		// Simulate hitting rate limit of 5
		store := &mockStore{count: 4} // Will become 5 after increment
		adapter := NewIncrementAdapter(store, time.Minute)

		result, err := adapter.Get("test-key")

		assert.NoError(t, err)
		count := int64(binary.BigEndian.Uint64(result))
		// After increment: count=5, returned=4 (which is < 5, so still passes)
		assert.Equal(t, int64(4), count)
	})

	t.Run("blocks after exceeding limit", func(t *testing.T) {
		// Simulate the blocking request
		store := &mockStore{count: 5} // Will become 6 after increment
		adapter := NewIncrementAdapter(store, time.Minute)

		result, err := adapter.Get("test-key")

		assert.NoError(t, err)
		count := int64(binary.BigEndian.Uint64(result))
		// After increment: count=6, returned=5 (which is >= 5, so blocks)
		assert.Equal(t, int64(5), count)
	})
}

func TestIncrementAdapter_Set(t *testing.T) {
	t.Run("set is a no-op", func(t *testing.T) {
		store := &mockStore{}
		adapter := NewIncrementAdapter(store, time.Minute)

		err := adapter.Set("test-key", []byte{1, 2, 3, 4, 5, 6, 7, 8}, time.Minute)

		assert.NoError(t, err)
	})
}

func TestIncrementAdapter_Delete(t *testing.T) {
	t.Run("calls reset on store", func(t *testing.T) {
		store := &mockStore{}
		adapter := NewIncrementAdapter(store, time.Minute)

		err := adapter.Delete("test-key")

		assert.NoError(t, err)
		assert.True(t, store.resetCalled)
	})
}

func TestIncrementAdapter_Reset(t *testing.T) {
	t.Run("reset returns nil for distributed stores", func(t *testing.T) {
		store := &mockStore{}
		adapter := NewIncrementAdapter(store, time.Minute)

		err := adapter.Reset()

		assert.NoError(t, err)
	})
}

func TestIncrementAdapter_Close(t *testing.T) {
	t.Run("closes underlying store", func(t *testing.T) {
		store := &mockStore{}
		adapter := NewIncrementAdapter(store, time.Minute)

		err := adapter.Close()

		assert.NoError(t, err)
		assert.True(t, store.closeCalled)
	})
}

// =============================================================================
// encodeInt64 Tests
// =============================================================================

func TestEncodeInt64(t *testing.T) {
	tests := []struct {
		name  string
		input int64
	}{
		{"zero", 0},
		{"positive", 42},
		{"large number", 1000000000},
		{"max int64", 9223372036854775807},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := encodeInt64(tt.input)

			require.Len(t, encoded, 8)

			// Decode and verify
			decoded := int64(binary.BigEndian.Uint64(encoded))
			assert.Equal(t, tt.input, decoded)
		})
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkFiberAdapter_Get(b *testing.B) {
	store := &mockStore{count: 100}
	adapter := NewFiberAdapter(store)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = adapter.Get("bench-key")
	}
}

func BenchmarkIncrementAdapter_Get(b *testing.B) {
	store := &mockStore{count: 0}
	adapter := NewIncrementAdapter(store, time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = adapter.Get("bench-key")
	}
}

func BenchmarkEncodeInt64(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = encodeInt64(int64(i))
	}
}
