# Tasks: Gatekeeper MVP Implementation

## Phase 1: Project Setup & Core Authentication

### Project Structure
- [ ] Initialize Go module
- [ ] Create standard Go project layout (cmd, internal, api, web)
- [ ] Set up .gitignore for Go and Node
- [ ] Create Makefile with common tasks
- [ ] Set up environment variable management

### Database Setup
- [ ] Create PostgreSQL schema
- [ ] Set up golang-migrate for migrations
- [ ] Create migration: 0001_init_schema.sql
  - Users table (for JWT subjects)
  - API keys table
  - Nonces table (with TTL)
- [ ] Implement database connection with connection pooling
- [ ] Add health check for database connectivity

### Configuration
- [ ] Create config package with environment loading
- [ ] Add config validation on startup
- [ ] Support .env file for development
- [ ] Document all environment variables in README

### Logging
- [ ] Set up structured logging with zap
- [ ] Configure log levels (debug, info, warn, error)
- [ ] Add request ID middleware for traceability
- [ ] Implement logging middleware for HTTP requests

### SIWE Authentication - Backend
- [ ] Install spruceid/siwe-go dependency
- [ ] Create auth/siwe.go with SIWEService
- [ ] Implement GenerateNonce() function
  - Generate cryptographically secure nonce
  - Store nonce in database/cache with 5min TTL
- [ ] Implement VerifyMessage(message, signature) function
  - Parse SIWE message
  - Validate domain matches
  - Check nonce exists and hasn't been used
  - Verify signature cryptographically
  - Validate expiration time
  - Consume nonce (prevent replay)
- [ ] Create nonce cleanup background job

### JWT Token Management
- [ ] Install golang-jwt/jwt dependency
- [ ] Create auth/jwt.go with JWT functions
- [ ] Implement GenerateJWT(address, scopes, expiry)
  - Create claims with wallet address as subject
  - Include scopes array
  - Add standard claims (iat, exp, nbf)
  - Sign with HMAC SHA-256
- [ ] Implement VerifyJWT(token) function
  - Parse and validate token
  - Check expiration
  - Return claims
- [ ] Create JWT middleware for protected routes
  - Extract Bearer token from Authorization header
  - Verify and parse JWT
  - Add claims to request context
  - Handle errors with 401 responses

### HTTP Handlers - Authentication
- [ ] Create http/handlers/auth.go
- [ ] Implement GET /auth/siwe/nonce handler
  - Generate nonce
  - Return JSON: {"nonce": "..."}
- [ ] Implement POST /auth/siwe/verify handler
  - Accept: {"message": "...", "signature": "..."}
  - Verify SIWE signature
  - Generate JWT on success
  - Return: {"token": "...", "address": "...", "expiresAt": "..."}
- [ ] Add request validation
- [ ] Add error handling with proper status codes

### Testing - Phase 1
- [ ] Write unit tests for SIWE verification
  - Valid signature passes
  - Invalid signature fails
  - Expired message fails
  - Nonce reuse fails
- [ ] Write unit tests for JWT generation/verification
- [ ] Write integration tests for auth endpoints
- [ ] Add table-driven tests for edge cases
- [ ] Achieve >80% coverage for auth package

---

## Phase 2: Policy Engine & Token-Gating

### Policy Engine Core
- [ ] Create policy/engine.go
- [ ] Define policy structures
  - Rule interface
  - RoutePolicy struct
  - RuleType enum (has_scope, in_allowlist, erc20_min_balance, erc721_owner)
- [ ] Implement Evaluate(policy, address, scopes) function
  - Support AND logic
  - Support OR logic
  - Return boolean result
- [ ] Create policy/rules.go with rule implementations
  - hasScope(scopes, required)
  - inAllowlist(address, addresses)
- [ ] Load policies from JSON configuration file
- [ ] Add policy validation on startup

