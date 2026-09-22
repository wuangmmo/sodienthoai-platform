package httpserver

import (
	"fmt"
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
			w.Header().Set("Retry-After", fmt.Sprintf("%d", maxInt64(1, int64(l.retryAfter(clientIP(r), time.Now()).Seconds()))))
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
		if len(l.buckets) > 10000 { for key, old := range l.buckets { if now.Sub(old.windowStart) >= l.window { delete(l.buckets,key) } } }
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

func (l *ipRateLimiter) retryAfter(ip string, now time.Time) time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()
	b, ok := l.buckets[ip]
	if !ok || b.windowStart.IsZero() { return l.window }
	remaining := l.window - now.Sub(b.windowStart)
	if remaining < time.Second { return time.Second }
	return remaining
}

func maxInt64(a,b int64) int64 { if a>b { return a }; return b }

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}
