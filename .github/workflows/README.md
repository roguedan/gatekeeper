# GitHub Actions Workflows

This directory contains all CI/CD workflows for the Gatekeeper project. These workflows automate testing, building, security scanning, and deployment processes.

## Workflows Overview

### 1. Test Workflow (`test.yml`)

**Purpose:** Automated testing and code quality checks
**Triggers:** Push to main/develop, Pull requests
**Duration:** ~3-5 minutes

**Jobs:**
- `test` - Runs Go tests with race detection and coverage analysis
  - PostgreSQL 15 test database
  - Minimum 85% coverage threshold
  - Uploads coverage to Codecov
  - Generates coverage reports
- `lint` - Code quality checks with golangci-lint

**Key Features:**
- Race condition detection
- Code coverage tracking
- Automated code formatting checks
- Test result summaries
- Coverage artifacts

**Status Badge:**
```markdown
![Test](https://github.com/roguedan/gatekeeper/workflows/Test/badge.svg)
```

---

### 2. Build Workflow (`build.yml`)

**Purpose:** Build binaries and Docker images
**Triggers:** Push to main, version tags (v*), manual dispatch
**Duration:** ~8-10 minutes

**Jobs:**
- `build-backend` - Cross-platform binary compilation
  - Linux (amd64, arm64)
  - macOS (amd64, arm64)
  - Windows (amd64)
  - Version embedding
  - Artifact upload

- `build-docker` - Multi-architecture Docker images
  - linux/amd64, linux/arm64
  - Pushed to GitHub Container Registry
  - Semantic versioning tags
  - Build caching for speed

- `scan-docker` - Container security scanning
  - Trivy vulnerability scanner
  - SARIF report to GitHub Security
  - Critical/High severity alerts

- `release` - GitHub release creation (on version tags)
  - Binary artifacts
  - Docker image tags
  - Changelog generation
  - Slack notifications

**Artifacts:**
- Cross-platform binaries (tar.gz)
- Docker images in GHCR
- Security scan reports

**Status Badge:**
```markdown
![Build](https://github.com/roguedan/gatekeeper/workflows/Build%20and%20Release/badge.svg)
```

---

### 3. Security Workflow (`security.yml`)

**Purpose:** Comprehensive security scanning
**Triggers:** Pull requests, Push, Daily schedule (2 AM UTC), Manual dispatch
**Duration:** ~2-3 minutes

**Jobs:**
- `gosec` - Go security analysis
  - Scans for security vulnerabilities
  - SARIF report to GitHub Security
  - JSON report artifact

- `govulncheck` - Dependency vulnerability scanning
  - Go vulnerability database
  - Checks all dependencies
  - Fails on any vulnerabilities

- `dependency-review` - PR dependency analysis
  - Checks new dependencies
  - License compliance
  - Vulnerability alerts

- `codeql` - Static analysis
  - Advanced code scanning
  - Security-extended queries
  - Pattern detection

- `secret-scan` - Credential leak detection
  - TruffleHog scanner
  - Verified secrets only
  - Git history scanning

- `license-check` - License compliance
  - go-licenses tool
  - Forbidden license detection
  - License report generation

- `security-summary` - Aggregated results

**Artifacts:**
- Security scan reports (JSON/SARIF)
- License compliance report
- Vulnerability reports

**Status Badge:**
```markdown
![Security](https://github.com/roguedan/gatekeeper/workflows/Security%20Scan/badge.svg)
```

---

### 4. Deploy Workflow (`deploy.yml`)

**Purpose:** Automated deployment to staging and production
**Triggers:** Release published, Manual dispatch
**Duration:** ~10-15 minutes

**Jobs:**
- `pre-deploy-checks` - Validation
  - Verifies Docker image exists
  - Checks release metadata

- `deploy-staging` - Staging environment deployment
  - SSH deployment
  - Docker Compose update
  - Smoke tests
  - API health checks
  - Deployment status tracking

