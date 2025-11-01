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
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *store.DB {
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

func TestMetricsCollector_RecordRequest(t *testing.T) {
	db := setupTestDB(t)

	collector := NewMetricsCollector(db)

	t.Run("records request count", func(t *testing.T) {
		collector.RecordRequest("GET /api/test", 200, 100*time.Millisecond)
		collector.RecordRequest("GET /api/test", 200, 150*time.Millisecond)
		collector.RecordRequest("GET /api/test", 404, 50*time.Millisecond)

		collector.mu.RLock()
		defer collector.mu.RUnlock()

		assert.Equal(t, int64(2), collector.requestCount["GET /api/test"][200])
		assert.Equal(t, int64(1), collector.requestCount["GET /api/test"][404])
	})

	t.Run("records request durations", func(t *testing.T) {
		collector := NewMetricsCollector(db)

		collector.RecordRequest("GET /api/test", 200, 100*time.Millisecond)
		collector.RecordRequest("GET /api/test", 200, 200*time.Millisecond)

		collector.mu.RLock()
		defer collector.mu.RUnlock()

		assert.Len(t, collector.requestDurations["GET /api/test"], 2)
		assert.InDelta(t, 0.1, collector.requestDurations["GET /api/test"][0], 0.01)
		assert.InDelta(t, 0.2, collector.requestDurations["GET /api/test"][1], 0.01)
	})

	t.Run("limits duration history", func(t *testing.T) {
		collector := NewMetricsCollector(db)

		// Record more than 1000 requests
		for i := 0; i < 1100; i++ {
			collector.RecordRequest("GET /api/test", 200, 100*time.Millisecond)
		}

		collector.mu.RLock()
		defer collector.mu.RUnlock()

		// Should keep only last 1000
		assert.Len(t, collector.requestDurations["GET /api/test"], 1000)
	})
}

func TestMetricsCollector_RecordError(t *testing.T) {
	db := setupTestDB(t)

	collector := NewMetricsCollector(db)

	collector.RecordError("not_found")
	collector.RecordError("not_found")
	collector.RecordError("server_error")

	collector.mu.RLock()
	defer collector.mu.RUnlock()

	assert.Equal(t, int64(2), collector.errorCount["not_found"])
	assert.Equal(t, int64(1), collector.errorCount["server_error"])
}

func TestMetricsCollector_RecordCache(t *testing.T) {
	db := setupTestDB(t)

	collector := NewMetricsCollector(db)

	t.Run("records cache hits", func(t *testing.T) {
		collector.RecordCacheHit()
		collector.RecordCacheHit()

		collector.mu.RLock()
		defer collector.mu.RUnlock()

		assert.Equal(t, int64(2), collector.cacheHits)
	})

	t.Run("records cache misses", func(t *testing.T) {
		collector.RecordCacheMiss()

		collector.mu.RLock()
		defer collector.mu.RUnlock()

		assert.Equal(t, int64(1), collector.cacheMisses)
	})
}

func TestMetricsCollector_ServeHTTP(t *testing.T) {
	db := setupTestDB(t)

	collector := NewMetricsCollector(db)

	t.Run("returns prometheus format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()

		collector.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/plain; version=0.0.4", w.Header().Get("Content-Type"))

		body := w.Body.String()
		assert.Contains(t, body, "# HELP http_requests_total")
		assert.Contains(t, body, "# TYPE http_requests_total counter")
	})

	t.Run("includes request counts", func(t *testing.T) {
		collector.RecordRequest("GET /api/test", 200, 100*time.Millisecond)
		collector.RecordRequest("GET /api/test", 404, 50*time.Millisecond)

		req := httptest.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()

		collector.ServeHTTP(w, req)

		body := w.Body.String()
		assert.Contains(t, body, `http_requests_total{endpoint="GET /api/test",status="200"} 1`)
		assert.Contains(t, body, `http_requests_total{endpoint="GET /api/test",status="404"} 1`)
	})

	t.Run("includes duration percentiles", func(t *testing.T) {
		collector := NewMetricsCollector(db)

		// Record various durations
		for i := 0; i < 100; i++ {
			collector.RecordRequest("GET /api/test", 200, time.Duration(i)*time.Millisecond)
		}

		req := httptest.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()

		collector.ServeHTTP(w, req)

		body := w.Body.String()
		assert.Contains(t, body, "# HELP http_request_duration_seconds")
		assert.Contains(t, body, `quantile="0.5"`)
		assert.Contains(t, body, `quantile="0.95"`)
		assert.Contains(t, body, `quantile="0.99"`)
		assert.Contains(t, body, `http_request_duration_seconds_count`)
		assert.Contains(t, body, `http_request_duration_seconds_sum`)
	})

	t.Run("includes error counts", func(t *testing.T) {
		collector.RecordError("not_found")
		collector.RecordError("server_error")

		req := httptest.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()

		collector.ServeHTTP(w, req)

		body := w.Body.String()
		assert.Contains(t, body, "# HELP http_errors_total")
		assert.Contains(t, body, `http_errors_total{type="not_found"}`)
		assert.Contains(t, body, `http_errors_total{type="server_error"}`)
	})

	t.Run("includes database pool stats", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()

		collector.ServeHTTP(w, req)

		body := w.Body.String()
		assert.Contains(t, body, "# HELP db_connections_max")
		assert.Contains(t, body, "# HELP db_connections_open")
		assert.Contains(t, body, "# HELP db_connections_in_use")
		assert.Contains(t, body, "# HELP db_connections_idle")
	})

	t.Run("includes cache metrics", func(t *testing.T) {
		collector.RecordCacheHit()
		collector.RecordCacheHit()
		collector.RecordCacheMiss()

		req := httptest.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()

		collector.ServeHTTP(w, req)

		body := w.Body.String()
		assert.Contains(t, body, "# HELP cache_hits_total")
		assert.Contains(t, body, "cache_hits_total 2")
		assert.Contains(t, body, "# HELP cache_misses_total")
		assert.Contains(t, body, "cache_misses_total 1")
		assert.Contains(t, body, "# HELP cache_hit_rate")
		assert.Contains(t, body, "cache_hit_rate")
	})

	t.Run("sanitizes labels", func(t *testing.T) {
		// Test label with quotes and backslashes
		collector.RecordRequest(`GET /api/"test"\path`, 200, 100*time.Millisecond)

		req := httptest.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()

		collector.ServeHTTP(w, req)

		body := w.Body.String()
		// Labels should be escaped
		assert.Contains(t, body, `endpoint="GET /api/\"test\"\\path"`)
	})
}

