package httpserver

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type rateBucket struct {
	windowStart time.Time
	count       int
}

type ipRateLimiter struct {
	mu      sync.Mutex
	buckets map[string]rateBucket
	limit   int
	window  time.Duration
}

func RateLimitByIP(next http.HandlerFunc, limit int, window time.Duration) http.HandlerFunc {
	if limit < 1 || window <= 0 {
		return next
	}
	l := &ipRateLimiter{buckets: make(map[string]rateBucket), limit: limit, window: window}
	return func(w http.ResponseWriter, r *http.Request) {
		if !l.allow(clientIP(r), time.Now()) {
			w.Header().Set("Retry-After", "60")
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate_limit_exceeded"})
			return
		}
		next(w, r)
	}
}

func (l *ipRateLimiter) allow(ip string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	b := l.buckets[ip]
	if b.windowStart.IsZero() || now.Sub(b.windowStart) >= l.window {
		l.buckets[ip] = rateBucket{windowStart: now, count: 1}
		return true
	}
	if b.count >= l.limit {
		return false
	}
	b.count++
	l.buckets[ip] = b
	return true
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}
