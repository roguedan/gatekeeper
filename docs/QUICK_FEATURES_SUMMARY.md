# Gatekeeper Features & Use Cases - Quick Reference

## 🎯 At a Glance

**Gatekeeper** = Wallet authentication + Token-gating + Access control

```
User connects wallet
        ↓
Signs SIWE message
        ↓
Gets JWT token
        ↓
System checks policies (Token balance? NFT ownership? Whitelisted?)
        ↓
Access granted or denied
```

---

## 🔐 Core Features (5 Minutes)

### 1. **Wallet Sign-In (SIWE)**
- Connect MetaMask, WalletConnect, Ledger, etc.
- Sign a message (no gas, no transaction)
- Get JWT token
- No passwords needed

### 2. **Token Balance Checking**
- "Users need 100+ USDC to access"
- Automatic on every request
- Works with any ERC20 token
- Multi-chain support

### 3. **NFT Ownership Verification**
- "BAYC holders only"
- Any ERC721 contract
- Check specific NFT or any in collection
- Instant verification

### 4. **Address Whitelisting**
- "Only these 10 addresses"
- Explicit allowlist
- Can combine with other rules
- Fast lookups

### 5. **API Key Management**
- Generate secure API keys
- Use instead of JWT for apps
- Track usage
- Revoke anytime

---

## 📊 Real-World Use Cases (Who Uses This?)

### **Community Access**
```
"Join our BAYC Discord"
→ Check if user owns BAYC NFT
→ Grant Discord access
→ Automatic verification
```

### **Premium Features**
```
"$1000+ USDC holders get pro tier"
→ Check user's USDC balance
→ Grant pro API access
→ Different rate limits
```

### **Exclusive Content**
```
"Pudgy Penguin owners get alpha access"
→ Verify NFT ownership
→ Grant content access
→ Track what they access
```

### **Multi-Criteria Access**
```
"Need: ETH + USDC + whitelist"
→ Check all 3 conditions
→ All must pass (AND logic)
→ Grant access if all met
```

### **Alternative Access**
```
"BAYC owners OR MAYC owners OR allowlist"
→ Check any of the conditions
→ Any one must pass (OR logic)
→ Grant access if any met
```

---

## 🚀 Technology Stack

```
Frontend                Backend              Blockchain
─────────────          ─────────────        ──────────────
React 18               Go 1.21+             Ethereum RPC
TypeScript             PostgreSQL           ERC20/ERC721
wagmi + viem           REST API             Multi-chain
RainbowKit             JWT/API Keys         Smart Contracts
Tailwind CSS           Rate Limiting
Vite                   Audit Logging
```

---

## 💡 Key Capabilities

| Feature | What It Does | Benefits |
|---------|--------------|----------|
| **SIWE Auth** | Sign in with wallet | No passwords, self-custody |
| **Token Gating** | Verify holdings | Automated access control |
| **NFT Verification** | Check ownership | Community membership |
| **Whitelisting** | Explicit allowlist | Simple & fast |
| **API Keys** | App authentication | Programmatic access |
| **Policy Engine** | Complex rules (AND/OR) | Flexible business logic |
| **Rate Limiting** | Control API usage | Prevent abuse |
| **Audit Logging** | Track everything | Compliance & security |
| **Multi-Chain** | Any EVM network | Users can sign from anywhere |
| **Caching** | 5-min cache | 80%+ faster responses |

---

## 📈 Use Case Comparison

```
Token-Holder Access     NFT-Gated Access      Hybrid Access
───────────────────     ────────────────      ─────────────
Requirements:           Requirements:          Requirements:
✓ Hold 100+ USDC        ✓ Own any BAYC        ✓ Hold 100+ USDC
                                               ✓ On whitelist
                                               ✓ Own NFT (any)

Access:                 Access:                Access:
✓ Basic tier           ✓ Private Discord      ✓ VIP tier
                       ✓ Exclusive content    ✓ Premium features

Example:               Example:                Example:
Yield dashboard        Creator community      DAO governance
```

---

## 🎮 Common Scenarios

### Scenario 1: "Web3 Game"
```
Requirement: Hold governance token (GOV)
Users: Play-to-earn players
Auth: SIWE sign-in
Verify: GOV token balance
Access: Play game, earn rewards
API: Game servers authenticate via JWT
```

### Scenario 2: "NFT Community"
```
Requirement: Own BAYC NFT
Users: Ape holders globally
Auth: SIWE sign-in
Verify: NFT ownership check
Access: Private Discord server
Integration: Auto-add to Discord role
```

### Scenario 3: "DeFi Dashboard"
```
Requirements: Hold 1000+ USDC OR 50+ ETH
Users: Traders & investors
Auth: SIWE sign-in
Verify: Token balances on any chain
Access: Real-time portfolio dashboard
API: Third-party apps can use API key
```

### Scenario 4: "Enterprise Blockchain"
```
Requirements: Employee + (CFO OR CEO roles)
Users: Company employees
Auth: SIWE with employee addresses
Verify: Allowlist by role
Access: Financial dashboards
Audit: Complete access logs
```

---

