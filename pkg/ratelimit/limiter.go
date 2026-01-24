package ratelimit

import (
	"context"
	"sync"
	"time"
)

// rule defines a rate limit: max requests within a time window.
type rule struct {
	limit  int
	window time.Duration
}

// bucket tracks request timestamps for a single rule.
type bucket struct {
	rule      rule
	timestamps []time.Time
}

// Limiter enforces rate limits using a sliding window algorithm.
type Limiter struct {
	mu      sync.Mutex
	buckets []*bucket
}

// Option configures a Limiter.
type Option func(*Limiter)

// WithLimit adds a rate limit rule.
func WithLimit(limit int, window time.Duration) Option {
	return func(l *Limiter) {
		l.buckets = append(l.buckets, &bucket{
			rule:       rule{limit: limit, window: window},
			timestamps: make([]time.Time, 0, limit),
		})
	}
}

// NewLimiter creates a new rate limiter with the given rules.
func NewLimiter(opts ...Option) *Limiter {
	l := &Limiter{
		buckets: make([]*bucket, 0),
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// Wait blocks until a request is allowed or the context is done.
func (l *Limiter) Wait(ctx context.Context) error {
	for {
		if l.TryAcquire() {
			return nil
		}

		waitDuration := l.nextAvailable()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitDuration):
		}
	}
}

// TryAcquire attempts to acquire a slot without blocking.
// Returns true if allowed, false if rate limited.
func (l *Limiter) TryAcquire() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	for _, b := range l.buckets {
		l.pruneExpired(b, now)
		if len(b.timestamps) >= b.rule.limit {
			return false
		}
	}

	for _, b := range l.buckets {
		b.timestamps = append(b.timestamps, now)
	}

	return true
}

// pruneExpired removes timestamps outside the window.
func (l *Limiter) pruneExpired(b *bucket, now time.Time) {
	cutoff := now.Add(-b.rule.window)
	i := 0
	for ; i < len(b.timestamps); i++ {
		if b.timestamps[i].After(cutoff) {
			break
		}
	}
	b.timestamps = b.timestamps[i:]
}

// nextAvailable returns the duration until the next slot is available.
func (l *Limiter) nextAvailable() time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	var maxWait time.Duration

	for _, b := range l.buckets {
		l.pruneExpired(b, now)
		if len(b.timestamps) >= b.rule.limit {
			oldest := b.timestamps[0]
			wait := oldest.Add(b.rule.window).Sub(now)
			if wait > maxWait {
				maxWait = wait
			}
		}
	}

	if maxWait < time.Millisecond {
		maxWait = time.Millisecond
	}

	return maxWait
}
