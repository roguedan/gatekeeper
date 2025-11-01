# Gatekeeper E2E Tests

Comprehensive end-to-end test suite for Gatekeeper using Playwright and Agentic QE framework best practices.

## Overview

This test suite provides complete coverage of the Gatekeeper authentication flow, including:
- Wallet connection (RainbowKit)
- SIWE (Sign-In with Ethereum) authentication
- API key management (creation, listing, revocation)
- Complete user journeys
- Security validations
- Mobile responsiveness

## Test Statistics

- **Total Tests:** 124 tests (62 per browser)
- **Test Files:** 4 spec files
- **Browsers:** Chromium, Firefox
- **Coverage Areas:** 8+ functional areas

## Directory Structure

```
tests/e2e/
├── README.md                          # This file
├── fixtures/
│   └── auth.ts                        # Authentication fixtures and helpers
├── helpers/
│   ├── api.ts                         # API interaction helpers
│   └── wallet.ts                      # Wallet mocking and interaction helpers
└── tests/
    ├── 01-wallet-connection.spec.ts   # Wallet connection flow tests (12 tests)
    ├── 02-siwe-authentication.spec.ts # SIWE authentication tests (14 tests)
    ├── 03-api-key-management.spec.ts  # API key management tests (20 tests)
    └── 04-complete-journey.spec.ts    # End-to-end user journeys (16 tests)
```

## Test Suites

### 01-wallet-connection.spec.ts (12 tests)
Tests wallet provider connection functionality:
- Wallet connection button visibility
- RainbowKit modal behavior
- Wallet provider options (MetaMask, WalletConnect)
- Connection state persistence
- Error handling and rejection flows
- Mobile responsiveness

### 02-siwe-authentication.spec.ts (14 tests)
Tests Sign-In with Ethereum authentication:
- SIWE nonce retrieval from backend
- Message generation and signing
- JWT token generation and validation
- Token persistence in localStorage
- Token expiration handling
- Security validations (no token exposure, logout cleanup)

### 03-api-key-management.spec.ts (20 tests)
Tests API key lifecycle management:
- API keys dashboard navigation
- Empty state display
- API key creation flow with name and scopes
- API key listing and display
- Key metadata (name, date, scopes)
- API key revocation with confirmation
- Permission-based access control

### 04-complete-journey.spec.ts (16 tests)
Tests complete user flows:
- New user journey: Connect → Sign → Create API Key
- Authentication persistence across navigation
- Browser refresh handling
- Returning user auto-authentication
- Logout and state cleanup
- Cross-page navigation
- Error handling (network errors, invalid tokens)
- Mobile experience
- Performance benchmarks

## Running Tests

### Prerequisites

```bash
cd /Users/danwilliams/Documents/web3/gatekeeper/web

# Install dependencies (if not already done)
npm install

# Install Playwright browsers
npx playwright install
```

### Run All Tests

```bash
# Run all E2E tests
npm run test:e2e

# Run with UI mode (interactive)
npm run test:e2e:ui

# Run in headed mode (see browser)
npm run test:e2e:headed

# Run in debug mode
npm run test:e2e:debug
```

### Run Specific Tests

```bash
# Run specific test file
npx playwright test tests/01-wallet-connection.spec.ts

# Run tests matching a pattern
npx playwright test --grep "wallet connection"

# Run single browser
npx playwright test --project=chromium
```

### View Test Reports

```bash
# Show HTML report
npm run test:e2e:report

# Generate and view report after test run
npx playwright show-report playwright-report
```

### Test Code Generation

```bash
# Record new tests interactively
npm run test:e2e:codegen
```

## Configuration

### Playwright Configuration

Located at: `/Users/danwilliams/Documents/web3/gatekeeper/web/playwright.config.ts`

Key settings:
- **Base URL:** http://localhost:3000
- **Timeout:** 30 seconds per test
- **Retries:** 2 on CI, 0 locally
- **Workers:** 1 on CI, unlimited locally
- **Reporters:** HTML, JSON, JUnit, List

### Environment Variables

```bash
# API base URL (default: http://localhost:8080)
export API_BASE_URL=http://localhost:8080

# Run in CI mode
export CI=true
```

## Helper Functions

### Authentication Helpers (fixtures/auth.ts)

```typescript
import { setAuthenticatedState, clearAuthState } from '../fixtures/auth';

// Set authenticated state
await setAuthenticatedState(context, token, address);

// Clear authentication
await clearAuthState(page);

// Check if authenticated
const isAuth = await isAuthenticated(page);
```

### API Helpers (helpers/api.ts)

