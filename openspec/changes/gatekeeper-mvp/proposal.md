# Proposal: Gatekeeper MVP - Wallet-Native Authentication Gateway

## Rationale

Current Web3 applications struggle with implementing production-ready authentication and authorization systems that bridge wallet-native identities with traditional API access patterns. Gatekeeper solves this by providing:

- **Wallet-Native Authentication**: Sign-In With Ethereum (SIWE) for passwordless login
- **Token-Gated Access**: Policy-based authorization using on-chain data (ERC20 balances, NFT ownership, allowlists)
- **API Key Management**: Scoped API keys for programmatic access
- **Production-Ready**: Tests, CI/CD, OpenAPI documentation, and demo frontend

This positions Gatekeeper as a portfolio project demonstrating high-signal skills for:
- Protocol infrastructure teams
- Web3 startups building authentication systems
- RWA projects requiring compliant access control
- Employers seeking senior engineers with Web3 + backend expertise

## Scope

### Phase 1: Core Authentication (Week 1)
- SIWE signature verification
- JWT token minting and validation
- Nonce management
- Authentication middleware
- OpenAPI documentation for auth endpoints

### Phase 2: Policy Engine & Token-Gating (Week 2)
- Policy engine with AND/OR logic
- ERC20 minimum balance checks
- NFT ownership verification
- Allowlist-based access
- Caching layer for blockchain queries
- Policy middleware
- Scoped API key system

### Phase 3: Demo Frontend & Polish (Week 3)
- React + TypeScript frontend
- wagmi + viem integration
- Wallet connection (MetaMask, WalletConnect)
- SIWE sign-in flow
- Protected route demonstrations
- CI/CD pipeline
- Docker deployment

## Out of Scope (Future)

- Multi-chain support (focus on Ethereum mainnet initially)
- Admin dashboard
- Usage analytics
- Rate limiting per API key
- Webhook integrations
- OAuth2 provider functionality

## Impact

### Users
- **Developers**: Can integrate wallet-based auth in minutes using OpenAPI specs and client SDKs
- **End Users**: Passwordless authentication using existing Web3 wallets
- **Protocol Teams**: Production-ready auth system with minimal configuration

### Technical Benefits
- **Security**: Industry-standard SIWE implementation with proper nonce handling
- **Flexibility**: Policy-based access control supports any ERC20/NFT requirement
- **Developer Experience**: OpenAPI docs, Postman collections, example code
- **Observability**: Structured logging, health checks, metrics endpoints

### Career Signal
- Demonstrates Go backend proficiency
- Shows Web3 authentication expertise
- Proves ability to build production systems
- Highlights compliance and security awareness

## Success Criteria

### Backend
- [x] SIWE nonce generation endpoint
- [x] SIWE signature verification endpoint
- [x] JWT minting with configurable expiry
- [x] JWT validation middleware
- [x] Policy engine evaluates rules correctly
- [x] ERC20 balance checking with caching
- [x] NFT ownership verification
- [x] Allowlist rule support
- [x] API key CRUD operations
- [x] API key authentication middleware
- [x] Protected route example
- [x] OpenAPI 3.0 specification
- [x] Auto-generated documentation (Redoc)
- [x] Health check endpoint
- [x] Tests with >80% coverage
- [x] Database migrations

### Frontend
- [x] Wallet connection component
- [x] SIWE sign-in flow
- [x] JWT storage and management
- [x] Protected route example
- [x] Error handling and loading states
- [x] Network switching support

### Infrastructure
- [x] Docker Compose setup
- [x] GitHub Actions CI
- [x] Automated testing
- [x] README with quickstart
- [x] Architecture documentation

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| RPC provider rate limits | High | Implement aggressive caching (5 min TTL), use fallback providers |
| Nonce replay attacks | Critical | Store nonces server-side with TTL, consume on use |
| JWT token theft | High | Short expiry (24h), refresh token flow, secure storage guidance |
| Policy bypass | Critical | Comprehensive tests for policy engine, security audit |
| Blockchain reorgs | Medium | Use finalized blocks, document limitations |

## Dependencies

### Backend
- Go 1.21+
- PostgreSQL 14+
- go-ethereum for ERC20/NFT checks
- spruceid/siwe-go for SIWE verification
- golang-jwt/jwt for JWT handling

### Frontend
- Node.js 20+
- React 18+
- wagmi for Ethereum hooks
- viem for Ethereum interactions
- SIWE library for message creation

### Infrastructure
- Docker & Docker Compose
- GitHub Actions
- Alchemy or Infura RPC endpoint

## Timeline

| Week | Deliverables | Status |
|------|--------------|--------|
| Week 1 | SIWE auth, JWT, OpenAPI docs, basic tests | Planned |
| Week 2 | Policy engine, token-gating, API keys, comprehensive tests | Planned |
| Week 3 | React frontend, Docker setup, CI/CD, polish | Planned |

## Metrics

Track these to validate success:

- **Code Quality**: Test coverage >80%, no critical linter warnings
- **Performance**: Auth flow <500ms, policy evaluation <200ms
- **Documentation**: OpenAPI spec 100% complete, README with working examples
- **Security**: No high/critical vulnerabilities from gosec/slither equivalent
- **Usability**: Demo app working end-to-end in <5 min setup

## Next Steps

1. Review and approve this proposal
2. Create detailed task breakdown
3. Define technical specifications for each component
4. Begin implementation with SIWE authentication
5. Iterate with security reviews at each phase

---

**Approval Required**: [ ] Technical Lead | [ ] Security Review | [ ] Product Owner
