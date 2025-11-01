# Phase 2 Execution Plan: Policy Engine & Token-Gating

## Overview
Implement a comprehensive policy engine supporting access control rules with blockchain integration, caching, and RPC management.

## Task Breakdown

### Block 1: Policy Configuration & Loading (Tasks 5.1 - 5.3)

#### Task 5.1: Policy Types & Configuration ⭐ TDD
- **Primary Skills**: `go-backend-development`, `test-driven-development`
- **TDD**: YES - Test configuration parsing and validation
- **Requirements**: REQ-POLICY-001
- **Key Implementation**:
  ```go
  // Policy types
  type Policy struct {
    Path    string     // Route pattern
    Methods []string   // HTTP methods
    Logic   string     // "AND" or "OR"
    Rules   []Rule
  }

  type Rule interface {
    Evaluate(ctx context.Context, address string, claims *auth.Claims) (bool, error)
  }

  type RuleType string
  const (
    HasScope    RuleType = "has_scope"
    InAllowlist RuleType = "in_allowlist"
    ERC20Min    RuleType = "erc20_min_balance"
    ERC721Owner RuleType = "erc721_owner"
  )
  ```

#### Task 5.2: Policy Loader & Validator ⭐ TDD
- **Primary Skills**: `go-backend-development`, `test-driven-development`
- **TDD**: YES - Test policy validation
- **Requirements**: REQ-POLICY-001
- **Key Implementation**:
  - Load policies from JSON config
  - Validate policy structure
  - Validate rule types and parameters
  - Clear error messages for invalid policies

#### Task 5.3: Policy Manager ⭐ TDD
- **Primary Skills**: `go-backend-development`, `test-driven-development`
- **TDD**: YES
- **Requirements**: REQ-POLICY-002
- **Key Implementation**:
  - Load all policies on startup
  - Match routes to policies
  - Support path patterns
  - Reload policies without downtime

### Block 2: Core Policy Evaluation (Tasks 6.1 - 6.4)

#### Task 6.1: AND/OR Evaluation Logic ⭐⭐ FULL TDD
- **Primary Skills**: `go-backend-development`, `test-driven-development`
- **TDD**: FULL RED-GREEN-REFACTOR
- **Requirements**: REQ-POLICY-002
- **Key Implementation**:
  ```go
  // AND logic - all rules must pass
  func (p *Policy) EvaluateAND(ctx context.Context, address string, claims *auth.Claims) (bool, error)

  // OR logic - any rule must pass
  func (p *Policy) EvaluateOR(ctx context.Context, address string, claims *auth.Claims) (bool, error)

  // Short-circuit evaluation for performance
  ```

#### Task 6.2: Scope-Based Rules ⭐ TDD
- **Primary Skills**: `go-backend-development`, `test-driven-development`
- **TDD**: YES
- **Requirements**: REQ-POLICY-003
- **Key Implementation**:
  ```go
  type HasScopeRule struct {
    Scope string
  }

  func (r *HasScopeRule) Evaluate(ctx context.Context, address string, claims *auth.Claims) (bool, error)
  ```

#### Task 6.3: Allowlist Rules ⭐ TDD
- **Primary Skills**: `go-backend-development`, `test-driven-development`
- **TDD**: YES
- **Requirements**: REQ-POLICY-004
- **Key Implementation**:
  - Case-insensitive address comparison
  - Checksum validation for Ethereum addresses
  - Efficient lookup (map-based)

#### Task 6.4: Default Policy Behavior ⭐ TDD
- **Primary Skills**: `go-backend-development`, `test-driven-development`
- **TDD**: YES
- **Requirements**: REQ-POLICY-002
- **Key Implementation**:
  - No matching policy = allow by default
  - Configurable default behavior
  - Logging for missing policies

### Block 3: Blockchain Integration (Tasks 7.1 - 7.4)

#### Task 7.1: ERC20 Balance Checking ⭐⭐ FULL TDD
- **Primary Skills**: `go-backend-development`, `test-driven-development`, `blockchain-testing`
- **TDD**: FULL RED-GREEN-REFACTOR
- **Requirements**: REQ-POLICY-005, REQ-POLICY-011
- **Dependencies**:
  - go-ethereum/ethclient for RPC
  - go-ethereum/accounts/abi for contract interaction
- **Key Implementation**:
  ```go
  type ERC20MinBalanceRule struct {
    ContractAddress string
    MinimumBalance  *big.Int
    ChainID         uint64
  }

  func (r *ERC20MinBalanceRule) Evaluate(ctx context.Context, address string, claims *auth.Claims) (bool, error)
  ```

#### Task 7.2: NFT Ownership Verification ⭐⭐ FULL TDD
- **Primary Skills**: `go-backend-development`, `test-driven-development`, `blockchain-testing`
- **TDD**: FULL RED-GREEN-REFACTOR
- **Requirements**: REQ-POLICY-006
- **Key Implementation**:
  ```go
  type ERC721OwnerRule struct {
    ContractAddress string
    TokenID         *big.Int
    ChainID         uint64
  }

  func (r *ERC721OwnerRule) Evaluate(ctx context.Context, address string, claims *auth.Claims) (bool, error)
  ```

