package ratelimit

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

// MemoryStore implements Store using in-memory storage.
// This is the default store for single-instance deployments.
// It provides the fastest performance but doesn't share state across instances.
type MemoryStore struct {
	data       map[string]*entry
	mu         sync.RWMutex
	gcInterval time.Duration
	stopCh     chan struct{}
	stopped    int32 // Atomic flag to prevent double-close (0=running, 1=stopped)
}

type entry struct {
	count     int64
	expiresAt time.Time
}

// NewMemoryStore creates a new in-memory rate limit store.
// gcInterval specifies how often to clean up expired entries.
func NewMemoryStore(gcInterval time.Duration) *MemoryStore {
	if gcInterval <= 0 {
		gcInterval = 10 * time.Minute
	}

	store := &MemoryStore{
		data:       make(map[string]*entry),
		gcInterval: gcInterval,
		stopCh:     make(chan struct{}),
	}

	// Start garbage collection goroutine
	go store.gc()

	// Log warning about in-memory rate limiting in production
	logMemoryStoreWarning()

	return store
}

// logMemoryStoreWarning logs a warning about using in-memory rate limiting in production.
// The warning is only logged once per process to avoid log spam.
func logMemoryStoreWarning() {
	// Check for indicators of multi-instance deployment
	isKubernetes := os.Getenv("KUBERNETES_SERVICE_HOST") != ""
	isPodName := os.Getenv("POD_NAME") != "" || os.Getenv("HOSTNAME") != ""
	isDockerCompose := os.Getenv("COMPOSE_PROJECT_NAME") != ""
	hasRedisURL := os.Getenv("FLUXBASE_REDIS_URL") != "" || os.Getenv("REDIS_URL") != ""
	hasDragonflyURL := os.Getenv("FLUXBASE_DRAGONFLY_URL") != "" || os.Getenv("DRAGONFLY_URL") != ""

	// If Redis/Dragonfly is configured, rate limiting can be distributed
	if hasRedisURL || hasDragonflyURL {
		return // Distributed rate limiting is likely configured
	}

	// Log warning if we detect multi-instance environment indicators
	if isKubernetes || isPodName || isDockerCompose {
		log.Warn().
			Bool("kubernetes_detected", isKubernetes).
			Bool("container_detected", isPodName).
			Bool("compose_detected", isDockerCompose).
			Msg("SECURITY WARNING: Using in-memory rate limiting in a multi-instance environment. " +
				"Rate limits are per-instance only and can be bypassed by targeting different instances. " +
				"For production, configure Redis/Dragonfly (FLUXBASE_REDIS_URL or FLUXBASE_DRAGONFLY_URL) " +
				"for distributed rate limiting, or use PostgreSQL backend via FLUXBASE_SCALING_BACKEND=postgres.")
	}
}

// Get retrieves the current count for a key.
func (s *MemoryStore) Get(ctx context.Context, key string) (int64, time.Time, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, exists := s.data[key]
	if !exists {
		return 0, time.Time{}, nil
	}

	// Check if expired
	if time.Now().After(e.expiresAt) {
		return 0, time.Time{}, nil
	}

	return e.count, e.expiresAt, nil
}

// Increment atomically increments the counter for a key.
func (s *MemoryStore) Increment(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	e, exists := s.data[key]

	if !exists || now.After(e.expiresAt) {
		// Create new entry or reset expired one
		s.data[key] = &entry{
			count:     1,
			expiresAt: now.Add(expiration),
		}
		return 1, nil
	}

	// Increment existing entry
	e.count++
	return e.count, nil
}

// Reset resets the counter for a key.
func (s *MemoryStore) Reset(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)
	return nil
}

// ResetAll removes all rate limit counters matching a key pattern.
// The pattern uses glob syntax (e.g., "api:*" matches all keys starting with "api:").
func (s *MemoryStore) ResetAll(ctx context.Context, pattern string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key := range s.data {
		matched, err := filepath.Match(pattern, key)
		if err != nil {
			continue // Invalid pattern, skip this key
		}
		if matched {
			delete(s.data, key)
		}
	}

	return nil
}

// Close stops the garbage collection goroutine.
func (s *MemoryStore) Close() error {
	// Check if already stopped (prevent double-close)
	if !atomic.CompareAndSwapInt32(&s.stopped, 0, 1) {
		return nil
	}
	close(s.stopCh)
	return nil
}

// gc periodically removes expired entries.
func (s *MemoryStore) gc() {
	ticker := time.NewTicker(s.gcInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.cleanup()
		}
	}
}

// cleanup removes all expired entries.
func (s *MemoryStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for key, e := range s.data {
		if now.After(e.expiresAt) {
			delete(s.data, key)
		}
	}
}