## 🔧 Configuration Examples

### Simple: Token Holder
```json
{
  "path": "/api/premium",
  "rules": [{
    "type": "ERC20MinBalance",
    "token": "0xUSDAC_ADDRESS",
    "minimum": "1000000000"
  }]
}
```

### Medium: NFT Owner
```json
{
  "path": "/api/exclusive",
  "rules": [{
    "type": "ERC721Owner",
    "contract": "0xBAYC_ADDRESS"
  }]
}
```

### Complex: Multiple Rules (ALL must pass)
```json
{
  "path": "/api/vip",
  "logic": "AND",
  "rules": [
    { "type": "ERC20MinBalance", ... },
    { "type": "ERC721Owner", ... },
    { "type": "InAllowlist", ... }
  ]
}
```

### Flexible: Alternative Rules (ANY can pass)
```json
{
  "path": "/api/club",
  "logic": "OR",
  "rules": [
    { "type": "ERC721Owner", "contract": "0xBAYC" },
    { "type": "ERC721Owner", "contract": "0xMAYC" },
    { "type": "InAllowlist", ... }
  ]
}
```

---

## 📊 By Industry

### **DeFi**
✅ Yield farm access by TVL
✅ Governance voting by token
✅ Premium analytics dashboard
✅ Multi-chain support

### **NFT/Gaming**
✅ NFT-gated Discord
✅ Play-to-earn access
✅ Exclusive content
✅ Verifiable ownership

### **Social/Community**
✅ Token-holder communities
✅ NFT-gated groups
✅ Verified member roles
✅ Cross-platform integration

### **Enterprise**
✅ Employee authentication
✅ Role-based access
✅ Audit trails
✅ Compliance logging

---

## 🔌 Integration Points

```
Your App
   │
   ├─→ Frontend: Connect wallet (wagmi)
   │           Sign SIWE message
   │           Get JWT token
   │
   ├─→ Backend: Validate JWT
   │           Check policies
   │           Return data or 403
   │
   └─→ Blockchain: Verify token/NFT
                  Cache results (5 min)
                  Fallback provider
```

---

## ⚡ Performance

| Operation | Time | Notes |
|-----------|------|-------|
| Wallet connect | 1-2s | User action |
| SIWE message | <100ms | Network |
| Signature | 1-5s | User confirms |
| JWT verify | <1ms | Cache hit |
| Policy eval (cached) | <5ms | In-memory |
| Policy eval (RPC) | <500ms | Blockchain call |
| Cache hit rate | 80%+ | 5-min TTL |

---

## 🛡️ Security Features

✅ **No Passwords** - SIWE authentication
✅ **No Private Keys** - Wallet handles signing
✅ **Fail-Closed** - Deny on blockchain error
✅ **Rate Limited** - Prevent abuse
✅ **Audit Logged** - Track everything
✅ **Encrypted Keys** - API keys hashed (SHA256)
✅ **Multi-Chain** - Works on any EVM network

---

## 📱 Supported Wallets

- ✅ MetaMask
- ✅ WalletConnect (Ledger, Trezor, etc.)
- ✅ Coinbase Wallet
- ✅ Rainbow
- ✅ Trust Wallet
- ✅ Any EIP-1193 wallet

---

## 🌍 Supported Networks

- ✅ Ethereum
- ✅ Polygon
- ✅ Arbitrum
- ✅ Optimism
- ✅ Base
- ✅ Sepolia (testnet)
- ✅ Any EVM-compatible chain

---

## 📦 Quick Deployment

```bash
# 1. Clone repo
git clone https://github.com/roguedan/gatekeeper

# 2. Start with Docker Compose
docker-compose up -d

# 3. Services running at:
Frontend: http://localhost:3000
Backend:  http://localhost:8080

# 4. Connect wallet & start testing!
```

---

## 📚 Documentation

**Get Started:**
- [Local Testing](../guides/LOCAL_TESTING.md) - Local development
- [Integration Guide](../guides/INTEGRATION_GUIDE.md) - Backend integration
- [API Reference](../api/API.md) - Complete endpoints

**Deploy:**
- [Docker Setup](../deployment/DOCKER_DEPLOYMENT.md) - Docker guide
- [CI/CD Guide](../deployment/CI_CD_GUIDE.md) - GitHub Actions

**Details:**
- [Full Features Guide](./FEATURES_AND_USECASES.md) - Comprehensive reference
- [Blockchain Rules](../api/BLOCKCHAIN_RULES_README.md) - Token-gating details

---

## 💬 Summary

**Gatekeeper simplifies:**

1. **Authentication** → "Who are you?" (SIWE wallet sign-in)
2. **Verification** → "What do you own?" (Token/NFT checks)
3. **Authorization** → "Can you access this?" (Policy evaluation)
4. **Management** → "Track what happens" (Audit logging)

**In one production-ready system.**

---

**Next Steps:**
1. Read [Local Testing Guide](../guides/LOCAL_TESTING.md)
2. Try [Docker setup](../deployment/DOCKER_DEPLOYMENT.md)
3. Check [API examples](../api/API.md)
4. Deploy to production

