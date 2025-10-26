# Design: Gatekeeper MVP Technical Architecture

## System Architecture

### High-Level Overview

```
┌─────────────────┐      ┌──────────────────┐      ┌─────────────────┐
│                 │      │                  │      │                 │
│  React Frontend │─────▶│  Go HTTP Server  │─────▶│   PostgreSQL    │
│  (wagmi/viem)   │      │  (Middleware)    │      │   Database      │
│                 │      │                  │      │                 │
└─────────────────┘      └──────────────────┘      └─────────────────┘
                                  │
                                  │
                                  ▼
                         ┌──────────────────┐
                         │                  │
                         │  Ethereum RPC    │
                         │  (Alchemy/Infura)│
                         │                  │
                         └──────────────────┘
```

### Technology Stack

**Backend**
- **Language**: Go 1.21+
- **HTTP Router**: gorilla/mux
- **Database**: PostgreSQL 14+
- **Migrations**: golang-migrate/migrate
- **Logging**: uber-go/zap
- **JWT**: golang-jwt/jwt v5
- **SIWE**: spruceid/siwe-go
- **Ethereum**: go-ethereum

**Frontend**
- **Framework**: React 18 + TypeScript
- **Build Tool**: Vite
- **Web3 Library**: wagmi + viem
- **State**: React Query (@tanstack/react-query)
- **SIWE**: siwe npm package

**Infrastructure**
- **Container**: Docker + Docker Compose
- **CI/CD**: GitHub Actions
- **API Docs**: OpenAPI 3.0 + Redoc

## Component Design

### 1. Authentication Service

**Responsibilities**:
- SIWE nonce generation and verification
- JWT token minting and validation
- Nonce lifecycle management

**Key Design Decisions**:

1. **Nonce Storage**: Use PostgreSQL with TTL
   - **Why**: Prevents replay attacks, ensures consistency across instances
   - **Alternative considered**: In-memory cache (rejected due to multi-instance issues)

2. **JWT Signing**: HMAC SHA-256
   - **Why**: Simple, fast, sufficient for single-server architecture
   - **Alternative**: RSA (overkill for MVP, adds complexity)

3. **Token Expiry**: 24 hours
   - **Why**: Balance between security and UX
   - **Future**: Add refresh token flow for longer sessions

**Implementation Notes**:
```go
// auth/siwe.go
type SIWEService struct {
    db     *sql.DB
    domain string
}

// auth/jwt.go
type Claims struct {
    Address string   `json:"address"`
    Scopes  []string `json:"scopes"`
    jwt.RegisteredClaims
}
```

### 2. Policy Engine

**Responsibilities**:
- Load and validate policy configurations
- Evaluate access rules with AND/OR logic
- Interface with blockchain for on-chain data

**Key Design Decisions**:

1. **Policy Format**: JSON configuration file
   - **Why**: Easy to modify without code changes, version controllable
   - **Alternative**: Database storage (future enhancement for dynamic policies)

2. **Rule Types**:
   - `has_scope`: JWT scope checking
   - `in_allowlist`: Address allowlist
   - `erc20_min_balance`: Minimum token balance
   - `erc721_owner`: NFT ownership

3. **Evaluation Logic**: Short-circuit evaluation
   - **AND**: Stop on first false
   - **OR**: Stop on first true
   - **Why**: Performance optimization, reduces unnecessary RPC calls

4. **Caching Strategy**: 5-minute TTL for blockchain data
   - **Why**: Balance freshness with RPC cost
   - **Risk**: Stale data during rapid balance changes (acceptable for MVP)

**Implementation Notes**:
```go
// policy/engine.go
type Engine struct {
    client *ethclient.Client
    cache  *Cache
}

func (e *Engine) Evaluate(ctx context.Context, policy RoutePolicy,
    address string, scopes []string) (bool, error) {
    // Implements AND/OR logic with rule evaluation
}
```

### 3. Blockchain Integration

**Responsibilities**:
- Manage RPC connections
- Execute smart contract calls
- Cache results

**Key Design Decisions**:

1. **RPC Provider**: Primary + Fallback
   - **Primary**: Alchemy (fast, reliable)
   - **Fallback**: Infura (redundancy)
   - **Why**: High availability, automatic failover

2. **Connection Pooling**: HTTP client with persistent connections
   - **Why**: Reduce connection overhead
   - **Configuration**: Max 10 connections per provider

