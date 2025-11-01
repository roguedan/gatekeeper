package http

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	requestIDKey contextKey = "request_id"
)

// MetricsMiddleware tracks HTTP request metrics
type MetricsMiddleware struct {
	collector *MetricsCollector
	logger    *zap.Logger
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware(collector *MetricsCollector, logger *zap.Logger) *MetricsMiddleware {
	return &MetricsMiddleware{
		collector: collector,
		logger:    logger,
	}
}

// Middleware returns the metrics collection middleware
func (m *MetricsMiddleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Generate request ID
			requestID := uuid.New().String()
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			r = r.WithContext(ctx)

			// Wrap response writer to capture status code
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // Default to 200
			}

			// Call next handler
			next.ServeHTTP(wrapped, r)

			// Calculate duration
			duration := time.Since(start)

			// Normalize endpoint for metrics (remove path parameters)
			endpoint := normalizeEndpoint(r.Method, r.URL.Path)

			// Record metrics
			m.collector.RecordRequest(endpoint, wrapped.statusCode, duration)

			// Log slow requests (> 1 second)
			if duration > time.Second {
				m.logger.Warn("slow request detected",
					zap.String("request_id", requestID),
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.Int("status", wrapped.statusCode),
					zap.Duration("duration", duration),
					zap.String("remote_addr", r.RemoteAddr),
				)
			}

			// Record error if status code indicates error
			if wrapped.statusCode >= 400 {
				errorType := getErrorType(wrapped.statusCode)
				m.collector.RecordError(errorType)
			}
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// WriteHeader captures the status code
func (w *responseWriter) WriteHeader(statusCode int) {
	if !w.written {
		w.statusCode = statusCode
		w.written = true
		w.ResponseWriter.WriteHeader(statusCode)
	}
}

// Write ensures WriteHeader is called
func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

// normalizeEndpoint normalizes endpoint paths for consistent metrics
// Removes path parameters and query strings
func normalizeEndpoint(method, path string) string {
	// Simple normalization - in production, you might want to use router patterns
	// For now, we'll keep the path as-is but could enhance to replace UUIDs, IDs, etc.

	// Common patterns to normalize
	normalized := path

	// Remove trailing slashes
	if len(normalized) > 1 && normalized[len(normalized)-1] == '/' {
		normalized = normalized[:len(normalized)-1]
	}

	// Combine with method for unique endpoint identification
	return method + " " + normalized
}

// getErrorType categorizes HTTP errors
func getErrorType(statusCode int) string {
	switch {
	case statusCode >= 400 && statusCode < 500:
		switch statusCode {
		case 400:
			return "bad_request"
		case 401:
			return "unauthorized"
		case 403:
			return "forbidden"
		case 404:
			return "not_found"
		case 429:
			return "rate_limit_exceeded"
		default:
			return "client_error"
		}
	case statusCode >= 500:
		return "server_error"
	default:
		return "unknown_error"
	}
}

// RequestIDFromContext extracts the request ID from context
func RequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return ""
}
