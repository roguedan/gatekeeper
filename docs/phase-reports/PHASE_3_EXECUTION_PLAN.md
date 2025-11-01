# Gatekeeper Phase 3 Execution Plan

**Phase 3 Focus:** Demo Frontend & Production Polish
**Estimated Duration:** 3-4 days (using parallel subagents)
**Target Completion:** Complete MVP with working frontend + Docker + CI/CD

---

## Executive Summary

Phase 3 implements a production-ready React frontend with wallet integration, Docker containerization, and CI/CD automation. The frontend will integrate with the Phase 2 backend to demonstrate the complete Gatekeeper authentication flow.

**MVP Components:**
- ✅ Phase 2 Backend (Complete - 87/100 grade, all critical fixes)
- 🔄 **Phase 3 Frontend** (New - React + TypeScript + wagmi)
- 🔄 **Containerization** (New - Docker + Docker Compose)
- 🔄 **CI/CD Pipeline** (New - GitHub Actions)
- 🔄 **Documentation** (New - Integration guide + deployment)

---

## Phase 3 Deliverables

### 1. React Frontend Application
**Directory:** `web/`
**Technology Stack:**
- React 18+ with TypeScript
- Vite for bundling (fast HMR)
- Tailwind CSS for styling
- wagmi + viem for Ethereum interaction
- RainbowKit for wallet connection
- TanStack React Query for data fetching

**Components to Build:**
```
web/src/
├── components/
│   ├── Layout/
│   │   ├── Header.tsx          # Navigation + wallet status
│   │   ├── Sidebar.tsx         # Menu navigation
│   │   └── Footer.tsx          # Footer info
│   ├── Auth/
│   │   ├── WalletConnect.tsx   # RainbowKit wallet button
│   │   ├── SignInFlow.tsx      # SIWE sign-in process
│   │   └── AuthGuard.tsx       # Protected route wrapper
│   ├── Pages/
│   │   ├── Home.tsx            # Landing page
│   │   ├── Dashboard.tsx       # User dashboard
│   │   ├── TokenGating.tsx     # Demo token-gated endpoint
│   │   ├── APIKeys.tsx         # API key management
│   │   ├── Policies.tsx        # Policy configuration UI
│   │   └── Status.tsx          # System status page
│   └── Common/
│       ├── Button.tsx
│       ├── Card.tsx
│       ├── Alert.tsx
│       └── LoadingSpinner.tsx
├── hooks/
│   ├── useAuth.ts              # Auth state management
│   ├── useSIWE.ts              # SIWE flow hook
│   ├── useAPI.ts               # API communication
│   └── usePolicy.ts            # Policy evaluation
├── services/
│   ├── api.ts                  # HTTP client
│   ├── auth.ts                 # Auth service
│   ├── siwe.ts                 # SIWE message creation
│   └── storage.ts              # Local storage management
├── types/
│   ├── auth.ts                 # Auth types
│   ├── api.ts                  # API response types
│   └── ethereum.ts             # Ethereum types
├── config/
│   ├── chains.ts               # Supported chains
│   ├── contracts.ts            # Contract addresses
│   └── env.ts                  # Environment config
├── pages/
│   ├── index.tsx               # App root
│   └── Layout.tsx              # Main layout
└── styles/
    └── globals.css             # Global styles
```

**Key Features:**
- ✅ Wallet connection (MetaMask, WalletConnect, Coinbase)
- ✅ SIWE sign-in flow with nonce-based verification
- ✅ JWT token management (store, refresh, logout)
- ✅ Protected routes with AuthGuard
- ✅ API key creation/management UI
- ✅ Token-gated content demo
- ✅ Policy configuration UI
- ✅ Responsive design with Tailwind CSS
- ✅ Error handling and loading states
- ✅ Dark mode support

---

### 2. Docker & Docker Compose
**Files to Create:**
- `Dockerfile` - Multi-stage build for Go backend
- `web/Dockerfile` - React frontend build
- `docker-compose.yml` - Complete stack (backend + frontend + postgres + redis)
- `.dockerignore` - Optimize build context
- `compose.override.yml` - Development overrides

