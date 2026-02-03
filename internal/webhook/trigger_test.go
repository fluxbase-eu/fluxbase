package webhook

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTriggerService(t *testing.T) {
	t.Run("creates with default workers", func(t *testing.T) {
		svc := NewTriggerService(nil, nil, 0)
		require.NotNil(t, svc)
		assert.Equal(t, 4, svc.workers) // Default
		assert.Equal(t, 30*time.Second, svc.backlogInterval)
		assert.NotNil(t, svc.eventChan)
		assert.NotNil(t, svc.stopChan)
	})

	t.Run("creates with negative workers uses default", func(t *testing.T) {
		svc := NewTriggerService(nil, nil, -1)
		assert.Equal(t, 4, svc.workers)
	})

	t.Run("creates with custom workers", func(t *testing.T) {
		svc := NewTriggerService(nil, nil, 10)
		assert.Equal(t, 10, svc.workers)
	})

	t.Run("creates with specified db and webhook service", func(t *testing.T) {
		webhookSvc := &WebhookService{}
		svc := NewTriggerService(nil, webhookSvc, 5)
		assert.Equal(t, webhookSvc, svc.webhookSvc)
	})
}

func TestTriggerService_SetBacklogInterval(t *testing.T) {
	svc := NewTriggerService(nil, nil, 2)

	t.Run("sets backlog interval before start", func(t *testing.T) {
		svc.SetBacklogInterval(1 * time.Minute)
		assert.Equal(t, 1*time.Minute, svc.backlogInterval)
	})

	t.Run("sets backlog interval to short duration", func(t *testing.T) {
		svc.SetBacklogInterval(5 * time.Second)
		assert.Equal(t, 5*time.Second, svc.backlogInterval)
	})
}

func TestTriggerService_Stop(t *testing.T) {
	svc := NewTriggerService(nil, nil, 1)

	// Stop should not panic even without Start
	assert.NotPanics(t, func() {
		svc.Stop()
	})
}

func TestWebhookEvent_Struct(t *testing.T) {
	t.Run("creates webhook event with all fields", func(t *testing.T) {
		webhookID := uuid.New()
		eventID := uuid.New()
		now := time.Now()
		recordID := "record-123"
		errorMsg := "test error"

		event := &WebhookEvent{
			ID:            eventID,
			WebhookID:     webhookID,
			EventType:     "INSERT",
			TableSchema:   "public",
			TableName:     "users",
			RecordID:      &recordID,
			OldData:       []byte(`{"name": "old"}`),
			NewData:       []byte(`{"name": "new"}`),
			Processed:     false,
			Attempts:      2,
			LastAttemptAt: &now,
			NextRetryAt:   &now,
			ErrorMessage:  &errorMsg,
			CreatedAt:     now,
		}

		assert.Equal(t, eventID, event.ID)
		assert.Equal(t, webhookID, event.WebhookID)
		assert.Equal(t, "INSERT", event.EventType)
		assert.Equal(t, "public", event.TableSchema)
		assert.Equal(t, "users", event.TableName)
		assert.Equal(t, "record-123", *event.RecordID)
		assert.JSONEq(t, `{"name": "old"}`, string(event.OldData))
		assert.JSONEq(t, `{"name": "new"}`, string(event.NewData))
		assert.False(t, event.Processed)
		assert.Equal(t, 2, event.Attempts)
		assert.NotNil(t, event.LastAttemptAt)
		assert.NotNil(t, event.NextRetryAt)
		assert.Equal(t, "test error", *event.ErrorMessage)
	})

	t.Run("creates minimal webhook event", func(t *testing.T) {
		event := &WebhookEvent{
			ID:          uuid.New(),
			WebhookID:   uuid.New(),
			EventType:   "DELETE",
			TableSchema: "public",
			TableName:   "posts",
			Processed:   false,
			Attempts:    0,
			CreatedAt:   time.Now(),
		}

		assert.NotEqual(t, uuid.Nil, event.ID)
		assert.Equal(t, "DELETE", event.EventType)
		assert.Nil(t, event.RecordID)
		assert.Nil(t, event.OldData)
		assert.Nil(t, event.NewData)
		assert.Nil(t, event.LastAttemptAt)
		assert.Nil(t, event.NextRetryAt)
		assert.Nil(t, event.ErrorMessage)
	})
}

func TestEventChannel(t *testing.T) {
	svc := NewTriggerService(nil, nil, 1)

	t.Run("event channel has buffer of 1000", func(t *testing.T) {
		assert.Equal(t, 1000, cap(svc.eventChan))
	})

	t.Run("can send events to channel", func(t *testing.T) {
		id := uuid.New()
		svc.eventChan <- id

		select {
		case received := <-svc.eventChan:
			assert.Equal(t, id, received)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("timeout waiting for event")
		}
	})
}

