# Gatekeeper Deployment Guide

Complete guide to deploying and configuring Gatekeeper in production.

## Prerequisites

- Go 1.20 or higher
- Access to Ethereum RPC endpoint (Infura, Alchemy, or self-hosted)
- Environment variables configured

## Environment Configuration

### Required Variables

```bash
# HTTP Server
PORT=8080                              # Port for HTTP server (default: 8080)

# Authentication & Authorization
JWT_SECRET=your-very-long-secret-key   # Secret for signing JWT tokens (min 32 characters)
                                       # Must be kept secure in production

# Blockchain
ETHEREUM_RPC=https://eth-mainnet.alchemyapi.io/v2/YOUR_API_KEY
                                       # Ethereum RPC endpoint
                                       # Use Infura: https://mainnet.infura.io/v3/YOUR_KEY
                                       # Use Alchemy: https://eth-mainnet.alchemyapi.io/v2/YOUR_KEY
                                       # Use your own node: http://localhost:8545

# Logging & Monitoring
LOG_LEVEL=info                         # Log level: debug, info, warn, error
                                       # Use 'debug' for development
                                       # Use 'info' for production

# Optional Configuration
NONCE_TTL_MINUTES=10                   # How long nonce is valid (default: 10 minutes)
JWT_EXPIRY_HOURS=1                     # JWT token expiration (default: 1 hour)
DATABASE_URL=postgresql://...          # Database URL (for future enhancements)
```

### Secure Secret Generation

Generate a secure JWT_SECRET:

```bash
# Using openssl
openssl rand -hex 32

# Using Go
go run -c 'import("crypto/rand", "encoding/hex", "fmt"); b := make([]byte, 32); rand.Read(b); fmt.Println(hex.EncodeToString(b))'

# Using Python
python3 -c "import secrets; print(secrets.token_hex(32))"
```

### Example .env File

```bash
# Server Configuration
PORT=8080
LOG_LEVEL=info

# Authentication
JWT_SECRET=abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789

# Blockchain
ETHEREUM_RPC=https://eth-mainnet.alchemyapi.io/v2/YOUR_ALCHEMY_KEY

# Nonce & Token Settings
NONCE_TTL_MINUTES=10
JWT_EXPIRY_HOURS=1
```

## Policy Configuration

Create a `policies.json` file to define access control rules:

```json
[
  {
    "path": "/api/health",
    "method": "GET",
    "logic": "AND",
    "rules": []
  },
  {
    "path": "/api/data",
    "method": "GET",
    "logic": "AND",
    "rules": [
      {
        "type": "has_scope",
        "scope": "read"
      }
    ]
  },
  {
    "path": "/api/admin",
    "method": "POST",
    "logic": "AND",
    "rules": [
      {
        "type": "has_scope",
        "scope": "admin"
      }
    ]
  }
]
```

### Policy Validation

Policies are validated on startup. Invalid configurations will cause the server to fail with a clear error message. Common issues:

- Missing required fields (path, method, logic, rules)
- Invalid logic operator (must be "AND" or "OR")
- Invalid rule types
- Missing rule parameters

## Local Development

### Quick Start

1. **Set environment variables:**
   ```bash
   export PORT=8080
   export JWT_SECRET=$(openssl rand -hex 32)
   export ETHEREUM_RPC=https://eth-sepolia.g.alchemy.com/v2/YOUR_KEY
   export LOG_LEVEL=debug
   ```

2. **Run tests:**
   ```bash
   go test ./internal/... -v
   ```

3. **Build the server:**
   ```bash
   go build -o gatekeeper ./cmd/server
   ```

4. **Run the server:**
   ```bash
   ./gatekeeper
   ```

5. **Test authentication:**
   ```bash
   # Get nonce
   curl http://localhost:8080/auth/siwe/nonce

   # Verify SIWE (with signed message)
   curl -X POST http://localhost:8080/auth/siwe/verify \
     -H "Content-Type: application/json" \
     -d '{"message": "...", "signature": "..."}'
   ```

## Docker Deployment

### Dockerfile

```dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o gatekeeper ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/gatekeeper .
COPY --from=builder /app/policies.json .

EXPOSE 8080
CMD ["./gatekeeper"]
```

### Docker Compose

```yaml
version: '3.8'

services:
  gatekeeper:
    build: .
    ports:
      - "8080:8080"
    environment:
      PORT: 8080
      JWT_SECRET: ${JWT_SECRET}
      ETHEREUM_RPC: ${ETHEREUM_RPC}
      LOG_LEVEL: info
      NONCE_TTL_MINUTES: 10
      JWT_EXPIRY_HOURS: 1
    volumes:
      - ./policies.json:/root/policies.json
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

### Running with Docker

```bash
# Build image
docker build -t gatekeeper:latest .

# Run container
docker run -d \
  --name gatekeeper \
  -p 8080:8080 \
  -e JWT_SECRET=$(openssl rand -hex 32) \
  -e ETHEREUM_RPC=https://eth-mainnet.alchemyapi.io/v2/YOUR_KEY \
  -e LOG_LEVEL=info \
  -v $(pwd)/policies.json:/root/policies.json \
  gatekeeper:latest

# View logs
docker logs -f gatekeeper

