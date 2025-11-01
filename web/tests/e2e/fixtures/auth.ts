import { test as base } from '@playwright/test';

/**
 * Authentication Fixtures for Playwright E2E Tests
 *
 * Purpose: Provide reusable authenticated page fixtures
 * Framework: Agentic QE with Playwright
 */

/**
 * Mock JWT token for testing
 * This is a sample token structure - in real tests, you'd use actual tokens
 */
export const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZGRyZXNzIjoiMHg3NDJkMzVDYzY2MzRDMDUzMjkyNWEzYjg0NEJjOWU3NTk1ZjBiRWIiLCJpYXQiOjE3MDAwMDAwMDAsImV4cCI6MTcwMDg2NDAwMH0.test_signature';

/**
 * Mock Ethereum address for testing
 */
export const mockAddress = '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb';

/**
 * Extended test fixture with authentication helpers
 */
type AuthFixtures = {
  authenticatedPage: any;
};

/**
 * Authenticated page fixture
 * Sets up a page with pre-authenticated state
 */
export const authenticatedPage = base.extend<AuthFixtures>({
  authenticatedPage: async ({ page, context }, use) => {
    // Set up authenticated state before each test
    await context.addInitScript(({ token, address }) => {
      // Set wallet connection state
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('wagmi.wallet', 'metaMask');
      localStorage.setItem('wagmi.account', address);

      // Set authentication token
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT, address: mockAddress });

    // Navigate to the application
    await page.goto('/');
    await page.waitForTimeout(1000);

    // Use the authenticated page in tests
    await use(page);
  },
});

/**
 * Helper to set authenticated state in any test
 */
export async function setAuthenticatedState(
  context: any,
  token: string = mockJWT,
  address: string = mockAddress
) {
  await context.addInitScript(
    ({ token, address }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('wagmi.wallet', 'metaMask');
      localStorage.setItem('wagmi.account', address);
      localStorage.setItem('gatekeeper_auth_token', token);
    },
    { token, address }
  );
}

/**
 * Helper to clear authentication state
 */
export async function clearAuthState(page: any) {
  await page.evaluate(() => {
    localStorage.removeItem('wagmi.connected');
    localStorage.removeItem('wagmi.wallet');
    localStorage.removeItem('wagmi.account');
    localStorage.removeItem('gatekeeper_auth_token');
    sessionStorage.clear();
  });
}

/**
 * Helper to check if page is authenticated
 */
export async function isAuthenticated(page: any): Promise<boolean> {
  const token = await page.evaluate(() => localStorage.getItem('gatekeeper_auth_token'));
  const connected = await page.evaluate(() => localStorage.getItem('wagmi.connected'));

  return !!(token && connected === 'true');
}

/**
 * Helper to get stored auth token
 */
export async function getStoredToken(page: any): Promise<string | null> {
  return await page.evaluate(() => localStorage.getItem('gatekeeper_auth_token'));
}

/**
 * Helper to get stored wallet address
 */
export async function getStoredAddress(page: any): Promise<string | null> {
  return await page.evaluate(() => localStorage.getItem('wagmi.account'));
}

/**
 * Mock expired JWT token
 */
export const expiredJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZGRyZXNzIjoiMHg3NDJkMzVDYzY2MzRDMDUzMjkyNWEzYjg0NEJjOWU3NTk1ZjBiRWIiLCJpYXQiOjE2MDAwMDAwMDAsImV4cCI6MTYwMDAwMDAwMX0.expired';

/**
 * Helper to set expired token state
 */
export async function setExpiredTokenState(context: any) {
  await context.addInitScript(({ token, address }) => {
    localStorage.setItem('wagmi.connected', 'true');
    localStorage.setItem('wagmi.wallet', 'metaMask');
    localStorage.setItem('wagmi.account', address);
    localStorage.setItem('gatekeeper_auth_token', token);
  }, { token: expiredJWT, address: mockAddress });
}

/**
 * Helper to wait for authentication to complete
 */
export async function waitForAuthentication(page: any, timeout: number = 5000): Promise<boolean> {
  try {
    await page.waitForFunction(
      () => {
        const token = localStorage.getItem('gatekeeper_auth_token');
        const connected = localStorage.getItem('wagmi.connected');
        return !!(token && connected === 'true');
      },
      { timeout }
    );
    return true;
  } catch {
    return false;
  }
}

/**
 * Helper to wait for logout to complete
 */
export async function waitForLogout(page: any, timeout: number = 5000): Promise<boolean> {
  try {
    await page.waitForFunction(
      () => {
        const token = localStorage.getItem('gatekeeper_auth_token');
        return !token;
      },
      { timeout }
    );
    return true;
  } catch {
    return false;
  }
}
