package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type RateLimiter interface {
	Limit(next http.Handler) http.Handler
}

type ipLimiter struct {
	limit   int
	window  time.Duration
	mu      sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	count int
	reset time.Time
}

func NewIPRateLimiter(limit int, window time.Duration) RateLimiter {
	return &ipLimiter{limit: limit, window: window, buckets: make(map[string]*bucket)}
}

func (l *ipLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		now := time.Now()
		l.mu.Lock()
		b, ok := l.buckets[ip]
		if !ok || now.After(b.reset) {
			b = &bucket{count: 0, reset: now.Add(l.window)}
			l.buckets[ip] = b
		}
		if b.count >= l.limit {
			w.Header().Set("Retry-After", time.Until(b.reset).Round(time.Second).String())
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			l.mu.Unlock()
			return
		}
		b.count++
		l.mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	// Trust X-Forwarded-For first, else RemoteAddr
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
