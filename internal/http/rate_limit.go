package http

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter defines the interface for rate limiting implementations
type RateLimiter interface {
	// Allow checks if a request from the given identifier should be allowed
	// Returns true if allowed, false if rate limit exceeded
	Allow(identifier string) bool

	// AllowN checks if N requests from the given identifier should be allowed
	AllowN(identifier string, n int) bool

	// Limit returns the configured rate limit
	Limit() rate.Limit

	// Burst returns the configured burst size
	Burst() int

	// Reset resets the rate limiter for the given identifier
	Reset(identifier string)

	// Clear removes old entries to prevent memory leaks
	Clear()
}

// InMemoryRateLimiter implements token bucket rate limiting in memory
// Uses golang.org/x/time/rate for thread-safe token bucket algorithm
type InMemoryRateLimiter struct {
	mu         sync.RWMutex
	limiters   map[string]*limiterEntry
	limit      rate.Limit // requests per second
	burst      int        // max burst size
	cleanupTTL time.Duration
}

// limiterEntry holds a rate limiter and its last access time
type limiterEntry struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

// NewInMemoryRateLimiter creates a new in-memory rate limiter
// Parameters:
//   - requestsPerWindow: number of requests allowed per time window
//   - window: time window for rate limiting (e.g., 1 hour, 1 minute)
//   - burst: maximum burst size (allows brief spikes above the rate)
func NewInMemoryRateLimiter(requestsPerWindow int, window time.Duration, burst int) *InMemoryRateLimiter {
	// Convert requests per window to requests per second (rate.Limit)
	rateLimit := rate.Limit(float64(requestsPerWindow) / window.Seconds())

	rl := &InMemoryRateLimiter{
		limiters:   make(map[string]*limiterEntry),
		limit:      rateLimit,
		burst:      burst,
		cleanupTTL: window * 2, // Keep entries for 2x the window
	}

	// Start background cleanup goroutine
	go rl.startCleanup()

	return rl
}

// Allow checks if a single request should be allowed for the given identifier
func (rl *InMemoryRateLimiter) Allow(identifier string) bool {
	return rl.AllowN(identifier, 1)
}

// AllowN checks if N requests should be allowed for the given identifier
func (rl *InMemoryRateLimiter) AllowN(identifier string, n int) bool {
	limiter := rl.getLimiter(identifier)
	return limiter.AllowN(time.Now(), n)
}

// Limit returns the configured rate limit (requests per second)
func (rl *InMemoryRateLimiter) Limit() rate.Limit {
	return rl.limit
}

// Burst returns the configured burst size
func (rl *InMemoryRateLimiter) Burst() int {
	return rl.burst
}

// Reset removes the rate limiter for the given identifier
func (rl *InMemoryRateLimiter) Reset(identifier string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.limiters, identifier)
}

// Clear removes entries that haven't been accessed recently
func (rl *InMemoryRateLimiter) Clear() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for id, entry := range rl.limiters {
		if now.Sub(entry.lastAccess) > rl.cleanupTTL {
			delete(rl.limiters, id)
		}
	}
}

// getLimiter gets or creates a rate limiter for the given identifier
func (rl *InMemoryRateLimiter) getLimiter(identifier string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, exists := rl.limiters[identifier]
	now := time.Now()

	if !exists {
		// Create new limiter for this identifier
		limiter := rate.NewLimiter(rl.limit, rl.burst)
		rl.limiters[identifier] = &limiterEntry{
			limiter:    limiter,
			lastAccess: now,
		}
		return limiter
	}

	// Update last access time
	entry.lastAccess = now
	return entry.limiter
}

// startCleanup runs periodic cleanup to remove old entries
func (rl *InMemoryRateLimiter) startCleanup() {
	ticker := time.NewTicker(rl.cleanupTTL)
	defer ticker.Stop()

	for range ticker.C {
		rl.Clear()
	}
}

// RateLimitConfig holds configuration for different rate limit tiers
type RateLimitConfig struct {
	// API Key creation limits
	APIKeyCreationLimit int           // requests per hour
	APIKeyCreationBurst int           // max burst
	APIKeyWindow        time.Duration // typically 1 hour

	// General API usage limits
	APIUsageLimit int           // requests per minute
	APIUsageBurst int           // max burst
	APIUsageWindow time.Duration // typically 1 minute
}

// DefaultRateLimitConfig returns sensible defaults
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		APIKeyCreationLimit: 10,
		APIKeyCreationBurst: 3,
		APIKeyWindow:        time.Hour,

		APIUsageLimit:  1000,
		APIUsageBurst:  100,
		APIUsageWindow: time.Minute,
	}
}

// Validate checks if the rate limit configuration is valid
func (c RateLimitConfig) Validate() error {
	if c.APIKeyCreationLimit <= 0 {
		return fmt.Errorf("APIKeyCreationLimit must be positive")
	}
	if c.APIKeyCreationBurst <= 0 {
		return fmt.Errorf("APIKeyCreationBurst must be positive")
	}
	if c.APIUsageLimit <= 0 {
		return fmt.Errorf("APIUsageLimit must be positive")
	}
	if c.APIUsageBurst <= 0 {
		return fmt.Errorf("APIUsageBurst must be positive")
	}
	if c.APIKeyWindow <= 0 {
		return fmt.Errorf("APIKeyWindow must be positive")
	}
	if c.APIUsageWindow <= 0 {
		return fmt.Errorf("APIUsageWindow must be positive")
	}
	return nil
}
