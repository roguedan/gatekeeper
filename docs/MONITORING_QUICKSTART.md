# Monitoring Quick Start Guide

This guide will help you quickly set up and use the health check and monitoring endpoints in Gatekeeper.

## Prerequisites

- Gatekeeper service running
- (Optional) Prometheus for metrics collection
- (Optional) Grafana for visualization

## Quick Access

Once the service is running, you can immediately access:

- **Health Check**: `http://localhost:8080/health`
- **Liveness Probe**: `http://localhost:8080/health/live`
- **Readiness Probe**: `http://localhost:8080/health/ready`
- **Metrics**: `http://localhost:8080/metrics`

## Testing Health Endpoints

### 1. Basic Health Check

```bash
curl http://localhost:8080/health | jq
```

Example response:
```json
{
  "status": "ok",
  "timestamp": "2024-11-01T12:00:00Z",
  "version": "1.0.0",
  "checks": {
    "database": {
      "status": "ok",
      "responseTime": 15,
      "message": "PostgreSQL connected"
    },
    "ethereum": {
      "status": "ok",
      "responseTime": 42,
      "chainId": "0x1",
      "message": "Ethereum RPC responding"
    },
    "uptime": 3600
  }
}
```

### 2. Liveness Probe

```bash
curl http://localhost:8080/health/live
```

Example response:
```json
{
  "status": "ok"
}
```

### 3. Readiness Probe

```bash
curl http://localhost:8080/health/ready
```

Example response when ready:
```json
{
  "status": "ready"
}
```

Example response when not ready:
```json
{
  "status": "not_ready",
  "reason": "database_down"
}
```

## Viewing Metrics

### View All Metrics

```bash
curl http://localhost:8080/metrics
```

### Filter Specific Metrics

```bash
# View HTTP request metrics
curl http://localhost:8080/metrics | grep http_requests_total

# View database metrics
curl http://localhost:8080/metrics | grep db_connections

# View cache metrics
curl http://localhost:8080/metrics | grep cache
```

## Setting Up Prometheus

### 1. Create prometheus.yml

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'gatekeeper'
    static_configs:
      - targets: ['host.docker.internal:8080']  # Use 'localhost:8080' if not using Docker
    metrics_path: /metrics
