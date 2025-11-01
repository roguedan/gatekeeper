package http

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/yourusername/gatekeeper/internal/store"
)

// MetricsCollector collects application metrics for Prometheus export
type MetricsCollector struct {
	mu sync.RWMutex

	// Request metrics
	requestCount      map[string]map[int]int64  // endpoint -> status_code -> count
	requestDurations  map[string][]float64      // endpoint -> durations in seconds
	errorCount        map[string]int64          // error_type -> count

	// Database metrics
	db *store.DB

	// Cache metrics
	cacheHits   int64
	cacheMisses int64
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(db *store.DB) *MetricsCollector {
	return &MetricsCollector{
		requestCount:     make(map[string]map[int]int64),
		requestDurations: make(map[string][]float64),
		errorCount:       make(map[string]int64),
		db:              db,
	}
}

// RecordRequest records a completed HTTP request
func (m *MetricsCollector) RecordRequest(endpoint string, statusCode int, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Record request count by endpoint and status code
	if m.requestCount[endpoint] == nil {
		m.requestCount[endpoint] = make(map[int]int64)
	}
	m.requestCount[endpoint][statusCode]++

	// Record duration (convert to seconds)
	durationSeconds := duration.Seconds()
	m.requestDurations[endpoint] = append(m.requestDurations[endpoint], durationSeconds)

	// Keep only last 1000 durations per endpoint to prevent unbounded memory growth
	if len(m.requestDurations[endpoint]) > 1000 {
		m.requestDurations[endpoint] = m.requestDurations[endpoint][1:]
	}
}

// RecordError records an error occurrence
func (m *MetricsCollector) RecordError(errorType string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.errorCount[errorType]++
}

// RecordCacheHit records a cache hit
func (m *MetricsCollector) RecordCacheHit() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cacheHits++
}

// RecordCacheMiss records a cache miss
func (m *MetricsCollector) RecordCacheMiss() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cacheMisses++
}