#### Task 7.3: RPC Client & Provider Management ⭐⭐ TDD
- **Primary Skills**: `go-backend-development`, `test-driven-development`
- **TDD**: YES
- **Requirements**: REQ-POLICY-008
- **Key Implementation**:
  - Primary and fallback RPC providers
  - Timeout handling (5 seconds)
  - Connection pooling
  - Automatic failover
  - Provider health checks

#### Task 7.4: Blockchain Query Error Handling ⭐ TDD
- **Primary Skills**: `go-backend-development`, `test-driven-development`
- **TDD**: YES
- **Requirements**: REQ-POLICY-005, REQ-POLICY-006
- **Key Implementation**:
  - Retry logic with fallback
  - Timeout recovery
  - Error logging
  - Fail-closed for security

### Block 4: Caching System (Tasks 8.1 - 8.3)

#### Task 8.1: Cache Key Generation ⭐ TDD
- **Primary Skills**: `go-backend-development`, `test-driven-development`
- **TDD**: YES
- **Requirements**: REQ-POLICY-007
- **Key Implementation**:
  ```go
  // Cache key format: {type}:{chainId}:{contract}:{identifier}:{address}
  func GenerateCacheKey(dataType, chainID, contract, identifier, address string) string
  ```

#### Task 8.2: TTL-Based Cache with Cleanup ⭐⭐ FULL TDD
- **Primary Skills**: `go-backend-development`, `test-driven-development`
- **TDD**: FULL RED-GREEN-REFACTOR
- **Requirements**: REQ-POLICY-007
- **Key Implementation**:
  ```go
  type Cache struct {
    data    map[string]*CacheEntry
    ttl     time.Duration
    mu      sync.RWMutex
    cleanup *time.Ticker
  }

  func (c *Cache) Get(key string) (interface{}, bool)
  func (c *Cache) Set(key string, value interface{})
  func (c *Cache) CleanupExpired()
  ```

#### Task 8.3: Cache Integration with Rules ⭐ TDD
- **Primary Skills**: `go-backend-development`, `test-driven-development`
- **TDD**: YES
- **Requirements**: REQ-POLICY-005, REQ-POLICY-006, REQ-POLICY-007
- **Key Implementation**:
  - Cache blockchain results
  - Cache bypass flag (future)
  - Separate cache for each chain/contract

### Block 5: Middleware & Logging (Tasks 9.1 - 9.3)

#### Task 9.1: Policy Evaluation Middleware ⭐⭐ FULL TDD
- **Primary Skills**: `go-backend-development`, `test-driven-development`
- **TDD**: FULL RED-GREEN-REFACTOR
- **Requirements**: REQ-POLICY-009
- **Key Implementation**:
  ```go
  func PolicyMiddleware(manager *PolicyManager, logger *log.Logger) Middleware
  ```

#### Task 9.2: Policy Decision Logging ⭐⭐ FULL TDD
- **Primary Skills**: `go-backend-development`, `test-driven-development`
- **TDD**: FULL RED-GREEN-REFACTOR
- **Requirements**: REQ-POLICY-010
- **Key Implementation**:
  - Log grants with details
  - Log denials with reasons
  - Log RPC errors
  - Structured logging for analysis

#### Task 9.3: Integration Tests ⭐⭐⭐ TDD
- **Primary Skills**: `test-driven-development`, `go-backend-development`
- **TDD**: YES - Full workflow tests
- **Key Implementation**:
  - End-to-end policy evaluation
  - Multiple rule combinations
  - Caching verification
  - Error scenarios

## Implementation Order

1. **Task 5.1**: Policy types and configuration structures
2. **Task 5.2**: Policy loader and validator
3. **Task 5.3**: Policy manager
4. **Task 6.1**: AND/OR evaluation logic
5. **Task 6.2**: Scope-based rules
6. **Task 6.3**: Allowlist rules
7. **Task 6.4**: Default behavior
8. **Task 7.3**: RPC client setup
9. **Task 7.1**: ERC20 balance checking
10. **Task 7.2**: NFT ownership
11. **Task 7.4**: Error handling
12. **Task 8.1**: Cache key generation
13. **Task 8.2**: Cache with cleanup
14. **Task 8.3**: Cache integration
15. **Task 9.1**: Policy middleware
16. **Task 9.2**: Decision logging
17. **Task 9.3**: Integration tests

## Coverage Goals

- **Unit Tests**: >85% for all policy components
- **Integration Tests**: End-to-end policy evaluation flows
- **Blockchain Tests**: Mock RPC calls and error scenarios
- **Run before commit**: `go test ./... -cover`

## Dependencies to Add

```bash
go get github.com/ethereum/go-ethereum
```

## Test Categories

- Configuration loading and validation
- Policy evaluation (AND/OR logic)
- Rule evaluation (scope, allowlist, ERC20, NFT)
- Caching behavior
- Error handling and recovery
- Logging verification
- Middleware integration

## Next Steps

Start with Block 1: Task 5.1 - Policy Types & Configuration
