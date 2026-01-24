package ratelimit_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/HatiCode/league-buddy/pkg/ratelimit"
)

func TestLimiter_AllowsWithinLimit(t *testing.T) {
	limiter := ratelimit.NewLimiter(
		ratelimit.WithLimit(5, time.Second),
	)

	// Should allow 5 requests
	for i := 0; i < 5; i++ {
		if err := limiter.Wait(context.Background()); err != nil {
			t.Fatalf("request %d should be allowed: %v", i+1, err)
		}
	}
}

func TestLimiter_BlocksOverLimit(t *testing.T) {
	limiter := ratelimit.NewLimiter(
		ratelimit.WithLimit(2, 100*time.Millisecond),
	)

	// Exhaust the limit
	limiter.Wait(context.Background())
	limiter.Wait(context.Background())

	// Third request should block until window resets
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := limiter.Wait(ctx)
	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestLimiter_ResetsAfterWindow(t *testing.T) {
	limiter := ratelimit.NewLimiter(
		ratelimit.WithLimit(2, 50*time.Millisecond),
	)

	// Exhaust the limit
	limiter.Wait(context.Background())
	limiter.Wait(context.Background())

	// Wait for window to reset
	time.Sleep(60 * time.Millisecond)

	// Should allow again
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	if err := limiter.Wait(ctx); err != nil {
		t.Errorf("should allow after window reset: %v", err)
	}
}

func TestLimiter_MultipleRules(t *testing.T) {
	limiter := ratelimit.NewLimiter(
		ratelimit.WithLimit(3, 50*time.Millisecond),  // 3 per 50ms
		ratelimit.WithLimit(5, 200*time.Millisecond), // 5 per 200ms
	)

	// Should allow 3 requests (first rule limits)
	for i := 0; i < 3; i++ {
		if err := limiter.Wait(context.Background()); err != nil {
			t.Fatalf("request %d should be allowed: %v", i+1, err)
		}
	}

	// 4th should block due to first rule
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := limiter.Wait(ctx)
	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded from first rule, got %v", err)
	}

	// Wait for first rule to reset
	time.Sleep(60 * time.Millisecond)

	// Should allow 2 more (second rule now limits at 5 total)
	for i := 0; i < 2; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		err := limiter.Wait(ctx)
		cancel()
		if err != nil {
			t.Fatalf("request %d after reset should be allowed: %v", i+4, err)
		}
	}

	// 6th should block due to second rule
	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel2()

	err = limiter.Wait(ctx2)
	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded from second rule, got %v", err)
	}
}

func TestLimiter_TryAcquire(t *testing.T) {
	limiter := ratelimit.NewLimiter(
		ratelimit.WithLimit(1, time.Second),
	)

	// First should succeed
	if !limiter.TryAcquire() {
		t.Error("first TryAcquire should succeed")
	}

	// Second should fail immediately (non-blocking)
	if limiter.TryAcquire() {
		t.Error("second TryAcquire should fail")
	}
}

func TestLimiter_ConcurrentAccess(t *testing.T) {
	limiter := ratelimit.NewLimiter(
		ratelimit.WithLimit(10, time.Second),
	)

	var wg sync.WaitGroup
	allowed := make(chan struct{}, 20)

	// Spawn 20 goroutines trying to acquire
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()
			if limiter.Wait(ctx) == nil {
				allowed <- struct{}{}
			}
		}()
	}

	wg.Wait()
	close(allowed)

	count := len(allowed)
	if count != 10 {
		t.Errorf("expected exactly 10 allowed, got %d", count)
	}
}

func TestLimiter_ContextCancellation(t *testing.T) {
	limiter := ratelimit.NewLimiter(
		ratelimit.WithLimit(1, time.Second),
	)

	// Exhaust limit
	limiter.Wait(context.Background())

	// Cancel context before waiting
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := limiter.Wait(ctx)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestLimiter_RiotAPILimits(t *testing.T) {
	// Simulate Riot API limits: 20/sec and 100/2min
	// We'll use smaller windows for testing
	limiter := ratelimit.NewLimiter(
		ratelimit.WithLimit(20, 100*time.Millisecond),  // 20 per 100ms (simulating 20/sec)
		ratelimit.WithLimit(100, 500*time.Millisecond), // 100 per 500ms (simulating 100/2min)
	)

	// Should allow 20 rapid requests
	for i := 0; i < 20; i++ {
		if err := limiter.Wait(context.Background()); err != nil {
			t.Fatalf("request %d should be allowed: %v", i+1, err)
		}
	}

	// 21st should be blocked by first rule
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	if err := limiter.Wait(ctx); err != context.DeadlineExceeded {
		t.Errorf("request 21 should be rate limited: %v", err)
	}
}
