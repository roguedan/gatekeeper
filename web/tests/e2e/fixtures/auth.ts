import { Page, BrowserContext } from '@playwright/test'

export interface AuthSetupOptions {
  address?: string
  token?: string
  waitForLoad?: boolean
}

/**
 * Setup authenticated state by setting JWT token in localStorage
 * IMPORTANT: Must be called with context.addInitScript BEFORE first page navigation
 * This function MUST be called BEFORE any page.goto() calls
 */
export async function setupAuthenticatedUser(
  page: Page,
  context: BrowserContext,
  options: AuthSetupOptions = {}
) {
  const {
    address = '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb',
    token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature',
    waitForLoad = true,
  } = options

  // CRITICAL: Setup auth in context BEFORE any navigation
  // This ensures localStorage is initialized when page loads and app initializes
  await context.addInitScript(
    ({ token, address }) => {
      // Set auth token (checked by authService.isAuthenticated())
      localStorage.setItem('gatekeeper_auth_token', token)
      // Set wallet address
      localStorage.setItem('gatekeeper_wallet_address', address)
      // Mark for E2E test environment
      ;(window as any).__E2E_TEST__ = true
      ;(window as any).__AUTH_ADDRESS__ = address
    },
    { token, address }
  )

  // Return without navigating - let the test handle navigation
  // This way the init script is applied to ALL subsequent page loads
  return { address, token }
}

/**
 * Verify user is authenticated by checking localStorage
 */
export async function verifyAuthenticated(page: Page) {
  const token = await page.evaluate(() => localStorage.getItem('gatekeeper_auth_token'))
  const address = await page.evaluate(() => localStorage.getItem('gatekeeper_wallet_address'))

  return {
    authenticated: !!token,
    token,
    address,
  }
}

/**
 * Verify user is NOT authenticated
 */
export async function verifyUnauthenticated(page: Page) {
  const token = await page.evaluate(() => localStorage.getItem('gatekeeper_auth_token'))
  const address = await page.evaluate(() => localStorage.getItem('gatekeeper_wallet_address'))

  return {
    authenticated: !token && !address,
    token: null,
    address: null,
  }
}

/**
 * Clear all auth state
 */
export async function clearAuth(page: Page) {
  await page.evaluate(() => {
    localStorage.removeItem('gatekeeper_auth_token')
    localStorage.removeItem('gatekeeper_wallet_address')
  })
}
