# Gatekeeper API Documentation

Gatekeeper is a wallet-native authentication gateway that combines Sign-In with Ethereum (SIWE) with blockchain-based access control policies.

## Quick Start

### 1. Get a Nonce for Signing In

First, request a unique nonce to include in your SIWE message:

```bash
curl http://localhost:8080/auth/siwe/nonce
```

Response:
```json
{
  "nonce": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
  "expiresIn": 600
}
```

Each nonce:
- Is cryptographically secure (128-bit entropy)
- Expires after 10 minutes
- Can only be used once
- Must be included in your SIWE message

### 2. Sign the SIWE Message

Create a SIWE message in EIP-4361 format and sign it with your wallet:

```
example.com wants you to sign in with your Ethereum account:
0x1234567890abcdef1234567890abcdef12345678

I accept the Terms of Service: https://example.com/tos

URI: https://example.com
Version: 1
Chain ID: 1
Nonce: 0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef
Issued At: 2024-01-15T12:00:00Z
Expiration Time: 2024-01-15T13:00:00Z
```

Sign this message with your Ethereum private key using EIP-191 personal_sign.

### 3. Verify Signature and Get JWT Token

Send the signed message to get a JWT token:

```bash
curl -X POST http://localhost:8080/auth/siwe/verify \
  -H "Content-Type: application/json" \
  -d '{
    "message": "example.com wants you to sign in...",
    "signature": "0x1234567890abcdef..."
  }'
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expiresIn": 3600,
  "address": "0x1234567890abcdef1234567890abcdef12345678"
}
```

The token is valid for 1 hour and contains:
- Your wallet address
- Your assigned scopes
- Expiration time

### 4. Use Token for Protected Requests

Include the token in the Authorization header for authenticated requests:

```bash
curl http://localhost:8080/api/data \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

Response (if policies allow):
```json
{
  "message": "Access granted",
  "data": {
    "userId": "0x1234567890abcdef1234567890abcdef12345678"
  }
}
```

## Access Control Policies

Gatekeeper uses flexible policy rules to control access to protected resources.

### Policy Types

#### HasScopeRule

Check if the user has a specific scope/permission:

```json
{
  "path": "/api/admin",
  "method": "GET",
  "logic": "AND",
  "rules": [
    {
      "type": "has_scope",
      "scope": "admin"
    }
  ]
}
```

User must have `admin` scope in their JWT token.

#### InAllowlistRule

Check if user's wallet address is in an allowlist:

```json
{
  "path": "/api/transfer",
  "method": "POST",
  "logic": "AND",
  "rules": [
    {
      "type": "in_allowlist",
      "addresses": [
        "0x1234567890abcdef1234567890abcdef12345678",
        "0xabcdef1234567890abcdef1234567890abcdef12"
      ]
    }
  ]
}
```

Only specified addresses can access this endpoint.

#### ERC20MinBalanceRule

Check if user holds minimum ERC20 token balance:

```json
{
  "path": "/api/claim-reward",
  "method": "POST",
  "logic": "AND",
  "rules": [
    {
      "type": "erc20_min_balance",
      "contractAddress": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
      "minimumBalance": "1000000000000000000",
      "chainId": 1
    }
  ]
}
```

Requires user to hold at least 1 token (18 decimals) of the specified ERC20 contract.

#### ERC721OwnerRule

Check if user owns a specific NFT:

```json
{
  "path": "/api/nft-exclusive",
  "method": "GET",
  "logic": "AND",
  "rules": [
    {
      "type": "erc721_owner",
      "contractAddress": "0xBC4CA0EdA7647A8aB7C2061c2E2ad1D8e83B4764",
      "tokenId": "1234",
      "chainId": 1
    }
  ]
}
```

Requires user to own a specific NFT token.

### Logic Operators

#### AND Logic

All rules must pass for access to be granted:

```json
{
  "logic": "AND",
  "rules": [
    { "type": "has_scope", "scope": "admin" },
    { "type": "in_allowlist", "addresses": ["0x..."] }
  ]
}
```

Access granted only if user has BOTH `admin` scope AND is in the allowlist.

#### OR Logic

At least one rule must pass:

```json
{
  "logic": "OR",
  "rules": [
    { "type": "has_scope", "scope": "premium" },
    { "type": "erc20_min_balance", "contractAddress": "0x...", "minimumBalance": "1000000000000000000" }
  ]
}
```

Access granted if user has `premium` scope OR holds minimum ERC20 balance.

### Configuration Example

Complete policy configuration file (`policies.json`):

```json
[
  {
    "path": "/api/public",
    "method": "GET",
    "logic": "AND",
    "rules": []
  },
  {
    "path": "/api/admin",
    "method": "GET",
    "logic": "AND",
    "rules": [
      {
        "type": "has_scope",
        "scope": "admin"
      }
    ]
  },
  {
    "path": "/api/premium-feature",
    "method": "POST",
    "logic": "OR",
    "rules": [
      {
        "type": "has_scope",
        "scope": "premium"
      },
      {
        "type": "erc20_min_balance",
        "contractAddress": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
        "minimumBalance": "10000000000000000000",
        "chainId": 1
      }
    ]
  }
]
```

## HTTP Status Codes

| Status Code | Meaning | Example |
|-------------|---------|---------|
| 200 OK | Request successful | Successfully retrieved protected data |
| 400 Bad Request | Invalid request format | Missing required fields in request |
| 401 Unauthorized | Authentication failed | Missing or invalid JWT token |
| 403 Forbidden | Access denied by policy | Policy evaluation failed |
| 500 Internal Server Error | Server error | Blockchain RPC call failed |

## Error Responses

All error responses follow this format:

```json
{
  "message": "Human-readable error message",
  "code": "ERROR_CODE",
  "details": {
    "additionalInfo": "..."
  }
}
```

Common error codes:
- `INVALID_NONCE`: Nonce not found or already used
- `EXPIRED_NONCE`: Nonce has expired
- `INVALID_SIGNATURE`: Signature verification failed
- `INVALID_TOKEN`: JWT token is invalid
- `POLICY_FAILED`: Access control policy denied access
- `RPC_ERROR`: Blockchain RPC call failed
- `INTERNAL_ERROR`: Internal server error

## Security Considerations

### Nonce Management
- Nonces expire after 10 minutes
- Each nonce can only be used once
- Nonces are single-use to prevent replay attacks
- Always validate that received nonce matches issued nonce

### Token Security
- JWT tokens are signed with HS256
- Tokens expire after 1 hour
- Always use HTTPS in production
- Never expose private keys
- Keep JWT_SECRET environment variable secure

### Policy Evaluation
- Policies are evaluated with fail-closed security
- If policy evaluation fails or errors occur, access is denied
- All policy decisions are logged for audit trail
- Blockchain errors (RPC failures) result in access denial

### Address Handling
- All Ethereum addresses are case-insensitive
- Addresses are normalized to lowercase for comparison
- Always validate addresses are valid Ethereum format (0x + 40 hex chars)

## Rate Limiting & Caching

### Nonce Generation
- No rate limiting on `/auth/siwe/nonce`
- Clients should implement reasonable request spacing
- Each nonce is independent

### Token Verification
- Token verification is fast (cryptographic check only)
- Consider implementing rate limiting on `/auth/siwe/verify` to prevent brute-force attacks

### Blockchain Queries
- Results are cached in-memory with TTL (configurable, default 5 minutes)
- Reduces RPC calls for repeated policy checks
- Cache is per-instance (not shared across servers)

## Environment Variables

Required configuration:

```bash
PORT=8080                              # HTTP server port
JWT_SECRET=your-secret-key             # SIWE message + JWT signing secret
ETHEREUM_RPC=https://eth.example.com   # Ethereum RPC endpoint
LOG_LEVEL=info                         # Log level: debug, info, warn, error
NONCE_TTL_MINUTES=10                   # Nonce expiration time
JWT_EXPIRY_HOURS=1                     # JWT token expiration time
```

## Examples

### TypeScript/Web3.js Example

```typescript
import { ethers } from 'ethers';