**Docker Compose Stack:**
```yaml
services:
  # Backend Go API
  backend:
    build: .
    ports: ["8080:8080"]
    environment: [DATABASE_URL, ETHEREUM_RPC, JWT_SECRET, ...]
    depends_on: [postgres]
    healthcheck: ["test", "curl -f http://localhost:8080/health"]

  # React Frontend
  frontend:
    build: ./web
    ports: ["3000:3000"]
    environment: [VITE_API_URL=http://localhost:8080]
    depends_on: [backend]

  # PostgreSQL Database
  postgres:
    image: postgres:15-alpine
    environment: [POSTGRES_DB=gatekeeper, POSTGRES_PASSWORD=dev]
    ports: ["5432:5432"]
    volumes: [postgres_data:/var/lib/postgresql/data]

  # Redis Cache (optional for future)
  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]
```

**Production Considerations:**
- Multi-stage builds to minimize image size
- Non-root user for security
- Health checks for orchestration
- Resource limits (CPU, memory)
- Secrets management via environment

---

### 3. GitHub Actions CI/CD Pipeline
**Files to Create:**
- `.github/workflows/test.yml` - Run tests on every PR
- `.github/workflows/build.yml` - Build and validate
- `.github/workflows/deploy.yml` - Deploy to staging/production
- `.github/workflows/security.yml` - Security scanning

**Pipeline Stages:**

#### Test Workflow (PR trigger)
```yaml
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
      - run: go test ./... -v -race -coverprofile=coverage.out
      - uses: codecov/codecov-action@v3
```

#### Build Workflow
```yaml
name: Build & Release
on: [push: main, tags: v*]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
      - run: go build -o gatekeeper ./cmd/server
      - uses: docker/build-push-action@v5
        with:
          push: true
          tags: [...docker registry...]
```

#### Security Workflow
```yaml
name: Security Scan
on: [pull_request]
jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: go install github.com/securego/gosec/v2/cmd/gosec@latest
      - run: gosec ./...
      - uses: dependabot/dependabot-action@v3
```

---

### 4. Documentation

#### PHASE_3_COMPLETION.md (New)
- Frontend architecture overview
- Component documentation
- Integration points with backend
- API client setup
- Testing guide
- Deployment instructions
- Performance metrics

#### FRONTEND_SETUP.md (New)
```markdown
# Frontend Setup Guide

## Prerequisites
- Node.js 18+ and npm/yarn
- MetaMask or wallet extension

## Installation
\`\`\`bash
cd web
npm install
npm run dev
\`\`\`

## Environment Variables
\`\`\`
VITE_API_URL=http://localhost:8080
VITE_WALLETCONNECT_ID=your_walletconnect_id
VITE_CHAIN_ID=1
\`\`\`

## Features
- Wallet Connection
- SIWE Sign-In
- Protected Routes
- API Key Management
- Token-Gating Demo
```

#### DOCKER_DEPLOYMENT.md (New)
- Docker image build process
- Docker Compose startup
- Environment configuration
- Scaling considerations
- Health checks
- Troubleshooting

#### INTEGRATION_GUIDE.md (New)
- Frontend-Backend integration points
- API authentication flow
- Error handling
- Real-world examples
- Debugging guide

---

## Implementation Strategy

### Parallel Subagent Approach (Recommended)

Use 3 concurrent subagents for maximum efficiency:

**Subagent 1: Frontend Implementation**
- Use `frontend-web3-integration` skill
- Create React component structure
- Implement SIWE sign-in flow
- Build protected routes
- Add API key management UI
- Setup Tailwind CSS styling
- Estimated: 6-8 hours

**Subagent 2: Docker & Infrastructure**
- Create Dockerfile with multi-stage build
- Create docker-compose.yml stack
- Setup .dockerignore
- Add health checks
- Document configuration
- Test local Docker deployment
- Estimated: 3-4 hours

**Subagent 3: CI/CD & Documentation**
- Create GitHub Actions workflows
- Setup security scanning
- Create deployment documentation
- Write integration guides
- Create troubleshooting docs
- Setup automated releases
- Estimated: 4-5 hours

**Sequential Dependency:** All three can run in parallel (minimal dependencies)

---

## Success Criteria

### Frontend (Subagent 1)
- ✅ Wallet connection working (MetaMask, WalletConnect)
- ✅ SIWE sign-in flow complete
- ✅ JWT token stored and refreshed
- ✅ Protected routes with AuthGuard
- ✅ API key CRUD interface
- ✅ Token-gating demo page
- ✅ Responsive design (mobile/desktop)
- ✅ Dark mode support
- ✅ >80% component test coverage
- ✅ No TypeScript errors
- ✅ Vite dev server runs without errors
- ✅ Production build <500KB (with gzip)

