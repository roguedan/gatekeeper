# Build Policy Blockchain Rules

Generate production-ready implementations for ERC20 and ERC721 token-gating rules in the Gatekeeper policy engine.

## Instructions

When the user requests policy rule implementations, help create complete blockchain rule evaluations:

### 1. ERC20 Minimum Balance Rule (`internal/policy/erc20_rule.go`)

**Struct Definition:**
```go
type ERC20MinBalanceRule struct {
    RuleType string  // "ERC20MinBalance"
    Params   struct {
        Token   string // Contract address (0x...)
        ChainID int    // Chain ID (1 for mainnet, etc.)
        Minimum string // Big integer as string (wei)
    }
}
```

**Evaluation Implementation:**

```go
func (r *ERC20MinBalanceRule) Evaluate(ctx context.Context, claims *Claims, provider *chain.Provider, cache *chain.Cache) (bool, error) {
    // 1. Validate inputs:
    //    - Claims.Address must be valid Ethereum address
    //    - Token address must be valid and checksummed
    //    - Minimum must be parseable as big.Int
    //    - Provider must not be nil

    // 2. Build cache key:
    //    key := chain.CacheKey("erc20_balance", chainID, token, address)

    // 3. Try cache:
    //    if cached, err := cache.Get(key); err == nil {
    //        return validateBalance(cached, minimum)
    //    }

    // 4. Build ERC20 contract call:
    //    - Call: balanceOf(address)
    //    - Selector: 0x70a08231 (standard ERC20)
    //    - Encoded: chain.EncodeERC20BalanceOf(token, address)

    // 5. Execute RPC call:
    //    result, err := provider.Call(ctx, &chain.CallRequest{
    //        To:   token,
    //        Data: encodedCall,
    //    })

    // 6. Parse response:
    //    balance, err := chain.DecodeUint256(result)

    // 7. Cache result:
    //    cache.Set(key, balance.String())

    // 8. Compare:
    //    return balance >= minimum, nil
}

func (r *ERC20MinBalanceRule) Validate() error {
    // Check that Params.Token is valid address
    // Check that Params.Minimum is valid big.Int
    // Check that Params.ChainID > 0
    // Return error if invalid
}
```

**Implementation Details:**

- **Address Validation:**
  - Check format: `0x[0-9a-f]{40}` (case-insensitive)
  - Checksum if necessary (use `go-ethereum/common`)
  - Normalize to lowercase for consistency

- **Balance Parsing:**
  - Response is 32-byte hex: "0x00000000...0000000a"
  - Use `chain.DecodeUint256()` to convert to big.Int
  - Handle overflow/underflow gracefully

- **Caching:**
  - Cache key format: `erc20_balance:{chainID}:{token}:{address}`
  - TTL: 5 minutes (configurable)
  - Cache hits reduce RPC calls significantly

- **Error Handling:**
  - ErrInvalidAddress if address is malformed
  - ErrInvalidContract if contract call fails
  - ErrNetworkError if provider is unreachable
  - Log all errors with context (rule, address, token)

- **RPC Call Details:**
  - Use eth_call (read-only, no gas cost)
  - Call blockNumber: "latest"
  - Timeout: 5 seconds (inherited from provider)
  - Retry on network errors (up to 2 times)

- **Security:**
  - Don't trust contract response format (validate parsing)
  - Handle zero balance gracefully (evaluate to false, not error)
  - Fail-closed on network errors (return false, log error)

### 2. ERC721 Owner Rule (`internal/policy/erc721_rule.go`)

**Struct Definition:**
```go
type ERC721OwnerRule struct {
    RuleType string  // "ERC721Owner"
    Params   struct {
        Token   string // Contract address (0x...)
        ChainID int    // Chain ID
        TokenID string // Big integer as string (token ID to check)
    }
}
```

**Evaluation Implementation:**

```go
func (r *ERC721OwnerRule) Evaluate(ctx context.Context, claims *Claims, provider *chain.Provider, cache *chain.Cache) (bool, error) {
    // 1. Validate inputs (similar to ERC20)

    // 2. Build cache key:
    //    key := chain.CacheKey("erc721_owner", chainID, token, tokenID)
    //    Note: Cache by token ID, not address (not user-specific)

    // 3. Try cache:
    //    if owner, err := cache.Get(key); err == nil {
    //        return owner == claims.Address, nil
    //    }

    // 4. Build ERC721 contract call:
    //    - Call: ownerOf(tokenID)
    //    - Selector: 0x6352211e (standard ERC721)
    //    - Encoded: chain.EncodeERC721OwnerOf(token, tokenID)

    // 5. Execute RPC call:
    //    result, err := provider.Call(ctx, &chain.CallRequest{
    //        To:   token,
    //        Data: encodedCall,
    //    })

    // 6. Parse response:
    //    owner, err := chain.DecodeAddress(result)

    // 7. Cache result:
    //    cache.Set(key, owner)

    // 8. Compare:
    //    return strings.EqualFold(owner, claims.Address), nil
}

func (r *ERC721OwnerRule) Validate() error {
    // Check that Params.Token is valid address
    // Check that Params.TokenID is valid big.Int
    // Check that Params.ChainID > 0
    // Return error if invalid
}
```

