package router

import (
	"net/http"
	"sync"
	"time"
)

// ipLimiter tracks per-IP request counts within a sliding window.
type ipLimiter struct {
	mu       sync.Mutex
	counts   map[string][]time.Time
	window   time.Duration
	maxReqs  int
}

func newIPLimiter(window time.Duration, maxReqs int) *ipLimiter {
	return &ipLimiter{
		counts:  make(map[string][]time.Time),
		window:  window,
		maxReqs: maxReqs,
	}
}

// allow returns true if the given IP is within the rate limit.
func (l *ipLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)

	times := l.counts[ip]
	valid := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= l.maxReqs {
		l.counts[ip] = valid
		return false
	}

	l.counts[ip] = append(valid, now)
	return true
}

// WithRateLimit returns middleware that limits requests per IP.
// window is the sliding window duration; maxReqs is the maximum allowed
// requests per IP within that window. Excess requests receive 429.
func WithRateLimit(window time.Duration, maxReqs int) func(http.Handler) http.Handler {
	if window <= 0 || maxReqs <= 0 {
		return func(next http.Handler) http.Handler { return next }
	}

	limiter := newIPLimiter(window, maxReqs)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := realIP(r)
			if !limiter.allow(ip) {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// realIP extracts the client IP from common headers or RemoteAddr.
func realIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}