### Docker & Compose (Subagent 2)
- ✅ Backend Docker image builds successfully
- ✅ Frontend Docker image builds successfully
- ✅ docker-compose up -d starts entire stack
- ✅ All services reach healthy state
- ✅ Frontend accessible at http://localhost:3000
- ✅ Backend accessible at http://localhost:8080
- ✅ Database migrations run automatically
- ✅ Environment variables properly configured
- ✅ Images follow security best practices
- ✅ Images <200MB (backend), <100MB (frontend)

### CI/CD Pipeline (Subagent 3)
- ✅ PR triggers test workflow
- ✅ Tests pass and coverage reported
- ✅ Security scanning runs
- ✅ Build workflow creates Docker images
- ✅ Tags trigger release workflow
- ✅ All workflows complete in <10 minutes
- ✅ Deployment documentation complete
- ✅ Rollback procedures documented
- ✅ Health checks verified in pipelines
- ✅ Slack/email notifications configured

### Overall Integration
- ✅ Frontend authenticates with backend
- ✅ API keys created via frontend UI
- ✅ Protected endpoints accessible with JWT
- ✅ Token-gating demo functional
- ✅ Full end-to-end flow tested
- ✅ No TypeScript errors in entire project
- ✅ All tests passing (backend + frontend)
- ✅ Documentation complete and accurate

---

## Technical Decisions

### Frontend Framework
- **React 18** - Industry standard, large ecosystem
- **TypeScript** - Type safety, better DX
- **Vite** - Fast bundler with HMR
- **Tailwind CSS** - Utility-first styling, responsive
- **wagmi** - React hooks for Ethereum
- **RainbowKit** - Beautiful wallet connection UI
- **TanStack Query** - Data fetching and caching

### Backend API Integration
- **Axios** - HTTP client with interceptors for JWT
- **API authentication** - Bearer token in Authorization header
- **Error handling** - Unified error response format
- **CORS** - Configured for localhost:3000

### Deployment
- **Docker** - Container standardization
- **Docker Compose** - Local development environment
- **GitHub Actions** - CI/CD automation
- **PostgreSQL** - Existing database
- **Redis** - Optional caching layer

---

## Testing Strategy

### Frontend Testing
- **Unit Tests** - Jest + React Testing Library
- **Component Tests** - Test isolated components
- **Integration Tests** - Test user flows
- **E2E Tests** - Playwright for full flows
- **Coverage Target** - >80% coverage

### Backend Integration Testing
- **API Tests** - Test endpoints with real frontend
- **Auth Flow Tests** - Complete SIWE + JWT + Protected Route
- **Token-Gating Tests** - Verify policy evaluation
- **Error Scenarios** - Network failures, invalid tokens

### Infrastructure Testing
- **Docker Build Tests** - Verify images build successfully
- **Compose Tests** - Test full stack startup
- **Health Check Tests** - Verify all services healthy
- **Security Tests** - Image vulnerability scanning

---

## Timeline & Milestones

### Day 1-2: Frontend Implementation
- **Day 1 Morning:** Component structure, wallet connection setup
- **Day 1 Afternoon:** SIWE sign-in flow, JWT management
- **Day 2 Morning:** Protected routes, API integration
- **Day 2 Afternoon:** UI refinement, styling, responsive design

### Day 2-3: Docker & Infrastructure
- **Day 2 Evening:** Dockerfile creation, compose setup
- **Day 3 Morning:** Environment configuration, health checks
- **Day 3 Afternoon:** Local testing, image optimization

### Day 3-4: CI/CD & Documentation
- **Day 3 Evening:** GitHub Actions workflows
- **Day 4 Morning:** Security scanning, release automation
- **Day 4 Afternoon:** Documentation completion, integration guide

### Day 4: Integration & Final Testing
- **Day 4 Evening:** Full end-to-end testing
- **Day 5 Morning:** Documentation review, final polish
- **Day 5 Afternoon:** Release preparation

---

## Files Structure After Phase 3

