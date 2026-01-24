package ratelimit

import (
	"net/http"
)

// Middleware wraps an http.Handler with rate limiting.
func Middleware(limiter *Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.TryAcquire() {
				w.Header().Set("Retry-After", "1")
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RoundTripper wraps an http.RoundTripper with rate limiting.
// Use this to rate limit outgoing HTTP requests (e.g., Riot API client).
type RoundTripper struct {
	limiter *Limiter
	next    http.RoundTripper
}

// NewRoundTripper creates a rate-limited RoundTripper.
func NewRoundTripper(limiter *Limiter, next http.RoundTripper) *RoundTripper {
	if next == nil {
		next = http.DefaultTransport
	}
	return &RoundTripper{
		limiter: limiter,
		next:    next,
	}
}

// RoundTrip implements http.RoundTripper.
func (rt *RoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	if err := rt.limiter.Wait(r.Context()); err != nil {
		return nil, err
	}
	return rt.next.RoundTrip(r)
}