func TestMetricsCollector_GetCacheStats(t *testing.T) {
	db := setupTestDB(t)

	collector := NewMetricsCollector(db)

	t.Run("calculates hit rate", func(t *testing.T) {
		collector.RecordCacheHit()
		collector.RecordCacheHit()
		collector.RecordCacheHit()
		collector.RecordCacheMiss()

		hits, misses, hitRate := collector.GetCacheStats()

		assert.Equal(t, int64(3), hits)
		assert.Equal(t, int64(1), misses)
		assert.InDelta(t, 0.75, hitRate, 0.01) // 3/4 = 0.75
	})

	t.Run("handles no data", func(t *testing.T) {
		hits, misses, hitRate := collector.GetCacheStats()

		assert.Equal(t, int64(0), hits)
		assert.Equal(t, int64(0), misses)
		assert.Equal(t, 0.0, hitRate)
	})
}

func TestPercentile(t *testing.T) {
	t.Run("calculates median", func(t *testing.T) {
		values := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
		p50 := percentile(values, 0.50)
		assert.InDelta(t, 3.0, p50, 0.1)
	})

	t.Run("calculates 95th percentile", func(t *testing.T) {
		values := make([]float64, 100)
		for i := 0; i < 100; i++ {
			values[i] = float64(i)
		}
		p95 := percentile(values, 0.95)
		assert.InDelta(t, 94.05, p95, 1.0)
	})

	t.Run("handles empty slice", func(t *testing.T) {
		values := []float64{}
		result := percentile(values, 0.50)
		assert.Equal(t, 0.0, result)
	})

	t.Run("handles single value", func(t *testing.T) {
		values := []float64{42.0}
		result := percentile(values, 0.95)
		assert.Equal(t, 42.0, result)
	})
}

func TestSum(t *testing.T) {
	t.Run("sums values", func(t *testing.T) {
		values := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
		result := sum(values)
		assert.Equal(t, 15.0, result)
	})

	t.Run("handles empty slice", func(t *testing.T) {
		values := []float64{}
		result := sum(values)
		assert.Equal(t, 0.0, result)
	})
}

func TestSanitizeLabel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no special characters",
			input:    "simple label",
			expected: "simple label",
		},
		{
			name:     "escapes quotes",
			input:    `label with "quotes"`,
			expected: `label with \"quotes\"`,
		},
		{
			name:     "escapes backslashes",
			input:    `path\to\file`,
			expected: `path\\to\\file`,
		},
		{
			name:     "escapes both",
			input:    `path\"with"quotes`,
			expected: `path\\\"with\"quotes`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeLabel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMetricsCollector_Concurrency(t *testing.T) {
	db := setupTestDB(t)

	collector := NewMetricsCollector(db)

	t.Run("handles concurrent writes", func(t *testing.T) {
		done := make(chan bool, 10)

		// Simulate concurrent metric recording
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 100; j++ {
					collector.RecordRequest("GET /api/test", 200, 100*time.Millisecond)
					collector.RecordError("test_error")
					collector.RecordCacheHit()
					collector.RecordCacheMiss()
				}
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify counts
		collector.mu.RLock()
		assert.Equal(t, int64(1000), collector.requestCount["GET /api/test"][200])
		assert.Equal(t, int64(1000), collector.errorCount["test_error"])
		assert.Equal(t, int64(1000), collector.cacheHits)
		assert.Equal(t, int64(1000), collector.cacheMisses)
		collector.mu.RUnlock()
	})

	t.Run("handles concurrent reads during writes", func(t *testing.T) {
		done := make(chan bool, 20)

		// Writers
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 50; j++ {
					collector.RecordRequest("GET /api/test", 200, 100*time.Millisecond)
				}
				done <- true
			}()
		}

		// Readers
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 50; j++ {
					req := httptest.NewRequest("GET", "/metrics", nil)
					w := httptest.NewRecorder()
					collector.ServeHTTP(w, req)
					require.Equal(t, http.StatusOK, w.Code)
				}
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 20; i++ {
			<-done
		}
	})
}

func BenchmarkMetricsCollector_RecordRequest(b *testing.B) {
	db := setupTestDB(&testing.T{})

	collector := NewMetricsCollector(db)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.RecordRequest("GET /api/test", 200, 100*time.Millisecond)
	}
}

func BenchmarkMetricsCollector_ServeHTTP(b *testing.B) {
	db := setupTestDB(&testing.T{})

	collector := NewMetricsCollector(db)

	// Populate with some data
	for i := 0; i < 100; i++ {
		collector.RecordRequest("GET /api/test", 200, 100*time.Millisecond)
	}

	req := httptest.NewRequest("GET", "/metrics", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		collector.ServeHTTP(w, req)
	}
}
