# Gatekeeper Phase 2 Implementation Guide

## Three Claude Skills Created for Phase 2

This guide coordinates three specialized Claude skills to implement the complete Phase 2 (Policy Engine & Token-Gating).

### Skill 1: Build Database Repositories
**File:** `.claude/commands/build-database-repositories.md`

**Implements:**
- User repository (create, read, update, delete)
- API Key repository (generate, validate, revoke, list)
- Allowlist repository (CRUD, address management)

**Output:** ~600 lines of production-ready Go code
**Estimated Time:** 4-6 hours
**Dependencies:** Database connection pool already in place

**Key Features:**
- Address validation and normalization
- API key hashing (SHA256)
- Allowlist batch operations
- Thread-safe database operations
- >85% test coverage

---

### Skill 2: Build API Key System
**File:** `.claude/commands/build-api-key-system.md`

**Implements:**
- API Key HTTP handlers (Create, List, Revoke)
- API Key validation middleware
- HTTP route handlers
- OpenAPI documentation
- Security best practices

**Output:** ~400 lines of HTTP handler code + middleware
**Estimated Time:** 3-4 hours
**Dependencies:** Database repositories (from Skill 1)

**Key Features:**
- Secure key generation (32-byte cryptographic random)
- Support for both X-API-Key and Bearer token formats
- Ownership verification (prevent cross-user access)
- Last-used tracking
- Expiration enforcement
- Clear error messages

---

### Skill 3: Build Policy Blockchain Rules
**File:** `.claude/commands/build-policy-blockchain-rules.md`

**Implements:**
- ERC20 minimum balance rule evaluation
- ERC721 token ownership rule evaluation
- Blockchain helper utilities
- Caching integration
- Error handling and logging

**Output:** ~350 lines of blockchain rule code
**Estimated Time:** 3-4 hours
**Dependencies:** RPC provider and cache already in place

**Key Features:**
- ERC20 balanceOf() contract calls
- ERC721 ownerOf() contract calls
- TTL-based result caching
- Multi-chain support (via chainID)
- Fail-closed error handling
- RPC timeout and retry logic

---

## Implementation Timeline

### Week 1: Database & API Keys
**Days 1-2:** Database Repositories
- Create User, APIKey, Allowlist repositories
- Write database tests
- Create migrations for allowlists table

**Days 3-4:** API Key System
- Implement HTTP handlers
- Add API key middleware
- Write integration tests
- Update OpenAPI spec

**Days 5-7:** Blockchain Rules
- Complete ERC20 rule evaluation
- Complete ERC721 rule evaluation
- Add caching integration
- Write rule evaluation tests

### Week 2: Integration & Polish
**Days 1-2:** Full Integration Testing
- End-to-end API key workflow
- End-to-end policy evaluation
- Blockchain rule testing with real contracts

**Days 3-4:** Performance & Security
- Load testing with multiple policies
- Security audit (token handling, address validation)
- Cache hit rate optimization

**Days 5-7:** Documentation & Examples
- Update API.md with examples
- Create Postman collection
- Document policy configuration
- TypeScript client examples

---

## Using the Skills

### For Database Repositories:

```bash
# In Claude Code terminal:
/build-database-repositories

# Then describe what you need:
# "Implement the User repository with CRUD operations"
# "Create the API Key repository with hashing"
# "Build the Allowlist repository with batch operations"
```

### For API Key System:

```bash
# In Claude Code terminal:
/build-api-key-system

# Then describe what you need:
# "Implement the CreateAPIKey HTTP handler"
# "Add API key validation middleware"
# "Create tests for API key generation workflow"
```

### For Blockchain Rules:

```bash
# In Claude Code terminal:
/build-policy-blockchain-rules

# Then describe what you need:
# "Implement the ERC20MinBalanceRule.Evaluate method"
# "Complete the ERC721OwnerRule evaluation"
# "Add caching integration for blockchain calls"
```

---

## Critical Path Dependencies

```
Database Repositories
    ↓
API Key System (depends on Repositories)
    ↓
Integration Tests (depends on Repositories + API Key System)

Policy Blockchain Rules (parallel with API Key System)
    ↓
Full Integration Testing (depends on all above)
```

**Recommendation:**
- Start Skill 1 (Database) first
- Once DB repos are done, start Skill 2 (API Keys) and Skill 3 (Blockchain) in parallel
- Combine all three in integration tests

---

## Database Schema Additions Needed

### Allowlists Table
```sql
CREATE TABLE allowlists (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL UNIQUE,
  description TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE allowlist_entries (
  id BIGSERIAL PRIMARY KEY,
  allowlist_id BIGINT NOT NULL REFERENCES allowlists(id) ON DELETE CASCADE,
  address VARCHAR(42) NOT NULL,
  added_at TIMESTAMP NOT NULL DEFAULT NOW(),
  UNIQUE(allowlist_id, address)
);

CREATE INDEX idx_allowlist_entries_address ON allowlist_entries(address);
```

---

## Environment Variables to Add

