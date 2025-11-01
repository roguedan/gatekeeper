package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/gatekeeper/internal/store"
	"go.uber.org/zap"
)

// setupTestDB creates a test database connection
func setupTestDB_middleware(t *testing.T) *store.DB {
	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://gatekeeper:change-this-secure-password@localhost:5432/gatekeeper_test?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolCfg := store.PoolConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}

	db, err := store.Connect(ctx, dbURL, poolCfg)
	if err != nil {
		t.Skipf("skipping test: failed to connect to test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func TestMetricsMiddleware(t *testing.T) {
	db := setupTestDB_middleware(t)

	logger := zap.NewNop()
	collector := NewMetricsCollector(db)
	middleware := NewMetricsMiddleware(collector, logger)

	t.Run("records successful request", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})

		wrapped := middleware.Middleware()(handler)

		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Check metrics were recorded
		collector.mu.RLock()
		defer collector.mu.RUnlock()

		assert.Equal(t, int64(1), collector.requestCount["GET /api/test"][200])
		assert.Len(t, collector.requestDurations["GET /api/test"], 1)
	})

	t.Run("records error request", func(t *testing.T) {
		collector := NewMetricsCollector(db)
		middleware := NewMetricsMiddleware(collector, logger)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("not found"))
		})

		wrapped := middleware.Middleware()(handler)

		req := httptest.NewRequest("GET", "/api/missing", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		// Check error was recorded
		collector.mu.RLock()
		defer collector.mu.RUnlock()

		assert.Equal(t, int64(1), collector.errorCount["not_found"])
	})

	t.Run("adds request ID to context", func(t *testing.T) {
		var capturedRequestID string

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedRequestID = RequestIDFromContext(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		wrapped := middleware.Middleware()(handler)

		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		assert.NotEmpty(t, capturedRequestID)
		assert.Len(t, capturedRequestID, 36) // UUID length
	})

	t.Run("measures request duration", func(t *testing.T) {
		collector := NewMetricsCollector(db)
		middleware := NewMetricsMiddleware(collector, logger)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(50 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		})

		wrapped := middleware.Middleware()(handler)

		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()

		start := time.Now()
		wrapped.ServeHTTP(w, req)
		elapsed := time.Since(start)

		assert.GreaterOrEqual(t, elapsed, 50*time.Millisecond)

		// Check duration was recorded
		collector.mu.RLock()
		defer collector.mu.RUnlock()

		require.Len(t, collector.requestDurations["GET /api/test"], 1)
		assert.GreaterOrEqual(t, collector.requestDurations["GET /api/test"][0], 0.05) // At least 50ms
	})

	t.Run("captures status code from WriteHeader", func(t *testing.T) {
		collector := NewMetricsCollector(db)
		middleware := NewMetricsMiddleware(collector, logger)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("created"))
		})

		wrapped := middleware.Middleware()(handler)

		req := httptest.NewRequest("POST", "/api/test", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		collector.mu.RLock()
		defer collector.mu.RUnlock()

		assert.Equal(t, int64(1), collector.requestCount["POST /api/test"][201])
	})

	t.Run("defaults to 200 if WriteHeader not called", func(t *testing.T) {
		collector := NewMetricsCollector(db)
		middleware := NewMetricsMiddleware(collector, logger)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Don't call WriteHeader, just write body
			w.Write([]byte("ok"))
		})

		wrapped := middleware.Middleware()(handler)

		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		collector.mu.RLock()
		defer collector.mu.RUnlock()

		assert.Equal(t, int64(1), collector.requestCount["GET /api/test"][200])
	})
}

func TestResponseWriter(t *testing.T) {
	t.Run("captures status code", func(t *testing.T) {
		w := httptest.NewRecorder()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

		wrapped.WriteHeader(http.StatusCreated)

		assert.Equal(t, http.StatusCreated, wrapped.statusCode)
		assert.True(t, wrapped.written)
	})

	t.Run("ignores subsequent WriteHeader calls", func(t *testing.T) {
		w := httptest.NewRecorder()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

		wrapped.WriteHeader(http.StatusOK)
		wrapped.WriteHeader(http.StatusInternalServerError) // Should be ignored

		assert.Equal(t, http.StatusOK, wrapped.statusCode)
	})

	t.Run("Write calls WriteHeader if not written", func(t *testing.T) {
		w := httptest.NewRecorder()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

		wrapped.Write([]byte("test"))

		assert.True(t, wrapped.written)
		assert.Equal(t, http.StatusOK, wrapped.statusCode)
	})
}

func TestNormalizeEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		expected string
	}{
		{
			name:     "simple path",
			method:   "GET",
			path:     "/api/test",
			expected: "GET /api/test",
		},
		{
			name:     "removes trailing slash",
			method:   "GET",
			path:     "/api/test/",
			expected: "GET /api/test",
		},
		{
			name:     "preserves root path",
			method:   "GET",
			path:     "/",
			expected: "GET /",
		},
		{
			name:     "different methods",
			method:   "POST",
			path:     "/api/test",
			expected: "POST /api/test",
		},
		{
			name:     "nested path",
			method:   "DELETE",
			path:     "/api/users/123/posts/456",
			expected: "DELETE /api/users/123/posts/456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeEndpoint(tt.method, tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetErrorType(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   string
	}{
		{400, "bad_request"},
		{401, "unauthorized"},
		{403, "forbidden"},
		{404, "not_found"},
		{429, "rate_limit_exceeded"},
		{405, "client_error"},
		{499, "client_error"},
		{500, "server_error"},
		{502, "server_error"},
		{503, "server_error"},
		{599, "server_error"},
		{200, "unknown_error"}, // Edge case
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.statusCode)), func(t *testing.T) {
			result := getErrorType(tt.statusCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRequestIDFromContext(t *testing.T) {
	t.Run("retrieves request ID", func(t *testing.T) {
		requestID := "test-request-id-123"
		ctx := context.WithValue(context.Background(), requestIDKey, requestID)

		result := RequestIDFromContext(ctx)
		assert.Equal(t, requestID, result)
	})

	t.Run("returns empty string if not set", func(t *testing.T) {
		ctx := context.Background()

		result := RequestIDFromContext(ctx)
		assert.Equal(t, "", result)
	})

	t.Run("returns empty string if wrong type", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), requestIDKey, 123)

		result := RequestIDFromContext(ctx)
		assert.Equal(t, "", result)
	})
}

func BenchmarkMetricsMiddleware(b *testing.B) {
	db := setupTestDB_middleware(&testing.T{})

	logger := zap.NewNop()
	collector := NewMetricsCollector(db)
	middleware := NewMetricsMiddleware(collector, logger)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.Middleware()(handler)

	req := httptest.NewRequest("GET", "/api/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		wrapped.ServeHTTP(w, req)
	}
}
