# Health Check and Monitoring Implementation Summary

This document summarizes the comprehensive health check and monitoring system implemented for the Gatekeeper backend service.

## Overview

A complete observability solution has been added to the Gatekeeper service, including:
- Health check endpoints for operational visibility
- Kubernetes-ready liveness and readiness probes
- Prometheus-compatible metrics collection
- Structured request logging with correlation IDs
- Comprehensive test coverage

## Files Created/Modified

### New Files Created

#### Health Check Handlers
- **`/internal/http/handlers/health.go`**
  - Implements comprehensive health checking
  - Provides detailed status for database and Ethereum RPC
  - Tracks service uptime and response times
  - Exports: `HealthHandler`, `Health()`, `Live()`, `Ready()`

- **`/internal/http/handlers/health_test.go`**
  - Comprehensive test coverage for health endpoints
  - Tests database and Ethereum connectivity checks
  - Validates response format and status codes
  - Tests context cancellation and timeout handling

#### Metrics Collection
- **`/internal/http/metrics.go`**
  - Prometheus-compatible metrics collector
  - Tracks HTTP requests, errors, durations
  - Monitors database connection pool
  - Tracks cache hit/miss rates
  - Exports: `MetricsCollector`, `ServeHTTP()`

- **`/internal/http/metrics_test.go`**
  - Tests metrics collection and aggregation
  - Validates Prometheus output format
  - Tests concurrency safety
  - Includes performance benchmarks

#### Middleware
- **`/internal/http/metrics_middleware.go`**
  - Middleware for automatic metrics collection
  - Generates unique request IDs (UUIDs)
  - Tracks request duration and status codes
  - Logs slow requests (>1 second)
  - Exports: `MetricsMiddleware`, `RequestIDFromContext()`

- **`/internal/http/metrics_middleware_test.go`**
  - Tests request tracking and metrics recording
  - Validates request ID generation and propagation
  - Tests error categorization

- **`/internal/http/logging_middleware.go`**
  - Structured JSON logging for all HTTP requests
  - Logs request start and completion
  - Includes request ID, user address, timing
  - Different log levels for errors (info/warn/error)
  - Exports: `LoggingMiddleware`

- **`/internal/http/logging_middleware_test.go`**
  - Tests logging behavior and output
  - Validates structured log format
  - Tests log level selection based on status codes

#### Documentation
- **`/docs/HEALTH_AND_MONITORING.md`**
  - Complete reference documentation
  - Detailed API specifications
  - Prometheus query examples
  - Kubernetes configuration examples
  - Alerting rule examples

- **`/docs/MONITORING_QUICKSTART.md`**
  - Quick start guide for developers
  - Step-by-step setup instructions
  - Docker Compose examples
  - Grafana dashboard queries
  - Troubleshooting guide

- **`/docs/HEALTH_MONITORING_IMPLEMENTATION.md`** (this file)
  - Implementation summary
  - Technical details
  - Integration guide

### Modified Files

#### Configuration
- **`/internal/config/config.go`**
  - Added `Version` field to Config struct
  - Loads VERSION environment variable (defaults to "dev")

- **`/go.mod`**
  - Added `github.com/google/uuid v1.6.0` dependency for request ID generation

#### Main Application
- **`/cmd/server/main.go`**
  - Initialized `MetricsCollector` with database connection
  - Initialized `HealthHandler` with database, provider, logger, and version
  - Created `MetricsMiddleware` and `LoggingMiddleware`
  - Applied global middleware to all routes (logging -> metrics)
  - Registered health endpoints: `/health`, `/health/live`, `/health/ready`
  - Registered metrics endpoint: `/metrics`
  - Removed old inline health check handler

## Endpoints Implemented

### 1. GET /health
**Comprehensive Health Check**

Returns detailed health information about all service dependencies.

**Response:** 200 OK (healthy/degraded) or 503 Service Unavailable (down)

```json
{
  "status": "ok|degraded|down",
  "timestamp": "2024-11-01T12:00:00Z",
  "version": "1.0.0",
  "checks": {
    "database": {
      "status": "ok|down",
      "responseTime": 15,
      "message": "PostgreSQL connected"
    },
    "ethereum": {
      "status": "ok|down",
      "responseTime": 42,
      "chainId": "0x1",
      "message": "Ethereum RPC responding"
    },
    "uptime": 3600
  }
}
```

