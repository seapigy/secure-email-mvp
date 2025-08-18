package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// rateLimitEntry holds request count and window start for an IP+endpoint
type rateLimitEntry struct {
	Count       int
	WindowStart time.Time
	LastSeen    time.Time
}

// IPRateLimiter is a reusable, thread-safe, in-memory rate limiter
// for per-IP per-endpoint rate limiting.
type IPRateLimiter struct {
	limit      int
	window     time.Duration
	store      sync.Map // key: ip+endpoint, value: *rateLimitEntry
	cleanupInt time.Duration
	quit       chan struct{}
	mu         sync.Mutex // Global mutex for thread safety
}

// NewIPRateLimitMiddleware returns a middleware that limits requests per IP per endpoint
func NewIPRateLimitMiddleware(limit int, window time.Duration) *IPRateLimiter {
	rl := &IPRateLimiter{
		limit:      limit,
		window:     window,
		cleanupInt: 5 * time.Minute,
		quit:       make(chan struct{}),
	}
	go rl.cleanupLoop()
	return rl
}

// Middleware returns a handler that enforces the rate limit
func (rl *IPRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := rl.getIP(r)
		endpoint := r.URL.Path
		key := ip + "|" + endpoint
		now := time.Now()

		// Use a global mutex for thread safety
		rl.mu.Lock()
		defer rl.mu.Unlock()

		val, _ := rl.store.LoadOrStore(key, &rateLimitEntry{Count: 0, WindowStart: now, LastSeen: now})
		entry := val.(*rateLimitEntry)

		// Check window
		if now.Sub(entry.WindowStart) > rl.window {
			entry.Count = 1
			entry.WindowStart = now
			entry.LastSeen = now
		} else {
			if entry.Count >= rl.limit {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]string{"error": "Too many requests. Please try again later."})
				log.Printf("[RATE LIMIT] IP %s exceeded limit on %s", ip, endpoint)
				return
			}
			entry.Count++
			entry.LastSeen = now
		}
		next.ServeHTTP(w, r)
	})
}

// getIP extracts the real client IP from the request
func (rl *IPRateLimiter) getIP(r *http.Request) string {
	// Check X-Forwarded-For for proxies
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		ip := strings.TrimSpace(parts[0])
		if net.ParseIP(ip) != nil {
			return ip
		}
	}
	// Fallback to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && net.ParseIP(ip) != nil {
		return ip
	}
	return r.RemoteAddr // as last resort
}

// cleanupLoop periodically removes stale entries
func (rl *IPRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanupInt)
	defer ticker.Stop()
	for {
		select {
		case <-rl.quit:
			return
		case now := <-ticker.C:
			rl.store.Range(func(key, value any) bool {
				entry := value.(*rateLimitEntry)
				if now.Sub(entry.LastSeen) > 10*time.Minute {
					rl.store.Delete(key)
				}
				return true
			})
		}
	}
}

// Stop stops the cleanup goroutine (for tests)
func (rl *IPRateLimiter) Stop() {
	close(rl.quit)
}
