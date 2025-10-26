# 🧭 Portfolio Strategy & Project Plan — Go + Web3 + RWA Niche

## 1. Context

We identified a **clear talent gap** in Web3 around **RWA (Real World Assets)**, compliance, and infrastructure:

- Institutional capital is entering the space.
- Demand for engineers who can **bridge TradFi compliance with on-chain logic** is surging.
- Supply of such talent is tiny.
- Go + infra + Web3 auth + RWA = high career leverage.

✅ **Your positioning**: senior fintech/platform engineer with leadership experience + new Web3/RWA niche skills.

---

## 2. Skills Checklist (Strategic Career Leverage)

| Area | Skill | Why It Matters |
|------|-------|---------------|
| Backend | Go, .NET | Core infra/backend skills for protocols |
| Auth | SIWE, JWT, scopes | Wallet-native auth |
| Infra | Docker, CI/CD, IaC | Mature engineering practices |
| Compliance | RWA rules, KYC/AML flows | Rare and valuable niche |
| Web3 | Solidity basics, ERC20/721, Dune | Credibility in crypto |
| Leadership | Architecture, team enablement | Moves you above IC-only hires |

**Outcome**: strong portfolio signal → fewer competitors → inbound job flow, not outbound applications.

---

## 3. Portfolio Project Concept — `gatekeeper`

A production-quality Go project that demonstrates:

- 🪙 **Sign-In With Ethereum (SIWE)** login
- 🔐 **JWT** + **scoped API keys**
- 🧭 **Token / NFT / allowlist** based policy enforcement
- 🧪 **Tests**, **CI**, **OpenAPI**
- 🖥️ A **demo frontend** to make the flow tangible

This project is designed to look **like something a protocol infra team would actually use.**

---

## 4. Repo Structure

```
gatekeeper/
  cmd/server/main.go
  internal/
    http/
      router.go
      handlers/
        auth.go
        keys.go
        demo.go
        docs.go
      middleware/
        auth_jwt.go
        api_key.go
        policy_gate.go
        ratelimit.go
    auth/
      siwe.go
      jwt.go
      nonces.go
      scopes.go
    policy/
      engine.go
      rules.go
      models.go
      config.json
    chain/
      ethclient.go
      cache.go
    store/
      db.go
      migrations/0001_init.sql
    config/config.go
    log/log.go
  api/
    openapi.yaml
    postman_collection.json
  web/
    index.html
    package.json
    vite.config.ts
    src/
      App.tsx
      main.tsx
      components/
        ConnectButton.tsx
        ProtectedRouteDemo.tsx
      hooks/useAuth.ts
      styles.css
  deployments/
    Dockerfile
    docker-compose.yml
  .github/workflows/ci.yaml
  Makefile
  README.md
  ARCHITECTURE.md
```

---

## 5. Backend Highlights

- `/auth/siwe/nonce` — issues nonce for SIWE  
- `/auth/siwe/verify` — verifies sig, mints JWT  
- `/keys` CRUD — scoped API key issuance  
- `/alpha/data` — example **protected route** (policy gated)
- `/openapi.yaml` + `/docs` — auto docs with Redoc

### Policy Config Example
```json
{
  "routes": [
    {
      "path": "/alpha/data",
      "methods": ["GET"],
      "logic": "OR",
      "rules": [
        { "type": "has_scope", "scope": "read:alpha" },
        { "type": "in_allowlist", "addresses": ["0xAbC..."] },
        { "type": "erc20_min_balance", "chainId": 1, "token": "0xToken", "min": "1000000000000000000" }
      ]
    }
  ]
}
```

---

## 6. Testing

✅ Built-in tests:

- Nonce → bad signature → verify returns 401  
- Policy engine AND/OR logic  
- JWT & API key middleware behavior  
- (Optional) e2e tests hitting local anvil chain

CI:
- `go vet`
- `go test -race -cover`
- `docker build`
- GitHub Actions on PR

