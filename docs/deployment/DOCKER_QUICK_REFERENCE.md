# Docker Quick Reference - Gatekeeper

Fast reference for common Docker operations with Gatekeeper.

---

## Initial Setup (First Time)

```bash
# 1. Clone repository
git clone https://github.com/yourusername/gatekeeper.git
cd gatekeeper

# 2. Create environment file
cp .env.example .env

# 3. Edit .env (REQUIRED)
nano .env  # Update these:
# - JWT_SECRET=$(openssl rand -base64 32)
# - POSTGRES_PASSWORD=strong-password
# - ETHEREUM_RPC=https://eth-mainnet.g.alchemy.com/v2/YOUR_KEY

# 4. Start services
docker compose up -d

# 5. Verify health
docker compose ps
curl http://localhost:8080/health
curl http://localhost:3000/
```

---

## Daily Commands

### Start/Stop

```bash
# Start all services (detached)
docker compose up -d

# Start and view logs
docker compose up

# Stop services (keep data)
docker compose stop

# Stop and remove containers (keep data)
docker compose down

# Stop and remove everything (DELETES DATA)
docker compose down -v
```

### Logs

```bash
# All logs (follow)
docker compose logs -f

# Backend only
docker compose logs -f backend

# Frontend only
docker compose logs -f frontend

# Last 100 lines
docker compose logs --tail=100

# Since specific time
docker compose logs --since=2024-01-01T10:00:00
```

### Status Checks

```bash
# Container status
docker compose ps

# Resource usage
docker stats

# Health checks
curl http://localhost:8080/health  # Backend
curl http://localhost:3000/health  # Frontend
docker compose exec postgres pg_isready -U gatekeeper
docker compose exec redis redis-cli ping
```

---

## Makefile Shortcuts

```bash
make docker-build          # Build all images
make docker-up             # Start services
make docker-down           # Stop services
make docker-logs           # View all logs
make docker-ps             # Show status
make docker-validate       # Run validation
make docker-restart        # Restart all
make docker-backend-logs   # Backend logs
make docker-frontend-logs  # Frontend logs
make docker-clean          # Remove all (interactive)
```

---

## Rebuild & Restart

```bash
# Rebuild specific service
docker compose build backend
docker compose up -d backend

# Rebuild all (no cache)
docker compose build --no-cache

# Full rebuild and restart
docker compose down
docker compose build
docker compose up -d
```

---

## Database Operations

```bash
# Connect to database
docker compose exec postgres psql -U gatekeeper -d gatekeeper

# List tables
docker compose exec postgres psql -U gatekeeper -d gatekeeper -c "\dt"

# Run query
docker compose exec postgres psql -U gatekeeper -d gatekeeper -c "SELECT COUNT(*) FROM users;"

# Backup database
docker compose exec postgres pg_dump -U gatekeeper gatekeeper > backup.sql

# Restore database
cat backup.sql | docker compose exec -T postgres psql -U gatekeeper gatekeeper

# Reset database (DELETES ALL DATA)
docker compose down -v
docker compose up -d
```

---

## Service-Specific Commands

### Backend

```bash
# Restart backend
docker compose restart backend

# View backend logs
docker compose logs -f backend

# Execute command in backend
docker compose exec backend sh

# Check environment
docker compose exec backend env | grep DATABASE_URL
```

### Frontend

```bash
# Restart frontend
docker compose restart frontend

# View frontend logs
docker compose logs -f frontend

# Execute command in frontend
docker compose exec frontend sh

# Check nginx config
docker compose exec frontend cat /etc/nginx/conf.d/default.conf
```

### PostgreSQL

```bash
# Restart database
docker compose restart postgres

# View PostgreSQL logs
docker compose logs -f postgres

# Database shell
docker compose exec postgres psql -U gatekeeper -d gatekeeper

# Check connections
docker compose exec postgres psql -U gatekeeper -d gatekeeper \
  -c "SELECT count(*) FROM pg_stat_activity;"
```

### Redis

```bash
# Restart Redis
docker compose restart redis

# Redis CLI
docker compose exec redis redis-cli

# View keys
docker compose exec redis redis-cli KEYS '*'

# Flush all data
docker compose exec redis redis-cli FLUSHALL
```

---

## Troubleshooting

### Service Won't Start

```bash
# Check logs
docker compose logs <service>

# Inspect container
docker compose ps
docker inspect gatekeeper-<service>

# Restart service
docker compose restart <service>

# Rebuild service
docker compose build --no-cache <service>
docker compose up -d <service>
```

### Port Already in Use

```bash
# Find process using port
lsof -i :8080
lsof -i :3000

# Change port in .env
BACKEND_PORT=8081
FRONTEND_PORT=3001

# Restart
docker compose down
docker compose up -d
```

### Out of Memory

```bash
# Check memory usage
docker stats

# Increase Docker memory limit
# Docker Desktop: Settings → Resources → Memory

# Clean up unused resources
docker system prune -a
```