3. **Timeout**: 5 seconds per RPC call
   - **Why**: Prevent hanging requests
   - **Behavior**: Fail to fallback on timeout

4. **Contract ABIs**: Embedded in binary
   - **Why**: No external dependencies at runtime
   - **Method**: `go:embed` for ABI JSON files

**Implementation Notes**:
```go
// chain/ethclient.go
type Client struct {
    primary  *ethclient.Client
    fallback *ethclient.Client
}

// chain/erc20.go
//go:embed abi/erc20.json
var erc20ABI string

func CheckBalance(ctx context.Context, token common.Address,
    owner common.Address) (*big.Int, error) {
    // Implements balance check with caching
}
```

### 4. API Key System

**Responsibilities**:
- Generate secure API keys
- Hash and store keys
- Authenticate requests via API key

**Key Design Decisions**:

1. **Key Generation**: Cryptographically secure random (32 bytes)
   - **Format**: `gk_` prefix + base64 encoded
   - **Why**: Identifiable, URL-safe, high entropy

2. **Storage**: bcrypt hashed
   - **Why**: Protect keys if database compromised
   - **Cost**: 10 rounds (balance security and performance)

3. **Scopes**: Same as JWT scopes
   - **Why**: Consistent authorization model
   - **Flexibility**: Different keys can have different scopes

4. **Revocation**: Delete from database
   - **Why**: Simple, immediate effect
   - **Future**: Consider soft delete for audit trail

**Implementation Notes**:
```go
// auth/apikeys.go
type APIKey struct {
    ID        string
    UserID    string
    Name      string
    KeyHash   string
    Scopes    []string
    CreatedAt time.Time
    ExpiresAt *time.Time
}

func GenerateAPIKey(userID, name string, scopes []string,
    expiry *time.Duration) (*APIKey, string, error) {
    // Returns both stored object and plain key (one time only)
}
```

### 5. Middleware Stack

**Request Flow**:
```
HTTP Request
    ↓
Logging Middleware (add request ID, log request)
    ↓
Recovery Middleware (catch panics)
    ↓
CORS Middleware (add CORS headers)
    ↓
Auth Middleware (JWT or API key)
    ↓
Policy Gate Middleware (evaluate access rules)
    ↓
Rate Limit Middleware (prevent abuse)
    ↓
Handler
    ↓
HTTP Response
```

**Key Design Decisions**:

1. **Middleware Ordering**: Critical for correctness
   - Logging first: capture all requests
   - Recovery early: prevent crashes
   - Auth before policy: need identity for rules

2. **Context Propagation**: Use request context
   - Store: User ID, address, scopes, request ID
   - **Why**: Type-safe, request-scoped, no global state

3. **Error Handling**: Consistent across middleware
   - All middleware returns errors in standard format
   - 401: Authentication failed
   - 403: Authorization failed
   - 429: Rate limited

**Implementation Notes**:
```go
// http/middleware/stack.go
func NewStack() []Middleware {
    return []Middleware{
        LoggingMiddleware,
        RecoveryMiddleware,
        CORSMiddleware,
        AuthMiddleware,
        PolicyGateMiddleware,
        RateLimitMiddleware,
    }
}
```

## Database Schema

### Tables

**users** (minimal for JWT subject tracking)
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    address TEXT NOT NULL UNIQUE,  -- Ethereum address
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_address ON users(address);
```

**nonces**
```sql
CREATE TABLE nonces (
    nonce TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_nonces_expires_at ON nonces(expires_at);
```

**api_keys**
```sql
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    key_hash TEXT NOT NULL,
    scopes TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP,
    last_used_at TIMESTAMP
);

CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_expires_at ON api_keys(expires_at);
```

**Design Decisions**:

1. **UUID vs Integer IDs**: UUIDs
   - **Why**: Non-sequential, harder to guess, distributed-friendly

2. **Address Storage**: TEXT with checksum
   - **Why**: Consistent format, case-insensitive comparison

3. **Scopes as Array**: PostgreSQL TEXT[]
   - **Why**: Native support, efficient querying

4. **Cascading Deletes**: ON DELETE CASCADE for user → api_keys
   - **Why**: Automatic cleanup, maintain referential integrity

## Security Considerations

### 1. Authentication Security

**SIWE Implementation**:
- ✅ Verify signature cryptographically
- ✅ Check domain matches (prevent phishing)
- ✅ Validate expiration time
- ✅ One-time nonce use (prevent replay)
- ✅ Nonce TTL (5 minutes, prevent stale attacks)

**JWT Security**:
- ✅ HMAC SHA-256 signing
- ✅ Secret key from environment (not hardcoded)
- ✅ Token expiration (24 hours)
- ✅ Not-before time validation
- ⚠️ Future: Refresh tokens, token revocation

### 2. Input Validation

**All Inputs Validated**:
- JSON body structure
- Field types and formats
- String lengths
- Address formats (checksum validation)
- Scope names (allowlist)

**SQL Injection Prevention**:
- ✅ Prepared statements only
- ✅ No string concatenation in queries
- ✅ ORM-like helpers for safety

### 3. Rate Limiting

**Per-IP Rate Limits**:
- Nonce generation: 60/minute
- Signature verification: 20/minute
- API calls: 100/minute (authenticated), 20/minute (unauthenticated)

**Why**: Prevent brute force, DoS attacks

### 4. Data Protection

**Secrets Management**:
- JWT secret: Environment variable, rotatable
- API keys: Bcrypt hashed, never logged
- Database credentials: Environment variable

**Logging**:
- ❌ Never log JWT tokens
- ❌ Never log API keys
- ❌ Never log full user signatures
- ✅ Log request IDs, addresses, outcomes

## Performance Considerations

### 1. Caching Strategy

**What to Cache**:
- ERC20 balances: 5 min TTL
- NFT ownership: 5 min TTL
- Not cached: JWT validation (crypto is fast)

**Cache Implementation**:
- In-memory map with mutex
- Background cleanup every minute
- Max 10,000 entries (LRU eviction)

### 2. Database Optimization

**Indexes**:
- users.address (frequent lookups)
- api_keys.user_id (list user's keys)
- nonces.expires_at (cleanup queries)

**Connection Pooling**:
- Max 25 connections
- Max 5 idle connections
- 5-minute max lifetime

### 3. RPC Optimization

**Minimize Calls**:
- Cache aggressively
- Batch calls when possible (future)
- Short-circuit policy evaluation

**Connection Management**:
- HTTP/2 when supported
- Persistent connections
- Timeout: 5 seconds

## Deployment Architecture

### Development
```
Docker Compose:
- postgres:14-alpine
- gatekeeper-backend:latest (local build)
- gatekeeper-frontend:latest (Vite dev server)
```

### Production (Future)
```
- Frontend: Vercel/Netlify (static hosting)
- Backend: Fly.io / Railway (container hosting)
- Database: Managed PostgreSQL (RDS/Neon/Supabase)
- RPC: Alchemy + Infura
```

## Testing Strategy

### Unit Tests (>80% coverage)
- auth/siwe.go: SIWE verification logic
- auth/jwt.go: JWT generation/validation
- policy/engine.go: Policy evaluation logic
- chain/erc20.go: Balance checking (mocked RPC)

### Integration Tests
- Full auth flow (nonce → verify → JWT)
- Protected route access
- Policy enforcement
- API key CRUD

### E2E Tests (Future)
- Frontend → Backend → Blockchain
- Using local testnet (Anvil)

## Monitoring & Observability

### Metrics (Future)
- Request count per endpoint
- Response time percentiles (p50, p95, p99)
- Auth success/failure rates
- Policy evaluation latency
- RPC call counts and latency
- Cache hit rate

### Logging
- Structured JSON logs (zap)
- Log levels: DEBUG, INFO, WARN, ERROR
- Request ID in all logs
- Fields: timestamp, level, message, context

### Health Checks
- `/health` endpoint
- Checks: database, RPC providers
- Used by: Load balancers, monitoring

## Future Enhancements

### Phase 2+
- [ ] Refresh token flow
- [ ] Multi-chain support (Polygon, Arbitrum, etc.)
- [ ] Admin dashboard for policy management
- [ ] Usage analytics per API key
- [ ] Webhook integrations
- [ ] OAuth2 provider functionality
- [ ] SDK generation from OpenAPI spec
- [ ] Metrics dashboard (Prometheus + Grafana)

### Scalability
- [ ] Horizontal scaling (load balancer + multiple instances)
- [ ] Redis for distributed caching
- [ ] Message queue for async tasks
- [ ] Read replicas for database

---

**Design Approved**: [ ] Tech Lead | [ ] Security | [ ] DevOps

**Ready for Implementation**: ✅
