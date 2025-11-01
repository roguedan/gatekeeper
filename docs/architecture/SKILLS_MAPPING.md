# Claude Skills Mapping for Gatekeeper Project

## Skills Used So Far ✅

### Phase 1: Core Authentication

#### 🔵 test-driven-development
**Status**: ✅ ACTIVELY USED
**Application**: Full RED-GREEN-REFACTOR cycle
- Task 1.2: Configuration management (11 tests)
- Task 2.3: Logging (10 tests)
- Task 3.1: SIWE nonce service (11 tests)
- Task 3.2: JWT service (10 tests)
- Task 4.1: JWT middleware (6 tests)
- Task 4.2: Auth handlers (11 tests)

**Reference**: `.claude/skills/test-driven-development/SKILL.md`
- Go testing patterns with testify
- Table-driven test approaches
- Test structure and organization

#### 🔵 go-backend-development
**Status**: ✅ ACTIVELY USED
**Application**: Core Go patterns and structure
- Project structure (cmd, internal, deployments)
- Package organization
- Configuration management patterns
- Database connection pooling
- Middleware composition
- Error handling
- Context usage

**Reference**: `.claude/skills/go-backend-development/SKILL.md`
- Standard Go project layout
- Connection pooling configuration
- Middleware patterns
- HTTP handler design

#### 🔵 web3-authentication-authorization
**Status**: ✅ ACTIVELY USED
**Application**: SIWE and JWT implementation
- SIWE nonce generation with entropy requirements
- JWT claims structure (address, scopes)
- Token-based authorization

**Reference**: `.claude/skills/web3-authentication-authorization/SKILL.md`
- SIWE best practices
- JWT security patterns
- Scope-based authorization fundamentals

### Phase 2 Part 1: Policy Engine Foundation

#### 🔵 test-driven-development
**Status**: ✅ ACTIVELY USED (continued)
**Application**: Policy engine with 65 tests
- Task 5.1: Policy types (14 tests)
- Task 5.2: Policy loader (20 tests)
- Task 5.3: Policy manager (16 tests)
- Task 6.1: AND/OR evaluation (15 tests including rule evaluation)

**Coverage**: 79.5% in policy package

#### 🔵 go-backend-development
**Status**: ✅ ACTIVELY USED (continued)
**Application**: Policy package design
- Interface design (Rule interface)
- Struct organization
- Thread-safe operations (sync.RWMutex)
- Configuration loading patterns
- Type validation

---

## Skills to Use for Next Tasks ⏳

### Phase 2 Part 2: Blockchain Integration & Caching

#### Task 7.3: RPC Provider Management
**Required Skills**:
1. **go-backend-development** - HTTP client patterns, connection pooling, error handling
2. **test-driven-development** - Mock RPC responses, failure scenarios
3. **blockchain-testing** - RPC call patterns, test fixtures

**Reference**:
- `.claude/skills/go-backend-development/SKILL.md` - Connection management
- `.claude/skills/blockchain-testing/SKILL.md` - Mock RPC responses

**Key Patterns**:
- HTTP client with timeout
- Failover logic
- Connection reuse

#### Task 7.1 & 7.2: ERC20/NFT Rules
**Required Skills**:
1. **go-backend-development** - Go patterns and structure
2. **test-driven-development** - Mock blockchain responses
3. **web3-authentication-authorization** - Token-gating patterns
4. **blockchain-testing** - Smart contract interaction patterns

**Reference**:
- `.claude/skills/blockchain-testing/SKILL.md` - Contract testing patterns
- `.claude/skills/web3-authentication-authorization/SKILL.md` - Token verification

**Key Patterns**:
- ERC20 balanceOf() calls
- ERC721 ownerOf() calls
- big.Int handling for amounts
- Decimal normalization

#### Task 8.1-8.3: Caching System
**Required Skills**:
1. **go-backend-development** - In-memory data structures, concurrency
2. **test-driven-development** - TTL testing, cache invalidation

**Key Patterns**:
- TTL-based cache entries
- Goroutine-safe operations
- Cleanup routines

#### Task 9.1-9.3: Middleware & Logging & Integration Tests
**Required Skills**:
1. **go-backend-development** - Middleware patterns, HTTP testing
2. **test-driven-development** - Integration testing patterns
3. **technical-documentation** - API documentation

