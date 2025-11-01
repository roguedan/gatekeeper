# Gatekeeper - Comprehensive Local Testing Guide

Complete guide to testing Gatekeeper locally with all features validated.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Environment Setup](#environment-setup)
3. [Starting Services](#starting-services)
4. [Testing Authentication](#testing-authentication)
5. [Testing Authorization](#testing-authorization)
6. [Testing API Keys](#testing-api-keys)
7. [Testing Policies](#testing-policies)
8. [Frontend Integration Testing](#frontend-integration-testing)
9. [Troubleshooting](#troubleshooting)
10. [Validation Checklist](#validation-checklist)

---

## Prerequisites

### Required Software

```bash
# Check versions
node --version          # Should be 18+
npm --version          # Should be 9+
go version            # Should be 1.21+
docker --version      # Should be 20.10+
docker-compose --version  # Should be 2.0+
```

### Required for Testing

```bash
# Install curl for API testing (usually pre-installed)
curl --version

# Install jq for JSON parsing (optional but helpful)
brew install jq  # macOS
apt-get install jq  # Linux

# Get MetaMask/Wallet for SIWE testing
# Download: https://metamask.io/
```

### Optional Tools

```bash
# For API documentation viewing
# Redoc: http://localhost:8080/docs
# Swagger: http://localhost:8080/swagger

# For database inspection
# Install PostgreSQL client
brew install postgresql  # macOS
apt-get install postgresql-client  # Linux
```

---

## Environment Setup

### Step 1: Create Environment File

```bash
cd /path/to/gatekeeper

# Create .env file with test values
cat > .env << 'EOF'
# Database
DATABASE_URL="postgresql://gatekeeper:gatekeeper@localhost:5432/gatekeeper"
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5

# JWT
JWT_SECRET=$(openssl rand -hex 32)
JWT_EXPIRY_HOURS=24

# SIWE
NONCE_TTL_MINUTES=5

# Ethereum RPC (Sepolia Testnet - FREE, no key needed)
ETHEREUM_RPC="https://eth-sepolia.g.alchemy.com/v2/demo"
ETHEREUM_RPC_FALLBACK="https://rpc.ankr.com/eth_sepolia"
CHAIN_ID=11155111

# Cache
CACHE_TTL=300

# Server
PORT=8080
LOG_LEVEL=debug

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_PER_USER_PER_MINUTE=1000
EOF
```

### Step 2: Verify Database Connection

```bash
# Check PostgreSQL is running
docker ps | grep postgres

# If not running, Docker Compose will start it
# Connection string: postgresql://gatekeeper:gatekeeper@localhost:5432/gatekeeper
```

### Step 3: Create Test Wallet

For local testing, you'll need a test wallet:

**Option A: Use MetaMask (Recommended)**
```
1. Install MetaMask browser extension
2. Create/import a test account
3. Switch to Sepolia testnet
4. Get free test ETH: https://sepoliafaucet.com/
5. Keep the private key for later (not needed for SIWE)
```

**Option B: Use Hardhat Test Account**
```bash
# If you have hardhat installed
npx hardhat accounts

# Use the first account address for testing
# Private key: 0xac0974bec39a17e36ba4a6b4d238ff944bacb476caded87d4b5c93f3aa52f46a (hardhat account 0)
# Address: 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
```

---

## Starting Services

### Step 1: Start Docker Compose Stack

```bash
# From the gatekeeper root directory
docker-compose up -d

# Verify all services started
docker-compose ps

# Output should show:
# NAME                    STATUS
# gatekeeper-backend      Up (healthy)
# gatekeeper-frontend     Up (healthy)
# gatekeeper-postgres     Up (healthy)
# gatekeeper-redis        Up (healthy)
```

### Step 2: Verify Backend is Running

```bash
# Check backend health
curl -s http://localhost:8080/health | jq .

# Expected response:
# {
#   "status": "healthy",
#   "timestamp": "2025-11-01T10:00:00Z",
#   "components": {
#     "database": "healthy",
#     "rpc_provider": "healthy",
#     "cache": "healthy"
#   }
# }
```

### Step 3: Verify Frontend is Running

```bash
# Open in browser or check with curl
curl -s http://localhost:3000/ | head -20

# Should return HTML starting with <!DOCTYPE html>

# Or open in browser:
# http://localhost:3000
```

### Step 4: Check Logs

```bash
# Backend logs
docker-compose logs backend -f

# Frontend logs
docker-compose logs frontend -f

# All services
docker-compose logs -f
```

---

## Testing Authentication

### Test 1: Get Nonce for SIWE

```bash
# Request nonce
curl -s http://localhost:8080/auth/siwe/nonce | jq .

# Expected response:
# {
#   "nonce": "a1b2c3d4e5f6g7h8..."
# }
```

### Test 2: Create SIWE Message (Manual)

```bash
# Store the nonce from previous test
NONCE="a1b2c3d4e5f6g7h8..."

# Create SIWE message (this is what the user signs)
cat > siwe_message.txt << EOF
example.com wants you to sign in with your Ethereum account:
0x1234567890123456789012345678901234567890

I accept the Terms of Service: https://example.com/tos

URI: http://localhost:3000
Version: 1
Chain ID: 11155111
Nonce: $NONCE
Issued At: $(date -u +%Y-%m-%dT%H:%M:%SZ)
EOF

cat siwe_message.txt
```

### Test 3: Sign Message and Get JWT

**Using MetaMask (Recommended for Testing):**

1. Open http://localhost:3000
2. Click "Connect Wallet"
3. Select MetaMask
4. Approve connection
5. Click "Sign In"
6. MetaMask shows message to sign
7. Click "Sign"
8. Frontend receives JWT token
9. Token stored in localStorage
10. You're authenticated!

**Using curl (For Headless Testing):**

```bash
# You need to sign the message with a private key
# This example uses a hardhat test account (DO NOT USE IN PRODUCTION)

# For testing, use the SIWE library to create proper message:
# npm install -g siwe

# Then sign with ethers.js:
cat > sign_message.js << 'EOF'
const { ethers } = require("ethers");

const PRIVATE_KEY = "0xac0974bec39a17e36ba4a6b4d238ff944bacb476caded87d4b5c93f3aa52f46a";
const wallet = new ethers.Wallet(PRIVATE_KEY);

const message = `example.com wants you to sign in with your Ethereum account:
${wallet.address}

I accept the Terms of Service: https://example.com/tos

URI: http://localhost:3000
Version: 1
Chain ID: 11155111
Nonce: test_nonce_12345
Issued At: 2025-11-01T10:00:00Z`;

wallet.signMessage(ethers.getBytes(message)).then(signature => {
  console.log(JSON.stringify({
    message: message,
    signature: signature
  }));
});
EOF

# Run with Node.js
node sign_message.js > signed_message.json

# Verify signature
curl -X POST http://localhost:8080/auth/siwe/verify \
  -H "Content-Type: application/json" \
  -d @signed_message.json | jq .

# Expected response:
# {
#   "token": "eyJhbGciOiJIUzI1NiIs...",
#   "expiresAt": "2025-11-02T10:00:00Z"
# }
```

### Test 4: Use JWT Token

```bash
# Store token from previous response
JWT_TOKEN="eyJhbGciOiJIUzI1NiIs..."

# Test protected endpoint with JWT
curl -s http://localhost:8080/api/data \
  -H "Authorization: Bearer $JWT_TOKEN" | jq .

# Expected: 200 OK with data

# Test with expired/invalid token
curl -s http://localhost:8080/api/data \
  -H "Authorization: Bearer invalid_token" | jq .

# Expected: 401 Unauthorized
```

---

## Testing Authorization

### Test 1: Check Blockchain Access (Read-Only)

```bash
# This doesn't require authentication
# Tests if we can connect to Ethereum RPC

curl -s http://localhost:8080/health | jq '.components.rpc_provider'

# Expected: "healthy"
```

### Test 2: Get User's Address

```bash
# After authentication with SIWE, you can get your address
JWT_TOKEN="your_jwt_token"

curl -s http://localhost:8080/api/profile \
  -H "Authorization: Bearer $JWT_TOKEN" | jq .

# Expected response:
# {
#   "address": "0x1234567890123456789012345678901234567890",
#   "createdAt": "2025-11-01T10:00:00Z"
# }
```

### Test 3: Test Policy Evaluation

```bash
# Create a policy that allows your address
# (This requires API key management - see next section)

curl -X POST http://localhost:8080/api/allowlists \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Allowlist"
  }' | jq .

# Note the allowlistId from response
ALLOWLIST_ID="list_abc123..."

# Add your address to allowlist
curl -X POST http://localhost:8080/api/allowlists/$ALLOWLIST_ID/entries \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "addresses": ["0x1234567890123456789012345678901234567890"]
  }' | jq .

# Verify it's in the list
curl -s http://localhost:8080/api/allowlists/$ALLOWLIST_ID/entries \
  -H "Authorization: Bearer $JWT_TOKEN" | jq .
```

---

## Testing API Keys

### Test 1: Create API Key

```bash
JWT_TOKEN="your_jwt_token"

# Create API key
curl -X POST http://localhost:8080/api/keys \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Key 1",
    "scopes": ["read", "write"],
    "expiresInSeconds": 86400
  }' | jq .

# Response (SAVE THE KEY VALUE - you won't see it again):
# {
#   "id": "key_abc123...",
#   "name": "Test Key 1",
#   "key": "gk_abc123...xyz789",  ⬅️ SAVE THIS SECURELY!
#   "createdAt": "2025-11-01T10:00:00Z",
#   "expiresAt": "2025-11-02T10:00:00Z",
#   "scopes": ["read", "write"]
# }
```

### Test 2: List API Keys

```bash
JWT_TOKEN="your_jwt_token"

# List all your API keys (shows metadata, not raw keys)
curl -s http://localhost:8080/api/keys \
  -H "Authorization: Bearer $JWT_TOKEN" | jq .

# Response:
# {
#   "keys": [
#     {
#       "id": "key_abc123...",
#       "name": "Test Key 1",
#       "createdAt": "2025-11-01T10:00:00Z",
#       "expiresAt": "2025-11-02T10:00:00Z",
#       "lastUsedAt": null,
#       "scopes": ["read", "write"]
#     }
#   ]
# }
```

### Test 3: Use API Key in Requests

```bash
API_KEY="gk_abc123...xyz789"

# Test with X-API-Key header
curl -s http://localhost:8080/api/data \
  -H "X-API-Key: $API_KEY" | jq .

# Test with Authorization: Bearer header
curl -s http://localhost:8080/api/data \
  -H "Authorization: Bearer $API_KEY" | jq .

# Both should return 200 OK
```

### Test 4: Test API Key Expiration

```bash
# Create API key that expires immediately
curl -X POST http://localhost:8080/api/keys \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Expiring Key",
    "scopes": ["read"],
    "expiresInSeconds": 1
  }' | jq .

# Wait 2 seconds
sleep 2

# Try to use expired key
curl -s http://localhost:8080/api/data \
  -H "X-API-Key: gk_expired_key" | jq .

# Expected: 401 Unauthorized (key expired)
```

### Test 5: Revoke API Key

```bash
JWT_TOKEN="your_jwt_token"
KEY_ID="key_abc123..."

# Revoke the key
curl -X DELETE http://localhost:8080/api/keys/$KEY_ID \
  -H "Authorization: Bearer $JWT_TOKEN"

# Expected: 204 No Content

# Try to use revoked key
curl -s http://localhost:8080/api/data \
  -H "X-API-Key: gk_abc123..." | jq .

# Expected: 401 Unauthorized (key not found)
```

---

## Testing Policies

### Test 1: Allowlist Policy

```bash
JWT_TOKEN="your_jwt_token"

# Create allowlist
ALLOWLIST=$(curl -X POST http://localhost:8080/api/allowlists \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "VIP Users"}' | jq -r '.id')

# Add your address
curl -X POST http://localhost:8080/api/allowlists/$ALLOWLIST/entries \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "addresses": ["0x1234567890123456789012345678901234567890"]
  }' | jq .

# Test protected endpoint with policy (once policies are configured)
curl -s http://localhost:8080/api/vip \
  -H "Authorization: Bearer $JWT_TOKEN" | jq .

# Expected: 200 OK (you're on allowlist)
```

### Test 2: Token Balance Policy (Requires Test Tokens)

```bash
# To test token balance checking, you need test tokens
# Option 1: Deploy test ERC20 contract locally
# Option 2: Use existing testnet token

# For Sepolia testnet, test USDC address:
# 0x1c7D4B196Cb0C6f48759e0024F1175C1103f6CF0

# Check if you have test USDC balance
# Your address should have token balance to test this

# The policy evaluation would be triggered by a protected endpoint
# configuration that checks for ERC20MinBalance
```

### Test 3: NFT Ownership Policy (Requires Test NFT)

```bash
# To test NFT ownership, you need to own an NFT on testnet
# Option 1: Mint a test NFT from a faucet
# Option 2: Deploy test ERC721 contract locally

# For Sepolia testnet, you can deploy a simple ERC721:
# Then test ownership verification through policy evaluation
```

### Test 4: Policy Engine (AND/OR Logic)

```bash
# Policies are configured via policy JSON file
# Default location: examples/policies.json

# Example policy configuration:
cat > policies.json << 'EOF'
{
  "policies": [
    {
      "path": "/api/vip",
      "method": "GET",
      "logic": "AND",
      "rules": [
        {
          "type": "HasScope",
          "params": { "scope": "read" }
        },
        {
          "type": "InAllowlist",
          "params": { "allowlistId": "your_list_id" }
        }
      ]
    }
  ]
}
EOF

# Load policy via API
curl -X POST http://localhost:8080/api/policies \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d @policies.json | jq .
```

---

## Frontend Integration Testing

### Test 1: Wallet Connection

1. Open http://localhost:3000
2. You should see "Connect Wallet" button
3. Click button → MetaMask opens
4. Select account and approve
5. Button changes to show connected address ✅

### Test 2: SIWE Sign-In

1. Connected wallet visible
2. Click "Sign In"
3. MetaMask shows message to sign
4. Click "Sign"
5. Frontend receives JWT
6. Redirects to dashboard ✅

### Test 3: Protected Routes

1. After sign-in, navigate to:
   - Dashboard (/dashboard)
   - API Keys (/api-keys)
   - Token Gating (/token-gating)
2. All pages should load
3. Logout and try to access → redirect to login ✅

### Test 4: API Key Management UI

1. Go to API Keys page
2. Click "Create New Key"
3. Enter name and scopes
4. Click "Create"
5. Raw key displayed in modal
6. Copy to clipboard
7. Click "Close"
8. Key appears in list (without raw value)
9. Click "Revoke" → confirm
10. Key removed from list ✅

### Test 5: Token Gating Demo

1. Go to Token Gating page
2. See demo message
3. Click "Check Access"
4. System checks policies
5. Shows result (granted or denied)
6. Mock balances displayed ✅

### Test 6: Error Handling

1. Try invalid JWT: Redirects to login
2. Try expired token: Shows error message
3. Network error: Shows friendly error
4. Invalid input: Shows validation error ✅

---

## Testing Rate Limiting

### Test 1: API Key Creation Rate Limit

```bash
JWT_TOKEN="your_jwt_token"

# Try to create many API keys rapidly
for i in {1..6}; do
  echo "Attempt $i:"
  curl -s -X POST http://localhost:8080/api/keys \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"name\": \"Key $i\",
      \"scopes\": [\"read\"]
    }" | jq '.error // .id'
done

# After 5 attempts, should get rate limit error:
# "rate_limit_exceeded"
```

### Test 2: Rate Limit Headers

```bash
JWT_TOKEN="your_jwt_token"

# Make a request and check rate limit headers
curl -s -i http://localhost:8080/api/keys \
  -H "Authorization: Bearer $JWT_TOKEN" | grep -i "x-ratelimit"

# Should see:
# X-RateLimit-Limit: 1000
# X-RateLimit-Remaining: 999
# X-RateLimit-Reset: [timestamp]
```

### Test 3: Rate Limit Retry-After

```bash
# When rate limited, response includes Retry-After
curl -s -i http://localhost:8080/api/data | grep -i "retry-after"

# Expected: Retry-After: 60 (seconds)
```

---

## Testing Error Handling

### Test 1: Invalid JWT

```bash
curl -s http://localhost:8080/api/data \
  -H "Authorization: Bearer invalid_token" | jq .

# Expected:
# {
#   "error": "unauthorized",
#   "message": "Invalid or missing token"
# }
```

### Test 2: Expired Token

```bash
# Create a token that expires in 1 second
# Wait 2+ seconds
# Try to use it

curl -s http://localhost:8080/api/data \
  -H "Authorization: Bearer expired_token" | jq .

# Expected:
# {
#   "error": "unauthorized",
#   "message": "Token has expired"
# }
```

### Test 3: Access Denied (Policy)

```bash
# Try to access endpoint you're not allowed to
curl -s http://localhost:8080/api/vip \
  -H "Authorization: Bearer $JWT_TOKEN" | jq .

# Expected:
# {
#   "error": "forbidden",
#   "message": "Access denied by policy"
# }
```

### Test 4: Invalid API Key

```bash
curl -s http://localhost:8080/api/data \
  -H "X-API-Key: invalid_key" | jq .

# Expected:
# {
#   "error": "unauthorized",
#   "message": "Invalid API key"
# }
```

### Test 5: Invalid Input

```bash
curl -X POST http://localhost:8080/api/keys \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": ""}' | jq .

# Expected:
# {
#   "error": "bad_request",
#   "message": "Invalid request: name is required"
# }
```

---

## Database Inspection

### Connect to PostgreSQL

```bash
# Using psql (if installed)
psql postgresql://gatekeeper:gatekeeper@localhost:5432/gatekeeper

# Or through Docker
docker exec -it gatekeeper-postgres psql -U gatekeeper -d gatekeeper
```

### View Tables

```sql
-- List all tables
\dt

-- Check users
SELECT * FROM users;

-- Check API keys
SELECT id, user_id, name, created_at, expires_at FROM api_keys;

-- Check allowlists
SELECT * FROM allowlists;

-- Check allowlist entries
SELECT * FROM allowlist_entries;
```

---

## Performance Testing

### Test 1: Policy Evaluation Speed

```bash
JWT_TOKEN="your_jwt_token"

# Time a policy evaluation
time curl -s http://localhost:8080/api/data \
  -H "Authorization: Bearer $JWT_TOKEN" > /dev/null

# Expected: <500ms for cached queries, <1s for RPC calls
```

### Test 2: Cache Hit Rate

```bash
# Make repeated requests
for i in {1..10}; do
  curl -s http://localhost:8080/health | jq '.timestamp' > /dev/null
done

# Check logs for cache hits
docker-compose logs backend | grep "cache"
```

### Test 3: Load Test (Optional)

```bash
# Install Apache Bench
brew install httpd  # macOS

# Run basic load test
ab -n 100 -c 10 \
  -H "Authorization: Bearer $JWT_TOKEN" \
  http://localhost:8080/health

# Should handle 100 requests in <5 seconds
```

---

## Validation Checklist

Use this checklist to validate all functionality:

### Authentication
- [ ] Can get nonce from /auth/siwe/nonce
- [ ] Can sign SIWE message with MetaMask
- [ ] Can verify signature and get JWT
- [ ] JWT is valid for protected endpoints
- [ ] Expired JWT is rejected
- [ ] Invalid JWT is rejected

### API Keys
- [ ] Can create API key via POST /api/keys
- [ ] Raw key shown only once
- [ ] Can list API keys via GET /api/keys
- [ ] Can use API key in X-API-Key header
- [ ] Can use API key in Authorization header
- [ ] Can revoke API key via DELETE /api/keys/{id}
- [ ] Revoked key is rejected
- [ ] Expired key is rejected

### Authorization
- [ ] Can add addresses to allowlist
- [ ] Can check allowlist membership
- [ ] Can remove addresses from allowlist
- [ ] Protected endpoints require JWT or API key
- [ ] Unauthenticated requests are denied

### Policies
- [ ] Policy with AND logic requires all rules to pass
- [ ] Policy with OR logic requires any rule to pass
- [ ] InAllowlist rule works
- [ ] HasScope rule works (if configured)
- [ ] Policy evaluation is fast (<500ms)

### Rate Limiting
- [ ] Rate limit headers present in responses
- [ ] Rapid requests are rate limited
- [ ] Rate limit exceeded returns 429
- [ ] Retry-After header provided

### Error Handling
- [ ] Invalid token returns 401
- [ ] Expired token returns 401
- [ ] Insufficient permissions return 403
- [ ] Invalid input returns 400
- [ ] Server errors return 500
- [ ] Error messages don't leak sensitive info

### Frontend
- [ ] Can connect wallet
- [ ] Can sign SIWE message
- [ ] Can access dashboard after sign-in
- [ ] Can create API key via UI
- [ ] Can list API keys
- [ ] Can revoke API key
- [ ] Protected routes are accessible
- [ ] Unauthenticated access redirects to login

### Performance
- [ ] Health check <10ms
- [ ] JWT verification <5ms
- [ ] Policy evaluation (cached) <50ms
- [ ] Policy evaluation (RPC) <500ms

### Security
- [ ] No sensitive data in error messages
- [ ] No sensitive data in logs
- [ ] API keys are hashed in database
- [ ] SIWE nonce is single-use
- [ ] SIWE message includes chain ID
- [ ] Signature verification works

---

## Troubleshooting

### Backend Won't Start

```bash
# Check logs
docker-compose logs backend

# Common issues:
# 1. Database not running: docker-compose up -d postgres
# 2. Invalid DATABASE_URL in .env
# 3. Port 8080 already in use: lsof -i :8080
```

### Can't Connect to Frontend

```bash
# Check frontend logs
docker-compose logs frontend

# Try rebuilding
docker-compose down
docker-compose up -d --build

# Check port 3000
lsof -i :3000
```

### Database Connection Error

```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Check credentials in .env match docker-compose.yml

# Try connecting directly
psql postgresql://gatekeeper:gatekeeper@localhost:5432/gatekeeper
```

### MetaMask Connection Issues

```bash
# Clear MetaMask cache
1. Open MetaMask extension
2. Settings → Advanced → Clear activity tab data
3. Reload page

# Ensure correct network (Sepolia for testnet)
1. MetaMask shows network selector
2. Click and select "Sepolia"
3. Try again
```

### RPC Provider Error

```bash
# Check RPC endpoint is accessible
curl -X POST https://eth-sepolia.g.alchemy.com/v2/demo \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq .

# If error, update ETHEREUM_RPC in .env

# Use fallback RPC
export ETHEREUM_RPC="https://rpc.ankr.com/eth_sepolia"
```

---

## Next Steps

After validating all functionality:

1. **Review logs**: Check for any errors or warnings
   ```bash
   docker-compose logs --tail=100
   ```

2. **Run test suite**: Execute full test coverage
   ```bash
   go test ./internal/... -v
   cd web && npm test
   ```

3. **Check documentation**: Review updated docs
   ```bash
   cat docs/QUICK_FEATURES_SUMMARY.md
   cat docs/FEATURES_AND_USECASES.md
   ```

4. **Prepare for production**: Follow deployment guide
   ```bash
   cat docs/deployment/DOCKER_DEPLOYMENT.md
   ```

---

## Quick Reference Commands

```bash
# Start services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f backend

# Stop services
docker-compose down

# Restart single service
docker-compose restart backend

# Remove all data (clean slate)
docker-compose down -v

# Run backend tests
go test ./internal/... -v

# Run frontend tests
cd web && npm test

# Check backend health
curl http://localhost:8080/health | jq .

# Check frontend
curl http://localhost:3000 | head -20
```

---

## Support

If you encounter issues:

1. Check [docs/README.md](./README.md) for documentation index
2. Review [docs/guides/LOCAL_TESTING.md](./guides/LOCAL_TESTING.md) for basic testing
3. Check logs: `docker-compose logs`
4. Verify environment: `cat .env`
5. Review [docs/api/API.md](./api/API.md) for endpoint details