// ServeHTTP serves metrics in Prometheus text format
// GET /metrics
func (m *MetricsCollector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")

	var output strings.Builder

	// Write request count metrics
	output.WriteString("# HELP http_requests_total Total number of HTTP requests\n")
	output.WriteString("# TYPE http_requests_total counter\n")

	// Sort endpoints for consistent output
	endpoints := make([]string, 0, len(m.requestCount))
	for endpoint := range m.requestCount {
		endpoints = append(endpoints, endpoint)
	}
	sort.Strings(endpoints)

	for _, endpoint := range endpoints {
		statusCodes := m.requestCount[endpoint]
		for statusCode, count := range statusCodes {
			output.WriteString(fmt.Sprintf(
				`http_requests_total{endpoint="%s",status="%d"} %d`+"\n",
				sanitizeLabel(endpoint), statusCode, count,
			))
		}
	}

	// Write request duration percentiles
	output.WriteString("\n# HELP http_request_duration_seconds HTTP request duration in seconds\n")
	output.WriteString("# TYPE http_request_duration_seconds summary\n")

	for _, endpoint := range endpoints {
		durations := m.requestDurations[endpoint]
		if len(durations) == 0 {
			continue
		}

		// Calculate percentiles
		p50 := percentile(durations, 0.50)
		p95 := percentile(durations, 0.95)
		p99 := percentile(durations, 0.99)
		sum := sum(durations)
		count := len(durations)

		sanitizedEndpoint := sanitizeLabel(endpoint)
		output.WriteString(fmt.Sprintf(
			`http_request_duration_seconds{endpoint="%s",quantile="0.5"} %.6f`+"\n",
			sanitizedEndpoint, p50,
		))
		output.WriteString(fmt.Sprintf(
			`http_request_duration_seconds{endpoint="%s",quantile="0.95"} %.6f`+"\n",
			sanitizedEndpoint, p95,
		))
		output.WriteString(fmt.Sprintf(
			`http_request_duration_seconds{endpoint="%s",quantile="0.99"} %.6f`+"\n",
			sanitizedEndpoint, p99,
		))
		output.WriteString(fmt.Sprintf(
			`http_request_duration_seconds_sum{endpoint="%s"} %.6f`+"\n",
			sanitizedEndpoint, sum,
		))
		output.WriteString(fmt.Sprintf(
			`http_request_duration_seconds_count{endpoint="%s"} %d`+"\n",
			sanitizedEndpoint, count,
		))
	}

	// Write error count metrics
	if len(m.errorCount) > 0 {
		output.WriteString("\n# HELP http_errors_total Total number of HTTP errors\n")
		output.WriteString("# TYPE http_errors_total counter\n")

		errorTypes := make([]string, 0, len(m.errorCount))
		for errorType := range m.errorCount {
			errorTypes = append(errorTypes, errorType)
		}
		sort.Strings(errorTypes)

		for _, errorType := range errorTypes {
			count := m.errorCount[errorType]
			output.WriteString(fmt.Sprintf(
				`http_errors_total{type="%s"} %d`+"\n",
				sanitizeLabel(errorType), count,
			))
		}
	}

	// Write database connection pool metrics
	if m.db != nil {
		stats := m.db.Stats()

		output.WriteString("\n# HELP db_connections_max Maximum number of database connections\n")
		output.WriteString("# TYPE db_connections_max gauge\n")
		output.WriteString(fmt.Sprintf("db_connections_max %d\n", stats.MaxOpenConnections))

		output.WriteString("\n# HELP db_connections_open Number of open database connections\n")
		output.WriteString("# TYPE db_connections_open gauge\n")
		output.WriteString(fmt.Sprintf("db_connections_open %d\n", stats.OpenConnections))

		output.WriteString("\n# HELP db_connections_in_use Number of database connections in use\n")
		output.WriteString("# TYPE db_connections_in_use gauge\n")
		output.WriteString(fmt.Sprintf("db_connections_in_use %d\n", stats.InUse))

		output.WriteString("\n# HELP db_connections_idle Number of idle database connections\n")
		output.WriteString("# TYPE db_connections_idle gauge\n")
		output.WriteString(fmt.Sprintf("db_connections_idle %d\n", stats.Idle))
	}

	// Write cache metrics
	totalCacheRequests := m.cacheHits + m.cacheMisses
	if totalCacheRequests > 0 {
		output.WriteString("\n# HELP cache_hits_total Total number of cache hits\n")
		output.WriteString("# TYPE cache_hits_total counter\n")
		output.WriteString(fmt.Sprintf("cache_hits_total %d\n", m.cacheHits))

		output.WriteString("\n# HELP cache_misses_total Total number of cache misses\n")
		output.WriteString("# TYPE cache_misses_total counter\n")
		output.WriteString(fmt.Sprintf("cache_misses_total %d\n", m.cacheMisses))

		hitRate := float64(m.cacheHits) / float64(totalCacheRequests)
		output.WriteString("\n# HELP cache_hit_rate Cache hit rate (0-1)\n")
		output.WriteString("# TYPE cache_hit_rate gauge\n")
		output.WriteString(fmt.Sprintf("cache_hit_rate %.4f\n", hitRate))
	}

	w.Write([]byte(output.String()))
}

// percentile calculates the nth percentile of a sorted slice
func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Make a copy and sort
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	// Calculate index
	index := p * float64(len(sorted)-1)
	lower := int(index)
	upper := lower + 1

	if upper >= len(sorted) {
		return sorted[len(sorted)-1]
	}

	// Linear interpolation
	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// sum calculates the sum of all values
func sum(values []float64) float64 {
	total := 0.0
	for _, v := range values {
		total += v
	}
	return total
}

// sanitizeLabel sanitizes a label value for Prometheus
func sanitizeLabel(label string) string {
	// Escape backslashes and quotes
	label = strings.ReplaceAll(label, `\`, `\\`)
	label = strings.ReplaceAll(label, `"`, `\"`)
	return label
}

// GetCacheStats returns current cache statistics
func (m *MetricsCollector) GetCacheStats() (hits int64, misses int64, hitRate float64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	hits = m.cacheHits
	misses = m.cacheMisses

	total := hits + misses
	if total > 0 {
		hitRate = float64(hits) / float64(total)
	}

	return
}
