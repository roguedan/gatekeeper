package http

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// LoggingMiddleware logs HTTP requests with structured context
type LoggingMiddleware struct {
	logger *zap.Logger
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(logger *zap.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// Middleware returns the logging middleware
func (m *LoggingMiddleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Get request ID from context (set by metrics middleware)
			requestID := RequestIDFromContext(r.Context())

			// Get user address from claims if authenticated
			userAddress := ""
			if claims := ClaimsFromContext(r); claims != nil {
				userAddress = claims.Address
			}

			// Wrap response writer to capture status
			wrapped := &loggingResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Log request start
			m.logger.Info("http request started",
				zap.String("request_id", requestID),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
				zap.String("user_address", userAddress),
			)

			// Call next handler
			next.ServeHTTP(wrapped, r)

			// Calculate duration
			duration := time.Since(start)

			// Log request completion
			logFunc := m.logger.Info
			if wrapped.statusCode >= 500 {
				logFunc = m.logger.Error
			} else if wrapped.statusCode >= 400 {
				logFunc = m.logger.Warn
			}

			logFunc("http request completed",
				zap.String("request_id", requestID),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", wrapped.statusCode),
				zap.Duration("duration", duration),
				zap.Int64("duration_ms", duration.Milliseconds()),
				zap.String("user_address", userAddress),
				zap.String("remote_addr", r.RemoteAddr),
			)
		})
	}
}

// loggingResponseWriter wraps http.ResponseWriter to capture status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// WriteHeader captures the status code
func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	if !w.written {
		w.statusCode = statusCode
		w.written = true
		w.ResponseWriter.WriteHeader(statusCode)
	}
}

// Write ensures WriteHeader is called
func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}