```bash
# Database
DATABASE_URL=postgresql://user:password@localhost:5432/gatekeeper

# Blockchain (for rules evaluation)
ETHEREUM_RPC=https://eth-mainnet.alchemyapi.io/v2/YOUR_KEY
ETHEREUM_RPC_FALLBACK=https://rpc.ankr.com/eth
CHAIN_ID=1  # 1=mainnet, 5=goerli, 11155111=sepolia

# Caching
CACHE_TTL=300  # seconds

# API Keys
API_KEY_SALT=your-random-salt-string

# Logging
LOG_LEVEL=info
```

---

## Testing Strategy

### Unit Tests (Per Skill)
- Database repository tests (fixtures, rollback)
- API key generation/validation tests
- Blockchain rule evaluation tests (with mock provider)

### Integration Tests
- API key workflow (generate -> list -> revoke)
- Policy evaluation with all rule types
- Blockchain calls with cached results
- User management across features

### End-to-End Tests
- Complete authentication flow + API key issuance
- Token-gating a protected endpoint
- Multi-chain policy evaluation
- Allowlist management

---

## Success Metrics for Phase 2

| Metric | Target | Status |
|--------|--------|--------|
| Code Coverage | >85% | ✅ Required |
| Database Consistency | 100% | ✅ Required |
| API Response Time | <500ms | ✅ Target |
| Cache Hit Rate | >80% | ✅ Target |
| RPC Call Failures (logged, not fatal) | 0 exposed errors | ✅ Required |
| API Key Security | No plaintext keys | ✅ Required |
| Test Pass Rate | 100% | ✅ Required |

---

## Files to Create (By Skill)

### Skill 1 Outputs:
- `internal/store/users.go` (~150 lines)
- `internal/store/api_keys.go` (~250 lines)
- `internal/store/allowlists.go` (~200 lines)
- `internal/store/errors.go` (~30 lines)
- Migration: `004_create_allowlists_table.sql`
- Migration: `005_create_allowlist_entries_table.sql`
- Tests for all repositories

### Skill 2 Outputs:
- `internal/http/api_key_handlers.go` (~250 lines)
- `internal/http/api_key_middleware.go` (~150 lines)
- Tests for handlers and middleware
- Updated `openapi.yaml` with API key endpoints

### Skill 3 Outputs:
- `internal/policy/erc20_rule.go` (~150 lines)
- `internal/policy/erc721_rule.go` (~150 lines)
- Updated `internal/policy/blockchain.go` (add helpers)
- Updated `internal/policy/manager.go` (wire in provider/cache)
- Tests for rule evaluation
- Example `policies.json` with blockchain rules

---

## Common Patterns Used

### Error Handling
All skills follow fail-closed pattern:
- Network errors → log, return false/error
- Invalid input → return error with context
- Missing data → return error with helpful message

### Logging
- Use existing Zap logger from `internal/log`
- Log authentication/authorization decisions
- Log errors with full context
- Never log raw keys or sensitive data

### Testing
- Table-driven tests for multiple scenarios
- Mock providers for isolation
- Database transaction rollback between tests
- >85% coverage target for all packages

### Security
- Address validation and normalization
- Key hashing before storage (SHA256)
- Ownership verification
- Fail-closed on errors
- No plaintext secrets in responses

---

## Integration Checklist

After implementing all three skills:

- [ ] Database migrations run successfully
- [ ] All repositories pass unit tests
- [ ] API key endpoints respond correctly
- [ ] API key middleware intercepts requests
- [ ] Blockchain rules evaluate with cached results
- [ ] Allowlist rules work with database
- [ ] OpenAPI spec is up to date
- [ ] Integration tests pass
- [ ] Test coverage >85% overall
- [ ] No raw keys in logs or responses
- [ ] RPC errors handled gracefully
- [ ] Ownership checks prevent cross-user access
- [ ] Performance targets met (<500ms auth flow)

---

## Getting Help

Each skill includes:
- ✅ Detailed struct definitions
- ✅ Implementation patterns
- ✅ Security considerations
- ✅ Testing requirements
- ✅ Error handling strategies
- ✅ Code examples
- ✅ Integration points

**Use each skill for:**
- Implementation guidance
- Code generation assistance
- Testing strategy
- Security review
- Documentation

---

## Phase 2 Completion Definition

Phase 2 is complete when:

✅ **Database Layer**
- User repository fully implemented and tested
- API Key repository with hashing and expiration
- Allowlist repository with address management
- All migrations running successfully

✅ **API Key Management**
- Generate, list, revoke endpoints working
- Middleware validates keys correctly
- Ownership checks prevent cross-user access
- Usage tracking (last_used_at) functional

✅ **Policy Engine**
- ERC20 balance rules evaluating correctly
- ERC721 ownership rules evaluating correctly
- All rule types (4 total) fully implemented
- Caching reducing RPC calls by 80%+

✅ **Testing**
- >85% code coverage overall
- All unit tests passing
- Integration tests for key workflows
- End-to-end testing completed

✅ **Documentation**
- OpenAPI spec fully updated
- API examples for all endpoints
- Configuration guide
- Policy rule documentation

---

**Generated:** October 26, 2025
**Target Completion:** November 2-6, 2025 (Week 2)
**Total Estimated Effort:** 14-16 hours