### Clean Slate Restart

```bash
# Stop everything
docker compose down -v

# Remove all images
docker compose down --rmi all

# Remove Docker system cache
docker system prune -a -f

# Rebuild and start fresh
docker compose build --no-cache
docker compose up -d
```

---

## Environment Variables

### Essential Variables (Update in .env)

```bash
# Security
JWT_SECRET=<generate with: openssl rand -base64 32>
POSTGRES_PASSWORD=<strong password>

# Blockchain
ETHEREUM_RPC=https://eth-mainnet.g.alchemy.com/v2/YOUR_API_KEY
CHAIN_ID=1  # 1=mainnet, 11155111=sepolia

# Frontend
VITE_API_URL=http://localhost:8080
VITE_WALLETCONNECT_PROJECT_ID=<your project id>
```

### Change Port

```bash
# In .env
BACKEND_PORT=8080
FRONTEND_PORT=3000
POSTGRES_PORT=5432
REDIS_PORT=6379

# Restart
docker compose down
docker compose up -d
```

### Change Log Level

```bash
# In .env
LOG_LEVEL=debug  # debug, info, warn, error

# Restart
docker compose restart backend
```

---

## Development Mode

### Enable Hot Reload

```bash
# compose.override.yml is automatically used
docker compose up -d

# Frontend uses Dockerfile.dev with Vite HMR
# Backend requires volume mounts + air (Go hot reload)
```

### Development vs Production

```bash
# Development (uses compose.override.yml)
docker compose up -d

# Production (ignore override)
docker compose -f docker-compose.yml up -d

# Or use production env file
docker compose --env-file .env.production up -d
```

---

## Scaling

### Horizontal Scaling

```bash
# Scale backend to 3 instances
docker compose up -d --scale backend=3

# Note: Requires load balancer configuration
```

### Resource Limits

Edit `docker-compose.yml`:

```yaml
services:
  backend:
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 1G
```

---

## Monitoring

### Real-time Monitoring

```bash
# All containers
docker stats

# Specific container
docker stats gatekeeper-backend

# Export metrics (Prometheus format)
# Add cAdvisor to docker-compose.yml
```

### Health Monitoring

```bash
# All services
docker compose ps

# Backend health
watch -n 5 'curl -s http://localhost:8080/health'

# Frontend health
watch -n 5 'curl -s http://localhost:3000/health'
```

---

## Security

### Scan Images

```bash
# Scan for vulnerabilities
docker scan gatekeeper-backend
docker scan gatekeeper-frontend

# Update base images
docker compose pull
docker compose build --no-cache
```

### Check for Updates

```bash
# Pull latest base images
docker compose pull

# Rebuild with updates
docker compose build --no-cache
docker compose up -d
```

---

## Access Points

| Service | URL | Health Check |
|---------|-----|--------------|
| Frontend | http://localhost:3000 | http://localhost:3000/health |
| Backend | http://localhost:8080 | http://localhost:8080/health |
| PostgreSQL | localhost:5432 | `pg_isready` |
| Redis | localhost:6379 | `redis-cli ping` |

---

## File Structure

```
gatekeeper/
├── Dockerfile                # Backend production
├── docker-compose.yml        # Production orchestration
├── compose.override.yml      # Development overrides
├── .dockerignore            # Backend build exclusions
├── .env.example             # Environment template
├── .env                     # Your configuration (gitignored)
├── web/
│   ├── Dockerfile           # Frontend production
│   ├── Dockerfile.dev       # Frontend development
│   ├── nginx.conf           # Nginx configuration
│   └── .dockerignore        # Frontend build exclusions
├── scripts/
│   ├── docker-build.sh      # Build validation
│   └── docker-validate.sh   # Service validation
└── DOCKER_DEPLOYMENT.md     # Full documentation
```

---

## Common Workflows

### Morning Startup

```bash
cd gatekeeper
docker compose up -d
docker compose logs -f
# Ctrl+C to stop following logs
```

### End of Day Shutdown

```bash
docker compose stop
```

### Deploy New Code

```bash
git pull
docker compose build
docker compose up -d
docker compose logs -f backend
```

### Reset Everything

```bash
docker compose down -v
docker compose build --no-cache
docker compose up -d
./scripts/docker-validate.sh
```

---

## Help & Support

```bash
# View Makefile help
make help

# Validate setup
./scripts/docker-validate.sh

# Build test
./scripts/docker-build.sh

# Docker Compose help
docker compose --help

# Service logs
docker compose logs -f <service>
```

---

## External Links

- [Docker Docs](https://docs.docker.com/)
- [Docker Compose Docs](https://docs.docker.com/compose/)
- [Full Deployment Guide](./DOCKER_DEPLOYMENT.md)
- [Gatekeeper README](./README.md)

---

**Last Updated**: November 2024
**Version**: 1.0.0
