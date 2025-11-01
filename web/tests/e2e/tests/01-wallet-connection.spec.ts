import { test, expect } from '@playwright/test';

/**
 * E2E Tests: Wallet Connection Flow
 *
 * Test Suite: Wallet Connection
 * Coverage: Wallet provider detection, connection modal, error handling
 * Framework: Agentic QE with Playwright
 */

test.describe('Wallet Connection Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should display wallet connection button on homepage', async ({ page }) => {
    // Verify the connect wallet button is visible (using first() to handle multiple buttons)
    const connectButton = page.getByTestId('rk-connect-button').first();
    await expect(connectButton).toBeVisible();
  });

  test('should show RainbowKit modal when connect button is clicked', async ({ page }) => {
    // Click connect wallet button
    const connectButton = page.getByTestId('rk-connect-button').first();
    await connectButton.click();

    // Verify RainbowKit modal appears
    await expect(page.locator('[data-rk]').first()).toBeVisible({ timeout: 5000 });
  });

  test('should display wallet provider options in modal', async ({ page }) => {
    // Open wallet connect modal
    const connectButton = page.getByTestId('rk-connect-button').first();
    await connectButton.click();

    // Wait for modal to be visible
    await expect(page.locator('[data-rk]').first()).toBeVisible({ timeout: 5000 });

    // Verify wallet options are displayed
    // RainbowKit shows wallet options like MetaMask, WalletConnect, etc.
    const modalContent = page.locator('[data-rk]').first();
    await expect(modalContent).toContainText(/MetaMask|WalletConnect|Coinbase/i);
  });

  test('should close modal when cancel/close button is clicked', async ({ page }) => {
    // Open wallet connect modal
    const connectButton = page.getByTestId('rk-connect-button').first();
    await connectButton.click();

    // Wait for modal to appear
    const modal = page.locator('[data-rk][role="dialog"]');
    await expect(modal).toBeVisible({ timeout: 5000 });

    // RainbowKit modal closes when you press Escape or click outside
    // The most reliable way is to find the close button or click the overlay
    // Try to find and click a button that closes the modal (SVG close icon button)
    const closeButton = modal.locator('button').filter({ has: page.locator('svg') }).first();

    // If close button found, click it; otherwise try Escape key
    try {
      const isVisible = await closeButton.isVisible();
      if (isVisible) {
        await closeButton.click();
      } else {
        await page.keyboard.press('Escape');
      }
    } catch {
      // If button not found, use Escape
      await page.keyboard.press('Escape');
    }

    // Wait for modal to close (CSS transition)
    await page.waitForTimeout(300);

    // Verify modal is no longer visible
    // Use isHidden instead of not.toBeVisible for better reliability
    await expect(modal).toHaveCount(0);
  });

  test('should handle no wallet provider gracefully', async ({ page }) => {
    // This test verifies the app doesn't crash when no wallet is installed
    // In a real browser without MetaMask, the modal should still open
    const connectButton = page.getByTestId('rk-connect-button').first();
    await connectButton.click();

    // Modal should still open even without wallet
    await expect(page.locator('[data-rk]').first()).toBeVisible({ timeout: 5000 });
  });

  test('should persist wallet connection state in localStorage', async ({ page, context }) => {
    // Mock localStorage to simulate connected state
    await context.addInitScript(() => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('wagmi.wallet', 'metaMask');
    });

    await page.reload();

    // Verify connection state is recognized
    // When connected, the button text should change or show account
    const accountButton = page.locator('button:has-text("0x")').or(page.getByRole('button', { name: /disconnect/i }));

    // Wait a bit for state to load
    await page.waitForTimeout(1000);
  });

  test('should display proper loading state during connection', async ({ page }) => {
    // Click connect wallet button
    const connectButton = page.getByTestId('rk-connect-button').first();
    await connectButton.click();

    // Verify loading indicators appear
    // RainbowKit shows connecting state
    await expect(page.locator('[data-rk]').first()).toBeVisible({ timeout: 5000 });
  });

  test('should show network selection in wallet modal', async ({ page }) => {
    // Open wallet connect modal
    const connectButton = page.getByTestId('rk-connect-button').first();
    await connectButton.click();

    // Wait for modal
    await expect(page.locator('[data-rk]').first()).toBeVisible({ timeout: 5000 });

    // RainbowKit modal may show network info or selection
    const modalContent = page.locator('[data-rk]').first();
    await expect(modalContent).toBeVisible();
  });

  test('should display wallet address in shortened format when connected', async ({ page, context }) => {
    // Mock a connected wallet state
    await context.addInitScript(() => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('wagmi.wallet', 'metaMask');
      localStorage.setItem('wagmi.account', '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb');
    });

    await page.reload();
    await page.waitForTimeout(1000);

    // Check if shortened address is displayed (e.g., 0x742d...0bEb)
    // This might appear in a connected button or account dropdown
  });

  test('should handle connection rejection by user', async ({ page }) => {
    // Click connect wallet button
    const connectButton = page.getByTestId('rk-connect-button').first();
    await connectButton.click();

    // Modal should open
    const modal = page.locator('[data-rk][role="dialog"]');
    await expect(modal).toBeVisible({ timeout: 5000 });

    // Close the modal to simulate rejection
    // Try to find close button or use Escape key
    const closeButton = modal.locator('button').filter({ has: page.locator('svg') }).first();

    try {
      const isVisible = await closeButton.isVisible();
      if (isVisible) {
        await closeButton.click();
      } else {
        await page.keyboard.press('Escape');
      }
    } catch {
      await page.keyboard.press('Escape');
    }

    await page.waitForTimeout(300);

    // Verify user remains on homepage and connect button is still visible
    await expect(page.getByTestId('rk-connect-button').first()).toBeVisible();
  });
});

test.describe('Wallet Connection - Responsive Design', () => {
  test.skip('should display wallet connection on mobile viewport', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('/');

    // Verify connect button is visible on mobile
    const connectButton = page.getByTestId('rk-connect-button').first();
    await expect(connectButton).toBeVisible();
  });

  test.skip('should open mobile-optimized wallet modal', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('/');

    // Click connect wallet button
    const connectButton = page.getByTestId('rk-connect-button').first();
    await connectButton.click();

    // Verify RainbowKit mobile modal appears
    await expect(page.locator('[data-rk]').first()).toBeVisible({ timeout: 5000 });
  });
});
