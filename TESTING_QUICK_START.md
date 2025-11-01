# Gatekeeper Testing - Quick Start (10 Minutes)

## Prerequisites Check

```bash
# Verify everything is installed
node --version       # 18+
npm --version       # 9+
go version         # 1.21+
docker --version   # 20.10+
curl --version     # (usually pre-installed)
```

## Step 1: Start Everything (2 min)

```bash
cd gatekeeper

# Start all services (backend, frontend, database, cache)
docker-compose up -d

# Wait for services to be healthy
sleep 5

# Verify all services running
docker-compose ps

# Should see 4 services with "Up" status
```

## Step 2: Test Backend (2 min)

```bash
# Check if backend is healthy
curl -s http://localhost:8080/health | jq .

# Expected: status: "healthy"

# Get nonce for SIWE
curl -s http://localhost:8080/auth/siwe/nonce | jq .

# Expected: nonce with random value
```

## Step 3: Test Frontend (1 min)

```bash
# Open in browser
open http://localhost:3000

# Or check with curl
curl -s http://localhost:3000 | head -5

# Expected: HTML with React app
```

## Step 4: Test SIWE Sign-In (3 min)

### With MetaMask (Recommended):

1. Open http://localhost:3000
2. Install MetaMask if needed: https://metamask.io/
3. Click "Connect Wallet"
4. Select MetaMask → Approve
5. Click "Sign In"
6. MetaMask shows message → Click "Sign"
7. ✅ You're authenticated!

### Or Test with curl:

```bash
# Get nonce
NONCE=$(curl -s http://localhost:8080/auth/siwe/nonce | jq -r '.nonce')

# Create SIWE message
MESSAGE="example.com wants you to sign in with your Ethereum account:
0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266

I accept the Terms of Service: https://example.com/tos

URI: http://localhost:3000
Version: 1
Chain ID: 11155111
Nonce: $NONCE
Issued At: $(date -u +%Y-%m-%dT%H:%M:%SZ)"

# Sign with ethers (requires Node.js)
# (Use MetaMask UI for easier testing)
```

## Step 5: Quick Feature Tests (2 min)

### Test API Keys

```bash
# Get JWT from MetaMask sign-in (see Step 4)
JWT_TOKEN="your_jwt_from_step_4"

# Create API key
API_KEY_RESPONSE=$(curl -s -X POST http://localhost:8080/api/keys \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Key",
    "scopes": ["read", "write"]
  }')

echo $API_KEY_RESPONSE | jq .

# Extract the key
API_KEY=$(echo $API_KEY_RESPONSE | jq -r '.key')

# Use API key to access protected endpoint
curl -s http://localhost:8080/api/data \
  -H "X-API-Key: $API_KEY" | jq .

# ✅ Should return 200 OK
```

### Test Allowlist

```bash
JWT_TOKEN="your_jwt"

# Create allowlist
ALLOWLIST=$(curl -s -X POST http://localhost:8080/api/allowlists \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Test List"}' | jq -r '.id')

echo "Allowlist ID: $ALLOWLIST"

# Add address to allowlist
curl -s -X POST http://localhost:8080/api/allowlists/$ALLOWLIST/entries \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "addresses": ["0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"]
  }' | jq .

# ✅ Address added successfully
```

### Test Error Handling

```bash
# Try with invalid JWT
curl -s http://localhost:8080/api/data \
  -H "Authorization: Bearer invalid_token" | jq .

# Expected: 401 Unauthorized

# Try with missing auth
curl -s http://localhost:8080/api/data | jq .

# Expected: 401 Unauthorized
```

## Step 6: View Full Documentation

```bash
# Quick overview (5 min read)
cat docs/QUICK_FEATURES_SUMMARY.md

# Complete features guide (30 min read)
cat docs/FEATURES_AND_USECASES.md

# Comprehensive testing (full reference)
cat docs/LOCAL_TESTING_COMPREHENSIVE.md

# API reference
cat docs/api/API.md
```

## Validation Checklist ✅

- [ ] All services running: `docker-compose ps`
- [ ] Backend healthy: `curl http://localhost:8080/health`
- [ ] Frontend loads: Open http://localhost:3000
- [ ] Can connect wallet: MetaMask extension
- [ ] Can sign SIWE: Click "Sign In"
- [ ] Get JWT token: Authenticated successfully
- [ ] Create API key: `curl -X POST /api/keys`
- [ ] Use API key: `curl -H "X-API-Key: ..."`
- [ ] Create allowlist: `curl -X POST /api/allowlists`
- [ ] Add to allowlist: `curl -X POST /api/allowlists/{id}/entries`
- [ ] Error handling: Invalid tokens rejected

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Services won't start | `docker-compose down -v && docker-compose up -d` |
| Backend health check fails | `docker-compose logs backend` |
| Can't connect MetaMask | Ensure Sepolia network selected in MetaMask |
| Port 3000 already in use | `lsof -i :3000` then kill the process |
| Database connection error | `docker-compose restart postgres` |

## Quick Commands

```bash
# View all logs
docker-compose logs -f

# Check specific service
docker-compose logs backend
docker-compose logs frontend

# Restart everything
docker-compose restart

# Stop everything
docker-compose down

# Clean slate (remove data)
docker-compose down -v

# Run tests
go test ./internal/... -v
cd web && npm test

# View API documentation
open http://localhost:8080/docs
```

## Next Steps

1. ✅ Everything working locally?
2. Read: [FEATURES_AND_USECASES.md](docs/FEATURES_AND_USECASES.md)
3. Follow: [LOCAL_TESTING_COMPREHENSIVE.md](docs/LOCAL_TESTING_COMPREHENSIVE.md) for full tests
4. Deploy: [DOCKER_DEPLOYMENT.md](docs/deployment/DOCKER_DEPLOYMENT.md)

## Common Endpoints to Test

```bash
# Health check
curl http://localhost:8080/health

# SIWE endpoints
curl http://localhost:8080/auth/siwe/nonce
curl -X POST http://localhost:8080/auth/siwe/verify

# API Key endpoints
curl http://localhost:8080/api/keys
curl -X POST http://localhost:8080/api/keys
curl -X DELETE http://localhost:8080/api/keys/{id}

# Allowlist endpoints
curl http://localhost:8080/api/allowlists
curl -X POST http://localhost:8080/api/allowlists
curl http://localhost:8080/api/allowlists/{id}/entries

# Protected endpoint
curl http://localhost:8080/api/data

# API Documentation
# Redoc: http://localhost:8080/docs
# Swagger: http://localhost:8080/swagger
# OpenAPI spec: http://localhost:8080/openapi.yaml
```

## Need More Help?

- Read full guide: `docs/LOCAL_TESTING_COMPREHENSIVE.md`
- API reference: `docs/api/API.md`
- Features overview: `docs/FEATURES_AND_USECASES.md`
- Deployment: `docs/deployment/DOCKER_DEPLOYMENT.md`

---

**That's it!** You now have Gatekeeper running locally with all features working. 🎉