### Blockchain Integration
- [ ] Create chain/ethclient.go
- [ ] Set up go-ethereum client connection
- [ ] Implement connection pooling for RPC
- [ ] Add fallback RPC provider support
- [ ] Create helper functions for contract calls
- [ ] Handle network errors gracefully

### ERC20 Balance Checking
- [ ] Create chain/erc20.go
- [ ] Implement CheckERC20Balance(chainID, token, address, minBalance)
  - Load ERC20 ABI
  - Call balanceOf(address)
  - Compare with minimum required
  - Return boolean
- [ ] Add caching layer (5 min TTL)
- [ ] Handle decimal conversion properly
- [ ] Add timeout for RPC calls

### NFT Ownership Verification
- [ ] Create chain/erc721.go
- [ ] Implement CheckERC721Ownership(chainID, contract, tokenID, address)
  - Load ERC721 ABI
  - Call ownerOf(tokenID)
  - Compare with address
  - Return boolean
- [ ] Add caching for ownership checks
- [ ] Handle non-existent tokens gracefully

### Caching Layer
- [ ] Create cache/cache.go
- [ ] Implement in-memory cache with TTL
  - Set(key, value, ttl)
  - Get(key) (value, ok)
  - Cleanup expired entries
- [ ] Add cache metrics (hits/misses)
- [ ] Make cache size configurable

### Policy Middleware
- [ ] Create http/middleware/policy_gate.go
- [ ] Implement PolicyGate(policies) middleware
  - Extract user claims from context
  - Find matching policy for route
  - Evaluate policy rules
  - Return 403 if policy fails
  - Allow through if policy passes
- [ ] Add logging for policy decisions
- [ ] Add metrics for policy evaluations

### API Key Management
- [ ] Create auth/apikeys.go
- [ ] Implement GenerateAPIKey(userID, name, scopes, expiry)
  - Generate secure random key
  - Hash key for storage (bcrypt)
  - Store in database
  - Return plain key (only time visible)
- [ ] Implement VerifyAPIKey(key, stored) function
  - Check expiration
  - Verify hash
  - Return scopes
- [ ] Create http/handlers/keys.go
  - GET /keys - List user's API keys
  - POST /keys - Create new API key
  - DELETE /keys/{id} - Revoke API key
- [ ] Add API key authentication middleware
  - Check X-API-Key header
  - Verify key
  - Add scopes to context

### Protected Route Example
- [ ] Create http/handlers/demo.go
- [ ] Implement GET /alpha/data handler
  - Protected by JWT or API key middleware
  - Protected by policy gate middleware
  - Returns sample data
  - Demonstrates successful authorization
- [ ] Create policy configuration example
  - Require ERC20 minimum balance OR
  - In allowlist OR
  - Has specific scope

### Testing - Phase 2
- [ ] Write unit tests for policy engine
  - AND logic works correctly
  - OR logic works correctly
  - Each rule type functions properly
- [ ] Write unit tests for blockchain checking
  - Mock RPC responses
  - Test caching behavior
  - Handle errors gracefully
- [ ] Write integration tests for protected routes
  - Access granted with valid token + policy
  - Access denied without token
  - Access denied with token but no policy match
- [ ] Test API key CRUD operations
- [ ] Achieve >80% coverage for policy package

---

## Phase 3: API Documentation & Frontend

### OpenAPI Specification
- [ ] Create api/openapi.yaml
- [ ] Document all endpoints
  - /health
  - /auth/siwe/nonce
  - /auth/siwe/verify
  - /keys (GET, POST, DELETE)
  - /alpha/data
- [ ] Define all schemas
  - Error response
  - SIWE verify request/response
  - API key models
- [ ] Define security schemes
  - bearerAuth (JWT)
  - apiKeyAuth (X-API-Key)
- [ ] Add examples for all requests/responses
- [ ] Validate spec with openapi-validator

