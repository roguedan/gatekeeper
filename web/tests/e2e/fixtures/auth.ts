import { Page, BrowserContext } from '@playwright/test'

export interface AuthSetupOptions {
  address?: string
  token?: string
  waitForLoad?: boolean
}

/**
 * Setup authenticated state by setting JWT token in localStorage
 * CRITICAL: Must be called BEFORE first page.goto() to inject auth into localStorage
 * The init script is applied to ALL subsequent page loads within the same context
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

  // CRITICAL: Setup auth in context BEFORE any page navigation
  // This ensures localStorage is initialized when page loads and app initializes
  // The init script runs on EVERY page load in this context
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

  // Navigate to home first to initialize auth context
  // The init script above will have already set localStorage for this navigation
  await page.goto('/')

  // Wait for auth to initialize and React to fully render
  if (waitForLoad) {
    try {
      // Wait for the root element to be interactive
      await page.waitForLoadState('networkidle')
    } catch {
      // Some routes may not complete networkidle, just wait for DOM to settle
      await page.waitForTimeout(2000)
    }

    // Additional check: verify auth token is actually in localStorage
    const tokenInStorage = await page.evaluate(() => localStorage.getItem('gatekeeper_auth_token'))
    if (!tokenInStorage) {
      throw new Error('Auth token not found in localStorage after setupAuthenticatedUser')
    }
  }

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