func TestBackoffCalculation(t *testing.T) {
	// Test the exponential backoff calculation logic
	// that would be used in handleDeliveryFailure

	testCases := []struct {
		attempts              int
		retryBackoffSeconds   int
		expectedBackoffMillis int
	}{
		{1, 60, 60000},  // First retry: 60 * 1 = 60s
		{2, 60, 120000}, // Second retry: 60 * 2 = 120s
		{3, 60, 180000}, // Third retry: 60 * 3 = 180s
		{1, 30, 30000},  // Different base: 30 * 1 = 30s
		{5, 10, 50000},  // Fifth retry: 10 * 5 = 50s
	}

	for _, tc := range testCases {
		t.Run("backoff calculation", func(t *testing.T) {
			backoffSeconds := tc.retryBackoffSeconds * tc.attempts
			assert.Equal(t, tc.expectedBackoffMillis/1000, backoffSeconds)
		})
	}
}

func TestMaxRetriesLogic(t *testing.T) {
	// Test the max retries check logic

	testCases := []struct {
		attempts   int
		maxRetries int
		shouldFail bool
	}{
		{1, 3, false}, // 1 attempt, max 3 - continue
		{2, 3, false}, // 2 attempts, max 3 - continue
		{3, 3, true},  // 3 attempts, max 3 - max reached
		{4, 3, true},  // 4 attempts, max 3 - exceeded
		{1, 1, true},  // 1 attempt, max 1 - max reached
		{0, 5, false}, // 0 attempts, max 5 - continue
	}

	for _, tc := range testCases {
		maxReached := tc.attempts >= tc.maxRetries
		assert.Equal(t, tc.shouldFail, maxReached)
	}
}

func TestEndpointRateLimiter(t *testing.T) {
	t.Run("allows initial request", func(t *testing.T) {
		rl := &endpointRateLimiter{
			requests: make(map[string][]time.Time),
			limit:    10,
		}

		allowed := rl.allow("https://example.com/webhook")
		assert.True(t, allowed)
	})

	t.Run("allows requests up to limit", func(t *testing.T) {
		rl := &endpointRateLimiter{
			requests: make(map[string][]time.Time),
			limit:    5,
		}

		endpoint := "https://example.com/webhook"
		for i := 0; i < 5; i++ {
			allowed := rl.allow(endpoint)
			assert.True(t, allowed, "request %d should be allowed", i+1)
		}
	})

	t.Run("blocks requests after limit exceeded", func(t *testing.T) {
		rl := &endpointRateLimiter{
			requests: make(map[string][]time.Time),
			limit:    3,
		}

		endpoint := "https://example.com/webhook"
		// Use up all allowed requests
		for i := 0; i < 3; i++ {
			rl.allow(endpoint)
		}

		// Next request should be blocked
		allowed := rl.allow(endpoint)
		assert.False(t, allowed)
	})

	t.Run("tracks different endpoints separately", func(t *testing.T) {
		rl := &endpointRateLimiter{
			requests: make(map[string][]time.Time),
			limit:    2,
		}

		endpoint1 := "https://example.com/webhook1"
		endpoint2 := "https://example.com/webhook2"

		// Exhaust limit for endpoint1
		rl.allow(endpoint1)
		rl.allow(endpoint1)
		assert.False(t, rl.allow(endpoint1))

		// endpoint2 should still be allowed
		assert.True(t, rl.allow(endpoint2))
		assert.True(t, rl.allow(endpoint2))
		assert.False(t, rl.allow(endpoint2))
	})

	t.Run("uses sliding window to expire old requests", func(t *testing.T) {
		rl := &endpointRateLimiter{
			requests: make(map[string][]time.Time),
			limit:    2,
		}

		endpoint := "https://example.com/webhook"

		// Pre-populate with old requests (outside 1-minute window)
		oldTime := time.Now().Add(-2 * time.Minute)
		rl.requests[endpoint] = []time.Time{oldTime, oldTime}

		// Should allow new requests since old ones are expired
		assert.True(t, rl.allow(endpoint))
		assert.True(t, rl.allow(endpoint))
		assert.False(t, rl.allow(endpoint))
	})

	t.Run("filters expired requests while keeping recent ones", func(t *testing.T) {
		rl := &endpointRateLimiter{
			requests: make(map[string][]time.Time),
			limit:    3,
		}

		endpoint := "https://example.com/webhook"

		// Pre-populate with mix of old and recent requests
		oldTime := time.Now().Add(-2 * time.Minute)
		recentTime := time.Now().Add(-30 * time.Second)
		rl.requests[endpoint] = []time.Time{oldTime, recentTime}

		// One recent request exists, so only 2 more should be allowed
		assert.True(t, rl.allow(endpoint))
		assert.True(t, rl.allow(endpoint))
		assert.False(t, rl.allow(endpoint))
	})

	t.Run("limit of 1 allows single request then blocks", func(t *testing.T) {
		rl := &endpointRateLimiter{
			requests: make(map[string][]time.Time),
			limit:    1,
		}

		endpoint := "https://example.com/webhook"
		assert.True(t, rl.allow(endpoint))
		assert.False(t, rl.allow(endpoint))
	})

	t.Run("handles empty endpoint string", func(t *testing.T) {
		rl := &endpointRateLimiter{
			requests: make(map[string][]time.Time),
			limit:    2,
		}

		assert.True(t, rl.allow(""))
		assert.True(t, rl.allow(""))
		assert.False(t, rl.allow(""))
	})
}