# Stop container
docker stop gatekeeper
```

## Kubernetes Deployment

### ConfigMap for Policies

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: gatekeeper-policies
data:
  policies.json: |
    [
      {
        "path": "/api/data",
        "method": "GET",
        "logic": "AND",
        "rules": []
      }
    ]
```

### Secrets for Sensitive Data

```bash
# Create secrets
kubectl create secret generic gatekeeper-secrets \
  --from-literal=jwt-secret=$(openssl rand -hex 32) \
  --from-literal=ethereum-rpc=https://eth-mainnet.alchemyapi.io/v2/YOUR_KEY
```

### Deployment Manifest

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gatekeeper
spec:
  replicas: 3
  selector:
    matchLabels:
      app: gatekeeper
  template:
    metadata:
      labels:
        app: gatekeeper
    spec:
      containers:
      - name: gatekeeper
        image: gatekeeper:latest
        ports:
        - containerPort: 8080
        env:
        - name: PORT
          value: "8080"
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: gatekeeper-secrets
              key: jwt-secret
        - name: ETHEREUM_RPC
          valueFrom:
            secretKeyRef:
              name: gatekeeper-secrets
              key: ethereum-rpc
        - name: LOG_LEVEL
          value: "info"
        volumeMounts:
        - name: policies
          mountPath: /root/policies.json
          subPath: policies.json
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
      volumes:
      - name: policies
        configMap:
          name: gatekeeper-policies
```

### Service and Ingress

```yaml
apiVersion: v1
kind: Service
metadata:
  name: gatekeeper-service
spec:
  selector:
    app: gatekeeper
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: ClusterIP

---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: gatekeeper-ingress
spec:
  rules:
  - host: auth.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: gatekeeper-service
            port:
              number: 80
```

## Production Checklist

- [ ] Environment variables are secure (not in version control)
- [ ] JWT_SECRET is strong (minimum 32 characters)
- [ ] HTTPS is enabled (use reverse proxy like nginx)
- [ ] CORS is configured appropriately
- [ ] Rate limiting is implemented
- [ ] Logging is aggregated (ELK, Datadog, etc.)
- [ ] Monitoring and alerting are in place
- [ ] Database backups are configured (if using DB)
- [ ] Security headers are set (HSTS, CSP, etc.)
- [ ] SIWE messages validate domain and URI
- [ ] Policy files are validated before deployment
- [ ] RPC endpoints have fallbacks
- [ ] Health checks are configured
- [ ] Graceful shutdown is implemented

## Monitoring

### Metrics to Track

1. **Authentication**
   - Nonces generated per minute
   - SIWE verification success/failure rate
   - JWT token verification success/failure rate

2. **Authorization**
   - Policy evaluations per minute
   - Allow/deny decisions by policy
   - Policy evaluation latency

3. **Blockchain**
   - RPC call latency
   - RPC call failure rate
   - Cache hit rate

4. **System**
   - Request latency (p50, p95, p99)
   - Error rate by status code
   - Memory usage
   - CPU usage

### Log Aggregation

Example ELK Stack configuration:

```yaml
# Filebeat config to collect logs
filebeat.inputs:
- type: log
  enabled: true
  paths:
    - /var/log/gatekeeper/*.log
  json.message_key: msg
  json.keys_under_root: true
  json.add_error_key: true

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
```

## Troubleshooting

### Common Issues

**Issue: "Invalid nonce" error**
- Ensure nonce is from the same server instance
- Check nonce hasn't expired (default 10 minutes)
- Verify nonce is included in SIWE message exactly as received

**Issue: "Invalid signature" error**
- Ensure SIWE message is signed exactly as created
- Verify signature is valid EIP-191 personal_sign format
- Check wallet is connected to correct network

**Issue: Policy evaluation timeout**
- Check Ethereum RPC endpoint is responsive
- Verify RPC endpoint has sufficient rate limits
- Consider adding fallback RPC endpoint

**Issue: High memory usage**
- Check policy file isn't excessively large
- Monitor cache size (cache TTL configuration)
- Consider horizontal scaling

## Scaling Considerations

### Horizontal Scaling

- Deploy multiple instances behind load balancer
- Share policies from central configuration service
- Use distributed cache (Redis) for blockchain query results
- Monitor RPC endpoint rate limits with multiple instances

### Vertical Scaling

- Increase memory for larger policy sets
- Increase CPU for high-frequency validation
- Optimize blockchain query caching

### Database (Future)

When persistent storage is needed:
- Use PostgreSQL or similar
- Implement connection pooling
- Monitor query performance
- Regular backups

## Security Best Practices

1. **Secrets Management**
   - Use environment variables or secrets manager
   - Rotate JWT_SECRET regularly
   - Use strong, random secrets
   - Never commit secrets to version control

2. **Network Security**
   - Use HTTPS in production
   - Implement rate limiting
   - Add WAF (Web Application Firewall)
   - Restrict RPC endpoint access

3. **Access Control**
   - Review policy configurations regularly
   - Audit policy changes
   - Monitor access logs
   - Test policies before deployment

4. **Blockchain Security**
   - Use trusted RPC endpoints
   - Verify contract addresses
   - Monitor for contract upgrades
   - Implement fallback mechanisms

## Support & Maintenance

- Monitor application logs for errors
- Review performance metrics regularly
- Update dependencies and security patches
- Test policy updates before production deployment
- Keep RPC endpoint configurations current
