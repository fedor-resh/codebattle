package httpapi

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type limitEntry struct {
	windowStart time.Time
	count       int
}

type rateLimiter struct {
	mu      sync.Mutex
	entries map[string]limitEntry
	now     func() time.Time
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{entries: map[string]limitEntry{}, now: time.Now}
}

func (l *rateLimiter) wrap(limit int, window time.Duration, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := l.now()
		key := clientIP(r) + "|" + r.URL.Path
		l.mu.Lock()
		entry := l.entries[key]
		if entry.windowStart.IsZero() || now.Sub(entry.windowStart) >= window {
			entry = limitEntry{windowStart: now}
		}
		entry.count++
		l.entries[key] = entry
		allowed := entry.count <= limit
		l.mu.Unlock()
		if !allowed {
			w.Header().Set("Retry-After", "60")
			writeError(w, r, http.StatusTooManyRequests, "RATE_LIMITED", "Слишком много запросов, попробуйте позже", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]); forwarded != "" {
		return forwarded
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
