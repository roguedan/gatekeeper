# Gatekeeper: Features & Use Cases

A comprehensive guide to Gatekeeper's capabilities and real-world applications.

---

## Table of Contents

1. [Core Features](#core-features)
2. [Authentication Features](#authentication-features)
3. [Authorization & Access Control](#authorization--access-control)
4. [Integration Features](#integration-features)
5. [Operational Features](#operational-features)
6. [Use Cases](#use-cases)
7. [Industry Applications](#industry-applications)
8. [Technical Capabilities](#technical-capabilities)

---

## Core Features

### 1. **Wallet-Native Authentication (SIWE)**

**What it does:**
Enable users to sign in using their Ethereum wallet instead of passwords.

**Key Capabilities:**
- ✅ Sign-In with Ethereum (EIP-4361) compliant
- ✅ No password management required
- ✅ Works with any Ethereum wallet
- ✅ Single-use, TTL-based nonce for replay prevention
- ✅ Message verification using EIP-191 personal_sign
- ✅ Automatic user creation on first sign-in

**Example:**
```
User Flow:
1. User connects wallet (MetaMask, WalletConnect, etc.)
2. System generates unique nonce
3. User signs message with wallet
4. System verifies signature and issues JWT
5. User accesses protected resources with JWT
```

**Benefits:**
- No passwords to remember or reset
- Wallet acts as identity provider
- User controls private keys (self-custody)
- Can't be phished for password
- Works across chains

---

### 2. **JWT Token Management**

**What it does:**
Issue and manage JSON Web Tokens (JWT) for authenticated sessions.

**Key Capabilities:**
- ✅ HS256 signed tokens
- ✅ Configurable expiration (default: 24 hours)
- ✅ Token validation on protected endpoints
- ✅ Automatic token refresh capability
- ✅ Revocation support
- ✅ Token payload includes user address and scopes

**Example Token Payload:**
```json
{
  "sub": "0x1234567890123456789012345678901234567890",
  "iat": 1698969600,
  "exp": 1699056000,
  "scopes": ["read", "write"],
  "iss": "gatekeeper"
}
```

**Benefits:**
- Stateless authentication (no session storage needed)
- Works across distributed systems
- Can be used for multiple APIs
- Includes user identity and permissions

---

### 3. **API Key Management System**

**What it does:**
Generate and manage API keys for programmatic access.

**Key Capabilities:**
- ✅ Cryptographically secure key generation (32 bytes)
- ✅ SHA256 hashing before storage
- ✅ Raw key displayed only once on creation
- ✅ Key expiration support
- ✅ Scopes for permission granularity
- ✅ Last-used tracking for audit trails
- ✅ Ownership verification
- ✅ Revocation (immediate invalidation)

**HTTP Endpoints:**
```
POST   /api/keys                 # Create new API key
GET    /api/keys                 # List user's keys
DELETE /api/keys/{id}            # Revoke API key
```

**Example Usage:**
```bash
# Create API key
curl -X POST http://localhost:8080/api/keys \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My App",
    "scopes": ["read", "write"],
    "expiresInSeconds": 2592000
  }'

# Response:
{
  "id": "key_abc123...",
  "name": "My App",
  "key": "gk_abc123...xyz789",  # Save this securely!
  "createdAt": "2025-11-01T10:00:00Z",
  "expiresAt": "2025-12-01T10:00:00Z",
  "scopes": ["read", "write"]
}

# Use API key in requests
curl http://localhost:8080/api/data \
  -H "X-API-Key: gk_abc123...xyz789"
```

**Benefits:**
- Programmatic access without exposing wallet
- Can be revoked without changing passwords
- Tracks usage patterns
- Supports expiration for security
- Different scopes for different apps

---

## Authentication Features

### 1. **Multi-Wallet Support**

**Supported Wallets:**
- ✅ MetaMask
- ✅ WalletConnect
- ✅ Coinbase Wallet
- ✅ Ledger (via WalletConnect)
- ✅ Trezor (via WalletConnect)
- ✅ Rainbow (via RainbowKit)
- ✅ Any EIP-1193 compatible wallet

**Benefits:**
- Users choose their preferred wallet
- Works with hardware wallets
- Mobile and desktop support
- No wallet dependency

---

### 2. **Multi-Chain Support**

**Supported Networks:**
- ✅ Ethereum Mainnet (Chain ID: 1)
- ✅ Polygon (Chain ID: 137)
- ✅ Arbitrum (Chain ID: 42161)
- ✅ Optimism (Chain ID: 10)
- ✅ Base (Chain ID: 8453)
- ✅ Sepolia Testnet (Chain ID: 11155111)
- ✅ Any EVM-compatible chain

**Example: Polygon Sign-In**
```typescript
// User signs message on Polygon
// System verifies on Polygon RPC
// JWT valid for all chains
const siweMessage = createSIWEMessage({
  address: userAddress,
  chainId: 137,  // Polygon
  nonce: nonce,
  version: '1'
});
```

**Benefits:**
- Users can sign in from any chain they own tokens on
- Single account works across chains
- Reduces friction for multi-chain users
- Can enforce chain-specific policies

---

### 3. **Nonce Management**

**What it does:**
Prevent replay attacks with single-use nonces.

**Key Capabilities:**
- ✅ Unique nonce per sign-in attempt
- ✅ TTL-based expiration (default: 5 minutes)
- ✅ Single-use enforcement
- ✅ Database-backed nonce storage

**Security:**
```
1. User requests nonce: GET /auth/siwe/nonce
   Response: { nonce: "abc123..." }

2. Nonce expires after 5 minutes (configurable)

3. Nonce can only be used once
   Reuse attempt: REJECTED

4. Second sign-in requires new nonce
```

**Benefits:**
- Prevents signature reuse attacks
- Prevents man-in-the-middle attacks
- Standard replay attack protection
- No additional user friction

---

## Authorization & Access Control

### 1. **Flexible Policy Engine**

**What it does:**
Define rule-based access control policies with complex logic.

**Supported Rule Types:**
- ✅ `HasScope` - Permission checks
- ✅ `InAllowlist` - Address whitelisting
- ✅ `ERC20MinBalance` - Token balance requirements
- ✅ `ERC721Owner` - NFT ownership verification

**Example Policy Configuration:**
```json
{
  "policies": [
    {
      "path": "/api/vip",
      "method": "GET",
      "logic": "AND",
      "rules": [
        {
          "type": "HasScope",
          "params": { "scope": "admin" }
        },
        {
          "type": "ERC20MinBalance",
          "params": {
            "token": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
            "chainId": 1,
            "minimum": "1000000000"  // 1000 USDC
          }
        }
      ]
    }
  ]
}
```

**Logic Types:**
- `AND` - All rules must pass
- `OR` - At least one rule must pass

**Benefits:**
- Expressive access control
- Combines multiple conditions
- Supports complex business logic
- Easily configurable

---

### 2. **ERC20 Token Balance Checking**

**What it does:**
Verify users hold minimum token balance.

**Capabilities:**
- ✅ Any ERC20 token on any EVM chain
- ✅ Minimum balance requirements
- ✅ RPC-based balance checking
- ✅ TTL-based caching (5 minutes default)
- ✅ Failover RPC provider support
- ✅ Fail-closed security (deny on error)

**Real-World Examples:**
```
USDC Balance ≥ $1000:
- token: 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48
- minimum: 1000000000 (1000 * 10^6 decimals)

DAI Balance ≥ 500:
- token: 0x6B175474E89094C44Da98b954EedeAC495271d0F
- minimum: 500000000000000000000 (500 * 10^18 decimals)

USDT Balance ≥ 100:
- token: 0xdAC17F958D2ee523a2206206994597C13D831ec7
- minimum: 100000000 (100 * 10^6 decimals)
```

**Benefits:**
- Gate access based on token holdings
- No smart contract required
- Works with any ERC20 token
- Cached for performance (80%+ cache hit rate)

---

### 3. **ERC721 NFT Ownership Verification**

**What it does:**
Verify users own specific NFTs.

**Capabilities:**
- ✅ Any ERC721 contract on any EVM chain
- ✅ Specific token ID ownership checks
- ✅ Collection-wide ownership checks
- ✅ RPC-based verification
- ✅ TTL-based caching
- ✅ Burned token handling

**Real-World Examples:**
```
Bored Ape Yacht Club (BAYC):
- contract: 0xBC4CA0EdA7647A8aB7C2061c2E2ad7d6d9e77241
- anyToken: true  # Check if user owns any BAYC

Specific Cryptopunk:
- contract: 0xb47e3cd837dDF8e4c57F05d70Ab865de6e193BBB
- tokenId: 1      # Check if user owns Cryptopunk #1

Pudgy Penguins:
- contract: 0xBd3531dA5DD0A74fb411a9b7FaD7d47B1B1395d4
- anyToken: true  # Check if user owns any Pudgy Penguin
```

**Benefits:**
- Gate access to NFT holders
- Verify specific NFT ownership
- Works with any ERC721 contract
- Supports burned token edge cases
- Cached for performance

---

### 4. **Address Whitelisting**

**What it does:**
Explicitly allow specific addresses.

**Capabilities:**
- ✅ Exact address matching
- ✅ Case-insensitive comparison
- ✅ Batch operations for performance
- ✅ Fast lookup using EXISTS queries
- ✅ Add/remove addresses dynamically

**HTTP Endpoints:**
```
GET    /api/allowlists           # List allowlists
POST   /api/allowlists           # Create allowlist
POST   /api/allowlists/{id}/entries  # Add addresses
DELETE /api/allowlists/{id}/entries/{address}  # Remove address
```

**Example Usage:**
```bash
# Create VIP allowlist
curl -X POST http://localhost:8080/api/allowlists \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "VIP Users"
  }'

# Add addresses to allowlist
curl -X POST http://localhost:8080/api/allowlists/list_123/entries \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "addresses": [
      "0x742d35Cc6634C0532925a3b844Bc9e7595f42600",
      "0x1234567890123456789012345678901234567890"
    ]
  }'
```

**Benefits:**
- Simple explicit access control
- Fast lookups
- Easy to manage
- Can combine with other rules (OR logic)

---

## Integration Features

### 1. **OpenAPI Documentation**

**What it does:**
Auto-generated API documentation with interactive testing.

**Capabilities:**
- ✅ Complete OpenAPI 3.0 specification
- ✅ Redoc interactive documentation
- ✅ Swagger UI for testing
- ✅ Request/response examples
- ✅ Schema validation
- ✅ Security scheme documentation

**Access Points:**
```
OpenAPI Spec:  GET /openapi.yaml
Redoc Docs:    GET /docs
Swagger UI:    GET /swagger
Health Check:  GET /health
```

**Benefits:**
- Self-documenting API
- Developers can test endpoints
- Clear contract definition
- Easy SDK generation

---

### 2. **REST API Endpoints**

**Authentication Endpoints:**
```
GET    /auth/siwe/nonce          # Get nonce for signing
POST   /auth/siwe/verify         # Verify signature & get JWT
GET    /auth/health              # Check auth service health
```

**API Key Endpoints:**
```
POST   /api/keys                 # Create new API key
GET    /api/keys                 # List user's API keys
DELETE /api/keys/{id}            # Revoke API key
```

**Allowlist Endpoints:**
```
GET    /api/allowlists           # List allowlists
POST   /api/allowlists           # Create allowlist
GET    /api/allowlists/{id}      # Get allowlist details
POST   /api/allowlists/{id}/entries     # Add addresses
DELETE /api/allowlists/{id}/entries/{addr}  # Remove address
GET    /api/allowlists/{id}/entries     # List addresses
```

**Protected Endpoint Example:**
```
GET    /api/data                 # Protected endpoint
       Requires: Authorization: Bearer {JWT}
       Response: 200 OK if authorized
       Response: 403 Forbidden if policy denies
```

**Benefits:**
- Standard RESTful design
- Easy to integrate
- Stateless (no session management)
- Works with any language/framework

---

### 3. **Rate Limiting**

**What it does:**
Protect against abuse and DoS attacks.

**Capabilities:**
- ✅ Per-user rate limiting
- ✅ Per-IP rate limiting
- ✅ Token bucket algorithm
- ✅ Configurable limits
- ✅ HTTP 429 Too Many Requests responses
- ✅ Retry-After headers

**Configuration:**
```
API Key Creation:
  - Max 10 API keys per user
  - Max 5 creation attempts per minute
  - Max 1000 requests per minute (overall)

Protected Endpoints:
  - User-based limits (depends on plan)
  - IP-based limits for unauthenticated
  - Backoff recommendations
```

**Example Response:**
```
HTTP/1.1 429 Too Many Requests
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1698969660
Retry-After: 60

{
  "error": "rate limit exceeded",
  "retryAfter": 60
}
```

**Benefits:**
- Prevents API abuse
- Protects backend resources
- Fair resource allocation
- Clear client feedback

---

## Operational Features

### 1. **Comprehensive Logging & Audit Trail**

**What it does:**
Track all security-relevant events.

**Logged Events:**
- ✅ SIWE message generation
- ✅ Signature verification
- ✅ JWT token issuance/validation
- ✅ API key creation/deletion
- ✅ API key usage
- ✅ Allowlist modifications
- ✅ Policy evaluation results
- ✅ Authentication failures
- ✅ Authorization denials
- ✅ Rate limit violations

**Log Format:**
```json
{
  "timestamp": "2025-11-01T10:00:00Z",
  "level": "info",
  "requestID": "req_abc123...",
  "userAddress": "0x1234567890123456789012345678901234567890",
  "action": "api_key_created",
  "resourceID": "key_xyz789...",
  "result": "success",
  "metadata": {
    "keyName": "My App",
    "scopes": ["read", "write"],
    "expiresAt": "2025-12-01T10:00:00Z"
  }
}
```

**Benefits:**
- Compliance and audit requirements
- Incident investigation
- Usage analytics
- Security monitoring

---

### 2. **Health Checks & Monitoring**

**What it does:**
Monitor system health and dependencies.

**Health Check Endpoint:**
```
GET /health

Response:
{
  "status": "healthy",
  "timestamp": "2025-11-01T10:00:00Z",
  "components": {
    "database": "healthy",
    "rpc_provider": "healthy",
    "cache": "healthy",
    "auth": "healthy",
    "policy_engine": "healthy"
  }
}
```

**Monitored Components:**
- ✅ Database connectivity
- ✅ RPC provider availability
- ✅ Cache functionality
- ✅ Authentication service
- ✅ Policy evaluation service

**Benefits:**
- Early problem detection
- Load balancer integration
- Automated recovery triggering
- Service dependency monitoring

---

### 3. **Error Handling & Security**

**What it does:**
Handle errors safely without leaking information.

**Error Response Examples:**
```
Authentication Error:
{
  "error": "invalid_token",
  "message": "Token validation failed"
}

Authorization Error:
{
  "error": "access_denied",
  "message": "Policy evaluation denied access"
}

Rate Limit Error:
{
  "error": "rate_limit_exceeded",
  "retryAfter": 60
}

Server Error:
{
  "error": "internal_server_error",
  "message": "An unexpected error occurred"
  # Note: Detailed error NOT included in response
}
```

**Security Practices:**
- ✅ No sensitive data in error messages
- ✅ No stack traces in responses
- ✅ Detailed logging for debugging (not in responses)
- ✅ Fail-closed for blockchain rules
- ✅ Input validation on all endpoints

**Benefits:**
- Won't leak information to attackers
- Clear client feedback
- Detailed logs for debugging
- Industry-standard security practices

---

## Use Cases

### **Use Case 1: Token-Holder Community Access**

**Scenario:**
Allow only USDC holders ($100+) to access premium features.

**Configuration:**
```json
{
  "policies": [{
    "path": "/api/premium/*",
    "method": "GET",
    "logic": "AND",
    "rules": [{
      "type": "ERC20MinBalance",
      "params": {
        "token": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
        "chainId": 1,
        "minimum": "100000000"
      }
    }]
  }]
}
```

**User Flow:**
1. User connects wallet
2. User signs SIWE message
3. User gets JWT token
4. User accesses `/api/premium`
5. System checks USDC balance
6. Access granted if ≥100 USDC

**Benefits:**
- Monetized access tier
- Automatic verification
- No manual approval needed
- User controls participation

---

### **Use Case 2: NFT-Gated Community**

**Scenario:**
Allow only Bored Ape Yacht Club (BAYC) owners to join Discord community.

**Configuration:**
```json
{
  "policies": [{
    "path": "/api/discord/webhook",
    "method": "POST",
    "logic": "AND",
    "rules": [{
      "type": "ERC721Owner",
      "params": {
        "contract": "0xBC4CA0EdA7647A8aB7C2061c2E2ad7d6d9e77241",
        "chainId": 1,
        "anyToken": true
      }
    }]
  }]
}
```

**User Flow:**
1. User visits community site
2. Connects wallet
3. Signs SIWE message
4. System verifies BAYC ownership
5. User granted Discord access
6. Automation adds user to Discord

**Benefits:**
- Exclusive community
- Automated membership
- Verifiable ownership
- Cross-platform integration

---

### **Use Case 3: Multi-Criteria Access (AND Logic)**

**Scenario:**
VIP tier requires: ETH holder + USDC holder + on whitelist

**Configuration:**
```json
{
  "policies": [{
    "path": "/api/vip/*",
    "method": "GET",
    "logic": "AND",
    "rules": [
      {
        "type": "ERC20MinBalance",
        "params": {
          "token": "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
          "chainId": 1,
          "minimum": "1000000000000000000"
        }
      },
      {
        "type": "ERC20MinBalance",
        "params": {
          "token": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
          "chainId": 1,
          "minimum": "10000000000"
        }
      },
      {
        "type": "InAllowlist",
        "params": {
          "allowlistId": "list_vip_users"
        }
      }
    ]
  }]
}
```

**Requirements:**
- ✅ Hold 1+ WETH (Ethereum)
- ✅ AND Hold 10,000+ USDC
- ✅ AND Be on VIP whitelist

**Benefits:**
- Complex access rules
- Multiple criteria
- Flexible business logic
- All automated

---

### **Use Case 4: Alternative Access (OR Logic)**

**Scenario:**
Access granted if: BAYC owner OR MAYC owner OR on allowlist

**Configuration:**
```json
{
  "policies": [{
    "path": "/api/ape-club/*",
    "method": "GET",
    "logic": "OR",
    "rules": [
      {
        "type": "ERC721Owner",
        "params": {
          "contract": "0xBC4CA0EdA7647A8aB7C2061c2E2ad7d6d9e77241",
          "chainId": 1,
          "anyToken": true
        }
      },
      {
        "type": "ERC721Owner",
        "params": {
          "contract": "0x60E4d786d1ad0075D4017F987210B6Bc415a9Bda",
          "chainId": 1,
          "anyToken": true
        }
      },
      {
        "type": "InAllowlist",
        "params": {
          "allowlistId": "list_approved_users"
        }
      }
    ]
  }]
}
```

**Requirements:**
- ✅ Own BAYC OR
- ✅ Own MAYC (Mutant Ape Yacht Club) OR
- ✅ Be on approved list

**Benefits:**
- Multiple paths to access
- Flexible membership
- Easy to update
- Community-driven

---

### **Use Case 5: Graduated Access Tiers**

**Scenario:**
Different endpoints require different token amounts (freemium model).

**Configuration:**
```json
{
  "policies": [
    {
      "path": "/api/basic/*",
      "method": "GET",
      "logic": "AND",
      "rules": [{
        "type": "ERC20MinBalance",
        "params": {
          "token": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
          "chainId": 1,
          "minimum": "100000000"
        }
      }]
    },
    {
      "path": "/api/pro/*",
      "method": "GET",
      "logic": "AND",
      "rules": [{
        "type": "ERC20MinBalance",
        "params": {
          "token": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
          "chainId": 1,
          "minimum": "10000000000"
        }
      }]
    },
    {
      "path": "/api/elite/*",
      "method": "GET",
      "logic": "AND",
      "rules": [{
        "type": "ERC20MinBalance",
        "params": {
          "token": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
          "chainId": 1,
          "minimum": "100000000000"
        }
      }]
    }
  ]
}
```

**Tiers:**
- Basic: 100 USDC (entry level)
- Pro: 10,000 USDC (mid-tier)
- Elite: 100,000 USDC (premium)

**Benefits:**
- Revenue-based tiers
- Automatic enforcement
- Clear user expectations
- Easy to adjust

---

### **Use Case 6: Cross-Chain Access**

**Scenario:**
Allow users to sign in from any chain and access all resources.

**Configuration:**
```json
{
  "policies": [{
    "path": "/api/app/*",
    "method": "GET",
    "logic": "OR",
    "rules": [
      {
        "type": "ERC20MinBalance",
        "params": {
          "token": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
          "chainId": 1,
          "minimum": "1000000000"
        }
      },
      {
        "type": "ERC20MinBalance",
        "params": {
          "token": "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174",
          "chainId": 137,
          "minimum": "1000000000"
        }
      },
      {
        "type": "ERC20MinBalance",
        "params": {
          "token": "0xFF970A61A04b1cA14834A43f5dE4533eBDDB5F86",
          "chainId": 42161,
          "minimum": "1000000000"
        }
      }
    ]
  }]
}
```

**Supported Chains:**
- ✅ Ethereum Mainnet (USDC)
- ✅ Polygon (USDC.e)
- ✅ Arbitrum (USDC)

**User Flow:**
1. User can sign in from any chain
2. System checks balances on all chains
3. Access if any chain meets requirement
4. Works globally across networks

**Benefits:**
- Multi-chain users
- Users on any chain can access
- Reduces network friction
- Broader user base

---

### **Use Case 7: API-First Integration**

**Scenario:**
Enable third-party apps to use API keys instead of SIWE.

**Workflow:**
```
1. User generates API key via frontend
   POST /api/keys
   Response: { key: "gk_abc123..." }

2. User saves key in app config
   GATEKEEPER_API_KEY=gk_abc123...

3. App uses key in API calls
   GET /api/data
   Headers: X-API-Key: gk_abc123...

4. System authenticates with API key
   Verifies key signature
   Looks up user permissions
   Returns data if authorized

5. Logs track API usage
   For audit and analytics
```

**Benefits:**
- Programmatic access
- API-first architecture
- Third-party integrations
- Usage tracking

---

## Industry Applications

### **DeFi Protocols**

**Example: Yield Farming Dashboard**
- Gate access to farmers with minimum locked value
- Require governance token holders for voting
- Track API key usage for analytics
- Multi-chain token support (ETH, Polygon, Arbitrum)

**Implementation:**
```json
{
  "path": "/api/dashboard/yields",
  "logic": "AND",
  "rules": [
    {
      "type": "ERC20MinBalance",
      "params": {
        "token": "0x...",
        "minimum": "1000000000000000000"
      }
    }
  ]
}
```

---

### **NFT Projects & Communities**

**Example: Creator Membership**
- Gate content behind NFT ownership
- Different content tiers for different NFTs
- Discord integration for verified members
- Cross-collection access (BAYC OR MAYC)

**Implementation:**
```json
{
  "path": "/api/exclusive-content",
  "logic": "OR",
  "rules": [
    {
      "type": "ERC721Owner",
      "params": {
        "contract": "0xBC4CA0EdA7647A8aB7C2061c2E2ad7d6d9e77241"
      }
    }
  ]
}
```

---

### **GameFi & Metaverse**

**Example: Play-to-Earn Game**
- Require governance token to play
- Different reward tiers by token balance
- Whitelist for beta access
- Multi-wallet support for accounts

**Features:**
- ✅ Wallet-native authentication
- ✅ Token-based access tiers
- ✅ Whitelist for early access
- ✅ API key for game servers

---

### **Web3 Social Platforms**

**Example: Decentralized Twitter**
- Gate profiles behind token holdings
- Verify NFT ownership for blue checkmark
- Different posting limits by tier
- API access for bot integrations

**Features:**
- ✅ SIWE for account creation
- ✅ Token balance verification
- ✅ NFT-based verification
- ✅ Rate limiting for API bots

---

### **DAO Governance**

**Example: Multi-Sig Voting**
- Gate voting to token holders
- Require minimum token balance
- Whitelist for executive board
- API access for governance tools

**Features:**
- ✅ Token balance verification
- ✅ Allowlist for council
- ✅ API keys for integrations
- ✅ Audit logging for compliance

---

### **Enterprise Web3 Integration**

**Example: Corporate Blockchain**
- Employee wallet authentication
- Role-based access (CEO, CFO, etc.)
- Department whitelists
- Compliance audit trails

**Features:**
- ✅ SIWE for employee onboarding
- ✅ Allowlist for org structure
- ✅ Rate limiting for resource protection
- ✅ Comprehensive audit logging

---

## Technical Capabilities

### **Scalability**

**Performance Metrics:**
- Policy Evaluation: <500ms
- Cache Hit: <5ms
- Database Query: <50ms
- JWT Verification: <1ms
- Health Check: 2ms

**Scaling Characteristics:**
- ✅ Stateless API (horizontal scaling)
- ✅ Database-backed storage
- ✅ Cached blockchain queries (80%+ hit rate)
- ✅ Connection pooling
- ✅ Rate limiting for resource protection

---

### **Security**

**Authentication Security:**
- ✅ SIWE (EIP-4361) compliant
- ✅ Signature verification (EIP-191)
- ✅ Nonce replay prevention
- ✅ JWT signing (HS256)
- ✅ Token expiration

**Authorization Security:**
- ✅ Fail-closed policy model
- ✅ Blockchain state verification
- ✅ Address normalization
- ✅ EIP-55 checksum validation
- ✅ Rate limiting

**Operational Security:**
- ✅ API key hashing (SHA256)
- ✅ No sensitive data in logs
- ✅ Audit trail logging
- ✅ Error message sanitization
- ✅ SQL injection prevention

---

### **Reliability**

**Blockchain Integration:**
- ✅ Primary + fallback RPC endpoints
- ✅ Automatic failover
- ✅ Request timeout (5 seconds)
- ✅ Network error handling
- ✅ Fail-closed on errors

**Database Reliability:**
- ✅ Connection pooling
- ✅ Transaction support
- ✅ Automatic migrations
- ✅ Backup support
- ✅ Graceful degradation

---

### **Operability**

**Deployment:**
- ✅ Docker containerization
- ✅ Docker Compose stack
- ✅ Kubernetes ready
- ✅ Environment configuration
- ✅ Health checks

**Monitoring:**
- ✅ Structured logging
- ✅ Request tracking
- ✅ Performance metrics
- ✅ Error tracking
- ✅ Health endpoints

---

## Summary

Gatekeeper provides a comprehensive solution for Web3 authentication and access control with:

**Core Strengths:**
1. **Wallet-Native Auth** - No passwords, uses SIWE
2. **Flexible Access Control** - Multiple rule types with AND/OR logic
3. **Token-Gating** - ERC20/ERC721 support with caching
4. **Production-Ready** - Comprehensive testing and documentation
5. **Developer-Friendly** - OpenAPI docs, REST API, multiple SDKs

**Ideal For:**
- Token-gated communities
- NFT-based access control
- DeFi protocol dashboards
- GameFi and metaverse
- Web3 social platforms
- DAO governance
- Enterprise blockchain

**Quick Start:**
1. Deploy with Docker Compose
2. Define access policies
3. Users connect wallets and sign in
4. System automatically verifies access
5. Grant or deny based on policies

---

**For implementation guides, see:**
- [docs/guides/LOCAL_TESTING.md](../guides/LOCAL_TESTING.md) - Local development
- [docs/guides/INTEGRATION_GUIDE.md](../guides/INTEGRATION_GUIDE.md) - Backend integration
- [docs/api/API.md](../api/API.md) - Complete API reference
- [docs/api/BLOCKCHAIN_RULES_README.md](../api/BLOCKCHAIN_RULES_README.md) - Token-gating details

