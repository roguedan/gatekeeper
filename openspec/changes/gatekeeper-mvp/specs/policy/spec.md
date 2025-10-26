# Policy Engine Specification

## ADDED Requirements

### REQ-POLICY-001: Policy Configuration
The system SHALL support configurable access control policies defined in JSON format.

**Scenario: Load policies from configuration**
- GIVEN a JSON configuration file with route policies
- WHEN the application starts
- THEN the system SHALL parse the policy configuration
- AND SHALL validate the policy structure
- AND SHALL reject invalid policy definitions with clear error messages
- AND SHALL load all valid policies into memory

**Scenario: Policy structure validation**
- GIVEN a policy configuration file
- WHEN the configuration is loaded
- THEN each policy SHALL have a "path" (route pattern)
- AND SHALL have "methods" array (HTTP methods)
- AND SHALL have "logic" field ("AND" or "OR")
- AND SHALL have "rules" array with at least one rule
- AND each rule SHALL have a valid "type"
- AND each rule SHALL have required parameters for its type

### REQ-POLICY-002: Policy Evaluation Engine
The system SHALL evaluate access policies using AND/OR logic composition.

**Scenario: Evaluate policy with AND logic**
- GIVEN a policy with logic: "AND" and multiple rules
- WHEN the policy is evaluated for a user
- THEN ALL rules MUST evaluate to true for access to be granted
- AND if ANY rule evaluates to false, access SHALL be denied
- AND the evaluation SHALL short-circuit on first false result

**Scenario: Evaluate policy with OR logic**
- GIVEN a policy with logic: "OR" and multiple rules
- WHEN the policy is evaluated for a user
- THEN ANY rule evaluating to true SHALL grant access
- AND if ALL rules evaluate to false, access SHALL be denied
- AND the evaluation SHALL short-circuit on first true result

**Scenario: No matching policy for route**
- GIVEN a request to a protected route
- WHEN no policy is defined for that route and method
- THEN the system SHALL allow access by default
- OR SHALL deny access by default based on configuration
- AND SHALL log the missing policy for administrator review

### REQ-POLICY-003: Scope-Based Authorization
The system SHALL support authorization based on JWT token scopes.

**Scenario: Check for required scope**
- GIVEN a rule of type "has_scope" with scope: "read:data"
- WHEN evaluated for a user with scopes: ["read:data", "write:data"]
- THEN the rule SHALL evaluate to true

**Scenario: Missing required scope**
- GIVEN a rule of type "has_scope" with scope: "admin:manage"
- WHEN evaluated for a user with scopes: ["read:data"]
- THEN the rule SHALL evaluate to false

**Scenario: Empty scopes**
- GIVEN a rule of type "has_scope" with scope: "any:scope"
- WHEN evaluated for a user with no scopes
- THEN the rule SHALL evaluate to false

### REQ-POLICY-004: Allowlist-Based Authorization
The system SHALL support address-based allowlists for access control.

**Scenario: Address in allowlist**
- GIVEN a rule of type "in_allowlist" with addresses: ["0xABC...", "0xDEF..."]
- WHEN evaluated for user address "0xABC..." (case-insensitive)
- THEN the rule SHALL evaluate to true

**Scenario: Address not in allowlist**
- GIVEN a rule of type "in_allowlist" with addresses: ["0xABC..."]
- WHEN evaluated for user address "0x123..."
- THEN the rule SHALL evaluate to false

**Scenario: Address comparison is case-insensitive**
- GIVEN a rule with allowlist containing "0xabc..." (lowercase)
- WHEN evaluated for user address "0xABC..." (uppercase)
- THEN the addresses SHALL be normalized (checksummed)
- AND the rule SHALL evaluate to true

### REQ-POLICY-005: ERC20 Minimum Balance Check
The system SHALL verify users hold minimum ERC20 token balances.

**Scenario: User meets minimum ERC20 balance**
- GIVEN a rule requiring minimum 1000 tokens (18 decimals)
- WHEN evaluated for a user with balance of 5000 tokens
- THEN the system SHALL call the token contract's balanceOf function
- AND SHALL compare the balance with the minimum
- AND the rule SHALL evaluate to true

**Scenario: User below minimum ERC20 balance**
- GIVEN a rule requiring minimum 1000 tokens
- WHEN evaluated for a user with balance of 500 tokens
- THEN the rule SHALL evaluate to false

**Scenario: ERC20 balance cached**
- GIVEN a balance check was performed 2 minutes ago
- WHEN the same balance is checked again
- THEN the system SHALL use the cached result
- AND SHALL NOT make a new RPC call
- AND the cache SHALL have 5-minute TTL

**Scenario: ERC20 RPC call fails**
- GIVEN an RPC call to check balance fails
- WHEN the network error occurs
- THEN the system SHALL retry once with fallback provider
- AND if both fail, SHALL evaluate to false
- AND SHALL log the error for monitoring

**Scenario: Multi-chain ERC20 support**
- GIVEN a rule specifies chainId: 1 (Ethereum mainnet)
- WHEN the balance is checked
- THEN the system SHALL use the RPC provider for chain ID 1
- AND SHALL include chain ID in cache key
- AND different chains SHALL have separate cached values

### REQ-POLICY-006: NFT Ownership Verification
The system SHALL verify users own specific NFTs.

**Scenario: User owns required NFT**
- GIVEN a rule of type "erc721_owner" for token ID 42
- WHEN evaluated for a user
- THEN the system SHALL call ownerOf(42) on the NFT contract
- AND SHALL compare the owner with the user's address
- AND the rule SHALL evaluate to true if addresses match