### 2. GET /health/live
**Kubernetes Liveness Probe**

Simple check to verify the service process is running.

**Response:** 200 OK

```json
{
  "status": "ok"
}
```

### 3. GET /health/ready
**Kubernetes Readiness Probe**

Checks if the service is ready to accept traffic (all dependencies healthy).

**Response:** 200 OK (ready) or 503 Service Unavailable (not ready)

```json
{
  "status": "ready"
}
```

Or when not ready:

```json
{
  "status": "not_ready",
  "reason": "database_down|ethereum_rpc_down"
}
```

### 4. GET /metrics
**Prometheus Metrics Endpoint**

Returns metrics in Prometheus text format.

**Content-Type:** `text/plain; version=0.0.4`

**Metrics Included:**
- `http_requests_total` - Total HTTP requests by endpoint and status
- `http_request_duration_seconds` - Request duration percentiles (p50, p95, p99)
- `http_errors_total` - Total errors by type
- `db_connections_max` - Maximum database connections
- `db_connections_open` - Current open connections
- `db_connections_in_use` - Connections currently in use
- `db_connections_idle` - Idle connections
- `cache_hits_total` - Total cache hits
- `cache_misses_total` - Total cache misses
- `cache_hit_rate` - Cache hit rate (0-1)

## Middleware Chain

The application applies middleware in the following order:

1. **LoggingMiddleware** - Logs all requests with structured data
2. **MetricsMiddleware** - Collects metrics and generates request IDs
3. **APIKeyMiddleware** - Authenticates API keys (for protected routes)
4. **JWTMiddleware** - Authenticates JWT tokens (for protected routes)
5. **RateLimitMiddleware** - Enforces rate limits

## Request Flow

For every HTTP request:

1. **LoggingMiddleware** logs request start with:
   - Request ID (from MetricsMiddleware)
   - HTTP method and path
   - Remote address and user agent
   - User address (if authenticated)

2. **MetricsMiddleware**:
   - Generates unique UUID request ID
   - Adds request ID to context
   - Measures request duration
   - Records metrics (count, duration, errors)
   - Logs slow requests (>1 second)

3. **Application Handler** processes the request

4. **LoggingMiddleware** logs request completion with:
   - Request ID
   - HTTP status code
   - Response duration (milliseconds)
   - User address (if authenticated)

5. **MetricsMiddleware** finalizes metrics:
   - Updates request count by endpoint and status
   - Records duration for percentile calculations
   - Records errors if status >= 400

## Health Check Behavior

### Database Check
- Executes `SELECT 1` query with 5-second timeout
- Measures response time in milliseconds
- Returns "ok" on success, "down" on failure
- Critical dependency - affects overall status

### Ethereum RPC Check
- Calls `eth_chainId` with 5-second timeout
- Measures response time in milliseconds
- Returns chain ID on success
- Non-critical dependency - affects degraded status only
- Only runs if provider is configured

### Overall Status Logic
- **ok**: All checks pass
- **degraded**: Ethereum down but database up (service still functional)
- **down**: Database down (service cannot operate)

## Metrics Collection

### Request Metrics
- **Count**: Tracked per endpoint and status code
- **Duration**: Stored for last 1000 requests per endpoint
- **Percentiles**: Calculated on-demand (p50, p95, p99)

### Error Metrics
- **Types**: bad_request, unauthorized, forbidden, not_found, rate_limit_exceeded, client_error, server_error
- **Count**: Tracked per error type

### Database Metrics
- **Real-time**: Queries actual database pool statistics
- **Updates**: On each /metrics request

### Cache Metrics
- **Hits/Misses**: Cumulative counters
- **Hit Rate**: Calculated as hits / (hits + misses)

## Logging Format

All logs use structured JSON format with zap logger:

```json
{
  "level": "info",
  "ts": 1698854400.123,
  "msg": "http request completed",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "method": "GET",
  "path": "/api/data",
  "status": 200,
  "duration": "111ms",
  "duration_ms": 111,
  "user_address": "0x1234567890abcdef",
  "remote_addr": "192.168.1.100:54321"
}
```