```typescript
import { getNonce, createAPIKey, deleteAPIKey } from '../helpers/api';

// Get SIWE nonce
const nonce = await getNonce(request);

// Create API key
const apiKey = await createAPIKey(request, token, 'My Key', ['read', 'write']);

// Delete API key
await deleteAPIKey(request, token, keyId);
```

### Wallet Helpers (helpers/wallet.ts)

```typescript
import { installMockWallet, signSIWEMessage } from '../helpers/wallet';

// Install mock wallet provider
await installMockWallet(page, address, chainId);

// Sign SIWE message
const signature = await signSIWEMessage(page, message);

// Connect mock wallet
const address = await connectMockWallet(page);
```

## Best Practices

### 1. Test Isolation
- Each test is independent and can run in any order
- Tests clean up their own state
- Use `beforeEach` for setup, `afterEach` for cleanup

### 2. Authentication
- Use fixtures for authenticated state
- Mock JWT tokens for testing
- Don't rely on real wallet connections in tests

### 3. Assertions
- Use Playwright's auto-waiting assertions
- Add explicit timeouts when needed
- Test both positive and negative cases

### 4. Selectors
- Prefer role-based selectors: `getByRole('button', { name: /connect/i })`
- Use data-testid for complex components
- Avoid CSS selectors when possible

### 5. Error Handling
- Test error states explicitly
- Verify error messages are displayed
- Test recovery from errors

## CI/CD Integration

### GitHub Actions Example

```yaml
name: E2E Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '18'

      - name: Install dependencies
        run: |
          cd web
          npm ci

      - name: Install Playwright Browsers
        run: |
          cd web
          npx playwright install --with-deps

      - name: Run E2E tests
        run: |
          cd web
          npm run test:e2e

      - name: Upload test results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: playwright-report
          path: web/playwright-report/
```

## Test Reports

After running tests, reports are generated in multiple formats:

- **HTML Report:** `playwright-report/index.html` (interactive)
- **JSON Report:** `test-results/results.json` (machine-readable)
- **JUnit Report:** `test-results/junit.xml` (CI integration)

## Debugging Tests

### Debug Mode

```bash
# Run in debug mode with inspector
npm run test:e2e:debug

# Debug specific test
npx playwright test --debug tests/01-wallet-connection.spec.ts
```

### Visual Debugging

```bash
# Run with UI mode
npm run test:e2e:ui

# Run headed (see browser)
npm run test:e2e:headed
```

### Traces

Traces are automatically captured on first retry. View them:

```bash
npx playwright show-trace trace.zip
```

## Troubleshooting

### Tests Timeout
- Increase timeout in playwright.config.ts
- Check if backend is running (http://localhost:8080)
- Verify frontend is accessible (http://localhost:3000)

### Backend Not Available
```bash
# Ensure backend is running
cd /Users/danwilliams/Documents/web3/gatekeeper
docker-compose up
```

### Frontend Not Available
```bash
# Start frontend dev server
cd /Users/danwilliams/Documents/web3/gatekeeper/web
npm run dev
```

### Browser Not Installed
```bash
# Install Playwright browsers
npx playwright install
```

## Writing New Tests

### Test Template

```typescript
import { test, expect } from '@playwright/test';

test.describe('Feature Name', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should do something', async ({ page }) => {
    // Arrange
    const button = page.getByRole('button', { name: /click me/i });

    // Act
    await button.click();

    // Assert
    await expect(page.getByText(/success/i)).toBeVisible();
  });
});
```

### Using Authenticated Fixture

```typescript
import { test, expect } from '@playwright/test';
import { setAuthenticatedState } from '../fixtures/auth';

test('authenticated test', async ({ page, context }) => {
  // Set up authentication
  await setAuthenticatedState(context);
  await page.goto('/dashboard');

  // Your test logic
  await expect(page.getByRole('heading', { name: /dashboard/i })).toBeVisible();
});
```

## Coverage Goals

- ✅ **Wallet Connection:** 100% coverage
- ✅ **SIWE Authentication:** 100% coverage
- ✅ **API Key Management:** 100% coverage
- ✅ **User Journeys:** 100% coverage
- ✅ **Error Handling:** 100% coverage
- ✅ **Mobile Responsiveness:** 100% coverage
- ✅ **Security Validations:** 100% coverage

## Contributing

When adding new tests:

1. Follow the existing file naming convention
2. Group related tests in describe blocks
3. Add descriptive test names
4. Include comments for complex logic
5. Update this README with new test counts
6. Ensure tests pass on all browsers

## Resources

- [Playwright Documentation](https://playwright.dev/docs/intro)
- [Agentic QE Framework](https://github.com/agentic-qe)
- [SIWE Specification](https://eips.ethereum.org/EIPS/eip-4361)
- [RainbowKit Documentation](https://www.rainbowkit.com/docs/introduction)

## License

Same as the Gatekeeper project.