- `deploy-production` - Production deployment (requires approval)
  - Database backup
  - Blue-green deployment
  - Comprehensive health checks
  - Automatic rollback on failure
  - Traffic switching

- `post-deploy` - Post-deployment tasks
  - Deployment report
  - Documentation updates

**Environments:**
- `staging` - Auto-deploy on releases
- `production` - Manual approval required

**Status Badge:**
```markdown
![Deploy](https://github.com/roguedan/gatekeeper/workflows/Deploy/badge.svg)
```

---

## Workflow Triggers Summary

| Workflow | Push (main) | Pull Request | Release | Schedule | Manual |
|----------|-------------|--------------|---------|----------|--------|
| Test | ✅ | ✅ | - | - | ✅ |
| Build | ✅ | - | ✅ (tags) | - | ✅ |
| Security | ✅ | ✅ | - | Daily 2 AM | ✅ |
| Deploy | - | - | ✅ | - | ✅ |

---

## Required Secrets

Configure these secrets in GitHub Settings → Secrets and variables → Actions:

### Docker Registry
- `GITHUB_TOKEN` - Auto-provided by GitHub (no setup needed)

### Code Coverage
- `CODECOV_TOKEN` - From codecov.io

### Deployment
- `STAGING_SSH_KEY` - SSH private key for staging server
- `STAGING_HOST` - Staging server hostname (format: user@host)
- `PRODUCTION_SSH_KEY` - SSH private key for production server
- `PRODUCTION_HOST` - Production server hostname
- `DEPLOYMENT_SSH_KEY` - Generic deployment key (if needed)

### Notifications
- `SLACK_WEBHOOK` - Slack webhook URL for notifications

### Optional
- `TEST_JWT` - JWT token for production smoke tests

---

## Environment Variables

Workflows use these environment variables:

```yaml
REGISTRY: ghcr.io
IMAGE_NAME: ${{ github.repository }}
DATABASE_URL: postgres://postgres:postgres@localhost:5432/gatekeeper_test?sslmode=disable
JWT_SECRET: test-secret-key-min-32-chars-long
ETHEREUM_RPC: https://eth-mainnet.g.alchemy.com/v2/demo
```

---

## Workflow Permissions

Each workflow declares minimum required permissions:

- **Test:** `contents: read`, `pull-requests: write`, `checks: write`
- **Build:** `contents: write`, `packages: write`, `id-token: write`
- **Security:** `contents: read`, `security-events: write`, `pull-requests: write`
- **Deploy:** `contents: read`, `deployments: write`, `id-token: write`

---

## Caching Strategy

Workflows use GitHub Actions caching for speed:

- **Go modules:** `go.sum` cache key
- **Docker layers:** GitHub Actions cache (build-push-action)
- **Go build cache:** Automatic via setup-go action

Average cache hit saves 30-60 seconds per workflow run.

---

## Artifacts Retention

| Artifact | Retention | Size (approx) |
|----------|-----------|---------------|
| Coverage reports | 30 days | < 1 MB |
| Binary builds | 30 days | 5-10 MB each |
| Security reports | 30 days | < 1 MB |
| License reports | 30 days | < 100 KB |

---

## Matrix Builds

### Backend Binary Matrix
```yaml
os: [linux, darwin, windows]
arch: [amd64, arm64]
# Excludes: windows/arm64
```

Produces 5 binary artifacts per build.

### Docker Image Platforms
```yaml
platforms: linux/amd64,linux/arm64
```

Multi-arch images for broad compatibility.

---

## Debugging Workflows

### View Workflow Runs
1. Go to **Actions** tab in GitHub
2. Select the workflow
3. Click on a specific run
4. Expand job steps for logs

### Download Artifacts
1. Navigate to workflow run
2. Scroll to **Artifacts** section
3. Click to download