### API Documentation Serving
- [ ] Create http/handlers/docs.go
- [ ] Embed openapi.yaml in binary
- [ ] Implement GET /openapi.yaml handler
- [ ] Implement GET /docs handler with Redoc
- [ ] Test documentation renders correctly

### Frontend Setup
- [ ] Create web/ directory
- [ ] Initialize Vite + React + TypeScript project
- [ ] Install dependencies
  - wagmi
  - viem
  - @tanstack/react-query
  - siwe
- [ ] Configure Vite dev proxy to backend
- [ ] Set up TypeScript strict mode
- [ ] Configure ESLint and Prettier

### wagmi Configuration
- [ ] Create src/config/wagmi.ts
- [ ] Configure chains (mainnet, sepolia)
- [ ] Set up connectors (MetaMask, WalletConnect)
- [ ] Configure transports (RPC providers)
- [ ] Add WagmiProvider to app

### Wallet Connection
- [ ] Create src/components/ConnectButton.tsx
- [ ] Implement wallet connection UI
  - Show connect options when disconnected
  - Show address and disconnect when connected
  - Format address (0x1234...5678)
  - Handle connection errors
- [ ] Add network switching dropdown
- [ ] Style with minimal CSS

### SIWE Authentication Flow
- [ ] Create src/hooks/useAuth.ts
- [ ] Implement signIn() function
  - Get nonce from backend
  - Create SIWE message
  - Sign message with wallet
  - Send to backend for verification
  - Store JWT in localStorage
  - Update auth state
- [ ] Implement signOut() function
  - Clear localStorage
  - Reset auth state
- [ ] Implement getAuthHeaders() helper
  - Return Authorization header with JWT

### Protected Route Demo
- [ ] Create src/components/ProtectedRoute.tsx
- [ ] Implement UI for calling protected endpoint
  - Show "Sign In" button if not authenticated
  - Show "Call Protected Route" button if authenticated
  - Display loading state during request
  - Show success response or error message
- [ ] Handle 401 (not authenticated)
- [ ] Handle 403 (policy failed)
- [ ] Show meaningful error messages

### Frontend Polish
- [ ] Add error boundaries
- [ ] Implement loading states for all async operations
- [ ] Add transaction feedback (pending, success, error)
- [ ] Test on multiple browsers
- [ ] Ensure mobile responsive
- [ ] Add basic accessibility (ARIA labels)

### Testing - Phase 3
- [ ] Write component tests
  - ConnectButton shows correct states
  - ProtectedRoute handles auth correctly
- [ ] Write integration tests
  - Full SIWE flow works end-to-end
  - Protected route call succeeds with auth
- [ ] Manual testing checklist
  - Wallet connection works
  - SIWE sign-in works
  - JWT stored correctly
  - Protected route accessible after auth
  - Logout clears state

---

## Phase 4: Infrastructure & Deployment

### Docker Configuration
- [ ] Create Dockerfile for Go backend
  - Multi-stage build
  - Use distroless base image
  - Copy binary only
  - Expose port 8080
- [ ] Create Dockerfile for frontend
  - Build static assets
  - Serve with nginx
- [ ] Create docker-compose.yml
  - PostgreSQL service
  - Backend service
  - Frontend service
  - Network configuration
  - Volume mounts
- [ ] Create .dockerignore
- [ ] Test local Docker Compose deployment

### CI/CD Pipeline
- [ ] Create .github/workflows/ci.yaml
- [ ] Add backend CI jobs
  - go vet
  - go test -race -cover
  - golangci-lint
  - gosec security scan
- [ ] Add frontend CI jobs
  - npm run lint
  - npm run type-check
  - npm test
- [ ] Add Docker build job
- [ ] Configure to run on pull requests
- [ ] Add status badge to README

### Environment Configuration
- [ ] Create .env.example file
- [ ] Document all environment variables
  - DATABASE_URL
  - JWT_SECRET
  - ETHEREUM_RPC
  - PORT
  - ENVIRONMENT