### Log Levels
- **info**: Successful requests (2xx)
- **warn**: Client errors (4xx) and slow requests (>1s)
- **error**: Server errors (5xx)

## Testing

### Test Coverage
- **Health Handler**: 8 test cases covering all endpoints and edge cases
- **Metrics Collector**: 11 test cases plus concurrency and benchmark tests
- **Metrics Middleware**: 7 test cases covering all middleware functionality
- **Logging Middleware**: 6 test cases covering log output and levels

### Running Tests

```bash
# Run all tests
go test ./internal/http/handlers/... ./internal/http/... -v

# Run with database (requires TEST_DATABASE_URL)
export TEST_DATABASE_URL="postgres://user:pass@localhost:5432/test_db?sslmode=disable"
go test ./internal/http/handlers/... ./internal/http/... -v

# Run benchmarks
go test ./internal/http/... -bench=. -benchmem
```

## Environment Variables

### New Variables
- **VERSION**: Service version (default: "dev")
  - Appears in health check responses
  - Used for deployment tracking

### Existing Variables (used by monitoring)
- **LOG_LEVEL**: Log level (default: "info")
- **DB_MAX_OPEN_CONNS**: Max database connections (default: 25)
- **DB_MAX_IDLE_CONNS**: Max idle connections (default: 5)
- **DATABASE_URL**: PostgreSQL connection string (required)
- **ETHEREUM_RPC**: Ethereum RPC endpoint (required)

## Integration Examples

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gatekeeper
spec:
  template:
    spec:
      containers:
      - name: gatekeeper
        image: gatekeeper:1.0.0
        env:
        - name: VERSION
          value: "1.0.0"

        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10

        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### Prometheus Scrape Config

```yaml
scrape_configs:
  - job_name: 'gatekeeper'
    static_configs:
      - targets: ['gatekeeper:8080']
    metrics_path: /metrics
    scrape_interval: 15s
```

### Alert Rules

```yaml
groups:
  - name: gatekeeper
    rules:
      - alert: GatekeeperDown
        expr: up{job="gatekeeper"} == 0
        for: 1m

      - alert: HighErrorRate
        expr: sum(rate(http_errors_total[5m])) > 10
        for: 5m

      - alert: SlowResponses
        expr: http_request_duration_seconds{quantile="0.95"} > 1
        for: 5m
```

## Performance Considerations

### Memory
- Request duration history limited to 1000 samples per endpoint
- Metrics stored in memory (suitable for moderate traffic)
- Structured logging uses efficient zap logger

### CPU
- Middleware overhead: ~20-30μs per request
- Metrics calculation: O(n log n) for percentiles (on /metrics endpoint only)
- Concurrent-safe with read/write mutexes

### Scalability
- Tested with concurrent requests
- Thread-safe metrics collection
- Efficient memory usage with bounded history

## Best Practices

1. **Health Checks**
   - Monitor `/health` for operational visibility
   - Use `/health/live` for Kubernetes liveness probes
   - Use `/health/ready` for Kubernetes readiness probes
   - Set reasonable timeout values (5-10 seconds)

2. **Metrics**
   - Scrape `/metrics` every 15-30 seconds
   - Set up alerts for critical metrics
   - Monitor trends, not just absolute values
   - Use Grafana for visualization

3. **Logging**
   - Use request IDs to correlate logs
   - Set appropriate log levels
   - Monitor slow requests
   - Aggregate logs with ELK or similar

4. **Alerting**
   - Alert on rate of change, not absolutes
   - Set meaningful thresholds
   - Include runbook links in annotations
   - Test alert rules regularly

## Future Enhancements

Potential improvements:
1. Distributed tracing with OpenTelemetry
2. Custom metrics exporters (InfluxDB, StatsD)
3. Request sampling for high-traffic scenarios
4. Metric persistence for historical analysis
5. Auto-scaling based on metrics
6. Circuit breaker integration with health checks

## Support and Documentation

- **Quick Start**: See `/docs/MONITORING_QUICKSTART.md`
- **API Reference**: See `/docs/HEALTH_AND_MONITORING.md`
- **Implementation**: See this document

For questions or issues, refer to the inline code documentation and test files for examples.