```

### 2. Run Prometheus with Docker

```bash
docker run -d \
  --name prometheus \
  -p 9090:9090 \
  -v $(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus
```

### 3. Access Prometheus UI

Open http://localhost:9090 and try these queries:

**Request Rate:**
```promql
rate(http_requests_total[5m])
```

**Error Rate:**
```promql
sum(rate(http_errors_total[5m])) by (type)
```

**P95 Latency:**
```promql
http_request_duration_seconds{quantile="0.95"}
```

## Setting Up Grafana

### 1. Run Grafana with Docker

```bash
docker run -d \
  --name grafana \
  -p 3000:3000 \
  grafana/grafana
```

### 2. Configure Prometheus Data Source

1. Open http://localhost:3000
2. Login (default: admin/admin)
3. Go to Configuration > Data Sources
4. Add Prometheus data source
5. URL: http://host.docker.internal:9090 (or http://localhost:9090)
6. Click "Save & Test"

### 3. Create Dashboard

Create a new dashboard with these panels:

**Request Rate:**
```promql
sum(rate(http_requests_total[5m])) by (endpoint)
```

**Error Rate:**
```promql
sum(rate(http_errors_total[5m])) by (type)
```

**Response Time (P50, P95, P99):**
```promql
http_request_duration_seconds{quantile="0.5"}
http_request_duration_seconds{quantile="0.95"}
http_request_duration_seconds{quantile="0.99"}
```

**Database Connections:**
```promql
db_connections_in_use
db_connections_idle
db_connections_max
```

**Cache Hit Rate:**
```promql
cache_hit_rate
```

## Kubernetes Integration

### Deploy with Health Checks

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gatekeeper
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: gatekeeper
        image: gatekeeper:latest
        ports:
        - containerPort: 8080
        env:
        - name: VERSION
          value: "1.0.0"

        # Liveness probe - restart if unhealthy
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3

        # Readiness probe - remove from load balancer if not ready
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 5
          failureThreshold: 3

        # Startup probe - give time to start
        startupProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 0
          periodSeconds: 5
          timeoutSeconds: 5
          failureThreshold: 12
```

### Prometheus ServiceMonitor (for Prometheus Operator)

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: gatekeeper
  labels:
    app: gatekeeper
spec:
  selector:
    matchLabels:
      app: gatekeeper
  endpoints:
  - port: http
    path: /metrics
    interval: 15s
```

## Monitoring Best Practices

### 1. Health Check Monitoring

Set up alerts for when health checks fail:

```yaml
# Alertmanager rule
- alert: GatekeeperDown
  expr: up{job="gatekeeper"} == 0
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "Gatekeeper service is down"
```

### 2. Request Rate Monitoring

Monitor for unusual traffic patterns:

```yaml
- alert: HighRequestRate
  expr: sum(rate(http_requests_total[5m])) > 1000
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "High request rate detected"
```

### 3. Error Rate Monitoring

Alert on elevated error rates:

```yaml
- alert: HighErrorRate
  expr: sum(rate(http_errors_total[5m])) / sum(rate(http_requests_total[5m])) > 0.05
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Error rate above 5%"
```

### 4. Latency Monitoring

Monitor response times:

```yaml
- alert: SlowResponses
  expr: http_request_duration_seconds{quantile="0.95"} > 1
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "P95 latency above 1 second"
```

### 5. Database Connection Pool

Monitor connection pool exhaustion:

```yaml
- alert: DatabasePoolExhausted
  expr: db_connections_in_use / db_connections_max > 0.9
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Database connection pool near exhaustion"
```

## Viewing Logs

All HTTP requests are logged with structured JSON format. Use these queries to filter logs:

### View All Requests

```bash
# If using JSON logging
docker logs gatekeeper | grep "http request completed" | jq
```

### View Slow Requests

```bash
# Requests taking > 1 second
docker logs gatekeeper | grep "slow request detected" | jq
```

### View Errors

```bash
# 5xx errors
docker logs gatekeeper | grep "http request completed" | jq 'select(.status >= 500)'

# 4xx errors
docker logs gatekeeper | grep "http request completed" | jq 'select(.status >= 400 and .status < 500)'
```

### Follow Specific Request

```bash
# Using request_id
REQUEST_ID="550e8400-e29b-41d4-a716-446655440000"
docker logs gatekeeper | grep $REQUEST_ID | jq
```

## Troubleshooting

### Health Check Returns "degraded"

This means Ethereum RPC is down but database is healthy. Service continues to operate with limited functionality.

**Solution:** Check your Ethereum RPC endpoint configuration and connectivity.

### Health Check Returns "down"

This means the database is unavailable. Service cannot operate.

**Solution:**
1. Check database connectivity
2. Verify DATABASE_URL environment variable
3. Check database logs

### Metrics Not Appearing in Prometheus

**Solution:**
1. Verify Prometheus can reach the metrics endpoint: `curl http://localhost:8080/metrics`
2. Check Prometheus targets page: http://localhost:9090/targets
3. Verify scrape_configs in prometheus.yml

### No Logs Appearing

**Solution:**
1. Check LOG_LEVEL environment variable (should be "info" or lower)
2. Verify application is running
3. Check container logs: `docker logs gatekeeper`

## Environment Variables

Configure monitoring behavior with these environment variables:

```bash
# Service version (appears in health checks)
VERSION="1.0.0"

# Log level (debug, info, warn, error)
LOG_LEVEL="info"

# Database connection pool (affects metrics)
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME_MINUTES=5
DB_CONN_MAX_IDLE_TIME_MINUTES=1
```

## Example Dashboard Queries

### Overall Service Health

```promql
# Service uptime
up{job="gatekeeper"}

# Request success rate
sum(rate(http_requests_total{status=~"2.."}[5m])) / sum(rate(http_requests_total[5m]))
```

### Performance Metrics

```promql
# Requests per second
sum(rate(http_requests_total[5m]))

# Average response time
rate(http_request_duration_seconds_sum[5m]) / rate(http_request_duration_seconds_count[5m])

# Response time percentiles
http_request_duration_seconds{quantile="0.5"}
http_request_duration_seconds{quantile="0.95"}
http_request_duration_seconds{quantile="0.99"}
```

### Resource Utilization

```promql
# Database connection usage
(db_connections_in_use / db_connections_max) * 100

# Cache performance
cache_hit_rate * 100
```

## Next Steps

1. Set up automated alerting based on metrics
2. Create custom Grafana dashboards for your use case
3. Integrate with your logging aggregation system
4. Configure retention policies for metrics and logs
5. Set up distributed tracing (optional)

For more details, see [HEALTH_AND_MONITORING.md](./HEALTH_AND_MONITORING.md)