- [ ] Add validation for required variables
- [ ] Create separate configs for dev/staging/prod

### Database Migrations
- [ ] Ensure all migrations are reversible
- [ ] Test migration up/down
- [ ] Add migration to Docker entrypoint
- [ ] Document migration process

### Health Checks & Monitoring
- [ ] Implement GET /health endpoint
  - Check database connectivity
  - Check RPC provider connectivity
  - Return version info
  - Return status: ok/degraded/down
- [ ] Add structured logging for errors
- [ ] Add request tracing
- [ ] Document observability setup

---

## Phase 5: Documentation & Polish

### README
- [ ] Create comprehensive README.md
- [ ] Add project description
- [ ] Add features list
- [ ] Add quickstart guide
  - Prerequisites
  - Installation steps
  - Running locally
  - Testing
- [ ] Add architecture overview
- [ ] Add API documentation link
- [ ] Add screenshots/demo video
- [ ] Add troubleshooting section
- [ ] Add contributing guidelines
- [ ] Add license

### Architecture Documentation
- [ ] Create docs/ARCHITECTURE.md
- [ ] Explain system design
- [ ] Add architecture diagrams
  - C4 Context diagram
  - C4 Container diagram
  - Sequence diagram for SIWE flow
  - Sequence diagram for policy evaluation
- [ ] Document design decisions
- [ ] Explain security considerations

### Code Documentation
- [ ] Add package documentation comments
- [ ] Add function documentation for public APIs
- [ ] Document complex algorithms
- [ ] Add inline comments for non-obvious code
- [ ] Generate godoc documentation

### Security Review
- [ ] Run gosec and fix critical issues
- [ ] Review SIWE implementation against spec
- [ ] Review JWT handling for best practices
- [ ] Check for hardcoded secrets
- [ ] Verify input validation everywhere
- [ ] Test auth bypass scenarios
- [ ] Document security assumptions

### Performance Testing
- [ ] Load test auth endpoints
  - Measure requests/second
  - Check response times under load
  - Verify no memory leaks
- [ ] Test policy engine performance
  - Benchmark rule evaluation
  - Verify cache effectiveness
- [ ] Optimize slow queries
- [ ] Profile with pprof

### Final Checklist
- [ ] All tests passing
- [ ] Test coverage >80%
- [ ] No linter warnings
- [ ] No security vulnerabilities
- [ ] Documentation complete
- [ ] Docker Compose works
- [ ] CI/CD pipeline green
- [ ] Demo app fully functional
- [ ] README has working examples
- [ ] Code reviewed

---

## Deployment Checklist

### Pre-Deployment
- [ ] Set production environment variables
- [ ] Generate strong JWT secret
- [ ] Configure production RPC provider (Alchemy/Infura)
- [ ] Set up production database
- [ ] Run database migrations
- [ ] Configure CORS for production domain
- [ ] Set up monitoring and alerts

### Deployment
- [ ] Deploy database (managed PostgreSQL)
- [ ] Deploy backend (container platform)
- [ ] Deploy frontend (Vercel/Netlify/CDN)
- [ ] Verify health checks pass
- [ ] Test SIWE flow in production
- [ ] Test policy enforcement
- [ ] Monitor logs for errors

### Post-Deployment
- [ ] Create demo accounts/keys
- [ ] Write blog post explaining architecture
- [ ] Share on Twitter/LinkedIn
- [ ] Add to portfolio site
- [ ] Update resume with project

---

**Task Summary**:
- **Total Tasks**: 150+
- **Phase 1**: 40 tasks (Project setup, SIWE, JWT)
- **Phase 2**: 45 tasks (Policy engine, token-gating, API keys)
- **Phase 3**: 35 tasks (OpenAPI, Frontend)
- **Phase 4**: 20 tasks (Docker, CI/CD)
- **Phase 5**: 25 tasks (Documentation, Security, Polish)
