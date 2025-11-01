# Docker Deployment Guide - Gatekeeper

Complete guide for deploying Gatekeeper using Docker and Docker Compose.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Building Images](#building-images)
- [Running Services](#running-services)
- [Health Checks](#health-checks)
- [Logs and Monitoring](#logs-and-monitoring)
- [Scaling](#scaling)
- [Production Deployment](#production-deployment)
- [Troubleshooting](#troubleshooting)
- [Security Checklist](#security-checklist)

---

## Prerequisites

### Required Software

- **Docker**: Version 20.10+ ([Install Docker](https://docs.docker.com/get-docker/))
- **Docker Compose**: Version 2.0+ ([Install Docker Compose](https://docs.docker.com/compose/install/))
- **Git**: For cloning the repository

### Verify Installation

```bash
docker --version
# Docker version 24.0.0+

docker compose version
# Docker Compose version v2.20.0+
```

### System Requirements

**Minimum:**
- 2 CPU cores
- 4GB RAM
- 10GB disk space

**Recommended:**
- 4 CPU cores
- 8GB RAM
- 20GB disk space (for logs and database)

---

## Quick Start

### 1. Clone Repository

```bash
git clone https://github.com/yourusername/gatekeeper.git
cd gatekeeper
```

### 2. Configure Environment

```bash
# Copy environment template
cp .env.example .env

# Edit configuration (see Configuration section)
nano .env  # or use your preferred editor
```

**Critical: Update these values in `.env`:**

```bash
# Generate a strong JWT secret
JWT_SECRET=$(openssl rand -base64 32)

# Set a secure database password
POSTGRES_PASSWORD=your-secure-password-here

# Add your Ethereum RPC endpoint
ETHEREUM_RPC=https://eth-mainnet.g.alchemy.com/v2/YOUR_API_KEY
```

### 3. Start Services

```bash
# Start all services (backend, frontend, postgres, redis)
docker compose up -d

# View logs
docker compose logs -f

# Check service health
docker compose ps
```

### 4. Verify Deployment

**Backend Health Check:**
```bash
curl http://localhost:8080/health
# Expected: {"status":"ok","port":"8080"}
```

**Frontend Access:**
```bash
# Open browser to: http://localhost:3000
```

**Database Connection:**
```bash
docker compose exec postgres psql -U gatekeeper -d gatekeeper -c "SELECT version();"
```

---

## Configuration

### Environment Variables

All configuration is managed through `.env` file. See `.env.example` for complete reference.

#### Essential Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `PORT` | Backend API port | `8080` | Yes |
| `DATABASE_URL` | PostgreSQL connection string | Auto-generated | Yes |
| `JWT_SECRET` | Secret key for JWT signing | None | **Yes** |
| `ETHEREUM_RPC` | Ethereum RPC endpoint | None | **Yes** |
| `POSTGRES_PASSWORD` | Database password | `devpassword` | **Yes** |

#### Blockchain Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `CHAIN_ID` | Ethereum chain ID (1=mainnet, 11155111=sepolia) | `1` |
| `ETHEREUM_RPC_FALLBACK` | Fallback RPC endpoint | None |
| `CACHE_TTL` | Blockchain cache TTL (seconds) | `300` |
| `RPC_TIMEOUT` | RPC call timeout (seconds) | `5` |

#### Rate Limiting

| Variable | Description | Default |
|----------|-------------|---------|
| `API_KEY_CREATION_RATE_LIMIT` | Max API keys per user/hour | `10` |
| `API_USAGE_RATE_LIMIT` | Max API requests per user/minute | `1000` |

#### Frontend Configuration

| Variable | Description | Required |
|----------|-------------|----------|
| `VITE_API_URL` | Backend API URL | Yes |
| `VITE_WALLETCONNECT_PROJECT_ID` | WalletConnect project ID | Recommended |
| `VITE_CHAIN_ID` | Chain ID for frontend | Yes |

---

## Building Images

### Build All Images

```bash
docker compose build
```

### Build Individual Services

```bash
# Backend only
docker compose build backend

# Frontend only
docker compose build frontend
```

### Build with No Cache (fresh build)

```bash
docker compose build --no-cache
```

### Expected Build Times

- **Backend**: 2-3 minutes (first build), 30s (cached)
- **Frontend**: 4-5 minutes (first build), 1 minute (cached)

### Image Sizes

- **Backend**: ~100MB (alpine-based)
- **Frontend**: ~80MB (nginx-based)
- **PostgreSQL**: ~240MB (official alpine image)
- **Redis**: ~35MB (alpine image)

---

## Running Services

### Start All Services

```bash
# Detached mode (background)
docker compose up -d

# Foreground with logs
docker compose up
```

### Start Specific Services

```bash
# Only database
docker compose up -d postgres

# Backend + dependencies
docker compose up -d backend
```

### Stop Services

```bash
# Stop all services (preserves data)
docker compose stop

# Stop and remove containers (preserves volumes)
docker compose down

# Stop and remove everything including volumes (DESTROYS DATA)
docker compose down -v
```

### Restart Services

```bash
# Restart all
docker compose restart

# Restart specific service
docker compose restart backend
```

---

## Health Checks

### Check Service Status

```bash
docker compose ps
```

Expected output:
```
NAME                    STATUS              PORTS
gatekeeper-backend      Up (healthy)        0.0.0.0:8080->8080/tcp
gatekeeper-frontend     Up (healthy)        0.0.0.0:3000->3000/tcp
gatekeeper-postgres     Up (healthy)        0.0.0.0:5432->5432/tcp
gatekeeper-redis        Up (healthy)        0.0.0.0:6379->6379/tcp
```

### Health Check Endpoints

**Backend:**
```bash
curl http://localhost:8080/health
```

**Frontend:**
```bash
curl http://localhost:3000/health
```

**PostgreSQL:**
```bash
docker compose exec postgres pg_isready -U gatekeeper
```

**Redis:**
```bash
docker compose exec redis redis-cli ping
```

### Health Check Timing

All services include health checks with these parameters:

- **Interval**: 30s (check every 30 seconds)
- **Timeout**: 3s (fail if check takes >3s)
- **Retries**: 3 (mark unhealthy after 3 failures)
- **Start Period**: 5-10s (grace period on startup)

**Expected startup time**: All services healthy within 30 seconds.

---

## Logs and Monitoring

### View Logs

```bash
# All services (follow mode)
docker compose logs -f

# Specific service
docker compose logs -f backend

# Last 100 lines
docker compose logs --tail=100

# Since specific time
docker compose logs --since 2024-01-01T10:00:00
```

### Log Levels

Configure via `LOG_LEVEL` environment variable:

- `debug` - Verbose logging (development)
- `info` - Standard logging (default)
- `warn` - Warnings only
- `error` - Errors only

### Container Stats

```bash
# Real-time resource usage
docker stats

# Specific container
docker stats gatekeeper-backend
```

### Inspect Service Configuration

```bash
# View service configuration
docker compose config

# View specific service
docker compose config backend
```

---

## Scaling

### Horizontal Scaling

**Scale backend instances:**
```bash
docker compose up -d --scale backend=3
```

**Note**: Requires load balancer configuration (nginx, HAProxy, etc.)

### Resource Limits

Edit `docker-compose.yml` to adjust resource limits:

```yaml
services:
  backend:
    deploy:
      resources:
        limits:
          cpus: '2.0'        # Increase CPU limit
          memory: 1G         # Increase memory limit
        reservations:
          cpus: '1.0'
          memory: 512M
```

### Database Scaling

**Increase connection pool:**
```bash
# In .env
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=10
```

**Monitor database connections:**
```bash
docker compose exec postgres psql -U gatekeeper -d gatekeeper -c \
  "SELECT count(*) FROM pg_stat_activity;"
```

---

## Production Deployment

### Production Checklist

- [ ] Use strong random `JWT_SECRET` (min 32 bytes)
- [ ] Set secure `POSTGRES_PASSWORD`
- [ ] Use HTTPS/TLS for all external connections
- [ ] Enable database SSL (`sslmode=require`)
- [ ] Configure firewall rules (restrict database/redis access)
- [ ] Set up reverse proxy (nginx/Caddy) with SSL
- [ ] Configure CORS for your domain
- [ ] Set up log aggregation (ELK, Datadog, etc.)
- [ ] Configure automated backups (database volumes)
- [ ] Set up monitoring and alerting
- [ ] Use Docker secrets or vault for sensitive data
- [ ] Enable resource limits for all services
- [ ] Set `LOG_LEVEL=info` or `LOG_LEVEL=warn`
- [ ] Review and harden nginx configuration
- [ ] Implement rate limiting at reverse proxy level

### Production Environment File

Create `.env.production`:

```bash
# Server
PORT=8080
ENV=production
LOG_LEVEL=info

# Database (use managed service in production)
DATABASE_URL=postgresql://user:pass@db.example.com:5432/gatekeeper?sslmode=require

# Security
JWT_SECRET=<use-vault-or-secrets-manager>
POSTGRES_PASSWORD=<use-vault-or-secrets-manager>

# Blockchain
ETHEREUM_RPC=https://eth-mainnet.g.alchemy.com/v2/PRODUCTION_KEY
ETHEREUM_RPC_FALLBACK=https://mainnet.infura.io/v3/FALLBACK_KEY
CHAIN_ID=1

# Frontend
VITE_API_URL=https://api.yourdomain.com
VITE_WALLETCONNECT_PROJECT_ID=<your-project-id>

# CORS
CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://app.yourdomain.com
```

### Use Production Environment

```bash
docker compose --env-file .env.production up -d
```

### Reverse Proxy Example (nginx)

```nginx
server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;

    ssl_certificate /etc/ssl/certs/yourdomain.crt;
    ssl_certificate_key /etc/ssl/private/yourdomain.key;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Database Backups

**Create backup:**
```bash
docker compose exec postgres pg_dump -U gatekeeper gatekeeper > backup.sql
```

**Restore backup:**
```bash
cat backup.sql | docker compose exec -T postgres psql -U gatekeeper gatekeeper
```

**Automated daily backups:**
```bash
# Add to crontab
0 2 * * * cd /path/to/gatekeeper && docker compose exec -T postgres \
  pg_dump -U gatekeeper gatekeeper | gzip > backups/backup-$(date +\%Y\%m\%d).sql.gz
```

---

## Troubleshooting

### Services Won't Start

**Check logs:**
```bash
docker compose logs
```

**Common issues:**

1. **Port already in use:**
   ```bash
   # Find process using port
   lsof -i :8080

   # Change port in .env
   BACKEND_PORT=8081
   ```

2. **Database connection fails:**
   ```bash
   # Check database is healthy
   docker compose ps postgres

   # Check connection string
   docker compose exec backend env | grep DATABASE_URL
   ```

3. **Missing environment variables:**
   ```bash
   # Verify .env exists
   ls -la .env

   # Check required variables
   grep JWT_SECRET .env
   ```

### Service Unhealthy

**Check health status:**
```bash
docker compose ps
```

**Inspect health check logs:**
```bash
docker inspect gatekeeper-backend | grep -A 10 Health
```

**Common fixes:**

- Increase `start_period` in health check
- Check service logs for errors
- Verify port bindings are correct

### Database Migration Errors

**Check migration status:**
```bash
docker compose exec postgres psql -U gatekeeper -d gatekeeper -c "\dt"
```

**Manually run migrations:**
```bash
# Migrations run automatically on container start
# To re-run, copy SQL files to init directory

docker compose down
docker volume rm gatekeeper-postgres-data
docker compose up -d
```

### Out of Memory

**Check memory usage:**
```bash
docker stats
```

**Increase Docker memory:**
- Docker Desktop: Settings → Resources → Memory
- Linux: Edit `/etc/docker/daemon.json`

### Cannot Connect to Ethereum RPC

**Test RPC endpoint:**
```bash
curl -X POST $ETHEREUM_RPC \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'
```

**Check backend logs:**
```bash
docker compose logs backend | grep -i rpc
```

### Frontend Can't Connect to Backend

**Verify backend is healthy:**
```bash
curl http://localhost:8080/health
```

**Check CORS configuration:**
```bash
# Frontend should use correct API URL
docker compose exec frontend env | grep VITE_API_URL
```

**Test from frontend container:**
```bash
docker compose exec frontend curl http://backend:8080/health
```

### Clean Restart

**Remove everything and start fresh:**
```bash
# Stop all services
docker compose down

# Remove all data (CAUTION: destroys database)
docker compose down -v

# Remove images
docker compose down --rmi all

# Rebuild and start
docker compose build --no-cache
docker compose up -d
```

---

## Security Checklist

### Pre-Deployment

- [ ] Review `.env` - no secrets committed to git
- [ ] Strong `JWT_SECRET` (32+ bytes, random)
- [ ] Secure `POSTGRES_PASSWORD`
- [ ] All images use non-root users
- [ ] Health checks configured
- [ ] Resource limits set
- [ ] `.dockerignore` excludes sensitive files

### Network Security

- [ ] Firewall rules restrict database access
- [ ] Redis not exposed to public internet
- [ ] HTTPS/TLS for all external traffic
- [ ] Database connections use SSL
- [ ] CORS configured for specific origins only

### Operational Security

- [ ] Regular security updates (rebuild images)
- [ ] Automated backups configured
- [ ] Log monitoring and alerting
- [ ] Secret rotation policy
- [ ] Incident response plan
- [ ] Regular vulnerability scanning

### Docker Security

```bash
# Scan images for vulnerabilities
docker scan gatekeeper-backend

# Check for updates
docker compose pull

# Remove unused images/volumes
docker system prune -a
```

---

## Support

### Getting Help

- **Documentation**: See `README.md`, `API.md`
- **Issues**: Open GitHub issue with logs and configuration
- **Logs**: Always include relevant logs when reporting issues

### Useful Commands Reference

```bash
# View all containers
docker compose ps

# View logs
docker compose logs -f <service>

# Restart service
docker compose restart <service>

# Execute command in container
docker compose exec <service> <command>

# View resource usage
docker stats

# Clean up
docker compose down
docker system prune
```

---

**Last Updated**: November 2024
**Version**: 1.0.0
**Gatekeeper**: Wallet-Native Authentication Gateway
