# Health Checks and Monitoring

This document describes the health check and monitoring endpoints available in the Gatekeeper service.

## Health Check Endpoints

### GET /health

Comprehensive health check that verifies all service dependencies and returns detailed status information.

**Authentication:** None required

**Response Format:**

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

**Status Codes:**

- `200 OK` - All systems operational or degraded but serving traffic
- `503 Service Unavailable` - Critical dependencies down

**Overall Status:**

- `ok` - All dependencies healthy
- `degraded` - Some non-critical dependencies down (e.g., Ethereum RPC)
- `down` - Critical dependencies down (e.g., database)

**Component Checks:**

1. **Database** (Critical)
   - Executes `SELECT 1` query with 5-second timeout
   - Measures response time in milliseconds
   - Returns PostgreSQL version information

2. **Ethereum** (Non-Critical)
   - Calls `eth_chainId` RPC method with 5-second timeout
   - Measures response time in milliseconds
   - Returns chain ID (e.g., "0x1" for mainnet)
   - Only present if Ethereum provider is configured

3. **Uptime**
   - Service uptime in seconds since startup

### GET /health/live

Kubernetes liveness probe - checks if the service process is running.

**Authentication:** None required

**Response:**

```json
{
  "status": "ok"
}
```

**Status Codes:**

- `200 OK` - Service is running
- `503 Service Unavailable` - Service is not responding

**Behavior:**

- Does NOT check external dependencies
- Simple check: is the HTTP server responding?
- Used by Kubernetes to determine if pod should be restarted

### GET /health/ready

Kubernetes readiness probe - checks if the service is ready to accept traffic.

**Authentication:** None required

**Response (Ready):**

```json
{
  "status": "ready"
}
```

**Response (Not Ready):**

```json
{
  "status": "not_ready",
  "reason": "database_down|ethereum_rpc_down"
}
```

**Status Codes:**

- `200 OK` - Service is ready to serve traffic
- `503 Service Unavailable` - Service is not ready

**Behavior:**

- Checks database connectivity (critical)
- Checks Ethereum RPC connectivity (critical, if configured)
- Used by Kubernetes to determine if pod should receive traffic

## Metrics Endpoint

### GET /metrics

Prometheus-compatible metrics endpoint exposing service metrics.

**Authentication:** None required

**Format:** Prometheus text format (`text/plain; version=0.0.4`)

**Available Metrics:**

#### HTTP Request Metrics

**http_requests_total** (counter)
```
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{endpoint="GET /api/data",status="200"} 1234
http_requests_total{endpoint="GET /api/data",status="404"} 56
http_requests_total{endpoint="POST /api/keys",status="201"} 789
```

**http_request_duration_seconds** (summary)
```
# HELP http_request_duration_seconds HTTP request duration in seconds
# TYPE http_request_duration_seconds summary
http_request_duration_seconds{endpoint="GET /api/data",quantile="0.5"} 0.023
http_request_duration_seconds{endpoint="GET /api/data",quantile="0.95"} 0.156
http_request_duration_seconds{endpoint="GET /api/data",quantile="0.99"} 0.342
http_request_duration_seconds_sum{endpoint="GET /api/data"} 45.678
http_request_duration_seconds_count{endpoint="GET /api/data"} 1234
```

#### Error Metrics

**http_errors_total** (counter)
```
# HELP http_errors_total Total number of HTTP errors
# TYPE http_errors_total counter
http_errors_total{type="not_found"} 56
http_errors_total{type="unauthorized"} 23
http_errors_total{type="rate_limit_exceeded"} 12
http_errors_total{type="server_error"} 3
```

**Error Types:**
- `bad_request` - HTTP 400
- `unauthorized` - HTTP 401
- `forbidden` - HTTP 403
- `not_found` - HTTP 404
- `rate_limit_exceeded` - HTTP 429
- `client_error` - Other 4xx errors
- `server_error` - 5xx errors

#### Database Metrics

**db_connections_max** (gauge)
```
# HELP db_connections_max Maximum number of database connections
# TYPE db_connections_max gauge
db_connections_max 25
```

**db_connections_open** (gauge)
```
# HELP db_connections_open Number of open database connections
# TYPE db_connections_open gauge
db_connections_open 8
```

**db_connections_in_use** (gauge)
```
# HELP db_connections_in_use Number of database connections in use
# TYPE db_connections_in_use gauge
db_connections_in_use 3
```

**db_connections_idle** (gauge)
```
# HELP db_connections_idle Number of idle database connections
# TYPE db_connections_idle gauge
db_connections_idle 5
```

#### Cache Metrics

**cache_hits_total** (counter)
```
# HELP cache_hits_total Total number of cache hits
# TYPE cache_hits_total counter
cache_hits_total 4567
```

