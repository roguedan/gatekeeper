package http

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/yourusername/gatekeeper/internal/log"
)

// RateLimitMiddleware provides rate limiting functionality for HTTP endpoints
type RateLimitMiddleware struct {
	limiter RateLimiter
	logger  *log.Logger
	// identifierFunc extracts the identifier from the request (user ID, IP, etc.)
	identifierFunc func(*http.Request) string
	// onRateLimit is called when rate limit is exceeded (for custom responses)
	onRateLimit func(http.ResponseWriter, *http.Request, string)
}

// RateLimitMiddlewareOption configures the rate limit middleware
type RateLimitMiddlewareOption func(*RateLimitMiddleware)

// WithIdentifierFunc sets a custom identifier extraction function
func WithIdentifierFunc(fn func(*http.Request) string) RateLimitMiddlewareOption {
	return func(m *RateLimitMiddleware) {
		m.identifierFunc = fn
	}
}

// WithOnRateLimit sets a custom rate limit exceeded handler
func WithOnRateLimit(fn func(http.ResponseWriter, *http.Request, string)) RateLimitMiddlewareOption {
	return func(m *RateLimitMiddleware) {
		m.onRateLimit = fn
	}
}

// NewRateLimitMiddleware creates a new rate limiting middleware
func NewRateLimitMiddleware(limiter RateLimiter, logger *log.Logger, opts ...RateLimitMiddlewareOption) *RateLimitMiddleware {
	m := &RateLimitMiddleware{
		limiter:        limiter,
		logger:         logger,
		identifierFunc: defaultIdentifier,
		onRateLimit:    defaultRateLimitResponse,
	}

	// Apply options
	for _, opt := range opts {
		opt(m)
	}

	return m
}

// Middleware returns the HTTP middleware function
func (m *RateLimitMiddleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract identifier (user ID or IP address)
			identifier := m.identifierFunc(r)

			// Check rate limit
			if !m.limiter.Allow(identifier) {
				// Log rate limit violation
				m.logger.Warn(fmt.Sprintf("Rate limit exceeded for %s on %s %s",
					identifier, r.Method, r.URL.Path))

				// Call custom or default rate limit handler
				m.onRateLimit(w, r, identifier)
				return
			}

			// Allow request to proceed
			next.ServeHTTP(w, r)
		})
	}
}

// MiddlewareFunc returns a Gorilla mux compatible middleware function
func (m *RateLimitMiddleware) MiddlewareFunc() func(http.Handler) http.Handler {
	return m.Middleware()
}

// defaultIdentifier extracts an identifier from the request
// Priority: User ID from JWT claims > IP address
func defaultIdentifier(r *http.Request) string {
	// Try to get user ID from JWT claims (if authenticated)
	claims := ClaimsFromContext(r)
	if claims != nil && claims.Address != "" {
		return "user:" + claims.Address
	}

	// Fallback to IP address
	ip := extractIP(r)
	return "ip:" + ip
}

// UserIdentifier creates an identifier extraction function that only uses user ID
// Falls back to IP if no user is authenticated
func UserIdentifier(r *http.Request) string {
	claims := ClaimsFromContext(r)
	if claims != nil && claims.Address != "" {
		return "user:" + claims.Address
	}
	// Still fallback to IP for unauthenticated requests
	return "ip:" + extractIP(r)
}

// IPIdentifier creates an identifier extraction function that only uses IP address
func IPIdentifier(r *http.Request) string {
	return "ip:" + extractIP(r)
}

// extractIP extracts the client IP address from the request
// Handles X-Forwarded-For and X-Real-IP headers for proxied requests
func extractIP(r *http.Request) string {
	// Check X-Forwarded-For header (comma-separated list, first is client)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// Take the first IP in the list
		if ip, _, err := net.SplitHostPort(forwarded); err == nil {
			return ip
		}
		// If no port, return as-is
		if ip := net.ParseIP(forwarded); ip != nil {
			return forwarded
		}
		// Multiple IPs, take first
		for idx := 0; idx < len(forwarded); idx++ {
			if forwarded[idx] == ',' {
				return forwarded[:idx]
			}
		}
		return forwarded
	}

	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		if ip := net.ParseIP(realIP); ip != nil {
			return realIP
		}
	}

	// Fall back to RemoteAddr
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return ip
	}

	return r.RemoteAddr
}

// defaultRateLimitResponse sends a 429 Too Many Requests response
func defaultRateLimitResponse(w http.ResponseWriter, r *http.Request, identifier string) {
	// Calculate retry-after based on rate limiter settings
	// For simplicity, we use a fixed 60 seconds, but this could be calculated
	// based on when the next token will be available
	retryAfter := 60

	// Set headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
	w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%.0f", float64(1))) // Simplified
	w.Header().Set("X-RateLimit-Remaining", "0")
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Unix()+int64(retryAfter), 10))

	// Send response
	w.WriteHeader(http.StatusTooManyRequests)
	fmt.Fprintf(w, `{"error":"Rate limit exceeded","message":"Too many requests. Please try again later.","retryAfter":%d}`, retryAfter)
}

// NewUserRateLimitMiddleware creates a middleware that rate limits by user ID
func NewUserRateLimitMiddleware(limiter RateLimiter, logger *log.Logger) *RateLimitMiddleware {
	return NewRateLimitMiddleware(limiter, logger, WithIdentifierFunc(UserIdentifier))
}

// NewIPRateLimitMiddleware creates a middleware that rate limits by IP address
func NewIPRateLimitMiddleware(limiter RateLimiter, logger *log.Logger) *RateLimitMiddleware {
	return NewRateLimitMiddleware(limiter, logger, WithIdentifierFunc(IPIdentifier))
}

// PerUserRateLimitMiddleware creates a middleware specifically for per-user rate limiting
// This is useful for authenticated endpoints like API key creation
func PerUserRateLimitMiddleware(requestsPerWindow int, window time.Duration, burst int, logger *log.Logger) func(http.Handler) http.Handler {
	limiter := NewInMemoryRateLimiter(requestsPerWindow, window, burst)
	middleware := NewUserRateLimitMiddleware(limiter, logger)
	return middleware.Middleware()
}

// PerIPRateLimitMiddleware creates a middleware specifically for per-IP rate limiting
// This is useful for public endpoints or fallback rate limiting
func PerIPRateLimitMiddleware(requestsPerWindow int, window time.Duration, burst int, logger *log.Logger) func(http.Handler) http.Handler {
	limiter := NewInMemoryRateLimiter(requestsPerWindow, window, burst)
	middleware := NewIPRateLimitMiddleware(limiter, logger)
	return middleware.Middleware()
}
