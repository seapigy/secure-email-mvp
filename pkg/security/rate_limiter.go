package security

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RateLimiterConfig defines rate limiting configuration
type RateLimiterConfig struct {
	RequestsPerMinute int           // Number of requests allowed per minute
	RequestsPerHour   int           // Number of requests allowed per hour
	BurstSize         int           // Maximum burst size
	WindowSize        time.Duration // Time window for rate limiting
	CleanupInterval   time.Duration // How often to clean up old entries
}

// RateLimitEntry represents a rate limit entry for a client
type RateLimitEntry struct {
	Count     int       `json:"count"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
	Blocked   bool      `json:"blocked"`
	BlockedAt *time.Time `json:"blocked_at,omitempty"`
}

// RateLimiter implements rate limiting functionality
type RateLimiter struct {
	config     RateLimiterConfig
	clients    map[string]*RateLimitEntry
	mutex      sync.RWMutex
	stopChan   chan struct{}
	cleanupTicker *time.Ticker
}

// NewRateLimiter creates a new rate limiter with the given configuration
func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	if config.RequestsPerMinute <= 0 {
		config.RequestsPerMinute = 60
	}
	if config.RequestsPerHour <= 0 {
		config.RequestsPerHour = 1000
	}
	if config.BurstSize <= 0 {
		config.BurstSize = 10
	}
	if config.WindowSize <= 0 {
		config.WindowSize = time.Minute
	}
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = 5 * time.Minute
	}

	rl := &RateLimiter{
		config:   config,
		clients:  make(map[string]*RateLimitEntry),
		stopChan: make(chan struct{}),
		cleanupTicker: time.NewTicker(config.CleanupInterval),
	}

	// Start cleanup goroutine
	go rl.cleanupRoutine()

	return rl
}

// GetClientIdentifier extracts a unique identifier for the client
func (rl *RateLimiter) GetClientIdentifier(r *http.Request) string {
	// Try to get real IP address
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Try X-Forwarded-For header
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		// Take the first IP in the chain
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Fall back to remote address
	remoteAddr := r.RemoteAddr
	if remoteAddr != "" {
		// Remove port if present
		if colonIndex := strings.LastIndex(remoteAddr, ":"); colonIndex != -1 {
			return remoteAddr[:colonIndex]
		}
		return remoteAddr
	}

	// Fallback to user agent if no IP available
	return r.Header.Get("User-Agent")
}

// IsAllowed checks if a request is allowed based on rate limiting rules
func (rl *RateLimiter) IsAllowed(clientID string) (bool, *RateLimitEntry, error) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	entry, exists := rl.clients[clientID]

	if !exists {
		// First request from this client
		entry = &RateLimitEntry{
			Count:     1,
			FirstSeen: now,
			LastSeen:  now,
			Blocked:   false,
		}
		rl.clients[clientID] = entry
		return true, entry, nil
	}

	// Check if client is blocked
	if entry.Blocked {
		// Check if block period has expired (1 hour)
		if now.Sub(*entry.BlockedAt) > time.Hour {
			// Unblock the client
			entry.Blocked = false
			entry.BlockedAt = nil
			entry.Count = 1
			entry.FirstSeen = now
			entry.LastSeen = now
			return true, entry, nil
		}
		return false, entry, fmt.Errorf("client is blocked due to rate limit violations")
	}

	// Check if we're in a new time window
	windowStart := now.Add(-rl.config.WindowSize)
	if entry.FirstSeen.Before(windowStart) {
		// Reset for new window
		entry.Count = 1
		entry.FirstSeen = now
		entry.LastSeen = now
		return true, entry, nil
	}

	// Update last seen time
	entry.LastSeen = now
	entry.Count++

	// Check rate limits
	if entry.Count > rl.config.RequestsPerMinute {
		// Block the client
		entry.Blocked = true
		entry.BlockedAt = &now
		return false, entry, fmt.Errorf("rate limit exceeded: %d requests in %v", entry.Count, rl.config.WindowSize)
	}

	// Check burst limit
	if entry.Count > rl.config.BurstSize {
		return false, entry, fmt.Errorf("burst limit exceeded: %d requests", entry.Count)
	}

	return true, entry, nil
}

// cleanupRoutine periodically cleans up old rate limit entries
func (rl *RateLimiter) cleanupRoutine() {
	for {
		select {
		case <-rl.cleanupTicker.C:
			rl.cleanup()
		case <-rl.stopChan:
			rl.cleanupTicker.Stop()
			return
		}
	}
}

// cleanup removes old rate limit entries to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	cutoff := time.Now().Add(-2 * rl.config.WindowSize) // Keep entries for 2x window size
	count := 0

	for clientID, entry := range rl.clients {
		if entry.LastSeen.Before(cutoff) {
			delete(rl.clients, clientID)
			count++
		}
	}

	if count > 0 {
		log.Printf("[RATE_LIMITER] Cleaned up %d old rate limit entries", count)
	}
}

// Stop stops the rate limiter and cleans up resources
func (rl *RateLimiter) Stop() {
	close(rl.stopChan)
}

// GetStats returns current rate limiter statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	totalClients := len(rl.clients)
	blockedClients := 0
	activeClients := 0

	for _, entry := range rl.clients {
		if entry.Blocked {
			blockedClients++
		} else {
			activeClients++
		}
	}

	return map[string]interface{}{
		"total_clients":   totalClients,
		"blocked_clients": blockedClients,
		"active_clients":  activeClients,
		"config":          rl.config,
	}
}

// RateLimitMiddleware creates a middleware that applies rate limiting
func RateLimitMiddleware(rl *RateLimiter) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			clientID := rl.GetClientIdentifier(r)
			
			allowed, entry, err := rl.IsAllowed(clientID)
			if !allowed {
				// Log the rate limit violation
				log.Printf("[RATE_LIMITER] Rate limit exceeded for client %s: %v", clientID, err)
				
				// Add rate limit headers
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.config.RequestsPerMinute))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(rl.config.WindowSize).Unix(), 10))
				w.Header().Set("Retry-After", strconv.Itoa(int(rl.config.WindowSize.Seconds())))
				
				// Return rate limit error
				http.Error(w, `{"error":"Rate limit exceeded","code":"RATE_LIMIT_EXCEEDED"}`, http.StatusTooManyRequests)
				return
			}

			// Add rate limit headers for successful requests
			remaining := rl.config.RequestsPerMinute - entry.Count
			if remaining < 0 {
				remaining = 0
			}
			
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.config.RequestsPerMinute))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(entry.FirstSeen.Add(rl.config.WindowSize).Unix(), 10))

			next.ServeHTTP(w, r)
		}
	}
}

// AdaptiveRateLimiter provides adaptive rate limiting based on client behavior
type AdaptiveRateLimiter struct {
	baseLimiter *RateLimiter
	config      AdaptiveRateLimiterConfig
	clients     map[string]*AdaptiveClientEntry
	mutex       sync.RWMutex
}

// AdaptiveRateLimiterConfig defines adaptive rate limiting configuration
type AdaptiveRateLimiterConfig struct {
	BaseRequestsPerMinute int           // Base rate limit
	MaxRequestsPerMinute  int           // Maximum rate limit
	MinRequestsPerMinute  int           // Minimum rate limit
	TrustThreshold        int           // Number of successful requests to increase limit
	PenaltyThreshold      int           // Number of violations to decrease limit
	AdjustmentFactor      float64       // Factor to adjust limits by
	TrustWindow           time.Duration // Time window for trust calculation
}

// AdaptiveClientEntry represents an adaptive rate limit entry
type AdaptiveClientEntry struct {
	BaseEntry        *RateLimitEntry
	TrustScore       int
	CurrentLimit     int
	SuccessfulReqs   int
	Violations       int
	LastAdjustment   time.Time
	TrustWindowStart time.Time
}

// NewAdaptiveRateLimiter creates a new adaptive rate limiter
func NewAdaptiveRateLimiter(config AdaptiveRateLimiterConfig) *AdaptiveRateLimiter {
	if config.BaseRequestsPerMinute <= 0 {
		config.BaseRequestsPerMinute = 60
	}
	if config.MaxRequestsPerMinute <= 0 {
		config.MaxRequestsPerMinute = 300
	}
	if config.MinRequestsPerMinute <= 0 {
		config.MinRequestsPerMinute = 10
	}
	if config.TrustThreshold <= 0 {
		config.TrustThreshold = 100
	}
	if config.PenaltyThreshold <= 0 {
		config.PenaltyThreshold = 5
	}
	if config.AdjustmentFactor <= 0 {
		config.AdjustmentFactor = 0.1
	}
	if config.TrustWindow <= 0 {
		config.TrustWindow = time.Hour
	}

	baseConfig := RateLimiterConfig{
		RequestsPerMinute: config.BaseRequestsPerMinute,
		RequestsPerHour:   config.BaseRequestsPerMinute * 60,
		BurstSize:         10,
		WindowSize:        time.Minute,
		CleanupInterval:   5 * time.Minute,
	}

	return &AdaptiveRateLimiter{
		baseLimiter: NewRateLimiter(baseConfig),
		config:      config,
		clients:     make(map[string]*AdaptiveClientEntry),
	}
}

// IsAllowed checks if a request is allowed with adaptive rate limiting
func (arl *AdaptiveRateLimiter) IsAllowed(clientID string) (bool, *RateLimitEntry, error) {
	arl.mutex.Lock()
	defer arl.mutex.Unlock()

	// Get or create adaptive entry
	entry, exists := arl.clients[clientID]
	if !exists {
		entry = &AdaptiveClientEntry{
			TrustScore:       0,
			CurrentLimit:     arl.config.BaseRequestsPerMinute,
			SuccessfulReqs:   0,
			Violations:       0,
			LastAdjustment:   time.Now(),
			TrustWindowStart: time.Now(),
		}
		arl.clients[clientID] = entry
	}

	// Check if trust window has expired
	now := time.Now()
	if now.Sub(entry.TrustWindowStart) > arl.config.TrustWindow {
		// Reset trust window
		entry.TrustWindowStart = now
		entry.SuccessfulReqs = 0
		entry.Violations = 0
	}

	// Temporarily set the base limiter's limit for this client
	originalLimit := arl.baseLimiter.config.RequestsPerMinute
	arl.baseLimiter.config.RequestsPerMinute = entry.CurrentLimit

	// Check with base limiter
	allowed, baseEntry, err := arl.baseLimiter.IsAllowed(clientID)
	entry.BaseEntry = baseEntry

	// Restore original limit
	arl.baseLimiter.config.RequestsPerMinute = originalLimit

	if allowed {
		// Successful request - increase trust
		entry.SuccessfulReqs++
		if entry.SuccessfulReqs >= arl.config.TrustThreshold {
			// Increase rate limit
			arl.adjustLimit(entry, true)
		}
	} else {
		// Violation - decrease trust
		entry.Violations++
		if entry.Violations >= arl.config.PenaltyThreshold {
			// Decrease rate limit
			arl.adjustLimit(entry, false)
		}
	}

	return allowed, baseEntry, err
}

// adjustLimit adjusts the rate limit for a client
func (arl *AdaptiveRateLimiter) adjustLimit(entry *AdaptiveClientEntry, increase bool) {
	now := time.Now()
	
	// Prevent too frequent adjustments
	if now.Sub(entry.LastAdjustment) < time.Minute {
		return
	}

	oldLimit := entry.CurrentLimit
	if increase {
		// Increase limit
		newLimit := int(float64(entry.CurrentLimit) * (1 + arl.config.AdjustmentFactor))
		if newLimit > arl.config.MaxRequestsPerMinute {
			newLimit = arl.config.MaxRequestsPerMinute
		}
		entry.CurrentLimit = newLimit
		entry.TrustScore++
	} else {
		// Decrease limit
		newLimit := int(float64(entry.CurrentLimit) * (1 - arl.config.AdjustmentFactor))
		if newLimit < arl.config.MinRequestsPerMinute {
			newLimit = arl.config.MinRequestsPerMinute
		}
		entry.CurrentLimit = newLimit
		entry.TrustScore--
	}

	entry.LastAdjustment = now
	entry.SuccessfulReqs = 0
	entry.Violations = 0

	log.Printf("[ADAPTIVE_RATE_LIMITER] Adjusted limit for client: %d -> %d (trust_score: %d)", 
		oldLimit, entry.CurrentLimit, entry.TrustScore)
}

// GetClientIdentifier delegates to base limiter
func (arl *AdaptiveRateLimiter) GetClientIdentifier(r *http.Request) string {
	return arl.baseLimiter.GetClientIdentifier(r)
}

// Stop stops the adaptive rate limiter
func (arl *AdaptiveRateLimiter) Stop() {
	arl.baseLimiter.Stop()
}

// GetStats returns adaptive rate limiter statistics
func (arl *AdaptiveRateLimiter) GetStats() map[string]interface{} {
	arl.mutex.RLock()
	defer arl.mutex.RUnlock()

	baseStats := arl.baseLimiter.GetStats()
	
	totalClients := len(arl.clients)
	trustedClients := 0
	penalizedClients := 0
	averageTrustScore := 0.0

	if totalClients > 0 {
		totalTrustScore := 0
		for _, entry := range arl.clients {
			if entry.TrustScore > 0 {
				trustedClients++
			} else if entry.TrustScore < 0 {
				penalizedClients++
			}
			totalTrustScore += entry.TrustScore
		}
		averageTrustScore = float64(totalTrustScore) / float64(totalClients)
	}

	return map[string]interface{}{
		"base_stats":        baseStats,
		"total_clients":     totalClients,
		"trusted_clients":   trustedClients,
		"penalized_clients": penalizedClients,
		"avg_trust_score":   averageTrustScore,
		"config":            arl.config,
	}
}

// AdaptiveRateLimitMiddleware creates a middleware that applies adaptive rate limiting
func AdaptiveRateLimitMiddleware(arl *AdaptiveRateLimiter) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			clientID := arl.GetClientIdentifier(r)
			
			allowed, entry, err := arl.IsAllowed(clientID)
			if !allowed {
				// Log the rate limit violation
				log.Printf("[ADAPTIVE_RATE_LIMITER] Rate limit exceeded for client %s: %v", clientID, err)
				
				// Add rate limit headers
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(entry.Count))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Minute).Unix(), 10))
				w.Header().Set("Retry-After", "60")
				
				// Return rate limit error
				http.Error(w, `{"error":"Rate limit exceeded","code":"RATE_LIMIT_EXCEEDED"}`, http.StatusTooManyRequests)
				return
			}

			// Add rate limit headers for successful requests
			remaining := entry.Count - 1
			if remaining < 0 {
				remaining = 0
			}
			
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(entry.Count))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(entry.FirstSeen.Add(time.Minute).Unix(), 10))

			next.ServeHTTP(w, r)
		}
	}
}