### Re-run Failed Jobs
1. Open failed workflow run
2. Click **Re-run jobs** → **Re-run failed jobs**

### Manual Workflow Dispatch
1. Go to **Actions** tab
2. Select workflow
3. Click **Run workflow**
4. Fill in parameters (if any)
5. Click **Run workflow**

---

## Common Issues

### Test Failures
- **Database connection issues:** Check PostgreSQL service health
- **Race conditions:** Review code with `go run -race`
- **Coverage below threshold:** Add tests to reach 85%

### Build Failures
- **Go version mismatch:** Ensure Go 1.21+ in workflow
- **Missing dependencies:** Run `go mod tidy` locally
- **Docker build fails:** Check Dockerfile syntax

### Security Failures
- **Vulnerabilities found:** Update dependencies with `go get -u`
- **Secrets detected:** Remove and rotate compromised secrets
- **License issues:** Review dependency licenses

### Deployment Failures
- **SSH connection:** Verify SSH keys and host access
- **Health check fails:** Check service logs on server
- **Rollback needed:** Automatic rollback on production failure

---

## Workflow Optimization

### Speed Improvements
- ✅ Go module caching enabled
- ✅ Docker layer caching configured
- ✅ Parallel job execution
- ✅ Minimal artifact retention

### Cost Optimization
- ✅ Conditional job execution
- ✅ Efficient artifact storage
- ✅ Scheduled scans (not on every commit)
- ✅ Public runners (no private runner costs)

### Best Practices
- ✅ Fail fast on critical errors
- ✅ Comprehensive test coverage
- ✅ Security scanning on PRs
- ✅ Production approval gates
- ✅ Automatic rollback on failure

---

## Monitoring and Notifications

### Slack Notifications
Configure Slack webhook to receive:
- Build completion status
- Deployment updates
- Security alerts
- Release notifications

### GitHub Notifications
Built-in notifications for:
- Workflow failures
- Security alerts
- Deployment status
- PR checks

### Status Badges
Add to README.md:
```markdown
![Test](https://github.com/roguedan/gatekeeper/workflows/Test/badge.svg)
![Build](https://github.com/roguedan/gatekeeper/workflows/Build%20and%20Release/badge.svg)
![Security](https://github.com/roguedan/gatekeeper/workflows/Security%20Scan/badge.svg)
```

---

## Release Process

### Creating a Release

1. **Tag the version:**
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. **Build workflow triggers automatically:**
   - Builds cross-platform binaries
   - Creates Docker images
   - Scans for vulnerabilities
   - Creates GitHub release

3. **Deploy workflow triggers:**
   - Deploys to staging
   - Runs smoke tests
   - Waits for production approval
   - Deploys to production

### Version Naming
- `v1.0.0` - Production release
- `v1.0.0-rc.1` - Release candidate
- `v1.0.0-beta.1` - Beta release
- `v1.0.0-alpha.1` - Alpha release

---

## Local Testing

### Test with act (GitHub Actions locally)

Install act:
```bash
brew install act  # macOS
# or
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash
```

Run workflows locally:
```bash
# Test workflow
act pull_request -W .github/workflows/test.yml

# Build workflow
act push -W .github/workflows/build.yml

# Security workflow
act pull_request -W .github/workflows/security.yml
```

**Note:** Some features may not work locally (secrets, services, etc.)

---

## Continuous Improvement

### Metrics to Track
- Workflow execution time
- Test coverage trends
- Security vulnerabilities over time
- Deployment frequency
- Failure rates

### Regular Maintenance
- [ ] Update action versions quarterly
- [ ] Review and optimize slow workflows
- [ ] Update security scanning tools
- [ ] Rotate deployment credentials
- [ ] Archive old artifacts

---

## Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [GitHub Container Registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [Workflow Syntax](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions)
- [Security Hardening](https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions)

---

**Last Updated:** November 1, 2025
**Maintained by:** Gatekeeper Team