**Implementation Details:**

- **Token ID Encoding:**
  - Can be 0-indexed or 1-indexed (implementation-specific)
  - Accept as string, convert to big.Int, encode as 32-byte hex
  - Use `chain.EncodeUint256()` for token ID

- **Owner Response Parsing:**
  - Response is 32-byte hex with address in last 20 bytes
  - Use `chain.DecodeAddress()` to extract
  - Compare case-insensitive (use `strings.EqualFold`)

- **Caching Strategy:**
  - Cache by token ID (not user-specific)
  - Owner doesn't change frequently
  - TTL: 5 minutes (balances caching strategy)
  - Can be shared across users checking same token

- **Error Handling:**
  - ErrTokenNotFound if ownerOf reverts (not minted)
  - ErrInvalidAddress if response doesn't contain address
  - Log with rule, token, tokenID context

- **Special Cases:**
  - Burned tokens: ownerOf reverts or returns zero address
  - Unburnable tokens: owner never changes (good cache hit rate)
  - Handle both cases gracefully (return false for burned)

### 3. Blockchain Utilities Updates (`internal/policy/blockchain.go`)

**Existing Functions (Verify These Work):**

```go
// Encode Ethereum address to 32-byte hex (left-padded)
func encodeAddress(addr string) (string, error) {
    // Validate address format
    // Strip 0x prefix
    // Left-pad with zeros to 32 bytes (64 hex chars)
    // Return 0x<64 hex chars>
}

// Decode 32-byte hex response to address
func decodeAddress(data string) (string, error) {
    // Strip 0x prefix
    // Take last 40 hex chars (20 bytes = address)
    // Validate hex format
    // Return 0x<40 hex chars> (checksummed)
}

// Encode uint256 to 32-byte hex
func encodeUint256(value string) (string, error) {
    // Parse as big.Int
    // Check not negative
    // Convert to 32-byte hex
    // Return 0x<64 hex chars>
}

// Decode 32-byte hex response to uint256
func decodeUint256(data string) (*big.Int, error) {
    // Parse hex string
    // Return as big.Int
}

// ERC20 balanceOf selector and encoding
const ERC20BalanceOfSelector = "0x70a08231"

func encodeERC20BalanceOf(token, account string) (string, error) {
    // Selector (4 bytes) + address (32 bytes)
    // ERC20BalanceOfSelector + encodeAddress(account)
    // Return complete calldata
}

// ERC721 ownerOf selector and encoding
const ERC721OwnerOfSelector = "0x6352211e"

func encodeERC721OwnerOf(token, tokenID string) (string, error) {
    // Selector (4 bytes) + uint256 (32 bytes)
    // ERC721OwnerOfSelector + encodeUint256(tokenID)
    // Return complete calldata
}

// Parse JSON-RPC response
func parseJSONRPCResponse(body []byte) (string, error) {
    // Unmarshal JSON
    // Check for error field (RPC error response)
    // Extract "result" field (hex string starting with 0x)
    // Validate it's valid hex
    // Return result
}
```

**New Helper Functions Needed:**

```go
// Cache key generation consistent format
func CacheKey(dataType, chainID, contract, identifier string) string {
    return fmt.Sprintf("%s:%d:%s:%s", dataType, chainID, contract, identifier)
}

// Validate Ethereum address format
func IsValidAddress(addr string) bool {
    // Check format: 0x + 40 hex chars
    // Return true/false
}

// Checksum address (optional, for display)
func ChecksumAddress(addr string) (string, error) {
    // Use go-ethereum/common to generate checksum
    // Return checksummed address
}
```

### 4. Integration with Policy Manager

**In `internal/policy/manager.go`:**

```go
func (m *PolicyManager) evaluateRule(ctx context.Context, rule Rule, claims *Claims) (bool, error) {
    switch r := rule.(type) {
    case *HasScopeRule:
        return r.Evaluate(ctx, claims, nil, nil)

    case *InAllowlistRule:
        return r.Evaluate(ctx, claims, nil, nil)

    case *ERC20MinBalanceRule:
        return r.Evaluate(ctx, claims, m.provider, m.cache)

    case *ERC721OwnerRule:
        return r.Evaluate(ctx, claims, m.provider, m.cache)

    default:
        return false, fmt.Errorf("unknown rule type: %T", rule)
    }
}
```

**PolicyManager Constructor Update:**

```go
type PolicyManager struct {
    policies []*Policy
    provider *chain.Provider   // NEW: for blockchain rules
    cache    *chain.Cache      // NEW: for caching results
}

func NewPolicyManager(provider *chain.Provider, cache *chain.Cache) *PolicyManager {
    return &PolicyManager{
        policies: []*Policy{},
        provider: provider,
        cache:    cache,
    }
}
```

### 5. Configuration for Blockchain Rules

**Add to `internal/config/config.go`:**