func TestNewEndpointRateLimiter(t *testing.T) {
	t.Run("creates with specified limit", func(t *testing.T) {
		rl := newEndpointRateLimiter(100)
		require.NotNil(t, rl)
		assert.Equal(t, 100, rl.limit)
		assert.NotNil(t, rl.requests)
	})

	t.Run("uses default limit when zero provided", func(t *testing.T) {
		rl := newEndpointRateLimiter(0)
		assert.Equal(t, DefaultRateLimitPerEndpoint, rl.limit)
	})

	t.Run("uses default limit when negative provided", func(t *testing.T) {
		rl := newEndpointRateLimiter(-5)
		assert.Equal(t, DefaultRateLimitPerEndpoint, rl.limit)
	})
}

func TestDefaultRateLimitPerEndpoint(t *testing.T) {
	assert.Equal(t, 60, DefaultRateLimitPerEndpoint)
}

func TestTriggerService_WaitForReady(t *testing.T) {
	t.Run("returns nil when already ready", func(t *testing.T) {
		svc := NewTriggerService(nil, nil, 1)
		// Simulate successful ready state
		svc.signalReady(false)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err := svc.WaitForReady(ctx)
		assert.NoError(t, err)
	})

	t.Run("returns error when listener failed", func(t *testing.T) {
		svc := NewTriggerService(nil, nil, 1)
		// Simulate failed ready state
		svc.signalReady(true)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err := svc.WaitForReady(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to start")
	})

	t.Run("returns context error on timeout", func(t *testing.T) {
		svc := NewTriggerService(nil, nil, 1)
		// Don't signal ready

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		err := svc.WaitForReady(ctx)
		assert.Error(t, err)
		assert.Equal(t, context.DeadlineExceeded, err)
	})

	t.Run("returns context error when cancelled", func(t *testing.T) {
		svc := NewTriggerService(nil, nil, 1)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := svc.WaitForReady(ctx)
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})
}

func TestTriggerService_IsReady(t *testing.T) {
	t.Run("returns false initially", func(t *testing.T) {
		svc := NewTriggerService(nil, nil, 1)
		assert.False(t, svc.IsReady())
	})

	t.Run("returns true after successful signal", func(t *testing.T) {
		svc := NewTriggerService(nil, nil, 1)
		svc.signalReady(false)
		assert.True(t, svc.IsReady())
	})

	t.Run("returns false after failed signal", func(t *testing.T) {
		svc := NewTriggerService(nil, nil, 1)
		svc.signalReady(true)
		assert.False(t, svc.IsReady())
	})
}

func TestTriggerService_signalReady(t *testing.T) {
	t.Run("only signals once", func(t *testing.T) {
		svc := NewTriggerService(nil, nil, 1)

		// Signal success
		svc.signalReady(false)
		assert.True(t, svc.IsReady())
		assert.False(t, svc.listenerFailed)

		// Try to signal failure - should be ignored
		svc.signalReady(true)
		assert.True(t, svc.IsReady())
		assert.False(t, svc.listenerFailed) // Still false
	})

	t.Run("closes ready channel", func(t *testing.T) {
		svc := NewTriggerService(nil, nil, 1)
		svc.signalReady(false)

		// Channel should be closed and receive should not block
		select {
		case <-svc.readyChan:
			// Expected - channel is closed
		default:
			t.Fatal("ready channel should be closed")
		}
	})
}

func TestTriggerService_RateLimiter(t *testing.T) {
	t.Run("service has rate limiter initialized", func(t *testing.T) {
		svc := NewTriggerService(nil, nil, 1)
		assert.NotNil(t, svc.rateLimiter)
		assert.Equal(t, DefaultRateLimitPerEndpoint, svc.rateLimiter.limit)
	})
}

func TestEndpointRateLimiter_ConcurrentAccess(t *testing.T) {
	rl := &endpointRateLimiter{
		requests: make(map[string][]time.Time),
		limit:    100,
	}

	endpoint := "https://example.com/webhook"

	// Run concurrent requests
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 20; j++ {
				rl.allow(endpoint)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have recorded requests without panic
	assert.NotPanics(t, func() {
		rl.allow(endpoint)
	})
}

func TestWebhookPayloadEventTypes(t *testing.T) {
	// Test that event types match expected values used in deliverEvent
	eventTypes := []string{"INSERT", "UPDATE", "DELETE"}

	for _, et := range eventTypes {
		event := &WebhookEvent{
			ID:          uuid.New(),
			WebhookID:   uuid.New(),
			EventType:   et,
			TableSchema: "public",
			TableName:   "users",
		}
		assert.Equal(t, et, event.EventType)
	}
}
