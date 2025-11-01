# Gatekeeper Testing Guide - Quick Reference

**Last Updated:** 2025-11-01
**Version:** 1.0

This guide provides quick reference instructions for running all types of tests in the Gatekeeper project.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Running Unit Tests (Go)](#running-unit-tests-go)
3. [Running E2E Tests (Playwright)](#running-e2e-tests-playwright)
4. [Running Integration Tests](#running-integration-tests)
5. [Generating Coverage Reports](#generating-coverage-reports)
6. [Debugging Tests](#debugging-tests)
7. [Test Organization](#test-organization)
8. [Common Issues](#common-issues)

---

## Prerequisites

### Required Software
- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- PostgreSQL 15+ (via Docker)
- Redis 7+ (via Docker)

### Environment Setup
```bash
# Clone repository
git clone <repository-url>
cd gatekeeper

# Install Go dependencies
go mod download

# Install Node.js dependencies
cd web
npm install
cd ..

# Start Docker services
docker-compose up -d

# Wait for services to be healthy
docker-compose ps
```

### Environment Variables
```bash
# Copy example environment file
cp .env.example .env

# Required variables
DATABASE_URL=postgresql://gatekeeper:gatekeeper@localhost:5432/gatekeeper?sslmode=disable
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-secret-key-change-in-production
ETHEREUM_RPC_URL=https://mainnet.infura.io/v3/YOUR-PROJECT-ID
```

---

## Running Unit Tests (Go)

### Run All Tests
```bash
# From project root
go test ./...

# With verbose output
go test -v ./...

# With coverage
go test -cover ./...
```

### Run Tests in Specific Package
```bash
# Test authentication package
go test ./internal/auth/...

# Test HTTP handlers
go test ./internal/http/...

# Test store layer
go test ./internal/store/...

# Test policy engine
go test ./internal/policy/...

# Test RPC proxy
go test ./internal/rpc/...

# Test rate limiting
go test ./internal/ratelimit/...
```

### Run Specific Test
```bash
# Run single test by name
go test ./internal/auth -run TestSIWEService_GenerateNonce

# Run tests matching pattern
go test ./internal/auth -run TestSIWE

# Run tests with timeout
go test -timeout 30s ./internal/auth
```

### Parallel Execution
```bash
# Run tests in parallel (default)
go test ./...

# Control parallelism
go test -parallel 4 ./...

# Run sequentially
go test -parallel 1 ./...
```

---

## Running E2E Tests (Playwright)

### Initial Setup
```bash
# Navigate to web directory
cd web

# Install Playwright browsers
npx playwright install

# Install system dependencies (Linux)
npx playwright install-deps
```

### Run All E2E Tests
```bash
# From web directory
npm run test:e2e

# With UI mode (interactive)
npm run test:e2e:ui

# With headed browsers (visible)
npm run test:e2e:headed

# With specific browser
npm run test:e2e -- --project=chromium
npm run test:e2e -- --project=firefox
npm run test:e2e -- --project=webkit
```

### Run Specific Test Files
```bash
# Run wallet connection tests
npm run test:e2e tests/e2e/01-wallet-connection.spec.ts

# Run authentication tests
npm run test:e2e tests/e2e/02-siwe-authentication.spec.ts

# Run API key management tests
npm run test:e2e tests/e2e/03-api-key-management.spec.ts

# Run complete user journey tests
npm run test:e2e tests/e2e/04-complete-user-journey.spec.ts
```

### Run Specific Test Cases
```bash
# Run tests matching title
npm run test:e2e -- --grep "should connect MetaMask"

# Run tests NOT matching title
npm run test:e2e -- --grep-invert "should handle error"

# Run tests by line number
npm run test:e2e tests/e2e/01-wallet-connection.spec.ts:15
```

### Debug Mode
```bash
# Debug mode with Playwright Inspector
npm run test:e2e -- --debug

# Debug specific test
npm run test:e2e tests/e2e/02-siwe-authentication.spec.ts --debug

# Debug with headed browser
PWDEBUG=1 npm run test:e2e
```

### View Test Reports
```bash
# Generate and open HTML report
npx playwright show-report

# Generate report without opening
npx playwright show-report --no-open

# Report is saved to: web/playwright-report/
```

---

## Running Integration Tests

### Start Required Services
```bash
# Start all services
docker-compose up -d

# Verify services are healthy
docker-compose ps

# View service logs
docker-compose logs -f postgres
docker-compose logs -f redis
```

### Run Integration Tests
```bash
# Integration tests are included in Go tests
# They use Docker services automatically

# Run all tests including integration
go test ./...

# Run only integration tests (tagged)
go test -tags=integration ./...

# Skip integration tests
go test -short ./...
```

### Test Database Migrations
```bash
# Migrations run automatically in tests
# To test manually:

# Connect to test database
docker exec -it gatekeeper-postgres psql -U gatekeeper -d gatekeeper

# View migrations
\dt

# Check migration status
SELECT * FROM schema_migrations;
```

---

## Generating Coverage Reports

### Go Coverage

#### Basic Coverage
```bash
# Show coverage percentage
go test -cover ./...

# Coverage by package
go test -cover ./internal/auth
go test -cover ./internal/http
go test -cover ./internal/store
```

#### Detailed Coverage Report
```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Open in browser
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

#### Coverage by Function
```bash
# Show coverage with function names
go tool cover -func=coverage.out

# Filter by package
go tool cover -func=coverage.out | grep "internal/auth"

# Sort by coverage
go tool cover -func=coverage.out | sort -k3 -n
```

### Playwright Coverage

Playwright focuses on E2E test coverage, not code coverage. To track E2E coverage:

```bash
# Run tests and generate report
npm run test:e2e

# View test results
npx playwright show-report

# Test statistics in report:
# - Total tests run
# - Pass/fail counts
# - Duration
# - Flaky tests
```

---

## Debugging Tests

### Debugging Go Tests

#### Print Debug Information
```go
func TestExample(t *testing.T) {
    // Use t.Log for debug output
    t.Logf("Debug: value = %v", value)

    // Use t.Error to fail but continue
    if value != expected {
        t.Errorf("Expected %v, got %v", expected, value)
    }

    // Use t.Fatal to fail and stop
    if critical {
        t.Fatalf("Critical failure: %v", err)
    }
}
```

#### Run with Verbose Output
```bash
# Verbose mode shows all t.Log output
go test -v ./internal/auth

# Show test names as they run
go test -v ./...
```

#### Debug with Delve
```bash
# Install delve debugger
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug test
dlv test ./internal/auth

# Debug specific test
dlv test ./internal/auth -- -test.run TestSIWEService_GenerateNonce

# Set breakpoints in code
# (dlv) break internal/auth/siwe.go:42
# (dlv) continue
```

### Debugging Playwright Tests

#### Debug Mode
```bash
# Open Playwright Inspector
npm run test:e2e -- --debug

# Debug specific test
npx playwright test tests/e2e/01-wallet-connection.spec.ts --debug

# Debug from specific line
npx playwright test tests/e2e/01-wallet-connection.spec.ts:15 --debug
```

#### Headed Mode (See Browser)
```bash
# Run with visible browser
npm run test:e2e -- --headed

# Slow down execution
npm run test:e2e -- --headed --slow-mo=1000
```

#### Screenshots and Videos
```bash
# Screenshots on failure (default)
npm run test:e2e

# Screenshots always
npx playwright test --screenshot=on

# Videos on failure
npx playwright test --video=retain-on-failure

# Videos always
npx playwright test --video=on

# Files saved to: web/test-results/
```

#### Console Output
```typescript
// Add debug logging in tests
test('example', async ({ page }) => {
  // Log page console
  page.on('console', msg => console.log('PAGE:', msg.text()));

  // Log network requests
  page.on('request', req => console.log('REQUEST:', req.url()));

  // Take screenshot
  await page.screenshot({ path: 'debug.png' });

  // Pause execution
  await page.pause();
});
```

---

## Test Organization

### Go Test Structure

```
gatekeeper/
├── internal/
│   ├── auth/
│   │   ├── siwe.go              # Implementation
│   │   ├── siwe_test.go         # Unit tests
│   │   └── jwt_test.go
│   ├── http/
│   │   ├── handlers.go
│   │   ├── handlers_test.go     # Unit tests
│   │   ├── middleware_test.go
│   │   └── api_key_handlers_test.go
│   ├── store/
│   │   ├── user_repository.go
│   │   ├── users_test.go        # Unit tests
│   │   ├── api_keys_test.go
│   │   ├── test_helpers.go      # Test utilities
│   │   └── db_test.go
│   ├── policy/
│   │   ├── engine.go
│   │   └── engine_test.go       # 145+ tests
│   └── rpc/
│       ├── proxy.go
│       └── proxy_test.go        # 36 tests, 95.9% coverage
```

### Playwright Test Structure

```
web/tests/
├── e2e/
│   ├── 01-wallet-connection.spec.ts      # 12 tests
│   ├── 02-siwe-authentication.spec.ts    # 14 tests
│   ├── 03-api-key-management.spec.ts     # 20 tests
│   └── 04-complete-user-journey.spec.ts  # 16 tests
├── fixtures/
│   ├── test-users.ts           # Test user data
│   ├── mock-wallet.ts          # MetaMask mocks
│   └── test-helpers.ts         # Utility functions
└── playwright.config.ts        # Playwright configuration
```

### Test Naming Conventions

#### Go Tests
```go
// Format: TestFunctionName_Scenario
func TestSIWEService_GenerateNonce(t *testing.T)
func TestSIWEService_VerifySignature_ValidSignature(t *testing.T)
func TestSIWEService_VerifySignature_InvalidSignature(t *testing.T)
func TestAPIKeyHandler_CreateKey_Success(t *testing.T)
func TestAPIKeyHandler_CreateKey_Unauthorized(t *testing.T)
```

#### Playwright Tests
```typescript
// Describe blocks for grouping
describe('Wallet Connection', () => {
  test('should display connection button', async ({ page }) => {
    // Test implementation
  });

  test('should connect MetaMask successfully', async ({ page }) => {
    // Test implementation
  });
});
```

---

## Common Issues

### Issue: Tests Fail Due to Missing Database

**Symptom:**
```
Error: failed to connect to database: connection refused
```

**Solution:**
```bash
# Start PostgreSQL
docker-compose up -d postgres

# Wait for health check
docker-compose ps postgres

# Verify connection
docker exec -it gatekeeper-postgres psql -U gatekeeper -c "SELECT 1"
```

### Issue: Migration Errors in Tests

**Symptom:**
```
Error: relation "users" already exists
```

**Solution:**
```bash
# Use test database isolation
# Tests should use IF NOT EXISTS in migrations

# Or reset test database
docker-compose down -v
docker-compose up -d
```

### Issue: Port Already in Use

**Symptom:**
```
Error: bind: address already in use
```

**Solution:**
```bash
# Find process using port
lsof -i :8080

# Kill process
kill -9 <PID>

# Or use different port in .env
PORT=8081
```

### Issue: Playwright Browsers Not Installed

**Symptom:**
```
Error: browserType.launch: Executable doesn't exist
```

**Solution:**
```bash
# Install browsers
cd web
npx playwright install

# Install system dependencies (Linux)
npx playwright install-deps
```

### Issue: Test Timeout

**Symptom:**
```
Error: test timeout of 30000ms exceeded
```

**Solution:**
```bash
# Increase timeout for Go tests
go test -timeout 60s ./...

# Increase timeout for Playwright
npx playwright test --timeout=60000

# Or in test file:
test.setTimeout(60000);
```

### Issue: Flaky Tests

**Symptom:**
Tests pass sometimes, fail other times

**Solution:**
```bash
# Run tests multiple times to identify flakes
go test -count=10 ./internal/auth

# Playwright retry configuration
npx playwright test --retries=3

# Or in config:
use: {
  retries: 3,
}
```

### Issue: Coverage Not Generated

**Symptom:**
```
Error: coverage: cannot use test profile flag with multiple packages
```

**Solution:**
```bash
# Use correct syntax for multiple packages
go test -coverprofile=coverage.out ./...

# Not this (wrong):
go test ./... -coverprofile=coverage.out
```

---

## Quick Reference Commands

### Most Common Commands
```bash
# Run all Go tests
go test ./...

# Run all E2E tests
cd web && npm run test:e2e

# Generate coverage report
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

# Debug single test
go test -v ./internal/auth -run TestSIWEService_GenerateNonce

# Debug Playwright test
cd web && npm run test:e2e -- --debug

# View test report
cd web && npx playwright show-report
```

### Service Management
```bash
# Start services
docker-compose up -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f

# Restart service
docker-compose restart postgres
```

---

## Additional Resources

- [Go Testing Documentation](https://pkg.go.dev/testing)
- [Playwright Documentation](https://playwright.dev)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Project README](./README.md)
- [Testing Summary](./TESTING_SUMMARY.md)

---

**For questions or issues, contact the development team or create an issue in the repository.**