**Scenario: User does not own NFT**
- GIVEN a rule requiring NFT token ID 42
- WHEN evaluated for a user who doesn't own it
- THEN the rule SHALL evaluate to false

**Scenario: NFT does not exist**
- GIVEN a rule for a non-existent token ID
- WHEN the ownerOf call reverts
- THEN the rule SHALL evaluate to false
- AND SHALL log the error for debugging

**Scenario: NFT ownership cached**
- GIVEN an ownership check was performed recently
- WHEN the same check is requested
- THEN the system SHALL use cached result with 5-minute TTL
- AND SHALL include chain ID and token ID in cache key

### REQ-POLICY-007: Cache Management
The system SHALL implement caching for blockchain data to minimize RPC calls.

**Scenario: Cache key generation**
- GIVEN a blockchain data request
- WHEN generating cache key
- THEN the key SHALL include: data type, chain ID, contract address, identifier, user address
- AND SHALL be deterministic for same inputs
- AND SHALL differentiate between different requests

**Scenario: Cache expiration**
- GIVEN a cached value with 5-minute TTL
- WHEN 5 minutes have elapsed since caching
- THEN the cached value SHALL be considered stale
- AND the next request SHALL trigger a fresh RPC call
- AND the cache SHALL be updated with new value

**Scenario: Cache cleanup**
- GIVEN expired cache entries
- WHEN cleanup runs (every 1 minute)
- THEN expired entries SHALL be removed
- AND memory SHALL be reclaimed
- AND cache size SHALL not grow unbounded

**Scenario: Cache bypass for critical operations**
- GIVEN a request with cache-bypass flag (if implemented)
- WHEN the request is processed
- THEN the system SHALL ignore cached values
- AND SHALL make a fresh RPC call
- AND SHALL update the cache with the new value

### REQ-POLICY-008: RPC Provider Management
The system SHALL manage blockchain RPC connections reliably.

**Scenario: Primary RPC provider unavailable**
- GIVEN a primary RPC provider fails
- WHEN a blockchain query is attempted
- THEN the system SHALL automatically retry with fallback provider
- AND SHALL log the failover event
- AND SHALL continue serving requests

**Scenario: RPC call timeout**
- GIVEN an RPC call exceeds timeout threshold (5 seconds)
- WHEN the timeout occurs
- THEN the system SHALL cancel the request
- AND SHALL try fallback provider
- AND SHALL log the timeout

**Scenario: Connection pooling**
- GIVEN multiple concurrent blockchain queries
- WHEN RPC calls are made
- THEN the system SHALL reuse HTTP connections
- AND SHALL limit concurrent connections to prevent overload
- AND SHALL queue requests if limit reached

### REQ-POLICY-009: Policy Middleware Integration
The system SHALL integrate policy enforcement as HTTP middleware.

**Scenario: Protected route with policy**
- GIVEN a request to /alpha/data
- WHEN a policy is configured for that route
- THEN the policy middleware SHALL intercept the request
- AND SHALL extract user identity from JWT claims
- AND SHALL evaluate the matching policy
- AND SHALL allow request to proceed if policy passes
- AND SHALL return HTTP 403 Forbidden if policy fails

**Scenario: Policy evaluation error**
- GIVEN policy evaluation encounters an error (e.g., RPC failure)
- WHEN the error occurs
- THEN the system SHALL default to denying access (fail-closed)
- AND SHALL return HTTP 500 Internal Server Error
- AND SHALL log the error with full context

**Scenario: Multiple policies for same route**
- GIVEN multiple policies could match a route
- WHEN determining which policy to apply
- THEN the system SHALL use the most specific path match
- OR SHALL apply policies in defined priority order
- AND SHALL document the policy selection algorithm

### REQ-POLICY-010: Policy Evaluation Logging
The system SHALL log policy evaluation decisions for auditing.

**Scenario: Log policy grant**
- GIVEN access is granted by policy
- WHEN the decision is made
- THEN the system SHALL log
  - Timestamp
  - User address
  - Route and method
  - Policy that granted access
  - Rules evaluated
  - Evaluation duration

**Scenario: Log policy denial**
- GIVEN access is denied by policy
- WHEN the decision is made
- THEN the system SHALL log
  - Timestamp
  - User address
  - Route and method
  - Policy that denied access
  - Which rules failed
  - Reason for failure (balance too low, not in allowlist, etc.)

**Scenario: Log RPC errors during evaluation**
- GIVEN an RPC call fails during policy evaluation
- WHEN the error occurs
- THEN the system SHALL log
  - Timestamp
  - User address
  - Policy being evaluated
  - RPC endpoint that failed
  - Error message
  - Whether fallback succeeded

### REQ-POLICY-011: Decimal Handling for Token Amounts
The system SHALL correctly handle token decimals in balance comparisons.

**Scenario: ERC20 token with 18 decimals**
- GIVEN a rule requiring minimum "1000000000000000000" (1 token with 18 decimals)
- WHEN checking a balance
- THEN the system SHALL compare raw values without decimal conversion
- AND the minimum SHALL be specified in base units (wei-equivalent)

**Scenario: Token with non-standard decimals**
- GIVEN an ERC20 token with 6 decimals
- WHEN specifying minimum balance in policy
- THEN the minimum SHALL be specified in base units (e.g., 1000000 for 1 USDC)
- AND the system SHALL NOT perform decimal conversion
- AND policy authors SHALL be responsible for correct base unit specification

## MODIFIED Requirements

None - this is a new service specification.

## REMOVED Requirements

None - this is a new service specification.
