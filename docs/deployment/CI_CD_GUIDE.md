# CI/CD Guide - Gatekeeper

Complete guide to Continuous Integration and Continuous Deployment workflows for the Gatekeeper project.

## Table of Contents

- [Overview](#overview)
- [Workflow Architecture](#workflow-architecture)
- [Setting Up CI/CD](#setting-up-cicd)
- [Workflow Details](#workflow-details)
- [Triggering Workflows](#triggering-workflows)
- [Viewing Results](#viewing-results)
- [Debugging Failed Workflows](#debugging-failed-workflows)
- [Managing Secrets](#managing-secrets)
- [Deployment Process](#deployment-process)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

---

## Overview

Gatekeeper uses GitHub Actions for automated CI/CD workflows:

- **Test Workflow** - Automated testing with >85% coverage
- **Build Workflow** - Cross-platform binaries and Docker images
- **Security Workflow** - Vulnerability scanning and security analysis
- **Deploy Workflow** - Staging and production deployments

All workflows are defined in `.github/workflows/` and run automatically on relevant events.

---

## Workflow Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    GitHub Repository                         │
└─────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        ▼                   ▼                   ▼
   ┌─────────┐        ┌─────────┐        ┌─────────┐
   │  Push   │        │   PR    │        │ Release │
   └─────────┘        └─────────┘        └─────────┘
        │                   │                   │
        ▼                   ▼                   ▼
   ┌─────────┐        ┌─────────┐        ┌─────────┐
   │  Test   │◄───────┤Security │        │  Build  │
   └─────────┘        └─────────┘        └─────────┘
        │                                       │
        ▼                                       ▼
   ┌─────────┐                            ┌─────────┐
   │  Build  │                            │ Deploy  │
   └─────────┘                            └─────────┘
                                               │
                                    ┌──────────┴──────────┐
                                    ▼                     ▼
                               ┌─────────┐          ┌─────────┐
                               │ Staging │          │  Prod   │
                               └─────────┘          └─────────┘
```

---

## Setting Up CI/CD

### 1. Configure GitHub Secrets

Navigate to **Settings** → **Secrets and variables** → **Actions** → **New repository secret**

#### Required Secrets

**Codecov Integration:**
```
CODECOV_TOKEN=<your-codecov-token>
```
Get from: https://codecov.io/

**Deployment Access:**
```
STAGING_SSH_KEY=<staging-private-key>
STAGING_HOST=user@staging.example.com
PRODUCTION_SSH_KEY=<production-private-key>
PRODUCTION_HOST=user@production.example.com
```

**Notifications:**
```
SLACK_WEBHOOK=https://hooks.slack.com/services/YOUR/WEBHOOK/URL
```

#### Optional Secrets
```
TEST_JWT=<jwt-for-smoke-tests>
DEPLOYMENT_SSH_KEY=<generic-deployment-key>
```

### 2. Set Up Environments

Create environments in **Settings** → **Environments**:

**Staging Environment:**
- Name: `staging`
- URL: `https://staging.gatekeeper.example.com`
- Protection rules: None (auto-deploy)

**Production Environment:**
- Name: `production`
- URL: `https://gatekeeper.example.com`
- Protection rules:
  - ✅ Required reviewers (1-2 people)
  - ✅ Wait timer (5 minutes)
  - ✅ Deployment branches (main only)

### 3. Configure Branch Protection

Navigate to **Settings** → **Branches** → **Add rule**

**Branch:** `main`

Protection rules:
- ✅ Require pull request reviews (1 approval)
- ✅ Require status checks to pass before merging
  - `test`
  - `lint`
  - `gosec`
  - `govulncheck`
- ✅ Require branches to be up to date
- ✅ Do not allow bypassing the above settings

---

## Workflow Details

### Test Workflow

**File:** `.github/workflows/test.yml`

**Purpose:** Ensures code quality and test coverage

**Runs on:**
- Every push to `main` or `develop`
- Every pull request
- Manual trigger

**Jobs:**

1. **test** - Run all tests
   - Sets up PostgreSQL 15 test database
   - Runs `go test` with race detector
   - Generates coverage report
   - Checks 85% coverage threshold
   - Uploads to Codecov

2. **lint** - Code quality checks
   - golangci-lint for code quality
   - go fmt for formatting
   - go vet for suspicious constructs

**Duration:** ~3-5 minutes

**Artifacts:**
- `coverage-report` - Coverage files (30 days)

---

### Build Workflow

**File:** `.github/workflows/build.yml`

**Purpose:** Build binaries and Docker images

**Runs on:**
- Push to `main` branch
- Version tags (v*)
- Manual trigger

**Jobs:**

1. **build-backend** - Cross-platform binaries
   - Matrix build: linux, darwin, windows
   - Architectures: amd64, arm64
   - Embeds version information
   - Creates compressed archives

2. **build-docker** - Container images
   - Multi-arch: linux/amd64, linux/arm64
   - Pushes to GitHub Container Registry
   - Tags: version, latest, SHA
   - Layer caching enabled

3. **scan-docker** - Security scanning
   - Trivy vulnerability scanner
   - Checks for CRITICAL and HIGH severity
   - Uploads SARIF to GitHub Security

4. **release** - Create GitHub release (on tags)
   - Attaches all binary artifacts
   - Docker image instructions
   - Changelog (if available)
   - Slack notification

**Duration:** ~8-10 minutes

**Artifacts:**
- Binary archives for all platforms (30 days)
- Docker images in GHCR

---

### Security Workflow

**File:** `.github/workflows/security.yml`

**Purpose:** Comprehensive security analysis

**Runs on:**
- Every push to `main` or `develop`
- Every pull request
- Daily at 2:00 AM UTC (schedule)
- Manual trigger

**Jobs:**

1. **gosec** - Go security scanner
   - Scans for security vulnerabilities
   - Uploads SARIF to GitHub Security
   - Creates JSON report

2. **govulncheck** - Vulnerability database
   - Checks Go vulnerability database
   - Scans all dependencies
   - Fails on any vulnerabilities

3. **dependency-review** - PR dependency check
   - Analyzes new dependencies in PRs
   - Checks licenses
   - Vulnerability alerts

4. **codeql** - Static analysis
   - Advanced pattern detection
   - Security-extended queries
   - Integrates with GitHub Security

5. **secret-scan** - Credential leak detection
   - TruffleHog scanner
   - Scans git history
   - Only verified secrets

6. **license-check** - License compliance
   - go-licenses tool
   - Checks forbidden licenses
   - Generates license report

7. **security-summary** - Aggregated results
   - Combined status
   - Summary table

**Duration:** ~2-3 minutes

**Artifacts:**
- `gosec-report` - Security scan results (30 days)
- `govulncheck-report` - Vulnerability check (30 days)
- `license-report` - License compliance (30 days)

---

### Deploy Workflow

**File:** `.github/workflows/deploy.yml`

**Purpose:** Automated deployment to environments

**Runs on:**
- Release published
- Manual trigger (with environment selection)

**Jobs:**

1. **pre-deploy-checks** - Validation
   - Verifies release exists
   - Checks Docker image availability
   - Validates version tags

2. **deploy-staging** - Staging deployment
   - SSH to staging server
   - Pull Docker images
   - Update docker-compose
   - Run smoke tests
   - Health checks
   - Deployment status tracking

3. **deploy-production** - Production deployment
   - **Requires manual approval**
   - Database backup
   - Blue-green deployment
   - Health checks
   - Automatic rollback on failure
   - Traffic switching

4. **post-deploy** - Post-deployment
   - Generate deployment report
   - Update documentation
   - Send notifications

**Duration:** ~10-15 minutes (excluding approval wait)

---

## Triggering Workflows

### Automatic Triggers

**On Push to Main:**
```bash
git push origin main
```
Triggers: Test, Build

**On Pull Request:**
```bash
gh pr create
```
Triggers: Test, Security, Dependency Review

**On Release:**
```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```
Triggers: Build, Deploy

**Scheduled (Daily 2 AM UTC):**
Triggers: Security

### Manual Triggers

**Via GitHub UI:**
1. Navigate to **Actions** tab
2. Select workflow from left sidebar
3. Click **Run workflow** button
4. Select branch and parameters
5. Click **Run workflow**

**Via GitHub CLI:**
```bash
# Test workflow
gh workflow run test.yml

# Build workflow
gh workflow run build.yml

# Security workflow
gh workflow run security.yml

# Deploy workflow (with parameters)
gh workflow run deploy.yml \
  -f environment=staging \
  -f skip_tests=false
```

---

## Viewing Results

### Workflow Status

**GitHub UI:**
1. Go to **Actions** tab
2. See all workflow runs
3. Filter by workflow, status, branch
4. Click run for detailed view

**Status Badges:**
Add to README.md:
```markdown
![Test](https://github.com/roguedan/gatekeeper/workflows/Test/badge.svg)
![Build](https://github.com/roguedan/gatekeeper/workflows/Build%20and%20Release/badge.svg)
![Security](https://github.com/roguekeeper/gatekeeper/workflows/Security%20Scan/badge.svg)
```

### Job Logs

1. Click on workflow run
2. Expand job in left sidebar
3. Click on step to view logs
4. Use search to find specific messages

### Artifacts

1. Navigate to workflow run
2. Scroll to **Artifacts** section
3. Click artifact name to download

**Example:**
```bash
# Download coverage report
gh run download <run-id> -n coverage-report

# List all artifacts
gh run view <run-id>
```

### Coverage Reports

**Codecov Dashboard:**
- Visit: https://codecov.io/gh/roguedan/gatekeeper
- View coverage trends
- Per-file coverage
- PR coverage diff

**GitHub Summary:**
- Coverage percentage in workflow summary
- Coverage report in test step

---

## Debugging Failed Workflows

### Common Failure Scenarios

#### Test Failures

**Symptom:** Test job fails

**Debug steps:**
```bash
# 1. Check logs for failing test
# 2. Run locally
go test ./internal/... -v -race

# 3. Check database connectivity
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/gatekeeper_test?sslmode=disable"
go test ./internal/store/... -v

# 4. Check coverage
go test ./internal/... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total
```

**Common causes:**
- Race conditions
- Database schema changes
- Missing test data
- Flaky tests

#### Build Failures

**Symptom:** Build job fails

**Debug steps:**
```bash
# 1. Verify Go version
go version  # Should be 1.21+

# 2. Check dependencies
go mod verify
go mod download

# 3. Build locally
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -o gatekeeper ./cmd/server

# 4. Test Docker build
docker build -t gatekeeper:test .
```

**Common causes:**
- Go version mismatch
- Missing dependencies
- Invalid Dockerfile
- Platform-specific code

#### Security Failures

**Symptom:** Security job fails

**Debug steps:**
```bash
# 1. Run gosec locally
go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec ./...

# 2. Check vulnerabilities
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# 3. Update dependencies
go get -u ./...
go mod tidy
```

**Common causes:**
- Known vulnerabilities in dependencies
- Security anti-patterns
- Exposed secrets
- Unsafe code patterns

#### Deploy Failures

**Symptom:** Deploy job fails

**Debug steps:**
1. Check SSH connectivity
   ```bash
   ssh -i ~/.ssh/deploy_key user@host
   ```

2. Verify Docker on server
   ```bash
   ssh user@host "docker ps"
   ```

3. Check health endpoint
   ```bash
   curl -f https://staging.gatekeeper.example.com/health
   ```

4. Review server logs
   ```bash
   ssh user@host "docker logs gatekeeper-backend"
   ```

**Common causes:**
- SSH key issues
- Network connectivity
- Service not starting
- Health check timeout

### Re-running Workflows

**Full re-run:**
```bash
gh run rerun <run-id>
```

**Failed jobs only:**
```bash
gh run rerun <run-id> --failed
```

**Via UI:**
1. Open failed run
2. Click **Re-run jobs** dropdown
3. Select **Re-run all jobs** or **Re-run failed jobs**

---

## Managing Secrets

### Adding Secrets

**Via GitHub UI:**
1. Settings → Secrets and variables → Actions
2. Click **New repository secret**
3. Name: `SECRET_NAME`
4. Value: `secret-value`
5. Click **Add secret**

**Via GitHub CLI:**
```bash
gh secret set CODECOV_TOKEN < token.txt
gh secret set SLACK_WEBHOOK --body "https://hooks.slack.com/..."
```

### Updating Secrets

**Via GitHub UI:**
1. Settings → Secrets and variables → Actions
2. Find secret
3. Click **Update**
4. Enter new value

**Via GitHub CLI:**
```bash
gh secret set SECRET_NAME --body "new-value"
```

### Using Secrets in Workflows

**In workflow YAML:**
```yaml
env:
  MY_SECRET: ${{ secrets.MY_SECRET }}

# Or in specific step
- name: Use secret
  run: echo "Using secret"
  env:
    TOKEN: ${{ secrets.TOKEN }}
```

**Security best practices:**
- ✅ Never echo secrets in logs
- ✅ Use minimal permissions
- ✅ Rotate secrets regularly
- ✅ Use environment-specific secrets
- ❌ Don't commit secrets to code
- ❌ Don't use secrets in URLs

### Secret Rotation

**Recommended rotation schedule:**
- SSH keys: Every 90 days
- API tokens: Every 60 days
- Webhooks: Every 180 days

**Rotation procedure:**
1. Generate new secret
2. Update in production
3. Update in GitHub Secrets
4. Test workflows
5. Revoke old secret

---

## Deployment Process

### Staging Deployment

**Automatic on release:**
```bash
# Create and push tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# Wait for build workflow
# Deploy workflow triggers automatically
```

**Manual deployment:**
```bash
gh workflow run deploy.yml -f environment=staging
```

**Steps:**
1. Pre-deployment checks
2. SSH to staging server
3. Pull Docker images
4. Update docker-compose
5. Restart services
6. Run smoke tests
7. Health checks
8. Notify team

**Rollback staging:**
```bash
# SSH to server
ssh user@staging.example.com

# Rollback to previous version
cd /opt/gatekeeper
docker-compose -f docker-compose.backup.yml up -d
```

### Production Deployment

**Requires approval:**
1. Release published → Build completes
2. Staging deployment succeeds
3. Production deployment waits for approval
4. Reviewer approves in GitHub
5. Production deployment proceeds

**Steps:**
1. Database backup
2. Blue-green deployment
   - Start new containers (green)
   - Health checks
   - Switch traffic
   - Stop old containers (blue)
3. Comprehensive health checks
4. Production smoke tests
5. Notify team
6. Auto-rollback on failure

**Manual production deployment:**
```bash
gh workflow run deploy.yml -f environment=production
```

**Production rollback:**
Automatic on health check failure, or manual:
```bash
# SSH to server
ssh user@production.example.com

# Rollback
cd /opt/gatekeeper
docker-compose -f docker-compose.backup.yml up -d

# Verify
curl -f https://gatekeeper.example.com/health
```

### Deployment Checklist

**Pre-deployment:**
- [ ] All tests passing
- [ ] Security scans clean
- [ ] Database migrations tested
- [ ] Changelog updated
- [ ] Team notified

**During deployment:**
- [ ] Monitor logs
- [ ] Watch metrics
- [ ] Verify health checks
- [ ] Test critical endpoints

**Post-deployment:**
- [ ] Smoke tests passed
- [ ] Metrics normal
- [ ] No error spikes
- [ ] Team notified
- [ ] Document issues

---

## Best Practices

### Workflow Development

**1. Test locally first:**
```bash
# Use act for local testing
act pull_request -W .github/workflows/test.yml

# Or run commands manually
go test ./internal/... -v
docker build -t test .
```

**2. Use matrix builds wisely:**
```yaml
strategy:
  matrix:
    os: [ubuntu-latest, macos-latest]
    go: ['1.21', '1.22']
```

**3. Fail fast:**
```yaml
strategy:
  fail-fast: true
  matrix: ...
```

**4. Cache dependencies:**
```yaml
- uses: actions/setup-go@v5
  with:
    go-version: '1.21'
    cache: true
```

### Security

**1. Minimal permissions:**
```yaml
permissions:
  contents: read
  # Only add what's needed
```

**2. Pin action versions:**
```yaml
# Bad
- uses: actions/checkout@main

# Good
- uses: actions/checkout@v4
```

**3. Use GITHUB_TOKEN:**
```yaml
# Automatically provided, no secret needed
token: ${{ secrets.GITHUB_TOKEN }}
```

**4. Validate inputs:**
```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Validate environment
        run: |
          if [[ ! "${{ inputs.environment }}" =~ ^(staging|production)$ ]]; then
            echo "Invalid environment"
            exit 1
          fi
```

### Performance

**1. Use artifacts efficiently:**
```yaml
- uses: actions/upload-artifact@v4
  with:
    name: my-artifact
    path: dist/
    retention-days: 7  # Don't keep forever
```

**2. Conditional jobs:**
```yaml
jobs:
  deploy:
    if: github.ref == 'refs/heads/main'
```

**3. Concurrent jobs:**
```yaml
jobs:
  test:
    runs-on: ubuntu-latest
  lint:
    runs-on: ubuntu-latest  # Runs in parallel
```

**4. Optimize Docker builds:**
```yaml
- uses: docker/build-push-action@v5
  with:
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

---

## Troubleshooting

### Workflow Not Triggering

**Check:**
1. Workflow file in `.github/workflows/`
2. Correct trigger events
3. Branch protection rules
4. Workflow enabled (not disabled)

**Fix:**
```bash
# Verify workflow syntax
gh workflow list

# Enable workflow if disabled
gh workflow enable <workflow-name>
```

### Slow Workflows

**Optimization strategies:**

1. **Enable caching:**
   ```yaml
   - uses: actions/cache@v3
     with:
       path: ~/go/pkg/mod
       key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
   ```

2. **Parallel jobs:**
   ```yaml
   jobs:
     job1:
       runs-on: ubuntu-latest
     job2:
       runs-on: ubuntu-latest  # Parallel
   ```

3. **Skip unnecessary steps:**
   ```yaml
   - name: Optional step
     if: github.event_name == 'release'
     run: ...
   ```

### Secret Access Issues

**Symptoms:**
- "Secret not found" error
- Empty secret value

**Debug:**
```yaml
- name: Check secret exists
  run: |
    if [ -z "${{ secrets.MY_SECRET }}" ]; then
      echo "Secret MY_SECRET not set"
      exit 1
    fi
```

**Fix:**
1. Verify secret name (case-sensitive)
2. Check repository vs environment secrets
3. Ensure secret is set for correct environment

### Permission Errors

**Symptoms:**
- "Resource not accessible by integration"
- "Permission denied"

**Fix:**
```yaml
permissions:
  contents: write  # If pushing to repo
  packages: write  # If publishing packages
  deployments: write  # If creating deployments
```

---

## Additional Resources

### Documentation
- [GitHub Actions Docs](https://docs.github.com/en/actions)
- [Workflow Syntax](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions)
- [Security Hardening](https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions)

### Tools
- [act - Local GitHub Actions](https://github.com/nektos/act)
- [actionlint - Workflow Linter](https://github.com/rhysd/actionlint)
- [GitHub CLI](https://cli.github.com/)

### Monitoring
- [GitHub Actions Dashboard](https://github.com/features/actions)
- [Codecov Dashboard](https://codecov.io/)
- [GitHub Security](https://github.com/security)

---

**Last Updated:** November 1, 2025
**Version:** 1.0.0
**Maintained by:** Gatekeeper Team
