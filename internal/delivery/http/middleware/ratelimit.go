package middleware

import (
	"geminiBackend/pkg/utils"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter interface {
	Limit(next http.Handler) http.Handler
	GinMiddleware() gin.HandlerFunc
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
			utils.Error(w, http.StatusTooManyRequests, "rate_limit", "rate limit exceeded")
			l.mu.Unlock()
			return
		}
		b.count++
		l.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

func (l *ipLimiter) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := clientIPGin(c)
		now := time.Now()

		l.mu.Lock()
		b, ok := l.buckets[ip]
		if !ok || now.After(b.reset) {
			b = &bucket{count: 0, reset: now.Add(l.window)}
			l.buckets[ip] = b
		}
		if b.count >= l.limit {
			c.Header("Retry-After", time.Until(b.reset).Round(time.Second).String())
			utils.Error(c.Writer, http.StatusTooManyRequests, "rate_limit", "rate limit exceeded")
			l.mu.Unlock()
			c.Abort()
			return
		}
		b.count++
		l.mu.Unlock()

		c.Next()
	}
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func clientIPGin(c *gin.Context) string {
	// Trust X-Forwarded-For first, else RemoteAddr
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		return xff
	}
	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return host
}