**Key Patterns**:
- Middleware composition
- Request/response flow
- End-to-end testing
- Structured logging

---

## Skills for API Documentation ⏳

### OpenAPI / API Documentation
**Requirement**: REQ-API-001 (OpenAPI documentation)
**Skill**: api-design-openapi
**Reference**: `.claude/skills/api-design-openapi/SKILL.md`

**What We Need**:
- OpenAPI 3.0 spec for:
  - GET /auth/siwe/nonce
  - POST /auth/siwe/verify
  - Future endpoints (policy-protected routes)
- Schema definitions for requests/responses
- Error response documentation

---

## Implementation Strategy Going Forward

### For Each Task:

1. **Identify Required Skills**
   - Check PHASE_2_EXECUTION_PLAN.md for Primary Skill assignments
   - Map to available `.claude/skills/` files

2. **Read Skill Documentation**
   - Reference relevant `.claude/skills/SKILL.md` files
   - Extract code patterns and best practices

3. **Apply TDD Methodology** (for most tasks)
   - Use test-driven-development skill
   - RED → GREEN → BLUE cycle
   - Verify tests pass

4. **Review Code Quality**
   - Consider using code-reviewer skill for significant changes
   - Ensure patterns match skill recommendations

5. **Document as We Go**
   - API changes documented in OpenAPI spec
   - Architecture updates to C4 diagrams
   - Technical decisions in design.md

---

## Skill Trigger Keywords for Future Prompts

When continuing to build, include relevant keywords to ensure skills are invoked:

### For RPC Provider Task:
```
"Build RPC provider management with failover and connection pooling"
→ Triggers: go-backend-development, blockchain-testing
```

### For ERC20/NFT Rules:
```
"Implement ERC20 balance checking and NFT ownership verification"
→ Triggers: go-backend-development, web3-authentication-authorization, blockchain-testing
```

### For Caching:
```
"Create TTL-based caching system for blockchain data"
→ Triggers: go-backend-development, test-driven-development
```

### For Middleware Integration:
```
"Implement policy enforcement middleware with comprehensive logging"
→ Triggers: go-backend-development, test-driven-development, technical-documentation
```

### For OpenAPI:
```
"Document API endpoints with OpenAPI specification"
→ Triggers: api-design-openapi, technical-documentation
```

---

## Confirmation Checklist ✅

- [x] Using test-driven-development for Phase 1 & 2a (59 + 65 tests)
- [x] Using go-backend-development for all backend code
- [x] Using web3-authentication-authorization for SIWE/JWT patterns
- [x] Mapped skills to OpenSpec requirements
- [x] Planning to use blockchain-testing for RPC mocking
- [x] Will use api-design-openapi for API documentation
- [x] Will use technical-documentation for comprehensive docs
- [x] All skills referenced from `experimental-claude-skills` repo

---

## Quick Reference

| Phase | Tasks | Primary Skills | TDD |
|-------|-------|-----------------|-----|
| 1 | All 9 tasks | `test-driven-development`, `go-backend-development`, `web3-auth` | YES |
| 2a | Tasks 5.1-6.3 | `test-driven-development`, `go-backend-development` | YES |
| 2b | Tasks 7.1-7.4 | `go-backend`, `test-driven-dev`, `blockchain-testing`, `web3-auth` | YES |
| 2c | Tasks 8.1-8.3 | `go-backend-development`, `test-driven-development` | YES |
| 2d | Tasks 9.1-9.3 | `go-backend`, `test-driven-dev`, `api-design-openapi`, `tech-docs` | YES |

---

## Summary

**We are CONFIRMED to be building with relevant skills**:

✅ **Phase 1 Complete**: All tasks used appropriate skills
✅ **Phase 2a Complete**: Policy foundation used appropriate skills
✅ **Phase 2b-2d Ready**: Skills identified and mapped
✅ **Keywords Ready**: Will use trigger keywords for skill invocation

**Current Approach**:
- Leading with `test-driven-development` for all core logic
- Using `go-backend-development` for implementation patterns
- Pulling from `web3-authentication-authorization` for Web3 patterns
- Ready to add `blockchain-testing` for RPC interactions
- Will use `api-design-openapi` when documenting endpoints