**cache_misses_total** (counter)
```
# HELP cache_misses_total Total number of cache misses
# TYPE cache_misses_total counter
cache_misses_total 1234
```

**cache_hit_rate** (gauge)
```
# HELP cache_hit_rate Cache hit rate (0-1)
# TYPE cache_hit_rate gauge
cache_hit_rate 0.7872
```

## Request Logging

All HTTP requests are logged with structured JSON format using zap logger.

### Log Format

**Request Started:**
```json
{
  "level": "info",
  "ts": 1698854400.123,
  "msg": "http request started",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "method": "GET",
  "path": "/api/data",
  "remote_addr": "192.168.1.100:54321",
  "user_agent": "Mozilla/5.0...",
  "user_address": "0x1234567890abcdef"
}
```

**Request Completed:**
```json
{
  "level": "info",
  "ts": 1698854400.234,
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

- `info` - Successful requests (2xx status codes)
- `warn` - Client errors (4xx status codes)
- `error` - Server errors (5xx status codes)

### Slow Request Detection

Requests taking longer than 1 second are logged with `warn` level:

```json
{
  "level": "warn",
  "ts": 1698854400.234,
  "msg": "slow request detected",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "method": "GET",
  "path": "/api/data",
  "status": 200,
  "duration": "1.234s",
  "remote_addr": "192.168.1.100:54321"
}
```

## Request ID

Every request is assigned a unique UUID request ID that:

- Is added to the request context
- Appears in all log entries for that request
- Can be retrieved using `RequestIDFromContext(ctx)`
- Enables correlation across logs and metrics

## Kubernetes Configuration

### Liveness Probe

```yaml
livenessProbe:
  httpGet:
    path: /health/live
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3
```

### Readiness Probe

```yaml
readinessProbe:
  httpGet:
    path: /health/ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 5
  failureThreshold: 3
```

### Startup Probe

```yaml
startupProbe:
  httpGet:
    path: /health/ready
    port: 8080
  initialDelaySeconds: 0
  periodSeconds: 5
  timeoutSeconds: 5
  failureThreshold: 12
```

## Prometheus Configuration

### Scrape Config

```yaml
scrape_configs:
  - job_name: 'gatekeeper'
    static_configs:
      - targets: ['gatekeeper:8080']
    metrics_path: /metrics
    scrape_interval: 15s
    scrape_timeout: 10s
```

### Example Queries

**Request Rate (per second):**
```promql
rate(http_requests_total[5m])
```

**Request Rate by Endpoint:**
```promql
sum(rate(http_requests_total[5m])) by (endpoint)
```

**Error Rate:**
```promql
sum(rate(http_errors_total[5m])) by (type)
```

**P95 Latency:**
```promql
http_request_duration_seconds{quantile="0.95"}
```

**Database Connection Pool Usage:**
```promql
db_connections_in_use / db_connections_max
```

**Cache Hit Rate:**
```promql
cache_hit_rate
```

## Alerting Rules

### Example Prometheus Alerts

```yaml
groups:
  - name: gatekeeper
    rules:
      - alert: HighErrorRate
        expr: sum(rate(http_errors_total[5m])) > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"

      - alert: ServiceDown
        expr: up{job="gatekeeper"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Gatekeeper service is down"

      - alert: SlowRequests
        expr: http_request_duration_seconds{quantile="0.95"} > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "95th percentile latency exceeds 1 second"

      - alert: DatabaseConnectionPoolExhausted
        expr: db_connections_in_use / db_connections_max > 0.9
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Database connection pool near exhaustion"
```

## Configuration

### Environment Variables

- `VERSION` - Service version string (default: "dev")
- `LOG_LEVEL` - Log level: debug, info, warn, error (default: "info")
- `DB_MAX_OPEN_CONNS` - Maximum database connections (default: 25)
- `DB_MAX_IDLE_CONNS` - Maximum idle connections (default: 5)

## Best Practices

1. **Health Checks**
   - Monitor `/health` for overall system health
   - Use `/health/live` for Kubernetes liveness
   - Use `/health/ready` for Kubernetes readiness
   - Check regularly (every 5-10 seconds)

2. **Metrics**
   - Scrape `/metrics` every 15-30 seconds
   - Set up alerts for error rates, latency, and availability
   - Monitor database connection pool usage
   - Track cache hit rates for optimization

3. **Logging**
   - Use request IDs to correlate logs
   - Set appropriate log levels (info/warn/error)
   - Monitor slow requests (>1s)
   - Include user context when available

4. **Performance**
   - Health checks timeout after 5 seconds
   - Metrics collection has minimal overhead
   - Request logging is structured and efficient
   - Connection pooling optimizes database access
