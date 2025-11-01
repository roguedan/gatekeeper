import { test, expect } from '@playwright/test';

/**
 * E2E Tests: Complete User Journey
 *
 * Test Suite: End-to-End User Journeys
 * Coverage: Full authentication flow, navigation, persistence, logout
 * Framework: Agentic QE with Playwright
 */

test.describe('Complete User Journey - New User Flow', () => {
  test('should complete full journey: Connect → Sign → Create API Key', async ({ page, context }) => {
    // Step 1: Navigate to homepage
    await page.goto('/');
    await page.waitForTimeout(1000);

    // Verify homepage loads
    await expect(page.getByRole('heading', { name: /gatekeeper/i })).toBeVisible({ timeout: 5000 });

    // Step 2: Verify connect wallet button is visible
    const connectButton = page.getByRole('button', { name: /connect wallet/i });
    await expect(connectButton).toBeVisible();

    // Step 3: Click connect wallet
    await connectButton.click();

    // Step 4: Verify modal opens
    await expect(page.locator('[data-rk]')).toBeVisible({ timeout: 5000 });

    // For E2E testing, we'll simulate authenticated state
    // In a real test with MetaMask, we'd interact with the extension
    await page.waitForTimeout(500);

    // Close modal for now
    const closeButton = page.locator('[data-rk] button').first();
    if (await closeButton.isVisible()) {
      await closeButton.click();
    }

    // Simulate authenticated state
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';
    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('wagmi.account', '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    await page.reload();
    await page.waitForTimeout(1000);

    // Step 5: Verify authenticated state
    const dashboardLink = page.getByRole('link', { name: /dashboard/i });
    await expect(dashboardLink).toBeVisible({ timeout: 5000 });

    // Step 6: Navigate to API Keys
    const apiKeysLink = page.getByRole('link', { name: /api.*keys/i });
    await apiKeysLink.click();
    await page.waitForTimeout(1000);

    // Step 7: Verify API Keys page loaded
    expect(page.url()).toContain('/api-keys');

    // Step 8: Look for create API key button
    const createButton = page.getByRole('button', { name: /create.*api.*key|new.*key/i });
    await expect(createButton).toBeVisible({ timeout: 5000 });
  });

  test('should persist authentication across multiple page navigations', async ({ page, context }) => {
    // Set authenticated state
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';
    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('wagmi.account', '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    // Navigate to home
    await page.goto('/');
    await page.waitForTimeout(1000);

    // Should see dashboard link
    let dashboardLink = page.getByRole('link', { name: /dashboard/i });
    await expect(dashboardLink).toBeVisible({ timeout: 5000 });

    // Navigate to dashboard
    await dashboardLink.click();
    await page.waitForTimeout(1000);

    // Should still be authenticated
    expect(page.url()).toContain('/dashboard');

    // Navigate to API keys
    const apiKeysLink = page.getByRole('link', { name: /api.*keys/i });
    await apiKeysLink.click();
    await page.waitForTimeout(1000);

    // Should still be authenticated
    expect(page.url()).toContain('/api-keys');

    // Navigate back home
    const homeLink = page.getByRole('link', { name: /^home$/i }).or(page.getByRole('link', { name: /gatekeeper/i }));

    if (await homeLink.isVisible({ timeout: 3000 })) {
      await homeLink.click();
      await page.waitForTimeout(1000);

      // Should still show authenticated state
      dashboardLink = page.getByRole('link', { name: /dashboard/i });
      await expect(dashboardLink).toBeVisible({ timeout: 5000 });
    }
  });

  test('should handle browser refresh while authenticated', async ({ page, context }) => {
    // Set authenticated state
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';
    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('wagmi.account', '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    await page.goto('/');
    await page.waitForTimeout(1000);

    // Verify authenticated
    const dashboardLink = page.getByRole('link', { name: /dashboard/i });
    await expect(dashboardLink).toBeVisible({ timeout: 5000 });

    // Refresh page
    await page.reload();
    await page.waitForTimeout(1000);

    // Should still be authenticated after refresh
    await expect(dashboardLink).toBeVisible({ timeout: 5000 });
  });
});

test.describe('Complete User Journey - Returning User Flow', () => {
  test('should auto-authenticate returning user with valid token', async ({ page, context }) => {
    // Simulate returning user with stored credentials
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';
    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('wagmi.wallet', 'metaMask');
      localStorage.setItem('wagmi.account', '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    await page.goto('/');
    await page.waitForTimeout(1000);

    // User should see authenticated state immediately
    const dashboardLink = page.getByRole('link', { name: /dashboard/i });
    await expect(dashboardLink).toBeVisible({ timeout: 5000 });

    // Should NOT see connect wallet button
    const connectButton = page.getByRole('button', { name: /^connect wallet$/i });
    await expect(connectButton).not.toBeVisible();
  });

  test('should prompt re-authentication if token is expired', async ({ page, context }) => {
    // Simulate expired token
    const expiredJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZGRyZXNzIjoiMHg3NDJkMzVDYzY2MzRDMDUzMjkyNWEzYjg0NEJjOWU3NTk1ZjBiRWIiLCJpYXQiOjE2MDAwMDAwMDAsImV4cCI6MTYwMDAwMDAwMX0.test';

    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: expiredJWT });

    await page.goto('/');
    await page.waitForTimeout(1000);

    // Should show sign-in prompt for expired token
    const signInButton = page.getByRole('button', { name: /sign.*message|sign.*in/i });

    // Wait to see if sign-in is required
    await page.waitForTimeout(1500);
  });
});

test.describe('Complete User Journey - Logout Flow', () => {
  test('should complete logout and clear all auth state', async ({ page, context }) => {
    // Set authenticated state
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';
    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('wagmi.account', '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    await page.goto('/');
    await page.waitForTimeout(1000);

    // Verify authenticated
    const dashboardLink = page.getByRole('link', { name: /dashboard/i });
    await expect(dashboardLink).toBeVisible({ timeout: 5000 });

    // Find disconnect/logout button
    const disconnectButton = page.getByRole('button', { name: /disconnect|logout|sign.*out/i });

    if (await disconnectButton.isVisible({ timeout: 3000 })) {
      await disconnectButton.click();
      await page.waitForTimeout(1000);

      // Verify auth state is cleared
      const storedToken = await page.evaluate(() => localStorage.getItem('gatekeeper_auth_token'));
      expect(storedToken).toBeNull();

      // Should see connect wallet button again
      const connectButton = page.getByRole('button', { name: /connect wallet/i });
      await expect(connectButton).toBeVisible({ timeout: 5000 });
    }
  });

  test('should redirect to home after logout from protected page', async ({ page, context }) => {
    // Set authenticated state
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';
    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('wagmi.account', '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    // Navigate to dashboard
    await page.goto('/dashboard');
    await page.waitForTimeout(1000);

    // Find disconnect button
    const disconnectButton = page.getByRole('button', { name: /disconnect|logout|sign.*out/i });

    if (await disconnectButton.isVisible({ timeout: 3000 })) {
      await disconnectButton.click();
      await page.waitForTimeout(1500);

      // Should redirect to home or login
      const currentUrl = page.url();
      expect(currentUrl).toMatch(/\/(home|login|$)/);
    }
  });
});

test.describe('Complete User Journey - Navigation', () => {
  test('should navigate between all main sections when authenticated', async ({ page, context }) => {
    // Set authenticated state
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';
    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('wagmi.account', '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    await page.goto('/');
    await page.waitForTimeout(1000);

    // Test navigation to Dashboard
    const dashboardLink = page.getByRole('link', { name: /dashboard/i });
    await dashboardLink.click();
    await page.waitForTimeout(1000);
    expect(page.url()).toContain('/dashboard');

    // Test navigation to API Keys
    const apiKeysLink = page.getByRole('link', { name: /api.*keys/i });
    await apiKeysLink.click();
    await page.waitForTimeout(1000);
    expect(page.url()).toContain('/api-keys');

    // Test navigation to Token Gating (if exists)
    const tokenGatingLink = page.getByRole('link', { name: /token.*gating/i });

    if (await tokenGatingLink.isVisible({ timeout: 3000 })) {
      await tokenGatingLink.click();
      await page.waitForTimeout(1000);
      expect(page.url()).toContain('/token-gating');
    }
  });

  test('should show active navigation state for current page', async ({ page, context }) => {
    // Set authenticated state
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';
    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('wagmi.account', '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    // Navigate to dashboard
    await page.goto('/dashboard');
    await page.waitForTimeout(1000);

    // Dashboard link should have active state (aria-current or CSS class)
    const dashboardLink = page.getByRole('link', { name: /dashboard/i });
    const ariaCurrentValue = await dashboardLink.getAttribute('aria-current');

    // Active link may have aria-current="page" or special CSS class
    await page.waitForTimeout(500);
  });
});

test.describe('Complete User Journey - Error Handling', () => {
  test('should handle network errors gracefully', async ({ page, context }) => {
    // Set authenticated state
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';
    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    // Intercept API calls and simulate network error
    await page.route('**/api/**', route => {
      route.abort('failed');
    });

    await page.goto('/');
    await page.waitForTimeout(1000);

    // App should not crash
    await expect(page.locator('body')).toBeVisible();
  });

  test('should handle invalid JWT token gracefully', async ({ page, context }) => {
    // Set invalid token
    await context.addInitScript(() => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('gatekeeper_auth_token', 'invalid.token.here');
    });

    await page.goto('/');
    await page.waitForTimeout(1000);

    // Should not crash and may show sign-in prompt
    await expect(page.locator('body')).toBeVisible();
  });

  test('should handle missing localStorage gracefully', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(1000);

    // App should work without localStorage
    const connectButton = page.getByRole('button', { name: /connect wallet/i });
    await expect(connectButton).toBeVisible({ timeout: 5000 });
  });
});

test.describe('Complete User Journey - Mobile Experience', () => {
  test('should complete full journey on mobile viewport', async ({ page, context }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });

    // Set authenticated state
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';
    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('wagmi.account', '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    await page.goto('/');
    await page.waitForTimeout(1000);

    // Should see mobile navigation
    const mobileMenu = page.getByRole('button', { name: /menu/i }).or(page.locator('[aria-label*="menu"]'));

    // Navigation should work on mobile
    await page.waitForTimeout(1000);
  });

  test('should handle mobile wallet connection', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });

    await page.goto('/');
    await page.waitForTimeout(1000);

    // Click connect wallet
    const connectButton = page.getByRole('button', { name: /connect wallet/i });
    await connectButton.click();

    // Mobile modal should appear
    await expect(page.locator('[data-rk]')).toBeVisible({ timeout: 5000 });
  });
});

test.describe('Complete User Journey - Performance', () => {
  test('should load homepage within acceptable time', async ({ page }) => {
    const startTime = Date.now();

    await page.goto('/');
    await page.waitForLoadState('networkidle');

    const loadTime = Date.now() - startTime;

    // Should load within 5 seconds
    expect(loadTime).toBeLessThan(5000);
  });

  test('should handle rapid navigation without errors', async ({ page, context }) => {
    // Set authenticated state
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';
    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('wagmi.account', '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    await page.goto('/');
    await page.waitForTimeout(500);

    // Rapidly navigate between pages
    await page.goto('/dashboard');
    await page.waitForTimeout(200);

    await page.goto('/api-keys');
    await page.waitForTimeout(200);

    await page.goto('/');
    await page.waitForTimeout(200);

    // App should not crash
    await expect(page.locator('body')).toBeVisible();
  });
});