```go
type Config struct {
    // ... existing fields ...

    // Blockchain configuration
    EthereumRPC     string  // Primary RPC endpoint
    EthereumRPCFallback string // Fallback RPC endpoint (optional)
    ChainID         int     // Chain ID (1=mainnet, 5=goerli, 11155111=sepolia)
    CacheTTL        int     // Cache time-to-live in seconds (default: 300)
    RpcTimeout      int     // RPC call timeout in seconds (default: 5)

    // Policy configuration
    PolicyConfigPath string // Path to policy JSON file
}

func (c *Config) Load() error {
    // ... existing code ...

    // Load blockchain settings
    c.EthereumRPC = getEnv("ETHEREUM_RPC", "")
    c.EthereumRPCFallback = getEnv("ETHEREUM_RPC_FALLBACK", "")
    c.ChainID = getEnvInt("CHAIN_ID", 1)
    c.CacheTTL = getEnvInt("CACHE_TTL", 300)
    c.RpcTimeout = getEnvInt("RPC_TIMEOUT", 5)

    return nil
}
```

### 6. Testing Requirements

**Unit Tests (`internal/policy/erc20_rule_test.go`):**
- Evaluate rule with sufficient balance → true
- Evaluate rule with insufficient balance → false
- Evaluate rule with zero balance → false
- Cache hit returns cached result
- Invalid address format → error
- Invalid token format → error
- RPC provider error → error logged, returns false (fail-closed)
- Decimal handling (wei vs tokens)

**Unit Tests (`internal/policy/erc721_rule_test.go`):**
- Evaluate rule with ownership → true
- Evaluate rule without ownership → false
- Evaluate rule with burned token → false (or error)
- Cache hit returns cached owner
- Invalid token ID → error
- RPC provider error → error logged, returns false

**Integration Tests (`internal/policy/blockchain_integration_test.go`):**
- Mock RPC provider for testing
- Test ERC20 rule against mock USDC balances
- Test ERC721 rule against mock NFT ownership
- Test cache behavior across multiple calls
- Test with real Ethereum testnet (optional, requires API key)

**Test Helpers:**

```go
// Mock provider for testing
type MockProvider struct {
    responses map[string]string  // calldata -> response
    errors    map[string]error
}

func (m *MockProvider) Call(ctx context.Context, req *CallRequest) (string, error) {
    if err, ok := m.errors[req.Data]; ok {
        return "", err
    }
    return m.responses[req.Data], nil
}

// Helper for simulating USDC balance
func mockUSDCBalance(address string, balance *big.Int) string {
    return "0x" + encodeUint256(balance.String())
}
```

### 7. Policy File Examples

**Example `policies.json` with blockchain rules:**

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
          "params": {
            "scope": "vip"
          }
        },
        {
          "type": "ERC20MinBalance",
          "params": {
            "token": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",  // USDC mainnet
            "chainId": 1,
            "minimum": "1000000000"  // 1000 USDC (6 decimals)
          }
        }
      ]
    },
    {
      "path": "/api/nft-holders",
      "method": "GET",
      "logic": "OR",
      "rules": [
        {
          "type": "ERC721Owner",
          "params": {
            "token": "0xBC4CA0EdA7647A8aB7C2061c2E9cDAFCAc3c7f70",  // BAYC mainnet
            "chainId": 1,
            "tokenId": "1"
          }
        },
        {
          "type": "ERC721Owner",
          "params": {
            "token": "0x60E4d786d1ad0075D4f8295a7cC202Aa3d52e6E7",  // MAYC mainnet
            "chainId": 1,
            "tokenId": "1"
          }
        }
      ]
    }
  ]
}
```

### Implementation Order

1. **Phase 1:** Implement `ERC20MinBalanceRule.Evaluate()`
2. **Phase 2:** Implement `ERC721OwnerRule.Evaluate()`
3. **Phase 3:** Update `PolicyManager` with provider/cache
4. **Phase 4:** Verify blockchain utilities work correctly
5. **Phase 5:** Write comprehensive unit tests
6. **Phase 6:** Integration tests with mock provider

### Success Criteria

- ✅ ERC20 balance checks work correctly
- ✅ ERC721 ownership checks work correctly
- ✅ Caching reduces RPC calls by 80%+
- ✅ Error handling is fail-closed (return false on network errors)
- ✅ Address validation prevents injection attacks
- ✅ RPC errors are logged with context
- ✅ >85% test coverage
- ✅ Policy evaluation <500ms for typical request
- ✅ Cache hits complete <5ms
- ✅ Support for multiple blockchains via chainID parameter

## When to Use This Skill

Use this skill when you need to:
- Implement ERC20 and ERC721 token-gating rules
- Complete blockchain rule evaluation logic
- Add caching for RPC calls
- Write tests for policy rules with blockchain integration
- Debug policy evaluation with real contracts
- Support multi-chain policies

---

**Generated for:** Gatekeeper MVP Phase 2 - Blockchain Rules
**Complexity:** High - requires careful RPC handling and error management
