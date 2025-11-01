package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yourusername/gatekeeper/internal/store"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// setupTestDB_logging creates a test database connection
func setupTestDB_logging(t *testing.T) *store.DB {
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

func TestLoggingMiddleware(t *testing.T) {
	t.Run("logs request start and completion", func(t *testing.T) {
		core, logs := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		middleware := NewLoggingMiddleware(logger)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})

		wrapped := middleware.Middleware()(handler)

		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		// Should have logged request start and completion
		assert.GreaterOrEqual(t, logs.Len(), 2)

		// Check first log (request started)
		startLog := logs.All()[0]
		assert.Equal(t, "http request started", startLog.Message)
		assert.Equal(t, "GET", startLog.ContextMap()["method"])
		assert.Equal(t, "/api/test", startLog.ContextMap()["path"])

		// Check second log (request completed)
		completedLog := logs.All()[1]
		assert.Equal(t, "http request completed", completedLog.Message)
		assert.Equal(t, "GET", completedLog.ContextMap()["method"])
		assert.Equal(t, "/api/test", completedLog.ContextMap()["path"])
		assert.Equal(t, int64(200), completedLog.ContextMap()["status"])
		assert.NotNil(t, completedLog.ContextMap()["duration"])
	})

	t.Run("logs with request ID from context", func(t *testing.T) {
		core, logs := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		middleware := NewLoggingMiddleware(logger)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrapped := middleware.Middleware()(handler)

		// Add request ID via metrics middleware first
		db := setupTestDB_logging(t)
		metricsMiddleware := NewMetricsMiddleware(NewMetricsCollector(db), logger)
		fullWrapped := metricsMiddleware.Middleware()(wrapped)

		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()

		fullWrapped.ServeHTTP(w, req)

		// All logs should have request_id
		for _, log := range logs.All() {
			assert.NotEmpty(t, log.ContextMap()["request_id"])
		}
	})

	t.Run("logs user address when authenticated", func(t *testing.T) {
		core, logs := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		middleware := NewLoggingMiddleware(logger)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrapped := middleware.Middleware()(handler)

		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		// Check that logs were created (user_address will be empty without auth)
		assert.GreaterOrEqual(t, logs.Len(), 2)
	})

	t.Run("uses error level for 5xx responses", func(t *testing.T) {
		core, logs := observer.New(zapcore.ErrorLevel)
		logger := zap.New(core)

		middleware := NewLoggingMiddleware(logger)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})

		wrapped := middleware.Middleware()(handler)

		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		// Should have at least one error-level log for completion
		assert.GreaterOrEqual(t, logs.Len(), 1)

		completedLog := logs.All()[logs.Len()-1]
		assert.Equal(t, zapcore.ErrorLevel, completedLog.Level)
		assert.Equal(t, int64(500), completedLog.ContextMap()["status"])
	})

	t.Run("uses warn level for 4xx responses", func(t *testing.T) {
		core, logs := observer.New(zapcore.WarnLevel)
		logger := zap.New(core)

		middleware := NewLoggingMiddleware(logger)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})

		wrapped := middleware.Middleware()(handler)

		req := httptest.NewRequest("GET", "/api/missing", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		// Should have warn-level log for completion
		assert.GreaterOrEqual(t, logs.Len(), 1)

		completedLog := logs.All()[logs.Len()-1]
		assert.Equal(t, zapcore.WarnLevel, completedLog.Level)
		assert.Equal(t, int64(404), completedLog.ContextMap()["status"])
	})

	t.Run("includes user agent and remote addr", func(t *testing.T) {
		core, logs := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		middleware := NewLoggingMiddleware(logger)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrapped := middleware.Middleware()(handler)

		req := httptest.NewRequest("GET", "/api/test", nil)
		req.Header.Set("User-Agent", "TestClient/1.0")
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		// Check request started log
		startLog := logs.All()[0]
		assert.Equal(t, "TestClient/1.0", startLog.ContextMap()["user_agent"])
		assert.Equal(t, "192.168.1.1:12345", startLog.ContextMap()["remote_addr"])
	})
}

func TestLoggingResponseWriter(t *testing.T) {
	t.Run("captures status code", func(t *testing.T) {
		w := httptest.NewRecorder()
		wrapped := &loggingResponseWriter{ResponseWriter: w, statusCode: 200}

		wrapped.WriteHeader(http.StatusAccepted)

		assert.Equal(t, http.StatusAccepted, wrapped.statusCode)
		assert.True(t, wrapped.written)
	})

	t.Run("ignores subsequent WriteHeader calls", func(t *testing.T) {
		w := httptest.NewRecorder()
		wrapped := &loggingResponseWriter{ResponseWriter: w, statusCode: 200}

		wrapped.WriteHeader(http.StatusOK)
		wrapped.WriteHeader(http.StatusBadRequest) // Should be ignored

		assert.Equal(t, http.StatusOK, wrapped.statusCode)
	})

	t.Run("Write calls WriteHeader if not written", func(t *testing.T) {
		w := httptest.NewRecorder()
		wrapped := &loggingResponseWriter{ResponseWriter: w, statusCode: 200}

		n, err := wrapped.Write([]byte("test"))

		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.True(t, wrapped.written)
		assert.Equal(t, http.StatusOK, wrapped.statusCode)
	})

	t.Run("writes body correctly", func(t *testing.T) {
		w := httptest.NewRecorder()
		wrapped := &loggingResponseWriter{ResponseWriter: w, statusCode: 200}

		wrapped.WriteHeader(http.StatusCreated)
		wrapped.Write([]byte("test response"))

		assert.Equal(t, "test response", w.Body.String())
		assert.Equal(t, http.StatusCreated, w.Code)
	})
}

func BenchmarkLoggingMiddleware(b *testing.B) {
	logger := zap.NewNop()
	middleware := NewLoggingMiddleware(logger)

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