---

## 7. OpenAPI

- `api/openapi.yaml` defines all routes, schemas, and security schemes.
- `/docs` endpoint serves Redoc UI.
- Ensures **developer friendliness** for protocol integrations.

---

## 8. Docker & CI

- Minimal Distroless container for Go binary
- Compose runs Postgres + app
- CI pipeline builds and tests on push

```bash
docker compose up -d
```

---

## 9. Demo Frontend

- Built with Vite + React
- wagmi + viem + siwe
- Login flow:
  1. Connect wallet
  2. Get nonce
  3. Sign SIWE message
  4. Verify with backend
  5. Call protected route

Minimal UI:  
- Connect button  
- Sign-In With Ethereum  
- “Call Protected Route”  
- Display access granted / forbidden

---

## 10. Frontend Tech & Structure

```
web/
  src/
    App.tsx
    hooks/useAuth.ts
    components/ConnectButton.tsx
    components/ProtectedRouteDemo.tsx
    styles.css
```

- Vite dev proxy forwards `/auth` and `/alpha` calls to backend
- Uses localStorage to persist JWT
- Button UX is minimal but effective

---

## 11. Suggested Repo Names

| Name | Style | Notes |
|------|-------|-------|
| `gatekeeper` | Professional | Recommended |
| `warden` | Brandable | Short, strong |
| `siwe-gateway` | Descriptive | Explicit |
| `tokenwall` | Playful | Good for demos |

**Pick:** `gatekeeper`

---

## 12. Implementation Flow (Fast with Claude Code)

| Week | Focus | Outcome |
|------|-------|---------|
| 1 | Project skeleton + SIWE endpoints + tests + OpenAPI | Running backend with docs |
| 2 | Policy engine + JWT/API key middleware + caching | Core logic complete |
| 3 | Demo frontend + polish + CI | Showcase-ready repo |

---

## 13. Why This Project Is High Signal for Hiring

| Signal | Why it matters |
|--------|---------------|
| ✅ Wallet-native auth | Shows Web3 fluency |
| ✅ Token/NFT gating | Real protocol relevance |
| ✅ Go backend | Solid engineering |
| ✅ Tests + CI + docs | Production hygiene |
| ✅ Demo frontend | Makes it tangible |
| ✅ Compliance/RWA positioning | Rare skill set |

---

## 14. Next Steps

- Flesh out SIWE signature verification and JWT minting
- Implement ERC-20 and NFT rules with caching
- Add proper DB persistence for API keys
- Polish frontend UX and deploy
- Publish on GitHub with a clear README
- Write 1–2 blog posts on architecture & RWA context

---

## 15. Strategic Angle

This project sits exactly at the intersection of:

- 🧰 **Backend engineering (Go)**  
- 🪙 **Web3 primitives (SIWE, ERC20, NFT)**  
- 🏦 **Regulatory and compliance trends (RWA)**  
- 🧠 **Senior-level signal (tests, docs, polish)**

👉 It’s designed to **stand out to protocols, infra teams, and RWA startups** hiring experienced engineers who understand both finance and crypto.

---

## 16. Useful AI Prompts for Claude Code

- “Implement SIWE signature verification for auth/siwe.go”
- “Generate JWT middleware with refresh support”
- “Implement ERC20 min balance rule using go-ethereum”
- “Write unit tests for middleware/auth_jwt.go”
- “Generate a Postman collection from openapi.yaml”
- “Add error handling and structured logging to router.go”

---

## 17. Demo Links (once running)

- Frontend: `http://localhost:5173`  
- Backend Docs: `http://localhost:8080/docs`  
- API Spec: `http://localhost:8080/openapi.yaml`

---

## 18. License & Attribution

This project scaffold is designed for **personal portfolio use** to demonstrate full-stack engineering in the Web3 + compliance domain.  
Feel free to fork, extend, and make it your own.