```
gatekeeper/
├── web/                           # React Frontend (NEW)
│   ├── src/
│   │   ├── components/            # React components
│   │   ├── hooks/                 # Custom React hooks
│   │   ├── services/              # API services
│   │   ├── types/                 # TypeScript types
│   │   ├── config/                # Configuration
│   │   ├── pages/                 # Page components
│   │   └── styles/                # CSS
│   ├── public/                    # Static assets
│   ├── package.json               # NPM dependencies
│   ├── tsconfig.json              # TypeScript config
│   ├── vite.config.ts             # Vite config
│   ├── vitest.config.ts           # Test config
│   ├── tailwind.config.js         # Tailwind config
│   └── Dockerfile                 # Frontend Docker image
│
├── Dockerfile                     # Backend Docker image (EXISTS)
├── docker-compose.yml             # Full stack (NEW)
├── .dockerignore                  # Docker ignore (NEW)
│
├── .github/
│   └── workflows/                 # CI/CD Pipelines (NEW)
│       ├── test.yml
│       ├── build.yml
│       ├── deploy.yml
│       └── security.yml
│
├── PHASE_3_EXECUTION_PLAN.md      # This file
├── PHASE_3_COMPLETION.md          # Phase 3 results (after)
├── FRONTEND_SETUP.md              # Frontend guide (NEW)
├── DOCKER_DEPLOYMENT.md           # Docker guide (NEW)
├── INTEGRATION_GUIDE.md           # Integration guide (NEW)
├── CI_CD_GUIDE.md                 # CI/CD guide (NEW)
│
└── [existing Phase 1-2 files...]
```

---

## Estimated Effort

| Task | Duration | Effort |
|------|----------|--------|
| React Frontend Components | 6-8 hours | Medium |
| SIWE Flow Implementation | 4-5 hours | Medium |
| Protected Routes & Auth | 3-4 hours | Medium |
| API Integration & Testing | 4-5 hours | Medium |
| Styling & Responsive Design | 3-4 hours | Low |
| Docker Setup | 2-3 hours | Low |
| Docker Compose Stack | 2-3 hours | Low |
| GitHub Actions Workflows | 3-4 hours | Low |
| Security Setup | 1-2 hours | Low |
| Documentation | 4-5 hours | Low |
| Integration Testing | 3-4 hours | Medium |
| **Total** | **35-43 hours** | **Parallel: 2-3 days** |

With parallel subagents: **2-3 days to completion**

---

## Dependencies & Prerequisites

### External Services
- ✅ Ethereum RPC endpoint (Alchemy, Infura, or local)
- ✅ MetaMask or wallet browser extension
- ✅ WalletConnect ID (optional)

### Software Requirements
- Node.js 18+
- npm or yarn
- Go 1.21+
- Docker & Docker Compose
- Git

### Existing Resources
- ✅ Phase 2 backend (complete and tested)
- ✅ OpenAPI specification
- ✅ Database schema and migrations
- ✅ Frontend Web3 integration skill
- ✅ Code review skill for verification

---

## Known Risks & Mitigations

### Risk 1: Frontend-Backend Incompatibility
- **Mitigation:** Start with API contract testing
- **Prevention:** Use OpenAPI spec as source of truth
- **Fallback:** Mock API server for frontend development

### Risk 2: Docker Build Failures
- **Mitigation:** Test multi-stage builds locally first
- **Prevention:** Use consistent base images
- **Fallback:** Document troubleshooting steps

### Risk 3: CI/CD Complexity
- **Mitigation:** Start with simple test workflow
- **Prevention:** Test locally with act (GitHub Actions locally)
- **Fallback:** Manual deployment steps documented

### Risk 4: Browser Compatibility Issues
- **Mitigation:** Test on Chrome, Firefox, Safari
- **Prevention:** Use modern standards, minimal polyfills
- **Fallback:** Document browser requirements

---

## Next Steps (After Phase 3)

### Phase 4 - Advanced Features (Future)
- [ ] Analytics dashboard
- [ ] Admin panel
- [ ] Rate limiting UI
- [ ] Webhook integrations
- [ ] Advanced policy builder
- [ ] Multi-signature support

### Phase 5 - Enterprise Features (Future)
- [ ] SAML/OAuth integration
- [ ] Advanced audit logging
- [ ] Compliance reporting
- [ ] SLA monitoring
- [ ] API versioning
- [ ] Developer portal

---

## Sign-Off & Approval

### Phase 3 Objective
Deliver a complete, production-ready MVP with working frontend, containerization, and CI/CD automation.

### Success Definition
- ✅ Frontend functional and integrated
- ✅ Docker deployment working
- ✅ CI/CD pipeline operational
- ✅ Full end-to-end flow tested
- ✅ Complete documentation

### Estimated Completion
**3-4 days with parallel subagents**

### Approval Gate
All success criteria met + documentation complete + tested end-to-end

---

**Created:** November 1, 2025
**Phase:** 3 of 5 (MVP)
**Status:** Planning Complete → Ready for Implementation
**Mode:** Parallel Subagent Execution Recommended