// 1. Get nonce
const nonceRes = await fetch('/auth/siwe/nonce');
const { nonce } = await nonceRes.json();

// 2. Create SIWE message
const siweMessage = `example.com wants you to sign in with your Ethereum account:
${walletAddress}

I accept the Terms of Service: https://example.com/tos

URI: https://example.com
Version: 1
Chain ID: 1
Nonce: ${nonce}
Issued At: ${new Date().toISOString()}`;

// 3. Sign message
const signer = provider.getSigner();
const signature = await signer.signMessage(siweMessage);

// 4. Verify and get token
const tokenRes = await fetch('/auth/siwe/verify', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ message: siweMessage, signature })
});
const { token } = await tokenRes.json();

// 5. Use token for authenticated requests
const dataRes = await fetch('/api/data', {
  headers: { 'Authorization': `Bearer ${token}` }
});
```

### cURL Example

```bash
# Get nonce
NONCE=$(curl http://localhost:8080/auth/siwe/nonce | jq -r '.nonce')

# Create SIWE message and sign (replace with actual signed message)
SIGNATURE="0x1234567890abcdef..."

# Verify and get token
TOKEN=$(curl -X POST http://localhost:8080/auth/siwe/verify \
  -H "Content-Type: application/json" \
  -d "{\"message\": \"...\", \"signature\": \"$SIGNATURE\"}" | jq -r '.token')

# Use token
curl http://localhost:8080/api/data \
  -H "Authorization: Bearer $TOKEN"
```

## Audit Logging

All policy decisions are logged with full context:

```json
{
  "level": "info",
  "timestamp": "2024-01-15T12:00:00Z",
  "message": "policy decision: access denied",
  "path": "/api/admin",
  "method": "GET",
  "address": "0x1234567890abcdef1234567890abcdef12345678",
  "policies": 1,
  "scopes": [],
  "decision": "DENIED",
  "reason": "policy_failed"
}
```

Use this audit log for:
- Security monitoring
- Access control debugging
- Compliance reporting
- Incident investigation

## Support

For issues, questions, or contributions, please refer to the main README and project documentation.
